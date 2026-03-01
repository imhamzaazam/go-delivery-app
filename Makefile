run:
	go run cmd/http/main.go

build:
	@go build -o bin/go-boilerplate cmd/http/main.go

test:
	@go test -v ./...

openapi:
	@echo "Using openapi.yaml"

# Validates openapi.yaml against OpenAPI schema.
validate-openapi:
	go run github.com/getkin/kin-openapi/cmd/validate@latest --defaults openapi.yaml

# Generates typed server/spec code from openapi.yaml.
generate: validate-openapi
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.4.1 \
		--package=v1 --generate=types,chi-server,spec,skip-prune \
		-o internal/adapters/http/v1/openapi.gen.go openapi.yaml

sqlc:
	sqlc generate

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

.PHONY:
	run build test openapi validate-openapi generate sqlc migrateup migrateup1 migratedown migratedown1 new_migration