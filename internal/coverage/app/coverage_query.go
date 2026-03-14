package app

import (
	"context"
	"errors"

	pgsqlc "github.com/horiondreher/go-web-api-boilerplate/internal/coverage/store"
	"github.com/jackc/pgx/v5"

	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
	coverage "github.com/horiondreher/go-web-api-boilerplate/internal/coverage"
	"github.com/horiondreher/go-web-api-boilerplate/internal/utils"
)

func (service *Service) ListAreas(ctx context.Context) ([]pgsqlc.Area, *domainerr.DomainError) {
	areas, err := service.store.ListAreas(ctx)
	if err != nil {
		return nil, domainerr.MatchPostgresError(err)
	}

	return areas, nil
}

func (service *Service) ListZonesByArea(ctx context.Context, areaID string) ([]pgsqlc.ListZonesByAreaRow, *domainerr.DomainError) {
	parsedAreaID, parseErr := utils.ParseUUID(areaID, "area id")
	if parseErr != nil {
		return nil, parseErr
	}

	zones, err := service.store.ListZonesByArea(ctx, parsedAreaID)
	if err != nil {
		return nil, domainerr.MatchPostgresError(err)
	}

	return zones, nil
}

func (service *Service) ListMerchantServiceZonesByMerchant(ctx context.Context, merchantID string) ([]pgsqlc.ListMerchantServiceZonesByMerchantRow, *domainerr.DomainError) {
	parsedMerchantID, parseErr := utils.ParseUUID(merchantID, "merchant id")
	if parseErr != nil {
		return nil, parseErr
	}

	rows, err := service.store.ListMerchantServiceZonesByMerchant(ctx, parsedMerchantID)
	if err != nil {
		return nil, domainerr.MatchPostgresError(err)
	}

	return rows, nil
}

func (service *Service) CheckMerchantServiceZoneCoverage(ctx context.Context, merchantID string, latitude float64, longitude float64) (coverage.CoverageResult, *domainerr.DomainError) {
	parsedMerchantID, parseErr := utils.ParseUUID(merchantID, "merchant id")
	if parseErr != nil {
		return coverage.CoverageResult{}, parseErr
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

	var result coverage.CoverageResult
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
			return coverage.CoverageResult{Covered: false}, nil
		}
		return coverage.CoverageResult{}, domainerr.NewInternalError(err)
	}

	result.Covered = true
	result.AreaCity = string(areaCity)
	return result, nil
}
