package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/coverage/store/generated"
)

type DBTX = coveragestore.DBTX

type Postgres struct {
	queries *coveragestore.Queries
}

func New(db DBTX) *Postgres {
	return &Postgres{queries: coveragestore.New(db)}
}

func (store *Postgres) CreateArea(ctx context.Context, arg CreateAreaParams) (Area, error) {
	return store.queries.CreateArea(ctx, coveragestore.CreateAreaParams(arg))
}

func (store *Postgres) GetArea(ctx context.Context, id uuid.UUID) (Area, error) {
	return store.queries.GetArea(ctx, id)
}

func (store *Postgres) CreateZone(ctx context.Context, arg CreateZoneParams) (CreateZoneRow, error) {
	return store.queries.CreateZone(ctx, coveragestore.CreateZoneParams(arg))
}

func (store *Postgres) GetZone(ctx context.Context, id uuid.UUID) (GetZoneRow, error) {
	return store.queries.GetZone(ctx, id)
}

func (store *Postgres) ListAreas(ctx context.Context) ([]Area, error) {
	areas, err := store.queries.ListAreas(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]Area, 0, len(areas))
	for _, area := range areas {
		result = append(result, area)
	}

	return result, nil
}

func (store *Postgres) ListZonesByArea(ctx context.Context, areaID uuid.UUID) ([]ListZonesByAreaRow, error) {
	zones, err := store.queries.ListZonesByArea(ctx, areaID)
	if err != nil {
		return nil, err
	}

	result := make([]ListZonesByAreaRow, 0, len(zones))
	for _, zone := range zones {
		result = append(result, zone)
	}

	return result, nil
}

func (store *Postgres) CreateMerchantServiceZone(ctx context.Context, arg CreateMerchantServiceZoneParams) (MerchantServiceZone, error) {
	return store.queries.CreateMerchantServiceZone(ctx, coveragestore.CreateMerchantServiceZoneParams(arg))
}

func (store *Postgres) ListMerchantServiceZonesByMerchant(ctx context.Context, merchantID uuid.UUID) ([]ListMerchantServiceZonesByMerchantRow, error) {
	rows, err := store.queries.ListMerchantServiceZonesByMerchant(ctx, merchantID)
	if err != nil {
		return nil, err
	}

	result := make([]ListMerchantServiceZonesByMerchantRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, row)
	}

	return result, nil
}
