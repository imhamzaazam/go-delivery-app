# Requirements

## 1) Goal
Provide a tenant-aware authentication and identity API for merchants and actors.
Actors are created within a merchant, can authenticate with JWT access/refresh tokens, and can fetch actor profile by UID.
Merchants can be created via dedicated endpoint and are used as the tenant boundary for actors.

## 2) Scope
### In Scope
- Merchant creation (`POST /api/v1/merchants`)
- Actor creation under merchant (`POST /api/v1/actors`)
- Actor login (`POST /api/v1/login`)
- Access token renewal (`POST /api/v1/renew-token`)
- Authenticated actor fetch by UID (`GET /api/v1/actor/{uid}`)
- Structured validation + domain error mapping
- Testcontainers-based integration tests for v1 HTTP layer
- RBAC/role assignment flows
- Merchant management updates/deletes
- Password reset/account recovery
- Pagination/list endpoints

### Out of Scope


## 3) Functional Requirements
List each requirement with an ID.

- FR-001: System MUST create merchants with required business fields (name, ntn, address, category, contact_number).
- FR-002: System MUST create actors only when `merchant_id` is provided by frontend and valid UUID.
- FR-003: System MUST hash actor password before persistence and split `full_name` into first/last name.
- FR-004: System MUST authenticate actor by email/password and issue access + refresh tokens.
- FR-005: System MUST persist refresh-token sessions and allow renewing access token.
- FR-006: System MUST return actor profile by UID including `merchant_id`.
- FR-007: System MUST create actor and assign roles to it ('admin', 'merchant', 'employee', 'customer') which can only be done my admin role.
- FR-008: System MUST only seed values for admin role, its not allowed to create through api
- FR-009: System MUST create 'merchant', 'employee', 'customer' these roles through API
- FR-010: System MUST allow updating and deleting merchants through API
- FR-011: System MUST allow password reset and account recovery flows for actors
- FR-012: System MUST have different apis for admin, merchant, employees, customer
- FR-013: A merchant can create product with images, category, price, description, stock quantity, and can update or delete products. Employees can only update products but not create or delete. Customers can only view products.

## 4) API Contract Changes
For each endpoint, define request/response fields, status codes, and auth requirements.

### Endpoint: `POST /api/v1/merchants`
- Purpose: Create merchant tenant.
- Auth: `none`
- Request body:
  - `name` (string, required)
  - `ntn` (string, required)
  - `address` (string, required)
  - `category` (string enum: `restaurant|pharma|bakery`, required)
  - `contact_number` (string, required)
- Success response:
  - Status: `201`
  - Body fields: `id`, `name`, `ntn`, `address`, `category`, `contact_number`
- Error responses:
  - `400`: malformed JSON
  - `409`: duplicate constraint conflict
  - `422`: validation error

### Endpoint: `POST /api/v1/actors`
- Purpose: Create actor under merchant.
- Auth: `none`
- Request body:
  - `merchant_id` (uuid, required)
  - `full_name` (string, required)
  - `email` (email, required)
  - `password` (string, required)
- Success response:
  - Status: `201`
  - Body fields: `merchant_id`, `uid`, `full_name`, `email`
- Error responses:
  - `400`: malformed JSON
  - `409`: duplicate email for merchant
  - `422`: validation error

### Endpoint: `POST /api/v1/login`
- Purpose: Actor login.
- Auth: `none`
- Request body:
  - `email` (email, required)
  - `password` (string, required)
- Success response:
  - Status: `200`
  - Body fields: `email`, `access_token`, `refresh_token`, `access_token_expires_at`, `refresh_token_expires_at`
- Error responses:
  - `400`: malformed JSON
  - `404`: actor not found
  - `422`: validation error

### Endpoint: `POST /api/v1/renew-token`
- Purpose: Refresh access token from refresh token.
- Auth: `none`
- Request body:
  - `refresh_token` (string, required)
- Success response:
  - Status: `200`
  - Body fields: `access_token`, `access_token_expires_at`
- Error responses:
  - `400`: malformed JSON
  - `401`: invalid/expired/blocked session or token mismatch
  - `422`: validation error

### Endpoint: `GET /api/v1/actor/{uid}`
- Purpose: Fetch actor profile.
- Auth: `BearerAuth`
- Path params:
  - `uid` (uuid string)
- Success response:
  - Status: `200`
  - Body fields: `merchant_id`, `uid`, `full_name`, `email`
- Error responses:
  - `401`: unauthorized
  - `404`: actor not found

## 5) Data / Database Requirements
Describe schema or query-level behavior needed.

- Tables impacted: `merchants`, `actors`, `sessions`
- New columns: none in current cycle (existing multitenant schema used)
- Query behavior:
  - `CreateActor` inserts explicit `merchant_id` from frontend request.
  - `GetActorByUID` returns `merchant_id` + actor fields.
  - `CreateSession` resolves actor by email and persists session with merchant/actor linkage.
- Migration needed: `no` (for latest `merchant_id` response addition)

## 6) Validation Rules
Explicit input validation rules.

- VR-001: `merchant_id` on create actor is `required,uuid4`.
- VR-002: Actor `email` is `required,email`.
- VR-003: Actor `full_name` and `password` are required.
- VR-004: Merchant `category` is `required,oneof=restaurant pharma bakery`.
- VR-005: Merchant create fields (`name`, `ntn`, `address`, `contact_number`) are required.
- VR-006: Login requires email + password.
- VR-007: Renew token requires `refresh_token`.

## 7) Business Rules
Domain rules and constraints.

- BR-001: Actor identity is tenant-bound by `merchant_id`.
- BR-002: Duplicate actor email is constrained per merchant.
- BR-003: Actor password is stored as bcrypt hash (never plaintext).
- BR-004: Refresh-token session must be unblocked, unexpired, and token-consistent to renew access token.
- BR-005: Merchant and actor code remain isolated by domain-specific files/services.

## 8) Non-Functional Requirements
- Performance: SQLC generated queries, indexed UUID-based lookups, lean HTTP handlers.
- Security: JWT bearer auth for actor profile endpoint; password hashing; refresh-session checks.
- Observability/logging: Request ID + structured request/error logs via middleware.
- Backward compatibility: Current API is actor/merchant naming (legacy user endpoints removed).

## 9) Testing / Acceptance Criteria
Write clear testable outcomes.

- AC-001: Given valid merchant + actor payload, when calling `POST /api/v1/actors`, then response is `201` and includes `merchant_id`, `uid`, `full_name`, `email`.
- AC-002: Given invalid actor email, when creating actor or logging in, then response is `422`.
- AC-003: Given missing `merchant_id`, when creating actor, then response is `422`.
- AC-004: Given valid login, when calling `POST /api/v1/login`, then response is `200` with access/refresh tokens and expirations.
- AC-005: Given valid bearer token, when calling `GET /api/v1/actor/{uid}`, then response is `200` and includes `merchant_id`.
- AC-006: Given invalid merchant category, when calling `POST /api/v1/merchants`, then response is `422`.

## 10) Rollout Notes (Optional)
- Feature flag needed: `no`
- Migration order: existing migration baseline + regenerate (`make sqlc`, `make generate`) for query/spec changes.
- Manual verification steps:
  1. Start postgres via `make run-services` (or run tests with Testcontainers).
  2. Create merchant via `POST /api/v1/merchants`.
  3. Create actor with returned `merchant_id`.
  4. Login actor and get bearer token.
  5. Call `GET /api/v1/actor/{uid}` and verify `merchant_id` is returned.

## 11) Open Questions
- Q-001: Should actor login also return `merchant_id` for multi-tenant clients?
- Q-002: Should `contact_number` and `ntn` enforce stricter format validation?
- Q-003: Do we need a `GET /api/v1/merchants/{id}` endpoint next?

---

## Implementation Mode (for Copilot)
Choose one:
- `MVP` (minimal implementation)
- `Full` (complete implementation)

Preferred mode: `MVP`
