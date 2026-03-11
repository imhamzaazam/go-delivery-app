package app

import (
	"context"

	merchantstore "github.com/horiondreher/go-web-api-boilerplate/internal/merchant/store"
	"github.com/horiondreher/go-web-api-boilerplate/internal/utils"

	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
)

func (service *Service) GetMerchant(ctx context.Context, merchantID string) (merchantstore.Merchant, *domainerr.DomainError) {
	parsedMerchantID, parseErr := utils.ParseUUID(merchantID, "merchant id")
	if parseErr != nil {
		return merchantstore.Merchant{}, parseErr
	}

	merchant, err := service.store.GetMerchant(ctx, parsedMerchantID)
	if err != nil {
		return merchantstore.Merchant{}, domainerr.MatchPostgresError(err)
	}

	return merchant, nil
}

func (service *Service) ListMerchants(ctx context.Context) ([]merchantstore.Merchant, *domainerr.DomainError) {
	merchants, err := service.store.ListMerchants(ctx)
	if err != nil {
		return nil, domainerr.MatchPostgresError(err)
	}

	return merchants, nil
}

func (service *Service) ListBranchesByMerchant(ctx context.Context, merchantID string) ([]merchantstore.Branch, *domainerr.DomainError) {
	parsedMerchantID, parseErr := utils.ParseUUID(merchantID, "merchant id")
	if parseErr != nil {
		return nil, parseErr
	}

	branches, err := service.store.ListBranchesByMerchant(ctx, parsedMerchantID)
	if err != nil {
		return nil, domainerr.MatchPostgresError(err)
	}

	result := make([]merchantstore.Branch, 0, len(branches))
	for _, branch := range branches {
		result = append(result, branchFromListRow(branch))
	}

	return result, nil
}

func (service *Service) ListDiscountsByMerchant(ctx context.Context, merchantID string) ([]merchantstore.MerchantDiscount, *domainerr.DomainError) {
	parsedMerchantID, parseErr := utils.ParseUUID(merchantID, "merchant id")
	if parseErr != nil {
		return nil, parseErr
	}

	discounts, err := service.store.ListDiscountsByMerchant(ctx, parsedMerchantID)
	if err != nil {
		return nil, domainerr.MatchPostgresError(err)
	}

	result := make([]merchantstore.MerchantDiscount, 0, len(discounts))
	for _, discount := range discounts {
		result = append(result, merchantDiscountFromListRow(discount))
	}

	return result, nil
}

func (service *Service) ListRolesByMerchant(ctx context.Context, merchantID string) ([]merchantstore.Role, *domainerr.DomainError) {
	parsedMerchantID, parseErr := utils.ParseUUID(merchantID, "merchant id")
	if parseErr != nil {
		return nil, parseErr
	}

	roles, err := service.store.ListRolesByMerchant(ctx, parsedMerchantID)
	if err != nil {
		return nil, domainerr.MatchPostgresError(err)
	}

	return roles, nil
}
