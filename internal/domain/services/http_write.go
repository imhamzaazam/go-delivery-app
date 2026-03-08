package services

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/pgsqlc"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/ports"
	"github.com/horiondreher/go-web-api-boilerplate/internal/utils"
	"github.com/jackc/pgx/v5"
)

func (service *CommerceManager) CreateActorByMerchant(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, newActor ports.NewActor, role string) (pgsqlc.CreateActorRow, *domainerr.DomainError) {
	parsedMerchantID, parseErr := parseUUID(merchantID, "merchant id")
	if parseErr != nil {
		return pgsqlc.CreateActorRow{}, parseErr
	}
	if parsedMerchantID != newActor.MerchantID {
		return pgsqlc.CreateActorRow{}, domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "merchant mismatch for actor creation", fmt.Errorf("merchant mismatch for actor creation"))
	}

	_, accessErr := service.requireMerchantManageAccess(ctx, viewerMerchantID, viewerEmail, newActor.MerchantID)
	if accessErr != nil {
		return pgsqlc.CreateActorRow{}, accessErr
	}

	roleType := pgsqlc.RoleType(strings.ToLower(strings.TrimSpace(role)))
	if roleType == "" {
		roleType = pgsqlc.RoleTypeCustomer
	}
	switch roleType {
	case pgsqlc.RoleTypeEmployee, pgsqlc.RoleTypeCustomer, pgsqlc.RoleTypeMerchant:
	default:
		return pgsqlc.CreateActorRow{}, domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "invalid actor role", fmt.Errorf("invalid actor role"))
	}

	hashedPassword, hashErr := utils.HashPasswordOrNoop(newActor.Password)
	if hashErr != nil {
		return pgsqlc.CreateActorRow{}, hashErr
	}

	firstName, lastName := splitFullName(newActor.FullName)
	var createdActor pgsqlc.CreateActorRow
	txErr := service.runInTx(ctx, func(tx pgx.Tx, store *pgsqlc.Queries) *domainerr.DomainError {
		for _, defaultRole := range []pgsqlc.RoleType{
			pgsqlc.RoleTypeMerchant,
			pgsqlc.RoleTypeEmployee,
			pgsqlc.RoleTypeCustomer,
		} {
			if _, roleErr := ensureMerchantRole(ctx, tx, store, newActor.MerchantID, defaultRole); roleErr != nil {
				return roleErr
			}
		}

		actor, createErr := store.CreateActor(ctx, pgsqlc.CreateActorParams{
			MerchantID:   newActor.MerchantID,
			Email:        newActor.Email,
			PasswordHash: hashedPassword,
			FirstName:    firstName,
			LastName:     lastName,
			IsActive:     true,
			LastLogin:    time.Now(),
		})
		if createErr != nil {
			return domainerr.MatchPostgresError(createErr)
		}
		createdActor = actor

		roleID, roleErr := ensureMerchantRole(ctx, tx, store, newActor.MerchantID, roleType)
		if roleErr != nil {
			return roleErr
		}
		if _, assignErr := store.AssignActorRole(ctx, pgsqlc.AssignActorRoleParams{
			MerchantID: newActor.MerchantID,
			ActorID:    createdActor.UID,
			RoleID:     roleID,
		}); assignErr != nil {
			return domainerr.MatchPostgresError(assignErr)
		}

		return nil
	})
	if txErr != nil {
		return pgsqlc.CreateActorRow{}, txErr
	}

	return createdActor, nil
}

func (service *CommerceManager) CreateBranchByMerchant(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, name string, address string, contactNumber string, city string) (pgsqlc.Branch, *domainerr.DomainError) {
	parsedMerchantID, parseErr := parseUUID(merchantID, "merchant id")
	if parseErr != nil {
		return pgsqlc.Branch{}, parseErr
	}
	if _, accessErr := service.requireMerchantManageAccess(ctx, viewerMerchantID, viewerEmail, parsedMerchantID); accessErr != nil {
		return pgsqlc.Branch{}, accessErr
	}

	createdBranch, createErr := service.store.CreateBranch(ctx, pgsqlc.CreateBranchParams{
		MerchantID:    parsedMerchantID,
		Name:          name,
		Address:       address,
		ContactNumber: textValue(contactNumber),
		City:          pgsqlc.CityType(city),
	})
	if createErr != nil {
		return pgsqlc.Branch{}, domainerr.MatchPostgresError(createErr)
	}

	return createdBranch, nil
}

func (service *CommerceManager) CreateMerchantServiceZoneByMerchant(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, zoneID string, branchID string) (pgsqlc.MerchantServiceZone, *domainerr.DomainError) {
	parsedMerchantID, merchantErr := parseUUID(merchantID, "merchant id")
	if merchantErr != nil {
		return pgsqlc.MerchantServiceZone{}, merchantErr
	}
	if _, accessErr := service.requireMerchantManageAccess(ctx, viewerMerchantID, viewerEmail, parsedMerchantID); accessErr != nil {
		return pgsqlc.MerchantServiceZone{}, accessErr
	}

	parsedZoneID, zoneErr := parseUUID(zoneID, "zone id")
	if zoneErr != nil {
		return pgsqlc.MerchantServiceZone{}, zoneErr
	}
	parsedBranchID, branchErr := parseUUID(branchID, "branch id")
	if branchErr != nil {
		return pgsqlc.MerchantServiceZone{}, branchErr
	}
	if _, getZoneErr := service.store.GetZone(ctx, parsedZoneID); getZoneErr != nil {
		return pgsqlc.MerchantServiceZone{}, domainerr.MatchPostgresError(getZoneErr)
	}
	branch, getBranchErr := service.store.GetBranch(ctx, pgsqlc.GetBranchParams{
		MerchantID: parsedMerchantID,
		ID:         parsedBranchID,
	})
	if getBranchErr != nil {
		return pgsqlc.MerchantServiceZone{}, domainerr.MatchPostgresError(getBranchErr)
	}
	if branch.MerchantID != parsedMerchantID {
		return pgsqlc.MerchantServiceZone{}, domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "branch does not belong to merchant", fmt.Errorf("branch does not belong to merchant"))
	}

	createdZone, createErr := service.store.CreateMerchantServiceZone(ctx, pgsqlc.CreateMerchantServiceZoneParams{
		MerchantID: parsedMerchantID,
		ZoneID:     parsedZoneID,
		BranchID:   parsedBranchID,
	})
	if createErr != nil {
		return pgsqlc.MerchantServiceZone{}, domainerr.MatchPostgresError(createErr)
	}

	return createdZone, nil
}

func (service *CommerceManager) CreateProductCategoryByMerchant(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, name string, description string) (pgsqlc.ProductCategory, *domainerr.DomainError) {
	parsedMerchantID, parseErr := parseUUID(merchantID, "merchant id")
	if parseErr != nil {
		return pgsqlc.ProductCategory{}, parseErr
	}
	if _, accessErr := service.requireMerchantManageAccess(ctx, viewerMerchantID, viewerEmail, parsedMerchantID); accessErr != nil {
		return pgsqlc.ProductCategory{}, accessErr
	}

	category, createErr := service.store.CreateProductCategory(ctx, pgsqlc.CreateProductCategoryParams{
		MerchantID:  parsedMerchantID,
		Name:        name,
		Description: textValue(description),
	})
	if createErr != nil {
		return pgsqlc.ProductCategory{}, domainerr.MatchPostgresError(createErr)
	}

	return category, nil
}

func (service *CommerceManager) CreateProductByMerchantHTTP(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, categoryID string, name string, description string, basePrice float64, imageURL string, trackInventory bool) (pgsqlc.Product, *domainerr.DomainError) {
	parsedMerchantID, merchantErr := parseUUID(merchantID, "merchant id")
	if merchantErr != nil {
		return pgsqlc.Product{}, merchantErr
	}
	viewer, accessErr := service.requireMerchantManageAccess(ctx, viewerMerchantID, viewerEmail, parsedMerchantID)
	if accessErr != nil {
		return pgsqlc.Product{}, accessErr
	}
	parsedCategoryID, categoryErr := parseUUID(categoryID, "category id")
	if categoryErr != nil {
		return pgsqlc.Product{}, categoryErr
	}

	return service.CreateProductByMerchant(ctx, viewer.UID, parsedMerchantID, parsedCategoryID, name, description, basePrice, imageURL, trackInventory)
}

func (service *CommerceManager) AddProductAddonByMerchantHTTP(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, productID string, name string, price float64) (pgsqlc.ProductAddon, *domainerr.DomainError) {
	parsedMerchantID, merchantErr := parseUUID(merchantID, "merchant id")
	if merchantErr != nil {
		return pgsqlc.ProductAddon{}, merchantErr
	}
	viewer, accessErr := service.requireMerchantManageAccess(ctx, viewerMerchantID, viewerEmail, parsedMerchantID)
	if accessErr != nil {
		return pgsqlc.ProductAddon{}, accessErr
	}
	parsedProductID, productErr := parseUUID(productID, "product id")
	if productErr != nil {
		return pgsqlc.ProductAddon{}, productErr
	}

	return service.AddProductAddonByMerchant(ctx, viewer.UID, parsedMerchantID, parsedProductID, name, price)
}

func (service *CommerceManager) CreateMerchantDiscountByMerchant(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, discountType string, value float64, description string) (pgsqlc.MerchantDiscount, *domainerr.DomainError) {
	parsedMerchantID, parseErr := parseUUID(merchantID, "merchant id")
	if parseErr != nil {
		return pgsqlc.MerchantDiscount{}, parseErr
	}
	if _, accessErr := service.requireMerchantManageAccess(ctx, viewerMerchantID, viewerEmail, parsedMerchantID); accessErr != nil {
		return pgsqlc.MerchantDiscount{}, accessErr
	}

	createdDiscount, createErr := service.store.CreateMerchantDiscount(ctx, pgsqlc.CreateMerchantDiscountParams{
		MerchantID:  parsedMerchantID,
		Type:        pgsqlc.DiscountType(strings.ToLower(strings.TrimSpace(discountType))),
		Value:       numericFromFloat(value),
		Description: textValue(description),
		ValidFrom:   time.Now().UTC(),
		ValidTo:     time.Now().UTC().Add(30 * 24 * time.Hour),
	})
	if createErr != nil {
		return pgsqlc.MerchantDiscount{}, domainerr.MatchPostgresError(createErr)
	}

	return createdDiscount, nil
}

func (service *CommerceManager) UpsertInventoryByMerchant(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, productID string, branchID string, quantity int32) (pgsqlc.ProductInventory, *domainerr.DomainError) {
	parsedMerchantID, merchantErr := parseUUID(merchantID, "merchant id")
	if merchantErr != nil {
		return pgsqlc.ProductInventory{}, merchantErr
	}
	if _, accessErr := service.requireMerchantManageAccess(ctx, viewerMerchantID, viewerEmail, parsedMerchantID); accessErr != nil {
		return pgsqlc.ProductInventory{}, accessErr
	}

	parsedProductID, productErr := parseUUID(productID, "product id")
	if productErr != nil {
		return pgsqlc.ProductInventory{}, productErr
	}
	parsedBranchID, branchErr := parseUUID(branchID, "branch id")
	if branchErr != nil {
		return pgsqlc.ProductInventory{}, branchErr
	}
	product, getProductErr := service.store.GetProduct(ctx, pgsqlc.GetProductParams{
		MerchantID: parsedMerchantID,
		ID:         parsedProductID,
	})
	if getProductErr != nil {
		return pgsqlc.ProductInventory{}, domainerr.MatchPostgresError(getProductErr)
	}
	branch, getBranchErr := service.store.GetBranch(ctx, pgsqlc.GetBranchParams{
		MerchantID: parsedMerchantID,
		ID:         parsedBranchID,
	})
	if getBranchErr != nil {
		return pgsqlc.ProductInventory{}, domainerr.MatchPostgresError(getBranchErr)
	}
	if product.MerchantID != parsedMerchantID || branch.MerchantID != parsedMerchantID {
		return pgsqlc.ProductInventory{}, domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "inventory references invalid merchant resources", fmt.Errorf("inventory references invalid merchant resources"))
	}

	inventory, upsertErr := service.store.UpsertProductInventory(ctx, pgsqlc.UpsertProductInventoryParams{
		ProductID: parsedProductID,
		BranchID:  parsedBranchID,
		Quantity:  quantity,
	})
	if upsertErr != nil {
		return pgsqlc.ProductInventory{}, domainerr.MatchPostgresError(upsertErr)
	}

	return inventory, nil
}

func (service *CommerceManager) CreateCartHTTP(ctx context.Context, merchantID string, branchID string, actorID string, cartID string) (pgsqlc.Cart, *domainerr.DomainError) {
	parsedMerchantID, merchantErr := parseUUID(merchantID, "merchant id")
	if merchantErr != nil {
		return pgsqlc.Cart{}, merchantErr
	}
	parsedBranchID, branchErr := parseUUID(branchID, "branch id")
	if branchErr != nil {
		return pgsqlc.Cart{}, branchErr
	}
	parsedCartID, cartErr := parseUUID(cartID, "cart id")
	if cartErr != nil {
		return pgsqlc.Cart{}, cartErr
	}

	if _, getBranchErr := service.store.GetBranch(ctx, pgsqlc.GetBranchParams{
		MerchantID: parsedMerchantID,
		ID:         parsedBranchID,
	}); getBranchErr != nil {
		return pgsqlc.Cart{}, domainerr.MatchPostgresError(getBranchErr)
	}

	if strings.TrimSpace(actorID) == "" {
		return service.CreateCart(ctx, parsedCartID, parsedMerchantID, parsedBranchID, uuid.Nil)
	}

	parsedActorID, actorErr := parseUUID(actorID, "actor id")
	if actorErr != nil {
		return pgsqlc.Cart{}, actorErr
	}

	actor, getActorErr := service.store.GetActorByUID(ctx, pgsqlc.GetActorByUIDParams{
		MerchantID: parsedMerchantID,
		ID:         parsedActorID,
	})
	if getActorErr != nil {
		return pgsqlc.Cart{}, domainerr.MatchPostgresError(getActorErr)
	}
	if actor.MerchantID != parsedMerchantID {
		return pgsqlc.Cart{}, domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "actor does not belong to merchant", fmt.Errorf("actor does not belong to merchant"))
	}

	return service.CreateCart(ctx, parsedCartID, parsedMerchantID, parsedBranchID, parsedActorID)
}

func (service *CommerceManager) AddItemToCartHTTP(ctx context.Context, cartID string, productID string, quantity int32, addonIDs []string, discountID string) (pgsqlc.CartItem, *domainerr.DomainError) {
	parsedCartID, cartErr := parseUUID(cartID, "cart id")
	if cartErr != nil {
		return pgsqlc.CartItem{}, cartErr
	}
	parsedProductID, productErr := parseUUID(productID, "product id")
	if productErr != nil {
		return pgsqlc.CartItem{}, productErr
	}

	parsedAddonIDs := make([]uuid.UUID, 0, len(addonIDs))
	for _, addonID := range addonIDs {
		parsedAddonID, addonErr := parseUUID(addonID, "addon id")
		if addonErr != nil {
			return pgsqlc.CartItem{}, addonErr
		}
		parsedAddonIDs = append(parsedAddonIDs, parsedAddonID)
	}

	parsedDiscountID := uuid.Nil
	if strings.TrimSpace(discountID) != "" {
		discountUUID, parsedErr := parseUUID(discountID, "discount id")
		if parsedErr != nil {
			return pgsqlc.CartItem{}, parsedErr
		}
		parsedDiscountID = discountUUID
	}

	return service.AddItemToCart(ctx, parsedCartID, parsedProductID, quantity, parsedAddonIDs, parsedDiscountID, 0)
}

func (service *CommerceManager) PlaceOrderFromCartHTTP(ctx context.Context, cartID string, paymentType string, deliveryAddress string, customerName string, customerPhone string) (ports.OrderBill, *domainerr.DomainError) {
	parsedCartID, parseErr := parseUUID(cartID, "cart id")
	if parseErr != nil {
		return ports.OrderBill{}, parseErr
	}

	cart, cartErr := service.getCartByID(ctx, parsedCartID)
	if cartErr != nil {
		return ports.OrderBill{}, cartErr
	}

	actorID := uuid.Nil
	if cart.ActorID.Valid {
		actorID = uuid.UUID(cart.ActorID.Bytes)
	}
	bill, billErr := service.PlaceOrderFromCart(ctx, actorID, parsedCartID, pgsqlc.PaymentType(strings.ToLower(strings.TrimSpace(paymentType))), deliveryAddress, customerName, customerPhone)
	if billErr != nil {
		return ports.OrderBill{}, billErr
	}

	lineItems := make([]ports.OrderLineBill, 0, len(bill.LineItems))
	for _, line := range bill.LineItems {
		lineItems = append(lineItems, ports.OrderLineBill{
			ProductID:      line.ProductID,
			Quantity:       line.Quantity,
			BaseAmount:     line.BaseAmount,
			AddonAmount:    line.AddonAmount,
			DiscountAmount: line.DiscountAmount,
			TaxAmount:      line.TaxAmount,
			LineTotal:      line.LineTotal,
		})
	}

	return ports.OrderBill{
		OrderID:   bill.OrderID,
		VatRate:   bill.VatRate,
		Subtotal:  bill.Subtotal,
		TotalTax:  bill.TotalTax,
		Total:     bill.Total,
		LineItems: lineItems,
	}, nil
}

func (service *CommerceManager) UpdateOrderStatusHTTP(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, orderID string, status string) (pgsqlc.Order, *domainerr.DomainError) {
	parsedMerchantID, merchantErr := parseUUID(merchantID, "merchant id")
	if merchantErr != nil {
		return pgsqlc.Order{}, merchantErr
	}
	viewer, accessErr := service.requireMerchantManageAccess(ctx, viewerMerchantID, viewerEmail, parsedMerchantID)
	if accessErr != nil {
		return pgsqlc.Order{}, accessErr
	}
	parsedOrderID, orderErr := parseUUID(orderID, "order id")
	if orderErr != nil {
		return pgsqlc.Order{}, orderErr
	}

	return service.UpdateOrderStatus(ctx, viewer.UID, parsedMerchantID, parsedOrderID, pgsqlc.OrderStatusType(strings.ToLower(strings.TrimSpace(status))))
}
