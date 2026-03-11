package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/actor/store/generated"
	commercestore "github.com/horiondreher/go-web-api-boilerplate/internal/commerce/store"

	actorapp "github.com/horiondreher/go-web-api-boilerplate/internal/actor/app"
	authapp "github.com/horiondreher/go-web-api-boilerplate/internal/auth/app"
	"github.com/horiondreher/go-web-api-boilerplate/internal/cart"
	cartapp "github.com/horiondreher/go-web-api-boilerplate/internal/cart/app"
	catalogapp "github.com/horiondreher/go-web-api-boilerplate/internal/catalog/app"
	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
	coverageapp "github.com/horiondreher/go-web-api-boilerplate/internal/coverage/app"
	merchantapp "github.com/horiondreher/go-web-api-boilerplate/internal/merchant/app"
	orderapp "github.com/horiondreher/go-web-api-boilerplate/internal/order/app"
	"github.com/horiondreher/go-web-api-boilerplate/internal/report"
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

type AuthService = authapp.Service
type CartService = cartapp.Service
type CatalogService = catalogapp.Service
type CoverageService = coverageapp.Service
type MerchantService = merchantapp.Service
type OrderService = orderapp.Service
type ReportService = reportapp.Service

type CommerceWrapper struct {
	*AuthService
	*CartService
	*CatalogService
	*CoverageService
	*MerchantService
	*OrderService
	*ReportService
}

func (wrapper *CommerceWrapper) CreateCartHTTP(ctx context.Context, merchantID string, branchID string, actorID string, cartID string) (cart.Cart, *domainerr.DomainError) {
	parsedMerchantID, merchantErr := utils.ParseUUID(merchantID, "merchant id")
	if merchantErr != nil {
		return cart.Cart{}, merchantErr
	}
	parsedBranchID, branchErr := utils.ParseUUID(branchID, "branch id")
	if branchErr != nil {
		return cart.Cart{}, branchErr
	}
	parsedCartID, cartErr := utils.ParseUUID(cartID, "cart id")
	if cartErr != nil {
		return cart.Cart{}, cartErr
	}

	parsedActorID := uuid.Nil
	if actorID != "" {
		value, actorErr := utils.ParseUUID(actorID, "actor id")
		if actorErr != nil {
			return cart.Cart{}, actorErr
		}
		parsedActorID = value
	}

	return wrapper.CartService.CreateCart(ctx, parsedCartID, parsedMerchantID, parsedBranchID, parsedActorID)
}

func (wrapper *CommerceWrapper) AddItemToCartHTTP(ctx context.Context, cartID string, productID string, quantity int32, addonIDs []string, discountID string) (cart.CartItem, *domainerr.DomainError) {
	parsedCartID, cartErr := utils.ParseUUID(cartID, "cart id")
	if cartErr != nil {
		return cart.CartItem{}, cartErr
	}
	parsedProductID, productErr := utils.ParseUUID(productID, "product id")
	if productErr != nil {
		return cart.CartItem{}, productErr
	}

	parsedAddonIDs := make([]uuid.UUID, 0, len(addonIDs))
	for _, addonID := range addonIDs {
		parsedAddonID, addonErr := utils.ParseUUID(addonID, "addon id")
		if addonErr != nil {
			return cart.CartItem{}, addonErr
		}
		parsedAddonIDs = append(parsedAddonIDs, parsedAddonID)
	}

	parsedDiscountID := uuid.Nil
	if discountID != "" {
		value, discountErr := utils.ParseUUID(discountID, "discount id")
		if discountErr != nil {
			return cart.CartItem{}, discountErr
		}
		parsedDiscountID = value
	}

	return wrapper.CartService.AddItemToCart(ctx, parsedCartID, parsedProductID, quantity, parsedAddonIDs, parsedDiscountID, 0)
}

func (wrapper *CommerceWrapper) UpdateCartItemQuantityHTTP(ctx context.Context, cartID string, itemID string, quantity int32) (cart.CartItem, *domainerr.DomainError) {
	parsedCartID, cartErr := utils.ParseUUID(cartID, "cart id")
	if cartErr != nil {
		return cart.CartItem{}, cartErr
	}
	parsedItemID, itemErr := utils.ParseUUID(itemID, "item id")
	if itemErr != nil {
		return cart.CartItem{}, itemErr
	}

	return wrapper.CartService.UpdateCartItemQuantity(ctx, parsedCartID, parsedItemID, quantity)
}

func (wrapper *CommerceWrapper) GetMonthlySalesReport(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, month int, year int) (report.SalesReport, *domainerr.DomainError) {
	result, err := wrapper.ReportService.GetMonthlySalesReport(ctx, viewerMerchantID, viewerEmail, merchantID, month, year)
	if err != nil {
		return report.SalesReport{}, err
	}

	return report.SalesReport{
		Month:          result.Month,
		Year:           result.Year,
		TotalSales:     result.TotalSales,
		TotalTax:       result.TotalTax,
		TotalDiscount:  result.TotalDiscount,
		ProfitEstimate: result.ProfitEstimate,
	}, nil
}

func (wrapper *CommerceWrapper) RemoveItemFromCartHTTP(ctx context.Context, cartID string, itemID string) *domainerr.DomainError {
	parsedCartID, cartErr := utils.ParseUUID(cartID, "cart id")
	if cartErr != nil {
		return cartErr
	}
	parsedItemID, itemErr := utils.ParseUUID(itemID, "item id")
	if itemErr != nil {
		return itemErr
	}

	return wrapper.CartService.RemoveItemFromCart(ctx, parsedCartID, parsedItemID)
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

	actorService := actorapp.NewService(actorStore)
	merchantService := merchantapp.NewService(database, commerceStore)

	wrapper := &CommerceWrapper{
		AuthService:     authapp.NewService(database, commerceStore),
		CartService:     cartapp.NewService(database, commerceStore),
		CatalogService:  catalogapp.NewService(database, commerceStore),
		CoverageService: coverageapp.NewService(database, commerceStore),
		MerchantService: merchantService,
		OrderService:    orderapp.NewService(database, commerceStore),
		ReportService:   reportapp.NewService(database, commerceStore),
	}

	server, err := httpV1.NewHTTPAdapter(httpV1.AdapterDependencies{
		ActorService:    actorService,
		CommerceService: wrapper,
		MerchantService: merchantService,
		ReadService:     wrapper,
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
