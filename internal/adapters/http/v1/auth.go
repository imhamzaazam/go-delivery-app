package v1

import (
	"errors"
	"net/http"
	"time"

	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/http/httperr"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/http/httputils"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/ports"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

type loginActorValidation struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required"`
}

type renewAccessTokenValidation struct {
	RefreshToken string `validate:"required"`
}

func (adapter *HTTPAdapter) loginActor(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
	requestBody, err := httputils.Decode[LoginActorRequest](r)
	if err != nil {
		return err
	}

	validationErr := validate.Struct(loginActorValidation{
		Email:    string(requestBody.Email),
		Password: requestBody.Password,
	})
	if validationErr != nil {
		return httperr.MatchValidationError(validationErr)
	}

	actor, err := adapter.actorService.LoginActor(r.Context(), ports.LoginActor{
		MerchantID: requestBody.MerchantId,
		Email:      string(requestBody.Email),
		Password:   requestBody.Password,
	})
	if err != nil {
		return err
	}

	accessToken, accessPayload, err := adapter.tokenMaker.CreateToken(actor.Email, "actor", actor.MerchantID, adapter.config.AccessTokenDuration)
	if err != nil {
		return err
	}

	refreshToken, refreshPayload, err := adapter.tokenMaker.CreateToken(actor.Email, "actor", actor.MerchantID, adapter.config.RefreshTokenDuration)
	if err != nil {
		return err
	}

	responseEmail := openapi_types.Email(actor.Email)
	responseMerchantID := openapi_types.UUID(actor.MerchantID)
	responseUID := openapi_types.UUID(actor.UID)
	loginRes := LoginActorResponse{
		MerchantId:            &responseMerchantID,
		Uid:                   &responseUID,
		Email:                 &responseEmail,
		AccessToken:           &accessToken,
		AccessTokenExpiresAt:  &accessPayload.ExpiredAt,
		RefreshToken:          &refreshToken,
		RefreshTokenExpiresAt: &refreshPayload.ExpiredAt,
	}

	_, err = adapter.actorService.CreateActorSession(r.Context(), ports.NewActorSession{
		RefreshTokenID:        refreshPayload.ID,
		MerchantID:            actor.MerchantID,
		ActorID:               actor.UID,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
		UserAgent:             r.UserAgent(),
		ClientIP:              r.RemoteAddr,
	})
	if err != nil {
		return err
	}

	err = httputils.Encode(w, r, http.StatusOK, loginRes)
	if err != nil {
		return err
	}

	return nil
}

func (adapter *HTTPAdapter) renewAccessToken(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
	renewAccessDto, err := httputils.Decode[RenewAccessTokenRequest](r)
	if err != nil {
		return err
	}

	validationErr := validate.Struct(renewAccessTokenValidation{
		RefreshToken: renewAccessDto.RefreshToken,
	})
	if validationErr != nil {
		return httperr.MatchValidationError(validationErr)
	}

	refreshPayload, err := adapter.tokenMaker.VerifyToken(renewAccessDto.RefreshToken)
	if err != nil {
		return err
	}

	session, err := adapter.actorService.GetActorSession(r.Context(), refreshPayload.ID)
	if err != nil {
		return err
	}

	if session.IsBlocked {
		return domainerr.NewDomainError(http.StatusUnauthorized, domainerr.UnauthorizedError, "session is blocked", errors.New("session is blocked"))
	}

	if session.ActorEmail != refreshPayload.Email {
		return domainerr.NewDomainError(http.StatusUnauthorized, domainerr.UnauthorizedError, "invalid session actor", errors.New("invalid session actor"))
	}

	if session.RefreshToken != renewAccessDto.RefreshToken {
		return domainerr.NewDomainError(http.StatusUnauthorized, domainerr.UnauthorizedError, "invalid refresh token", errors.New("invalid refresh token"))
	}

	if time.Now().After(session.ExpiresAt) {
		return domainerr.NewDomainError(http.StatusUnauthorized, domainerr.UnauthorizedError, "session expired", errors.New("session expired"))
	}

	accessToken, accessPayload, err := adapter.tokenMaker.CreateToken(session.ActorEmail, "actor", refreshPayload.MerchantID, adapter.config.AccessTokenDuration)
	if err != nil {
		return err
	}

	renewAccessTokenResponse := RenewAccessTokenResponse{
		AccessToken:          &accessToken,
		AccessTokenExpiresAt: &accessPayload.ExpiredAt,
	}

	err = httputils.Encode(w, r, http.StatusOK, renewAccessTokenResponse)
	if err != nil {
		return err
	}

	return nil
}

func (adapter *HTTPAdapter) LoginActor(w http.ResponseWriter, r *http.Request) {
	adapter.handlerWrapper(adapter.loginActor)(w, r)
}

func (adapter *HTTPAdapter) RenewAccessToken(w http.ResponseWriter, r *http.Request) {
	adapter.handlerWrapper(adapter.renewAccessToken)(w, r)
}
