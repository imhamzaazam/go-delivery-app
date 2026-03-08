package services

import (
	"context"

	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/pgsqlc"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
)

func (service *MerchantManager) GetMerchant(ctx context.Context, merchantID string) (pgsqlc.Merchant, *domainerr.DomainError) {
	parsedMerchantID, parseErr := parseUUID(merchantID, "merchant id")
	if parseErr != nil {
		return pgsqlc.Merchant{}, parseErr
	}

	merchant, err := service.store.GetMerchant(ctx, parsedMerchantID)
	if err != nil {
		return pgsqlc.Merchant{}, domainerr.MatchPostgresError(err)
	}

	return merchant, nil
}

func (service *MerchantManager) ListMerchants(ctx context.Context) ([]pgsqlc.Merchant, *domainerr.DomainError) {
	merchants, err := service.store.ListMerchants(ctx)
	if err != nil {
		return nil, domainerr.MatchPostgresError(err)
	}

	return merchants, nil
}

func (service *MerchantManager) ListBranchesByMerchant(ctx context.Context, merchantID string) ([]pgsqlc.Branch, *domainerr.DomainError) {
	parsedMerchantID, parseErr := parseUUID(merchantID, "merchant id")
	if parseErr != nil {
		return nil, parseErr
	}

	branches, err := service.store.ListBranchesByMerchant(ctx, parsedMerchantID)
	if err != nil {
		return nil, domainerr.MatchPostgresError(err)
	}

	return branches, nil
}

func (service *MerchantManager) ListDiscountsByMerchant(ctx context.Context, merchantID string) ([]pgsqlc.MerchantDiscount, *domainerr.DomainError) {
	parsedMerchantID, parseErr := parseUUID(merchantID, "merchant id")
	if parseErr != nil {
		return nil, parseErr
	}

	discounts, err := service.store.ListDiscountsByMerchant(ctx, parsedMerchantID)
	if err != nil {
		return nil, domainerr.MatchPostgresError(err)
	}

	return discounts, nil
}

func (service *MerchantManager) ListRolesByMerchant(ctx context.Context, merchantID string) ([]pgsqlc.Role, *domainerr.DomainError) {
	parsedMerchantID, parseErr := parseUUID(merchantID, "merchant id")
	if parseErr != nil {
		return nil, parseErr
	}

	roles, err := service.store.ListRolesByMerchant(ctx, parsedMerchantID)
	if err != nil {
		return nil, domainerr.MatchPostgresError(err)
	}

	return roles, nil
}
