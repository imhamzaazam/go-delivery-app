package core

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/rs/zerolog/log"

	"github.com/horiondreher/go-web-api-boilerplate/internal/actor"
	"github.com/horiondreher/go-web-api-boilerplate/internal/commerce"
	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
	"github.com/horiondreher/go-web-api-boilerplate/internal/merchant"
	"github.com/horiondreher/go-web-api-boilerplate/internal/utils"
	"github.com/horiondreher/go-web-api-boilerplate/pkg/http-tools/httputils"
	middleware "github.com/horiondreher/go-web-api-boilerplate/pkg/http-tools/middleware"
	"github.com/horiondreher/go-web-api-boilerplate/pkg/http-tools/token"
)

type DomainHandler func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError

const (
	GatewayMerchantIDHeader = "X-Merchant-Id"
	GatewaySecretHeader     = "X-Gateway-Secret"
)

type Shared struct {
	ActorService    actor.Service
	CommerceService commerce.Service
	MerchantService merchant.Service
	ReadService     commerce.ReadService
	Config          *utils.Config
	TokenMaker      *token.JWTMaker
	Validate        *validator.Validate
}

type AuthUser struct {
	Email      string
	MerchantID uuid.UUID
}

type ActorProfile struct {
	MerchantID uuid.UUID
	UID        uuid.UUID
	Email      string
}

func (shared *Shared) Wrap(handlerFn DomainHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if apiErr := handlerFn(w, r); apiErr != nil {
			var httpErrIntf *domainerr.DomainError
			var err *domainerr.DomainError

			requestID := middleware.GetRequestID(r.Context())

			if errors.As(apiErr, &httpErrIntf) {
				log.Info().Str("id", requestID).Str("error message", httpErrIntf.OriginalError).Msg("request error")
				err = httputils.Encode(w, r, httpErrIntf.HTTPCode, httpErrIntf.HTTPErrorBody)
			} else {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}

			if err != nil {
				log.Err(err).Msg("error encoding response")
			}
		}
	}
}

func (shared *Shared) ServeAuthenticated(w http.ResponseWriter, r *http.Request, handler DomainHandler) {
	authenticatedHandler := middleware.Authentication(shared.TokenMaker)(shared.Wrap(handler))
	authenticatedHandler.ServeHTTP(w, r)
}

func (shared *Shared) CurrentAuthUser(r *http.Request) (*AuthUser, *domainerr.DomainError) {
	authUser := middleware.GetAuthUser(r.Context())
	if authUser == nil {
		return nil, domainerr.NewDomainError(http.StatusUnauthorized, domainerr.UnauthorizedError, "unauthorized", errors.New("unauthorized"))
	}

	return &AuthUser{
		Email:      authUser.Email,
		MerchantID: authUser.MerchantID,
	}, nil
}

func (shared *Shared) CurrentActorProfile(r *http.Request) (*ActorProfile, *domainerr.DomainError) {
	authUser, authErr := shared.CurrentAuthUser(r)
	if authErr != nil {
		return nil, authErr
	}

	actor, err := shared.ActorService.GetActorProfileByMerchantAndEmail(r.Context(), authUser.MerchantID, authUser.Email)
	if err != nil {
		return nil, err
	}

	return &ActorProfile{
		MerchantID: actor.MerchantID,
		UID:        actor.UID,
		Email:      actor.Email,
	}, nil
}

func (shared *Shared) CurrentMerchantID(r *http.Request) (uuid.UUID, *domainerr.DomainError) {
	if strings.TrimSpace(r.Header.Get("Authorization")) != "" {
		authUser, authErr := shared.authUserFromAuthorizationHeader(r)
		if authErr != nil {
			return uuid.Nil, authErr
		}

		return authUser.MerchantID, nil
	}

	return shared.trustedGatewayMerchantID(r)
}

func (shared *Shared) authUserFromAuthorizationHeader(r *http.Request) (*AuthUser, *domainerr.DomainError) {
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	fields := strings.Fields(auth)
	if len(fields) < 2 {
		return nil, domainerr.NewDomainError(http.StatusUnauthorized, domainerr.UnauthorizedError, "invalid authorization header", errors.New("invalid authorization header"))
	}

	if strings.ToLower(fields[0]) != "bearer" {
		return nil, domainerr.NewDomainError(http.StatusUnauthorized, domainerr.UnauthorizedError, "invalid authorization type", errors.New("invalid authorization type"))
	}

	payload, err := shared.TokenMaker.VerifyToken(fields[1])
	if err != nil {
		return nil, err
	}

	return &AuthUser{
		Email:      payload.Email,
		MerchantID: payload.MerchantID,
	}, nil
}

func (shared *Shared) trustedGatewayMerchantID(r *http.Request) (uuid.UUID, *domainerr.DomainError) {
	if shared.Config == nil || strings.TrimSpace(shared.Config.GatewaySharedSecret) == "" {
		return uuid.Nil, domainerr.NewDomainError(http.StatusUnauthorized, domainerr.UnauthorizedError, "missing trusted merchant context", errors.New("gateway shared secret is not configured"))
	}

	providedSecret := strings.TrimSpace(r.Header.Get(GatewaySecretHeader))
	if providedSecret == "" || providedSecret != shared.Config.GatewaySharedSecret {
		return uuid.Nil, domainerr.NewDomainError(http.StatusUnauthorized, domainerr.UnauthorizedError, "missing trusted merchant context", errors.New("invalid gateway secret"))
	}

	merchantHeader := strings.TrimSpace(r.Header.Get(GatewayMerchantIDHeader))
	if merchantHeader == "" {
		return uuid.Nil, domainerr.NewDomainError(http.StatusUnauthorized, domainerr.UnauthorizedError, "missing trusted merchant context", errors.New("missing merchant header"))
	}

	merchantID, parseErr := uuid.Parse(merchantHeader)
	if parseErr != nil {
		return uuid.Nil, domainerr.NewDomainError(http.StatusUnauthorized, domainerr.UnauthorizedError, "invalid trusted merchant context", parseErr)
	}

	if _, err := shared.MerchantService.GetMerchant(r.Context(), merchantID.String()); err != nil {
		return uuid.Nil, domainerr.NewDomainError(http.StatusUnauthorized, domainerr.UnauthorizedError, "invalid trusted merchant context", err)
	}

	return merchantID, nil
}

func PtrUUID(id uuid.UUID) *openapi_types.UUID {
	value := id
	return &value
}

func NumericToFloat64(value pgtype.Numeric) float64 {
	number, err := value.Float64Value()
	if err != nil || !number.Valid {
		return 0
	}

	return number.Float64
}

func TextString(value pgtype.Text) string {
	if !value.Valid {
		return ""
	}

	return value.String
}

func MustString(value interface{}) string {
	return fmt.Sprint(value)
}

func PtrString(value string) *string {
	return &value
}
