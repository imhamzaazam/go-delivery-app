package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/pgsqlc"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/ports"
	service "github.com/horiondreher/go-web-api-boilerplate/internal/domain/services"
	"github.com/horiondreher/go-web-api-boilerplate/internal/utils"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestCartPayloadSnapshotsV1(t *testing.T) {
	ctx := context.Background()

	server, store, merchantID := newCartIntegrationServer(t, ctx)
	fixture := setupCartPayloadFixture(t, ctx, server, store, merchantID)

	createCartBody := fmt.Sprintf(`
{
  "merchant_id": "%s",
  "branch_id": "%s",
  "cart_id": "%s"
}
`, fixture.merchantID, fixture.branchID, fixture.cartID)

	createCartReq := httptest.NewRequest(http.MethodPost, "/api/v1/carts", strings.NewReader(createCartBody))
	createCartReq.Header.Set("Content-Type", "application/json")
	createCartRecorder := httptest.NewRecorder()

	server.router.ServeHTTP(createCartRecorder, createCartReq)

	require.Equal(t, http.StatusCreated, createCartRecorder.Code)
	createCartResponseBody := createCartRecorder.Body.String()
	createdCart := decodeJSONBody[CreateCartResponse](t, createCartRecorder)
	require.NotNil(t, createdCart.Id)
	require.NotNil(t, createdCart.BranchId)
	require.NotNil(t, createdCart.CreatedAt)
	require.Equal(t, fixture.cartID, uuid.UUID(*createdCart.Id))

	createCartExpected := fmt.Sprintf(`
{
  "id": "%s",
  "branch_id": "%s",
  "created_at": "%s"
}
`, fixture.cartID, fixture.branchID, createdCart.CreatedAt.Format(time.RFC3339Nano))

	require.JSONEq(t, createCartExpected, createCartResponseBody)

	addItemBody := fmt.Sprintf(`
{
  "product_id": "%s",
  "quantity": 2,
  "addon_ids": [
    "%s"
  ],
  "discount_id": "%s"
}
`, fixture.productID, fixture.addonID, fixture.discountID)

	addItemReq := httptest.NewRequest(http.MethodPost, "/api/v1/carts/"+createdCart.Id.String()+"/items", strings.NewReader(addItemBody))
	addItemReq.Header.Set("Content-Type", "application/json")
	addItemRecorder := httptest.NewRecorder()

	server.router.ServeHTTP(addItemRecorder, addItemReq)

	require.Equal(t, http.StatusCreated, addItemRecorder.Code)
	addItemResponseBody := addItemRecorder.Body.String()
	addedItem := decodeJSONBody[CartItemResponse](t, addItemRecorder)
	require.NotNil(t, addedItem.ProductId)
	require.NotNil(t, addedItem.Quantity)

	addItemExpected := fmt.Sprintf(`
{
  "product_id": "%s",
  "quantity": 2
}
`, fixture.productID)

	require.JSONEq(t, addItemExpected, addItemResponseBody)

	getCartReq := httptest.NewRequest(http.MethodGet, "/api/v1/carts/"+createdCart.Id.String()+"?payment_type=card", nil)
	getCartReq.Header.Set("Authorization", "Bearer "+fixture.accessToken)
	getCartRecorder := httptest.NewRecorder()

	server.router.ServeHTTP(getCartRecorder, getCartReq)

	require.Equal(t, http.StatusOK, getCartRecorder.Code)
	getCartResponseBody := getCartRecorder.Body.String()
	var cartDetail map[string]any
	err := json.NewDecoder(strings.NewReader(getCartResponseBody)).Decode(&cartDetail)
	require.NoError(t, err)
	products, ok := cartDetail["products"].([]any)
	require.True(t, ok)
	require.Len(t, products, 1)

	getCartExpected := fmt.Sprintf(`
{
	"cart_id": "%s",
	"total_price": 596,
	"discount": {
		"id": "%s",
		"type": "flat",
		"value": 20,
		"amount": 20,
		"description": "Payload cart discount"
	},
	"tax": {
		"rate": 10,
		"amount": 56
	},
	"products": [
    {
			"id": "%s",
			"name": "Payload Cart Product",
			"price": 250,
      "quantity": 2,
			"addons": [
				{
					"id": "%s",
					"name": "Payload Cart Addon",
					"price": 30
				}
			]
    }
  ]
}
`, createdCart.Id.String(), fixture.discountID, fixture.productID, fixture.addonID)

	require.JSONEq(t, getCartExpected, getCartResponseBody)
}

type cartPayloadFixture struct {
	merchantID       uuid.UUID
	branchID         uuid.UUID
	cartID           uuid.UUID
	categoryID       uuid.UUID
	productID        uuid.UUID
	productCreatedAt time.Time
	productUpdatedAt time.Time
	addonID          uuid.UUID
	addonCreatedAt   time.Time
	discountID       uuid.UUID
	accessToken      string
}

func newCartIntegrationServer(t *testing.T, ctx context.Context) (*HTTPAdapter, *pgsqlc.Queries, uuid.UUID) {
	t.Helper()

	utils.SetConfigFile("../../../../.env")
	config := utils.GetConfig()

	migrationsPath := filepath.Join("..", "..", "..", "..", "db", "postgres", "migration", "*.up.sql")
	upMigrations, err := filepath.Glob(migrationsPath)
	require.NoError(t, err)

	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgis/postgis:16-3.4"),
		postgres.WithInitScripts(upMigrations...),
		postgres.WithDatabase(config.DBName),
		postgres.WithUsername(config.DBUser),
		postgres.WithPassword(config.DBPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second),
		),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = pgContainer.Terminate(ctx)
	})

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	conn, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)

	t.Cleanup(func() {
		conn.Close()
	})

	store := pgsqlc.New(conn)
	merchant, err := store.CreateMerchant(ctx, pgsqlc.CreateMerchantParams{
		Name:          "Test Merchant",
		Ntn:           "TEST-NTN-0001",
		Address:       "Test Address",
		Category:      pgsqlc.MerchantCategoryRestaurant,
		ContactNumber: "12345678901234",
	})
	require.NoError(t, err)

	actorService := service.NewActorManager(store)
	merchantService := service.NewMerchantManager(conn, store)
	commerceService := service.NewCommerceManager(conn, store)

	server, err := NewHTTPAdapter(AdapterDependencies{
		ActorService:    actorService,
		CommerceService: commerceService,
		MerchantService: merchantService,
		ReadService:     commerceService,
	})
	require.NoError(t, err)

	return server, store, merchant.ID
}

func setupCartPayloadFixture(t *testing.T, ctx context.Context, server *HTTPAdapter, store *pgsqlc.Queries, merchantID uuid.UUID) cartPayloadFixture {
	t.Helper()

	actorService := service.NewActorManager(store)

	merchantActor, actorErr := actorService.CreateActor(ctx, ports.NewActor{
		MerchantID: merchantID,
		FullName:   "Payload Cart Merchant",
		Email:      fmt.Sprintf("cart-merchant-%s@test.local", uuid.NewString()),
		Password:   "Password#123",
	})
	require.Nil(t, actorErr)

	merchantRoleID := ensureRoleInStore(t, ctx, store, merchantID, pgsqlc.RoleTypeMerchant)
	_, assignRoleErr := store.AssignActorRole(ctx, pgsqlc.AssignActorRoleParams{
		MerchantID: merchantID,
		ActorID:    merchantActor.UID,
		RoleID:     merchantRoleID,
	})
	require.NoError(t, assignRoleErr)

	branch, branchErr := store.CreateBranch(ctx, pgsqlc.CreateBranchParams{
		MerchantID:    merchantID,
		Name:          "Payload Cart Branch",
		Address:       "Payload Cart Address",
		ContactNumber: testText("02100000000000"),
		City:          pgsqlc.CityTypeKarachi,
	})
	require.NoError(t, branchErr)

	category, categoryErr := store.CreateProductCategory(ctx, pgsqlc.CreateProductCategoryParams{
		MerchantID:  merchantID,
		Name:        "Payload Cart Category",
		Description: testText("Payload cart category"),
	})
	require.NoError(t, categoryErr)

	product, productErr := store.CreateProduct(ctx, pgsqlc.CreateProductParams{
		MerchantID:     merchantID,
		CategoryID:     category.ID,
		Name:           "Payload Cart Product",
		Description:    testText("Payload cart product"),
		BasePrice:      testNumeric(250),
		ImageUrl:       testText(""),
		TrackInventory: false,
		IsActive:       true,
	})
	require.NoError(t, productErr)

	addon, addonErr := store.CreateProductAddon(ctx, pgsqlc.CreateProductAddonParams{
		ProductID: product.ID,
		Name:      "Payload Cart Addon",
		Price:     testNumeric(30),
	})
	require.NoError(t, addonErr)

	discount, discountErr := store.CreateMerchantDiscount(ctx, pgsqlc.CreateMerchantDiscountParams{
		MerchantID:  merchantID,
		Type:        pgsqlc.DiscountTypeFlat,
		Value:       testNumeric(20),
		Description: testText("Payload cart discount"),
		ValidFrom:   time.Now().Add(-time.Hour),
		ValidTo:     time.Now().Add(24 * time.Hour),
	})
	require.NoError(t, discountErr)

	_, vatErr := store.UpsertVatRule(ctx, pgsqlc.UpsertVatRuleParams{
		MerchantID:  merchantID,
		PaymentType: pgsqlc.PaymentTypeCard,
		Rate:        testNumeric(10),
	})
	require.NoError(t, vatErr)

	customer, customerErr := actorService.CreateActor(ctx, ports.NewActor{
		MerchantID: merchantID,
		FullName:   "Payload Cart Customer",
		Email:      fmt.Sprintf("cart-customer-%s@test.local", uuid.NewString()),
		Password:   "Password#123",
	})
	require.Nil(t, customerErr)

	cartID := uuid.New()
	_, sessionErr := actorService.CreateActorSession(ctx, ports.NewActorSession{
		RefreshTokenID:        cartID,
		MerchantID:            merchantID,
		ActorID:               customer.UID,
		RefreshToken:          "payload-cart-refresh",
		UserAgent:             "test",
		ClientIP:              "127.0.0.1",
		RefreshTokenExpiresAt: time.Now().Add(24 * time.Hour),
	})
	require.Nil(t, sessionErr)

	accessToken, _, tokenErr := server.tokenMaker.CreateToken(merchantActor.Email, "actor", merchantID, server.config.AccessTokenDuration)
	require.Nil(t, tokenErr)

	return cartPayloadFixture{
		merchantID:       merchantID,
		branchID:         branch.ID,
		cartID:           cartID,
		categoryID:       category.ID,
		productID:        product.ID,
		productCreatedAt: product.CreatedAt,
		productUpdatedAt: product.UpdatedAt,
		addonID:          addon.ID,
		addonCreatedAt:   addon.CreatedAt,
		discountID:       discount.ID,
		accessToken:      accessToken,
	}
}

func ensureRoleInStore(t *testing.T, ctx context.Context, store *pgsqlc.Queries, merchantID uuid.UUID, roleType pgsqlc.RoleType) uuid.UUID {
	t.Helper()

	roles, err := store.ListRolesByMerchant(ctx, merchantID)
	require.NoError(t, err)

	for _, role := range roles {
		if role.RoleType == roleType {
			return role.ID
		}
	}

	role, createErr := store.CreateRole(ctx, pgsqlc.CreateRoleParams{
		MerchantID:  merchantID,
		RoleType:    roleType,
		Description: testText(string(roleType)),
	})
	require.NoError(t, createErr)

	return role.ID
}
