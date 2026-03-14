package presentation

import (
	"net/http"

	api "github.com/horiondreher/go-web-api-boilerplate/api"
	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
	"github.com/horiondreher/go-web-api-boilerplate/pkg/http-tools/httperr"
	"github.com/horiondreher/go-web-api-boilerplate/pkg/http-tools/httputils"
)

func (handler *Handler) CreateProductCategoryByMerchant(w http.ResponseWriter, r *http.Request) {
	handler.shared.ServeAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		requestBody, err := httputils.Decode[api.CreateCategoryRequest](r)
		if err != nil {
			return err
		}
		validationErr := handler.shared.Validate.Struct(createCategoryValidation{Name: requestBody.Name})
		if validationErr != nil {
			return httperr.MatchValidationError(validationErr)
		}

		authUser, authErr := handler.shared.CurrentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		description := ""
		if requestBody.Description != nil {
			description = *requestBody.Description
		}
		category, createErr := handler.shared.CatalogService.CreateProductCategoryByMerchant(r.Context(), authUser.MerchantID, authUser.Email, authUser.MerchantID.String(), requestBody.Name, description)
		if createErr != nil {
			return createErr
		}

		return httputils.Encode(w, r, http.StatusCreated, categoryResponse(category))
	})
}

func (handler *Handler) CreateProductByMerchant(w http.ResponseWriter, r *http.Request) {
	handler.shared.ServeAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		requestBody, err := httputils.Decode[api.CreateProductRequest](r)
		if err != nil {
			return err
		}
		validationErr := handler.shared.Validate.Struct(createProductValidation{
			CategoryID: requestBody.CategoryId.String(),
			Name:       requestBody.Name,
			BasePrice:  requestBody.BasePrice,
		})
		if validationErr != nil {
			return httperr.MatchValidationError(validationErr)
		}

		authUser, authErr := handler.shared.CurrentAuthUser(r)
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
		product, createErr := handler.shared.CatalogService.CreateProductByMerchantHTTP(r.Context(), authUser.MerchantID, authUser.Email, authUser.MerchantID.String(), requestBody.CategoryId.String(), requestBody.Name, description, requestBody.BasePrice, imageURL, trackInventory)
		if createErr != nil {
			return createErr
		}

		return httputils.Encode(w, r, http.StatusCreated, productResponse(product))
	})
}

func (handler *Handler) AddProductAddonByMerchant(w http.ResponseWriter, r *http.Request, productID string) {
	handler.shared.ServeAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		requestBody, err := httputils.Decode[api.CreateProductAddonRequest](r)
		if err != nil {
			return err
		}
		validationErr := handler.shared.Validate.Struct(createProductAddonValidation{
			Name:  requestBody.Name,
			Price: requestBody.Price,
		})
		if validationErr != nil {
			return httperr.MatchValidationError(validationErr)
		}

		authUser, authErr := handler.shared.CurrentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		addon, createErr := handler.shared.CatalogService.AddProductAddonByMerchantHTTP(r.Context(), authUser.MerchantID, authUser.Email, authUser.MerchantID.String(), productID, requestBody.Name, requestBody.Price)
		if createErr != nil {
			return createErr
		}

		return httputils.Encode(w, r, http.StatusCreated, productAddonResponse(addon))
	})
}

func (handler *Handler) UpsertInventoryByMerchant(w http.ResponseWriter, r *http.Request) {
	handler.shared.ServeAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		requestBody, err := httputils.Decode[api.UpsertInventoryRequest](r)
		if err != nil {
			return err
		}
		validationErr := handler.shared.Validate.Struct(upsertInventoryValidation{
			ProductID: requestBody.ProductId.String(),
			BranchID:  requestBody.BranchId.String(),
			Quantity:  requestBody.Quantity,
		})
		if validationErr != nil {
			return httperr.MatchValidationError(validationErr)
		}

		authUser, authErr := handler.shared.CurrentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		inventory, upsertErr := handler.shared.CatalogService.UpsertInventoryByMerchant(r.Context(), authUser.MerchantID, authUser.Email, authUser.MerchantID.String(), requestBody.ProductId.String(), requestBody.BranchId.String(), int32(requestBody.Quantity))
		if upsertErr != nil {
			return upsertErr
		}

		return httputils.Encode(w, r, http.StatusCreated, inventoryResponse(inventory))
	})
}
