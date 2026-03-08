package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/pgsqlc"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/ports"
)

func (service *CommerceManager) GetCartDetail(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, cartID string, paymentType string) (ports.CartDetail, *domainerr.DomainError) {
	parsedCartID, parseErr := parseUUID(cartID, "cart id")
	if parseErr != nil {
		return ports.CartDetail{}, parseErr
	}

	cart, cartErr := service.getCartByID(ctx, parsedCartID)
	if cartErr != nil {
		return ports.CartDetail{}, cartErr
	}

	viewer, viewerErr := service.resolveViewerActor(ctx, viewerMerchantID, viewerEmail)
	if viewerErr != nil {
		return ports.CartDetail{}, viewerErr
	}

	canViewMerchant, allowErr := service.canViewMerchant(ctx, viewer.UID, cart.MerchantID)
	if allowErr != nil {
		return ports.CartDetail{}, domainerr.NewInternalError(allowErr)
	}
	cartActorMatchesViewer := cart.ActorID.Valid && uuid.UUID(cart.ActorID.Bytes) == viewer.UID
	if !canViewMerchant && !cartActorMatchesViewer {
		return ports.CartDetail{}, domainerr.NewDomainError(403, domainerr.UnauthorizedError, "not allowed to view cart", fmt.Errorf("not allowed to view cart"))
	}

	items, err := service.store.ListCartItemsByCart(ctx, cart.ID)
	if err != nil {
		return ports.CartDetail{}, domainerr.MatchPostgresError(err)
	}

	detailItems := make([]ports.CartItemDetail, 0, len(items))
	vatRate := 0.0
	if paymentType != "" {
		vatRule, vatErr := service.store.GetVatRule(ctx, pgsqlc.GetVatRuleParams{
			MerchantID:  cart.MerchantID,
			PaymentType: pgsqlc.PaymentType(paymentType),
		})
		if vatErr == nil {
			vatRate = numericToFloat(vatRule.Rate)
		}
	}

	for _, item := range items {
		product, productErr := service.store.GetProduct(ctx, pgsqlc.GetProductParams{
			MerchantID: cart.MerchantID,
			ID:         item.ProductID,
		})
		if productErr != nil {
			return ports.CartDetail{}, domainerr.MatchPostgresError(productErr)
		}

		addons := make([]pgsqlc.ProductAddon, 0, len(item.AddonIds))
		for _, addonID := range item.AddonIds {
			addon, addonErr := service.store.GetProductAddon(ctx, addonID)
			if addonErr != nil {
				return ports.CartDetail{}, domainerr.MatchPostgresError(addonErr)
			}
			addons = append(addons, addon)
		}

		var discount *pgsqlc.MerchantDiscount
		if item.AppliedDiscountID != uuid.Nil {
			value, discountErr := service.store.GetMerchantDiscount(ctx, pgsqlc.GetMerchantDiscountParams{
				MerchantID: cart.MerchantID,
				ID:         item.AppliedDiscountID,
			})
			if discountErr == nil {
				discount = &value
			}
		}

		detailItems = append(detailItems, ports.CartItemDetail{
			Item:     item,
			Product:  product,
			Addons:   addons,
			Discount: discount,
		})
	}

	return ports.CartDetail{
		Cart:    cart,
		Items:   detailItems,
		VatRate: vatRate,
	}, nil
}
