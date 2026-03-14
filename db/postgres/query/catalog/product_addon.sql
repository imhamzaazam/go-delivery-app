-- name: CreateProductAddon :one
INSERT INTO product_addons (
    product_id,
    name,
    price
)
VALUES (
    $1,
    $2,
    $3
)
RETURNING id, product_id, name, price, created_at;

-- name: GetProductAddon :one
SELECT
    id,
    product_id,
    name,
    price,
    created_at
FROM product_addons
WHERE id = $1
LIMIT 1;

-- name: UpdateProductAddon :one
UPDATE product_addons
SET
    name = $2,
    price = $3
WHERE id = $1
RETURNING id, product_id, name, price, created_at;
