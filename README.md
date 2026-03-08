
# Go Web API Boilerplate

Just writing some golang with things that I like to see in a Web API

## Features

- Hexagonal Architecture (kinda overengineering but ok. Also, just wrote like this to see how it goes)
- Simple routing with chi
- Centralized encoding and decoding
- Centralized error handling
- Versioned HTTP Handler
- SQL type safety with SQLC
- Migrations with golang migrate
- JWT authentication with symmetric secret key
- Access and Refresh Tokens
- Tests that uses Testcontainers instead of mocks
- Testing scripts that uses cURL and jq (f* Postman)

## JWT setup

Set:

- `TOKEN_SYMMETRIC_KEY=super-secret-jwt-key-32-characters`

## Required dependencies

- jq
- golang-migrate
- docker
- sqlc

## Local database

- The schema requires PostGIS because `zones.coordinates` uses `GEOMETRY(POLYGON, 4326)`.
- Local Docker runtime is configured to use `postgis/postgis:16-3.4`.
- `make` now starts the local PostGIS container automatically before running the API.
- `make run` / `make` now stops an older `go run cmd/http/main.go` instance first so reruns do not fail with a duplicate `:8080` bind error.
- If you previously started the project with the plain `postgres` image, recreate the container before running again:

```sh
docker compose down
docker compose up -d --wait postgres
```

## OpenAPI workflow (spec-first)

- Edit `openapi.yaml` as the source of truth.
- Run `make validate-openapi` to validate the spec.
- Run `make generate` to generate `internal/adapters/http/v1/openapi.gen.go`.

## Codebase rules

### 1) Architecture boundaries

- Keep business logic in `internal/domain/services`.
- Keep transport concerns (HTTP, decode/encode, middleware) in `internal/adapters/http`.
- Keep persistence concerns in `internal/adapters/pgsqlc`.
- Domain packages must not import adapter packages.
- Keep domains isolated: merchant logic must live in merchant-specific files/services, and actor logic must live in actor-specific files/services (no cross-domain mixing in the same file).

### 2) API contract and handlers

- `openapi.yaml` is the API source of truth.
- Workflow for API changes:
	1. Change `openapi.yaml`.
	2. Run `make generate`.
	3. Implement the endpoint behavior using generated code from `internal/adapters/http/v1/openapi.gen.go`.
- Do not manually edit generated files (`openapi.gen.go`).
- Keep endpoint behavior consistent with OpenAPI status codes and models.

### 3) Validation and error handling

- Validate request payloads before calling services.
- Return domain errors through centralized HTTP error encoding.
- Prefer typed/structured error responses over ad-hoc strings.
- Avoid leaking internal errors/messages to API clients.

### 4) Database and migrations

- Schema changes must be done via SQL migrations in `db/postgres/migration`.
- Keep `.up.sql` and `.down.sql` pairs consistent and reversible.
- After query changes, regenerate sqlc code (`sqlc generate` or `make sqlc`).

### 5) Testing requirements

- Follow TDD by default: write/adjust a failing test first (`red`), implement the minimal code to pass (`green`), then refactor while keeping tests green (`refactor`).
- For every behavior change, the PR must include at least one test that would fail without the code change.
- Add/update tests for behavior changes.
- Keep tests focused on API behavior and domain logic, not implementation details.
- Use existing Testcontainers integration for DB-backed tests.
- Ensure affected tests pass before merging.

### 6) Change discipline

- Prefer small, focused PRs.
- Do not mix refactors with feature/fix changes unless necessary.
- Preserve existing naming/style unless there is a clear reason to change it.
- Update README/OpenAPI/docs when behavior or workflow changes.

### 7) File structure conventions

- Follow the existing project file structure.
- Every new domain must be added in separate files (do not mix multiple domains in a single file).
- Keep domain-specific code grouped under `internal/domain` and corresponding adapters in their own files under `internal/adapters`.
- For HTTP v1 handlers, keep one domain per file (e.g., `merchant.go` for merchant endpoints, `actor.go` for actor endpoints).

### List of bugs (fixed)

- [x] Product creation: added `image_url` and `track_inventory` (default false) to CreateProductRequest
- [x] Cart response: removed merchant_id, updated_at, applied_discount_*, discount; addons nested in product; only created_at

### HTTP examples and hostname header

- [docs/HTTP_EXAMPLES.md](docs/HTTP_EXAMPLES.md) – Request/response examples for catalog, cart, checkout (no need to run the flow)
- [docs/HOSTNAME_MERCHANT_HEADER.md](docs/HOSTNAME_MERCHANT_HEADER.md) – Hostname-to-`X-Merchant-Id` design for multi-tenant frontends (e.g. sujing.com.pk)