package app

import (
	"context"
	"strings"

	merchantstore "github.com/horiondreher/go-web-api-boilerplate/internal/merchant/store"
	"github.com/horiondreher/go-web-api-boilerplate/internal/utils"

	"github.com/google/uuid"

	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
	merchant "github.com/horiondreher/go-web-api-boilerplate/internal/merchant"
	pkgdb "github.com/horiondreher/go-web-api-boilerplate/pkg/db"
)

type merchantQueryStore interface {
	CreateMerchant(ctx context.Context, arg merchantstore.CreateMerchantParams) (merchantstore.Merchant, error)
	UpdateMerchant(ctx context.Context, arg merchantstore.UpdateMerchantParams) (merchantstore.Merchant, error)
	GetMerchant(ctx context.Context, id uuid.UUID) (merchantstore.Merchant, error)
	ListMerchants(ctx context.Context) ([]merchantstore.Merchant, error)
	ListBranchesByMerchant(ctx context.Context, merchantID uuid.UUID) ([]merchantstore.ListBranchesByMerchantRow, error)
	ListDiscountsByMerchant(ctx context.Context, merchantID uuid.UUID) ([]merchantstore.ListDiscountsByMerchantRow, error)
	ListRolesByMerchant(ctx context.Context, merchantID uuid.UUID) ([]merchantstore.Role, error)
}

type MerchantManager struct {
	db    *pkgdb.DB
	store merchantQueryStore
}

func NewMerchantManager(db *pkgdb.DB, store merchantQueryStore) *MerchantManager {
	return &MerchantManager{
		db:    db,
		store: store,
	}
}

func (service *Service) CreateMerchant(ctx context.Context, newMerchant merchant.NewMerchant) (merchantstore.Merchant, *domainerr.DomainError) {
	category := merchantstore.MerchantCategory(strings.ToLower(newMerchant.Category))

	merchant, err := service.store.CreateMerchant(ctx, merchantstore.CreateMerchantParams{
		Name:          newMerchant.Name,
		Ntn:           newMerchant.Ntn,
		Address:       newMerchant.Address,
		Category:      category,
		ContactNumber: newMerchant.ContactNumber,
	})
	if err != nil {
		return merchantstore.Merchant{}, domainerr.MatchPostgresError(err)
	}

	return merchant, nil
}

func (service *Service) UpdateMerchant(ctx context.Context, merchantID string, newMerchant merchant.NewMerchant) (merchantstore.Merchant, *domainerr.DomainError) {
	parsedMerchantID, parseErr := utils.ParseUUID(merchantID, "merchant id")
	if parseErr != nil {
		return merchantstore.Merchant{}, parseErr
	}

	merchant, err := service.store.UpdateMerchant(ctx, merchantstore.UpdateMerchantParams{
		ID:            parsedMerchantID,
		Name:          newMerchant.Name,
		Ntn:           newMerchant.Ntn,
		Address:       newMerchant.Address,
		Category:      merchantstore.MerchantCategory(strings.ToLower(newMerchant.Category)),
		ContactNumber: newMerchant.ContactNumber,
	})
	if err != nil {
		return merchantstore.Merchant{}, domainerr.MatchPostgresError(err)
	}

	return merchant, nil
}
