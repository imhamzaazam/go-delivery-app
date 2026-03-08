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

// TestCreateProduct_AcceptsImageUrlAndTrackInventory verifies product creation accepts image_url and track_inventory.
func TestCreateProduct_AcceptsImageUrlAndTrackInventory(t *testing.T) {
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
		FullName:   "Product Test Merchant",
		Email:      fmt.Sprintf("product-test-%s@test.local", uuid.NewString()),
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

	accessToken, _, tokenErr := server.tokenMaker.CreateToken(authActor.Email, "actor", testMerchantID, server.config.AccessTokenDuration)
	require.Nil(t, tokenErr)

	category, categoryErr := testStore.CreateProductCategory(ctx, pgsqlc.CreateProductCategoryParams{
		MerchantID:  testMerchantID,
		Name:        "Bug Fix Category",
		Description: testText("Category for bug fix tests"),
	})
	require.NoError(t, categoryErr)

	t.Run("creates product with image_url and track_inventory false", func(t *testing.T) {
		body := fmt.Sprintf(`{"category_id":"%s","name":"Product With Image","description":"Has image","base_price":99.50,"image_url":"https://example.com/image.png","track_inventory":false}`,
			category.ID)
		req, reqErr := http.NewRequest(http.MethodPost, "/api/v1/merchant/products", bytes.NewBufferString(body))
		require.NoError(t, reqErr)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("Content-Type", "application/json")

		recorder := httptest.NewRecorder()
		server.router.ServeHTTP(recorder, req)
		require.Equal(t, http.StatusCreated, recorder.Code)

		var product ProductResponse
		decodeErr := json.NewDecoder(recorder.Body).Decode(&product)
		require.NoError(t, decodeErr)
		require.NotNil(t, product.Id)
		require.Equal(t, "Product With Image", *product.Name)
		require.NotNil(t, product.ImageUrl)
		require.Equal(t, "https://example.com/image.png", *product.ImageUrl)
		require.NotNil(t, product.TrackInventory)
		require.False(t, *product.TrackInventory)
	})

	t.Run("creates product with track_inventory true when specified", func(t *testing.T) {
		body := fmt.Sprintf(`{"category_id":"%s","name":"Tracked Product","base_price":50,"track_inventory":true}`,
			category.ID)
		req, reqErr := http.NewRequest(http.MethodPost, "/api/v1/merchant/products", bytes.NewBufferString(body))
		require.NoError(t, reqErr)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("Content-Type", "application/json")

		recorder := httptest.NewRecorder()
		server.router.ServeHTTP(recorder, req)
		require.Equal(t, http.StatusCreated, recorder.Code)

		var product ProductResponse
		decodeErr := json.NewDecoder(recorder.Body).Decode(&product)
		require.NoError(t, decodeErr)
		require.NotNil(t, product.TrackInventory)
		require.True(t, *product.TrackInventory)
	})

	t.Run("creates product without image_url or track_inventory defaults to false", func(t *testing.T) {
		body := fmt.Sprintf(`{"category_id":"%s","name":"Minimal Product","base_price":10}`,
			category.ID)
		req, reqErr := http.NewRequest(http.MethodPost, "/api/v1/merchant/products", bytes.NewBufferString(body))
		require.NoError(t, reqErr)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("Content-Type", "application/json")

		recorder := httptest.NewRecorder()
		server.router.ServeHTTP(recorder, req)
		require.Equal(t, http.StatusCreated, recorder.Code)

		var product ProductResponse
		decodeErr := json.NewDecoder(recorder.Body).Decode(&product)
		require.NoError(t, decodeErr)
		require.NotNil(t, product.TrackInventory)
		require.False(t, *product.TrackInventory)
	})
}

// TestCartResponse_SimplifiedStructure verifies cart response has simplified structure:
// no merchant_id, updated_at, applied_discount_*, discount; addons nested in product.
func TestCartResponse_SimplifiedStructure(t *testing.T) {
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
		FullName:   "Cart Test Merchant",
		Email:      fmt.Sprintf("cart-test-%s@test.local", uuid.NewString()),
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

	customer, customerErr := testActorService.CreateActor(ctx, ports.NewActor{
		MerchantID: testMerchantID,
		FullName:   "Cart Customer",
		Email:      fmt.Sprintf("cart-customer-%s@test.local", uuid.NewString()),
		Password:   "Password#123",
	})
	require.Nil(t, customerErr)

	branch, branchErr := testStore.CreateBranch(ctx, pgsqlc.CreateBranchParams{
		MerchantID:    testMerchantID,
		Name:          "Cart Branch",
		Address:       "Cart Address",
		ContactNumber: testText("02100000000000"),
		City:          pgsqlc.CityTypeKarachi,
	})
	require.NoError(t, branchErr)

	category, categoryErr := testStore.CreateProductCategory(ctx, pgsqlc.CreateProductCategoryParams{
		MerchantID:  testMerchantID,
		Name:        "Cart Category",
		Description: testText("Cart category"),
	})
	require.NoError(t, categoryErr)

	product, productErr := testStore.CreateProduct(ctx, pgsqlc.CreateProductParams{
		MerchantID:     testMerchantID,
		CategoryID:     category.ID,
		Name:           "Cart Product",
		Description:    testText("Product for cart"),
		BasePrice:      testNumeric(100),
		ImageUrl:       testText("https://example.com/product.png"),
		TrackInventory: false,
		IsActive:       true,
	})
	require.NoError(t, productErr)

	addon, addonErr := testStore.CreateProductAddon(ctx, pgsqlc.CreateProductAddonParams{
		ProductID: product.ID,
		Name:      "Cart Addon",
		Price:     testNumeric(15),
	})
	require.NoError(t, addonErr)

	_, sessionErr := testActorService.CreateActorSession(ctx, ports.NewActorSession{
		RefreshTokenID:        uuid.New(),
		MerchantID:            testMerchantID,
		ActorID:               customer.UID,
		RefreshToken:          "cart-test-refresh",
		UserAgent:             "test",
		ClientIP:              "127.0.0.1",
		RefreshTokenExpiresAt: time.Now().Add(24 * time.Hour),
	})
	require.Nil(t, sessionErr)

	cart, cartErr := testReadService.CreateCart(ctx, uuid.New(), testMerchantID, branch.ID, customer.UID)
	require.Nil(t, cartErr)

	_, addItemErr := testReadService.AddItemToCart(ctx, cart.ID, product.ID, 2, []uuid.UUID{addon.ID}, uuid.Nil, 0)
	require.Nil(t, addItemErr)

	accessToken, _, tokenErr := server.tokenMaker.CreateToken(authActor.Email, "actor", testMerchantID, server.config.AccessTokenDuration)
	require.Nil(t, tokenErr)

	req, reqErr := http.NewRequest(http.MethodGet, "/api/v1/carts/"+cart.ID.String()+"?payment_type=card", nil)
	require.NoError(t, reqErr)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	recorder := httptest.NewRecorder()
	server.router.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]any
	decodeErr := json.NewDecoder(bytes.NewReader(recorder.Body.Bytes())).Decode(&response)
	require.NoError(t, decodeErr)
	require.Equal(t, cart.ID.String(), response["cart_id"])
	require.Equal(t, 230.0, response["total_price"])
	_, hasDiscount := response["discount"]
	require.False(t, hasDiscount)

	products, ok := response["products"].([]any)
	require.True(t, ok)
	require.Len(t, products, 1)
	productResponse, ok := products[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, product.ID.String(), productResponse["id"])
	require.Equal(t, "Cart Product", productResponse["name"])
	require.Equal(t, 100.0, productResponse["price"])
	require.Equal(t, 2.0, productResponse["quantity"])
	addons, ok := productResponse["addons"].([]any)
	require.True(t, ok)
	require.Len(t, addons, 1)
	addonResponse, ok := addons[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, addon.ID.String(), addonResponse["id"])
	require.Equal(t, "Cart Addon", addonResponse["name"])
	require.Equal(t, 15.0, addonResponse["price"])
}

// TestCreateCartResponse_SimplifiedStructure verifies CreateCart response is simplified.
func TestCreateCartResponse_SimplifiedStructure(t *testing.T) {
	server, err := NewHTTPAdapter(AdapterDependencies{
		ActorService:    testActorService,
		CommerceService: testReadService,
		MerchantService: testMerchantService,
		ReadService:     testReadService,
	})
	require.NoError(t, err)

	ctx := context.Background()
	branch, branchErr := testStore.CreateBranch(ctx, pgsqlc.CreateBranchParams{
		MerchantID:    testMerchantID,
		Name:          "Create Cart Branch",
		Address:       "Address",
		ContactNumber: testText("02100000000000"),
		City:          pgsqlc.CityTypeKarachi,
	})
	require.NoError(t, branchErr)

	cartID := uuid.New()
	body := fmt.Sprintf(`{"merchant_id":"%s","branch_id":"%s","cart_id":"%s"}`, testMerchantID, branch.ID, cartID)
	req, reqErr := http.NewRequest(http.MethodPost, "/api/v1/carts", bytes.NewBufferString(body))
	require.NoError(t, reqErr)
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	server.router.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusCreated, recorder.Code)

	var cartResp CreateCartResponse
	decodeErr := json.NewDecoder(recorder.Body).Decode(&cartResp)
	require.NoError(t, decodeErr)

	require.NotNil(t, cartResp.Id)
	require.Equal(t, cartID, uuid.UUID(*cartResp.Id))
	require.NotNil(t, cartResp.CreatedAt)
	require.NotNil(t, cartResp.BranchId)
	// Schema has no merchant_id or updated_at
}

// TestAddItemToCartResponse_SimplifiedStructure verifies AddItemToCart response is simplified.
func TestAddItemToCartResponse_SimplifiedStructure(t *testing.T) {
	server, err := NewHTTPAdapter(AdapterDependencies{
		ActorService:    testActorService,
		CommerceService: testReadService,
		MerchantService: testMerchantService,
		ReadService:     testReadService,
	})
	require.NoError(t, err)

	ctx := context.Background()
	customer, customerErr := testActorService.CreateActor(ctx, ports.NewActor{
		MerchantID: testMerchantID,
		FullName:   "Add Item Customer",
		Email:      fmt.Sprintf("additem-%s@test.local", uuid.NewString()),
		Password:   "Password#123",
	})
	require.Nil(t, customerErr)

	branch, branchErr := testStore.CreateBranch(ctx, pgsqlc.CreateBranchParams{
		MerchantID:    testMerchantID,
		Name:          "Add Item Branch",
		Address:       "Address",
		ContactNumber: testText("02100000000000"),
		City:          pgsqlc.CityTypeKarachi,
	})
	require.NoError(t, branchErr)

	category, categoryErr := testStore.CreateProductCategory(ctx, pgsqlc.CreateProductCategoryParams{
		MerchantID:  testMerchantID,
		Name:        "Add Item Category",
		Description: testText("Category"),
	})
	require.NoError(t, categoryErr)

	product, productErr := testStore.CreateProduct(ctx, pgsqlc.CreateProductParams{
		MerchantID:     testMerchantID,
		CategoryID:     category.ID,
		Name:           "Add Item Product",
		Description:    testText("Product"),
		BasePrice:      testNumeric(50),
		ImageUrl:       testText(""),
		TrackInventory: false,
		IsActive:       true,
	})
	require.NoError(t, productErr)

	_, sessionErr := testActorService.CreateActorSession(ctx, ports.NewActorSession{
		RefreshTokenID:        uuid.New(),
		MerchantID:            testMerchantID,
		ActorID:               customer.UID,
		RefreshToken:          "additem-refresh",
		UserAgent:             "test",
		ClientIP:              "127.0.0.1",
		RefreshTokenExpiresAt: time.Now().Add(24 * time.Hour),
	})
	require.Nil(t, sessionErr)

	cart, cartErr := testReadService.CreateCart(ctx, uuid.New(), testMerchantID, branch.ID, customer.UID)
	require.Nil(t, cartErr)

	body := fmt.Sprintf(`{"product_id":"%s","quantity":3}`, product.ID)
	req, reqErr := http.NewRequest(http.MethodPost, "/api/v1/carts/"+cart.ID.String()+"/items", bytes.NewBufferString(body))
	require.NoError(t, reqErr)
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	server.router.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusCreated, recorder.Code)

	var itemResp CartItemResponse
	decodeErr := json.NewDecoder(recorder.Body).Decode(&itemResp)
	require.NoError(t, decodeErr)

	require.NotNil(t, itemResp.ProductId)
	require.NotNil(t, itemResp.Quantity)
	require.Equal(t, 3, *itemResp.Quantity)
	// Schema has no cart_id, applied_discount_*, addon_ids at top level
}
