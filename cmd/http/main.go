package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	actorstore "github.com/horiondreher/go-web-api-boilerplate/internal/actor/store/generated"
	commercestore "github.com/horiondreher/go-web-api-boilerplate/internal/commerce/store"
	merchantapp "github.com/horiondreher/go-web-api-boilerplate/internal/merchant"
	merchantstore "github.com/horiondreher/go-web-api-boilerplate/internal/merchant/store"

	actorapp "github.com/horiondreher/go-web-api-boilerplate/internal/actor/app"
	cartapp "github.com/horiondreher/go-web-api-boilerplate/internal/cart/app"
	catalogapp "github.com/horiondreher/go-web-api-boilerplate/internal/catalog/app"
	coverageapp "github.com/horiondreher/go-web-api-boilerplate/internal/coverage/app"
	orderapp "github.com/horiondreher/go-web-api-boilerplate/internal/order/app"
	reportapp "github.com/horiondreher/go-web-api-boilerplate/internal/report/app"

	httpV1 "github.com/horiondreher/go-web-api-boilerplate/internal/http/v1"
	"github.com/horiondreher/go-web-api-boilerplate/internal/utils"
	pkgdb "github.com/horiondreher/go-web-api-boilerplate/pkg/db"

	"github.com/rs/zerolog/log"
)

var interruptSignals = []os.Signal{
	os.Interrupt,
	syscall.SIGTERM,
	syscall.SIGINT,
}

func main() {
	os.Setenv("TZ", "UTC")
	utils.StartLogger()

	ctx, stop := signal.NotifyContext(context.Background(), interruptSignals...)
	defer stop()

	config := utils.GetConfig()

	if err := pkgdb.RunMigrations(config.MigrationURL, config.DBSource); err != nil {
		log.Fatal().Err(err).Msg("failed to run migrations")
	}

	database, err := pkgdb.New(ctx, config.DBSource)
	if err != nil {
		log.Fatal().Err(err).Msg("error connecting to database")
	}
	defer database.Close()

	actorStore := actorstore.New(database.Pool)
	commerceStore := commercestore.New(database.Pool)
	merchantStore := merchantstore.New(database.Pool)

	actorService := actorapp.NewService(actorStore)
	cartService := cartapp.NewService(database, commerceStore)
	catalogService := catalogapp.NewService(database, commerceStore)
	coverageService := coverageapp.NewService(database, commerceStore)
	merchantService := merchantapp.NewService(database, merchantStore)
	orderService := orderapp.NewService(database, commerceStore)
	reportService := reportapp.NewService(database, commerceStore)

	server, err := httpV1.NewHTTPAdapter(httpV1.AdapterDependencies{
		ActorService:    actorService,
		CartService:     cartService,
		CatalogService:  catalogService,
		CoverageService: coverageService,
		MerchantService: merchantService,
		OrderService:    orderService,
		ReportService:   reportService,
	})
	if err != nil {
		log.Err(err).Msg("error creating server")
		stop()
	}

	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			log.Err(err).Msg("error starting server")
		}
	}()

	<-ctx.Done()
	server.Shutdown()
	log.Info().Msg("server stopped")
}
