package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/ports"
	"github.com/horiondreher/go-web-api-boilerplate/internal/utils"

	"github.com/stretchr/testify/require"
)

type testActor struct {
	full_name string
	email     string
	password  string
}

func TestCreateActorV1(t *testing.T) {
	server, serverErr := NewHTTPAdapter(AdapterDependencies{
		ActorService:    testActorService,
		CommerceService: testReadService,
		MerchantService: testMerchantService,
		ReadService:     testReadService,
	})
	require.NoError(t, serverErr)

	authActor := testActor{
		full_name: utils.RandomString(6),
		email:     utils.RandomEmail(),
		password:  utils.RandomString(8),
	}

	createdAuthActor, createAuthActorErr := testActorService.CreateActor(context.Background(), ports.NewActor{
		MerchantID: testMerchantID,
		FullName:   authActor.full_name,
		Email:      authActor.email,
		Password:   authActor.password,
	})
	require.Nil(t, createAuthActorErr)

	accessToken, _, tokenErr := server.tokenMaker.CreateToken(authActor.email, "actor", testMerchantID, server.config.AccessTokenDuration)
	require.Nil(t, tokenErr)
	require.NotEqual(t, uuid.Nil, createdAuthActor.UID)

	actor := testActor{
		full_name: utils.RandomString(6),
		email:     utils.RandomEmail(),
		password:  utils.RandomString(6),
	}

	tt := []struct {
		name          string
		body          string
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "CreateActor",
			body: fmt.Sprintf(`{"full_name": "%s", "email": "%s", "password": "%s"}`, actor.full_name, actor.email, actor.password),
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusCreated, recorder.Code)
				validateActorResponse(t, actor, recorder.Body)
			},
		},
		{
			name: "CreateActorWithInvalidEmail",
			body: fmt.Sprintf(`{"full_name": "%s", "email": "%s", "password": "%s"}`, actor.full_name, "invalid_email", actor.password),
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
		{
			name: "CreateActorWithoutName",
			body: fmt.Sprintf(`{"email": "%s", "password": "%s"}`, actor.email, actor.password),
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
		{
			name: "CreateActorWithoutEmail",
			body: fmt.Sprintf(`{"full_name": "%s", "password": "%s"}`, actor.full_name, actor.password),
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
		{
			name: "CreateActorWithoutPasswordNoRole",
			body: fmt.Sprintf(`{"full_name": "%s", "email": "%s"}`, actor.full_name, actor.email),
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "CreateGuestCustomerWithoutPasswordNoMerchantRole",
			body: fmt.Sprintf(`{"full_name": "%s", "email": "guest_%s", "role": "customer"}`, actor.full_name, actor.email),
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name: "CreateActorWithoutAuthorization",
			body: fmt.Sprintf(`{"full_name": "%s", "email": "%s", "password": "%s"}`, actor.full_name, "noauth_"+actor.email, actor.password),
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "CreateActorWithEmptyBody",
			body: `{}`,
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
		{
			name: "CreateActorWithInvalidJson",
			body: `{"full_name": "invalid_json}`,
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "CreateActorWithEmptyJson",
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/api/v1/actors", bytes.NewBufferString(tc.body))
			require.NoError(t, err)
			if tc.name != "CreateActorWithoutAuthorization" {
				req.Header.Set("Authorization", "Bearer "+accessToken)
			}

			recorder := httptest.NewRecorder()

			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(recorder)
		})
	}
}

func validateActorResponse(t *testing.T, response testActor, body *bytes.Buffer) {
	var responseUser ActorProfileResponse
	err := json.NewDecoder(body).Decode(&responseUser)
	require.NoError(t, err)

	require.NotNil(t, responseUser.FullName)
	require.NotNil(t, responseUser.Email)
	require.NotNil(t, responseUser.Uid)
	require.NotNil(t, responseUser.MerchantId)

	require.Equal(t, response.full_name, *responseUser.FullName)
	require.Equal(t, response.email, string(*responseUser.Email))
	require.Equal(t, testMerchantID, uuid.UUID(*responseUser.MerchantId))

	require.NotZero(t, *responseUser.Uid)
	require.IsType(t, uuid.UUID{}, uuid.UUID(*responseUser.Uid))
}

func TestGetActorReadEndpointsV1(t *testing.T) {
	server, serverErr := NewHTTPAdapter(AdapterDependencies{
		ActorService:    testActorService,
		CommerceService: testReadService,
		MerchantService: testMerchantService,
		ReadService:     testReadService,
	})
	require.NoError(t, serverErr)

	authActor := testActor{
		full_name: utils.RandomString(6),
		email:     utils.RandomEmail(),
		password:  utils.RandomString(8),
	}
	createdActor, createActorErr := testActorService.CreateActor(context.Background(), ports.NewActor{
		MerchantID: testMerchantID,
		FullName:   authActor.full_name,
		Email:      authActor.email,
		Password:   authActor.password,
	})
	require.Nil(t, createActorErr)

	accessToken, _, tokenErr := server.tokenMaker.CreateToken(authActor.email, "actor", testMerchantID, server.config.AccessTokenDuration)
	require.Nil(t, tokenErr)

	t.Run("GetActorByUID", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/v1/actors/"+createdActor.UID.String(), nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		recorder := httptest.NewRecorder()
		server.router.ServeHTTP(recorder, req)
		require.Equal(t, http.StatusOK, recorder.Code)
	})

	t.Run("GetAuthenticatedActor", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/v1/actors/me", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		recorder := httptest.NewRecorder()
		server.router.ServeHTTP(recorder, req)
		require.Equal(t, http.StatusOK, recorder.Code)
	})
}
