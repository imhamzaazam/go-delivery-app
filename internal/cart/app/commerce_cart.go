package app

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	pgsqlc "github.com/horiondreher/go-web-api-boilerplate/internal/commerce"
	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
	"github.com/horiondreher/go-web-api-boilerplate/internal/utils"
)

func cartItemFromCreateRow(row pgsqlc.CreateCartItemRow) pgsqlc.CartItem {
	return pgsqlc.CartItem{
		ID:                    row.ID,
		CartID:                row.CartID,
		ProductID:             row.ProductID,
		Quantity:              row.Quantity,
		AddonIds:              row.AddonIds,
		AppliedDiscountID:     row.AppliedDiscountID,
		AppliedDiscountAmount: row.AppliedDiscountAmount,
	}
}

func cartItemFromSignatureRow(row pgsqlc.GetCartItemBySignatureRow) pgsqlc.CartItem {
	return pgsqlc.CartItem{
		ID:                    row.ID,
		CartID:                row.CartID,
		ProductID:             row.ProductID,
		Quantity:              row.Quantity,
		AddonIds:              row.AddonIds,
		AppliedDiscountID:     row.AppliedDiscountID,
		AppliedDiscountAmount: row.AppliedDiscountAmount,
	}
}

func cartItemFromUpdateRow(row pgsqlc.UpdateCartItemByIDRow) pgsqlc.CartItem {
	return pgsqlc.CartItem{
		ID:                    row.ID,
		CartID:                row.CartID,
		ProductID:             row.ProductID,
		Quantity:              row.Quantity,
		AddonIds:              row.AddonIds,
		AppliedDiscountID:     row.AppliedDiscountID,
		AppliedDiscountAmount: row.AppliedDiscountAmount,
	}
}

func cartItemFromGetRow(row pgsqlc.GetCartItemByIDRow) pgsqlc.CartItem {
	return pgsqlc.CartItem{
		ID:                    row.ID,
		CartID:                row.CartID,
		ProductID:             row.ProductID,
		Quantity:              row.Quantity,
		AddonIds:              row.AddonIds,
		AppliedDiscountID:     row.AppliedDiscountID,
		AppliedDiscountAmount: row.AppliedDiscountAmount,
	}
}

func cartItemFromListRow(row pgsqlc.ListCartItemsByCartRow) pgsqlc.CartItem {
	return pgsqlc.CartItem{
		ID:                    row.ID,
		CartID:                row.CartID,
		ProductID:             row.ProductID,
		Quantity:              row.Quantity,
		AddonIds:              row.AddonIds,
		AppliedDiscountID:     row.AppliedDiscountID,
		AppliedDiscountAmount: row.AppliedDiscountAmount,
	}
}

func (service *Service) CreateCart(ctx context.Context, cartID uuid.UUID, merchantID uuid.UUID, branchID uuid.UUID, actorID uuid.UUID) (pgsqlc.Cart, *domainerr.DomainError) {
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

func (service *Service) AddItemToCart(ctx context.Context, cartID uuid.UUID, productID uuid.UUID, quantity int32, addonIDs []uuid.UUID, discountID uuid.UUID, discountAmount float64) (pgsqlc.CartItem, *domainerr.DomainError) {
	if quantity <= 0 {
		return pgsqlc.CartItem{}, domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "quantity must be greater than zero", fmt.Errorf("quantity must be greater than zero"))
	}

	normalizedAddonIDs := normalizeAddonIDs(addonIDs)

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

	addonUnitSubtotal := 0.0
	for _, addonID := range normalizedAddonIDs {
		addon, addonErr := service.store.GetProductAddon(ctx, addonID)
		if addonErr != nil {
			return pgsqlc.CartItem{}, domainerr.MatchPostgresError(addonErr)
		}
		if addon.ProductID != product.ID {
			return pgsqlc.CartItem{}, domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "addon does not belong to product", fmt.Errorf("addon %s does not belong to product %s", addonID, productID))
		}
		addonUnitSubtotal += utils.NumericToFloat(addon.Price)
	}

	existingItem, existingErr := service.store.GetCartItemBySignature(ctx, pgsqlc.GetCartItemBySignatureParams{
		CartID:    cartID,
		ProductID: productID,
		AddonIds:  normalizedAddonIDs,
	})
	hasExistingItem := existingErr == nil
	var storedItem pgsqlc.CartItem
	if hasExistingItem {
		storedItem = cartItemFromSignatureRow(existingItem)
	}
	if existingErr != nil && existingErr != pgx.ErrNoRows {
		return pgsqlc.CartItem{}, domainerr.MatchPostgresError(existingErr)
	}

	mergedQuantity := quantity
	if hasExistingItem {
		mergedQuantity += storedItem.Quantity
	}

	lineSubtotal := utils.Round2((float64(mergedQuantity) * utils.NumericToFloat(product.BasePrice)) + (addonUnitSubtotal * float64(mergedQuantity)))
	appliedDiscountID := discountID
	var resolvedDiscount *pgsqlc.MerchantDiscount
	if discountID == uuid.Nil {
		resolved, resolveErr := service.resolveBestDiscountForProduct(ctx, cart.MerchantID, product.ID, product.CategoryID, time.Now())
		if resolveErr != nil {
			return pgsqlc.CartItem{}, resolveErr
		}
		resolvedDiscount = resolved
		if resolvedDiscount != nil {
			appliedDiscountID = resolvedDiscount.ID
		}
	}
	if appliedDiscountID == uuid.Nil {
		noDiscountID, noDiscountErr := service.ensureNoDiscountID(ctx, cartID)
		if noDiscountErr != nil {
			return pgsqlc.CartItem{}, domainerr.NewInternalError(noDiscountErr)
		}
		appliedDiscountID = noDiscountID
		discountAmount = 0
	} else {
		merchantDiscountRow, merchantDiscountErr := service.store.GetMerchantDiscount(ctx, pgsqlc.GetMerchantDiscountParams{
			MerchantID: cart.MerchantID,
			ID:         appliedDiscountID,
		})
		merchantDiscount := merchantDiscountFromGetRow(merchantDiscountRow)
		if resolvedDiscount != nil && resolvedDiscount.ID == appliedDiscountID {
			merchantDiscount = *resolvedDiscount
			merchantDiscountErr = nil
		}
		if merchantDiscountErr != nil {
			return pgsqlc.CartItem{}, domainerr.MatchPostgresError(merchantDiscountErr)
		}

		now := time.Now()
		if !isDiscountActive(merchantDiscount, now) && !merchantDiscount.ValidTo.IsZero() {
			if now.Before(merchantDiscount.ValidFrom) {
				return pgsqlc.CartItem{}, domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "discount is not active yet", fmt.Errorf("discount is not active yet"))
			}
			return pgsqlc.CartItem{}, domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "discount has expired", fmt.Errorf("discount has expired"))
		}

		if priority, matches := discountPriority(merchantDiscount, product.ID, product.CategoryID); priority < 1 || !matches {
			return pgsqlc.CartItem{}, domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "discount does not apply to product", fmt.Errorf("discount does not apply to product"))
		}

		discountAmount = calculateDiscountAmount(merchantDiscount, lineSubtotal)
	}
	if discountAmount < 0 || discountAmount > lineSubtotal {
		return pgsqlc.CartItem{}, domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "discount exceeds subtotal", fmt.Errorf("discount exceeds subtotal"))
	}

	if hasExistingItem {
		updatedItem, updateErr := service.store.UpdateCartItemByID(ctx, pgsqlc.UpdateCartItemByIDParams{
			CartID:                cartID,
			ID:                    storedItem.ID,
			Quantity:              mergedQuantity,
			AddonIds:              normalizedAddonIDs,
			AppliedDiscountID:     appliedDiscountID,
			AppliedDiscountAmount: utils.NumericFromFloat(discountAmount),
		})
		if updateErr != nil {
			return pgsqlc.CartItem{}, domainerr.MatchPostgresError(updateErr)
		}

		return cartItemFromUpdateRow(updatedItem), nil
	}

	createdItem, createErr := service.store.CreateCartItem(ctx, pgsqlc.CreateCartItemParams{
		CartID:                cartID,
		ProductID:             productID,
		Quantity:              mergedQuantity,
		AddonIds:              normalizedAddonIDs,
		AppliedDiscountID:     appliedDiscountID,
		AppliedDiscountAmount: utils.NumericFromFloat(discountAmount),
	})
	if createErr != nil {
		return pgsqlc.CartItem{}, domainerr.MatchPostgresError(createErr)
	}

	return cartItemFromCreateRow(createdItem), nil
}

func (service *Service) UpdateCartItemQuantity(ctx context.Context, cartID uuid.UUID, itemID uuid.UUID, quantity int32) (pgsqlc.CartItem, *domainerr.DomainError) {
	if quantity <= 0 {
		return pgsqlc.CartItem{}, domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "quantity must be greater than zero", fmt.Errorf("quantity must be greater than zero"))
	}

	cart, cartErr := service.getCartByID(ctx, cartID)
	if cartErr != nil {
		return pgsqlc.CartItem{}, cartErr
	}

	storedRow, storedErr := service.store.GetCartItemByID(ctx, pgsqlc.GetCartItemByIDParams{
		CartID: cartID,
		ID:     itemID,
	})
	if storedErr != nil {
		return pgsqlc.CartItem{}, domainerr.MatchPostgresError(storedErr)
	}
	storedItem := cartItemFromGetRow(storedRow)

	product, productErr := service.store.GetProduct(ctx, pgsqlc.GetProductParams{
		MerchantID: cart.MerchantID,
		ID:         storedItem.ProductID,
	})
	if productErr != nil {
		return pgsqlc.CartItem{}, domainerr.MatchPostgresError(productErr)
	}

	addonUnitSubtotal := 0.0
	for _, addonID := range storedItem.AddonIds {
		addon, addonErr := service.store.GetProductAddon(ctx, addonID)
		if addonErr != nil {
			return pgsqlc.CartItem{}, domainerr.MatchPostgresError(addonErr)
		}
		if addon.ProductID != product.ID {
			return pgsqlc.CartItem{}, domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "addon does not belong to product", fmt.Errorf("addon %s does not belong to product %s", addonID, product.ID))
		}
		addonUnitSubtotal += utils.NumericToFloat(addon.Price)
	}

	lineSubtotal := utils.Round2((float64(quantity) * utils.NumericToFloat(product.BasePrice)) + (addonUnitSubtotal * float64(quantity)))
	appliedDiscountID := storedItem.AppliedDiscountID
	discountAmount := 0.0

	if appliedDiscountID == uuid.Nil {
		noDiscountID, noDiscountErr := service.ensureNoDiscountID(ctx, cartID)
		if noDiscountErr != nil {
			return pgsqlc.CartItem{}, domainerr.NewInternalError(noDiscountErr)
		}
		appliedDiscountID = noDiscountID
	} else {
		merchantDiscountRow, merchantDiscountErr := service.store.GetMerchantDiscount(ctx, pgsqlc.GetMerchantDiscountParams{
			MerchantID: cart.MerchantID,
			ID:         appliedDiscountID,
		})
		if merchantDiscountErr != nil {
			return pgsqlc.CartItem{}, domainerr.MatchPostgresError(merchantDiscountErr)
		}
		merchantDiscount := merchantDiscountFromGetRow(merchantDiscountRow)

		now := time.Now()
		if !isDiscountActive(merchantDiscount, now) && !merchantDiscount.ValidTo.IsZero() {
			if now.Before(merchantDiscount.ValidFrom) {
				return pgsqlc.CartItem{}, domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "discount is not active yet", fmt.Errorf("discount is not active yet"))
			}
			return pgsqlc.CartItem{}, domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "discount has expired", fmt.Errorf("discount has expired"))
		}

		if priority, matches := discountPriority(merchantDiscount, product.ID, product.CategoryID); priority < 1 || !matches {
			return pgsqlc.CartItem{}, domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "discount does not apply to product", fmt.Errorf("discount does not apply to product"))
		}

		discountAmount = calculateDiscountAmount(merchantDiscount, lineSubtotal)
	}

	if discountAmount < 0 || discountAmount > lineSubtotal {
		return pgsqlc.CartItem{}, domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "discount exceeds subtotal", fmt.Errorf("discount exceeds subtotal"))
	}

	updatedItem, updateErr := service.store.UpdateCartItemByID(ctx, pgsqlc.UpdateCartItemByIDParams{
		CartID:                cartID,
		ID:                    itemID,
		Quantity:              quantity,
		AddonIds:              storedItem.AddonIds,
		AppliedDiscountID:     appliedDiscountID,
		AppliedDiscountAmount: utils.NumericFromFloat(discountAmount),
	})
	if updateErr != nil {
		return pgsqlc.CartItem{}, domainerr.MatchPostgresError(updateErr)
	}

	return cartItemFromUpdateRow(updatedItem), nil
}

func (service *Service) RemoveItemFromCart(ctx context.Context, cartID uuid.UUID, itemID uuid.UUID) *domainerr.DomainError {
	_, cartErr := service.getCartByID(ctx, cartID)
	if cartErr != nil {
		return cartErr
	}

	deletedRows, deleteErr := service.store.DeleteCartItem(ctx, pgsqlc.DeleteCartItemParams{
		CartID: cartID,
		ID:     itemID,
	})
	if deleteErr != nil {
		return domainerr.MatchPostgresError(deleteErr)
	}
	if deletedRows == 0 {
		return domainerr.NewDomainError(http.StatusNotFound, domainerr.NotFoundError, "cart item not found", pgx.ErrNoRows)
	}

	return nil
}

func (service *Service) getCartByID(ctx context.Context, cartID uuid.UUID) (pgsqlc.Cart, *domainerr.DomainError) {
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

func (service *Service) ensureNoDiscountID(ctx context.Context, cartID uuid.UUID) (uuid.UUID, error) {
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

func normalizeAddonIDs(addonIDs []uuid.UUID) []uuid.UUID {
	if len(addonIDs) == 0 {
		return []uuid.UUID{}
	}

	normalized := make([]uuid.UUID, 0, len(addonIDs))
	seen := make(map[uuid.UUID]struct{}, len(addonIDs))
	for _, addonID := range addonIDs {
		if addonID == uuid.Nil {
			continue
		}
		if _, ok := seen[addonID]; ok {
			continue
		}
		seen[addonID] = struct{}{}
		normalized = append(normalized, addonID)
	}

	sort.Slice(normalized, func(i int, j int) bool {
		return normalized[i].String() < normalized[j].String()
	})

	return normalized
}

func calculateDiscountAmount(discount pgsqlc.MerchantDiscount, subtotal float64) float64 {
	discountValue := utils.NumericToFloat(discount.Value)
	switch discount.Type {
	case pgsqlc.DiscountTypePercentage:
		return utils.Round2((subtotal * discountValue) / 100)
	case pgsqlc.DiscountTypeFlat:
		fallthrough
	default:
		if discountValue > subtotal {
			return utils.Round2(subtotal)
		}
		return utils.Round2(discountValue)
	}
}
