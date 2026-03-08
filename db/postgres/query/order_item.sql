-- name: UpsertOrderItem :one
INSERT INTO order_items (
    order_id,
    product_id,
    quantity,
    price,
    base_amount,
    addon_amount,
    discount_amount,
    tax_amount,
    line_total
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9
)
ON CONFLICT (order_id, product_id)
DO UPDATE
SET
    quantity = EXCLUDED.quantity,
    price = EXCLUDED.price,
    base_amount = EXCLUDED.base_amount,
    addon_amount = EXCLUDED.addon_amount,
    discount_amount = EXCLUDED.discount_amount,
    tax_amount = EXCLUDED.tax_amount,
    line_total = EXCLUDED.line_total
RETURNING order_id, product_id, quantity, price, base_amount, addon_amount, discount_amount, tax_amount, line_total;

-- name: GetOrderItem :one
SELECT
    order_id,
    product_id,
    quantity,
    price,
    base_amount,
    addon_amount,
    discount_amount,
    tax_amount,
    line_total
FROM order_items
WHERE order_id = $1 AND product_id = $2
LIMIT 1;

-- name: UpdateOrderItem :one
UPDATE order_items
SET
    quantity = $3,
    price = $4,
    base_amount = $5,
    addon_amount = $6,
    discount_amount = $7,
    tax_amount = $8,
    line_total = $9
WHERE order_id = $1 AND product_id = $2
RETURNING order_id, product_id, quantity, price, base_amount, addon_amount, discount_amount, tax_amount, line_total;
