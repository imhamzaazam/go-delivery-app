package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/cart/store/generated"
)

type DBTX = cartstore.DBTX

type Postgres struct {
	queries *cartstore.Queries
}

func New(db DBTX) *Postgres {
	return &Postgres{queries: cartstore.New(db)}
}

func (store *Postgres) CreateCart(ctx context.Context, arg CreateCartParams) (Cart, error) {
	return store.queries.CreateCart(ctx, cartstore.CreateCartParams(arg))
}

func (store *Postgres) CreateGuestCart(ctx context.Context, arg CreateGuestCartParams) (Cart, error) {
	return store.queries.CreateGuestCart(ctx, cartstore.CreateGuestCartParams(arg))
}

func (store *Postgres) GetCart(ctx context.Context, arg GetCartParams) (Cart, error) {
	return store.queries.GetCart(ctx, cartstore.GetCartParams(arg))
}

func (store *Postgres) UpdateCart(ctx context.Context, arg UpdateCartParams) (Cart, error) {
	return store.queries.UpdateCart(ctx, cartstore.UpdateCartParams(arg))
}

func (store *Postgres) CreateCartItem(ctx context.Context, arg CreateCartItemParams) (CartItem, error) {
	return store.queries.CreateCartItem(ctx, cartstore.CreateCartItemParams(arg))
}

func (store *Postgres) GetCartItemBySignature(ctx context.Context, arg GetCartItemBySignatureParams) (CartItem, error) {
	return store.queries.GetCartItemBySignature(ctx, cartstore.GetCartItemBySignatureParams(arg))
}

func (store *Postgres) GetCartItemByID(ctx context.Context, arg GetCartItemByIDParams) (CartItem, error) {
	return store.queries.GetCartItemByID(ctx, cartstore.GetCartItemByIDParams(arg))
}

func (store *Postgres) UpdateCartItemByID(ctx context.Context, arg UpdateCartItemByIDParams) (CartItem, error) {
	return store.queries.UpdateCartItemByID(ctx, cartstore.UpdateCartItemByIDParams(arg))
}

func (store *Postgres) DeleteCartItem(ctx context.Context, arg DeleteCartItemParams) (int64, error) {
	return store.queries.DeleteCartItem(ctx, cartstore.DeleteCartItemParams(arg))
}

func (store *Postgres) ListCartItemsByCart(ctx context.Context, cartID uuid.UUID) ([]ListCartItemsByCartRow, error) {
	items, err := store.queries.ListCartItemsByCart(ctx, cartID)
	if err != nil {
		return nil, err
	}

	result := make([]ListCartItemsByCartRow, 0, len(items))
	for _, item := range items {
		result = append(result, item)
	}

	return result, nil
}
