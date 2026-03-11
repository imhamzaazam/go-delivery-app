package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	catalogstore "github.com/horiondreher/go-web-api-boilerplate/internal/catalog/store"
	"github.com/jackc/pgx/v5/pgtype"

	commerce "github.com/horiondreher/go-web-api-boilerplate/internal/commerce"
	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
	"github.com/horiondreher/go-web-api-boilerplate/internal/utils"
)

func (service *Service) CreateProductCategoryByMerchant(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, name string, description string) (catalogstore.ProductCategory, *domainerr.DomainError) {
	parsedMerchantID, parseErr := utils.ParseUUID(merchantID, "merchant id")
	if parseErr != nil {
		return catalogstore.ProductCategory{}, parseErr
	}

	if _, accessErr := service.requireMerchantManageAccess(ctx, viewerMerchantID, viewerEmail, parsedMerchantID); accessErr != nil {
		return catalogstore.ProductCategory{}, accessErr
	}

	category, err := service.store.CreateProductCategory(ctx, commerce.CreateProductCategoryParams{
		MerchantID:  parsedMerchantID,
		Name:        name,
		Description: optionalText(description),
	})
	if err != nil {
		return catalogstore.ProductCategory{}, domainerr.MatchPostgresError(err)
	}

	return category, nil
}

func (service *Service) CreateProductByMerchant(ctx context.Context, merchantActorID uuid.UUID, merchantID uuid.UUID, categoryID uuid.UUID, productName string, description string, basePrice float64, imageURL string, trackInventory bool) (catalogstore.Product, *domainerr.DomainError) {
	allowed, allowErr := service.canViewMerchant(ctx, merchantActorID, merchantID)
	if allowErr != nil {
		return catalogstore.Product{}, domainerr.NewInternalError(allowErr)
	}
	if !allowed {
		return catalogstore.Product{}, domainerr.NewDomainError(http.StatusForbidden, domainerr.UnauthorizedError, "admin or merchant role required", fmt.Errorf("admin or merchant role required"))
	}

	if _, err := service.store.GetProductCategory(ctx, commerce.GetProductCategoryParams{MerchantID: merchantID, ID: categoryID}); err != nil {
		return catalogstore.Product{}, domainerr.MatchPostgresError(err)
	}

	product, err := service.store.CreateProduct(ctx, commerce.CreateProductParams{
		MerchantID:     merchantID,
		CategoryID:     categoryID,
		Name:           productName,
		Description:    optionalText(description),
		BasePrice:      utils.NumericFromFloat(basePrice),
		ImageUrl:       optionalText(imageURL),
		TrackInventory: trackInventory,
		IsActive:       true,
	})
	if err != nil {
		return catalogstore.Product{}, domainerr.MatchPostgresError(err)
	}

	return product, nil
}

func (service *Service) CreateProductByMerchantHTTP(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, categoryID string, name string, description string, basePrice float64, imageURL string, trackInventory bool) (catalogstore.Product, *domainerr.DomainError) {
	parsedMerchantID, merchantErr := utils.ParseUUID(merchantID, "merchant id")
	if merchantErr != nil {
		return catalogstore.Product{}, merchantErr
	}
	parsedCategoryID, categoryErr := utils.ParseUUID(categoryID, "category id")
	if categoryErr != nil {
		return catalogstore.Product{}, categoryErr
	}

	viewer, accessErr := service.requireMerchantManageAccess(ctx, viewerMerchantID, viewerEmail, parsedMerchantID)
	if accessErr != nil {
		return catalogstore.Product{}, accessErr
	}

	return service.CreateProductByMerchant(ctx, viewer.UID, parsedMerchantID, parsedCategoryID, name, description, basePrice, imageURL, trackInventory)
}

func (service *Service) AddProductAddonByMerchant(ctx context.Context, merchantActorID uuid.UUID, merchantID uuid.UUID, productID uuid.UUID, addonName string, addonPrice float64) (catalogstore.ProductAddon, *domainerr.DomainError) {
	allowed, allowErr := service.canViewMerchant(ctx, merchantActorID, merchantID)
	if allowErr != nil {
		return catalogstore.ProductAddon{}, domainerr.NewInternalError(allowErr)
	}
	if !allowed {
		return catalogstore.ProductAddon{}, domainerr.NewDomainError(http.StatusForbidden, domainerr.UnauthorizedError, "admin or merchant role required", fmt.Errorf("admin or merchant role required"))
	}

	if _, err := service.store.GetProduct(ctx, commerce.GetProductParams{MerchantID: merchantID, ID: productID}); err != nil {
		return catalogstore.ProductAddon{}, domainerr.MatchPostgresError(err)
	}

	addon, err := service.store.CreateProductAddon(ctx, commerce.CreateProductAddonParams{
		ProductID: productID,
		Name:      addonName,
		Price:     utils.NumericFromFloat(addonPrice),
	})
	if err != nil {
		return catalogstore.ProductAddon{}, domainerr.MatchPostgresError(err)
	}

	return addon, nil
}

func (service *Service) AddProductAddonByMerchantHTTP(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, productID string, name string, price float64) (catalogstore.ProductAddon, *domainerr.DomainError) {
	parsedMerchantID, merchantErr := utils.ParseUUID(merchantID, "merchant id")
	if merchantErr != nil {
		return catalogstore.ProductAddon{}, merchantErr
	}
	parsedProductID, productErr := utils.ParseUUID(productID, "product id")
	if productErr != nil {
		return catalogstore.ProductAddon{}, productErr
	}

	viewer, accessErr := service.requireMerchantManageAccess(ctx, viewerMerchantID, viewerEmail, parsedMerchantID)
	if accessErr != nil {
		return catalogstore.ProductAddon{}, accessErr
	}

	return service.AddProductAddonByMerchant(ctx, viewer.UID, parsedMerchantID, parsedProductID, name, price)
}

func (service *Service) UpsertInventoryByMerchant(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, productID string, branchID string, quantity int32) (catalogstore.ProductInventory, *domainerr.DomainError) {
	parsedMerchantID, merchantErr := utils.ParseUUID(merchantID, "merchant id")
	if merchantErr != nil {
		return catalogstore.ProductInventory{}, merchantErr
	}
	parsedProductID, productErr := utils.ParseUUID(productID, "product id")
	if productErr != nil {
		return catalogstore.ProductInventory{}, productErr
	}
	parsedBranchID, branchErr := utils.ParseUUID(branchID, "branch id")
	if branchErr != nil {
		return catalogstore.ProductInventory{}, branchErr
	}

	if _, accessErr := service.requireMerchantManageAccess(ctx, viewerMerchantID, viewerEmail, parsedMerchantID); accessErr != nil {
		return catalogstore.ProductInventory{}, accessErr
	}

	product, err := service.store.GetProduct(ctx, commerce.GetProductParams{MerchantID: parsedMerchantID, ID: parsedProductID})
	if err != nil {
		return catalogstore.ProductInventory{}, domainerr.MatchPostgresError(err)
	}
	if !product.TrackInventory {
		return catalogstore.ProductInventory{}, domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "inventory tracking is disabled for product", fmt.Errorf("inventory tracking is disabled for product"))
	}

	if _, err := service.store.GetBranch(ctx, commerce.GetBranchParams{MerchantID: parsedMerchantID, ID: parsedBranchID}); err != nil {
		return catalogstore.ProductInventory{}, domainerr.MatchPostgresError(err)
	}

	inventory, err := service.store.UpsertProductInventory(ctx, commerce.UpsertProductInventoryParams{
		ProductID: parsedProductID,
		BranchID:  parsedBranchID,
		Quantity:  quantity,
	})
	if err != nil {
		return catalogstore.ProductInventory{}, domainerr.MatchPostgresError(err)
	}

	return inventory, nil
}

func optionalText(value string) pgtype.Text {
	if value == "" {
		return pgtype.Text{}
	}

	return pgtype.Text{String: value, Valid: true}
}
