ifeq ($(OS),Windows_NT)
STOP_APP_CMD=powershell -NoProfile -Command "$$ErrorActionPreference = 'SilentlyContinue'; $$procs = Get-CimInstance Win32_Process | Where-Object { $$_.Name -eq 'go.exe' -and $$_.CommandLine -like '*cmd/http/main.go*' }; foreach ($$proc in $$procs) { taskkill /PID $$proc.ProcessId /T /F | Out-Null }; $$listeners = Get-NetTCPConnection -LocalPort 8080 -State Listen | Select-Object -ExpandProperty OwningProcess -Unique; foreach ($$pid in $$listeners) { taskkill /PID $$pid /T /F | Out-Null }; exit 0"
else
STOP_APP_CMD=pkill -f "go run cmd/http/main.go" || true; fuser -k 8080/tcp >/dev/null 2>&1 || true
endif

run: stop-app run-services
	go run cmd/http/main.go

stop-app:
	@$(STOP_APP_CMD)

run-services:
	docker compose up -d --wait postgres

stop-services:
	docker compose stop postgres

build:
	@go build -o bin/go-boilerplate cmd/http/main.go

test:
	@go test -v ./...

test-http:
	@go test -v ./internal/http/v1/...

test-http-domains:
	@go test -v ./internal/http/v1/actor ./internal/http/v1/auth ./internal/http/v1/cart ./internal/http/v1/catalog ./internal/http/v1/coverage ./internal/http/v1/merchant ./internal/http/v1/order ./internal/http/v1/report

test-http-flow:
	@go test -v ./internal/http/v1/... -run "(Flow|Payload|Snapshots)"

test-http-e2e:
	@go test -v ./internal/http/v1 -run "TestHTTP_"

test-http-smoke:
	@go test -v ./internal/http/v1 -run "TestReadSurfaceSmokeV1"

test-http-regression:
	@go test -v ./internal/http/v1 -run "(Regression|SimplifiedStructure|AcceptsImageUrlAndTrackInventory|CartResponse|CreateCartResponse|AddItemToCartResponse)"

test-domain:
	@go test -v ./internal/domain/...

test-fast:
	@go test -v ./internal/domain/... ./internal/http/v1/actor ./internal/http/v1/auth ./internal/http/v1/cart ./internal/http/v1/catalog ./internal/http/v1/coverage ./internal/http/v1/merchant ./internal/http/v1/order ./internal/http/v1/report

generate:
	$(MAKE) -C api generate

sqlc:
	sqlc generate

DB_SOURCE ?= postgresql://pguser:pgpassword@localhost:5432/go_boilerplate?sslmode=disable

migrateup:
	migrate -path db/postgres/migration -database "$(DB_SOURCE)" -verbose up

migrateup1:
	migrate -path db/postgres/migration -database "$(DB_SOURCE)" -verbose up 1

migratedown:
	migrate -path db/postgres/migration -database "$(DB_SOURCE)" -verbose down

migratedown1:
	migrate -path db/postgres/migration -database "$(DB_SOURCE)" -verbose down 1

new_migration:
	migrate create -ext sql -dir db/postgres/migration -seq $(name)

.PHONY: run stop-app run-services stop-services build test test-http test-http-domains test-http-flow test-http-e2e test-http-smoke test-http-regression test-domain test-fast generate sqlc migrateup migrateup1 migratedown migratedown1 new_migration