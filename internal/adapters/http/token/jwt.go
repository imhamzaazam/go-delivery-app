package token

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
)

type jwtClaims struct {
	Email      string `json:"email"`
	Role       string `json:"role"`
	MerchantID string `json:"merchant_id"`
	jwt.RegisteredClaims
}

type JWTMaker struct {
	secretKey string
}

const minJWTSecretKeyLength = 32

func NewJWTMaker(secretKey string) (*JWTMaker, error) {
	if len(secretKey) < minJWTSecretKeyLength {
		return nil, fmt.Errorf("secret key must be at least %d characters", minJWTSecretKeyLength)
	}

	return &JWTMaker{
		secretKey: secretKey,
	}, nil
}

func (maker *JWTMaker) CreateToken(email string, role string, merchantID uuid.UUID, duration time.Duration) (string, *Payload, *domainerr.DomainError) {
	payload, payloadErr := NewPayload(email, role, merchantID, duration)
	if payloadErr != nil {
		return "", payload, payloadErr
	}

	claims := jwtClaims{
		Email:      email,
		Role:       role,
		MerchantID: merchantID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        payload.ID.String(),
			IssuedAt:  jwt.NewNumericDate(payload.IssuedAt),
			ExpiresAt: jwt.NewNumericDate(payload.ExpiredAt),
		},
	}

	tkn := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	encodedToken, err := tkn.SignedString([]byte(maker.secretKey))
	if err != nil {
		return "", nil, domainerr.NewInternalError(err)
	}

	return encodedToken, payload, nil
}

func (maker *JWTMaker) VerifyToken(token string) (*Payload, *domainerr.DomainError) {
	if maker == nil || len(maker.secretKey) == 0 {
		return nil, domainerr.NewDomainError(http.StatusInternalServerError, domainerr.UnexpectedError, "internal server error", ErrInvalidInstance)
	}

	claims := &jwtClaims{}
	decodedToken, err := jwt.ParseWithClaims(token, claims, func(parsedToken *jwt.Token) (interface{}, error) {
		if _, ok := parsedToken.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %s", parsedToken.Method.Alg())
		}

		return []byte(maker.secretKey), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, domainerr.NewDomainError(http.StatusUnauthorized, domainerr.ExpiredToken, "Expired Token", ErrExpiredToken)
		}

		return nil, domainerr.NewDomainError(http.StatusUnauthorized, domainerr.UnauthorizedError, "invalid token", ErrInvalidToken)
	}

	if decodedToken == nil || !decodedToken.Valid {
		return nil, domainerr.NewDomainError(http.StatusUnauthorized, domainerr.UnauthorizedError, "invalid token", ErrInvalidToken)
	}

	tokenID, parseErr := uuid.Parse(claims.ID)
	if parseErr != nil {
		return nil, domainerr.NewDomainError(http.StatusUnauthorized, domainerr.UnauthorizedError, "invalid token", ErrInvalidToken)
	}

	merchantID, merchantParseErr := uuid.Parse(claims.MerchantID)
	if merchantParseErr != nil {
		return nil, domainerr.NewDomainError(http.StatusUnauthorized, domainerr.UnauthorizedError, "invalid token", ErrInvalidToken)
	}

	payload := &Payload{
		ID:         tokenID,
		Email:      claims.Email,
		Role:       claims.Role,
		MerchantID: merchantID,
		IssuedAt:   claims.IssuedAt.Time,
		ExpiredAt:  claims.ExpiresAt.Time,
	}

	if validationErr := payload.Valid(); validationErr != nil {
		return nil, validationErr
	}

	return payload, nil
}
