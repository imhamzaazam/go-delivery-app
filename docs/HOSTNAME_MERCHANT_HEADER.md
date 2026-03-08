# Hostname-to-Merchant Header Design

For multi-tenant storefronts like [Sujing (sujing.com.pk)](https://sujing.com.pk/), the frontend should **not** send `merchant_id`. The backend derives it from the request hostname via a trusted header injected by the API Gateway.

## Problem

- Frontend at `https://sujing.com.pk` needs to show Suijing products and addons.
- Catalog endpoints (`/merchant/categories`, `/merchant/products`, `/products/{id}/addons`) currently require Bearer auth and use `merchant_id` from the JWT.
- Guest users browsing the menu have no JWT.
- We don’t want the frontend to hardcode or send `merchant_id` (security and flexibility).

## Solution: `X-Merchant-Id` header

### Flow

```
┌─────────────┐     ┌─────────────────┐     ┌─────────────┐
│   Browser   │     │   API Gateway   │     │   Backend   │
│ sujing.com  │────▶│ Host: sujing... │────▶│ Validate &  │
│             │     │ Inject header   │     │ Use header  │
└─────────────┘     └─────────────────┘     └─────────────┘
                           │
                           │ Host → merchant lookup
                           │ X-Merchant-Id: 0f9f93c2-...
                           ▼
```

### 1. API Gateway (or reverse proxy)

- Reads `Host` (e.g. `sujing.com.pk`).
- Maps hostname → `merchant_id` (config or DB).
- Adds `X-Merchant-Id: <uuid>` to the request.
- Optionally strips or overwrites `X-Merchant-Id` from the client so it cannot be spoofed.

**Example mapping:**

| Hostname       | Merchant ID                              |
|----------------|------------------------------------------|
| sujing.com.pk  | 0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11    |
| other-store.com| &lt;other-merchant-uuid&gt;               |

### 2. Backend validation

- Only trust `X-Merchant-Id` when the request comes from the API Gateway (e.g. internal network, specific proxy, or signed header).
- Validate that the UUID exists in `merchants`.
- Use it for public catalog endpoints instead of JWT `merchant_id`.

### 3. Public catalog endpoints (to implement)

New or extended endpoints that:

- Do **not** require Bearer auth.
- Require `X-Merchant-Id` (injected by API GW).
- Return products, categories, and addons for that merchant.

**Option A: New public routes**

```
GET /api/v1/catalog/categories     # X-Merchant-Id required
GET /api/v1/catalog/products       # X-Merchant-Id required
GET /api/v1/catalog/products/{id}  # X-Merchant-Id required
GET /api/v1/catalog/products/{id}/addons  # X-Merchant-Id required
```

**Option B: Extend existing routes**

- If `Authorization` is absent but `X-Merchant-Id` is present and valid, treat as public catalog request.
- If `Authorization` is present, use JWT `merchant_id` as today.

### 4. Security

- **Never** trust `X-Merchant-Id` from the public internet; only from the API Gateway.
- Use one of:
  - Internal network (backend not exposed publicly).
  - API Gateway in front that strips client `X-Merchant-Id` and injects its own.
  - Signed header or HMAC so backend can verify the header came from the gateway.

### 5. Frontend usage

Frontend at `sujing.com.pk`:

- Calls `https://api.example.com/api/v1/catalog/products` (or equivalent).
- Does **not** send `X-Merchant-Id` or `merchant_id`.
- API Gateway sees `Host: sujing.com.pk` (or Referer/Origin) and injects `X-Merchant-Id`.
- Backend returns Suijing products.

---

## Implementation checklist

- [ ] API Gateway: hostname → merchant_id mapping
- [ ] API Gateway: inject `X-Merchant-Id` for catalog requests
- [ ] Backend: middleware to read and validate `X-Merchant-Id`
- [ ] Backend: public catalog endpoints (or optional auth on existing ones)
- [ ] Backend: ensure `X-Merchant-Id` is only accepted from trusted gateway
