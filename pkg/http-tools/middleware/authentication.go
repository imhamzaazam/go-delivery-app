package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
	"github.com/horiondreher/go-web-api-boilerplate/pkg/http-tools/httputils"
	token2 "github.com/horiondreher/go-web-api-boilerplate/pkg/http-tools/token"
)

const (
	bearerAuth = "bearer"
)

func Authentication(tokenMaker *token2.JWTMaker) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")

			requestID := GetRequestID(r.Context())

			if len(auth) == 0 {
				log.Info().Str("id", requestID).Str("error message", "empty authorization header").Msg("request error")
				_ = httputils.Encode(w, r, http.StatusUnauthorized, domainerr.HTTPErrorBody{
					Code:   domainerr.UnauthorizedError,
					Errors: "Empty Authorization Header",
				})
				return
			}

			fields := strings.Fields(auth)

			if len(fields) < 2 {
				log.Info().Str("id", requestID).Str("error message", "invalid authorization header").Msg("request error")
				_ = httputils.Encode(w, r, http.StatusUnauthorized, domainerr.HTTPErrorBody{
					Code:   domainerr.UnauthorizedError,
					Errors: "Invalid Authorization Header",
				})
				return
			}

			authorizationType := strings.ToLower(fields[0])
			if authorizationType != bearerAuth {
				log.Info().Str("id", requestID).Str("error message", "invalid authorization type").Msg("request error")
				_ = httputils.Encode(w, r, http.StatusUnauthorized, domainerr.HTTPErrorBody{
					Code:   domainerr.UnauthorizedError,
					Errors: "Invalid Authorization Type",
				})
				return
			}

			accessToken := fields[1]
			payload, err := tokenMaker.VerifyToken(accessToken)
			if err != nil {
				log.Info().Str("id", requestID).Str("error message", err.Error()).Msg("request error")
				_ = httputils.Encode(w, r, http.StatusUnauthorized, err.HTTPErrorBody)
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, KeyAuthUser, payload)

			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}

func GetAuthUser(httpCtx context.Context) (payload *token2.Payload) {
	if authPayload, ok := httpCtx.Value(KeyAuthUser).(*token2.Payload); ok {
		payload = authPayload
	}

	return
}
