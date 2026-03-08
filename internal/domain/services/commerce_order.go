package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/pgsqlc"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (service *CommerceManager) PlaceOrderFromCart(ctx context.Context, merchantActorID uuid.UUID, cartID uuid.UUID, paymentType pgsqlc.PaymentType, deliveryAddress string, customerName string, customerPhone string) (OrderBill, *domainerr.DomainError) {
	var cart pgsqlc.Cart
	cartErr := service.db.QueryRow(ctx, `
		SELECT id, merchant_id, branch_id, actor_id, created_at, updated_at
		FROM carts
		WHERE id = $1
	`, cartID).Scan(&cart.ID, &cart.MerchantID, &cart.BranchID, &cart.ActorID, &cart.CreatedAt, &cart.UpdatedAt)
	if cartErr != nil {
		return OrderBill{}, domainerr.MatchPostgresError(cartErr)
	}

	isGuestCart := !cart.ActorID.Valid
	if !isGuestCart {
		isMerchant, merchantErr := service.hasRole(ctx, merchantActorID, pgsqlc.RoleTypeMerchant, cart.MerchantID)
		if merchantErr != nil {
			return OrderBill{}, domainerr.NewInternalError(merchantErr)
		}
		if !isMerchant {
			isCustomer, customerErr := service.hasRole(ctx, merchantActorID, pgsqlc.RoleTypeCustomer, cart.MerchantID)
			if customerErr != nil {
				return OrderBill{}, domainerr.NewInternalError(customerErr)
			}
			cartActorID := uuid.UUID(cart.ActorID.Bytes)
			if !isCustomer || cartActorID != merchantActorID {
				return OrderBill{}, domainerr.NewDomainError(http.StatusForbidden, domainerr.UnauthorizedError, "merchant or cart customer role required", fmt.Errorf("merchant or cart customer role required"))
			}
		}
	}

	type addonSnapshot struct {
		addon pgsqlc.ProductAddon
		total float64
	}

	type cartLine struct {
		product        pgsqlc.Product
		quantity       int32
		discountAmount float64
		addons         []addonSnapshot
		availableStock int32
		trackInventory bool
	}

	type rawCartLine struct {
		productID      uuid.UUID
		quantity       int32
		addonIDs       []uuid.UUID
		discountAmount float64
	}

	var bill OrderBill
	txErr := service.runInTx(ctx, func(tx pgx.Tx, store *pgsqlc.Queries) *domainerr.DomainError {
		vatRate := 0.0
		vatRule, vatErr := store.GetVatRule(ctx, pgsqlc.GetVatRuleParams{MerchantID: cart.MerchantID, PaymentType: paymentType})
		if vatErr == nil {
			vatRate = numericToFloat(vatRule.Rate)
		}

		rows, rowsErr := tx.Query(ctx, `
        SELECT product_id, quantity, addon_ids, COALESCE(applied_discount_amount, 0)
        FROM cart_items
        WHERE cart_id = $1
    `, cartID)
		if rowsErr != nil {
			return domainerr.NewInternalError(rowsErr)
		}
		defer rows.Close()

		rawCartLines := make([]rawCartLine, 0)
		for rows.Next() {
			var productID uuid.UUID
			var quantity int32
			var addonIDs []uuid.UUID
			var discountNumeric pgtype.Numeric
			if scanErr := rows.Scan(&productID, &quantity, &addonIDs, &discountNumeric); scanErr != nil {
				return domainerr.NewInternalError(scanErr)
			}
			if quantity <= 0 {
				return domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "quantity must be greater than zero", fmt.Errorf("quantity must be greater than zero"))
			}

			rawCartLines = append(rawCartLines, rawCartLine{
				productID:      productID,
				quantity:       quantity,
				addonIDs:       addonIDs,
				discountAmount: round2(numericToFloat(discountNumeric)),
			})
		}

		if err := rows.Err(); err != nil {
			return domainerr.NewInternalError(err)
		}
		rows.Close()

		cartLines := make([]cartLine, 0, len(rawCartLines))
		for _, rawLine := range rawCartLines {
			product, productErr := store.GetProduct(ctx, pgsqlc.GetProductParams{MerchantID: cart.MerchantID, ID: rawLine.productID})
			if productErr != nil {
				return domainerr.MatchPostgresError(productErr)
			}

			line := cartLine{
				product:        product,
				quantity:       rawLine.quantity,
				discountAmount: rawLine.discountAmount,
				addons:         make([]addonSnapshot, 0, len(rawLine.addonIDs)),
				trackInventory: product.TrackInventory,
			}

			if product.TrackInventory && cart.BranchID != uuid.Nil {
				inventory, inventoryErr := store.GetProductInventory(ctx, pgsqlc.GetProductInventoryParams{
					ProductID: product.ID,
					BranchID:  cart.BranchID,
				})
				if inventoryErr != nil {
					return domainerr.MatchPostgresError(inventoryErr)
				}
				line.availableStock = inventory.Quantity
				if inventory.Quantity < rawLine.quantity {
					return domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "insufficient inventory", fmt.Errorf("insufficient inventory"))
				}
			}

			for _, addonID := range rawLine.addonIDs {
				addon, addonErr := store.GetProductAddon(ctx, addonID)
				if addonErr != nil {
					return domainerr.MatchPostgresError(addonErr)
				}
				if addon.ProductID != product.ID {
					return domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "addon does not belong to product", fmt.Errorf("addon %s does not belong to product %s", addonID, product.ID))
				}

				lineAddOnTotal := round2(numericToFloat(addon.Price) * float64(rawLine.quantity))
				line.addons = append(line.addons, addonSnapshot{
					addon: addon,
					total: lineAddOnTotal,
				})
			}

			cartLines = append(cartLines, line)
		}

		if len(cartLines) == 0 {
			return domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "cart has no items", fmt.Errorf("cart has no items"))
		}

		lineItems := make([]OrderLineBill, 0, len(cartLines))
		total := 0.0
		totalTax := 0.0
		subtotal := 0.0

		var createdOrder pgsqlc.Order
		if isGuestCart {
			order, createErr := store.CreateOrderGuest(ctx, pgsqlc.CreateOrderGuestParams{
				CartID:          cart.ID,
				MerchantID:      cart.MerchantID,
				BranchID:        cart.BranchID,
				PaymentType:     paymentType,
				VatRate:         numericFromFloat(vatRate),
				TotalAmount:     numericFromFloat(0),
				Status:          pgsqlc.OrderStatusTypePending,
				DeliveryAddress: deliveryAddress,
				CustomerName:    customerName,
				CustomerPhone:   customerPhone,
			})
			if createErr != nil {
				return domainerr.MatchPostgresError(createErr)
			}
			createdOrder = order
		} else {
			order, createErr := store.CreateOrder(ctx, pgsqlc.CreateOrderParams{
				CartID:          cart.ID,
				MerchantID:      cart.MerchantID,
				BranchID:        cart.BranchID,
				ActorID:         cart.ActorID,
				PaymentType:     paymentType,
				VatRate:         numericFromFloat(vatRate),
				TotalAmount:     numericFromFloat(0),
				Status:          pgsqlc.OrderStatusTypePending,
				DeliveryAddress: deliveryAddress,
				CustomerName:    customerName,
				CustomerPhone:   customerPhone,
			})
			if createErr != nil {
				return domainerr.MatchPostgresError(createErr)
			}
			createdOrder = order
		}

		for _, line := range cartLines {
			baseAmount := round2(float64(line.quantity) * numericToFloat(line.product.BasePrice))
			addonAmount := 0.0
			for _, addon := range line.addons {
				addonAmount += addon.total
			}
			addonAmount = round2(addonAmount)

			if line.discountAmount < 0 {
				return domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "discount amount cannot be negative", fmt.Errorf("discount amount cannot be negative"))
			}

			lineSubtotal := round2(baseAmount + addonAmount - line.discountAmount)
			if lineSubtotal < 0 {
				return domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "discount exceeds line subtotal", fmt.Errorf("discount exceeds line subtotal"))
			}

			taxableAmount := round2(baseAmount + addonAmount)
			taxAmount := round2(taxableAmount * (vatRate / 100.0))
			lineTotal := round2(lineSubtotal + taxAmount)

			lineBill := OrderLineBill{
				ProductID:      line.product.ID,
				Quantity:       line.quantity,
				BaseAmount:     baseAmount,
				AddonAmount:    addonAmount,
				DiscountAmount: line.discountAmount,
				TaxAmount:      taxAmount,
				LineTotal:      lineTotal,
			}
			lineItems = append(lineItems, lineBill)

			_, upsertErr := store.UpsertOrderItem(ctx, pgsqlc.UpsertOrderItemParams{
				OrderID:        createdOrder.ID,
				ProductID:      line.product.ID,
				Quantity:       line.quantity,
				Price:          numericFromFloat(numericToFloat(line.product.BasePrice)),
				BaseAmount:     numericFromFloat(baseAmount),
				AddonAmount:    numericFromFloat(addonAmount),
				DiscountAmount: numericFromFloat(line.discountAmount),
				TaxAmount:      numericFromFloat(taxAmount),
				LineTotal:      numericFromFloat(lineTotal),
			})
			if upsertErr != nil {
				return domainerr.MatchPostgresError(upsertErr)
			}

			for _, addon := range line.addons {
				_, addonErr := store.UpsertOrderItemAddon(ctx, pgsqlc.UpsertOrderItemAddonParams{
					OrderID:        createdOrder.ID,
					ProductID:      line.product.ID,
					AddonID:        addon.addon.ID,
					AddonName:      addon.addon.Name,
					AddonPrice:     addon.addon.Price,
					Quantity:       line.quantity,
					LineAddonTotal: numericFromFloat(addon.total),
				})
				if addonErr != nil {
					return domainerr.MatchPostgresError(addonErr)
				}
			}

			if line.trackInventory && cart.BranchID != uuid.Nil {
				updatedQuantity := line.availableStock - line.quantity
				if _, inventoryErr := store.UpdateProductInventoryQuantity(ctx, pgsqlc.UpdateProductInventoryQuantityParams{
					ProductID: line.product.ID,
					BranchID:  cart.BranchID,
					Quantity:  updatedQuantity,
				}); inventoryErr != nil {
					return domainerr.MatchPostgresError(inventoryErr)
				}
			}

			subtotal += lineSubtotal
			totalTax += taxAmount
			total += lineTotal
		}

		if _, updateErr := store.UpdateOrder(ctx, pgsqlc.UpdateOrderParams{
			MerchantID:      createdOrder.MerchantID,
			ID:              createdOrder.ID,
			BranchID:        createdOrder.BranchID,
			ActorID:         createdOrder.ActorID,
			PaymentType:     createdOrder.PaymentType,
			VatRate:         createdOrder.VatRate,
			TotalAmount:     numericFromFloat(round2(total)),
			Status:          createdOrder.Status,
			DeliveryAddress: createdOrder.DeliveryAddress,
			CustomerName:    createdOrder.CustomerName,
			CustomerPhone:   createdOrder.CustomerPhone,
		}); updateErr != nil {
			return domainerr.MatchPostgresError(updateErr)
		}

		bill = OrderBill{
			OrderID:   createdOrder.ID,
			VatRate:   vatRate,
			Subtotal:  round2(subtotal),
			TotalTax:  round2(totalTax),
			Total:     round2(total),
			LineItems: lineItems,
		}

		return nil
	})
	if txErr != nil {
		return OrderBill{}, txErr
	}

	return bill, nil
}

func (service *CommerceManager) UpdateOrderStatus(ctx context.Context, merchantActorID uuid.UUID, merchantID uuid.UUID, orderID uuid.UUID, status pgsqlc.OrderStatusType) (pgsqlc.Order, *domainerr.DomainError) {
	allowed, allowErr := service.hasRole(ctx, merchantActorID, pgsqlc.RoleTypeMerchant, merchantID)
	if allowErr != nil {
		return pgsqlc.Order{}, domainerr.NewInternalError(allowErr)
	}
	if !allowed {
		return pgsqlc.Order{}, domainerr.NewDomainError(http.StatusForbidden, domainerr.UnauthorizedError, "merchant role required", fmt.Errorf("merchant role required"))
	}

	existingOrder, getErr := service.store.GetOrder(ctx, pgsqlc.GetOrderParams{MerchantID: merchantID, ID: orderID})
	if getErr != nil {
		return pgsqlc.Order{}, domainerr.MatchPostgresError(getErr)
	}
	if !isValidOrderStatusTransition(existingOrder.Status, status) {
		return pgsqlc.Order{}, domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "invalid order status transition", fmt.Errorf("cannot move order from %s to %s", existingOrder.Status, status))
	}

	updatedOrder, updateErr := service.store.UpdateOrder(ctx, pgsqlc.UpdateOrderParams{
		MerchantID:      existingOrder.MerchantID,
		ID:              existingOrder.ID,
		BranchID:        existingOrder.BranchID,
		ActorID:         existingOrder.ActorID,
		PaymentType:     existingOrder.PaymentType,
		VatRate:         existingOrder.VatRate,
		TotalAmount:     existingOrder.TotalAmount,
		Status:          status,
		DeliveryAddress: existingOrder.DeliveryAddress,
		CustomerName:    existingOrder.CustomerName,
		CustomerPhone:   existingOrder.CustomerPhone,
	})
	if updateErr != nil {
		return pgsqlc.Order{}, domainerr.MatchPostgresError(updateErr)
	}

	return updatedOrder, nil
}

func isValidOrderStatusTransition(from pgsqlc.OrderStatusType, to pgsqlc.OrderStatusType) bool {
	allowedTransitions := map[pgsqlc.OrderStatusType]map[pgsqlc.OrderStatusType]bool{
		pgsqlc.OrderStatusTypePending: {
			pgsqlc.OrderStatusTypeAccepted:  true,
			pgsqlc.OrderStatusTypeCancelled: true,
		},
		pgsqlc.OrderStatusTypeAccepted: {
			pgsqlc.OrderStatusTypeOutForDelivery: true,
			pgsqlc.OrderStatusTypeCancelled:      true,
		},
		pgsqlc.OrderStatusTypeOutForDelivery: {
			pgsqlc.OrderStatusTypeDelivered: true,
			pgsqlc.OrderStatusTypeCancelled: true,
		},
		pgsqlc.OrderStatusTypeDelivered: {
			pgsqlc.OrderStatusTypeRefunded: true,
		},
	}

	return allowedTransitions[from][to]
}
