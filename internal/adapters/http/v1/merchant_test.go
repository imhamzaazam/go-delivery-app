package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/ports"
	"github.com/horiondreher/go-web-api-boilerplate/internal/utils"

	"github.com/stretchr/testify/require"
)

type testMerchant struct {
	name          string
	ntn           string
	address       string
	category      string
	contactNumber string
}

func TestMerchantReadEndpointsV1(t *testing.T) {
	server, err := NewHTTPAdapter(AdapterDependencies{
		ActorService:    testActorService,
		CommerceService: testReadService,
		MerchantService: testMerchantService,
		ReadService:     testReadService,
	})
	require.NoError(t, err)

	authActor := testMerchant{
		name:          "Reader " + utils.RandomString(6),
		ntn:           "NTN-" + utils.RandomString(10),
		address:       "Reader Address",
		category:      "restaurant",
		contactNumber: "12345678901234",
	}

	actor, actorErr := testActorService.CreateActor(context.Background(), ports.NewActor{
		MerchantID: testMerchantID,
		FullName:   authActor.name,
		Email:      utils.RandomEmail(),
		Password:   utils.RandomString(8),
	})
	require.Nil(t, actorErr)

	accessToken, _, tokenErr := server.tokenMaker.CreateToken(actor.Email, "actor", testMerchantID, server.config.AccessTokenDuration)
	require.Nil(t, tokenErr)

	for _, route := range []string{"/api/v1/merchants", "/api/v1/merchant"} {
		req, reqErr := http.NewRequest("GET", route, nil)
		require.NoError(t, reqErr)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		recorder := httptest.NewRecorder()
		server.router.ServeHTTP(recorder, req)
		require.Equal(t, http.StatusOK, recorder.Code)
	}
}

func TestCreateMerchantV1(t *testing.T) {
	merchant := testMerchant{
		name:          "Merchant " + utils.RandomString(6),
		ntn:           "NTN-" + utils.RandomString(10),
		address:       "Test Address",
		category:      "restaurant",
		contactNumber: "12345678901234",
	}

	tt := []struct {
		name          string
		body          string
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "CreateMerchant",
			body: fmt.Sprintf(`{"name":"%s","ntn":"%s","address":"%s","category":"%s","contact_number":"%s"}`,
				merchant.name,
				merchant.ntn,
				merchant.address,
				merchant.category,
				merchant.contactNumber,
			),
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusCreated, recorder.Code)
			},
		},
		{
			name: "CreateMerchantWithInvalidCategory",
			body: fmt.Sprintf(`{"name":"%s","ntn":"%s","address":"%s","category":"invalid","contact_number":"%s"}`,
				merchant.name,
				"NTN-"+utils.RandomString(10),
				merchant.address,
				merchant.contactNumber,
			),
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/api/v1/merchants", bytes.NewBufferString(tc.body))
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			server, err := NewHTTPAdapter(AdapterDependencies{
				ActorService:    testActorService,
				CommerceService: testReadService,
				MerchantService: testMerchantService,
				ReadService:     testReadService,
			})
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(recorder)
		})
	}
}

func TestBootstrapMerchantActorFlowV1(t *testing.T) {
	server, err := NewHTTPAdapter(AdapterDependencies{
		ActorService:    testActorService,
		CommerceService: testReadService,
		MerchantService: testMerchantService,
		ReadService:     testReadService,
	})
	require.NoError(t, err)

	merchantBody := fmt.Sprintf(`{"name":"%s","ntn":"%s","address":"%s","category":"%s","contact_number":"%s"}`,
		"Merchant "+utils.RandomString(6),
		"NTN-"+utils.RandomString(10),
		"Bootstrap Address",
		"restaurant",
		"12345678901234",
	)

	createMerchantReq, reqErr := http.NewRequest("POST", "/api/v1/merchants", bytes.NewBufferString(merchantBody))
	require.NoError(t, reqErr)

	createMerchantRecorder := httptest.NewRecorder()
	server.router.ServeHTTP(createMerchantRecorder, createMerchantReq)
	require.Equal(t, http.StatusCreated, createMerchantRecorder.Code)

	var merchantResponse MerchantResponse
	decodeMerchantErr := json.NewDecoder(createMerchantRecorder.Body).Decode(&merchantResponse)
	require.NoError(t, decodeMerchantErr)
	require.NotNil(t, merchantResponse.Id)

	bootstrapEmail := utils.RandomEmail()
	bootstrapPassword := "Password#123"
	bootstrapBody := fmt.Sprintf(`{"full_name":"%s","email":"%s","password":"%s","role":"merchant"}`,
		"Owner "+utils.RandomString(6),
		bootstrapEmail,
		bootstrapPassword,
	)

	bootstrapReq, bootstrapReqErr := http.NewRequest("POST", "/api/v1/merchants/"+merchantResponse.Id.String()+"/bootstrap-actor", bytes.NewBufferString(bootstrapBody))
	require.NoError(t, bootstrapReqErr)

	bootstrapRecorder := httptest.NewRecorder()
	server.router.ServeHTTP(bootstrapRecorder, bootstrapReq)
	require.Equal(t, http.StatusCreated, bootstrapRecorder.Code)

	var actorResponse ActorProfileResponse
	decodeActorErr := json.NewDecoder(bootstrapRecorder.Body).Decode(&actorResponse)
	require.NoError(t, decodeActorErr)
	require.NotNil(t, actorResponse.MerchantId)
	require.Equal(t, merchantResponse.Id.String(), actorResponse.MerchantId.String())

	loginBody := fmt.Sprintf(`{"merchant_id":"%s","email":"%s","password":"%s"}`,
		merchantResponse.Id.String(),
		bootstrapEmail,
		bootstrapPassword,
	)

	loginReq, loginReqErr := http.NewRequest("POST", "/api/v1/login", bytes.NewBufferString(loginBody))
	require.NoError(t, loginReqErr)

	loginRecorder := httptest.NewRecorder()
	server.router.ServeHTTP(loginRecorder, loginReq)
	require.Equal(t, http.StatusOK, loginRecorder.Code)

	repeatBootstrapReq, repeatBootstrapReqErr := http.NewRequest("POST", "/api/v1/merchants/"+merchantResponse.Id.String()+"/bootstrap-actor", bytes.NewBufferString(bootstrapBody))
	require.NoError(t, repeatBootstrapReqErr)

	repeatBootstrapRecorder := httptest.NewRecorder()
	server.router.ServeHTTP(repeatBootstrapRecorder, repeatBootstrapReq)
	require.Equal(t, http.StatusConflict, repeatBootstrapRecorder.Code)
}
