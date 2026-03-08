package services

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/pgsqlc"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/ports"
	"github.com/stretchr/testify/require"
)

type actorCreateRequest struct {
	MerchantID uuid.UUID `json:"merchant_id"`
	FullName   string    `json:"full_name"`
	Email      string    `json:"email"`
	Password   string    `json:"password"`
}

type actorCreateResponse struct {
	UID      uuid.UUID `json:"uid"`
	Email    string    `json:"email"`
	FullName string    `json:"full_name"`
}

type sessionCreateRequest struct {
	RefreshTokenID        uuid.UUID `json:"refresh_token_id"`
	MerchantID            uuid.UUID `json:"merchant_id"`
	ActorID               uuid.UUID `json:"actor_id"`
	RefreshToken          string    `json:"refresh_token"`
	UserAgent             string    `json:"user_agent"`
	ClientIP              string    `json:"client_ip"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
}

type sessionCreateResponse struct {
	ID           uuid.UUID `json:"id"`
	ActorEmail   string    `json:"actor_email"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type cartAddItemRequest struct {
	CartID      uuid.UUID   `json:"cart_id"`
	ProductID   uuid.UUID   `json:"product_id"`
	Quantity    int32       `json:"quantity"`
	AddonIDs    []uuid.UUID `json:"addon_ids"`
	DiscountID  uuid.UUID   `json:"discount_id"`
	DiscountRaw float64     `json:"discount_amount_input"`
}

type cartAddItemResponse struct {
	CartID                uuid.UUID   `json:"cart_id"`
	ProductID             uuid.UUID   `json:"product_id"`
	Quantity              int32       `json:"quantity"`
	AddonIDs              []uuid.UUID `json:"addon_ids"`
	AppliedDiscountID     uuid.UUID   `json:"applied_discount_id"`
	AppliedDiscountAmount float64     `json:"applied_discount_amount"`
}

type placeOrderRequest struct {
	MerchantActorID uuid.UUID          `json:"merchant_actor_id"`
	CartID          uuid.UUID          `json:"cart_id"`
	PaymentType     pgsqlc.PaymentType `json:"payment_type"`
	DeliveryAddress string             `json:"delivery_address"`
	CustomerName    string             `json:"customer_name"`
	CustomerPhone   string             `json:"customer_phone"`
}

type inventorySnapshot struct {
	ProductID   uuid.UUID `json:"product_id"`
	ProductName string    `json:"product_name"`
	Quantity    int32     `json:"quantity"`
}

type salesReportSnapshot struct {
	Month          int     `json:"month"`
	Year           int     `json:"year"`
	TotalSales     float64 `json:"total_sales"`
	TotalTax       float64 `json:"total_tax"`
	TotalDiscount  float64 `json:"total_discount"`
	ProfitEstimate float64 `json:"profit_estimate"`
}

func TestE2E_ServiceRequestResponse_MainCommerceFlow(t *testing.T) {
	fx := setupCommerceFixture(t)

	category, categoryErr := fx.store.CreateProductCategory(fx.ctx, pgsqlc.CreateProductCategoryParams{
		MerchantID:  fx.merchantID,
		Name:        "E2E Category",
		Description: textValue("E2E category"),
	})
	require.NoError(t, categoryErr)

	product, productErr := fx.commerceService.CreateProductByMerchant(
		fx.ctx,
		fx.merchantOwnerID,
		fx.merchantID,
		category.ID,
		"Combo Burger",
		"Double patty burger",
		1200,
		"",
		true, // track_inventory for inventory decrement test
	)
	require.Nil(t, productErr)

	addonA, addonAErr := fx.commerceService.AddProductAddonByMerchant(
		fx.ctx,
		fx.merchantOwnerID,
		fx.merchantID,
		product.ID,
		"Extra Cheese",
		100,
	)
	require.Nil(t, addonAErr)

	addonB, addonBErr := fx.commerceService.AddProductAddonByMerchant(
		fx.ctx,
		fx.merchantOwnerID,
		fx.merchantID,
		product.ID,
		"Fries Upgrade",
		150,
	)
	require.Nil(t, addonBErr)

	_, inventoryErr := fx.store.UpsertProductInventory(fx.ctx, pgsqlc.UpsertProductInventoryParams{
		ProductID: product.ID,
		BranchID:  fx.merchantBranch,
		Quantity:  25,
	})
	require.NoError(t, inventoryErr)

	discount, discountErr := fx.store.CreateMerchantDiscount(fx.ctx, pgsqlc.CreateMerchantDiscountParams{
		MerchantID:  fx.merchantID,
		Type:        pgsqlc.DiscountTypeFlat,
		Value:       numericFromFloat(200),
		Description: textValue("Launch discount"),
		ValidFrom:   time.Now().Add(-time.Hour),
		ValidTo:     time.Now().Add(24 * time.Hour),
	})
	require.NoError(t, discountErr)

	_, vatErr := fx.store.UpsertVatRule(fx.ctx, pgsqlc.UpsertVatRuleParams{
		MerchantID:  fx.merchantID,
		PaymentType: pgsqlc.PaymentTypeCard,
		Rate:        numericFromFloat(10),
	})
	require.NoError(t, vatErr)

	createActorReq := actorCreateRequest{
		MerchantID: fx.merchantID,
		FullName:   "Snapshot Customer",
		Email:      "snapshot.customer@test.local",
		Password:   "Password#123",
	}
	createdActor, createActorDomainErr := fx.actorService.CreateActor(fx.ctx, ports.NewActor{
		MerchantID: createActorReq.MerchantID,
		FullName:   createActorReq.FullName,
		Email:      createActorReq.Email,
		Password:   createActorReq.Password,
	})
	require.Nil(t, createActorDomainErr)

	createActorRes := actorCreateResponse{
		UID:      createdActor.UID,
		Email:    createdActor.Email,
		FullName: createdActor.FullName,
	}

	sessionReq := sessionCreateRequest{
		RefreshTokenID:        uuid.New(),
		MerchantID:            fx.merchantID,
		ActorID:               createdActor.UID,
		RefreshToken:          "e2e-visible-refresh-token",
		UserAgent:             "testcontainers-e2e",
		ClientIP:              "127.0.0.1",
		RefreshTokenExpiresAt: time.Now().Add(24 * time.Hour),
	}
	session, sessionDomainErr := fx.actorService.CreateActorSession(fx.ctx, ports.NewActorSession{
		RefreshTokenID:        sessionReq.RefreshTokenID,
		MerchantID:            sessionReq.MerchantID,
		ActorID:               sessionReq.ActorID,
		RefreshToken:          sessionReq.RefreshToken,
		UserAgent:             sessionReq.UserAgent,
		ClientIP:              sessionReq.ClientIP,
		RefreshTokenExpiresAt: sessionReq.RefreshTokenExpiresAt,
	})
	require.Nil(t, sessionDomainErr)

	sessionRes := sessionCreateResponse{
		ID:           session.ID,
		ActorEmail:   session.ActorEmail,
		RefreshToken: session.RefreshToken,
		ExpiresAt:    session.ExpiresAt,
	}

	cart, cartErr := fx.commerceService.CreateCart(fx.ctx, uuid.New(), fx.merchantID, fx.merchantBranch, createdActor.UID)
	require.Nil(t, cartErr)

	addItemReq := cartAddItemRequest{
		CartID:      cart.ID,
		ProductID:   product.ID,
		Quantity:    2,
		AddonIDs:    []uuid.UUID{addonA.ID, addonB.ID},
		DiscountID:  discount.ID,
		DiscountRaw: 0,
	}
	cartItem, addItemErr := fx.commerceService.AddItemToCart(
		fx.ctx,
		addItemReq.CartID,
		addItemReq.ProductID,
		addItemReq.Quantity,
		addItemReq.AddonIDs,
		addItemReq.DiscountID,
		addItemReq.DiscountRaw,
	)
	require.Nil(t, addItemErr)

	addItemRes := cartAddItemResponse{
		CartID:                cartItem.CartID,
		ProductID:             cartItem.ProductID,
		Quantity:              cartItem.Quantity,
		AddonIDs:              cartItem.AddonIds,
		AppliedDiscountID:     cartItem.AppliedDiscountID,
		AppliedDiscountAmount: numericToFloat(cartItem.AppliedDiscountAmount),
	}

	orderReq := placeOrderRequest{
		MerchantActorID: fx.merchantOwnerID,
		CartID:          cart.ID,
		PaymentType:     pgsqlc.PaymentTypeCard,
		DeliveryAddress: "DHA Karachi",
		CustomerName:    "Snapshot Customer",
		CustomerPhone:   "03000000000",
	}
	orderRes, orderDomainErr := fx.commerceService.PlaceOrderFromCart(
		fx.ctx,
		orderReq.MerchantActorID,
		orderReq.CartID,
		orderReq.PaymentType,
		orderReq.DeliveryAddress,
		orderReq.CustomerName,
		orderReq.CustomerPhone,
	)
	require.Nil(t, orderDomainErr)

	_, statusErr := fx.commerceService.UpdateOrderStatus(
		fx.ctx,
		fx.merchantOwnerID,
		fx.merchantID,
		orderRes.OrderID,
		pgsqlc.OrderStatusTypeAccepted,
	)
	require.Nil(t, statusErr)

	reportMonth := int(time.Now().Month())
	reportYear := time.Now().Year()
	report, reportErr := fx.commerceService.SalesReportByMonth(fx.ctx, fx.merchantOwnerID, fx.merchantID, reportMonth, reportYear)
	require.Nil(t, reportErr)

	inventory, inventoryViewErr := fx.commerceService.ViewInventory(fx.ctx, fx.merchantOwnerID, fx.merchantID)
	require.Nil(t, inventoryViewErr)

	require.Equal(t, "snapshot.customer@test.local", createActorRes.Email)
	require.Equal(t, createdActor.UID, sessionReq.ActorID)
	require.Equal(t, sessionReq.RefreshToken, sessionRes.RefreshToken)
	require.Equal(t, int32(2), addItemRes.Quantity)
	require.Len(t, addItemRes.AddonIDs, 2)
	require.Equal(t, 200.00, addItemRes.AppliedDiscountAmount)
	require.Len(t, orderRes.LineItems, 1)
	require.Equal(t, 2400.00, orderRes.LineItems[0].BaseAmount)
	require.Equal(t, 500.00, orderRes.LineItems[0].AddonAmount)
	require.Equal(t, 200.00, orderRes.LineItems[0].DiscountAmount)
	require.Equal(t, 270.00, orderRes.LineItems[0].TaxAmount)
	require.Equal(t, 2970.00, orderRes.Total)
	require.Greater(t, report.TotalSales, 0.0)

	var burgerInventory *inventorySnapshot
	for _, item := range inventory {
		if item.ProductID == product.ID {
			burgerInventory = &inventorySnapshot{
				ProductID:   item.ProductID,
				ProductName: item.ProductName,
				Quantity:    item.Quantity,
			}
			break
		}
	}
	require.NotNil(t, burgerInventory)
	require.Equal(t, int32(23), burgerInventory.Quantity)

	reportSnapshot := salesReportSnapshot{
		Month:          report.Month,
		Year:           report.Year,
		TotalSales:     report.TotalSales,
		TotalTax:       report.TotalTax,
		TotalDiscount:  report.TotalDiscount,
		ProfitEstimate: report.ProfitEstimate,
	}

	logJSONSnapshot(t, "create_actor_request", createActorReq)
	logJSONSnapshot(t, "create_actor_response", createActorRes)
	logJSONSnapshot(t, "create_session_request", sessionReq)
	logJSONSnapshot(t, "create_session_response", sessionRes)
	logJSONSnapshot(t, "add_item_request", addItemReq)
	logJSONSnapshot(t, "add_item_response", addItemRes)
	logJSONSnapshot(t, "place_order_request", orderReq)
	logJSONSnapshot(t, "place_order_response", orderRes)
	logJSONSnapshot(t, "sales_report_response", reportSnapshot)
	logJSONSnapshot(t, "inventory_response", burgerInventory)
}

func logJSONSnapshot(t *testing.T, label string, value any) {
	t.Helper()

	payload, err := json.MarshalIndent(value, "", "  ")
	require.NoError(t, err)

	t.Logf("%s:\n%s", label, string(payload))
}
