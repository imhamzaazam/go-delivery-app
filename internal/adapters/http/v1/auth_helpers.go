package v1

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/http/middleware"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
)

func (adapter *HTTPAdapter) serveAuthenticated(w http.ResponseWriter, r *http.Request, handler HandlerWrapper) {
	authenticatedHandler := middleware.Authentication(adapter.tokenMaker)(adapter.handlerWrapper(handler))
	authenticatedHandler.ServeHTTP(w, r)
}

func (adapter *HTTPAdapter) currentAuthUser(r *http.Request) (*middlewareAuthUser, *domainerr.DomainError) {
	authUser := middleware.GetAuthUser(r.Context())
	if authUser == nil {
		return nil, domainerr.NewDomainError(http.StatusUnauthorized, domainerr.UnauthorizedError, "unauthorized", errors.New("unauthorized"))
	}

	return &middlewareAuthUser{
		Email:      authUser.Email,
		MerchantID: authUser.MerchantID,
	}, nil
}

type middlewareAuthUser struct {
	Email      string
	MerchantID uuid.UUID
}
