package v1

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/http/api"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/http/httperr"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/http/httputils"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/http/middleware"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/http/token"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/ports"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/rs/zerolog/log"
)

type createUserValidation struct {
	FullName string `validate:"required"`
	Email    string `validate:"required,email"`
	Password string `validate:"required"`
}

type createUserJSONBody struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginUserValidation struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required"`
}

type loginUserJSONBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type renewAccessTokenValidation struct {
	RefreshToken string `validate:"required"`
}

func (adapter *HTTPAdapter) createUser(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
	requestBody, err := httputils.Decode[createUserJSONBody](r)
	if err != nil {
		return err
	}

	reqUser := api.CreateUserRequest{
		FullName: requestBody.FullName,
		Email:    openapi_types.Email(requestBody.Email),
		Password: requestBody.Password,
	}

	validationErr := validate.Struct(createUserValidation{
		FullName: requestBody.FullName,
		Email:    requestBody.Email,
		Password: requestBody.Password,
	})
	if validationErr != nil {
		return httperr.MatchValidationError(validationErr)
	}

	createdUser, err := adapter.userService.CreateUser(r.Context(), ports.NewUser{
		FullName: reqUser.FullName,
		Email:    string(reqUser.Email),
		Password: reqUser.Password,
	})
	if err != nil {
		return err
	}

	log.Info().Msg("AAAAAAAAAAAA")

	responseEmail := openapi_types.Email(createdUser.Email)
	responseFullName := createdUser.FullName
	responseUID := openapi_types.UUID(createdUser.UID)

	err = httputils.Encode(w, r, http.StatusCreated, api.CreateUserResponse{
		Uid:      &responseUID,
		FullName: &responseFullName,
		Email:    &responseEmail,
	})
	if err != nil {
		return err
	}

	return nil
}

func (adapter *HTTPAdapter) loginUser(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
	requestBody, err := httputils.Decode[loginUserJSONBody](r)
	if err != nil {
		return err
	}

	reqUser := api.LoginUserRequest{
		Email:    openapi_types.Email(requestBody.Email),
		Password: requestBody.Password,
	}

	validationErr := validate.Struct(loginUserValidation{
		Email:    requestBody.Email,
		Password: requestBody.Password,
	})
	if validationErr != nil {
		return httperr.MatchValidationError(validationErr)
	}

	user, err := adapter.userService.LoginUser(r.Context(), ports.LoginUser{
		Email:    string(reqUser.Email),
		Password: reqUser.Password,
	})
	if err != nil {
		return err
	}

	accessToken, accessPayload, err := adapter.tokenMaker.CreateToken(user.Email, "user", adapter.config.AccessTokenDuration)
	if err != nil {
		return err
	}

	refreshToken, refreshPayload, err := adapter.tokenMaker.CreateToken(user.Email, "user", adapter.config.RefreshTokenDuration)
	if err != nil {
		return err
	}

	responseEmail := openapi_types.Email(user.Email)
	loginRes := api.LoginUserResponse{
		Email:                 &responseEmail,
		AccessToken:           &accessToken,
		AccessTokenExpiresAt:  &accessPayload.ExpiredAt,
		RefreshToken:          &refreshToken,
		RefreshTokenExpiresAt: &refreshPayload.ExpiredAt,
	}

	_, err = adapter.userService.CreateUserSession(r.Context(), ports.NewUserSession{
		RefreshTokenID:        refreshPayload.ID,
		Email:                 user.Email,
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
	renewAccessDto, err := httputils.Decode[api.RenewAccessTokenRequest](r)
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

	session, err := adapter.userService.GetUserSession(r.Context(), refreshPayload.ID)
	if err != nil {
		return err
	}

	if session.IsBlocked {
		return domainerr.NewDomainError(http.StatusUnauthorized, domainerr.UnauthorizedError, "session is blocked", errors.New("session is blocked"))
	}

	if session.UserEmail != refreshPayload.Email {
		return domainerr.NewDomainError(http.StatusUnauthorized, domainerr.UnauthorizedError, "invalid session user", errors.New("invalid session user"))
	}

	if session.RefreshToken != renewAccessDto.RefreshToken {
		return domainerr.NewDomainError(http.StatusUnauthorized, domainerr.UnauthorizedError, "invalid refresh token", errors.New("invalid refresh token"))
	}

	if time.Now().After(session.ExpiresAt) {
		return domainerr.NewDomainError(http.StatusUnauthorized, domainerr.UnauthorizedError, "session expired", errors.New("session expired"))
	}

	accessToken, accessPayload, err := adapter.tokenMaker.CreateToken(session.UserEmail, "user", adapter.config.AccessTokenDuration)
	if err != nil {
		return err
	}

	renewAccessTokenResponse := api.RenewAccessTokenResponse{
		AccessToken:          &accessToken,
		AccessTokenExpiresAt: &accessPayload.ExpiredAt,
	}

	err = httputils.Encode(w, r, http.StatusOK, renewAccessTokenResponse)
	if err != nil {
		return err
	}

	return nil
}

func (adapter *HTTPAdapter) getUserByUID(w http.ResponseWriter, r *http.Request, uid string) *domainerr.DomainError {
	payload := r.Context().Value(middleware.KeyAuthUser).(*token.Payload)
	requestID := middleware.GetRequestID(r.Context())

	fmt.Println(payload)
	fmt.Println(requestID)

	user, serviceErr := adapter.userService.GetUserByUID(r.Context(), uid)
	if serviceErr != nil {
		return serviceErr
	}

	responseEmail := openapi_types.Email(user.Email)
	responseFullName := user.FullName
	responseUID := openapi_types.UUID(user.UID)

	err := httputils.Encode(w, r, http.StatusOK, api.CreateUserResponse{
		Uid:      &responseUID,
		Email:    &responseEmail,
		FullName: &responseFullName,
	})
	if err != nil {
		return err
	}

	return nil
}
