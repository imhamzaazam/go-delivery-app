package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/pgsqlc"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/ports"
	"github.com/stretchr/testify/require"
)

func TestGetCartDetail_ReturnsCartSummaryResponse(t *testing.T) {
	server, err := NewHTTPAdapter(AdapterDependencies{
		ActorService:    testActorService,
		CommerceService: testReadService,
		MerchantService: testMerchantService,
		ReadService:     testReadService,
	})
	require.NoError(t, err)

	ctx := context.Background()
	authActor, actorErr := testActorService.CreateActor(ctx, ports.NewActor{
		MerchantID: testMerchantID,
		FullName:   "Cart Summary Merchant",
		Email:      fmt.Sprintf("cart-summary-%s@test.local", uuid.NewString()),
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
		Name:          "Cart Summary Branch",
		Address:       "Cart Summary Address",
		ContactNumber: testText("02100000000000"),
		City:          pgsqlc.CityTypeKarachi,
	})
	require.NoError(t, branchErr)

	category, categoryErr := testStore.CreateProductCategory(ctx, pgsqlc.CreateProductCategoryParams{
		MerchantID:  testMerchantID,
		Name:        "Cart Summary Category",
		Description: testText("Cart summary category"),
	})
	require.NoError(t, categoryErr)

	product, productErr := testStore.CreateProduct(ctx, pgsqlc.CreateProductParams{
		MerchantID:     testMerchantID,
		CategoryID:     category.ID,
		Name:           "Cart Summary Product",
		Description:    testText("Cart summary product"),
		BasePrice:      testNumeric(100),
		ImageUrl:       testText(""),
		TrackInventory: false,
		IsActive:       true,
	})
	require.NoError(t, productErr)

	addon, addonErr := testStore.CreateProductAddon(ctx, pgsqlc.CreateProductAddonParams{
		ProductID: product.ID,
		Name:      "Cart Summary Addon",
		Price:     testNumeric(15),
	})
	require.NoError(t, addonErr)

	discount, discountErr := testStore.CreateMerchantDiscount(ctx, pgsqlc.CreateMerchantDiscountParams{
		MerchantID:  testMerchantID,
		Type:        pgsqlc.DiscountTypeFlat,
		Value:       testNumeric(10),
		Description: testText("Cart summary discount"),
		ValidFrom:   time.Now().Add(-time.Hour),
		ValidTo:     time.Now().Add(24 * time.Hour),
	})
	require.NoError(t, discountErr)

	_, vatErr := testStore.UpsertVatRule(ctx, pgsqlc.UpsertVatRuleParams{
		MerchantID:  testMerchantID,
		PaymentType: pgsqlc.PaymentTypeCard,
		Rate:        testNumeric(10),
	})
	require.NoError(t, vatErr)

	cartID := uuid.New()
	createCartBody := fmt.Sprintf(`{"merchant_id":"%s","branch_id":"%s","cart_id":"%s"}`, testMerchantID, branch.ID, cartID)
	createCartReq := httptest.NewRequest(http.MethodPost, "/api/v1/carts", bytes.NewBufferString(createCartBody))
	createCartReq.Header.Set("Content-Type", "application/json")
	createCartRecorder := httptest.NewRecorder()
	server.router.ServeHTTP(createCartRecorder, createCartReq)
	require.Equal(t, http.StatusCreated, createCartRecorder.Code)

	addItemBody := fmt.Sprintf(`{"product_id":"%s","quantity":2,"addon_ids":["%s"],"discount_id":"%s"}`, product.ID, addon.ID, discount.ID)
	addItemReq := httptest.NewRequest(http.MethodPost, "/api/v1/carts/"+cartID.String()+"/items", bytes.NewBufferString(addItemBody))
	addItemReq.Header.Set("Content-Type", "application/json")
	addItemRecorder := httptest.NewRecorder()
	server.router.ServeHTTP(addItemRecorder, addItemReq)
	require.Equal(t, http.StatusCreated, addItemRecorder.Code)

	accessToken, _, tokenErr := server.tokenMaker.CreateToken(authActor.Email, "actor", testMerchantID, server.config.AccessTokenDuration)
	require.Nil(t, tokenErr)

	getCartReq := httptest.NewRequest(http.MethodGet, "/api/v1/carts/"+cartID.String()+"?payment_type=card", nil)
	getCartReq.Header.Set("Authorization", "Bearer "+accessToken)
	getCartRecorder := httptest.NewRecorder()
	server.router.ServeHTTP(getCartRecorder, getCartReq)

	require.Equal(t, http.StatusOK, getCartRecorder.Code)

	expected := fmt.Sprintf(`
{
  "cart_id": "%s",
	"total_price": 243,
  "discount": {
    "id": "%s",
    "type": "flat",
    "value": 10,
    "amount": 10,
    "description": "Cart summary discount"
  },
  "tax": {
		"rate": 10,
		"amount": 23
  },
  "products": [
    {
      "id": "%s",
      "name": "Cart Summary Product",
      "price": 100,
      "quantity": 2,
      "addons": [
        {
          "id": "%s",
          "name": "Cart Summary Addon",
          "price": 15
        }
      ]
    }
  ]
}
`, cartID, discount.ID, product.ID, addon.ID)

	require.JSONEq(t, expected, getCartRecorder.Body.String())
}

func TestGetCartDetail_TaxDoesNotChangeWithDiscount(t *testing.T) {
	server, err := NewHTTPAdapter(AdapterDependencies{
		ActorService:    testActorService,
		CommerceService: testReadService,
		MerchantService: testMerchantService,
		ReadService:     testReadService,
	})
	require.NoError(t, err)

	ctx := context.Background()
	authActor, actorErr := testActorService.CreateActor(ctx, ports.NewActor{
		MerchantID: testMerchantID,
		FullName:   "Cart Tax Merchant",
		Email:      fmt.Sprintf("cart-tax-%s@test.local", uuid.NewString()),
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
		Name:          "Cart Tax Branch",
		Address:       "Cart Tax Address",
		ContactNumber: testText("02100000000000"),
		City:          pgsqlc.CityTypeKarachi,
	})
	require.NoError(t, branchErr)

	category, categoryErr := testStore.CreateProductCategory(ctx, pgsqlc.CreateProductCategoryParams{
		MerchantID:  testMerchantID,
		Name:        "Cart Tax Category",
		Description: testText("Cart tax category"),
	})
	require.NoError(t, categoryErr)

	product, productErr := testStore.CreateProduct(ctx, pgsqlc.CreateProductParams{
		MerchantID:     testMerchantID,
		CategoryID:     category.ID,
		Name:           "Cart Tax Product",
		Description:    testText("Cart tax product"),
		BasePrice:      testNumeric(100),
		ImageUrl:       testText(""),
		TrackInventory: false,
		IsActive:       true,
	})
	require.NoError(t, productErr)

	addon, addonErr := testStore.CreateProductAddon(ctx, pgsqlc.CreateProductAddonParams{
		ProductID: product.ID,
		Name:      "Cart Tax Addon",
		Price:     testNumeric(15),
	})
	require.NoError(t, addonErr)

	discount, discountErr := testStore.CreateMerchantDiscount(ctx, pgsqlc.CreateMerchantDiscountParams{
		MerchantID:  testMerchantID,
		Type:        pgsqlc.DiscountTypeFlat,
		Value:       testNumeric(10),
		Description: testText("Cart tax discount"),
		ValidFrom:   time.Now().Add(-time.Hour),
		ValidTo:     time.Now().Add(24 * time.Hour),
	})
	require.NoError(t, discountErr)

	_, vatErr := testStore.UpsertVatRule(ctx, pgsqlc.UpsertVatRuleParams{
		MerchantID:  testMerchantID,
		PaymentType: pgsqlc.PaymentTypeCard,
		Rate:        testNumeric(10),
	})
	require.NoError(t, vatErr)

	accessToken, _, tokenErr := server.tokenMaker.CreateToken(authActor.Email, "actor", testMerchantID, server.config.AccessTokenDuration)
	require.Nil(t, tokenErr)

	createCart := func(t *testing.T) uuid.UUID {
		t.Helper()
		cartID := uuid.New()
		createCartBody := fmt.Sprintf(`{"merchant_id":"%s","branch_id":"%s","cart_id":"%s"}`, testMerchantID, branch.ID, cartID)
		createCartReq := httptest.NewRequest(http.MethodPost, "/api/v1/carts", bytes.NewBufferString(createCartBody))
		createCartReq.Header.Set("Content-Type", "application/json")
		createCartRecorder := httptest.NewRecorder()
		server.router.ServeHTTP(createCartRecorder, createCartReq)
		require.Equal(t, http.StatusCreated, createCartRecorder.Code)
		return cartID
	}

	getCart := func(t *testing.T, cartID uuid.UUID) map[string]any {
		t.Helper()
		getCartReq := httptest.NewRequest(http.MethodGet, "/api/v1/carts/"+cartID.String()+"?payment_type=card", nil)
		getCartReq.Header.Set("Authorization", "Bearer "+accessToken)
		getCartRecorder := httptest.NewRecorder()
		server.router.ServeHTTP(getCartRecorder, getCartReq)
		require.Equal(t, http.StatusOK, getCartRecorder.Code)

		var payload map[string]any
		decodeErr := json.NewDecoder(bytes.NewReader(getCartRecorder.Body.Bytes())).Decode(&payload)
		require.NoError(t, decodeErr)
		return payload
	}

	discountedCartID := createCart(t)
	discountedBody := fmt.Sprintf(`{"product_id":"%s","quantity":2,"addon_ids":["%s"],"discount_id":"%s"}`, product.ID, addon.ID, discount.ID)
	discountedReq := httptest.NewRequest(http.MethodPost, "/api/v1/carts/"+discountedCartID.String()+"/items", bytes.NewBufferString(discountedBody))
	discountedReq.Header.Set("Content-Type", "application/json")
	discountedRecorder := httptest.NewRecorder()
	server.router.ServeHTTP(discountedRecorder, discountedReq)
	require.Equal(t, http.StatusCreated, discountedRecorder.Code)

	fullPriceCartID := createCart(t)
	fullPriceBody := fmt.Sprintf(`{"product_id":"%s","quantity":2,"addon_ids":["%s"]}`, product.ID, addon.ID)
	fullPriceReq := httptest.NewRequest(http.MethodPost, "/api/v1/carts/"+fullPriceCartID.String()+"/items", bytes.NewBufferString(fullPriceBody))
	fullPriceReq.Header.Set("Content-Type", "application/json")
	fullPriceRecorder := httptest.NewRecorder()
	server.router.ServeHTTP(fullPriceRecorder, fullPriceReq)
	require.Equal(t, http.StatusCreated, fullPriceRecorder.Code)

	discountedCart := getCart(t, discountedCartID)
	fullPriceCart := getCart(t, fullPriceCartID)

	discountedTax := discountedCart["tax"].(map[string]any)
	fullPriceTax := fullPriceCart["tax"].(map[string]any)
	require.Equal(t, 23.0, discountedTax["amount"])
	require.Equal(t, 23.0, fullPriceTax["amount"])
	require.Equal(t, discountedTax["amount"], fullPriceTax["amount"])

	require.Equal(t, 243.0, discountedCart["total_price"])
	require.Equal(t, 253.0, fullPriceCart["total_price"])
	_, hasDiscount := discountedCart["discount"]
	require.True(t, hasDiscount)
	_, hasFullPriceDiscount := fullPriceCart["discount"]
	require.False(t, hasFullPriceDiscount)
}
