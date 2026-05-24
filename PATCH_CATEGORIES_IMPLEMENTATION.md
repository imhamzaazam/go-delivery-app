# PATCH Category Availability API Implementation

## Overview

Created a PATCH API endpoint for `/merchant/categories/{categoryID}` to allow merchants to update the availability status of product categories.

## Changes Made

### 1. Database Migration

**File:** `db/postgres/migration/000002_add_category_availability.up.sql`

- Added `is_available BOOLEAN NOT NULL DEFAULT true` column to `product_categories` table
- Down migration provided in `000002_add_category_availability.down.sql`

### 2. SQL Queries

**File:** `db/postgres/query/catalog/product_category.sql`

- Added new query: `UpdateProductCategoryAvailability`
  - Updates the `is_available` field for a specific category
  - Returns the updated category with all fields including `is_available`
- Updated `CreateProductCategory` and `GetProductCategory` to include `is_available` field
- Updated `ListProductCategoriesByMerchant` in `catalog_read.sql` to include `is_available`

### 3. Generated Database Code

**Files:** `internal/catalog/store/generated/`

- `product_category.sql.go`: Added `UpdateProductCategoryAvailability()` function
- `models.go`: Added `IsAvailable bool` field to `ProductCategory` struct
- `catalog_read.sql.go`: Updated scan operations to include `IsAvailable`
- `querier.go`: Added method signature for `UpdateProductCategoryAvailability`

### 4. Store Layer

**Files:** `internal/catalog/store/`

- `dto.go`: Added type alias for `UpdateProductCategoryAvailabilityParams`
- `postgres.go`: Added `UpdateProductCategoryAvailability()` wrapper method

### 5. Service Layer

**Files:**

- `internal/catalog/service.go`:
  - Added `UpdateProductCategoryAvailability` to Service interface
  - Added type alias for `UpdateProductCategoryAvailabilityParams`

- `internal/catalog/app/catalog_query.go`:
  - Implemented `UpdateProductCategoryAvailability()` service method
  - Handles UUID parsing and domain error mapping

### 6. Commerce Store

**File:** `internal/commerce/store/postgres.go`

- Added delegation method `UpdateProductCategoryAvailability()` to commerce store

**File:** `internal/commerce/types.go`

- Added type alias for `UpdateProductCategoryAvailabilityParams`

### 7. HTTP Handler

**File:** `internal/catalog/presentation/write_handler.go`

- Implemented `UpdateProductCategoryAvailability()` handler
- Accepts `categoryID` as UUID parameter
- Deserializes request body to `UpdateCategoryAvailabilityRequest`
- Authenticates user and updates category availability
- Returns updated category with HTTP 200 status

### 8. API Schema

**File:** `api/openapi.yaml`

- Added `UpdateCategoryAvailabilityRequest` schema with required `is_available: boolean` field
- Added `is_available` field to `ProductCategoryResponse` schema
- Added PATCH endpoint: `/api/v1/merchant/categories/{categoryID}`
  - Parameters: `categoryID` (path, UUID, required)
  - Request body: `UpdateCategoryAvailabilityRequest`
  - Security: BearerAuth token required
  - Response: `ProductCategoryResponse` (200 OK)

## API Usage

### Request

```bash
PATCH /api/v1/merchant/categories/{categoryID}
Authorization: Bearer <token>
Content-Type: application/json

{
  "is_available": false
}
```

### Response

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "merchant_id": "550e8400-e29b-41d4-a716-446655440001",
  "name": "Beverages",
  "description": "Cold drinks and beverages",
  "is_available": false,
  "created_at": "2024-01-15T10:30:00Z"
}
```

## Testing

The implementation:

- ✅ Compiles successfully
- ✅ Follows existing code patterns and conventions
- ✅ Includes proper authentication/authorization
- ✅ Includes input validation
- ✅ Supports proper error handling with domain errors
- ✅ Includes type-safe UUID handling via OpenAPI code generation

## Next Steps

1. Run database migrations: `make run` will automatically apply migrations
2. The API is ready to be tested with the running application
3. Category availability can now be toggled via the PATCH endpoint
