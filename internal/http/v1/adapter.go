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
	cartservice "github.com/horiondreher/go-web-api-boilerplate/internal/cart"
	cart "github.com/horiondreher/go-web-api-boilerplate/internal/cart/presentation"
	catalogservice "github.com/horiondreher/go-web-api-boilerplate/internal/catalog"
	catalog "github.com/horiondreher/go-web-api-boilerplate/internal/catalog/presentation"
	coverageservice "github.com/horiondreher/go-web-api-boilerplate/internal/coverage"
	coverage "github.com/horiondreher/go-web-api-boilerplate/internal/coverage/presentation"
	merchant "github.com/horiondreher/go-web-api-boilerplate/internal/merchant"
	orderservice "github.com/horiondreher/go-web-api-boilerplate/internal/order"
	order "github.com/horiondreher/go-web-api-boilerplate/internal/order/presentation"
	reportservice "github.com/horiondreher/go-web-api-boilerplate/internal/report"
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
	cartService     cartservice.Service
	catalogService  catalogservice.Service
	coverageService coverageservice.Service
	merchantService *merchant.MerchantService
	orderService    orderservice.Service
	reportService   reportservice.Service

	config *utils.Config
	router *chi.Mux
	server *http.Server

	tokenMaker *token.JWTMaker
	shared     *core.Shared
	openapi    *openAPIServer
}

type actorAPIServer = actor.Handler
type authAPIServer = auth.Handler
type cartAPIServer = cart.Handler
type catalogAPIServer = catalog.Handler
type coverageAPIServer = coverage.Handler
type merchantAPIServer = merchant.Handler
type orderAPIServer = order.Handler
type reportAPIServer = report.Handler

type openAPIServer struct {
	*actorAPIServer
	*authAPIServer
	*cartAPIServer
	*catalogAPIServer
	*coverageAPIServer
	*merchantAPIServer
	*orderAPIServer
	*reportAPIServer
}

var _ api.ServerInterface = (*openAPIServer)(nil)

type AdapterDependencies struct {
	ActorService    actorservice.Service
	CartService     cartservice.Service
	CatalogService  catalogservice.Service
	CoverageService coverageservice.Service
	MerchantService *merchant.MerchantService
	OrderService    orderservice.Service
	ReportService   reportservice.Service
}

func NewHTTPAdapter(deps AdapterDependencies) (*HTTPAdapter, error) {
	adapter := &HTTPAdapter{
		actorService:    deps.ActorService,
		cartService:     deps.CartService,
		catalogService:  deps.CatalogService,
		coverageService: deps.CoverageService,
		merchantService: deps.MerchantService,
		orderService:    deps.OrderService,
		reportService:   deps.ReportService,
		config:          utils.GetConfig(),
	}

	if err := adapter.setupTokenMaker(); err != nil {
		log.Err(err).Msg("error setting up server")
		return nil, err
	}

	adapter.shared = &core.Shared{
		ActorService:    adapter.actorService,
		CartService:     adapter.cartService,
		CatalogService:  adapter.catalogService,
		CoverageService: adapter.coverageService,
		MerchantService: adapter.merchantService,
		OrderService:    adapter.orderService,
		ReportService:   adapter.reportService,
		Config:          adapter.config,
		TokenMaker:      adapter.tokenMaker,
		Validate:        newValidator(),
	}
	adapter.openapi = &openAPIServer{
		actorAPIServer:    actor.New(adapter.shared),
		authAPIServer:     auth.New(adapter.shared),
		cartAPIServer:     cart.New(adapter.shared),
		catalogAPIServer:  catalog.New(adapter.shared),
		coverageAPIServer: coverage.New(adapter.shared),
		merchantAPIServer: merchant.New(adapter.merchantService, adapter.config, adapter.tokenMaker, adapter.shared.Validate),
		orderAPIServer:    order.New(adapter.shared),
		reportAPIServer:   report.New(adapter.shared),
	}

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

	api.HandlerFromMux(adapter.openapi, router)
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
