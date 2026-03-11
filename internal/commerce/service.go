package commerce

import (
	"context"

	"github.com/google/uuid"

	"github.com/horiondreher/go-web-api-boilerplate/internal/actor"
	"github.com/horiondreher/go-web-api-boilerplate/internal/cart"
	"github.com/horiondreher/go-web-api-boilerplate/internal/catalog"
	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
	"github.com/horiondreher/go-web-api-boilerplate/internal/coverage"
	"github.com/horiondreher/go-web-api-boilerplate/internal/merchant"
	"github.com/horiondreher/go-web-api-boilerplate/internal/order"
	"github.com/horiondreher/go-web-api-boilerplate/internal/report"
)

type Service interface {
	CreateAreaHTTP(ctx context.Context, viewerActorID uuid.UUID, viewerMerchantID uuid.UUID, name string, city string) (coverage.Area, *domainerr.DomainError)
	CreateZoneHTTP(ctx context.Context, viewerActorID uuid.UUID, viewerMerchantID uuid.UUID, areaID string, name string, coordinatesWKT string) (coverage.CreateZoneRow, *domainerr.DomainError)
	CreateActorByMerchant(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, newActor actor.NewActor, role string) (actor.CreateActorRow, *domainerr.DomainError)
	CreateBranchByMerchant(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, name string, address string, contactNumber string, city string, openingTime string, closingTime string) (merchant.Branch, *domainerr.DomainError)
	CreateMerchantServiceZoneByMerchant(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, zoneID string, branchID string) (coverage.MerchantServiceZone, *domainerr.DomainError)
	CreateProductCategoryByMerchant(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, name string, description string) (catalog.ProductCategory, *domainerr.DomainError)
	CreateProductByMerchantHTTP(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, categoryID string, name string, description string, basePrice float64, imageURL string, trackInventory bool) (catalog.Product, *domainerr.DomainError)
	AddProductAddonByMerchantHTTP(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, productID string, name string, price float64) (catalog.ProductAddon, *domainerr.DomainError)
	CreateMerchantDiscountByMerchant(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, discountType string, value float64, description string, productID string, categoryID string) (merchant.MerchantDiscount, *domainerr.DomainError)
	UpsertInventoryByMerchant(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, productID string, branchID string, quantity int32) (catalog.ProductInventory, *domainerr.DomainError)
	CreateCartHTTP(ctx context.Context, merchantID string, branchID string, actorID string, cartID string) (cart.Cart, *domainerr.DomainError)
	AddItemToCartHTTP(ctx context.Context, cartID string, productID string, quantity int32, addonIDs []string, discountID string) (cart.CartItem, *domainerr.DomainError)
	UpdateCartItemQuantityHTTP(ctx context.Context, cartID string, itemID string, quantity int32) (cart.CartItem, *domainerr.DomainError)
	RemoveItemFromCartHTTP(ctx context.Context, cartID string, itemID string) *domainerr.DomainError
	PlaceOrderFromCartHTTP(ctx context.Context, cartID string, paymentType string, deliveryAddress string, customerName string, customerPhone string) (order.Bill, *domainerr.DomainError)
	UpdateOrderStatusHTTP(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, orderID string, status string) (order.Order, *domainerr.DomainError)
}

type ReadService interface {
	ListProductCategoriesByMerchant(ctx context.Context, merchantID string) ([]catalog.ProductCategory, *domainerr.DomainError)
	ListProductsByMerchant(ctx context.Context, merchantID string) ([]catalog.Product, *domainerr.DomainError)
	GetProductDetail(ctx context.Context, merchantID string, productID string) (catalog.GetProductDetailRow, *domainerr.DomainError)
	ListProductAddonsByProduct(ctx context.Context, merchantID string, productID string) ([]catalog.ProductAddon, *domainerr.DomainError)
	ListInventoryByMerchant(ctx context.Context, viewerActorID uuid.UUID, merchantID string) ([]catalog.ListInventoryByMerchantRow, *domainerr.DomainError)
	GetCartDetail(ctx context.Context, cartID string) (cart.Detail, *domainerr.DomainError)
	GetPublicOrderDetail(ctx context.Context, orderID string) (order.Detail, *domainerr.DomainError)
	GetOrderDetail(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, orderID string) (order.Detail, *domainerr.DomainError)
	ListOrdersByMerchant(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string) ([]order.Order, *domainerr.DomainError)
	GetMonthlySalesReport(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, month int, year int) (report.SalesReport, *domainerr.DomainError)
	ListAreas(ctx context.Context) ([]coverage.Area, *domainerr.DomainError)
	ListZonesByArea(ctx context.Context, areaID string) ([]coverage.ListZonesByAreaRow, *domainerr.DomainError)
	ListMerchantServiceZonesByMerchant(ctx context.Context, merchantID string) ([]coverage.ListMerchantServiceZonesByMerchantRow, *domainerr.DomainError)
	CheckMerchantServiceZoneCoverage(ctx context.Context, merchantID string, latitude float64, longitude float64) (coverage.CoverageResult, *domainerr.DomainError)
	GetBranchAvailability(ctx context.Context, merchantID string, branchID string) (merchant.BranchAvailability, *domainerr.DomainError)
}
