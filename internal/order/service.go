package order

import (
	"context"

	"github.com/google/uuid"
	orderstore "github.com/horiondreher/go-web-api-boilerplate/internal/order/store"

	catalog "github.com/horiondreher/go-web-api-boilerplate/internal/catalog"
	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
)

type Order = orderstore.Order
type OrderItem = orderstore.OrderItem
type OrderItemAddon = orderstore.OrderItemAddon
type OrderStatusType = orderstore.OrderStatusType
type PaymentType = orderstore.PaymentType
type VatRule = orderstore.VatRule
type CreateOrderParams = orderstore.CreateOrderParams
type CreateOrderGuestParams = orderstore.CreateOrderGuestParams
type GetOrderParams = orderstore.GetOrderParams
type UpdateOrderParams = orderstore.UpdateOrderParams
type UpsertOrderItemParams = orderstore.UpsertOrderItemParams
type GetOrderItemParams = orderstore.GetOrderItemParams
type UpdateOrderItemParams = orderstore.UpdateOrderItemParams
type UpsertOrderItemAddonParams = orderstore.UpsertOrderItemAddonParams
type GetOrderItemAddonParams = orderstore.GetOrderItemAddonParams
type UpdateOrderItemAddonParams = orderstore.UpdateOrderItemAddonParams
type GetVatRuleParams = orderstore.GetVatRuleParams
type UpdateVatRuleByIDParams = orderstore.UpdateVatRuleByIDParams
type UpsertVatRuleParams = orderstore.UpsertVatRuleParams

const OrderStatusTypeAccepted = orderstore.OrderStatusTypeAccepted
const OrderStatusTypeCancelled = orderstore.OrderStatusTypeCancelled
const OrderStatusTypeDelivered = orderstore.OrderStatusTypeDelivered
const OrderStatusTypeOutForDelivery = orderstore.OrderStatusTypeOutForDelivery
const OrderStatusTypePending = orderstore.OrderStatusTypePending
const OrderStatusTypeRefunded = orderstore.OrderStatusTypeRefunded

const PaymentTypeCard = orderstore.PaymentTypeCard
const PaymentTypeCash = orderstore.PaymentTypeCash

type LineBill struct {
	ProductID      uuid.UUID
	ProductName    string
	BasePrice      float64
	PaymentMethod  string
	Quantity       int32
	BaseAmount     float64
	AddonAmount    float64
	DiscountAmount float64
	FinalPrice     float64
	TaxAmount      float64
	Vat            float64
	LineTotal      float64
	TotalPrice     float64
}

type Bill struct {
	OrderID     uuid.UUID
	PaymentType string
	VatRate     float64
	Subtotal    float64
	TotalTax    float64
	Total       float64
	LineItems   []LineBill
}

type Detail struct {
	Order Order
	Items []ItemDetail
}

type ItemDetail struct {
	Item    OrderItem
	Product catalog.Product
	Addons  []OrderItemAddon
}

type Service interface {
	PlaceOrderFromCart(ctx context.Context, merchantActorID uuid.UUID, cartID uuid.UUID, paymentType PaymentType, deliveryAddress string, customerName string, customerPhone string) (Bill, *domainerr.DomainError)
	UpdateOrderStatus(ctx context.Context, merchantActorID uuid.UUID, merchantID uuid.UUID, orderID uuid.UUID, status OrderStatusType) (Order, *domainerr.DomainError)
	GetOrderDetail(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, orderID string) (Detail, *domainerr.DomainError)
	ListOrdersByMerchant(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string) ([]Order, *domainerr.DomainError)
}
