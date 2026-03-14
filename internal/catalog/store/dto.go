package store

import (
	"github.com/horiondreher/go-web-api-boilerplate/internal/catalog/store/generated"
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
