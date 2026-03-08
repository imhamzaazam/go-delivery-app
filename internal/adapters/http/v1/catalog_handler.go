package v1

import (
	"net/http"

	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/http/httputils"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (adapter *HTTPAdapter) ListProductCategoriesByMerchant(w http.ResponseWriter, r *http.Request) {
	adapter.serveAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		authUser, authErr := adapter.currentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		categories, err := adapter.readService.ListProductCategoriesByMerchant(r.Context(), authUser.MerchantID.String())
		if err != nil {
			return err
		}

		response := make([]ProductCategoryResponse, 0, len(categories))
		for _, category := range categories {
			description := textString(category.Description)
			createdAt := category.CreatedAt
			item := ProductCategoryResponse{
				Id:          ptrUUID(category.ID),
				MerchantId:  ptrUUID(category.MerchantID),
				Name:        &category.Name,
				Description: &description,
				CreatedAt:   &createdAt,
			}
			response = append(response, item)
		}

		return httputils.Encode(w, r, http.StatusOK, response)
	})
}

func (adapter *HTTPAdapter) ListProductsByMerchant(w http.ResponseWriter, r *http.Request) {
	adapter.serveAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		authUser, authErr := adapter.currentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		products, err := adapter.readService.ListProductsByMerchant(r.Context(), authUser.MerchantID.String())
		if err != nil {
			return err
		}

		response := make([]ProductResponse, 0, len(products))
		for _, product := range products {
			description := textString(product.Description)
			imageURL := textString(product.ImageUrl)
			basePrice := numericToFloat64(product.BasePrice)
			createdAt := product.CreatedAt
			updatedAt := product.UpdatedAt
			categoryID := openapi_types.UUID(product.CategoryID)
			item := ProductResponse{
				Id:             ptrUUID(product.ID),
				MerchantId:     ptrUUID(product.MerchantID),
				CategoryId:     &categoryID,
				Name:           &product.Name,
				Description:    &description,
				BasePrice:      &basePrice,
				ImageUrl:       &imageURL,
				TrackInventory: &product.TrackInventory,
				IsActive:       &product.IsActive,
				CreatedAt:      &createdAt,
				UpdatedAt:      &updatedAt,
			}
			response = append(response, item)
		}

		return httputils.Encode(w, r, http.StatusOK, response)
	})
}

func (adapter *HTTPAdapter) GetProductDetail(w http.ResponseWriter, r *http.Request, productID string) {
	adapter.serveAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		authUser, authErr := adapter.currentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		product, err := adapter.readService.GetProductDetail(r.Context(), authUser.MerchantID.String(), productID)
		if err != nil {
			return err
		}

		description := textString(product.Description)
		imageURL := textString(product.ImageUrl)
		basePrice := numericToFloat64(product.BasePrice)
		createdAt := product.CreatedAt
		updatedAt := product.UpdatedAt
		categoryName := product.CategoryName
		categoryID := openapi_types.UUID(product.CategoryID)
		response := ProductDetailResponse{
			Id:             ptrUUID(product.ID),
			MerchantId:     ptrUUID(product.MerchantID),
			CategoryId:     &categoryID,
			Name:           &product.Name,
			Description:    &description,
			BasePrice:      &basePrice,
			ImageUrl:       &imageURL,
			TrackInventory: &product.TrackInventory,
			IsActive:       &product.IsActive,
			CreatedAt:      &createdAt,
			UpdatedAt:      &updatedAt,
			CategoryName:   &categoryName,
		}

		return httputils.Encode(w, r, http.StatusOK, response)
	})
}

func (adapter *HTTPAdapter) ListProductAddonsByProduct(w http.ResponseWriter, r *http.Request, productID string) {
	adapter.serveAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		authUser, authErr := adapter.currentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		addons, err := adapter.readService.ListProductAddonsByProduct(r.Context(), authUser.MerchantID.String(), productID)
		if err != nil {
			return err
		}

		response := make([]ProductAddonResponse, 0, len(addons))
		for _, addon := range addons {
			price := numericToFloat64(addon.Price)
			createdAt := addon.CreatedAt
			item := ProductAddonResponse{
				Id:        ptrUUID(addon.ID),
				ProductId: ptrUUID(addon.ProductID),
				Name:      &addon.Name,
				Price:     &price,
				CreatedAt: &createdAt,
			}
			response = append(response, item)
		}

		return httputils.Encode(w, r, http.StatusOK, response)
	})
}

func (adapter *HTTPAdapter) ListInventoryByMerchant(w http.ResponseWriter, r *http.Request) {
	adapter.serveAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		viewer, viewerErr := adapter.currentActorProfile(r)
		if viewerErr != nil {
			return viewerErr
		}

		items, err := adapter.readService.ListInventoryByMerchant(r.Context(), viewer.UID, viewer.MerchantID.String())
		if err != nil {
			return err
		}

		response := make([]InventoryItemResponse, 0, len(items))
		for _, item := range items {
			quantity := int(item.Quantity)
			row := InventoryItemResponse{
				ProductId:   ptrUUID(item.ProductID),
				ProductName: &item.ProductName,
				Quantity:    &quantity,
			}
			response = append(response, row)
		}

		return httputils.Encode(w, r, http.StatusOK, response)
	})
}
