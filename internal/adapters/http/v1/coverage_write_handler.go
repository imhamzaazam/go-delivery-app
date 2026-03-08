package v1

import (
	"net/http"

	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/http/httperr"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/http/httputils"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
)

type createAreaValidation struct {
	Name string `validate:"required"`
	City string `validate:"required,oneof=Karachi Lahore"`
}

type createZoneValidation struct {
	Name           string `validate:"required"`
	CoordinatesWKT string `validate:"required"`
}

func (adapter *HTTPAdapter) CreateArea(w http.ResponseWriter, r *http.Request) {
	adapter.serveAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		requestBody, err := httputils.Decode[CreateAreaRequest](r)
		if err != nil {
			return err
		}

		validationErr := validate.Struct(createAreaValidation{
			Name: requestBody.Name,
			City: string(requestBody.City),
		})
		if validationErr != nil {
			return httperr.MatchValidationError(validationErr)
		}

		viewer, viewerErr := adapter.currentActorProfile(r)
		if viewerErr != nil {
			return viewerErr
		}

		area, createErr := adapter.commerceService.CreateAreaHTTP(r.Context(), viewer.UID, viewer.MerchantID, requestBody.Name, string(requestBody.City))
		if createErr != nil {
			return createErr
		}

		return httputils.Encode(w, r, http.StatusCreated, areaResponse(area))
	})
}

func (adapter *HTTPAdapter) CreateZone(w http.ResponseWriter, r *http.Request, areaID string) {
	adapter.serveAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		requestBody, err := httputils.Decode[CreateZoneRequest](r)
		if err != nil {
			return err
		}

		validationErr := validate.Struct(createZoneValidation{
			Name:           requestBody.Name,
			CoordinatesWKT: requestBody.CoordinatesWkt,
		})
		if validationErr != nil {
			return httperr.MatchValidationError(validationErr)
		}

		viewer, viewerErr := adapter.currentActorProfile(r)
		if viewerErr != nil {
			return viewerErr
		}

		zone, createErr := adapter.commerceService.CreateZoneHTTP(r.Context(), viewer.UID, viewer.MerchantID, areaID, requestBody.Name, requestBody.CoordinatesWkt)
		if createErr != nil {
			return createErr
		}

		return httputils.Encode(w, r, http.StatusCreated, zoneResponse(zone.ID, zone.AreaID, zone.Name, mustString(zone.CoordinatesWkt), zone.CreatedAt))
	})
}
