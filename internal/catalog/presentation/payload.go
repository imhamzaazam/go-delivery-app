package presentation

import (
	catalogstore "github.com/horiondreher/go-web-api-boilerplate/internal/catalog/store"
	openapi_types "github.com/oapi-codegen/runtime/types"

	api "github.com/horiondreher/go-web-api-boilerplate/api"
	"github.com/horiondreher/go-web-api-boilerplate/internal/core"
)

type createCategoryValidation struct {
	Name string `validate:"required"`
}

type createProductValidation struct {
	CategoryID string  `validate:"required,uuid"`
	Name       string  `validate:"required"`
	BasePrice  float64 `validate:"required,gt=0"`
}

type createProductAddonValidation struct {
	Name  string  `validate:"required"`
	Price float64 `validate:"required,gt=0"`
}

type upsertInventoryValidation struct {
	ProductID string `validate:"required,uuid"`
	BranchID  string `validate:"required,uuid"`
	Quantity  int    `validate:"gte=0"`
}

func categoryResponse(category catalogstore.ProductCategory) api.ProductCategoryResponse {
	description := core.TextString(category.Description)
	createdAt := category.CreatedAt
	return api.ProductCategoryResponse{
		Id:          core.PtrUUID(category.ID),
		MerchantId:  core.PtrUUID(category.MerchantID),
		Name:        &category.Name,
		Description: &description,
		CreatedAt:   &createdAt,
	}
}

func productResponse(product catalogstore.Product) api.ProductResponse {
	description := core.TextString(product.Description)
	imageURL := core.TextString(product.ImageUrl)
	basePrice := core.NumericToFloat64(product.BasePrice)
	createdAt := product.CreatedAt
	updatedAt := product.UpdatedAt
	categoryID := openapi_types.UUID(product.CategoryID)
	return api.ProductResponse{
		Id:             core.PtrUUID(product.ID),
		MerchantId:     core.PtrUUID(product.MerchantID),
		CategoryId:     &categoryID,
		Name:           &product.Name,
		Description:    &description,
		BasePrice:      &basePrice,
		ImageUrl:       &imageURL,
		TrackInventory: &product.TrackInventory,
		IsActive:       &product.IsActive,
		CreatedAt:      &createdAt,
		UpdatedAt:      &updatedAt,
	}
}

func productDetailResponse(product catalogstore.GetProductDetailRow) api.ProductDetailResponse {
	description := core.TextString(product.Description)
	imageURL := core.TextString(product.ImageUrl)
	basePrice := core.NumericToFloat64(product.BasePrice)
	createdAt := product.CreatedAt
	updatedAt := product.UpdatedAt
	categoryName := product.CategoryName
	categoryID := openapi_types.UUID(product.CategoryID)
	return api.ProductDetailResponse{
		Id:             core.PtrUUID(product.ID),
		MerchantId:     core.PtrUUID(product.MerchantID),
		CategoryId:     &categoryID,
		Name:           &product.Name,
		Description:    &description,
		BasePrice:      &basePrice,
		ImageUrl:       &imageURL,
		TrackInventory: &product.TrackInventory,
		IsActive:       &product.IsActive,
		CreatedAt:      &createdAt,
		UpdatedAt:      &updatedAt,
		CategoryName:   &categoryName,
	}
}

func productAddonResponse(addon catalogstore.ProductAddon) api.ProductAddonResponse {
	price := core.NumericToFloat64(addon.Price)
	createdAt := addon.CreatedAt
	return api.ProductAddonResponse{
		Id:        core.PtrUUID(addon.ID),
		ProductId: core.PtrUUID(addon.ProductID),
		Name:      &addon.Name,
		Price:     &price,
		CreatedAt: &createdAt,
	}
}

func inventoryResponse(inventory catalogstore.ProductInventory) api.ProductInventoryResponse {
	quantity := int(inventory.Quantity)
	createdAt := inventory.CreatedAt
	updatedAt := inventory.UpdatedAt
	return api.ProductInventoryResponse{
		Id:        core.PtrUUID(inventory.ID),
		ProductId: core.PtrUUID(inventory.ProductID),
		BranchId:  core.PtrUUID(inventory.BranchID),
		Quantity:  &quantity,
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
	}
}

func inventoryItemResponse(item catalogstore.ListInventoryByMerchantRow) api.InventoryItemResponse {
	quantity := int(item.Quantity)
	return api.InventoryItemResponse{
		ProductId:   core.PtrUUID(item.ProductID),
		ProductName: &item.ProductName,
		Quantity:    &quantity,
	}
}
