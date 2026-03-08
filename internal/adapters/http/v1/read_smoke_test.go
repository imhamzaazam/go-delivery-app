package v1

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/pgsqlc"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/ports"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func TestReadSurfaceSmokeV1(t *testing.T) {
	server, serverErr := NewHTTPAdapter(AdapterDependencies{
		ActorService:    testActorService,
		CommerceService: testReadService,
		MerchantService: testMerchantService,
		ReadService:     testReadService,
	})
	require.NoError(t, serverErr)

	ctx := context.Background()

	authActor, actorErr := testActorService.CreateActor(ctx, ports.NewActor{
		MerchantID: testMerchantID,
		FullName:   "Merchant Reader",
		Email:      fmt.Sprintf("reader-%s@test.local", uuid.NewString()),
		Password:   "Password#123",
	})
	require.Nil(t, actorErr)

	merchantRoleID := ensureRole(t, ctx, testMerchantID, pgsqlc.RoleTypeMerchant)
	_, assignRoleErr := testStore.AssignActorRole(ctx, pgsqlc.AssignActorRoleParams{
		MerchantID: testMerchantID,
		ActorID:    authActor.UID,
		RoleID:     merchantRoleID,
	})
	require.NoError(t, assignRoleErr)

	branch, branchErr := testStore.CreateBranch(ctx, pgsqlc.CreateBranchParams{
		MerchantID:    testMerchantID,
		Name:          "Smoke Branch",
		Address:       "Smoke Address",
		ContactNumber: testText("02100000000000"),
		City:          pgsqlc.CityTypeKarachi,
	})
	require.NoError(t, branchErr)

	discount, discountErr := testStore.CreateMerchantDiscount(ctx, pgsqlc.CreateMerchantDiscountParams{
		MerchantID:  testMerchantID,
		Type:        pgsqlc.DiscountTypeFlat,
		Value:       testNumeric(25),
		Description: testText("Smoke discount"),
		ValidFrom:   time.Now().Add(-time.Hour),
		ValidTo:     time.Now().Add(24 * time.Hour),
	})
	require.NoError(t, discountErr)
	require.NotEqual(t, uuid.Nil, discount.ID)

	category, categoryErr := testStore.CreateProductCategory(ctx, pgsqlc.CreateProductCategoryParams{
		MerchantID:  testMerchantID,
		Name:        "Smoke Category",
		Description: testText("Category"),
	})
	require.NoError(t, categoryErr)

	product, productErr := testStore.CreateProduct(ctx, pgsqlc.CreateProductParams{
		MerchantID:     testMerchantID,
		CategoryID:     category.ID,
		Name:           "Smoke Product",
		Description:    testText("Product"),
		BasePrice:      testNumeric(100),
		ImageUrl:       testText(""),
		TrackInventory: true,
		IsActive:       true,
	})
	require.NoError(t, productErr)

	addon, addonErr := testStore.CreateProductAddon(ctx, pgsqlc.CreateProductAddonParams{
		ProductID: product.ID,
		Name:      "Smoke Addon",
		Price:     testNumeric(10),
	})
	require.NoError(t, addonErr)

	_, inventoryErr := testStore.UpsertProductInventory(ctx, pgsqlc.UpsertProductInventoryParams{
		ProductID: product.ID,
		BranchID:  branch.ID,
		Quantity:  20,
	})
	require.NoError(t, inventoryErr)

	area, areaErr := testStore.CreateArea(ctx, pgsqlc.CreateAreaParams{
		Name: "Smoke Area",
		City: pgsqlc.CityTypeKarachi,
	})
	require.NoError(t, areaErr)

	zone, zoneErr := testStore.CreateZone(ctx, pgsqlc.CreateZoneParams{
		AreaID:         area.ID,
		Name:           "Smoke Zone",
		StGeomfromtext: "POLYGON((67.00 24.00, 67.01 24.00, 67.01 24.01, 67.00 24.01, 67.00 24.00))",
	})
	require.NoError(t, zoneErr)

	_, serviceZoneErr := testStore.CreateMerchantServiceZone(ctx, pgsqlc.CreateMerchantServiceZoneParams{
		MerchantID: testMerchantID,
		ZoneID:     zone.ID,
		BranchID:   branch.ID,
	})
	require.NoError(t, serviceZoneErr)

	customer, customerErr := testActorService.CreateActor(ctx, ports.NewActor{
		MerchantID: testMerchantID,
		FullName:   "Smoke Customer",
		Email:      fmt.Sprintf("customer-%s@test.local", uuid.NewString()),
		Password:   "Password#123",
	})
	require.Nil(t, customerErr)

	_, sessionErr := testActorService.CreateActorSession(ctx, ports.NewActorSession{
		RefreshTokenID:        uuid.New(),
		MerchantID:            testMerchantID,
		ActorID:               customer.UID,
		RefreshToken:          "smoke-refresh-token",
		UserAgent:             "test",
		ClientIP:              "127.0.0.1",
		RefreshTokenExpiresAt: time.Now().Add(24 * time.Hour),
	})
	require.Nil(t, sessionErr)

	cart, cartErr := testReadService.CreateCart(ctx, uuid.New(), testMerchantID, branch.ID, customer.UID)
	require.Nil(t, cartErr)

	_, cartItemErr := testReadService.AddItemToCart(ctx, cart.ID, product.ID, 2, []uuid.UUID{addon.ID}, discount.ID, 0)
	require.Nil(t, cartItemErr)

	order, orderErr := testStore.CreateOrder(ctx, pgsqlc.CreateOrderParams{
		CartID:          cart.ID,
		MerchantID:      testMerchantID,
		BranchID:        branch.ID,
		ActorID:         pgtype.UUID{Bytes: customer.UID, Valid: true},
		PaymentType:     pgsqlc.PaymentTypeCard,
		VatRate:         testNumeric(5),
		TotalAmount:     testNumeric(220),
		Status:          pgsqlc.OrderStatusTypeAccepted,
		DeliveryAddress: "Smoke Address",
		CustomerName:    "Smoke Customer",
		CustomerPhone:   "03000000000",
	})
	require.NoError(t, orderErr)

	_, orderItemErr := testStore.UpsertOrderItem(ctx, pgsqlc.UpsertOrderItemParams{
		OrderID:        order.ID,
		ProductID:      product.ID,
		Quantity:       2,
		Price:          testNumeric(100),
		BaseAmount:     testNumeric(200),
		AddonAmount:    testNumeric(20),
		DiscountAmount: testNumeric(0),
		TaxAmount:      testNumeric(10),
		LineTotal:      testNumeric(230),
	})
	require.NoError(t, orderItemErr)

	_, orderAddonErr := testStore.UpsertOrderItemAddon(ctx, pgsqlc.UpsertOrderItemAddonParams{
		OrderID:        order.ID,
		ProductID:      product.ID,
		AddonID:        addon.ID,
		AddonName:      addon.Name,
		AddonPrice:     addon.Price,
		Quantity:       2,
		LineAddonTotal: testNumeric(20),
	})
	require.NoError(t, orderAddonErr)

	accessToken, _, tokenErr := server.tokenMaker.CreateToken(authActor.Email, "actor", testMerchantID, server.config.AccessTokenDuration)
	require.Nil(t, tokenErr)

	routes := []string{
		"/api/v1/actor/" + authActor.UID.String(),
		"/api/v1/actors/me",
		"/api/v1/actors/" + authActor.UID.String(),
		"/api/v1/merchants",
		"/api/v1/merchant",
		"/api/v1/merchant/actors",
		"/api/v1/merchant/employees",
		"/api/v1/merchant/branches",
		"/api/v1/merchant/categories",
		"/api/v1/merchant/products",
		"/api/v1/products/" + product.ID.String(),
		"/api/v1/products/" + product.ID.String() + "/addons",
		"/api/v1/merchant/discounts",
		"/api/v1/merchant/roles",
		"/api/v1/carts/" + cart.ID.String() + "?payment_type=card",
		"/api/v1/orders/" + order.ID.String(),
		"/api/v1/merchant/orders",
		"/api/v1/merchant/inventory",
		fmt.Sprintf("/api/v1/merchant/reports/sales?month=%d&year=%d", time.Now().Month(), time.Now().Year()),
		"/api/v1/merchant/service-zones",
		"/api/v1/areas",
		"/api/v1/areas/" + area.ID.String() + "/zones",
	}

	for _, route := range routes {
		req, reqErr := http.NewRequest("GET", route, nil)
		require.NoError(t, reqErr)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		recorder := httptest.NewRecorder()
		server.router.ServeHTTP(recorder, req)
		require.Equal(t, http.StatusOK, recorder.Code, route)
	}
}

func ensureRole(t *testing.T, ctx context.Context, merchantID uuid.UUID, roleType pgsqlc.RoleType) uuid.UUID {
	t.Helper()

	roles, err := testStore.ListRolesByMerchant(ctx, merchantID)
	require.NoError(t, err)

	for _, role := range roles {
		if role.RoleType == roleType {
			return role.ID
		}
	}

	role, createErr := testStore.CreateRole(ctx, pgsqlc.CreateRoleParams{
		MerchantID:  merchantID,
		RoleType:    roleType,
		Description: testText(string(roleType)),
	})
	require.NoError(t, createErr)

	return role.ID
}

func testNumeric(value float64) pgtype.Numeric {
	var numeric pgtype.Numeric
	_ = numeric.Scan(fmt.Sprintf("%.2f", value))
	return numeric
}

func testText(value string) pgtype.Text {
	return pgtype.Text{String: value, Valid: true}
}
