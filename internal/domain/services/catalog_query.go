package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/pgsqlc"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
)

func (service *CommerceManager) ListProductCategoriesByMerchant(ctx context.Context, merchantID string) ([]pgsqlc.ProductCategory, *domainerr.DomainError) {
	parsedMerchantID, parseErr := parseUUID(merchantID, "merchant id")
	if parseErr != nil {
		return nil, parseErr
	}

	categories, err := service.store.ListProductCategoriesByMerchant(ctx, parsedMerchantID)
	if err != nil {
		return nil, domainerr.MatchPostgresError(err)
	}

	return categories, nil
}

func (service *CommerceManager) ListProductsByMerchant(ctx context.Context, merchantID string) ([]pgsqlc.Product, *domainerr.DomainError) {
	parsedMerchantID, parseErr := parseUUID(merchantID, "merchant id")
	if parseErr != nil {
		return nil, parseErr
	}

	products, err := service.store.ListProductsByMerchant(ctx, parsedMerchantID)
	if err != nil {
		return nil, domainerr.MatchPostgresError(err)
	}

	return products, nil
}

func (service *CommerceManager) GetProductDetail(ctx context.Context, merchantID string, productID string) (pgsqlc.GetProductDetailRow, *domainerr.DomainError) {
	parsedMerchantID, merchantErr := parseUUID(merchantID, "merchant id")
	if merchantErr != nil {
		return pgsqlc.GetProductDetailRow{}, merchantErr
	}
	parsedProductID, productErr := parseUUID(productID, "product id")
	if productErr != nil {
		return pgsqlc.GetProductDetailRow{}, productErr
	}

	product, err := service.store.GetProductDetail(ctx, pgsqlc.GetProductDetailParams{
		MerchantID: parsedMerchantID,
		ID:         parsedProductID,
	})
	if err != nil {
		return pgsqlc.GetProductDetailRow{}, domainerr.MatchPostgresError(err)
	}

	return product, nil
}

func (service *CommerceManager) ListProductAddonsByProduct(ctx context.Context, merchantID string, productID string) ([]pgsqlc.ProductAddon, *domainerr.DomainError) {
	product, productErr := service.GetProductDetail(ctx, merchantID, productID)
	if productErr != nil {
		return nil, productErr
	}

	addons, err := service.store.ListProductAddonsByProduct(ctx, product.ID)
	if err != nil {
		return nil, domainerr.MatchPostgresError(err)
	}

	return addons, nil
}

func (service *CommerceManager) ListInventoryByMerchant(ctx context.Context, viewerActorID uuid.UUID, merchantID string) ([]pgsqlc.ListInventoryByMerchantRow, *domainerr.DomainError) {
	parsedMerchantID, parseErr := parseUUID(merchantID, "merchant id")
	if parseErr != nil {
		return nil, parseErr
	}

	allowed, allowErr := service.canViewMerchant(ctx, viewerActorID, parsedMerchantID)
	if allowErr != nil {
		return nil, domainerr.NewInternalError(allowErr)
	}
	if !allowed {
		return nil, domainerr.NewDomainError(403, domainerr.UnauthorizedError, "admin or merchant role required", fmt.Errorf("admin or merchant role required"))
	}

	items, err := service.store.ListInventoryByMerchant(ctx, parsedMerchantID)
	if err != nil {
		return nil, domainerr.MatchPostgresError(err)
	}

	return items, nil
}
