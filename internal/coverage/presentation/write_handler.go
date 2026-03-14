package presentation

import (
	"net/http"

	api "github.com/horiondreher/go-web-api-boilerplate/api"
	"github.com/horiondreher/go-web-api-boilerplate/internal/core"
	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
	"github.com/horiondreher/go-web-api-boilerplate/pkg/http-tools/httperr"
	"github.com/horiondreher/go-web-api-boilerplate/pkg/http-tools/httputils"
)

func (handler *Handler) CreateArea(w http.ResponseWriter, r *http.Request) {
	handler.shared.ServeAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		requestBody, err := httputils.Decode[api.CreateAreaRequest](r)
		if err != nil {
			return err
		}

		validationErr := handler.shared.Validate.Struct(createAreaValidation{
			Name: requestBody.Name,
			City: string(requestBody.City),
		})
		if validationErr != nil {
			return httperr.MatchValidationError(validationErr)
		}

		viewer, viewerErr := handler.shared.CurrentActorProfile(r)
		if viewerErr != nil {
			return viewerErr
		}

		area, createErr := handler.shared.CoverageService.CreateAreaHTTP(r.Context(), viewer.UID, viewer.MerchantID, requestBody.Name, string(requestBody.City))
		if createErr != nil {
			return createErr
		}

		return httputils.Encode(w, r, http.StatusCreated, areaResponse(area))
	})
}

func (handler *Handler) CreateZone(w http.ResponseWriter, r *http.Request, areaID string) {
	handler.shared.ServeAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		requestBody, err := httputils.Decode[api.CreateZoneRequest](r)
		if err != nil {
			return err
		}

		validationErr := handler.shared.Validate.Struct(createZoneValidation{
			Name:           requestBody.Name,
			CoordinatesWKT: requestBody.CoordinatesWkt,
		})
		if validationErr != nil {
			return httperr.MatchValidationError(validationErr)
		}

		viewer, viewerErr := handler.shared.CurrentActorProfile(r)
		if viewerErr != nil {
			return viewerErr
		}

		zone, createErr := handler.shared.CoverageService.CreateZoneHTTP(r.Context(), viewer.UID, viewer.MerchantID, areaID, requestBody.Name, requestBody.CoordinatesWkt)
		if createErr != nil {
			return createErr
		}

		return httputils.Encode(w, r, http.StatusCreated, zoneResponse(zone.ID, zone.AreaID, zone.Name, core.MustString(zone.CoordinatesWkt), zone.CreatedAt))
	})
}

func (handler *Handler) CreateMerchantServiceZoneByMerchant(w http.ResponseWriter, r *http.Request) {
	handler.shared.ServeAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		requestBody, err := httputils.Decode[api.CreateMerchantServiceZoneRequest](r)
		if err != nil {
			return err
		}
		validationErr := handler.shared.Validate.Struct(createMerchantServiceZoneValidation{
			ZoneID:   requestBody.ZoneId.String(),
			BranchID: requestBody.BranchId.String(),
		})
		if validationErr != nil {
			return httperr.MatchValidationError(validationErr)
		}

		authUser, authErr := handler.shared.CurrentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		serviceZone, createErr := handler.shared.CoverageService.CreateMerchantServiceZoneByMerchant(r.Context(), authUser.MerchantID, authUser.Email, authUser.MerchantID.String(), requestBody.ZoneId.String(), requestBody.BranchId.String())
		if createErr != nil {
			return createErr
		}

		return httputils.Encode(w, r, http.StatusCreated, merchantServiceZoneResponse(serviceZone))
	})
}
