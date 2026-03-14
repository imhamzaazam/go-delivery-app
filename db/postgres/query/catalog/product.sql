-- name: CreateProduct :one
INSERT INTO products (
    merchant_id,
    category_id,
    name,
    description,
    base_price,
    image_url,
    track_inventory,
    is_active
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
)
RETURNING id, merchant_id, category_id, name, description, base_price, image_url, track_inventory, is_active, created_at, updated_at;

-- name: GetProduct :one
SELECT
    id,
    merchant_id,
    category_id,
    name,
    description,
    base_price,
    image_url,
    track_inventory,
    is_active,
    created_at,
    updated_at
FROM products
WHERE merchant_id = $1 AND id = $2
LIMIT 1;

-- name: UpdateProduct :one
UPDATE products
SET
    category_id = $3,
    name = $4,
    description = $5,
    base_price = $6,
    image_url = $7,
    track_inventory = $8,
    is_active = $9,
    updated_at = NOW()
WHERE merchant_id = $1 AND id = $2
RETURNING id, merchant_id, category_id, name, description, base_price, image_url, track_inventory, is_active, created_at, updated_at;
