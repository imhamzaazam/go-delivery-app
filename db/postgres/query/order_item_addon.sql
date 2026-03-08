-- name: UpsertOrderItemAddon :one
INSERT INTO order_item_addons (
    order_id,
    product_id,
    addon_id,
    addon_name,
    addon_price,
    quantity,
    line_addon_total
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7
)
ON CONFLICT (order_id, product_id, addon_id)
DO UPDATE
SET
    addon_name = EXCLUDED.addon_name,
    addon_price = EXCLUDED.addon_price,
    quantity = EXCLUDED.quantity,
    line_addon_total = EXCLUDED.line_addon_total
RETURNING order_id, product_id, addon_id, addon_name, addon_price, quantity, line_addon_total;

-- name: GetOrderItemAddon :one
SELECT
    order_id,
    product_id,
    addon_id,
    addon_name,
    addon_price,
    quantity,
    line_addon_total
FROM order_item_addons
WHERE order_id = $1 AND product_id = $2 AND addon_id = $3
LIMIT 1;

-- name: UpdateOrderItemAddon :one
UPDATE order_item_addons
SET
    addon_name = $4,
    addon_price = $5,
    quantity = $6,
    line_addon_total = $7
WHERE order_id = $1 AND product_id = $2 AND addon_id = $3
RETURNING order_id, product_id, addon_id, addon_name, addon_price, quantity, line_addon_total;