package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/pgsqlc"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
)

type OrderLineBill struct {
	ProductID      uuid.UUID
	Quantity       int32
	BaseAmount     float64
	AddonAmount    float64
	DiscountAmount float64
	TaxAmount      float64
	LineTotal      float64
}

type OrderBill struct {
	OrderID   uuid.UUID
	VatRate   float64
	Subtotal  float64
	TotalTax  float64
	Total     float64
	LineItems []OrderLineBill
}

type OrderDetail struct {
	Order pgsqlc.Order
	Items []OrderItemDetail
}

type OrderItemDetail struct {
	Item    pgsqlc.OrderItem
	Product pgsqlc.Product
	Addons  []pgsqlc.OrderItemAddon
}

type OrderService interface {
	PlaceOrderFromCart(ctx context.Context, merchantActorID uuid.UUID, cartID uuid.UUID, paymentType pgsqlc.PaymentType, deliveryAddress string, customerName string, customerPhone string) (OrderBill, *domainerr.DomainError)
	UpdateOrderStatus(ctx context.Context, merchantActorID uuid.UUID, merchantID uuid.UUID, orderID uuid.UUID, status pgsqlc.OrderStatusType) (pgsqlc.Order, *domainerr.DomainError)
	GetOrderDetail(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, orderID string) (OrderDetail, *domainerr.DomainError)
	ListOrdersByMerchant(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string) ([]pgsqlc.Order, *domainerr.DomainError)
}
