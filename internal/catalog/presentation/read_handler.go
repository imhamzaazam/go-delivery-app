package presentation

import (
	"net/http"

	api "github.com/horiondreher/go-web-api-boilerplate/api"
	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
	"github.com/horiondreher/go-web-api-boilerplate/pkg/http-tools/httputils"
)

func (handler *Handler) ListProductCategoriesByMerchant(w http.ResponseWriter, r *http.Request) {
	handler.shared.Wrap(func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		merchantID, merchantErr := handler.shared.CurrentMerchantID(r)
		if merchantErr != nil {
			return merchantErr
		}

		categories, err := handler.shared.ReadService.ListProductCategoriesByMerchant(r.Context(), merchantID.String())
		if err != nil {
			return err
		}

		response := make([]api.ProductCategoryResponse, 0, len(categories))
		for _, category := range categories {
			response = append(response, categoryResponse(category))
		}

		return httputils.Encode(w, r, http.StatusOK, response)
	})(w, r)
}

func (handler *Handler) ListProductsByMerchant(w http.ResponseWriter, r *http.Request) {
	handler.shared.Wrap(func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		merchantID, merchantErr := handler.shared.CurrentMerchantID(r)
		if merchantErr != nil {
			return merchantErr
		}

		products, err := handler.shared.ReadService.ListProductsByMerchant(r.Context(), merchantID.String())
		if err != nil {
			return err
		}

		response := make([]api.ProductResponse, 0, len(products))
		for _, product := range products {
			response = append(response, productResponse(product))
		}

		return httputils.Encode(w, r, http.StatusOK, response)
	})(w, r)
}

func (handler *Handler) GetProductDetail(w http.ResponseWriter, r *http.Request, productID string) {
	handler.shared.Wrap(func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		merchantID, merchantErr := handler.shared.CurrentMerchantID(r)
		if merchantErr != nil {
			return merchantErr
		}

		product, err := handler.shared.ReadService.GetProductDetail(r.Context(), merchantID.String(), productID)
		if err != nil {
			return err
		}

		return httputils.Encode(w, r, http.StatusOK, productDetailResponse(product))
	})(w, r)
}

func (handler *Handler) ListProductAddonsByProduct(w http.ResponseWriter, r *http.Request, productID string) {
	handler.shared.Wrap(func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		merchantID, merchantErr := handler.shared.CurrentMerchantID(r)
		if merchantErr != nil {
			return merchantErr
		}

		addons, err := handler.shared.ReadService.ListProductAddonsByProduct(r.Context(), merchantID.String(), productID)
		if err != nil {
			return err
		}

		response := make([]api.ProductAddonResponse, 0, len(addons))
		for _, addon := range addons {
			response = append(response, productAddonResponse(addon))
		}

		return httputils.Encode(w, r, http.StatusOK, response)
	})(w, r)
}

func (handler *Handler) ListInventoryByMerchant(w http.ResponseWriter, r *http.Request) {
	handler.shared.ServeAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		viewer, viewerErr := handler.shared.CurrentActorProfile(r)
		if viewerErr != nil {
			return viewerErr
		}

		items, err := handler.shared.ReadService.ListInventoryByMerchant(r.Context(), viewer.UID, viewer.MerchantID.String())
		if err != nil {
			return err
		}

		response := make([]api.InventoryItemResponse, 0, len(items))
		for _, item := range items {
			response = append(response, inventoryItemResponse(item))
		}

		return httputils.Encode(w, r, http.StatusOK, response)
	})
}
