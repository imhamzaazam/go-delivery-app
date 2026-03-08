-- name: UpsertCartItem :one
INSERT INTO cart_items (
    cart_id,
    product_id,
    quantity,
    addon_ids,
    applied_discount_id,
    applied_discount_amount
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
ON CONFLICT (cart_id, product_id)
DO UPDATE
SET
    quantity = EXCLUDED.quantity,
    addon_ids = EXCLUDED.addon_ids,
    applied_discount_id = EXCLUDED.applied_discount_id,
    applied_discount_amount = EXCLUDED.applied_discount_amount
RETURNING cart_id, product_id, quantity, addon_ids, applied_discount_id, applied_discount_amount;

-- name: GetCartItem :one
SELECT
    cart_id,
    product_id,
    quantity,
    addon_ids,
    applied_discount_id,
    applied_discount_amount
FROM cart_items
WHERE cart_id = $1 AND product_id = $2
LIMIT 1;

-- name: UpdateCartItem :one
UPDATE cart_items
SET
    quantity = $3,
    addon_ids = $4,
    applied_discount_id = $5,
    applied_discount_amount = $6
WHERE cart_id = $1 AND product_id = $2
RETURNING cart_id, product_id, quantity, addon_ids, applied_discount_id, applied_discount_amount;
