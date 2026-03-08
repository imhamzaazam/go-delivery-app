package v1

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type merchantPayloadCollection struct {
	Headers struct {
		ContentType   string `json:"content_type"`
		Authorization string `json:"authorization"`
	} `json:"headers"`
	CreateMerchant struct {
		Request          json.RawMessage `json:"request"`
		ExpectedResponse json.RawMessage `json:"expected_response"`
	} `json:"create_merchant"`
	BootstrapActor struct {
		Request json.RawMessage `json:"request"`
	} `json:"bootstrap_actor"`
	LoginActor struct {
		Request json.RawMessage `json:"request"`
	} `json:"login_actor"`
	UpdateMerchant struct {
		Request          json.RawMessage `json:"request"`
		ExpectedResponse json.RawMessage `json:"expected_response"`
	} `json:"update_merchant"`
	GetMerchant struct {
		ExpectedResponse json.RawMessage `json:"expected_response"`
	} `json:"get_merchant"`
	ListMerchants struct {
		ExpectedResponse json.RawMessage `json:"expected_response"`
	} `json:"list_merchants"`
}

func TestMerchantPayloadSnapshotsV1(t *testing.T) {
	server, err := NewHTTPAdapter(AdapterDependencies{
		ActorService:    testActorService,
		CommerceService: testReadService,
		MerchantService: testMerchantService,
		ReadService:     testReadService,
	})
	require.NoError(t, err)

	payloads := loadMerchantPayloads(t)

	createRecorder := doJSONRequestWithHeaders(t, server, http.MethodPost, "/api/v1/merchants", string(payloads.CreateMerchant.Request), map[string]string{
		"Content-Type": payloads.Headers.ContentType,
	})
	require.Equal(t, http.StatusCreated, createRecorder.Code)
	createResponseBody := createRecorder.Body.String()

	createdMerchant := decodeJSONBody[MerchantResponse](t, createRecorder)
	require.NotNil(t, createdMerchant.Id)
	require.NotEmpty(t, createdMerchant.Id.String())
	require.JSONEq(t,
		renderPayloadTemplate(t, payloads.CreateMerchant.ExpectedResponse, map[string]string{
			"merchant_id": createdMerchant.Id.String(),
		}),
		createResponseBody,
	)

	bootstrapRecorder := doJSONRequestWithHeaders(t, server, http.MethodPost, "/api/v1/merchants/"+createdMerchant.Id.String()+"/bootstrap-actor", string(payloads.BootstrapActor.Request), map[string]string{
		"Content-Type": payloads.Headers.ContentType,
	})
	require.Equal(t, http.StatusCreated, bootstrapRecorder.Code)

	bootstrappedActor := decodeJSONBody[ActorProfileResponse](t, bootstrapRecorder)
	require.NotNil(t, bootstrappedActor.MerchantId)
	require.Equal(t, createdMerchant.Id.String(), bootstrappedActor.MerchantId.String())
	require.NotNil(t, bootstrappedActor.Uid)

	loginRecorder := doJSONRequestWithHeaders(t, server, http.MethodPost, "/api/v1/login", renderPayloadTemplate(t, payloads.LoginActor.Request, map[string]string{
		"merchant_id": createdMerchant.Id.String(),
	}), map[string]string{
		"Content-Type": payloads.Headers.ContentType,
	})
	require.Equal(t, http.StatusOK, loginRecorder.Code)

	loginResponse := decodeJSONBody[LoginActorResponse](t, loginRecorder)
	require.NotNil(t, loginResponse.AccessToken)
	require.NotEmpty(t, *loginResponse.AccessToken)
	require.NotNil(t, loginResponse.RefreshToken)
	require.NotEmpty(t, *loginResponse.RefreshToken)
	require.NotNil(t, loginResponse.MerchantId)
	require.Equal(t, createdMerchant.Id.String(), loginResponse.MerchantId.String())

	authHeader := strings.ReplaceAll(payloads.Headers.Authorization, "{{access_token}}", *loginResponse.AccessToken)

	getRecorder := doJSONRequestWithHeaders(t, server, http.MethodGet, "/api/v1/merchant", "", map[string]string{
		"Authorization": authHeader,
	})
	require.Equal(t, http.StatusOK, getRecorder.Code)

	getMerchant := decodeJSONBody[MerchantResponse](t, getRecorder)
	require.NotNil(t, getMerchant.Id)
	require.Equal(t, createdMerchant.Id.String(), getMerchant.Id.String())
	require.NotNil(t, getMerchant.CreatedAt)
	require.NotNil(t, getMerchant.UpdatedAt)
	require.True(t, getMerchant.UpdatedAt.Equal(*getMerchant.CreatedAt) || getMerchant.UpdatedAt.After(*getMerchant.CreatedAt))

	updateRecorder := doJSONRequestWithHeaders(t, server, http.MethodPatch, "/api/v1/merchant", string(payloads.UpdateMerchant.Request), map[string]string{
		"Authorization": authHeader,
		"Content-Type":  payloads.Headers.ContentType,
	})
	require.Equal(t, http.StatusOK, updateRecorder.Code)
	updateResponseBody := updateRecorder.Body.String()

	updatedMerchant := decodeJSONBody[MerchantResponse](t, updateRecorder)
	require.NotNil(t, updatedMerchant.Id)
	require.Equal(t, createdMerchant.Id.String(), updatedMerchant.Id.String())
	require.NotNil(t, updatedMerchant.CreatedAt)
	require.NotNil(t, updatedMerchant.UpdatedAt)
	require.True(t, updatedMerchant.UpdatedAt.Equal(*updatedMerchant.CreatedAt) || updatedMerchant.UpdatedAt.After(*updatedMerchant.CreatedAt))
	require.JSONEq(t,
		renderPayloadTemplate(t, payloads.UpdateMerchant.ExpectedResponse, map[string]string{
			"merchant_id": createdMerchant.Id.String(),
			"created_at":  updatedMerchant.CreatedAt.Format(time.RFC3339Nano),
			"updated_at":  updatedMerchant.UpdatedAt.Format(time.RFC3339Nano),
		}),
		updateResponseBody,
	)

	getUpdatedRecorder := doJSONRequestWithHeaders(t, server, http.MethodGet, "/api/v1/merchant", "", map[string]string{
		"Authorization": authHeader,
	})
	require.Equal(t, http.StatusOK, getUpdatedRecorder.Code)
	require.JSONEq(t,
		renderPayloadTemplate(t, payloads.GetMerchant.ExpectedResponse, map[string]string{
			"merchant_id": createdMerchant.Id.String(),
			"created_at":  updatedMerchant.CreatedAt.Format(time.RFC3339Nano),
			"updated_at":  updatedMerchant.UpdatedAt.Format(time.RFC3339Nano),
		}),
		getUpdatedRecorder.Body.String(),
	)

	listRecorder := doJSONRequestWithHeaders(t, server, http.MethodGet, "/api/v1/merchants", "", map[string]string{
		"Authorization": authHeader,
	})
	require.Equal(t, http.StatusOK, listRecorder.Code)
	require.JSONEq(t,
		renderPayloadTemplate(t, payloads.ListMerchants.ExpectedResponse, map[string]string{
			"merchant_id": createdMerchant.Id.String(),
			"created_at":  updatedMerchant.CreatedAt.Format(time.RFC3339Nano),
			"updated_at":  updatedMerchant.UpdatedAt.Format(time.RFC3339Nano),
		}),
		listRecorder.Body.String(),
	)
}

func loadMerchantPayloads(t *testing.T) merchantPayloadCollection {
	t.Helper()

	filePath := filepath.Join("testdata", "merchant_api_payloads.json")
	contents, err := os.ReadFile(filePath)
	require.NoError(t, err)

	var payloads merchantPayloadCollection
	err = json.Unmarshal(contents, &payloads)
	require.NoError(t, err)

	return payloads
}

func renderPayloadTemplate(t *testing.T, template json.RawMessage, replacements map[string]string) string {
	t.Helper()

	rendered := string(template)
	for key, value := range replacements {
		rendered = strings.ReplaceAll(rendered, "{{"+key+"}}", value)
	}

	var normalized any
	err := json.Unmarshal([]byte(rendered), &normalized)
	require.NoError(t, err)

	output, err := json.Marshal(normalized)
	require.NoError(t, err)

	return string(output)
}

func doJSONRequestWithHeaders(t *testing.T, server *HTTPAdapter, method string, route string, body string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()

	req, err := http.NewRequest(method, route, strings.NewReader(body))
	require.NoError(t, err)

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	recorder := httptest.NewRecorder()
	server.router.ServeHTTP(recorder, req)
	return recorder
}
