package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/order/store/generated"
)

type DBTX = orderstore.DBTX

type Postgres struct {
	queries *orderstore.Queries
}

func New(db DBTX) *Postgres {
	return &Postgres{queries: orderstore.New(db)}
}

func (store *Postgres) CreateOrder(ctx context.Context, arg CreateOrderParams) (Order, error) {
	return store.queries.CreateOrder(ctx, orderstore.CreateOrderParams(arg))
}

func (store *Postgres) CreateOrderGuest(ctx context.Context, arg CreateOrderGuestParams) (Order, error) {
	return store.queries.CreateOrderGuest(ctx, orderstore.CreateOrderGuestParams(arg))
}

func (store *Postgres) GetOrder(ctx context.Context, arg GetOrderParams) (Order, error) {
	return store.queries.GetOrder(ctx, orderstore.GetOrderParams(arg))
}

func (store *Postgres) UpdateOrder(ctx context.Context, arg UpdateOrderParams) (Order, error) {
	return store.queries.UpdateOrder(ctx, orderstore.UpdateOrderParams(arg))
}

func (store *Postgres) UpsertOrderItem(ctx context.Context, arg UpsertOrderItemParams) (OrderItem, error) {
	return store.queries.UpsertOrderItem(ctx, orderstore.UpsertOrderItemParams(arg))
}

func (store *Postgres) GetOrderItem(ctx context.Context, arg GetOrderItemParams) (OrderItem, error) {
	return store.queries.GetOrderItem(ctx, orderstore.GetOrderItemParams(arg))
}

func (store *Postgres) UpdateOrderItem(ctx context.Context, arg UpdateOrderItemParams) (OrderItem, error) {
	return store.queries.UpdateOrderItem(ctx, orderstore.UpdateOrderItemParams(arg))
}

func (store *Postgres) UpsertOrderItemAddon(ctx context.Context, arg UpsertOrderItemAddonParams) (OrderItemAddon, error) {
	return store.queries.UpsertOrderItemAddon(ctx, orderstore.UpsertOrderItemAddonParams(arg))
}

func (store *Postgres) GetOrderItemAddon(ctx context.Context, arg GetOrderItemAddonParams) (OrderItemAddon, error) {
	return store.queries.GetOrderItemAddon(ctx, orderstore.GetOrderItemAddonParams(arg))
}

func (store *Postgres) UpdateOrderItemAddon(ctx context.Context, arg UpdateOrderItemAddonParams) (OrderItemAddon, error) {
	return store.queries.UpdateOrderItemAddon(ctx, orderstore.UpdateOrderItemAddonParams(arg))
}

func (store *Postgres) ListOrdersByMerchant(ctx context.Context, merchantID uuid.UUID) ([]Order, error) {
	items, err := store.queries.ListOrdersByMerchant(ctx, merchantID)
	if err != nil {
		return nil, err
	}

	result := make([]Order, 0, len(items))
	for _, item := range items {
		result = append(result, item)
	}

	return result, nil
}

func (store *Postgres) ListOrderItemsByOrder(ctx context.Context, orderID uuid.UUID) ([]OrderItem, error) {
	items, err := store.queries.ListOrderItemsByOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}

	result := make([]OrderItem, 0, len(items))
	for _, item := range items {
		result = append(result, item)
	}

	return result, nil
}

func (store *Postgres) ListOrderItemAddonsByOrder(ctx context.Context, orderID uuid.UUID) ([]OrderItemAddon, error) {
	items, err := store.queries.ListOrderItemAddonsByOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}

	result := make([]OrderItemAddon, 0, len(items))
	for _, item := range items {
		result = append(result, item)
	}

	return result, nil
}

func (store *Postgres) GetVatRule(ctx context.Context, arg GetVatRuleParams) (VatRule, error) {
	return store.queries.GetVatRule(ctx, orderstore.GetVatRuleParams(arg))
}

func (store *Postgres) UpdateVatRuleByID(ctx context.Context, arg UpdateVatRuleByIDParams) (VatRule, error) {
	return store.queries.UpdateVatRuleByID(ctx, orderstore.UpdateVatRuleByIDParams(arg))
}

func (store *Postgres) UpsertVatRule(ctx context.Context, arg UpsertVatRuleParams) (VatRule, error) {
	return store.queries.UpsertVatRule(ctx, orderstore.UpsertVatRuleParams(arg))
}
