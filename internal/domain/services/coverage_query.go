package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/pgsqlc"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/ports"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
)

func (service *CommerceManager) ListAreas(ctx context.Context) ([]pgsqlc.Area, *domainerr.DomainError) {
	areas, err := service.store.ListAreas(ctx)
	if err != nil {
		return nil, domainerr.MatchPostgresError(err)
	}

	return areas, nil
}

func (service *CommerceManager) ListZonesByArea(ctx context.Context, areaID string) ([]pgsqlc.ListZonesByAreaRow, *domainerr.DomainError) {
	parsedAreaID, parseErr := parseUUID(areaID, "area id")
	if parseErr != nil {
		return nil, parseErr
	}

	zones, err := service.store.ListZonesByArea(ctx, parsedAreaID)
	if err != nil {
		return nil, domainerr.MatchPostgresError(err)
	}

	return zones, nil
}

func (service *CommerceManager) ListMerchantServiceZonesByMerchant(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string) ([]pgsqlc.ListMerchantServiceZonesByMerchantRow, *domainerr.DomainError) {
	parsedMerchantID, parseErr := parseUUID(merchantID, "merchant id")
	if parseErr != nil {
		return nil, parseErr
	}

	if _, accessErr := service.requireMerchantViewAccess(ctx, viewerMerchantID, viewerEmail, parsedMerchantID); accessErr != nil {
		return nil, accessErr
	}

	rows, err := service.store.ListMerchantServiceZonesByMerchant(ctx, parsedMerchantID)
	if err != nil {
		return nil, domainerr.MatchPostgresError(err)
	}

	return rows, nil
}

func (service *CommerceManager) CheckMerchantServiceZoneCoverage(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, latitude float64, longitude float64) (ports.ServiceZoneCoverageResult, *domainerr.DomainError) {
	parsedMerchantID, parseErr := parseUUID(merchantID, "merchant id")
	if parseErr != nil {
		return ports.ServiceZoneCoverageResult{}, parseErr
	}

	if _, accessErr := service.requireMerchantViewAccess(ctx, viewerMerchantID, viewerEmail, parsedMerchantID); accessErr != nil {
		return ports.ServiceZoneCoverageResult{}, accessErr
	}

	const coverageQuery = `
SELECT
    msz.merchant_id,
    msz.zone_id,
    msz.branch_id,
    z.name AS zone_name,
    COALESCE(b.name, '') AS branch_name,
    a.id AS area_id,
    a.name AS area_name,
    a.city AS area_city
FROM merchant_service_zones msz
JOIN zones z
  ON z.id = msz.zone_id
JOIN areas a
  ON a.id = z.area_id
LEFT JOIN branches b
  ON b.id = msz.branch_id
WHERE msz.merchant_id = $1
  AND ST_Covers(
        z.coordinates,
        ST_SetSRID(ST_MakePoint($2, $3), 4326)
      )
ORDER BY msz.created_at DESC, msz.id
LIMIT 1;
`

	var result ports.ServiceZoneCoverageResult
	var areaCity pgsqlc.CityType
	err := service.db.QueryRow(ctx, coverageQuery, parsedMerchantID, longitude, latitude).Scan(
		&result.MerchantID,
		&result.ZoneID,
		&result.BranchID,
		&result.ZoneName,
		&result.BranchName,
		&result.AreaID,
		&result.AreaName,
		&areaCity,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ports.ServiceZoneCoverageResult{Covered: false}, nil
		}
		return ports.ServiceZoneCoverageResult{}, domainerr.NewInternalError(err)
	}

	result.Covered = true
	result.AreaCity = string(areaCity)
	return result, nil
}
