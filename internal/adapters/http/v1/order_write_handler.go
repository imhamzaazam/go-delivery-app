package v1

import (
	"net/http"

	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/http/httperr"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/http/httputils"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
)

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

func (adapter *HTTPAdapter) PlaceOrderFromCart(w http.ResponseWriter, r *http.Request) {
	adapter.handlerWrapper(func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		requestBody, err := httputils.Decode[CreateOrderRequest](r)
		if err != nil {
			return err
		}
		validationErr := validate.Struct(createOrderValidation{
			CartID:          requestBody.CartId.String(),
			PaymentType:     string(requestBody.PaymentType),
			DeliveryAddress: requestBody.DeliveryAddress,
			CustomerName:    requestBody.CustomerName,
			CustomerPhone:   requestBody.CustomerPhone,
		})
		if validationErr != nil {
			return httperr.MatchValidationError(validationErr)
		}

		orderBill, createErr := adapter.commerceService.PlaceOrderFromCartHTTP(r.Context(), requestBody.CartId.String(), string(requestBody.PaymentType), requestBody.DeliveryAddress, requestBody.CustomerName, requestBody.CustomerPhone)
		if createErr != nil {
			return createErr
		}

		return httputils.Encode(w, r, http.StatusCreated, orderBillResponse(orderBill))
	})(w, r)
}

func (adapter *HTTPAdapter) UpdateOrderStatus(w http.ResponseWriter, r *http.Request, orderID string) {
	adapter.serveAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		requestBody, err := httputils.Decode[UpdateOrderStatusRequest](r)
		if err != nil {
			return err
		}
		validationErr := validate.Struct(updateOrderStatusValidation{
			Status: string(requestBody.Status),
		})
		if validationErr != nil {
			return httperr.MatchValidationError(validationErr)
		}

		authUser, authErr := adapter.currentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		order, updateErr := adapter.commerceService.UpdateOrderStatusHTTP(r.Context(), authUser.MerchantID, authUser.Email, authUser.MerchantID.String(), orderID, string(requestBody.Status))
		if updateErr != nil {
			return updateErr
		}

		return httputils.Encode(w, r, http.StatusOK, orderSummaryResponse(order))
	})
}
