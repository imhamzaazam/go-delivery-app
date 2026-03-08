package v1

import (
	"net/http"

	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/http/httperr"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/http/httputils"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
)

type createCategoryValidation struct {
	Name string `validate:"required"`
}

type createProductValidation struct {
	CategoryID string  `validate:"required,uuid"`
	Name       string  `validate:"required"`
	BasePrice  float64 `validate:"required,gt=0"`
}

type createProductAddonValidation struct {
	Name       string  `validate:"required"`
	Price      float64 `validate:"required,gt=0"`
}

func (adapter *HTTPAdapter) CreateProductCategoryByMerchant(w http.ResponseWriter, r *http.Request) {
	adapter.serveAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		requestBody, err := httputils.Decode[CreateCategoryRequest](r)
		if err != nil {
			return err
		}
		validationErr := validate.Struct(createCategoryValidation{Name: requestBody.Name})
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
		category, createErr := adapter.commerceService.CreateProductCategoryByMerchant(r.Context(), authUser.MerchantID, authUser.Email, authUser.MerchantID.String(), requestBody.Name, description)
		if createErr != nil {
			return createErr
		}

		return httputils.Encode(w, r, http.StatusCreated, categoryResponse(category))
	})
}

func (adapter *HTTPAdapter) CreateProductByMerchant(w http.ResponseWriter, r *http.Request) {
	adapter.serveAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		requestBody, err := httputils.Decode[CreateProductRequest](r)
		if err != nil {
			return err
		}
		validationErr := validate.Struct(createProductValidation{
			CategoryID: requestBody.CategoryId.String(),
			Name:       requestBody.Name,
			BasePrice:  requestBody.BasePrice,
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
		imageURL := ""
		if requestBody.ImageUrl != nil {
			imageURL = *requestBody.ImageUrl
		}
		trackInventory := false
		if requestBody.TrackInventory != nil {
			trackInventory = *requestBody.TrackInventory
		}
		product, createErr := adapter.commerceService.CreateProductByMerchantHTTP(r.Context(), authUser.MerchantID, authUser.Email, authUser.MerchantID.String(), requestBody.CategoryId.String(), requestBody.Name, description, requestBody.BasePrice, imageURL, trackInventory)
		if createErr != nil {
			return createErr
		}

		return httputils.Encode(w, r, http.StatusCreated, productResponse(product))
	})
}

func (adapter *HTTPAdapter) AddProductAddonByMerchant(w http.ResponseWriter, r *http.Request, productID string) {
	adapter.serveAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		requestBody, err := httputils.Decode[CreateProductAddonRequest](r)
		if err != nil {
			return err
		}
		validationErr := validate.Struct(createProductAddonValidation{
			Name:       requestBody.Name,
			Price:      requestBody.Price,
		})
		if validationErr != nil {
			return httperr.MatchValidationError(validationErr)
		}

		authUser, authErr := adapter.currentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		addon, createErr := adapter.commerceService.AddProductAddonByMerchantHTTP(r.Context(), authUser.MerchantID, authUser.Email, authUser.MerchantID.String(), productID, requestBody.Name, requestBody.Price)
		if createErr != nil {
			return createErr
		}

		return httputils.Encode(w, r, http.StatusCreated, productAddonResponse(addon))
	})
}
