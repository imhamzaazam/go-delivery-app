package cart

import (
	"context"

	"github.com/google/uuid"
	cartstore "github.com/horiondreher/go-web-api-boilerplate/internal/cart/store"

	catalog "github.com/horiondreher/go-web-api-boilerplate/internal/catalog"
	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
	merchant "github.com/horiondreher/go-web-api-boilerplate/internal/merchant"
)

type Cart = cartstore.Cart
type CartItem = cartstore.CartItem
type CreateCartParams = cartstore.CreateCartParams
type CreateGuestCartParams = cartstore.CreateGuestCartParams
type GetCartParams = cartstore.GetCartParams
type UpdateCartParams = cartstore.UpdateCartParams
type CreateCartItemParams = cartstore.CreateCartItemParams
type CreateCartItemRow = cartstore.CreateCartItemRow
type GetCartItemBySignatureParams = cartstore.GetCartItemBySignatureParams
type GetCartItemBySignatureRow = cartstore.GetCartItemBySignatureRow
type GetCartItemByIDParams = cartstore.GetCartItemByIDParams
type GetCartItemByIDRow = cartstore.GetCartItemByIDRow
type UpdateCartItemByIDParams = cartstore.UpdateCartItemByIDParams
type UpdateCartItemByIDRow = cartstore.UpdateCartItemByIDRow
type DeleteCartItemParams = cartstore.DeleteCartItemParams
type ListCartItemsByCartRow = cartstore.ListCartItemsByCartRow

type Detail struct {
	Cart  Cart
	Items []ItemDetail
}

type ItemDetail struct {
	Item     CartItem
	Product  catalog.Product
	Addons   []catalog.ProductAddon
	Discount *merchant.MerchantDiscount
}

type Service interface {
	CreateCart(ctx context.Context, cartID uuid.UUID, merchantID uuid.UUID, branchID uuid.UUID, actorID uuid.UUID) (Cart, *domainerr.DomainError)
	AddItemToCart(ctx context.Context, cartID uuid.UUID, productID uuid.UUID, quantity int32, addonIDs []uuid.UUID, discountID uuid.UUID, discountAmount float64) (CartItem, *domainerr.DomainError)
	RemoveItemFromCart(ctx context.Context, cartID uuid.UUID, itemID uuid.UUID) *domainerr.DomainError
	GetCartDetail(ctx context.Context, cartID string) (Detail, *domainerr.DomainError)
}
