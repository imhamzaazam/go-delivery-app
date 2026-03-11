package presentation

import (
	"net/http"

	api "github.com/horiondreher/go-web-api-boilerplate/api"
	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
	"github.com/horiondreher/go-web-api-boilerplate/pkg/http-tools/httperr"
	"github.com/horiondreher/go-web-api-boilerplate/pkg/http-tools/httputils"
)

func (handler *Handler) ListAreas(w http.ResponseWriter, r *http.Request) {
	handler.shared.ServeAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		areas, err := handler.shared.ReadService.ListAreas(r.Context())
		if err != nil {
			return err
		}

		response := make([]api.AreaResponse, 0, len(areas))
		for _, area := range areas {
			response = append(response, areaResponse(area))
		}

		return httputils.Encode(w, r, http.StatusOK, response)
	})
}

func (handler *Handler) ListZonesByArea(w http.ResponseWriter, r *http.Request, areaID string) {
	handler.shared.ServeAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		zones, err := handler.shared.ReadService.ListZonesByArea(r.Context(), areaID)
		if err != nil {
			return err
		}

		response := make([]api.ZoneResponse, 0, len(zones))
		for _, zone := range zones {
			response = append(response, zoneReadResponse(zone))
		}

		return httputils.Encode(w, r, http.StatusOK, response)
	})
}

func (handler *Handler) ListMerchantServiceZonesByMerchant(w http.ResponseWriter, r *http.Request) {
	handler.shared.Wrap(func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		merchantID, merchantErr := handler.shared.CurrentMerchantID(r)
		if merchantErr != nil {
			return merchantErr
		}

		rows, err := handler.shared.ReadService.ListMerchantServiceZonesByMerchant(r.Context(), merchantID.String())
		if err != nil {
			return err
		}

		response := make([]api.MerchantServiceZoneResponse, 0, len(rows))
		for _, row := range rows {
			response = append(response, merchantServiceZoneReadResponse(row))
		}

		return httputils.Encode(w, r, http.StatusOK, response)
	})(w, r)
}

func (handler *Handler) CheckMerchantServiceZoneCoverage(w http.ResponseWriter, r *http.Request) {
	handler.shared.Wrap(func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		merchantID, merchantErr := handler.shared.CurrentMerchantID(r)
		if merchantErr != nil {
			return merchantErr
		}

		requestBody, err := httputils.Decode[api.ServiceZoneCoverageCheckRequest](r)
		if err != nil {
			return err
		}

		validationErr := handler.shared.Validate.Struct(serviceZoneCoverageCheckValidation{
			Latitude:  requestBody.Latitude,
			Longitude: requestBody.Longitude,
		})
		if validationErr != nil {
			return httperr.MatchValidationError(validationErr)
		}

		result, checkErr := handler.shared.ReadService.CheckMerchantServiceZoneCoverage(r.Context(), merchantID.String(), requestBody.Latitude, requestBody.Longitude)
		if checkErr != nil {
			return checkErr
		}

		return httputils.Encode(w, r, http.StatusOK, serviceZoneCoverageResponse(result))
	})(w, r)
}
