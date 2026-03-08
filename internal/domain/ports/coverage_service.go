package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/pgsqlc"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
)

type CoverageService interface {
	ListAreas(ctx context.Context) ([]pgsqlc.Area, *domainerr.DomainError)
	ListZonesByArea(ctx context.Context, areaID string) ([]pgsqlc.ListZonesByAreaRow, *domainerr.DomainError)
	ListMerchantServiceZonesByMerchant(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string) ([]pgsqlc.ListMerchantServiceZonesByMerchantRow, *domainerr.DomainError)
}
