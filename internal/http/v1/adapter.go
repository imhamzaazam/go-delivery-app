package v1

import (
	"context"
	"net/http"
	"reflect"
	"strings"
	"time"

	actorservice "github.com/horiondreher/go-web-api-boilerplate/internal/actor"
	actor "github.com/horiondreher/go-web-api-boilerplate/internal/actor/presentation"
	auth "github.com/horiondreher/go-web-api-boilerplate/internal/auth/presentation"
	cart "github.com/horiondreher/go-web-api-boilerplate/internal/cart/presentation"
	catalog "github.com/horiondreher/go-web-api-boilerplate/internal/catalog/presentation"
	commerce "github.com/horiondreher/go-web-api-boilerplate/internal/commerce"
	coverage "github.com/horiondreher/go-web-api-boilerplate/internal/coverage/presentation"
	merchantservice "github.com/horiondreher/go-web-api-boilerplate/internal/merchant"
	merchant "github.com/horiondreher/go-web-api-boilerplate/internal/merchant/presentation"
	order "github.com/horiondreher/go-web-api-boilerplate/internal/order/presentation"
	report "github.com/horiondreher/go-web-api-boilerplate/internal/report/presentation"

	"github.com/google/uuid"

	api "github.com/horiondreher/go-web-api-boilerplate/api"
	"github.com/horiondreher/go-web-api-boilerplate/internal/core"
	"github.com/horiondreher/go-web-api-boilerplate/internal/utils"
	middleware2 "github.com/horiondreher/go-web-api-boilerplate/pkg/http-tools/middleware"
	"github.com/horiondreher/go-web-api-boilerplate/pkg/http-tools/token"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

type HTTPAdapter struct {
	actorService    actorservice.Service
	commerceService commerce.Service
	merchantService merchantservice.Service
	readService     commerce.ReadService

	config *utils.Config
	router *chi.Mux
	server *http.Server

	tokenMaker *token.JWTMaker
	shared     *core.Shared

	actorHandler    *actor.Handler
	authHandler     *auth.Handler
	cartHandler     *cart.Handler
	catalogHandler  *catalog.Handler
	coverageHandler *coverage.Handler
	merchantHandler *merchant.Handler
	orderHandler    *order.Handler
	reportHandler   *report.Handler
}

type AdapterDependencies struct {
	ActorService    actorservice.Service
	CommerceService commerce.Service
	MerchantService merchantservice.Service
	ReadService     commerce.ReadService
}

func NewHTTPAdapter(deps AdapterDependencies) (*HTTPAdapter, error) {
	adapter := &HTTPAdapter{
		actorService:    deps.ActorService,
		commerceService: deps.CommerceService,
		merchantService: deps.MerchantService,
		readService:     deps.ReadService,
		config:          utils.GetConfig(),
	}

	if err := adapter.setupTokenMaker(); err != nil {
		log.Err(err).Msg("error setting up server")
		return nil, err
	}

	adapter.shared = &core.Shared{
		ActorService:    adapter.actorService,
		CommerceService: adapter.commerceService,
		MerchantService: adapter.merchantService,
		ReadService:     adapter.readService,
		Config:          adapter.config,
		TokenMaker:      adapter.tokenMaker,
		Validate:        newValidator(),
	}
	adapter.actorHandler = actor.New(adapter.shared)
	adapter.authHandler = auth.New(adapter.shared)
	adapter.cartHandler = cart.New(adapter.shared)
	adapter.catalogHandler = catalog.New(adapter.shared)
	adapter.coverageHandler = coverage.New(adapter.shared)
	adapter.merchantHandler = merchant.New(adapter.shared)
	adapter.orderHandler = order.New(adapter.shared)
	adapter.reportHandler = report.New(adapter.shared)

	adapter.setupRouter()
	adapter.setupServer()

	return adapter, nil
}

func newValidator() *validator.Validate {
	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	return validate
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

func (adapter *HTTPAdapter) Router() *chi.Mux {
	return adapter.router
}

func (adapter *HTTPAdapter) CreateActorAccessToken(email string, merchantID uuid.UUID) (string, error) {
	tokenValue, _, err := adapter.tokenMaker.CreateToken(email, "actor", merchantID, adapter.config.AccessTokenDuration)
	if err != nil {
		return "", err
	}

	return tokenValue, nil
}

func (adapter *HTTPAdapter) setupTokenMaker() error {
	tokenMaker, err := token.NewJWTMaker(adapter.config.TokenSymmetricKey)
	if err != nil {
		return err
	}

	adapter.tokenMaker = tokenMaker
	return nil
}

func (adapter *HTTPAdapter) setupRouter() {
	router := chi.NewRouter()

	router.Use(chiMiddleware.Recoverer)
	router.Use(chiMiddleware.RedirectSlashes)
	router.NotFound(notFoundHandler)
	router.MethodNotAllowed(methodNotAllowedHandler)
	router.Use(middleware2.RequestID)
	router.Use(middleware2.Logger)

	api.HandlerFromMux(adapter, router)
	adapter.router = router
}

func (adapter *HTTPAdapter) setupServer() {
	adapter.server = &http.Server{
		Addr:    adapter.config.HTTPServerAddress,
		Handler: adapter.router,
	}
}

func (adapter *HTTPAdapter) printRoutes(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
	log.Info().Str("method", method).Str("route", route).Msg("route registered")
	return nil
}
