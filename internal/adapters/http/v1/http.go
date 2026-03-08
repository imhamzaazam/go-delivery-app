package v1

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/http/httputils"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/http/middleware"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/http/token"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/ports"
	"github.com/horiondreher/go-web-api-boilerplate/internal/utils"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

var validate *validator.Validate

func setupValidator() {
	validate = validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]

		if name == "-" {
			return ""
		}
		return name
	})
}

type HTTPAdapter struct {
	actorService    ports.ActorService
	commerceService ports.CommerceService
	merchantService ports.MerchantService
	readService     ports.ReadService

	config *utils.Config
	router *chi.Mux
	server *http.Server

	tokenMaker *token.JWTMaker
}

type AdapterDependencies struct {
	ActorService    ports.ActorService
	CommerceService ports.CommerceService
	MerchantService ports.MerchantService
	ReadService     ports.ReadService
}

func NewHTTPAdapter(deps AdapterDependencies) (*HTTPAdapter, error) {
	httpAdapter := &HTTPAdapter{
		actorService:    deps.ActorService,
		commerceService: deps.CommerceService,
		merchantService: deps.MerchantService,
		readService:     deps.ReadService,
		config:          utils.GetConfig(),
	}

	setupValidator()

	err := httpAdapter.setupTokenMaker()
	if err != nil {
		log.Err(err).Msg("error setting up server")
		return nil, err
	}

	httpAdapter.setupRouter()
	httpAdapter.setupServer()

	return httpAdapter, nil
}

func (adapter *HTTPAdapter) Start() error {
	log.Info().Str("address", adapter.server.Addr).Msg("starting server")

	_ = chi.Walk(adapter.router, adapter.printRoutes)

	return adapter.server.ListenAndServe()
}

func (adapter *HTTPAdapter) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := adapter.server.Shutdown(ctx); err != nil {
		log.Err(err).Msg("error shutting down server")
	}
}

func (adapter *HTTPAdapter) setupRouter() {
	router := chi.NewRouter()

	router.Use(chiMiddleware.Recoverer)
	router.Use(chiMiddleware.RedirectSlashes)

	router.NotFound(notFoundResponse)
	router.MethodNotAllowed(methodNotAllowedResponse)

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)

	HandlerFromMux(adapter, router)

	adapter.router = router
}

type HandlerWrapper func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError

func (adapter *HTTPAdapter) handlerWrapper(handlerFn HandlerWrapper) http.HandlerFunc {
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

func (adapter *HTTPAdapter) printRoutes(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
	log.Info().Str("method", method).Str("route", route).Msg("route registered")
	return nil
}

func (adapter *HTTPAdapter) setupTokenMaker() error {
	tokenMaker, err := token.NewJWTMaker(adapter.config.TokenSymmetricKey)
	if err != nil {
		return err
	}

	adapter.tokenMaker = tokenMaker

	return nil
}

func (adapter *HTTPAdapter) setupServer() {
	server := &http.Server{
		Addr:    adapter.config.HTTPServerAddress,
		Handler: adapter.router,
	}

	adapter.server = server
}

func (adapter *HTTPAdapter) Router() *chi.Mux {
	return adapter.router
}
