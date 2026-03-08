package services

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/pgsqlc"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (service *CommerceManager) CreateCart(ctx context.Context, cartID uuid.UUID, merchantID uuid.UUID, branchID uuid.UUID, actorID uuid.UUID) (pgsqlc.Cart, *domainerr.DomainError) {
	actor := pgtype.UUID{Valid: false}
	if actorID != uuid.Nil {
		actor = pgtype.UUID{Bytes: actorID, Valid: true}
	}

	createdCart, createErr := service.store.CreateCart(ctx, pgsqlc.CreateCartParams{
		ID:         cartID,
		MerchantID: merchantID,
		BranchID:   branchID,
		ActorID:    actor,
	})
	if createErr != nil {
		return pgsqlc.Cart{}, domainerr.MatchPostgresError(createErr)
	}
	return createdCart, nil
}

func (service *CommerceManager) AddItemToCart(ctx context.Context, cartID uuid.UUID, productID uuid.UUID, quantity int32, addonIDs []uuid.UUID, discountID uuid.UUID, discountAmount float64) (pgsqlc.CartItem, *domainerr.DomainError) {
	if quantity <= 0 {
		return pgsqlc.CartItem{}, domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "quantity must be greater than zero", fmt.Errorf("quantity must be greater than zero"))
	}

	cart, cartErr := service.getCartByID(ctx, cartID)
	if cartErr != nil {
		return pgsqlc.CartItem{}, cartErr
	}

	product, productErr := service.store.GetProduct(ctx, pgsqlc.GetProductParams{
		MerchantID: cart.MerchantID,
		ID:         productID,
	})
	if productErr != nil {
		return pgsqlc.CartItem{}, domainerr.MatchPostgresError(productErr)
	}

	addonSubtotal := 0.0
	for _, addonID := range addonIDs {
		addon, addonErr := service.store.GetProductAddon(ctx, addonID)
		if addonErr != nil {
			return pgsqlc.CartItem{}, domainerr.MatchPostgresError(addonErr)
		}
		if addon.ProductID != product.ID {
			return pgsqlc.CartItem{}, domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "addon does not belong to product", fmt.Errorf("addon %s does not belong to product %s", addonID, productID))
		}
		addonSubtotal += numericToFloat(addon.Price) * float64(quantity)
	}

	lineSubtotal := round2((float64(quantity) * numericToFloat(product.BasePrice)) + addonSubtotal)
	appliedDiscountID := discountID
	if discountID == uuid.Nil {
		noDiscountID, noDiscountErr := service.ensureNoDiscountID(ctx, cartID)
		if noDiscountErr != nil {
			return pgsqlc.CartItem{}, domainerr.NewInternalError(noDiscountErr)
		}
		appliedDiscountID = noDiscountID
		discountAmount = 0
	} else {
		merchantDiscount, merchantDiscountErr := service.store.GetMerchantDiscount(ctx, pgsqlc.GetMerchantDiscountParams{
			MerchantID: cart.MerchantID,
			ID:         discountID,
		})
		if merchantDiscountErr != nil {
			return pgsqlc.CartItem{}, domainerr.MatchPostgresError(merchantDiscountErr)
		}

		now := time.Now()
		if !merchantDiscount.ValidFrom.IsZero() && now.Before(merchantDiscount.ValidFrom) {
			return pgsqlc.CartItem{}, domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "discount is not active yet", fmt.Errorf("discount is not active yet"))
		}
		if !merchantDiscount.ValidTo.IsZero() && now.After(merchantDiscount.ValidTo) {
			return pgsqlc.CartItem{}, domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "discount has expired", fmt.Errorf("discount has expired"))
		}

		discountAmount = calculateDiscountAmount(merchantDiscount, lineSubtotal)
	}
	if discountAmount < 0 || discountAmount > lineSubtotal {
		return pgsqlc.CartItem{}, domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "discount exceeds subtotal", fmt.Errorf("discount exceeds subtotal"))
	}

	cartItem, upsertErr := service.store.UpsertCartItem(ctx, pgsqlc.UpsertCartItemParams{
		CartID:                cartID,
		ProductID:             productID,
		Quantity:              quantity,
		AddonIds:              addonIDs,
		AppliedDiscountID:     appliedDiscountID,
		AppliedDiscountAmount: numericFromFloat(discountAmount),
	})
	if upsertErr != nil {
		return pgsqlc.CartItem{}, domainerr.MatchPostgresError(upsertErr)
	}

	return cartItem, nil
}

func (service *CommerceManager) getCartByID(ctx context.Context, cartID uuid.UUID) (pgsqlc.Cart, *domainerr.DomainError) {
	var cart pgsqlc.Cart
	err := service.db.QueryRow(ctx, `
		SELECT id, merchant_id, branch_id, actor_id, created_at, updated_at
		FROM carts
		WHERE id = $1
	`, cartID).Scan(&cart.ID, &cart.MerchantID, &cart.BranchID, &cart.ActorID, &cart.CreatedAt, &cart.UpdatedAt)
	if err != nil {
		return pgsqlc.Cart{}, domainerr.MatchPostgresError(err)
	}

	return cart, nil
}

func (service *CommerceManager) ensureNoDiscountID(ctx context.Context, cartID uuid.UUID) (uuid.UUID, error) {
	cart, cartErr := service.getCartByID(ctx, cartID)
	if cartErr != nil {
		return uuid.Nil, cartErr
	}

	merchantID := cart.MerchantID

	var discountID uuid.UUID
	err := service.db.QueryRow(ctx, `
		SELECT id
		FROM merchant_discounts
		WHERE merchant_id = $1 AND description = 'NO_DISCOUNT'
		LIMIT 1
	`, merchantID).Scan(&discountID)
	if err == nil {
		return discountID, nil
	}
	if err != pgx.ErrNoRows {
		return uuid.Nil, err
	}

	err = service.db.QueryRow(ctx, `
		INSERT INTO merchant_discounts (merchant_id, type, value, description, valid_from, valid_to)
		VALUES ($1, 'flat', 0, 'NO_DISCOUNT', NOW() - INTERVAL '1 day', NOW() + INTERVAL '100 year')
		RETURNING id
	`, merchantID).Scan(&discountID)
	if err != nil {
		return uuid.Nil, err
	}
	return discountID, nil
}

func calculateDiscountAmount(discount pgsqlc.MerchantDiscount, subtotal float64) float64 {
	discountValue := numericToFloat(discount.Value)
	switch discount.Type {
	case pgsqlc.DiscountTypePercentage:
		return round2((subtotal * discountValue) / 100)
	case pgsqlc.DiscountTypeFlat:
		fallthrough
	default:
		if discountValue > subtotal {
			return round2(subtotal)
		}
		return round2(discountValue)
	}
}
