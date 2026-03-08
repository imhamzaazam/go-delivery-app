# HTTP Request/Response Examples

Base URL: `https://api.example.com` (replace with your API host)  
Suijing hostname: `https://sujing.com.pk` ([Chinese Cuisine in Karachi](https://sujing.com.pk/))

---

## 1. Hostname-to-Merchant Header (API Gateway / Backend)

For multi-tenant frontends like **sujing.com.pk**, the frontend does **not** send `merchant_id`. The API Gateway or backend derives it from the request hostname and injects a trusted header.

### Flow

```
Browser (sujing.com.pk) → API Gateway → Backend
                              │
                              ├─ Host: sujing.com.pk
                              ├─ Lookup: sujing.com.pk → merchant_id = 0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11
                              └─ Inject: X-Merchant-Id: 0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11
```

### Header

| Header        | Source        | Purpose                                      |
|---------------|---------------|----------------------------------------------|
| `X-Merchant-Id` | API Gateway   | Injected from hostname; backend validates it |

### API Gateway config (example)

```yaml
# Nginx / Kong / AWS API GW / etc.
hostname_to_merchant:
  sujing.com.pk: "0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11"
  # other-merchant.com: "<merchant-uuid>"
```

### Backend validation

- Reject requests with `X-Merchant-Id` from untrusted sources (only accept when from API GW).
- Validate `X-Merchant-Id` is a valid UUID and exists in `merchants`.
- Use it for public catalog endpoints (products, categories, addons) instead of JWT `merchant_id`.

---

## 2. Login (get access token)

**Request**

```http
POST /api/v1/login HTTP/1.1
Host: api.example.com
Content-Type: application/json

{
  "merchant_id": "0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11",
  "email": "owner@suijing.com.pk",
  "password": "Password#123"
}
```

**Response (200 OK)**

```json
{
  "merchant_id": "0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11",
  "uid": "52ea5f5f-d993-4052-8308-0c5bc6f27801",
  "email": "owner@suijing.com.pk",
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "access_token_expires_at": "2026-03-08T00:30:00Z",
  "refresh_token_expires_at": "2026-03-14T23:30:00Z"
}
```

---

## 3. List categories (authenticated)

**Request**

```http
GET /api/v1/merchant/categories HTTP/1.1
Host: api.example.com
Authorization: Bearer <access_token>
```

**Response (200 OK)**

```json
[
  {
    "id": "84d25368-0c95-4e77-98d6-44e4970c9301",
    "merchant_id": "0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11",
    "name": "Sushi Rolls",
    "description": "Signature sushi roll selection",
    "created_at": "2026-03-01T10:00:00Z"
  },
  {
    "id": "84d25368-0c95-4e77-98d6-44e4970c9302",
    "merchant_id": "0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11",
    "name": "Ramen Bowls",
    "description": "Japanese ramen variants",
    "created_at": "2026-03-01T10:00:00Z"
  },
  {
    "id": "84d25368-0c95-4e77-98d6-44e4970c9303",
    "merchant_id": "0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11",
    "name": "Beverages",
    "description": "Drinks and tea",
    "created_at": "2026-03-01T10:00:00Z"
  }
]
```

---

## 4. List products (authenticated)

**Request**

```http
GET /api/v1/merchant/products HTTP/1.1
Host: api.example.com
Authorization: Bearer <access_token>
```

**Response (200 OK)**

```json
[
  {
    "id": "f5adf2e2-8b67-4f07-a365-9fefac3f0b01",
    "merchant_id": "0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11",
    "category_id": "84d25368-0c95-4e77-98d6-44e4970c9301",
    "name": "Dragon Roll",
    "description": "Prawn tempura roll with avocado and eel sauce",
    "base_price": 1690,
    "image_url": "https://suijing.com.pk/menu/dragon-roll.jpg",
    "track_inventory": true,
    "is_active": true,
    "created_at": "2026-03-01T10:00:00Z",
    "updated_at": "2026-03-01T10:00:00Z"
  },
  {
    "id": "f5adf2e2-8b67-4f07-a365-9fefac3f0b02",
    "merchant_id": "0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11",
    "category_id": "84d25368-0c95-4e77-98d6-44e4970c9302",
    "name": "Chicken Miso Ramen",
    "description": "Miso broth with grilled chicken and soft egg",
    "base_price": 1450,
    "image_url": "https://suijing.com.pk/menu/chicken-miso-ramen.jpg",
    "track_inventory": true,
    "is_active": true,
    "created_at": "2026-03-01T10:00:00Z",
    "updated_at": "2026-03-01T10:00:00Z"
  }
]
```

---

## 5. Get product detail with addons (authenticated)

**Request**

```http
GET /api/v1/products/f5adf2e2-8b67-4f07-a365-9fefac3f0b01 HTTP/1.1
Host: api.example.com
Authorization: Bearer <access_token>
```

**Response (200 OK)**

```json
{
  "id": "f5adf2e2-8b67-4f07-a365-9fefac3f0b01",
  "merchant_id": "0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11",
  "category_id": "84d25368-0c95-4e77-98d6-44e4970c9301",
  "category_name": "Sushi Rolls",
  "name": "Dragon Roll",
  "description": "Prawn tempura roll with avocado and eel sauce",
  "base_price": 1690,
  "image_url": "https://suijing.com.pk/menu/dragon-roll.jpg",
  "track_inventory": true,
  "is_active": true,
  "created_at": "2026-03-01T10:00:00Z",
  "updated_at": "2026-03-01T10:00:00Z"
}
```

**Request (product addons)**

```http
GET /api/v1/products/f5adf2e2-8b67-4f07-a365-9fefac3f0b01/addons HTTP/1.1
Host: api.example.com
Authorization: Bearer <access_token>
```

**Response (200 OK)**

```json
[
  {
    "id": "e7f6f82f-8dc2-40f6-9c9f-80abf5eec001",
    "product_id": "f5adf2e2-8b67-4f07-a365-9fefac3f0b01",
    "name": "Extra Wasabi",
    "price": 120,
    "created_at": "2026-03-01T10:00:00Z"
  }
]
```

---

## 6. Create cart (guest, no auth)

**Request**

```http
POST /api/v1/carts HTTP/1.1
Host: api.example.com
Content-Type: application/json

{
  "merchant_id": "0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11",
  "branch_id": "3a957f6d-6ee1-47fd-95e7-f96bc66c0c01",
  "cart_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
}
```

Optional `actor_id` for logged-in users; omit for guest checkout.

**Response (201 Created)**

```json
{
  "id": "8df06d24-15ee-4079-8e3b-cd16579d9d79",
  "branch_id": "3a957f6d-6ee1-47fd-95e7-f96bc66c0c01",
  "created_at": "2026-03-07T22:58:38.664473+03:00"
}
```

---

## 7. Add item to cart (no auth)

**Request**

```http
POST /api/v1/carts/8df06d24-15ee-4079-8e3b-cd16579d9d79/items HTTP/1.1
Host: api.example.com
Content-Type: application/json

{
  "product_id": "f5adf2e2-8b67-4f07-a365-9fefac3f0b01",
  "quantity": 2,
  "addon_ids": ["e7f6f82f-8dc2-40f6-9c9f-80abf5eec001"]
}
```

Optional `discount_id` for item-level discount.

**Response (201 Created)**

```json
{
  "product_id": "f5adf2e2-8b67-4f07-a365-9fefac3f0b01",
  "quantity": 2
}
```

---

## 8. Get cart detail (authenticated)

**Request**

```http
GET /api/v1/carts/8df06d24-15ee-4079-8e3b-cd16579d9d79 HTTP/1.1
Host: api.example.com
Authorization: Bearer <access_token>
```

**Response (200 OK)**

```json
{
  "id": "8df06d24-15ee-4079-8e3b-cd16579d9d79",
  "branch_id": "3a957f6d-6ee1-47fd-95e7-f96bc66c0c01",
  "created_at": "2026-03-07T22:58:38.664473+03:00",
  "items": [
    {
      "product_id": "f5adf2e2-8b67-4f07-a365-9fefac3f0b01",
      "quantity": 2,
      "product": {
        "id": "f5adf2e2-8b67-4f07-a365-9fefac3f0b01",
        "merchant_id": "0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11",
        "category_id": "84d25368-0c95-4e77-98d6-44e4970c9301",
        "name": "Dragon Roll",
        "description": "Prawn tempura roll with avocado and eel sauce",
        "base_price": 1690,
        "image_url": "https://suijing.com.pk/menu/dragon-roll.jpg",
        "track_inventory": true,
        "is_active": true,
        "created_at": "2026-03-01T10:00:00Z",
        "updated_at": "2026-03-01T10:00:00Z",
        "addons": [
          {
            "id": "e7f6f82f-8dc2-40f6-9c9f-80abf5eec001",
            "product_id": "f5adf2e2-8b67-4f07-a365-9fefac3f0b01",
            "name": "Extra Wasabi",
            "price": 120,
            "created_at": "2026-03-01T10:00:00Z"
          }
        ]
      }
    }
  ]
}
```

---

## 9. Place order (no auth)

**Request**

```http
POST /api/v1/orders HTTP/1.1
Host: api.example.com
Content-Type: application/json

{
  "cart_id": "8df06d24-15ee-4079-8e3b-cd16579d9d79",
  "payment_type": "card",
  "delivery_address": "DHA Phase 6, Karachi",
  "customer_name": "Sara Ahmed",
  "customer_phone": "03001234567"
}
```

**Response (201 Created)**

```json
{
  "order_id": "b5cd7f84-183a-4284-8443-2984b8bf5ccb",
  "vat_rate": 10,
  "subtotal": 3620,
  "total_tax": 362,
  "total": 3982,
  "line_items": [
    {
      "product_id": "f5adf2e2-8b67-4f07-a365-9fefac3f0b01",
      "quantity": 2,
      "base_amount": 3380,
      "addon_amount": 240,
      "discount_amount": 0,
      "tax_amount": 362,
      "line_total": 3982
    }
  ]
}
```

---

## 10. Create product (with image_url and track_inventory)

**Request**

```http
POST /api/v1/merchant/products HTTP/1.1
Host: api.example.com
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "category_id": "84d25368-0c95-4e77-98d6-44e4970c9301",
  "name": "Chinese Rice",
  "description": "Chinese rice",
  "base_price": 10,
  "image_url": "https://suijing.com.pk/menu/chinese-rice.jpg",
  "track_inventory": false
}
```

**Response (201 Created)**

```json
{
  "id": "726ce295-1ae6-4abc-a400-270475c27cb5",
  "merchant_id": "0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11",
  "category_id": "84d25368-0c95-4e77-98d6-44e4970c9301",
  "name": "Chinese Rice",
  "description": "Chinese rice",
  "base_price": 10,
  "image_url": "https://suijing.com.pk/menu/chinese-rice.jpg",
  "track_inventory": false,
  "is_active": true,
  "created_at": "2026-03-07T22:53:30.870334+03:00",
  "updated_at": "2026-03-07T22:53:30.870334+03:00"
}
```

---

## 11. Check coverage (delivery zone)

**Request**

```http
POST /api/v1/merchant/service-zones/check HTTP/1.1
Host: api.example.com
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "latitude": 24.912500,
  "longitude": 67.128500
}
```

**Response (200 OK)**

```json
{
  "covered": true,
  "merchant_id": "0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11",
  "zone_id": "uuid-of-zone",
  "zone_name": "Gulistan-e-Jauhar Block 1",
  "branch_id": "3a957f6d-6ee1-47fd-95e7-f96bc66c0c01",
  "branch_name": "Suijing DHA Branch",
  "area_id": "uuid-of-area",
  "area_name": "Gulistan-e-Jauhar",
  "area_city": "Karachi"
}
```

---

## Suijing seed IDs (reference)

| Entity   | ID                                   |
|----------|--------------------------------------|
| Merchant | `0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11` |
| Branch DHA | `3a957f6d-6ee1-47fd-95e7-f96bc66c0c01` |
| Branch Clifton | `3a957f6d-6ee1-47fd-95e7-f96bc66c0c02` |
| Category Sushi | `84d25368-0c95-4e77-98d6-44e4970c9301` |
| Category Ramen | `84d25368-0c95-4e77-98d6-44e4970c9302` |
| Category Beverages | `84d25368-0c95-4e77-98d6-44e4970c9303` |
| Product Dragon Roll | `f5adf2e2-8b67-4f07-a365-9fefac3f0b01` |
| Product Chicken Miso Ramen | `f5adf2e2-8b67-4f07-a365-9fefac3f0b02` |
| Product Iced Matcha | `f5adf2e2-8b67-4f07-a365-9fefac3f0b03` |
| Addon Extra Wasabi | `e7f6f82f-8dc2-40f6-9c9f-80abf5eec001` |
| Addon Soft Egg | `e7f6f82f-8dc2-40f6-9c9f-80abf5eec002` |
