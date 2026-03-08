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
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/pgsqlc"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/ports"
	"github.com/stretchr/testify/require"
)

func TestCreateCoverageEndpointsV1(t *testing.T) {
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
		FullName:   "Coverage Merchant",
		Email:      fmt.Sprintf("coverage-%s@test.local", uuid.NewString()),
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

	createAreaReq, reqErr := http.NewRequest(http.MethodPost, "/api/v1/areas", bytes.NewBufferString(`{"name":"Coverage Area","city":"Karachi"}`))
	require.NoError(t, reqErr)
	createAreaReq.Header.Set("Authorization", "Bearer "+accessToken)
	createAreaReq.Header.Set("Content-Type", "application/json")
	createAreaRecorder := httptest.NewRecorder()
	server.router.ServeHTTP(createAreaRecorder, createAreaReq)
	require.Equal(t, http.StatusCreated, createAreaRecorder.Code)

	var area AreaResponse
	decodeAreaErr := json.NewDecoder(createAreaRecorder.Body).Decode(&area)
	require.NoError(t, decodeAreaErr)
	require.NotNil(t, area.Id)
	require.Equal(t, "Coverage Area", *area.Name)

	createZoneBody := `{"name":"Coverage Zone","coordinates_wkt":"POLYGON((67.10 24.10, 67.11 24.10, 67.11 24.11, 67.10 24.11, 67.10 24.10))"}`
	createZoneReq, zoneReqErr := http.NewRequest(http.MethodPost, "/api/v1/areas/"+uuid.UUID(*area.Id).String()+"/zones", bytes.NewBufferString(createZoneBody))
	require.NoError(t, zoneReqErr)
	createZoneReq.Header.Set("Authorization", "Bearer "+accessToken)
	createZoneReq.Header.Set("Content-Type", "application/json")
	createZoneRecorder := httptest.NewRecorder()
	server.router.ServeHTTP(createZoneRecorder, createZoneReq)
	require.Equal(t, http.StatusCreated, createZoneRecorder.Code)

	var zone ZoneResponse
	decodeZoneErr := json.NewDecoder(createZoneRecorder.Body).Decode(&zone)
	require.NoError(t, decodeZoneErr)
	require.NotNil(t, zone.Id)
	require.Equal(t, "Coverage Zone", *zone.Name)
	require.NotNil(t, zone.CoordinatesWkt)
	require.Contains(t, *zone.CoordinatesWkt, "POLYGON")

	branchReq, branchReqErr := http.NewRequest(http.MethodPost, "/api/v1/merchant/branches", bytes.NewBufferString(`{"name":"Coverage Branch","address":"Coverage Address","contact_number":"02100000000000","city":"Karachi"}`))
	require.NoError(t, branchReqErr)
	branchReq.Header.Set("Authorization", "Bearer "+accessToken)
	branchReq.Header.Set("Content-Type", "application/json")
	branchRecorder := httptest.NewRecorder()
	server.router.ServeHTTP(branchRecorder, branchReq)
	require.Equal(t, http.StatusCreated, branchRecorder.Code)

	var branch BranchResponse
	decodeBranchErr := json.NewDecoder(branchRecorder.Body).Decode(&branch)
	require.NoError(t, decodeBranchErr)
	require.NotNil(t, branch.Id)

	attachReq, attachReqErr := http.NewRequest(http.MethodPost, "/api/v1/merchant/service-zones", bytes.NewBufferString(fmt.Sprintf(`{"zone_id":"%s","branch_id":"%s"}`, uuid.UUID(*zone.Id), uuid.UUID(*branch.Id))))
	require.NoError(t, attachReqErr)
	attachReq.Header.Set("Authorization", "Bearer "+accessToken)
	attachReq.Header.Set("Content-Type", "application/json")
	attachRecorder := httptest.NewRecorder()
	server.router.ServeHTTP(attachRecorder, attachReq)
	require.Equal(t, http.StatusCreated, attachRecorder.Code)

	listZonesReq, listReqErr := http.NewRequest(http.MethodGet, "/api/v1/areas/"+uuid.UUID(*area.Id).String()+"/zones", nil)
	require.NoError(t, listReqErr)
	listZonesReq.Header.Set("Authorization", "Bearer "+accessToken)
	listZonesRecorder := httptest.NewRecorder()
	server.router.ServeHTTP(listZonesRecorder, listZonesReq)
	require.Equal(t, http.StatusOK, listZonesRecorder.Code)

	coveredReq, coveredReqErr := http.NewRequest(http.MethodPost, "/api/v1/merchant/service-zones/check", bytes.NewBufferString(`{"latitude":24.1050,"longitude":67.1050}`))
	require.NoError(t, coveredReqErr)
	coveredReq.Header.Set("Authorization", "Bearer "+accessToken)
	coveredReq.Header.Set("Content-Type", "application/json")
	coveredRecorder := httptest.NewRecorder()
	server.router.ServeHTTP(coveredRecorder, coveredReq)
	require.Equal(t, http.StatusOK, coveredRecorder.Code)

	var covered ServiceZoneCoverageCheckResponse
	decodeCoveredErr := json.NewDecoder(coveredRecorder.Body).Decode(&covered)
	require.NoError(t, decodeCoveredErr)
	require.NotNil(t, covered.Covered)
	require.True(t, *covered.Covered)
	require.NotNil(t, covered.ZoneId)
	require.Equal(t, uuid.UUID(*zone.Id), uuid.UUID(*covered.ZoneId))

	outsideReq, outsideReqErr := http.NewRequest(http.MethodPost, "/api/v1/merchant/service-zones/check", bytes.NewBufferString(`{"latitude":24.3000,"longitude":67.3000}`))
	require.NoError(t, outsideReqErr)
	outsideReq.Header.Set("Authorization", "Bearer "+accessToken)
	outsideReq.Header.Set("Content-Type", "application/json")
	outsideRecorder := httptest.NewRecorder()
	server.router.ServeHTTP(outsideRecorder, outsideReq)
	require.Equal(t, http.StatusOK, outsideRecorder.Code)

	var outside ServiceZoneCoverageCheckResponse
	decodeOutsideErr := json.NewDecoder(outsideRecorder.Body).Decode(&outside)
	require.NoError(t, decodeOutsideErr)
	require.NotNil(t, outside.Covered)
	require.False(t, *outside.Covered)
}
