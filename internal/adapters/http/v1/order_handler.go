package v1

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/http/httputils"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (adapter *HTTPAdapter) GetOrderDetail(w http.ResponseWriter, r *http.Request, orderID string) {
	adapter.serveAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		authUser, authErr := adapter.currentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		order, err := adapter.readService.GetOrderDetail(r.Context(), authUser.MerchantID, authUser.Email, orderID)
		if err != nil {
			return err
		}

		items := make([]OrderLineResponse, 0, len(order.Items))
		for _, item := range order.Items {
			quantity := int(item.Item.Quantity)
			price := numericToFloat64(item.Item.Price)
			baseAmount := numericToFloat64(item.Item.BaseAmount)
			addonAmount := numericToFloat64(item.Item.AddonAmount)
			discountAmount := numericToFloat64(item.Item.DiscountAmount)
			taxAmount := numericToFloat64(item.Item.TaxAmount)
			lineTotal := numericToFloat64(item.Item.LineTotal)
			row := OrderLineResponse{
				OrderId:        ptrUUID(item.Item.OrderID),
				ProductId:      ptrUUID(item.Item.ProductID),
				Quantity:       &quantity,
				Price:          &price,
				BaseAmount:     &baseAmount,
				AddonAmount:    &addonAmount,
				DiscountAmount: &discountAmount,
				TaxAmount:      &taxAmount,
				LineTotal:      &lineTotal,
			}

			productDescription := textString(item.Product.Description)
			productImageURL := textString(item.Product.ImageUrl)
			productBasePrice := numericToFloat64(item.Product.BasePrice)
			productCreatedAt := item.Product.CreatedAt
			productUpdatedAt := item.Product.UpdatedAt
			productCategoryID := openapi_types.UUID(item.Product.CategoryID)
			row.Product = &ProductResponse{
				Id:             ptrUUID(item.Product.ID),
				MerchantId:     ptrUUID(item.Product.MerchantID),
				CategoryId:     &productCategoryID,
				Name:           &item.Product.Name,
				Description:    &productDescription,
				BasePrice:      &productBasePrice,
				ImageUrl:       &productImageURL,
				TrackInventory: &item.Product.TrackInventory,
				IsActive:       &item.Product.IsActive,
				CreatedAt:      &productCreatedAt,
				UpdatedAt:      &productUpdatedAt,
			}

			if len(item.Addons) > 0 {
				addons := make([]OrderLineAddonResponse, 0, len(item.Addons))
				for _, addon := range item.Addons {
					addonQuantity := int(addon.Quantity)
					addonPrice := numericToFloat64(addon.AddonPrice)
					lineAddonTotal := numericToFloat64(addon.LineAddonTotal)
					addons = append(addons, OrderLineAddonResponse{
						OrderId:        ptrUUID(addon.OrderID),
						ProductId:      ptrUUID(addon.ProductID),
						AddonId:        ptrUUID(addon.AddonID),
						AddonName:      &addon.AddonName,
						AddonPrice:     &addonPrice,
						Quantity:       &addonQuantity,
						LineAddonTotal: &lineAddonTotal,
					})
				}
				row.Addons = &addons
			}

			items = append(items, row)
		}

		totalAmount := numericToFloat64(order.Order.TotalAmount)
		vatRate := numericToFloat64(order.Order.VatRate)
		paymentType := string(order.Order.PaymentType)
		status := string(order.Order.Status)
		createdAt := order.Order.CreatedAt
		updatedAt := order.Order.UpdatedAt
		response := OrderDetailResponse{
			Id:              ptrUUID(order.Order.ID),
			CartId:          ptrUUID(order.Order.CartID),
			MerchantId:      ptrUUID(order.Order.MerchantID),
			PaymentType:     &paymentType,
			VatRate:         &vatRate,
			TotalAmount:     &totalAmount,
			Status:          &status,
			DeliveryAddress: &order.Order.DeliveryAddress,
			CustomerName:    &order.Order.CustomerName,
			CustomerPhone:   &order.Order.CustomerPhone,
			CreatedAt:       &createdAt,
			UpdatedAt:       &updatedAt,
			Items:           &items,
		}
		if order.Order.ActorID.Valid {
			actorID := order.Order.ActorID.Bytes
			response.ActorId = (*openapi_types.UUID)(&actorID)
		}
		if order.Order.BranchID != uuid.Nil {
			branchID := order.Order.BranchID
			response.BranchId = &branchID
		}

		return httputils.Encode(w, r, http.StatusOK, response)
	})
}

func (adapter *HTTPAdapter) ListOrdersByMerchant(w http.ResponseWriter, r *http.Request) {
	adapter.serveAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		authUser, authErr := adapter.currentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		orders, err := adapter.readService.ListOrdersByMerchant(r.Context(), authUser.MerchantID, authUser.Email, authUser.MerchantID.String())
		if err != nil {
			return err
		}

		response := make([]OrderSummaryResponse, 0, len(orders))
		for _, order := range orders {
			totalAmount := numericToFloat64(order.TotalAmount)
			vatRate := numericToFloat64(order.VatRate)
			paymentType := string(order.PaymentType)
			status := string(order.Status)
			createdAt := order.CreatedAt
			updatedAt := order.UpdatedAt
			item := OrderSummaryResponse{
				Id:              ptrUUID(order.ID),
				CartId:          ptrUUID(order.CartID),
				MerchantId:      ptrUUID(order.MerchantID),
				PaymentType:     &paymentType,
				VatRate:         &vatRate,
				TotalAmount:     &totalAmount,
				Status:          &status,
				DeliveryAddress: &order.DeliveryAddress,
				CustomerName:    &order.CustomerName,
				CustomerPhone:   &order.CustomerPhone,
				CreatedAt:       &createdAt,
				UpdatedAt:       &updatedAt,
			}
			if order.ActorID.Valid {
				actorID := order.ActorID.Bytes
				item.ActorId = (*openapi_types.UUID)(&actorID)
			}
			if order.BranchID != uuid.Nil {
				branchID := order.BranchID
				item.BranchId = &branchID
			}
			response = append(response, item)
		}

		return httputils.Encode(w, r, http.StatusOK, response)
	})
}
