package app

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
	orderdomain "github.com/horiondreher/go-web-api-boilerplate/internal/order"
	"github.com/horiondreher/go-web-api-boilerplate/internal/utils"
)

func (service *Service) PlaceOrderFromCartHTTP(ctx context.Context, cartID string, paymentType string, deliveryAddress string, customerName string, customerPhone string) (orderdomain.Bill, *domainerr.DomainError) {
	parsedCartID, cartErr := utils.ParseUUID(cartID, "cart id")
	if cartErr != nil {
		return orderdomain.Bill{}, cartErr
	}

	parsedPaymentType := parsePaymentType(paymentType)
	if paymentErr := validateParsedPaymentType(parsedPaymentType); paymentErr != nil {
		return orderdomain.Bill{}, paymentErr
	}

	orderBill, placeErr := service.PlaceOrderFromCart(ctx, uuid.Nil, parsedCartID, parsedPaymentType, deliveryAddress, customerName, customerPhone)
	if placeErr != nil {
		return orderdomain.Bill{}, placeErr
	}

	return toDomainBill(orderBill), nil
}

func (service *Service) UpdateOrderStatusHTTP(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, orderID string, status string) (orderdomain.Order, *domainerr.DomainError) {
	parsedMerchantID, merchantErr := utils.ParseUUID(merchantID, "merchant id")
	if merchantErr != nil {
		return orderdomain.Order{}, merchantErr
	}
	parsedOrderID, orderErr := utils.ParseUUID(orderID, "order id")
	if orderErr != nil {
		return orderdomain.Order{}, orderErr
	}

	viewer, accessErr := service.requireMerchantManageAccess(ctx, viewerMerchantID, viewerEmail, parsedMerchantID)
	if accessErr != nil {
		return orderdomain.Order{}, accessErr
	}

	parsedStatus := parseOrderStatus(status)
	if statusErr := validateParsedOrderStatus(parsedStatus); statusErr != nil {
		return orderdomain.Order{}, statusErr
	}

	updatedOrder, updateErr := service.UpdateOrderStatus(ctx, viewer.UID, parsedMerchantID, parsedOrderID, parsedStatus)
	if updateErr != nil {
		return orderdomain.Order{}, updateErr
	}

	return updatedOrder, nil
}

func parsePaymentType(value string) orderdomain.PaymentType {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(orderdomain.PaymentTypeCard):
		return orderdomain.PaymentTypeCard
	case string(orderdomain.PaymentTypeCash):
		return orderdomain.PaymentTypeCash
	default:
		return orderdomain.PaymentType("")
	}
}

func parseOrderStatus(value string) orderdomain.OrderStatusType {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(orderdomain.OrderStatusTypePending):
		return orderdomain.OrderStatusTypePending
	case string(orderdomain.OrderStatusTypeAccepted):
		return orderdomain.OrderStatusTypeAccepted
	case string(orderdomain.OrderStatusTypeOutForDelivery):
		return orderdomain.OrderStatusTypeOutForDelivery
	case string(orderdomain.OrderStatusTypeDelivered):
		return orderdomain.OrderStatusTypeDelivered
	case string(orderdomain.OrderStatusTypeRefunded):
		return orderdomain.OrderStatusTypeRefunded
	case string(orderdomain.OrderStatusTypeCancelled):
		return orderdomain.OrderStatusTypeCancelled
	default:
		return orderdomain.OrderStatusType("")
	}
}

func validateParsedPaymentType(paymentType orderdomain.PaymentType) *domainerr.DomainError {
	if strings.TrimSpace(string(paymentType)) == "" {
		return domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "invalid payment type", fmt.Errorf("invalid payment type"))
	}

	return nil
}

func validateParsedOrderStatus(status orderdomain.OrderStatusType) *domainerr.DomainError {
	if strings.TrimSpace(string(status)) == "" {
		return domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "invalid order status", fmt.Errorf("invalid order status"))
	}

	return nil
}

func toDomainBill(bill OrderBill) orderdomain.Bill {
	lineItems := make([]orderdomain.LineBill, 0, len(bill.LineItems))
	for _, line := range bill.LineItems {
		lineItems = append(lineItems, orderdomain.LineBill{
			ProductID:      line.ProductID,
			ProductName:    line.ProductName,
			BasePrice:      line.BasePrice,
			PaymentMethod:  line.PaymentMethod,
			Quantity:       line.Quantity,
			BaseAmount:     line.BaseAmount,
			AddonAmount:    line.AddonAmount,
			DiscountAmount: line.DiscountAmount,
			FinalPrice:     line.FinalPrice,
			TaxAmount:      line.TaxAmount,
			Vat:            line.Vat,
			LineTotal:      line.LineTotal,
			TotalPrice:     line.TotalPrice,
		})
	}

	return orderdomain.Bill{
		OrderID:     bill.OrderID,
		PaymentType: bill.PaymentType,
		VatRate:     bill.VatRate,
		Subtotal:    bill.Subtotal,
		TotalTax:    bill.TotalTax,
		Total:       bill.Total,
		LineItems:   lineItems,
	}
}
