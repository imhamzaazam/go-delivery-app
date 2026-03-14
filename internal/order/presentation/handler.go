package presentation

import (
	"net/http"

	"github.com/google/uuid"
	orderstore "github.com/horiondreher/go-web-api-boilerplate/internal/order/store"
	openapi_types "github.com/oapi-codegen/runtime/types"

	api "github.com/horiondreher/go-web-api-boilerplate/api"
	"github.com/horiondreher/go-web-api-boilerplate/internal/core"
	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
	order "github.com/horiondreher/go-web-api-boilerplate/internal/order"
	"github.com/horiondreher/go-web-api-boilerplate/pkg/http-tools/httperr"
	"github.com/horiondreher/go-web-api-boilerplate/pkg/http-tools/httputils"
)

type Handler struct {
	shared *core.Shared
}

type createOrderValidation struct {
	CartID          string `validate:"required,uuid"`
	PaymentType     string `validate:"required,oneof=card cash"`
	DeliveryAddress string `validate:"required"`
	CustomerName    string `validate:"required"`
	CustomerPhone   string `validate:"required"`
}

type updateOrderStatusValidation struct {
	Status string `validate:"required,oneof=pending accepted out_for_delivery delivered refunded cancelled"`
}

func New(shared *core.Shared) *Handler {
	return &Handler{shared: shared}
}

func orderBillResponse(bill order.Bill) api.OrderBillResponse {
	paymentMethod := bill.PaymentType
	subtotal := bill.Subtotal
	totalTax := bill.TotalTax
	total := bill.Total
	vatRate := bill.VatRate
	lines := make([]api.OrderBillLineResponse, 0, len(bill.LineItems))
	for _, line := range bill.LineItems {
		productName := line.ProductName
		basePrice := line.BasePrice
		linePaymentMethod := line.PaymentMethod
		quantity := int(line.Quantity)
		baseAmount := line.BaseAmount
		addonAmount := line.AddonAmount
		discountAmount := line.DiscountAmount
		finalPrice := line.FinalPrice
		taxAmount := line.TaxAmount
		vat := line.Vat
		lineTotal := line.LineTotal
		totalPrice := line.TotalPrice
		lines = append(lines, api.OrderBillLineResponse{
			Name:           &productName,
			ProductId:      core.PtrUUID(line.ProductID),
			BasePrice:      &basePrice,
			PaymentMethod:  &linePaymentMethod,
			Quantity:       &quantity,
			BaseAmount:     &baseAmount,
			AddonAmount:    &addonAmount,
			DiscountAmount: &discountAmount,
			FinalPrice:     &finalPrice,
			TaxAmount:      &taxAmount,
			Vat:            &vat,
			LineTotal:      &lineTotal,
			TotalPrice:     &totalPrice,
		})
	}

	return api.OrderBillResponse{
		OrderId:       core.PtrUUID(bill.OrderID),
		PaymentMethod: &paymentMethod,
		Subtotal:      &subtotal,
		TotalTax:      &totalTax,
		Total:         &total,
		VatRate:       &vatRate,
		LineItems:     &lines,
	}
}

func orderSummaryResponse(order orderstore.Order) api.OrderSummaryResponse {
	totalAmount := core.NumericToFloat64(order.TotalAmount)
	vatRate := core.NumericToFloat64(order.VatRate)
	paymentType := string(order.PaymentType)
	status := string(order.Status)
	createdAt := order.CreatedAt
	updatedAt := order.UpdatedAt
	response := api.OrderSummaryResponse{
		Id:              core.PtrUUID(order.ID),
		CartId:          core.PtrUUID(order.CartID),
		MerchantId:      core.PtrUUID(order.MerchantID),
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
		actorID := openapi_types.UUID(uuid.UUID(order.ActorID.Bytes))
		response.ActorId = &actorID
	}
	if order.BranchID != uuid.Nil {
		branchID := openapi_types.UUID(order.BranchID)
		response.BranchId = &branchID
	}
	return response
}

func (handler *Handler) PlaceOrderFromCart(w http.ResponseWriter, r *http.Request) {
	handler.shared.Wrap(func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		requestBody, err := httputils.Decode[api.CreateOrderRequest](r)
		if err != nil {
			return err
		}
		validationErr := handler.shared.Validate.Struct(createOrderValidation{
			CartID:          requestBody.CartId.String(),
			PaymentType:     string(requestBody.PaymentType),
			DeliveryAddress: requestBody.DeliveryAddress,
			CustomerName:    requestBody.CustomerName,
			CustomerPhone:   requestBody.CustomerPhone,
		})
		if validationErr != nil {
			return httperr.MatchValidationError(validationErr)
		}

		orderBill, createErr := handler.shared.OrderService.PlaceOrderFromCartHTTP(r.Context(), requestBody.CartId.String(), string(requestBody.PaymentType), requestBody.DeliveryAddress, requestBody.CustomerName, requestBody.CustomerPhone)
		if createErr != nil {
			return createErr
		}

		return httputils.Encode(w, r, http.StatusCreated, orderBillResponse(orderBill))
	})(w, r)
}

func (handler *Handler) GetOrderDetail(w http.ResponseWriter, r *http.Request, orderID string) {
	handler.shared.Wrap(func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		order, err := handler.shared.OrderService.GetPublicOrderDetail(r.Context(), orderID)
		if err != nil {
			return err
		}

		items := make([]api.OrderLineResponse, 0, len(order.Items))
		for _, item := range order.Items {
			quantity := int(item.Item.Quantity)
			price := core.NumericToFloat64(item.Item.Price)
			baseAmount := core.NumericToFloat64(item.Item.BaseAmount)
			addonAmount := core.NumericToFloat64(item.Item.AddonAmount)
			discountAmount := core.NumericToFloat64(item.Item.DiscountAmount)
			taxAmount := core.NumericToFloat64(item.Item.TaxAmount)
			lineTotal := core.NumericToFloat64(item.Item.LineTotal)
			row := api.OrderLineResponse{
				OrderId:        core.PtrUUID(item.Item.OrderID),
				ProductId:      core.PtrUUID(item.Item.ProductID),
				Quantity:       &quantity,
				Price:          &price,
				BaseAmount:     &baseAmount,
				AddonAmount:    &addonAmount,
				DiscountAmount: &discountAmount,
				TaxAmount:      &taxAmount,
				LineTotal:      &lineTotal,
			}

			productDescription := core.TextString(item.Product.Description)
			productImageURL := core.TextString(item.Product.ImageUrl)
			productBasePrice := core.NumericToFloat64(item.Product.BasePrice)
			productCreatedAt := item.Product.CreatedAt
			productUpdatedAt := item.Product.UpdatedAt
			productCategoryID := openapi_types.UUID(item.Product.CategoryID)
			row.Product = &api.ProductResponse{
				Id:             core.PtrUUID(item.Product.ID),
				MerchantId:     core.PtrUUID(item.Product.MerchantID),
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
				addons := make([]api.OrderLineAddonResponse, 0, len(item.Addons))
				for _, addon := range item.Addons {
					addonQuantity := int(addon.Quantity)
					addonPrice := core.NumericToFloat64(addon.AddonPrice)
					lineAddonTotal := core.NumericToFloat64(addon.LineAddonTotal)
					addons = append(addons, api.OrderLineAddonResponse{
						OrderId:        core.PtrUUID(addon.OrderID),
						ProductId:      core.PtrUUID(addon.ProductID),
						AddonId:        core.PtrUUID(addon.AddonID),
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

		totalAmount := core.NumericToFloat64(order.Order.TotalAmount)
		vatRate := core.NumericToFloat64(order.Order.VatRate)
		paymentType := string(order.Order.PaymentType)
		status := string(order.Order.Status)
		createdAt := order.Order.CreatedAt
		updatedAt := order.Order.UpdatedAt
		response := api.OrderDetailResponse{
			Id:              core.PtrUUID(order.Order.ID),
			CartId:          core.PtrUUID(order.Order.CartID),
			MerchantId:      core.PtrUUID(order.Order.MerchantID),
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
			actorID := openapi_types.UUID(uuid.UUID(order.Order.ActorID.Bytes))
			response.ActorId = &actorID
		}
		if order.Order.BranchID != uuid.Nil {
			branchID := openapi_types.UUID(order.Order.BranchID)
			response.BranchId = &branchID
		}

		return httputils.Encode(w, r, http.StatusOK, response)
	})(w, r)
}

func (handler *Handler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request, orderID string) {
	handler.shared.ServeAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		requestBody, err := httputils.Decode[api.UpdateOrderStatusRequest](r)
		if err != nil {
			return err
		}
		validationErr := handler.shared.Validate.Struct(updateOrderStatusValidation{Status: string(requestBody.Status)})
		if validationErr != nil {
			return httperr.MatchValidationError(validationErr)
		}

		authUser, authErr := handler.shared.CurrentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		order, updateErr := handler.shared.OrderService.UpdateOrderStatusHTTP(r.Context(), authUser.MerchantID, authUser.Email, authUser.MerchantID.String(), orderID, string(requestBody.Status))
		if updateErr != nil {
			return updateErr
		}

		return httputils.Encode(w, r, http.StatusOK, orderSummaryResponse(order))
	})
}

func (handler *Handler) ListOrdersByMerchant(w http.ResponseWriter, r *http.Request) {
	handler.shared.ServeAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		authUser, authErr := handler.shared.CurrentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		orders, err := handler.shared.OrderService.ListOrdersByMerchant(r.Context(), authUser.MerchantID, authUser.Email, authUser.MerchantID.String())
		if err != nil {
			return err
		}

		response := make([]api.OrderSummaryResponse, 0, len(orders))
		for _, order := range orders {
			response = append(response, orderSummaryResponse(order))
		}

		return httputils.Encode(w, r, http.StatusOK, response)
	})
}
