package store

import (
	"github.com/horiondreher/go-web-api-boilerplate/internal/cart/store/generated"
)

type Cart = cartstore.Cart
type CartItem = cartstore.CartItem
type CreateCartParams = cartstore.CreateCartParams
type CreateGuestCartParams = cartstore.CreateGuestCartParams
type GetCartParams = cartstore.GetCartParams
type UpdateCartParams = cartstore.UpdateCartParams
type CreateCartItemParams = cartstore.CreateCartItemParams
type CreateCartItemRow = cartstore.CartItem
type GetCartItemBySignatureParams = cartstore.GetCartItemBySignatureParams
type GetCartItemBySignatureRow = cartstore.CartItem
type GetCartItemByIDParams = cartstore.GetCartItemByIDParams
type GetCartItemByIDRow = cartstore.CartItem
type UpdateCartItemByIDParams = cartstore.UpdateCartItemByIDParams
type UpdateCartItemByIDRow = cartstore.CartItem
type DeleteCartItemParams = cartstore.DeleteCartItemParams
type ListCartItemsByCartRow = cartstore.CartItem
