package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/pgsqlc"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
)

type CommerceService interface {
	CreateAreaHTTP(ctx context.Context, viewerActorID uuid.UUID, viewerMerchantID uuid.UUID, name string, city string) (pgsqlc.Area, *domainerr.DomainError)
	CreateZoneHTTP(ctx context.Context, viewerActorID uuid.UUID, viewerMerchantID uuid.UUID, areaID string, name string, coordinatesWKT string) (pgsqlc.CreateZoneRow, *domainerr.DomainError)
	CreateActorByMerchant(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, newActor NewActor, role string) (pgsqlc.CreateActorRow, *domainerr.DomainError)
	CreateBranchByMerchant(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, name string, address string, contactNumber string, city string) (pgsqlc.Branch, *domainerr.DomainError)
	CreateMerchantServiceZoneByMerchant(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, zoneID string, branchID string) (pgsqlc.MerchantServiceZone, *domainerr.DomainError)
	CreateProductCategoryByMerchant(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, name string, description string) (pgsqlc.ProductCategory, *domainerr.DomainError)
	CreateProductByMerchantHTTP(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, categoryID string, name string, description string, basePrice float64, imageURL string, trackInventory bool) (pgsqlc.Product, *domainerr.DomainError)
	AddProductAddonByMerchantHTTP(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, productID string, name string, price float64) (pgsqlc.ProductAddon, *domainerr.DomainError)
	CreateMerchantDiscountByMerchant(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, discountType string, value float64, description string) (pgsqlc.MerchantDiscount, *domainerr.DomainError)
	UpsertInventoryByMerchant(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, productID string, branchID string, quantity int32) (pgsqlc.ProductInventory, *domainerr.DomainError)
	CreateCartHTTP(ctx context.Context, merchantID string, branchID string, actorID string, cartID string) (pgsqlc.Cart, *domainerr.DomainError)
	AddItemToCartHTTP(ctx context.Context, cartID string, productID string, quantity int32, addonIDs []string, discountID string) (pgsqlc.CartItem, *domainerr.DomainError)
	PlaceOrderFromCartHTTP(ctx context.Context, cartID string, paymentType string, deliveryAddress string, customerName string, customerPhone string) (OrderBill, *domainerr.DomainError)
	UpdateOrderStatusHTTP(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, orderID string, status string) (pgsqlc.Order, *domainerr.DomainError)
}
