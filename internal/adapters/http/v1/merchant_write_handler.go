package v1

import (
	"net/http"

	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/http/httperr"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/http/httputils"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
)

type createBranchValidation struct {
	Name          string `validate:"required"`
	Address       string `validate:"required"`
	ContactNumber string `validate:"required"`
	City          string `validate:"required,oneof=Karachi Lahore"`
}

type createDiscountValidation struct {
	Type  string  `validate:"required,oneof=flat percentage"`
	Value float64 `validate:"required,gt=0"`
}

type createMerchantServiceZoneValidation struct {
	ZoneID   string `validate:"required,uuid"`
	BranchID string `validate:"required,uuid"`
}

type upsertInventoryValidation struct {
	ProductID string `validate:"required,uuid"`
	BranchID  string `validate:"required,uuid"`
	Quantity  int    `validate:"gte=0"`
}

func (adapter *HTTPAdapter) CreateBranchByMerchant(w http.ResponseWriter, r *http.Request) {
	adapter.serveAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		requestBody, err := httputils.Decode[CreateBranchRequest](r)
		if err != nil {
			return err
		}
		validationErr := validate.Struct(createBranchValidation{
			Name:          requestBody.Name,
			Address:       requestBody.Address,
			ContactNumber: requestBody.ContactNumber,
			City:          string(requestBody.City),
		})
		if validationErr != nil {
			return httperr.MatchValidationError(validationErr)
		}

		authUser, authErr := adapter.currentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		branch, createErr := adapter.commerceService.CreateBranchByMerchant(r.Context(), authUser.MerchantID, authUser.Email, authUser.MerchantID.String(), requestBody.Name, requestBody.Address, requestBody.ContactNumber, string(requestBody.City))
		if createErr != nil {
			return createErr
		}

		return httputils.Encode(w, r, http.StatusCreated, branchResponse(branch))
	})
}

func (adapter *HTTPAdapter) CreateDiscountByMerchant(w http.ResponseWriter, r *http.Request) {
	adapter.serveAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		requestBody, err := httputils.Decode[CreateDiscountRequest](r)
		if err != nil {
			return err
		}
		validationErr := validate.Struct(createDiscountValidation{
			Type:  string(requestBody.Type),
			Value: requestBody.Value,
		})
		if validationErr != nil {
			return httperr.MatchValidationError(validationErr)
		}

		authUser, authErr := adapter.currentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		description := ""
		if requestBody.Description != nil {
			description = *requestBody.Description
		}
		discount, createErr := adapter.commerceService.CreateMerchantDiscountByMerchant(r.Context(), authUser.MerchantID, authUser.Email, authUser.MerchantID.String(), string(requestBody.Type), requestBody.Value, description)
		if createErr != nil {
			return createErr
		}

		return httputils.Encode(w, r, http.StatusCreated, discountResponse(discount))
	})
}

func (adapter *HTTPAdapter) CreateMerchantServiceZoneByMerchant(w http.ResponseWriter, r *http.Request) {
	adapter.serveAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		requestBody, err := httputils.Decode[CreateMerchantServiceZoneRequest](r)
		if err != nil {
			return err
		}
		validationErr := validate.Struct(createMerchantServiceZoneValidation{
			ZoneID:   requestBody.ZoneId.String(),
			BranchID: requestBody.BranchId.String(),
		})
		if validationErr != nil {
			return httperr.MatchValidationError(validationErr)
		}

		authUser, authErr := adapter.currentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		serviceZone, createErr := adapter.commerceService.CreateMerchantServiceZoneByMerchant(r.Context(), authUser.MerchantID, authUser.Email, authUser.MerchantID.String(), requestBody.ZoneId.String(), requestBody.BranchId.String())
		if createErr != nil {
			return createErr
		}

		return httputils.Encode(w, r, http.StatusCreated, merchantServiceZoneResponse(serviceZone))
	})
}

func (adapter *HTTPAdapter) UpsertInventoryByMerchant(w http.ResponseWriter, r *http.Request) {
	adapter.serveAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		requestBody, err := httputils.Decode[UpsertInventoryRequest](r)
		if err != nil {
			return err
		}
		validationErr := validate.Struct(upsertInventoryValidation{
			ProductID: requestBody.ProductId.String(),
			BranchID:  requestBody.BranchId.String(),
			Quantity:  requestBody.Quantity,
		})
		if validationErr != nil {
			return httperr.MatchValidationError(validationErr)
		}

		authUser, authErr := adapter.currentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		inventory, upsertErr := adapter.commerceService.UpsertInventoryByMerchant(r.Context(), authUser.MerchantID, authUser.Email, authUser.MerchantID.String(), requestBody.ProductId.String(), requestBody.BranchId.String(), int32(requestBody.Quantity))
		if upsertErr != nil {
			return upsertErr
		}

		return httputils.Encode(w, r, http.StatusCreated, inventoryResponse(inventory))
	})
}
