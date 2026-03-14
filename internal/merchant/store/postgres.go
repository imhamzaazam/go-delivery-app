package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/merchant/store/generated"
)

type DBTX = merchantstore.DBTX

type Postgres struct {
	queries *merchantstore.Queries
}

func New(db DBTX) *Postgres {
	return &Postgres{queries: merchantstore.New(db)}
}

func (store *Postgres) CreateMerchant(ctx context.Context, arg CreateMerchantParams) (Merchant, error) {
	return store.queries.CreateMerchant(ctx, merchantstore.CreateMerchantParams(arg))
}

func (store *Postgres) UpdateMerchant(ctx context.Context, arg UpdateMerchantParams) (Merchant, error) {
	return store.queries.UpdateMerchant(ctx, merchantstore.UpdateMerchantParams(arg))
}

func (store *Postgres) GetMerchant(ctx context.Context, id uuid.UUID) (Merchant, error) {
	return store.queries.GetMerchant(ctx, id)
}

func (store *Postgres) ListMerchants(ctx context.Context) ([]Merchant, error) {
	items, err := store.queries.ListMerchants(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]Merchant, 0, len(items))
	for _, item := range items {
		result = append(result, item)
	}

	return result, nil
}

func (store *Postgres) CreateRole(ctx context.Context, arg CreateRoleParams) (Role, error) {
	return store.queries.CreateRole(ctx, merchantstore.CreateRoleParams(arg))
}

func (store *Postgres) ListRolesByMerchant(ctx context.Context, merchantID uuid.UUID) ([]Role, error) {
	roles, err := store.queries.ListRolesByMerchant(ctx, merchantID)
	if err != nil {
		return nil, err
	}

	result := make([]Role, 0, len(roles))
	for _, role := range roles {
		result = append(result, role)
	}

	return result, nil
}

func (store *Postgres) CreateBranch(ctx context.Context, arg CreateBranchParams) (Branch, error) {
	return store.queries.CreateBranch(ctx, merchantstore.CreateBranchParams(arg))
}

func (store *Postgres) GetBranch(ctx context.Context, arg GetBranchParams) (Branch, error) {
	return store.queries.GetBranch(ctx, merchantstore.GetBranchParams(arg))
}

func (store *Postgres) ListBranchesByMerchant(ctx context.Context, merchantID uuid.UUID) ([]ListBranchesByMerchantRow, error) {
	branches, err := store.queries.ListBranchesByMerchant(ctx, merchantID)
	if err != nil {
		return nil, err
	}

	result := make([]ListBranchesByMerchantRow, 0, len(branches))
	for _, branch := range branches {
		result = append(result, branch)
	}

	return result, nil
}

func (store *Postgres) CreateMerchantDiscount(ctx context.Context, arg CreateMerchantDiscountParams) (CreateMerchantDiscountRow, error) {
	return store.queries.CreateMerchantDiscount(ctx, merchantstore.CreateMerchantDiscountParams(arg))
}

func (store *Postgres) GetMerchantDiscount(ctx context.Context, arg GetMerchantDiscountParams) (GetMerchantDiscountRow, error) {
	return store.queries.GetMerchantDiscount(ctx, merchantstore.GetMerchantDiscountParams(arg))
}

func (store *Postgres) ListDiscountsByMerchant(ctx context.Context, merchantID uuid.UUID) ([]ListDiscountsByMerchantRow, error) {
	discounts, err := store.queries.ListDiscountsByMerchant(ctx, merchantID)
	if err != nil {
		return nil, err
	}

	result := make([]ListDiscountsByMerchantRow, 0, len(discounts))
	for _, discount := range discounts {
		result = append(result, discount)
	}

	return result, nil
}
