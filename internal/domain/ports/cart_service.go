package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/pgsqlc"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
)

type CartDetail struct {
	Cart    pgsqlc.Cart
	Items   []CartItemDetail
	VatRate float64
}

type CartItemDetail struct {
	Item     pgsqlc.CartItem
	Product  pgsqlc.Product
	Addons   []pgsqlc.ProductAddon
	Discount *pgsqlc.MerchantDiscount
}

type CartService interface {
	CreateCart(ctx context.Context, cartID uuid.UUID, merchantID uuid.UUID, branchID uuid.UUID, actorID uuid.UUID) (pgsqlc.Cart, *domainerr.DomainError)
	AddItemToCart(ctx context.Context, cartID uuid.UUID, productID uuid.UUID, quantity int32, addonIDs []uuid.UUID, discountID uuid.UUID, discountAmount float64) (pgsqlc.CartItem, *domainerr.DomainError)
	GetCartDetail(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, cartID string, paymentType string) (CartDetail, *domainerr.DomainError)
}
