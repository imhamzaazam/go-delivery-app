package services

import (
	"context"
	"strings"

	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/pgsqlc"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/ports"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MerchantManager struct {
	db    *pgxpool.Pool
	store pgsqlc.Querier
}

func NewMerchantManager(db *pgxpool.Pool, store pgsqlc.Querier) *MerchantManager {
	return &MerchantManager{
		db:    db,
		store: store,
	}
}

func (service *MerchantManager) CreateMerchant(ctx context.Context, newMerchant ports.NewMerchant) (pgsqlc.Merchant, *domainerr.DomainError) {
	category := pgsqlc.MerchantCategory(strings.ToLower(newMerchant.Category))

	merchant, err := service.store.CreateMerchant(ctx, pgsqlc.CreateMerchantParams{
		Name:          newMerchant.Name,
		Ntn:           newMerchant.Ntn,
		Address:       newMerchant.Address,
		Category:      category,
		ContactNumber: newMerchant.ContactNumber,
	})
	if err != nil {
		return pgsqlc.Merchant{}, domainerr.MatchPostgresError(err)
	}

	return merchant, nil
}

func (service *MerchantManager) UpdateMerchant(ctx context.Context, merchantID string, newMerchant ports.NewMerchant) (pgsqlc.Merchant, *domainerr.DomainError) {
	parsedMerchantID, parseErr := parseUUID(merchantID, "merchant id")
	if parseErr != nil {
		return pgsqlc.Merchant{}, parseErr
	}

	merchant, err := service.store.UpdateMerchant(ctx, pgsqlc.UpdateMerchantParams{
		ID:            parsedMerchantID,
		Name:          newMerchant.Name,
		Ntn:           newMerchant.Ntn,
		Address:       newMerchant.Address,
		Category:      pgsqlc.MerchantCategory(strings.ToLower(newMerchant.Category)),
		ContactNumber: newMerchant.ContactNumber,
	})
	if err != nil {
		return pgsqlc.Merchant{}, domainerr.MatchPostgresError(err)
	}

	return merchant, nil
}
