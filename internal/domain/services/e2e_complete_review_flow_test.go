package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/http/token"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/pgsqlc"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/ports"
	"github.com/stretchr/testify/require"
)

func TestE2E_Reviewable_MerchantCustomerCompleteFlow(t *testing.T) {
	ctx := context.Background()
	pool := setupServiceTestDB(t, ctx)
	store := pgsqlc.New(pool)

	actorService := NewActorManager(store)
	merchantService := NewMerchantManager(pool, store)
	commerceService := NewCommerceManager(pool, store)

	tokenMaker, tokenMakerErr := token.NewJWTMaker("test-super-secret-jwt-key-32-chars")
	require.NoError(t, tokenMakerErr)

	merchantReq := ports.NewMerchant{
		Name:          "Suijing",
		Ntn:           "NTN-SUIJING-REVIEW-001",
		Address:       "Jauhar",
		Category:      "restaurant",
		ContactNumber: "+92336123123",
	}
	logJSONSnapshot(t, "step_01_create_merchant_request", merchantReq)

	merchant, merchantErr := merchantService.CreateMerchant(ctx, merchantReq)
	require.Nil(t, merchantErr)
	require.Equal(t, "Suijing", merchant.Name)
	logJSONSnapshot(t, "step_01_create_merchant_response", merchant)

	merchantActorReq := ports.BootstrapActor{
		FullName: "Suijing Owner",
		Email:    "owner@suijing.test",
		Password: "Password#123",
		Role:     "merchant",
	}
	logJSONSnapshot(t, "step_02_bootstrap_merchant_actor_request", merchantActorReq)

	merchantActor, bootstrapErr := merchantService.BootstrapActor(ctx, merchant.ID.String(), merchantActorReq)
	require.Nil(t, bootstrapErr)
	require.Equal(t, merchantActorReq.Email, merchantActor.Email)
	logJSONSnapshot(t, "step_02_bootstrap_merchant_actor_response", merchantActor)

	merchantLoginPayload, merchantSession := loginAndCreateSession(
		t,
		ctx,
		tokenMaker,
		actorService,
		merchant.ID,
		merchantActorReq.Email,
		merchantActorReq.Password,
		string(pgsqlc.RoleTypeMerchant),
		"merchant-login-session",
	)
	require.Equal(t, merchant.ID, merchantLoginPayload.MerchantID)
	require.Equal(t, string(pgsqlc.RoleTypeMerchant), merchantLoginPayload.Role)
	require.Equal(t, merchantActorReq.Email, merchantSession.ActorEmail)
	logJSONSnapshot(t, "step_03_merchant_login_claims", merchantLoginPayload)
	logJSONSnapshot(t, "step_03_merchant_login_session", merchantSession)

	branchA, branchAErr := store.CreateBranch(ctx, pgsqlc.CreateBranchParams{
		MerchantID:    merchantLoginPayload.MerchantID,
		Name:          "Suijing Jauhar Branch",
		Address:       "Jauhar Block 13",
		ContactNumber: textValue("02111111111111"),
		City:          pgsqlc.CityTypeKarachi,
	})
	require.NoError(t, branchAErr)

	branchB, branchBErr := store.CreateBranch(ctx, pgsqlc.CreateBranchParams{
		MerchantID:    merchantLoginPayload.MerchantID,
		Name:          "Suijing Gulshan Branch",
		Address:       "Gulshan-e-Iqbal",
		ContactNumber: textValue("02122222222222"),
		City:          pgsqlc.CityTypeKarachi,
	})
	require.NoError(t, branchBErr)
	logJSONSnapshot(t, "step_04_create_branches_response", []pgsqlc.Branch{branchA, branchB})

	area, areaErr := store.CreateArea(ctx, pgsqlc.CreateAreaParams{
		Name: "Gulshan/Jauhar Coverage",
		City: pgsqlc.CityTypeKarachi,
	})
	require.NoError(t, areaErr)

	zone, zoneErr := store.CreateZone(ctx, pgsqlc.CreateZoneParams{
		AreaID:         area.ID,
		Name:           "Suijing Coverage Zone",
		StGeomfromtext: "POLYGON((67.1000 24.9000, 67.1500 24.9000, 67.1500 24.9400, 67.1000 24.9400, 67.1000 24.9000))",
	})
	require.NoError(t, zoneErr)

	serviceZone, serviceZoneErr := store.CreateMerchantServiceZone(ctx, pgsqlc.CreateMerchantServiceZoneParams{
		MerchantID: merchantLoginPayload.MerchantID,
		ZoneID:     zone.ID,
		BranchID:   branchA.ID,
	})
	require.NoError(t, serviceZoneErr)
	logJSONSnapshot(t, "step_05_create_merchant_service_zone_response", serviceZone)

	category, categoryErr := store.CreateProductCategory(ctx, pgsqlc.CreateProductCategoryParams{
		MerchantID:  merchantLoginPayload.MerchantID,
		Name:        "Sushi",
		Description: textValue("Signature sushi menu"),
	})
	require.NoError(t, categoryErr)

	product, productErr := commerceService.CreateProductByMerchant(
		ctx,
		merchantActor.UID,
		merchantLoginPayload.MerchantID,
		category.ID,
		"Dragon Roll",
		"Best seller roll",
		1690,
		"",
		false,
	)
	require.Nil(t, productErr)

	addonA, addonAErr := commerceService.AddProductAddonByMerchant(
		ctx,
		merchantActor.UID,
		merchantLoginPayload.MerchantID,
		product.ID,
		"Extra Wasabi",
		120,
	)
	require.Nil(t, addonAErr)

	addonB, addonBErr := commerceService.AddProductAddonByMerchant(
		ctx,
		merchantActor.UID,
		merchantLoginPayload.MerchantID,
		product.ID,
		"Spicy Mayo",
		180,
	)
	require.Nil(t, addonBErr)

	_, inventoryErr := store.UpsertProductInventory(ctx, pgsqlc.UpsertProductInventoryParams{
		ProductID: product.ID,
		BranchID:  branchA.ID,
		Quantity:  50,
	})
	require.NoError(t, inventoryErr)

	discount, discountErr := store.CreateMerchantDiscount(ctx, pgsqlc.CreateMerchantDiscountParams{
		MerchantID:  merchantLoginPayload.MerchantID,
		Type:        pgsqlc.DiscountTypeFlat,
		Value:       numericFromFloat(200),
		Description: textValue("Launch discount"),
		ValidFrom:   time.Now().Add(-time.Hour),
		ValidTo:     time.Now().Add(24 * time.Hour),
	})
	require.NoError(t, discountErr)

	_, vatErr := store.UpsertVatRule(ctx, pgsqlc.UpsertVatRuleParams{
		MerchantID:  merchantLoginPayload.MerchantID,
		PaymentType: pgsqlc.PaymentTypeCard,
		Rate:        numericFromFloat(5),
	})
	require.NoError(t, vatErr)
	logJSONSnapshot(t, "step_06_catalog_and_discount_response", map[string]any{
		"category": category,
		"product":  product,
		"addons":   []pgsqlc.ProductAddon{addonA, addonB},
		"discount": discount,
	})

	// Guest checkout is still actor-backed in the current schema because sessions reference actor_id.
	guestCustomerReq := ports.NewActor{
		MerchantID: merchantLoginPayload.MerchantID,
		FullName:   "Guest Customer",
		Email:      "guest.customer@suijing.test",
		Password:   "Password#123",
	}
	guestCustomerActor, guestCustomerActorErr := actorService.CreateActor(ctx, guestCustomerReq)
	require.Nil(t, guestCustomerActorErr)

	customerRoleID, customerRoleErr := commerceService.getRoleIDByType(ctx, merchantLoginPayload.MerchantID, pgsqlc.RoleTypeCustomer)
	require.NoError(t, customerRoleErr)

	customerRole, assignCustomerRoleErr := store.AssignActorRole(ctx, pgsqlc.AssignActorRoleParams{
		MerchantID: merchantLoginPayload.MerchantID,
		ActorID:    guestCustomerActor.UID,
		RoleID:     customerRoleID,
	})
	require.NoError(t, assignCustomerRoleErr)
	logJSONSnapshot(t, "step_07_customer_role_assignment_response", customerRole)

	guestSessionReq := ports.NewActorSession{
		RefreshTokenID:        uuid.New(),
		MerchantID:            merchantLoginPayload.MerchantID,
		ActorID:               guestCustomerActor.UID,
		RefreshToken:          "guest-fe-session-" + uuid.NewString(),
		UserAgent:             "guest-checkout",
		ClientIP:              "127.0.0.1",
		RefreshTokenExpiresAt: time.Now().Add(24 * time.Hour),
	}
	customerSession, customerSessionErr := actorService.CreateActorSession(ctx, guestSessionReq)
	require.Nil(t, customerSessionErr)
	require.Equal(t, guestCustomerReq.Email, customerSession.ActorEmail)
	logJSONSnapshot(t, "step_08_guest_customer_session_request", guestSessionReq)
	logJSONSnapshot(t, "step_08_guest_customer_session_response", customerSession)

	cart, cartErr := commerceService.CreateCart(ctx, uuid.New(), merchantLoginPayload.MerchantID, branchA.ID, guestCustomerActor.UID)
	require.Nil(t, cartErr)
	require.NotEqual(t, uuid.Nil, cart.ID)

	cartItem, addItemErr := commerceService.AddItemToCart(
		ctx,
		cart.ID,
		product.ID,
		2,
		[]uuid.UUID{addonA.ID, addonB.ID},
		discount.ID,
		0,
	)
	require.Nil(t, addItemErr)
	require.Equal(t, int32(2), cartItem.Quantity)
	require.Equal(t, discount.ID, cartItem.AppliedDiscountID)
	logJSONSnapshot(t, "step_09_customer_cart_response", map[string]any{
		"cart":      cart,
		"cart_item": cartItem,
	})

	orderBill, orderErr := commerceService.PlaceOrderFromCart(
		ctx,
		guestCustomerActor.UID,
		cart.ID,
		pgsqlc.PaymentTypeCard,
		"Jauhar Karachi",
		"Sara Ahmed",
		"03001234567",
	)
	require.Nil(t, orderErr)
	require.Len(t, orderBill.LineItems, 1)
	require.Greater(t, orderBill.Total, 0.0)
	logJSONSnapshot(t, "step_10_customer_place_order_response", orderBill)

	acceptedOrder, acceptErr := commerceService.UpdateOrderStatus(ctx, merchantActor.UID, merchantLoginPayload.MerchantID, orderBill.OrderID, pgsqlc.OrderStatusTypeAccepted)
	require.Nil(t, acceptErr)
	require.Equal(t, pgsqlc.OrderStatusTypeAccepted, acceptedOrder.Status)

	outForDeliveryOrder, outForDeliveryErr := commerceService.UpdateOrderStatus(ctx, merchantActor.UID, merchantLoginPayload.MerchantID, orderBill.OrderID, pgsqlc.OrderStatusTypeOutForDelivery)
	require.Nil(t, outForDeliveryErr)
	require.Equal(t, pgsqlc.OrderStatusTypeOutForDelivery, outForDeliveryOrder.Status)

	deliveredOrder, deliveredErr := commerceService.UpdateOrderStatus(ctx, merchantActor.UID, merchantLoginPayload.MerchantID, orderBill.OrderID, pgsqlc.OrderStatusTypeDelivered)
	require.Nil(t, deliveredErr)
	require.Equal(t, pgsqlc.OrderStatusTypeDelivered, deliveredOrder.Status)
	logJSONSnapshot(t, "step_11_merchant_order_state_updates", []pgsqlc.Order{acceptedOrder, outForDeliveryOrder, deliveredOrder})

	reportMonth := int(time.Now().Month())
	reportYear := time.Now().Year()
	report, reportErr := commerceService.SalesReportByMonth(ctx, merchantActor.UID, merchantLoginPayload.MerchantID, reportMonth, reportYear)
	require.Nil(t, reportErr)
	require.Equal(t, reportMonth, report.Month)
	require.Equal(t, reportYear, report.Year)
	require.Greater(t, report.TotalSales, 0.0)

	inventory, inventoryErr := commerceService.ViewInventory(ctx, merchantActor.UID, merchantLoginPayload.MerchantID)
	require.Nil(t, inventoryErr)
	require.NotEmpty(t, inventory)
	logJSONSnapshot(t, "step_12_merchant_report_response", report)
	logJSONSnapshot(t, "step_12_inventory_response", inventory)
}

func loginAndCreateSession(
	t *testing.T,
	ctx context.Context,
	tokenMaker *token.JWTMaker,
	actorService *ActorManager,
	merchantID uuid.UUID,
	email string,
	password string,
	role string,
	sessionLabel string,
) (*token.Payload, pgsqlc.CreateSessionRow) {
	t.Helper()

	loginActor, loginErr := actorService.LoginActor(ctx, ports.LoginActor{
		MerchantID: merchantID,
		Email:      email,
		Password:   password,
	})
	require.Nil(t, loginErr)

	accessToken, _, accessTokenErr := tokenMaker.CreateToken(loginActor.Email, role, loginActor.MerchantID, 15*time.Minute)
	require.Nil(t, accessTokenErr)
	require.NotEmpty(t, accessToken)

	refreshToken, refreshPayload, refreshTokenErr := tokenMaker.CreateToken(loginActor.Email, role, loginActor.MerchantID, 24*time.Hour)
	require.Nil(t, refreshTokenErr)
	require.NotEmpty(t, refreshToken)

	session, sessionErr := actorService.CreateActorSession(ctx, ports.NewActorSession{
		RefreshTokenID:        refreshPayload.ID,
		MerchantID:            loginActor.MerchantID,
		ActorID:               loginActor.UID,
		RefreshToken:          refreshToken,
		UserAgent:             sessionLabel,
		ClientIP:              "127.0.0.1",
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
	})
	require.Nil(t, sessionErr)

	storedSession, storedSessionErr := actorService.GetActorSession(ctx, refreshPayload.ID)
	require.Nil(t, storedSessionErr)
	require.Equal(t, refreshToken, storedSession.RefreshToken)
	require.Equal(t, loginActor.Email, storedSession.ActorEmail)

	verifiedPayload, verifyErr := tokenMaker.VerifyToken(accessToken)
	require.Nil(t, verifyErr)
	require.Equal(t, loginActor.Email, verifiedPayload.Email)
	require.Equal(t, role, verifiedPayload.Role)
	require.Equal(t, merchantID, verifiedPayload.MerchantID)

	return verifiedPayload, session
}
