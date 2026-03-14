package app

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/google/uuid"
	commercestore "github.com/horiondreher/go-web-api-boilerplate/internal/commerce/store"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	pgsqlc "github.com/horiondreher/go-web-api-boilerplate/internal/commerce"
	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
	orderdomain "github.com/horiondreher/go-web-api-boilerplate/internal/order"
	"github.com/horiondreher/go-web-api-boilerplate/internal/utils"
)

func defaultVATRate(paymentType orderdomain.PaymentType) (float64, *domainerr.DomainError) {
	switch paymentType {
	case pgsqlc.PaymentTypeCard:
		return 8, nil
	case pgsqlc.PaymentTypeCash:
		return 15, nil
	default:
		return 0, domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "invalid payment type", fmt.Errorf("invalid payment type %q", paymentType))
	}
}

func (service *Service) PlaceOrderFromCart(ctx context.Context, merchantActorID uuid.UUID, cartID uuid.UUID, paymentType orderdomain.PaymentType, deliveryAddress string, customerName string, customerPhone string) (orderdomain.Bill, *domainerr.DomainError) {
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

	branch, branchErr := service.store.GetBranch(ctx, pgsqlc.GetBranchParams{
		MerchantID: cart.MerchantID,
		ID:         cart.BranchID,
	})
	if branchErr != nil {
		return OrderBill{}, domainerr.MatchPostgresError(branchErr)
	}
	if !utils.IsBranchOpenAt(branch.OpeningTimeMinutes, branch.ClosingTimeMinutes, time.Now()) {
		return OrderBill{}, domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "branch is closed", fmt.Errorf("cannot place order: branch is not open"))
	}

	type addonSnapshot struct {
		addon    pgsqlc.ProductAddon
		quantity int32
		total    float64
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

	type aggregatedCartLine struct {
		productID       uuid.UUID
		quantity        int32
		discountAmount  float64
		addonQuantities map[uuid.UUID]int32
	}

	var bill orderdomain.Bill
	txErr := service.runInTx(ctx, func(tx pgx.Tx, store *commercestore.Postgres) *domainerr.DomainError {
		vatRate, vatErr := defaultVATRate(paymentType)
		if vatErr != nil {
			return vatErr
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
				discountAmount: utils.Round2(utils.NumericToFloat(discountNumeric)),
			})
		}

		if err := rows.Err(); err != nil {
			return domainerr.NewInternalError(err)
		}
		rows.Close()

		aggregatedLines := make([]aggregatedCartLine, 0, len(rawCartLines))
		aggregatedLineIndex := make(map[uuid.UUID]int, len(rawCartLines))
		for _, rawLine := range rawCartLines {
			index, ok := aggregatedLineIndex[rawLine.productID]
			if !ok {
				aggregatedLineIndex[rawLine.productID] = len(aggregatedLines)
				aggregatedLines = append(aggregatedLines, aggregatedCartLine{
					productID:       rawLine.productID,
					addonQuantities: make(map[uuid.UUID]int32),
				})
				index = len(aggregatedLines) - 1
			}

			aggregatedLines[index].quantity += rawLine.quantity
			aggregatedLines[index].discountAmount = utils.Round2(aggregatedLines[index].discountAmount + rawLine.discountAmount)
			for _, addonID := range rawLine.addonIDs {
				aggregatedLines[index].addonQuantities[addonID] += rawLine.quantity
			}
		}

		cartLines := make([]cartLine, 0, len(aggregatedLines))
		for _, aggregatedLine := range aggregatedLines {
			product, productErr := store.GetProduct(ctx, pgsqlc.GetProductParams{MerchantID: cart.MerchantID, ID: aggregatedLine.productID})
			if productErr != nil {
				return domainerr.MatchPostgresError(productErr)
			}

			line := cartLine{
				product:        product,
				quantity:       aggregatedLine.quantity,
				discountAmount: aggregatedLine.discountAmount,
				addons:         make([]addonSnapshot, 0, len(aggregatedLine.addonQuantities)),
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
				if inventory.Quantity < aggregatedLine.quantity {
					return domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "insufficient inventory", fmt.Errorf("insufficient inventory"))
				}
			}

			addonIDs := make([]uuid.UUID, 0, len(aggregatedLine.addonQuantities))
			for addonID := range aggregatedLine.addonQuantities {
				addonIDs = append(addonIDs, addonID)
			}
			sort.Slice(addonIDs, func(i int, j int) bool {
				return addonIDs[i].String() < addonIDs[j].String()
			})

			for _, addonID := range addonIDs {
				addon, addonErr := store.GetProductAddon(ctx, addonID)
				if addonErr != nil {
					return domainerr.MatchPostgresError(addonErr)
				}
				if addon.ProductID != product.ID {
					return domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "addon does not belong to product", fmt.Errorf("addon %s does not belong to product %s", addonID, product.ID))
				}

				addonQuantity := aggregatedLine.addonQuantities[addonID]
				lineAddOnTotal := utils.Round2(utils.NumericToFloat(addon.Price) * float64(addonQuantity))
				line.addons = append(line.addons, addonSnapshot{
					addon:    addon,
					quantity: addonQuantity,
					total:    lineAddOnTotal,
				})
			}

			cartLines = append(cartLines, line)
		}

		if len(cartLines) == 0 {
			return domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "cart has no items", fmt.Errorf("cart has no items"))
		}

		lineItems := make([]orderdomain.LineBill, 0, len(cartLines))
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
				VatRate:         utils.NumericFromFloat(vatRate),
				TotalAmount:     utils.NumericFromFloat(0),
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
				VatRate:         utils.NumericFromFloat(vatRate),
				TotalAmount:     utils.NumericFromFloat(0),
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
			unitBasePrice := utils.Round2(utils.NumericToFloat(line.product.BasePrice))
			baseAmount := utils.Round2(float64(line.quantity) * utils.NumericToFloat(line.product.BasePrice))
			addonAmount := 0.0
			for _, addon := range line.addons {
				addonAmount += addon.total
			}
			addonAmount = utils.Round2(addonAmount)

			if line.discountAmount < 0 {
				return domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "discount amount cannot be negative", fmt.Errorf("discount amount cannot be negative"))
			}

			lineSubtotal := utils.Round2(baseAmount + addonAmount - line.discountAmount)
			if lineSubtotal < 0 {
				return domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "discount exceeds line subtotal", fmt.Errorf("discount exceeds line subtotal"))
			}

			taxAmount := utils.Round2(lineSubtotal * (vatRate / 100.0))
			lineTotal := utils.Round2(lineSubtotal + taxAmount)

			lineBill := orderdomain.LineBill{
				ProductID:      line.product.ID,
				ProductName:    line.product.Name,
				BasePrice:      unitBasePrice,
				PaymentMethod:  string(paymentType),
				Quantity:       line.quantity,
				BaseAmount:     baseAmount,
				AddonAmount:    addonAmount,
				DiscountAmount: line.discountAmount,
				FinalPrice:     lineSubtotal,
				TaxAmount:      taxAmount,
				Vat:            taxAmount,
				LineTotal:      lineTotal,
				TotalPrice:     lineTotal,
			}
			lineItems = append(lineItems, lineBill)

			_, upsertErr := store.UpsertOrderItem(ctx, pgsqlc.UpsertOrderItemParams{
				OrderID:        createdOrder.ID,
				ProductID:      line.product.ID,
				Quantity:       line.quantity,
				Price:          utils.NumericFromFloat(utils.NumericToFloat(line.product.BasePrice)),
				BaseAmount:     utils.NumericFromFloat(baseAmount),
				AddonAmount:    utils.NumericFromFloat(addonAmount),
				DiscountAmount: utils.NumericFromFloat(line.discountAmount),
				TaxAmount:      utils.NumericFromFloat(taxAmount),
				LineTotal:      utils.NumericFromFloat(lineTotal),
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
					Quantity:       addon.quantity,
					LineAddonTotal: utils.NumericFromFloat(addon.total),
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
			TotalAmount:     utils.NumericFromFloat(utils.Round2(total)),
			Status:          createdOrder.Status,
			DeliveryAddress: createdOrder.DeliveryAddress,
			CustomerName:    createdOrder.CustomerName,
			CustomerPhone:   createdOrder.CustomerPhone,
		}); updateErr != nil {
			return domainerr.MatchPostgresError(updateErr)
		}

		bill = orderdomain.Bill{
			OrderID:     createdOrder.ID,
			PaymentType: string(paymentType),
			VatRate:     vatRate,
			Subtotal:    utils.Round2(subtotal),
			TotalTax:    utils.Round2(totalTax),
			Total:       utils.Round2(total),
			LineItems:   lineItems,
		}

		return nil
	})
	if txErr != nil {
		return orderdomain.Bill{}, txErr
	}

	return bill, nil
}

func (service *Service) UpdateOrderStatus(ctx context.Context, merchantActorID uuid.UUID, merchantID uuid.UUID, orderID uuid.UUID, status pgsqlc.OrderStatusType) (pgsqlc.Order, *domainerr.DomainError) {
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
