package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/catalog/store/generated"
)

type DBTX = catalogstore.DBTX

type Postgres struct {
	queries *catalogstore.Queries
}

func New(db DBTX) *Postgres {
	return &Postgres{queries: catalogstore.New(db)}
}

func (store *Postgres) CreateProductCategory(ctx context.Context, arg CreateProductCategoryParams) (ProductCategory, error) {
	return store.queries.CreateProductCategory(ctx, catalogstore.CreateProductCategoryParams(arg))
}

func (store *Postgres) GetProductCategory(ctx context.Context, arg GetProductCategoryParams) (ProductCategory, error) {
	return store.queries.GetProductCategory(ctx, catalogstore.GetProductCategoryParams(arg))
}

func (store *Postgres) CreateProduct(ctx context.Context, arg CreateProductParams) (Product, error) {
	return store.queries.CreateProduct(ctx, catalogstore.CreateProductParams(arg))
}

func (store *Postgres) GetProduct(ctx context.Context, arg GetProductParams) (Product, error) {
	return store.queries.GetProduct(ctx, catalogstore.GetProductParams(arg))
}

func (store *Postgres) GetProductAddon(ctx context.Context, id uuid.UUID) (ProductAddon, error) {
	return store.queries.GetProductAddon(ctx, id)
}

func (store *Postgres) CreateProductAddon(ctx context.Context, arg CreateProductAddonParams) (ProductAddon, error) {
	return store.queries.CreateProductAddon(ctx, catalogstore.CreateProductAddonParams(arg))
}

func (store *Postgres) ListProductCategoriesByMerchant(ctx context.Context, merchantID uuid.UUID) ([]ProductCategory, error) {
	items, err := store.queries.ListProductCategoriesByMerchant(ctx, merchantID)
	if err != nil {
		return nil, err
	}

	result := make([]ProductCategory, 0, len(items))
	for _, item := range items {
		result = append(result, item)
	}

	return result, nil
}

func (store *Postgres) ListProductsByMerchant(ctx context.Context, merchantID uuid.UUID) ([]Product, error) {
	items, err := store.queries.ListProductsByMerchant(ctx, merchantID)
	if err != nil {
		return nil, err
	}

	result := make([]Product, 0, len(items))
	for _, item := range items {
		result = append(result, item)
	}

	return result, nil
}

func (store *Postgres) GetProductDetail(ctx context.Context, arg GetProductDetailParams) (GetProductDetailRow, error) {
	return store.queries.GetProductDetail(ctx, catalogstore.GetProductDetailParams(arg))
}

func (store *Postgres) ListProductAddonsByProduct(ctx context.Context, productID uuid.UUID) ([]ProductAddon, error) {
	items, err := store.queries.ListProductAddonsByProduct(ctx, productID)
	if err != nil {
		return nil, err
	}

	result := make([]ProductAddon, 0, len(items))
	for _, item := range items {
		result = append(result, item)
	}

	return result, nil
}

func (store *Postgres) ListInventoryByMerchant(ctx context.Context, merchantID uuid.UUID) ([]ListInventoryByMerchantRow, error) {
	items, err := store.queries.ListInventoryByMerchant(ctx, merchantID)
	if err != nil {
		return nil, err
	}

	result := make([]ListInventoryByMerchantRow, 0, len(items))
	for _, item := range items {
		result = append(result, item)
	}

	return result, nil
}

func (store *Postgres) GetProductInventory(ctx context.Context, arg GetProductInventoryParams) (ProductInventory, error) {
	return store.queries.GetProductInventory(ctx, catalogstore.GetProductInventoryParams(arg))
}

func (store *Postgres) UpdateProductInventoryQuantity(ctx context.Context, arg UpdateProductInventoryQuantityParams) (ProductInventory, error) {
	return store.queries.UpdateProductInventoryQuantity(ctx, catalogstore.UpdateProductInventoryQuantityParams(arg))
}

func (store *Postgres) UpsertProductInventory(ctx context.Context, arg UpsertProductInventoryParams) (ProductInventory, error) {
	return store.queries.UpsertProductInventory(ctx, catalogstore.UpsertProductInventoryParams(arg))
}
