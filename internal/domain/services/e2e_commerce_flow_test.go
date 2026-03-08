package services

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/pgsqlc"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/ports"
	"github.com/stretchr/testify/require"
)

func TestTDD_ProductAndAddonsFlow(t *testing.T) {
	fx := setupCommerceFixture(t)

	product, addonA, addonB := createProductAndAddons(t, fx)
	require.NotEqual(t, uuid.Nil, product.ID)
	require.Equal(t, product.ID, addonA.ProductID)
	require.Equal(t, product.ID, addonB.ProductID)
	require.NotEqual(t, numericToFloat(addonA.Price), numericToFloat(addonB.Price))

	_, forbiddenProductErr := fx.commerceService.CreateProductByMerchant(fx.ctx, fx.customerID, fx.merchantID, product.CategoryID, "Forbidden", "No access", 500, "", false)
	require.NotNil(t, forbiddenProductErr)
	require.Equal(t, 403, forbiddenProductErr.HTTPCode)
}

func TestTDD_ActorRoleAssignmentIsOneToOne(t *testing.T) {
	fx := setupCommerceFixture(t)

	employeeRoleID, roleErr := fx.commerceService.getRoleIDByType(fx.ctx, fx.merchantID, pgsqlc.RoleTypeEmployee)
	require.NoError(t, roleErr)

	assignedRole, assignErr := fx.store.AssignActorRole(fx.ctx, pgsqlc.AssignActorRoleParams{
		MerchantID: fx.merchantID,
		ActorID:    fx.merchantOwnerID,
		RoleID:     employeeRoleID,
	})
	require.NoError(t, assignErr)
	require.Equal(t, employeeRoleID, assignedRole.RoleID)

	storedRole, getRoleErr := fx.store.GetActorRole(fx.ctx, pgsqlc.GetActorRoleParams{
		MerchantID: fx.merchantID,
		ActorID:    fx.merchantOwnerID,
	})
	require.NoError(t, getRoleErr)
	require.Equal(t, employeeRoleID, storedRole.RoleID)

	category, categoryErr := fx.store.CreateProductCategory(fx.ctx, pgsqlc.CreateProductCategoryParams{
		MerchantID:  fx.merchantID,
		Name:        "Forbidden Category",
		Description: textValue("merchant access required"),
	})
	require.NoError(t, categoryErr)

	_, createErr := fx.commerceService.CreateProductByMerchant(
		fx.ctx,
		fx.merchantOwnerID,
		fx.merchantID,
		category.ID,
		"Should Fail",
		"role changed to employee",
		100,
		"",
		false,
	)
	require.NotNil(t, createErr)
	require.Equal(t, 403, createErr.HTTPCode)
}

func TestE2E_Commerce_PlaceOrderFromCart_WithVAT(t *testing.T) {
	fx := setupCommerceFixture(t)
	product, addonA, addonB := createProductAndAddons(t, fx)

	cartItem, addCartErr := fx.commerceService.AddItemToCart(fx.ctx, fx.cartID, product.ID, 2, []uuid.UUID{addonA.ID, addonB.ID}, uuid.Nil, 0)
	require.Nil(t, addCartErr)
	require.EqualValues(t, 2, cartItem.Quantity)
	require.Len(t, cartItem.AddonIds, 2)

	_, vatErr := fx.store.UpsertVatRule(fx.ctx, pgsqlc.UpsertVatRuleParams{
		MerchantID:  fx.merchantID,
		PaymentType: pgsqlc.PaymentTypeCard,
		Rate:        numericFromFloat(10),
	})
	require.NoError(t, vatErr)

	bill, orderErr := fx.commerceService.PlaceOrderFromCart(fx.ctx, fx.merchantOwnerID, fx.cartID, pgsqlc.PaymentTypeCard, "Karachi", "Sara", "03000000000")
	require.Nil(t, orderErr)
	require.Len(t, bill.LineItems, 1)
	require.Equal(t, 2000.00, bill.LineItems[0].BaseAmount)
	require.Equal(t, 260.00, bill.LineItems[0].AddonAmount)
	require.Equal(t, 226.00, bill.LineItems[0].TaxAmount)
	require.Equal(t, 2486.00, bill.Total)

	orderItem, getItemErr := fx.store.GetOrderItem(fx.ctx, pgsqlc.GetOrderItemParams{OrderID: bill.OrderID, ProductID: product.ID})
	require.NoError(t, getItemErr)
	require.Equal(t, 2486.00, numericToFloat(orderItem.LineTotal))
}

func TestTDD_PlaceOrder_NoInventoryDecrementWhenTrackInventoryFalse(t *testing.T) {
	fx := setupCommerceFixture(t)

	category, categoryErr := fx.store.CreateProductCategory(fx.ctx, pgsqlc.CreateProductCategoryParams{
		MerchantID:  fx.merchantID,
		Name:        "No-Track Category",
		Description: textValue("Products without inventory tracking"),
	})
	require.NoError(t, categoryErr)

	product, productErr := fx.commerceService.CreateProductByMerchant(fx.ctx, fx.merchantOwnerID, fx.merchantID, category.ID, "Untracked Product", "No inventory", 100, "", false)
	require.Nil(t, productErr)
	require.False(t, product.TrackInventory)

	_, inventoryErr := fx.store.UpsertProductInventory(fx.ctx, pgsqlc.UpsertProductInventoryParams{
		ProductID: product.ID,
		BranchID:  fx.merchantBranch,
		Quantity:  50,
	})
	require.NoError(t, inventoryErr)

	_, addCartErr := fx.commerceService.AddItemToCart(fx.ctx, fx.cartID, product.ID, 5, nil, uuid.Nil, 0)
	require.Nil(t, addCartErr)

	_, placeErr := fx.commerceService.PlaceOrderFromCart(fx.ctx, fx.merchantOwnerID, fx.cartID, pgsqlc.PaymentTypeCard, "Karachi", "Sara", "03000000000")
	require.Nil(t, placeErr)

	inventory, getInvErr := fx.store.GetProductInventory(fx.ctx, pgsqlc.GetProductInventoryParams{
		ProductID: product.ID,
		BranchID:  fx.merchantBranch,
	})
	require.NoError(t, getInvErr)
	require.Equal(t, int32(50), inventory.Quantity, "inventory must not decrement when track_inventory is false")
}

func TestTDD_PlaceOrderForbiddenForDifferentCustomer(t *testing.T) {
	fx := setupCommerceFixture(t)
	product, addonA, addonB := createProductAndAddons(t, fx)

	_, addCartErr := fx.commerceService.AddItemToCart(fx.ctx, fx.cartID, product.ID, 2, []uuid.UUID{addonA.ID, addonB.ID}, uuid.Nil, 0)
	require.Nil(t, addCartErr)

	otherCustomer, otherCustomerErr := fx.actorService.CreateActor(fx.ctx, ports.NewActor{
		MerchantID: fx.merchantID,
		FullName:   "Other Customer",
		Email:      "other.customer@test.local",
		Password:   "Password#123",
	})
	require.Nil(t, otherCustomerErr)

	customerRoleID, roleErr := fx.commerceService.getRoleIDByType(fx.ctx, fx.merchantID, pgsqlc.RoleTypeCustomer)
	require.NoError(t, roleErr)

	_, assignRoleErr := fx.store.AssignActorRole(fx.ctx, pgsqlc.AssignActorRoleParams{
		MerchantID: fx.merchantID,
		ActorID:    otherCustomer.UID,
		RoleID:     customerRoleID,
	})
	require.NoError(t, assignRoleErr)

	_, placeErr := fx.commerceService.PlaceOrderFromCart(fx.ctx, otherCustomer.UID, fx.cartID, pgsqlc.PaymentTypeCard, "Karachi", "Sara", "03000000000")
	require.NotNil(t, placeErr)
	require.Equal(t, 403, placeErr.HTTPCode)
	require.Equal(t, domainerr.UnauthorizedError, placeErr.HTTPErrorBody.Code)
}

func TestTDD_PlaceOrderFailsOnEmptyCart(t *testing.T) {
	fx := setupCommerceFixture(t)

	emptyCart, createCartErr := fx.commerceService.CreateCart(fx.ctx, uuid.New(), fx.merchantID, fx.merchantBranch, fx.customerID)
	require.Nil(t, createCartErr)

	_, placeErr := fx.commerceService.PlaceOrderFromCart(fx.ctx, fx.merchantOwnerID, emptyCart.ID, pgsqlc.PaymentTypeCard, "Karachi", "Sara", "03000000000")
	require.NotNil(t, placeErr)
	require.Equal(t, 400, placeErr.HTTPCode)
	require.Equal(t, domainerr.ValidationError, placeErr.HTTPErrorBody.Code)
}

func TestTDD_DiscountTaxAndStatusFlow(t *testing.T) {
	fx := setupCommerceFixture(t)
	product, addonA, addonB := createProductAndAddons(t, fx)

	discount, discountErr := fx.store.CreateMerchantDiscount(fx.ctx, pgsqlc.CreateMerchantDiscountParams{
		MerchantID:  fx.merchantID,
		Type:        pgsqlc.DiscountTypeFlat,
		Value:       numericFromFloat(100),
		Description: textValue("TDD flat discount"),
		ValidFrom:   time.Now().Add(-1 * time.Hour),
		ValidTo:     time.Now().Add(24 * time.Hour),
	})
	require.NoError(t, discountErr)

	_, addCartErr := fx.commerceService.AddItemToCart(fx.ctx, fx.cartID, product.ID, 2, []uuid.UUID{addonA.ID, addonB.ID}, discount.ID, 100)
	require.Nil(t, addCartErr)

	_, vatErr := fx.store.UpsertVatRule(fx.ctx, pgsqlc.UpsertVatRuleParams{
		MerchantID:  fx.merchantID,
		PaymentType: pgsqlc.PaymentTypeCard,
		Rate:        numericFromFloat(10),
	})
	require.NoError(t, vatErr)

	bill, placeErr := fx.commerceService.PlaceOrderFromCart(fx.ctx, fx.merchantOwnerID, fx.cartID, pgsqlc.PaymentTypeCard, "Karachi", "Sara", "03000000000")
	require.Nil(t, placeErr)
	require.Equal(t, 100.00, bill.LineItems[0].DiscountAmount)
	require.Equal(t, 216.00, bill.LineItems[0].TaxAmount)
	require.Equal(t, 2376.00, bill.Total)

	acceptedOrder, acceptedErr := fx.commerceService.UpdateOrderStatus(fx.ctx, fx.merchantOwnerID, fx.merchantID, bill.OrderID, pgsqlc.OrderStatusTypeAccepted)
	require.Nil(t, acceptedErr)
	require.Equal(t, pgsqlc.OrderStatusTypeAccepted, acceptedOrder.Status)

	outForDeliveryOrder, outForDeliveryErr := fx.commerceService.UpdateOrderStatus(fx.ctx, fx.merchantOwnerID, fx.merchantID, bill.OrderID, pgsqlc.OrderStatusTypeOutForDelivery)
	require.Nil(t, outForDeliveryErr)
	require.Equal(t, pgsqlc.OrderStatusTypeOutForDelivery, outForDeliveryOrder.Status)

	updatedOrder, statusErr := fx.commerceService.UpdateOrderStatus(fx.ctx, fx.merchantOwnerID, fx.merchantID, bill.OrderID, pgsqlc.OrderStatusTypeDelivered)
	require.Nil(t, statusErr)
	require.Equal(t, pgsqlc.OrderStatusTypeDelivered, updatedOrder.Status)
}

func TestTDD_SalesReportAndInventoryFlow(t *testing.T) {
	fx := setupCommerceFixture(t)
	product, addonA, addonB := createProductAndAddons(t, fx)

	_, addCartErr := fx.commerceService.AddItemToCart(fx.ctx, fx.cartID, product.ID, 2, []uuid.UUID{addonA.ID, addonB.ID}, uuid.Nil, 0)
	require.Nil(t, addCartErr)

	_, vatErr := fx.store.UpsertVatRule(fx.ctx, pgsqlc.UpsertVatRuleParams{
		MerchantID:  fx.merchantID,
		PaymentType: pgsqlc.PaymentTypeCard,
		Rate:        numericFromFloat(10),
	})
	require.NoError(t, vatErr)

	bill, orderErr := fx.commerceService.PlaceOrderFromCart(fx.ctx, fx.merchantOwnerID, fx.cartID, pgsqlc.PaymentTypeCard, "Karachi", "Sara", "03000000000")
	require.Nil(t, orderErr)
	require.Greater(t, bill.Total, 0.0)

	acceptedOrder, statusErr := fx.commerceService.UpdateOrderStatus(fx.ctx, fx.merchantOwnerID, fx.merchantID, bill.OrderID, pgsqlc.OrderStatusTypeAccepted)
	require.Nil(t, statusErr)
	require.Equal(t, pgsqlc.OrderStatusTypeAccepted, acceptedOrder.Status)

	_, inventoryErr := fx.store.UpsertProductInventory(fx.ctx, pgsqlc.UpsertProductInventoryParams{
		ProductID: product.ID,
		BranchID:  fx.merchantBranch,
		Quantity:  40,
	})
	require.NoError(t, inventoryErr)

	reportMonth := int(time.Now().Month())
	reportYear := time.Now().Year()

	merchantReport, merchantReportErr := fx.commerceService.SalesReportByMonth(fx.ctx, fx.merchantOwnerID, fx.merchantID, reportMonth, reportYear)
	require.Nil(t, merchantReportErr)
	require.Greater(t, merchantReport.TotalSales, 0.0)
	require.Greater(t, merchantReport.ProfitEstimate, 0.0)

	adminReport, adminReportErr := fx.commerceService.SalesReportByMonth(fx.ctx, fx.adminActorID, fx.merchantID, reportMonth, reportYear)
	require.Nil(t, adminReportErr)
	require.Equal(t, merchantReport.TotalSales, adminReport.TotalSales)

	merchantInventory, merchantInvErr := fx.commerceService.ViewInventory(fx.ctx, fx.merchantOwnerID, fx.merchantID)
	require.Nil(t, merchantInvErr)
	require.NotEmpty(t, merchantInventory)

	adminInventory, adminInvErr := fx.commerceService.ViewInventory(fx.ctx, fx.adminActorID, fx.merchantID)
	require.Nil(t, adminInvErr)
	require.NotEmpty(t, adminInventory)
}

func TestTDD_ReportAndInventoryForbiddenForCustomer(t *testing.T) {
	fx := setupCommerceFixture(t)

	reportMonth := int(time.Now().Month())
	reportYear := time.Now().Year()

	_, reportErr := fx.commerceService.SalesReportByMonth(fx.ctx, fx.customerID, fx.merchantID, reportMonth, reportYear)
	require.NotNil(t, reportErr)
	require.Equal(t, 403, reportErr.HTTPCode)
	require.Equal(t, domainerr.UnauthorizedError, reportErr.HTTPErrorBody.Code)

	_, inventoryErr := fx.commerceService.ViewInventory(fx.ctx, fx.customerID, fx.merchantID)
	require.NotNil(t, inventoryErr)
	require.Equal(t, 403, inventoryErr.HTTPCode)
	require.Equal(t, domainerr.UnauthorizedError, inventoryErr.HTTPErrorBody.Code)
}

func createProductAndAddons(t *testing.T, fx *commerceFixture) (pgsqlc.Product, pgsqlc.ProductAddon, pgsqlc.ProductAddon) {
	t.Helper()

	category, categoryErr := fx.store.CreateProductCategory(fx.ctx, pgsqlc.CreateProductCategoryParams{
		MerchantID:  fx.merchantID,
		Name:        "Sushi",
		Description: textValue("Sushi category"),
	})
	require.NoError(t, categoryErr)

	product, productErr := fx.commerceService.CreateProductByMerchant(fx.ctx, fx.merchantOwnerID, fx.merchantID, category.ID, "Dragon Roll", "Best seller", 1000.00, "", false)
	require.Nil(t, productErr)

	addonA, addonAErr := fx.commerceService.AddProductAddonByMerchant(fx.ctx, fx.merchantOwnerID, fx.merchantID, product.ID, "Extra Wasabi", 50.00)
	require.Nil(t, addonAErr)
	addonB, addonBErr := fx.commerceService.AddProductAddonByMerchant(fx.ctx, fx.merchantOwnerID, fx.merchantID, product.ID, "Spicy Mayo", 80.00)
	require.Nil(t, addonBErr)

	_, inventoryErr := fx.store.UpsertProductInventory(fx.ctx, pgsqlc.UpsertProductInventoryParams{
		ProductID: product.ID,
		BranchID:  fx.merchantBranch,
		Quantity:  50,
	})
	require.NoError(t, inventoryErr)

	return product, addonA, addonB
}
