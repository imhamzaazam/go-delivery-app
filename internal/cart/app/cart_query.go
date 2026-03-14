package app

import (
	"context"

	"github.com/google/uuid"

	cartdomain "github.com/horiondreher/go-web-api-boilerplate/internal/cart"
	commerce "github.com/horiondreher/go-web-api-boilerplate/internal/commerce"
	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
	"github.com/horiondreher/go-web-api-boilerplate/internal/utils"
)

func (service *Service) GetCartDetail(ctx context.Context, cartID string) (cartdomain.Detail, *domainerr.DomainError) {
	parsedCartID, parseErr := utils.ParseUUID(cartID, "cart id")
	if parseErr != nil {
		return cartdomain.Detail{}, parseErr
	}

	cart, cartErr := service.getCartByID(ctx, parsedCartID)
	if cartErr != nil {
		return cartdomain.Detail{}, cartErr
	}

	items, err := service.store.ListCartItemsByCart(ctx, cart.ID)
	if err != nil {
		return cartdomain.Detail{}, domainerr.MatchPostgresError(err)
	}

	detailItems := make([]cartdomain.ItemDetail, 0, len(items))

	for _, item := range items {
		product, productErr := service.store.GetProduct(ctx, commerce.GetProductParams{
			MerchantID: cart.MerchantID,
			ID:         item.ProductID,
		})
		if productErr != nil {
			return cartdomain.Detail{}, domainerr.MatchPostgresError(productErr)
		}

		addons := make([]commerce.ProductAddon, 0, len(item.AddonIds))
		for _, addonID := range item.AddonIds {
			addon, addonErr := service.store.GetProductAddon(ctx, addonID)
			if addonErr != nil {
				return cartdomain.Detail{}, domainerr.MatchPostgresError(addonErr)
			}
			addons = append(addons, addon)
		}

		var discount *commerce.MerchantDiscount
		if item.AppliedDiscountID != uuid.Nil {
			value, discountErr := service.store.GetMerchantDiscount(ctx, commerce.GetMerchantDiscountParams{
				MerchantID: cart.MerchantID,
				ID:         item.AppliedDiscountID,
			})
			if discountErr == nil {
				resolved := merchantDiscountFromGetRow(value)
				discount = &resolved
			}
		}

		detailItems = append(detailItems, cartdomain.ItemDetail{
			Item:     cartItemFromListRow(item),
			Product:  product,
			Addons:   addons,
			Discount: discount,
		})
	}

	return cartdomain.Detail{
		Cart:  cart,
		Items: detailItems,
	}, nil
}
