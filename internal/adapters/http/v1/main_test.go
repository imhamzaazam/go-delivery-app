package v1

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/pgsqlc"
	service "github.com/horiondreher/go-web-api-boilerplate/internal/domain/services"
	"github.com/horiondreher/go-web-api-boilerplate/internal/utils"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/jackc/pgx/v5/pgxpool"
)

var testActorService *service.ActorManager
var testMerchantService *service.MerchantManager
var testReadService *service.CommerceManager
var testStore *pgsqlc.Queries
var testMerchantID uuid.UUID

func TestMain(m *testing.M) {
	ctx := context.Background()

	setEnvIfUnset("ENVIRONMENT", "test")
	setEnvIfUnset("HTTP_SERVER_ADDRESS", "0.0.0.0:8080")
	setEnvIfUnset("POSTGRES_DB", "go_boilerplate")
	setEnvIfUnset("POSTGRES_USER", "pguser")
	setEnvIfUnset("POSTGRES_PASSWORD", "pgpassword")
	setEnvIfUnset("DB_SOURCE", "postgresql://pguser:pgpassword@localhost:5432/go_boilerplate?sslmode=disable")
	setEnvIfUnset("MIGRATION_URL", "file://db/postgres/migration")
	setEnvIfUnset("TOKEN_SYMMETRIC_KEY", "test-super-secret-jwt-key-32-chars")
	setEnvIfUnset("ACCESS_TOKEN_DURATION", "15m")
	setEnvIfUnset("REFRESH_TOKEN_DURATION", "24h")

	utils.SetConfigFile("../../../../.env")
	config := utils.GetConfig()

	migrationsPath := filepath.Join("..", "..", "..", "..", "db", "postgres", "migration", "*.up.sql")
	upMigrations, err := filepath.Glob(migrationsPath)
	if err != nil {
		log.Fatalf("cannot find up migrations: %v", err)
	}

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
	if err != nil {
		log.Fatalf("cannot start postgres container: %v", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("cannot get connection string: %v", err)
	}

	conn, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("cannot connect to database: %v", err)
	}
	defer conn.Close()

	testStore = pgsqlc.New(conn)

	merchant, err := testStore.CreateMerchant(ctx, pgsqlc.CreateMerchantParams{
		Name:          "Test Merchant",
		Ntn:           "TEST-NTN-0001",
		Address:       "Test Address",
		Category:      pgsqlc.MerchantCategoryRestaurant,
		ContactNumber: "12345678901234",
	})
	if err != nil {
		log.Fatalf("cannot seed merchant: %v", err)
	}

	if merchant.ID.String() == "" {
		log.Fatalf("seeded merchant has empty id")
	}

	if merchant.Name != "Test Merchant" {
		log.Fatalf("seeded merchant has unexpected name: %s", merchant.Name)
	}

	if merchant.Ntn != "TEST-NTN-0001" {
		log.Fatalf("seeded merchant has unexpected ntn: %s", merchant.Ntn)
	}

	testMerchantID = merchant.ID

	testActorService = service.NewActorManager(testStore)
	testMerchantService = service.NewMerchantManager(conn, testStore)
	testReadService = service.NewCommerceManager(conn, testStore)

	os.Exit(m.Run())
}

func setEnvIfUnset(name string, value string) {
	if _, exists := os.LookupEnv(name); exists {
		return
	}

	if err := os.Setenv(name, value); err != nil {
		log.Fatalf("cannot set env var %s: %v", name, err)
	}
}
