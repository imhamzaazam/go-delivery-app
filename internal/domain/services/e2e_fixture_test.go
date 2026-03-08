package services

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/pgsqlc"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/ports"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type commerceFixture struct {
	ctx             context.Context
	store           *pgsqlc.Queries
	actorService    *ActorManager
	commerceService *CommerceManager
	adminActorID    uuid.UUID
	merchantID      uuid.UUID
	merchantOwnerID uuid.UUID
	merchantBranch  uuid.UUID
	customerID      uuid.UUID
	cartID          uuid.UUID
}

func setupCommerceFixture(t *testing.T) *commerceFixture {
	t.Helper()

	ctx := context.Background()
	pool := setupServiceTestDB(t, ctx)
	store := pgsqlc.New(pool)

	actorService := NewActorManager(store)
	merchantService := NewMerchantManager(pool, store)
	commerceService := NewCommerceManager(pool, store)

	platformMerchant, platformErr := merchantService.CreateMerchant(ctx, ports.NewMerchant{
		Name:          "Platform HQ",
		Ntn:           "NTN-PLATFORM-001",
		Address:       "HQ Address",
		Category:      "restaurant",
		ContactNumber: "03111111111111",
	})
	require.Nil(t, platformErr)

	_, branchErr := store.CreateBranch(ctx, pgsqlc.CreateBranchParams{
		MerchantID:    platformMerchant.ID,
		Name:          "HQ Branch",
		Address:       "HQ Branch Address",
		ContactNumber: textValue("02100000000000"),
		City:          pgsqlc.CityTypeKarachi,
	})
	require.NoError(t, branchErr)

	adminRole, adminRoleErr := store.CreateRole(ctx, pgsqlc.CreateRoleParams{
		MerchantID:  platformMerchant.ID,
		RoleType:    pgsqlc.RoleTypeAdmin,
		Description: textValue("Platform admin"),
	})
	require.NoError(t, adminRoleErr)

	adminActor, adminActorErr := actorService.CreateActor(ctx, ports.NewActor{
		MerchantID: platformMerchant.ID,
		FullName:   "Platform Admin",
		Email:      "admin@platform.test",
		Password:   "Password#123",
	})
	require.Nil(t, adminActorErr)

	_, assignAdminErr := store.AssignActorRole(ctx, pgsqlc.AssignActorRoleParams{
		MerchantID: platformMerchant.ID,
		ActorID:    adminActor.UID,
		RoleID:     adminRole.ID,
	})
	require.NoError(t, assignAdminErr)

	merchant, merchantErr := commerceService.CreateMerchantByAdmin(ctx, adminActor.UID, ports.NewMerchant{
		Name:          "Suijing Tenant",
		Ntn:           "NTN-SUIJING-900",
		Address:       "DHA Karachi",
		Category:      "restaurant",
		ContactNumber: "03222222222222",
	})
	require.Nil(t, merchantErr)

	merchantBranch, merchantBranchErr := store.CreateBranch(ctx, pgsqlc.CreateBranchParams{
		MerchantID:    merchant.ID,
		Name:          "Main Branch",
		Address:       "Main Branch Address",
		ContactNumber: textValue("02133333333333"),
		City:          pgsqlc.CityTypeKarachi,
	})
	require.NoError(t, merchantBranchErr)

	merchantOwner, merchantOwnerErr := actorService.CreateActor(ctx, ports.NewActor{
		MerchantID: merchant.ID,
		FullName:   "Merchant Owner",
		Email:      "owner@suijing.test",
		Password:   "Password#123",
	})
	require.Nil(t, merchantOwnerErr)

	merchantRoleID, merchantRoleErr := commerceService.getRoleIDByType(ctx, merchant.ID, pgsqlc.RoleTypeMerchant)
	require.NoError(t, merchantRoleErr)

	_, assignMerchantErr := store.AssignActorRole(ctx, pgsqlc.AssignActorRoleParams{
		MerchantID: merchant.ID,
		ActorID:    merchantOwner.UID,
		RoleID:     merchantRoleID,
	})
	require.NoError(t, assignMerchantErr)

	customer, customerErr := actorService.CreateActor(ctx, ports.NewActor{
		MerchantID: merchant.ID,
		FullName:   "End Customer",
		Email:      "customer@suijing.test",
		Password:   "Password#123",
	})
	require.Nil(t, customerErr)

	customerRoleID, customerRoleErr := commerceService.getRoleIDByType(ctx, merchant.ID, pgsqlc.RoleTypeCustomer)
	require.NoError(t, customerRoleErr)

	_, assignCustomerErr := store.AssignActorRole(ctx, pgsqlc.AssignActorRoleParams{
		MerchantID: merchant.ID,
		ActorID:    customer.UID,
		RoleID:     customerRoleID,
	})
	require.NoError(t, assignCustomerErr)

	_, sessionErr := actorService.CreateActorSession(ctx, ports.NewActorSession{
		RefreshTokenID:        uuid.New(),
		MerchantID:            merchant.ID,
		ActorID:               customer.UID,
		RefreshToken:          "rt_customer_fixture_001",
		UserAgent:             "e2e-test",
		ClientIP:              "127.0.0.1",
		RefreshTokenExpiresAt: time.Now().Add(24 * time.Hour),
	})
	require.Nil(t, sessionErr)

	cart, cartErr := commerceService.CreateCart(ctx, uuid.New(), merchant.ID, merchantBranch.ID, customer.UID)
	require.Nil(t, cartErr)

	return &commerceFixture{
		ctx:             ctx,
		store:           store,
		actorService:    actorService,
		commerceService: commerceService,
		adminActorID:    adminActor.UID,
		merchantID:      merchant.ID,
		merchantOwnerID: merchantOwner.UID,
		merchantBranch:  merchantBranch.ID,
		customerID:      customer.UID,
		cartID:          cart.ID,
	}
}

func setupServiceTestDB(t *testing.T, ctx context.Context) *pgxpool.Pool {
	t.Helper()

	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgis/postgis:16-3.4"),
		postgres.WithDatabase("go_boilerplate"),
		postgres.WithUsername("pguser"),
		postgres.WithPassword("pgpassword"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(10*time.Second),
		),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = pgContainer.Terminate(ctx)
	})

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)

	_, currentFile, _, ok := runtime.Caller(0)
	require.True(t, ok)

	rootDir := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..", ".."))
	migrationPaths, globErr := filepath.Glob(filepath.Join(rootDir, "db", "postgres", "migration", "*.up.sql"))
	require.NoError(t, globErr)
	require.NotEmpty(t, migrationPaths)
	sort.Strings(migrationPaths)

	for _, migrationPath := range migrationPaths {
		_, statErr := os.Stat(migrationPath)
		require.NoError(t, statErr)

		migrationSQL, readErr := os.ReadFile(migrationPath)
		require.NoError(t, readErr)

		_, execErr := pool.Exec(ctx, string(migrationSQL))
		require.NoError(t, execErr)
	}

	t.Cleanup(func() {
		pool.Close()
	})

	return pool
}
