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
	"github.com/stretchr/testify/require"
)

func TestHTTP_MerchantGuestCheckoutCompleteFlow(t *testing.T) {
	server, err := NewHTTPAdapter(AdapterDependencies{
		ActorService:    testActorService,
		CommerceService: testReadService,
		MerchantService: testMerchantService,
		ReadService:     testReadService,
	})
	require.NoError(t, err)

	createMerchantBody := fmt.Sprintf(`{"name":"%s","ntn":"%s","address":"%s","category":"restaurant","contact_number":"%s"}`,
		"HTTP Merchant "+uuid.NewString()[:8],
		"NTN-"+uuid.NewString()[:8],
		"Flow Address",
		"12345678901234",
	)
	merchantRecorder := doJSONRequest(t, server, http.MethodPost, "/api/v1/merchants", createMerchantBody, "")
	require.Equal(t, http.StatusCreated, merchantRecorder.Code)
	merchant := decodeJSONBody[MerchantResponse](t, merchantRecorder)
	require.NotNil(t, merchant.Id)

	merchantID := uuid.UUID(*merchant.Id)
	bootstrapBody := `{"full_name":"Merchant Owner","email":"merchant.owner@test.local","password":"Password#123","role":"merchant"}`
	bootstrapRecorder := doJSONRequest(t, server, http.MethodPost, "/api/v1/merchants/"+merchantID.String()+"/bootstrap-actor", bootstrapBody, "")
	require.Equal(t, http.StatusCreated, bootstrapRecorder.Code)
	bootstrapActor := decodeJSONBody[ActorProfileResponse](t, bootstrapRecorder)
	require.NotNil(t, bootstrapActor.Uid)

	loginBody := fmt.Sprintf(`{"merchant_id":"%s","email":"%s","password":"%s"}`, merchantID, "merchant.owner@test.local", "Password#123")
	loginRecorder := doJSONRequest(t, server, http.MethodPost, "/api/v1/login", loginBody, "")
	require.Equal(t, http.StatusOK, loginRecorder.Code)
	loginResponse := decodeJSONBody[LoginActorResponse](t, loginRecorder)
	require.NotNil(t, loginResponse.AccessToken)
	accessToken := *loginResponse.AccessToken

	createAreaRecorder := doJSONRequest(t, server, http.MethodPost, "/api/v1/areas", `{"name":"HTTP Flow Area","city":"Karachi"}`, accessToken)
	require.Equal(t, http.StatusCreated, createAreaRecorder.Code)
	createdArea := decodeJSONBody[AreaResponse](t, createAreaRecorder)
	require.NotNil(t, createdArea.Id)

	createZoneRecorder := doJSONRequest(t, server, http.MethodPost, "/api/v1/areas/"+uuid.UUID(*createdArea.Id).String()+"/zones", `{"name":"HTTP Flow Zone","coordinates_wkt":"POLYGON((67.00 24.00, 67.01 24.00, 67.01 24.01, 67.00 24.01, 67.00 24.00))"}`, accessToken)
	require.Equal(t, http.StatusCreated, createZoneRecorder.Code)
	createdZone := decodeJSONBody[ZoneResponse](t, createZoneRecorder)
	require.NotNil(t, createdZone.Id)

	areasRecorder := doJSONRequest(t, server, http.MethodGet, "/api/v1/areas", "", accessToken)
	require.Equal(t, http.StatusOK, areasRecorder.Code)
	areas := decodeJSONBody[[]AreaResponse](t, areasRecorder)
	require.NotEmpty(t, areas)

	zonesRecorder := doJSONRequest(t, server, http.MethodGet, "/api/v1/areas/"+uuid.UUID(*createdArea.Id).String()+"/zones", "", accessToken)
	require.Equal(t, http.StatusOK, zonesRecorder.Code)
	zones := decodeJSONBody[[]ZoneResponse](t, zonesRecorder)
	require.NotEmpty(t, zones)

	branchBody := `{"name":"Main Branch","address":"Branch Address","contact_number":"02100000000000","city":"Karachi"}`
	branchRecorder := doJSONRequest(t, server, http.MethodPost, "/api/v1/merchant/branches", branchBody, accessToken)
	require.Equal(t, http.StatusCreated, branchRecorder.Code)
	branch := decodeJSONBody[BranchResponse](t, branchRecorder)
	require.NotNil(t, branch.Id)

	serviceZoneBody := fmt.Sprintf(`{"zone_id":"%s","branch_id":"%s"}`, uuid.UUID(*createdZone.Id), uuid.UUID(*branch.Id))
	serviceZoneRecorder := doJSONRequest(t, server, http.MethodPost, "/api/v1/merchant/service-zones", serviceZoneBody, accessToken)
	require.Equal(t, http.StatusCreated, serviceZoneRecorder.Code)
	serviceZone := decodeJSONBody[MerchantServiceZoneResponse](t, serviceZoneRecorder)
	require.NotNil(t, serviceZone.Id)

	coverageRecorder := doJSONRequest(t, server, http.MethodPost, "/api/v1/merchant/service-zones/check", `{"latitude":24.0050,"longitude":67.0050}`, accessToken)
	require.Equal(t, http.StatusOK, coverageRecorder.Code)
	coverage := decodeJSONBody[ServiceZoneCoverageCheckResponse](t, coverageRecorder)
	require.NotNil(t, coverage.Covered)
	require.True(t, *coverage.Covered)

	categoryRecorder := doJSONRequest(t, server, http.MethodPost, "/api/v1/merchant/categories", `{"name":"Burgers","description":"Main menu"}`, accessToken)
	require.Equal(t, http.StatusCreated, categoryRecorder.Code)
	category := decodeJSONBody[ProductCategoryResponse](t, categoryRecorder)
	require.NotNil(t, category.Id)

	productBody := fmt.Sprintf(`{"category_id":"%s","name":"Zinger Burger","description":"Signature","base_price":250}`, uuid.UUID(*category.Id))
	productRecorder := doJSONRequest(t, server, http.MethodPost, "/api/v1/merchant/products", productBody, accessToken)
	require.Equal(t, http.StatusCreated, productRecorder.Code)
	product := decodeJSONBody[ProductResponse](t, productRecorder)
	require.NotNil(t, product.Id)

	addonBody := `{"name":"Extra Cheese","price":30}`
	addonRecorder := doJSONRequest(t, server, http.MethodPost, "/api/v1/products/"+uuid.UUID(*product.Id).String()+"/addons", addonBody, accessToken)
	require.Equal(t, http.StatusCreated, addonRecorder.Code)
	addon := decodeJSONBody[ProductAddonResponse](t, addonRecorder)
	require.NotNil(t, addon.Id)

	discountRecorder := doJSONRequest(t, server, http.MethodPost, "/api/v1/merchant/discounts", `{"type":"flat","value":25,"description":"Flow discount"}`, accessToken)
	require.Equal(t, http.StatusCreated, discountRecorder.Code)
	discount := decodeJSONBody[DiscountResponse](t, discountRecorder)
	require.NotNil(t, discount.Id)

	inventoryBody := fmt.Sprintf(`{"product_id":"%s","branch_id":"%s","quantity":20}`, uuid.UUID(*product.Id), uuid.UUID(*branch.Id))
	inventoryRecorder := doJSONRequest(t, server, http.MethodPost, "/api/v1/merchant/inventory", inventoryBody, accessToken)
	require.Equal(t, http.StatusCreated, inventoryRecorder.Code)
	inventory := decodeJSONBody[ProductInventoryResponse](t, inventoryRecorder)
	require.NotNil(t, inventory.Quantity)
	require.Equal(t, 20, *inventory.Quantity)

	cartID := uuid.New()
	createCartBody := fmt.Sprintf(`{"merchant_id":"%s","branch_id":"%s","cart_id":"%s"}`, merchantID, uuid.UUID(*branch.Id), cartID)
	createCartRecorder := doJSONRequest(t, server, http.MethodPost, "/api/v1/carts", createCartBody, "")
	require.Equal(t, http.StatusCreated, createCartRecorder.Code)
	cart := decodeJSONBody[CreateCartResponse](t, createCartRecorder)
	require.NotNil(t, cart.Id)
	require.Equal(t, cartID, uuid.UUID(*cart.Id))

	addItemBody := fmt.Sprintf(`{"product_id":"%s","quantity":2,"addon_ids":["%s"],"discount_id":"%s"}`, uuid.UUID(*product.Id), uuid.UUID(*addon.Id), uuid.UUID(*discount.Id))
	addItemRecorder := doJSONRequest(t, server, http.MethodPost, "/api/v1/carts/"+uuid.UUID(*cart.Id).String()+"/items", addItemBody, "")
	require.Equal(t, http.StatusCreated, addItemRecorder.Code)
	cartItem := decodeJSONBody[CartItemResponse](t, addItemRecorder)
	require.NotNil(t, cartItem.Quantity)
	require.Equal(t, 2, *cartItem.Quantity)

	_, vatErr := testStore.UpsertVatRule(context.Background(), pgsqlc.UpsertVatRuleParams{
		MerchantID:  merchantID,
		PaymentType: pgsqlc.PaymentTypeCard,
		Rate:        testNumeric(10),
	})
	require.NoError(t, vatErr)

	getCartRecorder := doJSONRequest(t, server, http.MethodGet, "/api/v1/carts/"+uuid.UUID(*cart.Id).String()+"?payment_type=card", "", accessToken)
	require.Equal(t, http.StatusOK, getCartRecorder.Code)
	var cartDetail map[string]any
	err = json.NewDecoder(getCartRecorder.Body).Decode(&cartDetail)
	require.NoError(t, err)
	products, ok := cartDetail["products"].([]any)
	require.True(t, ok)
	require.Len(t, products, 1)

	createOrderBody := fmt.Sprintf(`{"cart_id":"%s","payment_type":"card","delivery_address":"Customer Address","customer_name":"Sara Ahmed","customer_phone":"03000000000"}`, uuid.UUID(*cart.Id))
	createOrderRecorder := doJSONRequest(t, server, http.MethodPost, "/api/v1/orders", createOrderBody, "")
	require.Equal(t, http.StatusCreated, createOrderRecorder.Code)
	orderBill := decodeJSONBody[OrderBillResponse](t, createOrderRecorder)
	require.NotNil(t, orderBill.OrderId)
	require.NotNil(t, orderBill.Total)
	require.Greater(t, *orderBill.Total, 0.0)

	orderID := uuid.UUID(*orderBill.OrderId)
	orderDetailRecorder := doJSONRequest(t, server, http.MethodGet, "/api/v1/orders/"+orderID.String(), "", accessToken)
	require.Equal(t, http.StatusOK, orderDetailRecorder.Code)
	orderDetail := decodeJSONBody[OrderDetailResponse](t, orderDetailRecorder)
	require.NotNil(t, orderDetail.Status)
	require.Equal(t, "pending", *orderDetail.Status)

	for _, status := range []string{"accepted", "out_for_delivery", "delivered"} {
		updateStatusRecorder := doJSONRequest(t, server, http.MethodPatch, "/api/v1/orders/"+orderID.String(), fmt.Sprintf(`{"status":"%s"}`, status), accessToken)
		require.Equal(t, http.StatusOK, updateStatusRecorder.Code)
		updatedOrder := decodeJSONBody[OrderSummaryResponse](t, updateStatusRecorder)
		require.NotNil(t, updatedOrder.Status)
		require.Equal(t, status, *updatedOrder.Status)
	}

	listOrdersRecorder := doJSONRequest(t, server, http.MethodGet, "/api/v1/merchant/orders", "", accessToken)
	require.Equal(t, http.StatusOK, listOrdersRecorder.Code)
	orders := decodeJSONBody[[]OrderSummaryResponse](t, listOrdersRecorder)
	require.NotEmpty(t, orders)

	listInventoryRecorder := doJSONRequest(t, server, http.MethodGet, "/api/v1/merchant/inventory", "", accessToken)
	require.Equal(t, http.StatusOK, listInventoryRecorder.Code)
	inventoryItems := decodeJSONBody[[]InventoryItemResponse](t, listInventoryRecorder)
	require.NotEmpty(t, inventoryItems)

	reportRoute := fmt.Sprintf("/api/v1/merchant/reports/sales?month=%d&year=%d", time.Now().Month(), time.Now().Year())
	reportRecorder := doJSONRequest(t, server, http.MethodGet, reportRoute, "", accessToken)
	require.Equal(t, http.StatusOK, reportRecorder.Code)
	report := decodeJSONBody[SalesReportResponse](t, reportRecorder)
	require.NotNil(t, report.TotalSales)
	require.Greater(t, *report.TotalSales, 0.0)
}

func doJSONRequest(t *testing.T, server *HTTPAdapter, method string, route string, body string, accessToken string) *httptest.ResponseRecorder {
	t.Helper()

	var requestBody *bytes.Buffer
	if body == "" {
		requestBody = bytes.NewBuffer(nil)
	} else {
		requestBody = bytes.NewBufferString(body)
	}

	req, err := http.NewRequest(method, route, requestBody)
	require.NoError(t, err)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}

	recorder := httptest.NewRecorder()
	server.router.ServeHTTP(recorder, req)
	return recorder
}

func decodeJSONBody[T any](t *testing.T, recorder *httptest.ResponseRecorder) T {
	t.Helper()

	var response T
	err := json.NewDecoder(recorder.Body).Decode(&response)
	require.NoError(t, err)
	return response
}
