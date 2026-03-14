package presentation

import (
	"errors"
	"net/http"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"

	api "github.com/horiondreher/go-web-api-boilerplate/api"
	actordomain "github.com/horiondreher/go-web-api-boilerplate/internal/actor"
	"github.com/horiondreher/go-web-api-boilerplate/internal/core"
	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
	"github.com/horiondreher/go-web-api-boilerplate/pkg/http-tools/httperr"
	"github.com/horiondreher/go-web-api-boilerplate/pkg/http-tools/httputils"
)

type Handler struct {
	shared *core.Shared
}

type loginActorValidation struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required"`
}

type renewAccessTokenValidation struct {
	RefreshToken string `validate:"required"`
}

func New(shared *core.Shared) *Handler {
	return &Handler{shared: shared}
}

func (handler *Handler) loginActor(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
	requestBody, err := httputils.Decode[api.LoginActorRequest](r)
	if err != nil {
		return err
	}

	validationErr := handler.shared.Validate.Struct(loginActorValidation{
		Email:    string(requestBody.Email),
		Password: requestBody.Password,
	})
	if validationErr != nil {
		return httperr.MatchValidationError(validationErr)
	}

	actor, err := handler.shared.ActorService.LoginActor(r.Context(), actordomain.LoginActor{
		MerchantID: requestBody.MerchantId,
		Email:      string(requestBody.Email),
		Password:   requestBody.Password,
	})
	if err != nil {
		return err
	}

	accessToken, accessPayload, err := handler.shared.TokenMaker.CreateToken(actor.Email, "actor", actor.MerchantID, handler.shared.Config.AccessTokenDuration)
	if err != nil {
		return err
	}

	refreshToken, refreshPayload, err := handler.shared.TokenMaker.CreateToken(actor.Email, "actor", actor.MerchantID, handler.shared.Config.RefreshTokenDuration)
	if err != nil {
		return err
	}

	responseEmail := openapi_types.Email(actor.Email)
	responseMerchantID := openapi_types.UUID(actor.MerchantID)
	responseUID := openapi_types.UUID(actor.UID)
	loginRes := api.LoginActorResponse{
		MerchantId:            &responseMerchantID,
		Uid:                   &responseUID,
		Email:                 &responseEmail,
		AccessToken:           &accessToken,
		AccessTokenExpiresAt:  &accessPayload.ExpiredAt,
		RefreshToken:          &refreshToken,
		RefreshTokenExpiresAt: &refreshPayload.ExpiredAt,
	}

	_, err = handler.shared.ActorService.CreateActorSession(r.Context(), actordomain.NewActorSession{
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

	return httputils.Encode(w, r, http.StatusOK, loginRes)
}

func (handler *Handler) renewAccessToken(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
	renewAccessDTO, err := httputils.Decode[api.RenewAccessTokenRequest](r)
	if err != nil {
		return err
	}

	validationErr := handler.shared.Validate.Struct(renewAccessTokenValidation{
		RefreshToken: renewAccessDTO.RefreshToken,
	})
	if validationErr != nil {
		return httperr.MatchValidationError(validationErr)
	}

	refreshPayload, err := handler.shared.TokenMaker.VerifyToken(renewAccessDTO.RefreshToken)
	if err != nil {
		return err
	}

	session, err := handler.shared.ActorService.GetActorSession(r.Context(), refreshPayload.ID)
	if err != nil {
		return err
	}

	if session.IsBlocked {
		return domainerr.NewDomainError(http.StatusUnauthorized, domainerr.UnauthorizedError, "session is blocked", errors.New("session is blocked"))
	}
	if session.ActorEmail != refreshPayload.Email {
		return domainerr.NewDomainError(http.StatusUnauthorized, domainerr.UnauthorizedError, "invalid session actor", errors.New("invalid session actor"))
	}
	if session.RefreshToken != renewAccessDTO.RefreshToken {
		return domainerr.NewDomainError(http.StatusUnauthorized, domainerr.UnauthorizedError, "invalid refresh token", errors.New("invalid refresh token"))
	}
	if time.Now().After(session.ExpiresAt) {
		return domainerr.NewDomainError(http.StatusUnauthorized, domainerr.UnauthorizedError, "session expired", errors.New("session expired"))
	}

	accessToken, accessPayload, err := handler.shared.TokenMaker.CreateToken(session.ActorEmail, "actor", refreshPayload.MerchantID, handler.shared.Config.AccessTokenDuration)
	if err != nil {
		return err
	}

	response := api.RenewAccessTokenResponse{
		AccessToken:          &accessToken,
		AccessTokenExpiresAt: &accessPayload.ExpiredAt,
	}

	return httputils.Encode(w, r, http.StatusOK, response)
}

func (handler *Handler) LoginActor(w http.ResponseWriter, r *http.Request) {
	handler.shared.Wrap(handler.loginActor)(w, r)
}

func (handler *Handler) RenewAccessToken(w http.ResponseWriter, r *http.Request) {
	handler.shared.Wrap(handler.renewAccessToken)(w, r)
}
