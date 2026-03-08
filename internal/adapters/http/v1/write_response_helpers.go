package v1

import (
	"time"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/pgsqlc"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/ports"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func areaResponse(area pgsqlc.Area) AreaResponse {
	city := string(area.City)
	createdAt := area.CreatedAt
	return AreaResponse{
		Id:        ptrUUID(area.ID),
		Name:      &area.Name,
		City:      &city,
		CreatedAt: &createdAt,
	}
}

func zoneResponse(id uuid.UUID, areaID uuid.UUID, name string, coordinatesWKT string, createdAt time.Time) ZoneResponse {
	return ZoneResponse{
		Id:             ptrUUID(id),
		AreaId:         ptrUUID(areaID),
		Name:           &name,
		CoordinatesWkt: ptrString(coordinatesWKT),
		CreatedAt:      &createdAt,
	}
}

func branchResponse(branch pgsqlc.Branch) BranchResponse {
	contactNumber := textString(branch.ContactNumber)
	city := string(branch.City)
	createdAt := branch.CreatedAt
	updatedAt := branch.UpdatedAt
	return BranchResponse{
		Id:            ptrUUID(branch.ID),
		MerchantId:    ptrUUID(branch.MerchantID),
		Name:          &branch.Name,
		Address:       &branch.Address,
		ContactNumber: &contactNumber,
		City:          &city,
		CreatedAt:     &createdAt,
		UpdatedAt:     &updatedAt,
	}
}

func categoryResponse(category pgsqlc.ProductCategory) ProductCategoryResponse {
	description := textString(category.Description)
	createdAt := category.CreatedAt
	return ProductCategoryResponse{
		Id:          ptrUUID(category.ID),
		MerchantId:  ptrUUID(category.MerchantID),
		Name:        &category.Name,
		Description: &description,
		CreatedAt:   &createdAt,
	}
}

func productResponse(product pgsqlc.Product) ProductResponse {
	description := textString(product.Description)
	imageURL := textString(product.ImageUrl)
	basePrice := numericToFloat64(product.BasePrice)
	createdAt := product.CreatedAt
	updatedAt := product.UpdatedAt
	categoryID := openapi_types.UUID(product.CategoryID)
	return ProductResponse{
		Id:             ptrUUID(product.ID),
		MerchantId:     ptrUUID(product.MerchantID),
		CategoryId:     &categoryID,
		Name:           &product.Name,
		Description:    &description,
		BasePrice:      &basePrice,
		ImageUrl:       &imageURL,
		TrackInventory: &product.TrackInventory,
		IsActive:       &product.IsActive,
		CreatedAt:      &createdAt,
		UpdatedAt:      &updatedAt,
	}
}

func productAddonResponse(addon pgsqlc.ProductAddon) ProductAddonResponse {
	price := numericToFloat64(addon.Price)
	createdAt := addon.CreatedAt
	return ProductAddonResponse{
		Id:        ptrUUID(addon.ID),
		ProductId: ptrUUID(addon.ProductID),
		Name:      &addon.Name,
		Price:     &price,
		CreatedAt: &createdAt,
	}
}

func discountResponse(discount pgsqlc.MerchantDiscount) DiscountResponse {
	description := textString(discount.Description)
	discountType := string(discount.Type)
	value := numericToFloat64(discount.Value)
	createdAt := discount.CreatedAt
	response := DiscountResponse{
		Id:          ptrUUID(discount.ID),
		MerchantId:  ptrUUID(discount.MerchantID),
		Type:        &discountType,
		Value:       &value,
		Description: &description,
		CreatedAt:   &createdAt,
	}
	if !discount.ValidFrom.IsZero() {
		validFrom := discount.ValidFrom
		response.ValidFrom = &validFrom
	}
	if !discount.ValidTo.IsZero() {
		validTo := discount.ValidTo
		response.ValidTo = &validTo
	}
	return response
}

func inventoryResponse(inventory pgsqlc.ProductInventory) ProductInventoryResponse {
	quantity := int(inventory.Quantity)
	createdAt := inventory.CreatedAt
	updatedAt := inventory.UpdatedAt
	return ProductInventoryResponse{
		Id:        ptrUUID(inventory.ID),
		ProductId: ptrUUID(inventory.ProductID),
		BranchId:  ptrUUID(inventory.BranchID),
		Quantity:  &quantity,
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
	}
}

func cartCreatedResponse(cart pgsqlc.Cart) CreateCartResponse {
	createdAt := cart.CreatedAt
	response := CreateCartResponse{
		Id:        ptrUUID(cart.ID),
		CreatedAt: &createdAt,
	}
	if cart.BranchID != uuid.Nil {
		branchID := openapi_types.UUID(cart.BranchID)
		response.BranchId = &branchID
	}
	return response
}

func cartItemCreatedResponse(item pgsqlc.CartItem) CartItemResponse {
	quantity := int(item.Quantity)
	return CartItemResponse{
		ProductId: ptrUUID(item.ProductID),
		Quantity:  &quantity,
	}
}

func orderBillResponse(bill ports.OrderBill) OrderBillResponse {
	subtotal := bill.Subtotal
	totalTax := bill.TotalTax
	total := bill.Total
	vatRate := bill.VatRate
	lines := make([]OrderBillLineResponse, 0, len(bill.LineItems))
	for _, line := range bill.LineItems {
		quantity := int(line.Quantity)
		baseAmount := line.BaseAmount
		addonAmount := line.AddonAmount
		discountAmount := line.DiscountAmount
		taxAmount := line.TaxAmount
		lineTotal := line.LineTotal
		lines = append(lines, OrderBillLineResponse{
			ProductId:      ptrUUID(line.ProductID),
			Quantity:       &quantity,
			BaseAmount:     &baseAmount,
			AddonAmount:    &addonAmount,
			DiscountAmount: &discountAmount,
			TaxAmount:      &taxAmount,
			LineTotal:      &lineTotal,
		})
	}

	response := OrderBillResponse{
		OrderId:   ptrUUID(bill.OrderID),
		Subtotal:  &subtotal,
		TotalTax:  &totalTax,
		Total:     &total,
		VatRate:   &vatRate,
		LineItems: &lines,
	}
	return response
}

func orderSummaryResponse(order pgsqlc.Order) OrderSummaryResponse {
	totalAmount := numericToFloat64(order.TotalAmount)
	vatRate := numericToFloat64(order.VatRate)
	paymentType := string(order.PaymentType)
	status := string(order.Status)
	createdAt := order.CreatedAt
	updatedAt := order.UpdatedAt
	response := OrderSummaryResponse{
		Id:              ptrUUID(order.ID),
		CartId:          ptrUUID(order.CartID),
		MerchantId:      ptrUUID(order.MerchantID),
		PaymentType:     &paymentType,
		VatRate:         &vatRate,
		TotalAmount:     &totalAmount,
		Status:          &status,
		DeliveryAddress: &order.DeliveryAddress,
		CustomerName:    &order.CustomerName,
		CustomerPhone:   &order.CustomerPhone,
		CreatedAt:       &createdAt,
		UpdatedAt:       &updatedAt,
	}
	if order.ActorID.Valid {
		actorID := openapi_types.UUID(uuid.UUID(order.ActorID.Bytes))
		response.ActorId = &actorID
	}
	if order.BranchID != uuid.Nil {
		branchID := openapi_types.UUID(order.BranchID)
		response.BranchId = &branchID
	}
	return response
}

func merchantServiceZoneResponse(zone pgsqlc.MerchantServiceZone) MerchantServiceZoneResponse {
	createdAt := zone.CreatedAt
	response := MerchantServiceZoneResponse{
		Id:         ptrUUID(zone.ID),
		MerchantId: ptrUUID(zone.MerchantID),
		ZoneId:     ptrUUID(zone.ZoneID),
		CreatedAt:  &createdAt,
	}
	if zone.BranchID != uuid.Nil {
		branchID := openapi_types.UUID(zone.BranchID)
		response.BranchId = &branchID
	}
	return response
}
