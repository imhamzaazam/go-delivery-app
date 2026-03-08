package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/pgsqlc"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
)

func (service *CommerceManager) CreateProductByMerchant(ctx context.Context, merchantActorID uuid.UUID, merchantID uuid.UUID, categoryID uuid.UUID, productName string, description string, basePrice float64, imageURL string, trackInventory bool) (pgsqlc.Product, *domainerr.DomainError) {
	allowed, allowErr := service.hasRole(ctx, merchantActorID, pgsqlc.RoleTypeMerchant, merchantID)
	if allowErr != nil {
		return pgsqlc.Product{}, domainerr.NewInternalError(allowErr)
	}
	if !allowed {
		return pgsqlc.Product{}, domainerr.NewDomainError(http.StatusForbidden, domainerr.UnauthorizedError, "merchant role required", fmt.Errorf("merchant role required"))
	}
	if _, categoryErr := service.store.GetProductCategory(ctx, pgsqlc.GetProductCategoryParams{
		MerchantID: merchantID,
		ID:         categoryID,
	}); categoryErr != nil {
		return pgsqlc.Product{}, domainerr.MatchPostgresError(categoryErr)
	}

	createdProduct, createErr := service.store.CreateProduct(ctx, pgsqlc.CreateProductParams{
		MerchantID:     merchantID,
		CategoryID:     categoryID,
		Name:           productName,
		Description:    textValue(description),
		BasePrice:      numericFromFloat(basePrice),
		ImageUrl:       textValue(imageURL),
		TrackInventory: trackInventory,
		IsActive:       true,
	})
	if createErr != nil {
		return pgsqlc.Product{}, domainerr.MatchPostgresError(createErr)
	}

	return createdProduct, nil
}

func (service *CommerceManager) AddProductAddonByMerchant(ctx context.Context, merchantActorID uuid.UUID, merchantID uuid.UUID, productID uuid.UUID, addonName string, addonPrice float64) (pgsqlc.ProductAddon, *domainerr.DomainError) {
	allowed, allowErr := service.hasRole(ctx, merchantActorID, pgsqlc.RoleTypeMerchant, merchantID)
	if allowErr != nil {
		return pgsqlc.ProductAddon{}, domainerr.NewInternalError(allowErr)
	}
	if !allowed {
		return pgsqlc.ProductAddon{}, domainerr.NewDomainError(http.StatusForbidden, domainerr.UnauthorizedError, "merchant role required", fmt.Errorf("merchant role required"))
	}
	product, productErr := service.store.GetProduct(ctx, pgsqlc.GetProductParams{
		MerchantID: merchantID,
		ID:         productID,
	})
	if productErr != nil {
		return pgsqlc.ProductAddon{}, domainerr.MatchPostgresError(productErr)
	}
	if product.MerchantID != merchantID {
		return pgsqlc.ProductAddon{}, domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "product does not belong to merchant", fmt.Errorf("product does not belong to merchant"))
	}

	createdAddon, createErr := service.store.CreateProductAddon(ctx, pgsqlc.CreateProductAddonParams{
		ProductID: productID,
		Name:      addonName,
		Price:     numericFromFloat(addonPrice),
	})
	if createErr != nil {
		return pgsqlc.ProductAddon{}, domainerr.MatchPostgresError(createErr)
	}

	return createdAddon, nil
}
