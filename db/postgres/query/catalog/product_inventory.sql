-- name: UpsertProductInventory :one
INSERT INTO product_inventory (
    product_id,
    branch_id,
    quantity
)
VALUES (
    $1,
    $2,
    $3
)
ON CONFLICT (product_id, branch_id)
DO UPDATE
SET
    quantity = EXCLUDED.quantity,
    updated_at = NOW()
RETURNING id, product_id, branch_id, quantity, created_at, updated_at;

-- name: GetProductInventory :one
SELECT
    id,
    product_id,
    branch_id,
    quantity,
    created_at,
    updated_at
FROM product_inventory
WHERE product_id = $1 AND branch_id = $2
LIMIT 1;

-- name: UpdateProductInventoryQuantity :one
UPDATE product_inventory
SET
    quantity = $3,
    updated_at = NOW()
WHERE product_id = $1 AND branch_id = $2
RETURNING id, product_id, branch_id, quantity, created_at, updated_at;
