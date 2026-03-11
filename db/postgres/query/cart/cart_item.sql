-- name: CreateCartItem :one
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
RETURNING id, cart_id, product_id, quantity, addon_ids, applied_discount_id, applied_discount_amount;

-- name: GetCartItemBySignature :one
SELECT
    id,
    cart_id,
    product_id,
    quantity,
    addon_ids,
    applied_discount_id,
    applied_discount_amount
FROM cart_items
WHERE cart_id = $1 AND product_id = $2 AND addon_ids = $3
LIMIT 1;

-- name: GetCartItemByID :one
SELECT
    id,
    cart_id,
    product_id,
    quantity,
    addon_ids,
    applied_discount_id,
    applied_discount_amount
FROM cart_items
WHERE cart_id = $1 AND id = $2
LIMIT 1;

-- name: UpdateCartItemByID :one
UPDATE cart_items
SET
    quantity = $3,
    addon_ids = $4,
    applied_discount_id = $5,
    applied_discount_amount = $6
WHERE cart_id = $1 AND id = $2
RETURNING id, cart_id, product_id, quantity, addon_ids, applied_discount_id, applied_discount_amount;

-- name: DeleteCartItem :execrows
DELETE FROM cart_items
WHERE cart_id = $1 AND id = $2;
