-- name: CreateProductCategory :one
INSERT INTO product_categories (
    merchant_id,
    name,
    description
)
VALUES (
    $1,
    $2,
    $3
)
RETURNING id, merchant_id, name, description, created_at;

-- name: GetProductCategory :one
SELECT
    id,
    merchant_id,
    name,
    description,
    created_at
FROM product_categories
WHERE merchant_id = $1 AND id = $2
LIMIT 1;

-- name: UpdateProductCategory :one
UPDATE product_categories
SET
    name = $3,
    description = $4
WHERE merchant_id = $1 AND id = $2
RETURNING id, merchant_id, name, description, created_at;
