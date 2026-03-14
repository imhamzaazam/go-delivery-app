package presentation

import (
	"time"

	"github.com/google/uuid"
	coveragestore "github.com/horiondreher/go-web-api-boilerplate/internal/coverage/store"
	openapi_types "github.com/oapi-codegen/runtime/types"

	api "github.com/horiondreher/go-web-api-boilerplate/api"
	"github.com/horiondreher/go-web-api-boilerplate/internal/core"
	coverage "github.com/horiondreher/go-web-api-boilerplate/internal/coverage"
)

type serviceZoneCoverageCheckValidation struct {
	Latitude  float64 `validate:"required,gte=-90,lte=90"`
	Longitude float64 `validate:"required,gte=-180,lte=180"`
}

type createAreaValidation struct {
	Name string `validate:"required"`
	City string `validate:"required,oneof=Karachi Lahore"`
}

type createZoneValidation struct {
	Name           string `validate:"required"`
	CoordinatesWKT string `validate:"required"`
}

type createMerchantServiceZoneValidation struct {
	ZoneID   string `validate:"required,uuid"`
	BranchID string `validate:"required,uuid"`
}

func areaResponse(area coveragestore.Area) api.AreaResponse {
	city := string(area.City)
	createdAt := area.CreatedAt
	return api.AreaResponse{
		Id:        core.PtrUUID(area.ID),
		Name:      &area.Name,
		City:      &city,
		CreatedAt: &createdAt,
	}
}

func zoneResponse(id uuid.UUID, areaID uuid.UUID, name string, coordinatesWKT string, createdAt time.Time) api.ZoneResponse {
	return api.ZoneResponse{
		Id:             core.PtrUUID(id),
		AreaId:         core.PtrUUID(areaID),
		Name:           &name,
		CoordinatesWkt: core.PtrString(coordinatesWKT),
		CreatedAt:      &createdAt,
	}
}

func zoneReadResponse(zone coveragestore.ListZonesByAreaRow) api.ZoneResponse {
	createdAt := zone.CreatedAt
	return api.ZoneResponse{
		Id:             core.PtrUUID(zone.ID),
		AreaId:         core.PtrUUID(zone.AreaID),
		Name:           &zone.Name,
		CoordinatesWkt: core.PtrString(core.MustString(zone.CoordinatesWkt)),
		CreatedAt:      &createdAt,
	}
}

func merchantServiceZoneResponse(zone coveragestore.MerchantServiceZone) api.MerchantServiceZoneResponse {
	createdAt := zone.CreatedAt
	response := api.MerchantServiceZoneResponse{
		Id:         core.PtrUUID(zone.ID),
		MerchantId: core.PtrUUID(zone.MerchantID),
		ZoneId:     core.PtrUUID(zone.ZoneID),
		CreatedAt:  &createdAt,
	}
	if zone.BranchID != uuid.Nil {
		branchID := openapi_types.UUID(zone.BranchID)
		response.BranchId = &branchID
	}
	return response
}

func merchantServiceZoneReadResponse(row coveragestore.ListMerchantServiceZonesByMerchantRow) api.MerchantServiceZoneResponse {
	createdAt := row.CreatedAt
	item := api.MerchantServiceZoneResponse{
		Id:                 core.PtrUUID(row.ID),
		MerchantId:         core.PtrUUID(row.MerchantID),
		ZoneId:             core.PtrUUID(row.ZoneID),
		CreatedAt:          &createdAt,
		ZoneName:           &row.ZoneName,
		ZoneCoordinatesWkt: core.PtrString(core.MustString(row.ZoneCoordinatesWkt)),
		AreaId:             core.PtrUUID(row.AreaID),
		AreaName:           &row.AreaName,
		AreaCity:           core.PtrString(core.MustString(row.AreaCity)),
		BranchName:         &row.BranchName,
	}
	if row.BranchID != uuid.Nil {
		item.BranchId = core.PtrUUID(row.BranchID)
	}
	return item
}

func serviceZoneCoverageResponse(result coverage.CoverageResult) api.ServiceZoneCoverageCheckResponse {
	response := api.ServiceZoneCoverageCheckResponse{Covered: &result.Covered}
	if result.Covered {
		response.MerchantId = core.PtrUUID(result.MerchantID)
		response.ZoneId = core.PtrUUID(result.ZoneID)
		response.ZoneName = &result.ZoneName
		response.AreaId = core.PtrUUID(result.AreaID)
		response.AreaName = &result.AreaName
		response.AreaCity = &result.AreaCity
		if result.BranchID != uuid.Nil {
			response.BranchId = core.PtrUUID(result.BranchID)
		}
		if result.BranchName != "" {
			response.BranchName = &result.BranchName
		}
	}
	return response
}
