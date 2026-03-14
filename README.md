
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

- Edit `api/openapi.yaml`.
- Run `make generate -C api`.
- That regenerates [api/openapi.gen.go](d:/projects/go-delivery-app/api/openapi.gen.go).

## Test commands

- `make test` runs the full repository test suite.
- `make test-fast` runs the fast local suite: domain tests plus HTTP domain-package tests.
- `make test-domain` runs domain-layer tests only.
- `make test-http` runs all HTTP adapter tests under `internal/http/v1/...`.
- `make test-http-domains` runs the per-domain HTTP adapter packages only.
- `make test-http-flow` runs HTTP flow and payload snapshot tests using current naming conventions.
- `make test-http-e2e` runs root-level HTTP end-to-end flow tests.
- `make test-http-smoke` runs the root-level HTTP smoke test.
- `make test-http-regression` runs the root-level HTTP regression-focused tests.

These targets follow the current naming/layout conventions in the repository. As the HTTP adapter packages continue to move toward the preferred domain-first structure, keep new flow tests in `*_flow_test.go` and keep root-level cross-domain checks in the dedicated E2E and smoke files.

## Target architecture

New code should follow the DDD + Hexagonal structure defined in [.cursorrules](d:/projects/go-delivery-app/.cursorrules).

Dependency direction must remain:

- HTTP -> Application -> Domain
- Postgres/Redis/Kafka adapters -> Domain interfaces

The repository is still transitioning toward this shape. For changes in existing areas, prefer incremental refactors that preserve behavior instead of broad rewrites.

Target structure for new features:

```text
cmd/
	server/main.go

internal/
	domain/
		<aggregate>/
			entity.go
			repository.go
			errors.go

	application/
		<aggregate>/
			create_<aggregate>.go
			get_<aggregate>.go
			update_<aggregate>.go
			service.go

	adapters/
		http/
			router.go
			middleware/
			v1/
				<aggregate>_handler.go
		postgres/
			<aggregate>_repository.go
		redis/
			<feature>_repository.go

	platform/
		config/
		logger/
		clock/

	testsupport/

api/
	Makefile
	openapi.gen.go
	openapi.yaml
```

Preferred HTTP adapter shape for new or migrated v1 work:

```text
internal/
	adapters/
		http/
			v1/
				actor/
					handler.go
					payload.go
					handler_test.go
				auth/
					handler.go
					helpers.go
					handler_test.go
				cart/
					read/
						handler.go
						handler_test.go
					write/
						handler.go
						handler_test.go
					payload_flow_test.go
				catalog/
					read/
						handler.go
						handler_test.go
					write/
						handler.go
						handler_test.go
				coverage/
					read/
						handler.go
						handler_test.go
					write/
						handler.go
						handler_test.go
				merchant/
					read/
						handler.go
						handler_test.go
					write/
						handler.go
						handler_test.go
					payload_flow_test.go
				order/
					handler.go
					write_handler.go
					handler_test.go
				report/
					handler.go
					handler_test.go
				shared/
					response_helpers.go
					write_response_helpers.go
				http.go
				http_e2e_flow_test.go
				main_test.go
				read_smoke_test.go
```

This is the preferred target shape for HTTP adapter code. The repository is still migrating, so apply it incrementally instead of moving everything at once.

## Codebase rules

### 1) Architecture boundaries

- Keep business logic in `internal/domain`.
- Put use cases in `internal/application`.
- Keep transport concerns (HTTP, decode/encode, middleware) in `internal/adapters/http`.
- Keep persistence concerns in adapter packages such as `internal/adapters/postgres` or repository implementations that satisfy domain interfaces.
- Domain packages must not import adapter packages, frameworks, or transport/database concerns.
- Keep domains isolated: merchant logic must live in merchant-specific files/services, and actor logic must live in actor-specific files/services.

### 2) API contract and handlers

- `api/openapi.yaml` is the editable API source of truth.
- Workflow for API changes:
	1. Change `api/openapi.yaml`.
	2. Run `make generate -C api`.
	3. Implement the endpoint behavior against the generated types and server surface in [api/openapi.gen.go](d:/projects/go-delivery-app/api/openapi.gen.go).
- Do not manually edit generated files (`api/openapi.gen.go`).
- Keep endpoint behavior consistent with OpenAPI status codes and models.
- HTTP handlers should only parse requests, call application use cases, and return responses.
- Do not put business logic in handlers.

### 3) Validation and error handling

- Validate request payloads before calling services.
- Return domain errors through centralized HTTP error encoding.
- Prefer typed/structured error responses over ad-hoc strings.
- Avoid leaking internal errors/messages to API clients.

### 4) Database and migrations

- Schema changes must be done via SQL migrations in `db/postgres/migration`.
- The repo currently keeps a single baseline migration pair there; update that baseline when resetting schema history.
- Keep `.up.sql` and `.down.sql` pairs consistent and reversible.
- After query changes, regenerate sqlc code (`sqlc generate` or `make sqlc`).

### 5) Testing requirements

- Follow TDD by default: write/adjust a failing test first (`red`), implement the minimal code to pass (`green`), then refactor while keeping tests green (`refactor`).
- For every behavior change, the PR must include at least one test that would fail without the code change.
- Add/update tests for behavior changes.
- Keep tests focused on API behavior and domain logic, not implementation details.
- Domain tests should cover business rules.
- Application tests should mock repositories.
- Integration tests should live in adapters and use Testcontainers when DB-backed behavior is required.
- Ensure affected tests pass before merging.

HTTP adapter testing strategy:

- Unit tests: test each handler independently with mocked dependencies using `httptest` and `testify`; focus on input -> output mapping, payload validation, and HTTP error handling.
- Flow tests: use `*_flow_test.go` for domain-level HTTP flows that exercise routing, payload conversion, and handler/domain integration together.
- E2E and smoke tests: keep cross-domain critical path checks at the `v1/` level in files such as `http_e2e_flow_test.go` and `read_smoke_test.go`.
- Shared helper tests: add focused tests for serialization, response helpers, and edge cases.
- Coverage priorities: handlers and domain logic first; do not chase coverage in generated code or external libraries.
- CI cadence: unit tests on every commit, integration tests on the main branch, E2E tests in nightly or staging pipelines.

### 6) Change discipline

- Prefer small, focused PRs.
- Do not mix refactors with feature/fix changes unless necessary.
- Preserve existing naming/style unless there is a clear reason to change it.
- Update README/OpenAPI/docs when behavior or workflow changes.

### 6.1) Documentation consistency

- Treat `api/openapi.yaml` as the source of truth, and keep generated code in sync with it.
- Keep `docs/*.md` aligned with implemented behavior.
- If a doc describes proposed behavior that is not implemented yet, mark it clearly as `draft` or `planned`.
- Do not leave stale examples in `docs/HTTP_EXAMPLES.md` after endpoint or payload changes.
- Design docs should explicitly separate current behavior from future design.
- When changing API behavior, update these in the same change when relevant:
	1. `api/openapi.yaml`
	2. generated code
	3. implementation and tests
	4. `README.md`
	5. affected files under `docs/`

### 7) File structure conventions

- Follow the target structure documented above for new code.
- Every new domain must be added in separate files.
- Keep domain-specific code grouped under `internal/` by domain. The v1 HTTP runtime lives under `internal/http/v1`.
- Prefer vertical slices for new features so a feature is represented consistently across domain, application, and adapters.
- Keep one domain per handler file or handler package.
- For HTTP adapters, prefer self-contained domain folders with handlers, payloads/DTOs, and tests together.
- When a domain has distinct read and write concerns, prefer `read/` and `write/` subfolders under that domain.
- Put cross-domain HTTP helper code under `internal/core`.
- Keep root-level `v1/` tests for cross-domain integration, smoke, and E2E coverage.

### List of bugs (fixed)

- [x] Product creation: added `image_url` and `track_inventory` (default false) to CreateProductRequest
- [x] Cart response: removed merchant_id, updated_at, applied_discount_*, discount; addons nested in product; only created_at
 

### List of features (to be added)

- [ ] Discount -> with code name to add. Validate from date and to date. 
- [ ] Discount -> Discount on products, order, and by cash/card based on entity type rules.
- [ ] Analytics -> Report on orders, products, and merchants. For example, top-selling products, total sales per merchant, etc.
- [ ] Socket -> for real time order tracking on change status, emit events to subscribed clients. For example, when an order status changes from "preparing" to "out for delivery", emit an event that the frontend can listen to and update the UI accordingly. But it should be consistent, no duplication


### HTTP examples and hostname header

- [docs/HTTP_EXAMPLES.md](docs/HTTP_EXAMPLES.md) – Request/response examples for catalog, cart, checkout (no need to run the flow)
- [docs/HOSTNAME_MERCHANT_HEADER.md](docs/HOSTNAME_MERCHANT_HEADER.md) – Hostname-to-`X-Merchant-Id` design for multi-tenant frontends (e.g. sujing.com.pk)