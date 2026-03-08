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

type testLoginActor struct {
	full_name string
	email     string
	password  string
}

func TestLoginActor(t *testing.T) {
	actor := testLoginActor{
		full_name: utils.RandomString(6),
		email:     utils.RandomEmail(),
		password:  utils.RandomString(6),
	}

	_, err := testActorService.CreateActor(context.Background(), ports.NewActor{
		MerchantID: testMerchantID,
		FullName:   actor.full_name,
		Email:      actor.email,
		Password:   actor.password,
	})
	require.Nil(t, err)

	tt := []struct {
		name          string
		body          string
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "LoginActor",
			body: fmt.Sprintf(`{"merchant_id": "%s", "email": "%s", "password": "%s"}`, testMerchantID, actor.email, actor.password),
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "LoginActorWithInvalidEmail",
			body: fmt.Sprintf(`{"merchant_id": "%s", "email": "%s", "password": "%s"}`, testMerchantID, "invalid_email", actor.password),
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
		{
			name: "LoginActorWithoutEmail",
			body: fmt.Sprintf(`{"merchant_id": "%s", "password": "%s"}`, testMerchantID, actor.password),
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
		{
			name: "LoginActorWithEmptyPassword",
			body: fmt.Sprintf(`{"merchant_id": "%s", "email": "%s"}`, testMerchantID, actor.email),
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
		{
			name: "LoginActorWithEmptyBody",
			body: `{}`,
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
		{
			name: "LoginActorWithInvalidJson",
			body: `{"email": "invalid_json}`,
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "LoginActorWithEmptyJson",
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			req, reqErr := http.NewRequest("POST", "/api/v1/login", bytes.NewBufferString(tc.body))
			require.NoError(t, reqErr)

			recorder := httptest.NewRecorder()
			server, serverErr := NewHTTPAdapter(AdapterDependencies{
				ActorService:    testActorService,
				CommerceService: testReadService,
				MerchantService: testMerchantService,
				ReadService:     testReadService,
			})
			require.NoError(t, serverErr)

			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(recorder)
		})
	}
}

func TestRenewAccessToken(t *testing.T) {
	actor := testLoginActor{
		full_name: utils.RandomString(6),
		email:     utils.RandomEmail(),
		password:  utils.RandomString(8),
	}

	_, err := testActorService.CreateActor(context.Background(), ports.NewActor{
		MerchantID: testMerchantID,
		FullName:   actor.full_name,
		Email:      actor.email,
		Password:   actor.password,
	})
	require.Nil(t, err)

	server, serverErr := NewHTTPAdapter(AdapterDependencies{
		ActorService:    testActorService,
		CommerceService: testReadService,
		MerchantService: testMerchantService,
		ReadService:     testReadService,
	})
	require.NoError(t, serverErr)

	loginReq, reqErr := http.NewRequest("POST", "/api/v1/login", bytes.NewBufferString(fmt.Sprintf(`{"merchant_id": "%s", "email": "%s", "password": "%s"}`, testMerchantID, actor.email, actor.password)))
	require.NoError(t, reqErr)

	loginRecorder := httptest.NewRecorder()
	server.router.ServeHTTP(loginRecorder, loginReq)
	require.Equal(t, http.StatusOK, loginRecorder.Code)

	var loginResponse LoginActorResponse
	decodeErr := json.NewDecoder(loginRecorder.Body).Decode(&loginResponse)
	require.NoError(t, decodeErr)
	require.NotNil(t, loginResponse.RefreshToken)

	renewReq, renewReqErr := http.NewRequest("POST", "/api/v1/renew-token", bytes.NewBufferString(fmt.Sprintf(`{"refresh_token": "%s"}`, *loginResponse.RefreshToken)))
	require.NoError(t, renewReqErr)

	renewRecorder := httptest.NewRecorder()
	server.router.ServeHTTP(renewRecorder, renewReq)
	require.Equal(t, http.StatusOK, renewRecorder.Code)
}
