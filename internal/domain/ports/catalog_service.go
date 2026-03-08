package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/pgsqlc"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
)

type CatalogService interface {
	CreateProductByMerchant(ctx context.Context, merchantActorID uuid.UUID, merchantID uuid.UUID, categoryID uuid.UUID, productName string, description string, basePrice float64, imageURL string, trackInventory bool) (pgsqlc.Product, *domainerr.DomainError)
	AddProductAddonByMerchant(ctx context.Context, merchantActorID uuid.UUID, merchantID uuid.UUID, productID uuid.UUID, addonName string, addonPrice float64) (pgsqlc.ProductAddon, *domainerr.DomainError)
	ListProductCategoriesByMerchant(ctx context.Context, merchantID string) ([]pgsqlc.ProductCategory, *domainerr.DomainError)
	ListProductsByMerchant(ctx context.Context, merchantID string) ([]pgsqlc.Product, *domainerr.DomainError)
	GetProductDetail(ctx context.Context, merchantID string, productID string) (pgsqlc.GetProductDetailRow, *domainerr.DomainError)
	ListProductAddonsByProduct(ctx context.Context, merchantID string, productID string) ([]pgsqlc.ProductAddon, *domainerr.DomainError)
	ListInventoryByMerchant(ctx context.Context, viewerActorID uuid.UUID, merchantID string) ([]pgsqlc.ListInventoryByMerchantRow, *domainerr.DomainError)
}
