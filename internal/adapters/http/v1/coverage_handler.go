package v1

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/http/httperr"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/http/httputils"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
)

func (adapter *HTTPAdapter) ListAreas(w http.ResponseWriter, r *http.Request) {
	adapter.serveAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		areas, err := adapter.readService.ListAreas(r.Context())
		if err != nil {
			return err
		}

		response := make([]AreaResponse, 0, len(areas))
		for _, area := range areas {
			city := string(area.City)
			createdAt := area.CreatedAt
			response = append(response, AreaResponse{
				Id:        ptrUUID(area.ID),
				Name:      &area.Name,
				City:      &city,
				CreatedAt: &createdAt,
			})
		}

		return httputils.Encode(w, r, http.StatusOK, response)
	})
}

func (adapter *HTTPAdapter) ListZonesByArea(w http.ResponseWriter, r *http.Request, areaID string) {
	adapter.serveAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		zones, err := adapter.readService.ListZonesByArea(r.Context(), areaID)
		if err != nil {
			return err
		}

		response := make([]ZoneResponse, 0, len(zones))
		for _, zone := range zones {
			createdAt := zone.CreatedAt
			response = append(response, ZoneResponse{
				Id:             ptrUUID(zone.ID),
				AreaId:         ptrUUID(zone.AreaID),
				Name:           &zone.Name,
				CoordinatesWkt: ptrString(mustString(zone.CoordinatesWkt)),
				CreatedAt:      &createdAt,
			})
		}

		return httputils.Encode(w, r, http.StatusOK, response)
	})
}

func (adapter *HTTPAdapter) ListMerchantServiceZonesByMerchant(w http.ResponseWriter, r *http.Request) {
	adapter.serveAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		authUser, authErr := adapter.currentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		rows, err := adapter.readService.ListMerchantServiceZonesByMerchant(r.Context(), authUser.MerchantID, authUser.Email, authUser.MerchantID.String())
		if err != nil {
			return err
		}

		response := make([]MerchantServiceZoneResponse, 0, len(rows))
		for _, row := range rows {
			createdAt := row.CreatedAt
			item := MerchantServiceZoneResponse{
				Id:                 ptrUUID(row.ID),
				MerchantId:         ptrUUID(row.MerchantID),
				ZoneId:             ptrUUID(row.ZoneID),
				CreatedAt:          &createdAt,
				ZoneName:           &row.ZoneName,
				ZoneCoordinatesWkt: ptrString(mustString(row.ZoneCoordinatesWkt)),
				AreaId:             ptrUUID(row.AreaID),
				AreaName:           &row.AreaName,
				AreaCity:           ptrString(mustString(row.AreaCity)),
				BranchName:         &row.BranchName,
			}
			if row.BranchID != uuid.Nil {
				item.BranchId = ptrUUID(row.BranchID)
			}
			response = append(response, item)
		}

		return httputils.Encode(w, r, http.StatusOK, response)
	})
}

type serviceZoneCoverageCheckValidation struct {
	Latitude  float64 `validate:"required,gte=-90,lte=90"`
	Longitude float64 `validate:"required,gte=-180,lte=180"`
}

func (adapter *HTTPAdapter) CheckMerchantServiceZoneCoverage(w http.ResponseWriter, r *http.Request) {
	adapter.serveAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		requestBody, err := httputils.Decode[ServiceZoneCoverageCheckRequest](r)
		if err != nil {
			return err
		}

		validationErr := validate.Struct(serviceZoneCoverageCheckValidation{
			Latitude:  requestBody.Latitude,
			Longitude: requestBody.Longitude,
		})
		if validationErr != nil {
			return httperr.MatchValidationError(validationErr)
		}

		authUser, authErr := adapter.currentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		result, checkErr := adapter.readService.CheckMerchantServiceZoneCoverage(r.Context(), authUser.MerchantID, authUser.Email, authUser.MerchantID.String(), requestBody.Latitude, requestBody.Longitude)
		if checkErr != nil {
			return checkErr
		}

		response := ServiceZoneCoverageCheckResponse{
			Covered: &result.Covered,
		}
		if result.Covered {
			response.MerchantId = ptrUUID(result.MerchantID)
			response.ZoneId = ptrUUID(result.ZoneID)
			response.ZoneName = &result.ZoneName
			response.AreaId = ptrUUID(result.AreaID)
			response.AreaName = &result.AreaName
			response.AreaCity = &result.AreaCity
			if result.BranchID != uuid.Nil {
				response.BranchId = ptrUUID(result.BranchID)
			}
			if result.BranchName != "" {
				response.BranchName = &result.BranchName
			}
		}

		return httputils.Encode(w, r, http.StatusOK, response)
	})
}

func ptrString(value string) *string {
	return &value
}
