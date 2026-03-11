package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	commerce "github.com/horiondreher/go-web-api-boilerplate/internal/commerce"
	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
	orderdomain "github.com/horiondreher/go-web-api-boilerplate/internal/order"
	"github.com/horiondreher/go-web-api-boilerplate/internal/utils"
)

func (service *Service) GetPublicOrderDetail(ctx context.Context, orderID string) (orderdomain.Detail, *domainerr.DomainError) {
	parsedOrderID, parseErr := utils.ParseUUID(orderID, "order id")
	if parseErr != nil {
		return orderdomain.Detail{}, parseErr
	}

	order, orderErr := service.getOrderByID(ctx, parsedOrderID)
	if orderErr != nil {
		return orderdomain.Detail{}, orderErr
	}

	return service.buildOrderDetail(ctx, order)
}

func (service *Service) GetOrderDetail(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, orderID string) (orderdomain.Detail, *domainerr.DomainError) {
	parsedOrderID, parseErr := utils.ParseUUID(orderID, "order id")
	if parseErr != nil {
		return orderdomain.Detail{}, parseErr
	}

	viewer, viewerErr := service.resolveViewerActor(ctx, viewerMerchantID, viewerEmail)
	if viewerErr != nil {
		return orderdomain.Detail{}, viewerErr
	}

	isAdmin, adminErr := service.hasRole(ctx, viewer.UID, commerce.RoleTypeAdmin, uuid.Nil)
	if adminErr != nil {
		return orderdomain.Detail{}, domainerr.NewInternalError(adminErr)
	}

	var orderRow commerce.Order
	if isAdmin {
		orderByID, orderErr := service.getOrderByID(ctx, parsedOrderID)
		if orderErr != nil {
			return orderdomain.Detail{}, orderErr
		}
		orderRow = orderByID
	} else {
		allOrders, orderErr := service.store.ListOrdersByMerchant(ctx, viewerMerchantID)
		if orderErr != nil {
			return orderdomain.Detail{}, domainerr.MatchPostgresError(orderErr)
		}
		for _, candidate := range allOrders {
			if candidate.ID == parsedOrderID {
				orderRow = candidate
				break
			}
		}
		if orderRow.ID == uuid.Nil {
			return orderdomain.Detail{}, domainerr.NewDomainError(404, domainerr.NotFoundError, "not found", fmt.Errorf("order not found"))
		}
	}

	canViewMerchant, allowErr := service.canViewMerchant(ctx, viewer.UID, orderRow.MerchantID)
	if allowErr != nil {
		return orderdomain.Detail{}, domainerr.NewInternalError(allowErr)
	}
	orderActorMatchesViewer := orderRow.ActorID.Valid && uuid.UUID(orderRow.ActorID.Bytes) == viewer.UID
	if !canViewMerchant && !orderActorMatchesViewer {
		return orderdomain.Detail{}, domainerr.NewDomainError(403, domainerr.UnauthorizedError, "not allowed to view order", fmt.Errorf("not allowed to view order"))
	}

	return service.buildOrderDetail(ctx, orderRow)
}

func (service *Service) getOrderByID(ctx context.Context, orderID uuid.UUID) (commerce.Order, *domainerr.DomainError) {
	rows, queryErr := service.db.Query(ctx, `
        SELECT id, cart_id, merchant_id, branch_id, actor_id, payment_type, vat_rate, total_amount, status, delivery_address, customer_name, customer_phone, created_at, updated_at
        FROM orders
        WHERE id = $1
        LIMIT 1
    `, orderID)
	if queryErr != nil {
		return commerce.Order{}, domainerr.NewInternalError(queryErr)
	}
	defer rows.Close()

	var order commerce.Order
	if !rows.Next() {
		return commerce.Order{}, domainerr.NewDomainError(404, domainerr.NotFoundError, "not found", fmt.Errorf("order not found"))
	}
	if scanErr := rows.Scan(&order.ID, &order.CartID, &order.MerchantID, &order.BranchID, &order.ActorID, &order.PaymentType, &order.VatRate, &order.TotalAmount, &order.Status, &order.DeliveryAddress, &order.CustomerName, &order.CustomerPhone, &order.CreatedAt, &order.UpdatedAt); scanErr != nil {
		return commerce.Order{}, domainerr.NewInternalError(scanErr)
	}

	return order, nil
}

func (service *Service) buildOrderDetail(ctx context.Context, orderRow commerce.Order) (orderdomain.Detail, *domainerr.DomainError) {

	items, err := service.store.ListOrderItemsByOrder(ctx, orderRow.ID)
	if err != nil {
		return orderdomain.Detail{}, domainerr.MatchPostgresError(err)
	}
	allAddons, addonErr := service.store.ListOrderItemAddonsByOrder(ctx, orderRow.ID)
	if addonErr != nil {
		return orderdomain.Detail{}, domainerr.MatchPostgresError(addonErr)
	}

	itemDetails := make([]orderdomain.ItemDetail, 0, len(items))
	for _, item := range items {
		product, productErr := service.store.GetProduct(ctx, commerce.GetProductParams{
			MerchantID: orderRow.MerchantID,
			ID:         item.ProductID,
		})
		if productErr != nil {
			return orderdomain.Detail{}, domainerr.MatchPostgresError(productErr)
		}

		itemAddons := make([]commerce.OrderItemAddon, 0)
		for _, addon := range allAddons {
			if addon.ProductID == item.ProductID {
				itemAddons = append(itemAddons, addon)
			}
		}

		itemDetails = append(itemDetails, orderdomain.ItemDetail{
			Item:    item,
			Product: product,
			Addons:  itemAddons,
		})
	}

	return orderdomain.Detail{
		Order: orderRow,
		Items: itemDetails,
	}, nil
}

func (service *Service) ListOrdersByMerchant(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string) ([]commerce.Order, *domainerr.DomainError) {
	parsedMerchantID, parseErr := utils.ParseUUID(merchantID, "merchant id")
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
