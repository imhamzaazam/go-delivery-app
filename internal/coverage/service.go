package coverage

import (
	"context"

	"github.com/google/uuid"
	coveragestore "github.com/horiondreher/go-web-api-boilerplate/internal/coverage/store"

	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
)

type Area = coveragestore.Area
type CreateAreaParams = coveragestore.CreateAreaParams
type CreateZoneParams = coveragestore.CreateZoneParams
type CreateZoneRow = coveragestore.CreateZoneRow
type GetZoneRow = coveragestore.GetZoneRow
type ListMerchantServiceZonesByMerchantRow = coveragestore.ListMerchantServiceZonesByMerchantRow
type ListZonesByAreaRow = coveragestore.ListZonesByAreaRow
type MerchantServiceZone = coveragestore.MerchantServiceZone
type CreateMerchantServiceZoneParams = coveragestore.CreateMerchantServiceZoneParams

type CoverageResult struct {
	Covered    bool
	MerchantID uuid.UUID
	ZoneID     uuid.UUID
	ZoneName   string
	BranchID   uuid.UUID
	BranchName string
	AreaID     uuid.UUID
	AreaName   string
	AreaCity   string
}

type Service interface {
	CreateAreaHTTP(ctx context.Context, viewerActorID uuid.UUID, viewerMerchantID uuid.UUID, name string, city string) (Area, *domainerr.DomainError)
	CreateZoneHTTP(ctx context.Context, viewerActorID uuid.UUID, viewerMerchantID uuid.UUID, areaID string, name string, coordinatesWKT string) (CreateZoneRow, *domainerr.DomainError)
	CreateMerchantServiceZoneByMerchant(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, zoneID string, branchID string) (MerchantServiceZone, *domainerr.DomainError)
	ListAreas(ctx context.Context) ([]Area, *domainerr.DomainError)
	ListZonesByArea(ctx context.Context, areaID string) ([]ListZonesByAreaRow, *domainerr.DomainError)
	ListMerchantServiceZonesByMerchant(ctx context.Context, merchantID string) ([]ListMerchantServiceZonesByMerchantRow, *domainerr.DomainError)
	CheckMerchantServiceZoneCoverage(ctx context.Context, merchantID string, latitude float64, longitude float64) (CoverageResult, *domainerr.DomainError)
}
