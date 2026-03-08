package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/pgsqlc"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/ports"
)

func (service *CommerceManager) GetOrderDetail(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, orderID string) (ports.OrderDetail, *domainerr.DomainError) {
	parsedOrderID, parseErr := parseUUID(orderID, "order id")
	if parseErr != nil {
		return ports.OrderDetail{}, parseErr
	}

	viewer, viewerErr := service.resolveViewerActor(ctx, viewerMerchantID, viewerEmail)
	if viewerErr != nil {
		return ports.OrderDetail{}, viewerErr
	}

	isAdmin, adminErr := service.hasRole(ctx, viewer.UID, pgsqlc.RoleTypeAdmin, uuid.Nil)
	if adminErr != nil {
		return ports.OrderDetail{}, domainerr.NewInternalError(adminErr)
	}

	var order pgsqlc.Order
	if isAdmin {
		rows, queryErr := service.db.Query(ctx, `
        SELECT id, cart_id, merchant_id, branch_id, actor_id, payment_type, vat_rate, total_amount, status, delivery_address, customer_name, customer_phone, created_at, updated_at
        FROM orders
        WHERE id = $1
        LIMIT 1
    `, parsedOrderID)
		if queryErr != nil {
			return ports.OrderDetail{}, domainerr.NewInternalError(queryErr)
		}
		defer rows.Close()

		if !rows.Next() {
			return ports.OrderDetail{}, domainerr.NewDomainError(404, domainerr.NotFoundError, "not found", fmt.Errorf("order not found"))
		}
		if scanErr := rows.Scan(&order.ID, &order.CartID, &order.MerchantID, &order.BranchID, &order.ActorID, &order.PaymentType, &order.VatRate, &order.TotalAmount, &order.Status, &order.DeliveryAddress, &order.CustomerName, &order.CustomerPhone, &order.CreatedAt, &order.UpdatedAt); scanErr != nil {
			return ports.OrderDetail{}, domainerr.NewInternalError(scanErr)
		}
	} else {
		allOrders, orderErr := service.store.ListOrdersByMerchant(ctx, viewerMerchantID)
		if orderErr != nil {
			return ports.OrderDetail{}, domainerr.MatchPostgresError(orderErr)
		}
		for _, candidate := range allOrders {
			if candidate.ID == parsedOrderID {
				order = candidate
				break
			}
		}
		if order.ID == uuid.Nil {
			return ports.OrderDetail{}, domainerr.NewDomainError(404, domainerr.NotFoundError, "not found", fmt.Errorf("order not found"))
		}
	}

	canViewMerchant, allowErr := service.canViewMerchant(ctx, viewer.UID, order.MerchantID)
	if allowErr != nil {
		return ports.OrderDetail{}, domainerr.NewInternalError(allowErr)
	}
	orderActorMatchesViewer := order.ActorID.Valid && uuid.UUID(order.ActorID.Bytes) == viewer.UID
	if !canViewMerchant && !orderActorMatchesViewer {
		return ports.OrderDetail{}, domainerr.NewDomainError(403, domainerr.UnauthorizedError, "not allowed to view order", fmt.Errorf("not allowed to view order"))
	}

	items, err := service.store.ListOrderItemsByOrder(ctx, order.ID)
	if err != nil {
		return ports.OrderDetail{}, domainerr.MatchPostgresError(err)
	}
	allAddons, addonErr := service.store.ListOrderItemAddonsByOrder(ctx, order.ID)
	if addonErr != nil {
		return ports.OrderDetail{}, domainerr.MatchPostgresError(addonErr)
	}

	itemDetails := make([]ports.OrderItemDetail, 0, len(items))
	for _, item := range items {
		product, productErr := service.store.GetProduct(ctx, pgsqlc.GetProductParams{
			MerchantID: order.MerchantID,
			ID:         item.ProductID,
		})
		if productErr != nil {
			return ports.OrderDetail{}, domainerr.MatchPostgresError(productErr)
		}

		itemAddons := make([]pgsqlc.OrderItemAddon, 0)
		for _, addon := range allAddons {
			if addon.ProductID == item.ProductID {
				itemAddons = append(itemAddons, addon)
			}
		}

		itemDetails = append(itemDetails, ports.OrderItemDetail{
			Item:    item,
			Product: product,
			Addons:  itemAddons,
		})
	}

	return ports.OrderDetail{
		Order: order,
		Items: itemDetails,
	}, nil
}

func (service *CommerceManager) ListOrdersByMerchant(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string) ([]pgsqlc.Order, *domainerr.DomainError) {
	parsedMerchantID, parseErr := parseUUID(merchantID, "merchant id")
	if parseErr != nil {
		return nil, parseErr
	}

	if _, accessErr := service.requireMerchantViewAccess(ctx, viewerMerchantID, viewerEmail, parsedMerchantID); accessErr != nil {
		return nil, accessErr
	}

	orders, err := service.store.ListOrdersByMerchant(ctx, parsedMerchantID)
	if err != nil {
		return nil, domainerr.MatchPostgresError(err)
	}

	return orders, nil
}
