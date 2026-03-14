package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	pgsqlc "github.com/horiondreher/go-web-api-boilerplate/internal/catalog/store"

	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
	"github.com/horiondreher/go-web-api-boilerplate/internal/utils"
)

func (service *Service) ListProductCategoriesByMerchant(ctx context.Context, merchantID string) ([]pgsqlc.ProductCategory, *domainerr.DomainError) {
	parsedMerchantID, parseErr := utils.ParseUUID(merchantID, "merchant id")
	if parseErr != nil {
		return nil, parseErr
	}

	categories, err := service.store.ListProductCategoriesByMerchant(ctx, parsedMerchantID)
	if err != nil {
		return nil, domainerr.MatchPostgresError(err)
	}

	return categories, nil
}

func (service *Service) ListProductsByMerchant(ctx context.Context, merchantID string) ([]pgsqlc.Product, *domainerr.DomainError) {
	parsedMerchantID, parseErr := utils.ParseUUID(merchantID, "merchant id")
	if parseErr != nil {
		return nil, parseErr
	}

	products, err := service.store.ListProductsByMerchant(ctx, parsedMerchantID)
	if err != nil {
		return nil, domainerr.MatchPostgresError(err)
	}

	return products, nil
}

func (service *Service) GetProductDetail(ctx context.Context, merchantID string, productID string) (pgsqlc.GetProductDetailRow, *domainerr.DomainError) {
	parsedMerchantID, merchantErr := utils.ParseUUID(merchantID, "merchant id")
	if merchantErr != nil {
		return pgsqlc.GetProductDetailRow{}, merchantErr
	}
	parsedProductID, productErr := utils.ParseUUID(productID, "product id")
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

func (service *Service) ListProductAddonsByProduct(ctx context.Context, merchantID string, productID string) ([]pgsqlc.ProductAddon, *domainerr.DomainError) {
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

func (service *Service) ListInventoryByMerchant(ctx context.Context, viewerActorID uuid.UUID, merchantID string) ([]pgsqlc.ListInventoryByMerchantRow, *domainerr.DomainError) {
	parsedMerchantID, parseErr := utils.ParseUUID(merchantID, "merchant id")
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
