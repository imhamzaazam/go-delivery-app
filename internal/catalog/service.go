package catalog

import (
	"context"

	"github.com/google/uuid"
	catalogstore "github.com/horiondreher/go-web-api-boilerplate/internal/catalog/store"

	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
)

type Product = catalogstore.Product
type ProductAddon = catalogstore.ProductAddon
type ProductCategory = catalogstore.ProductCategory
type ProductInventory = catalogstore.ProductInventory
type CreateProductParams = catalogstore.CreateProductParams
type CreateProductAddonParams = catalogstore.CreateProductAddonParams
type CreateProductCategoryParams = catalogstore.CreateProductCategoryParams
type GetProductParams = catalogstore.GetProductParams
type GetProductCategoryParams = catalogstore.GetProductCategoryParams
type GetProductDetailParams = catalogstore.GetProductDetailParams
type GetProductDetailRow = catalogstore.GetProductDetailRow
type GetProductInventoryParams = catalogstore.GetProductInventoryParams
type ListInventoryByMerchantRow = catalogstore.ListInventoryByMerchantRow
type UpdateProductInventoryQuantityParams = catalogstore.UpdateProductInventoryQuantityParams
type UpsertProductInventoryParams = catalogstore.UpsertProductInventoryParams

type Service interface {
	CreateProductByMerchant(ctx context.Context, merchantActorID uuid.UUID, merchantID uuid.UUID, categoryID uuid.UUID, productName string, description string, basePrice float64, imageURL string, trackInventory bool) (Product, *domainerr.DomainError)
	AddProductAddonByMerchant(ctx context.Context, merchantActorID uuid.UUID, merchantID uuid.UUID, productID uuid.UUID, addonName string, addonPrice float64) (ProductAddon, *domainerr.DomainError)
	ListProductCategoriesByMerchant(ctx context.Context, merchantID string) ([]ProductCategory, *domainerr.DomainError)
	ListProductsByMerchant(ctx context.Context, merchantID string) ([]Product, *domainerr.DomainError)
	GetProductDetail(ctx context.Context, merchantID string, productID string) (GetProductDetailRow, *domainerr.DomainError)
	ListProductAddonsByProduct(ctx context.Context, merchantID string, productID string) ([]ProductAddon, *domainerr.DomainError)
	ListInventoryByMerchant(ctx context.Context, viewerActorID uuid.UUID, merchantID string) ([]ListInventoryByMerchantRow, *domainerr.DomainError)
}
