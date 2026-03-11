package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	pgsqlc "github.com/horiondreher/go-web-api-boilerplate/internal/coverage/store"

	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
	"github.com/horiondreher/go-web-api-boilerplate/internal/utils"
)

func (service *Service) CreateAreaHTTP(ctx context.Context, viewerActorID uuid.UUID, viewerMerchantID uuid.UUID, name string, city string) (pgsqlc.Area, *domainerr.DomainError) {
	allowed, allowErr := service.canViewMerchant(ctx, viewerActorID, viewerMerchantID)
	if allowErr != nil {
		return pgsqlc.Area{}, domainerr.NewInternalError(allowErr)
	}
	if !allowed {
		return pgsqlc.Area{}, domainerr.NewDomainError(http.StatusForbidden, domainerr.UnauthorizedError, "admin or merchant role required", fmt.Errorf("admin or merchant role required"))
	}

	createdArea, createErr := service.store.CreateArea(ctx, pgsqlc.CreateAreaParams{
		Name: name,
		City: pgsqlc.CityType(city),
	})
	if createErr != nil {
		return pgsqlc.Area{}, domainerr.MatchPostgresError(createErr)
	}

	return createdArea, nil
}

func (service *Service) CreateZoneHTTP(ctx context.Context, viewerActorID uuid.UUID, viewerMerchantID uuid.UUID, areaID string, name string, coordinatesWKT string) (pgsqlc.CreateZoneRow, *domainerr.DomainError) {
	allowed, allowErr := service.canViewMerchant(ctx, viewerActorID, viewerMerchantID)
	if allowErr != nil {
		return pgsqlc.CreateZoneRow{}, domainerr.NewInternalError(allowErr)
	}
	if !allowed {
		return pgsqlc.CreateZoneRow{}, domainerr.NewDomainError(http.StatusForbidden, domainerr.UnauthorizedError, "admin or merchant role required", fmt.Errorf("admin or merchant role required"))
	}

	parsedAreaID, parseErr := utils.ParseUUID(areaID, "area id")
	if parseErr != nil {
		return pgsqlc.CreateZoneRow{}, parseErr
	}
	if _, getAreaErr := service.store.GetArea(ctx, parsedAreaID); getAreaErr != nil {
		return pgsqlc.CreateZoneRow{}, domainerr.MatchPostgresError(getAreaErr)
	}

	createdZone, createErr := service.store.CreateZone(ctx, pgsqlc.CreateZoneParams{
		AreaID:         parsedAreaID,
		Name:           name,
		StGeomfromtext: coordinatesWKT,
	})
	if createErr != nil {
		return pgsqlc.CreateZoneRow{}, domainerr.MatchPostgresError(createErr)
	}

	return createdZone, nil
}
