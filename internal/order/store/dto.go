package store

import (
	"github.com/horiondreher/go-web-api-boilerplate/internal/order/store/generated"
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
