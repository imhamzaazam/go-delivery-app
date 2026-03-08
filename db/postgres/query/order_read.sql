-- name: ListOrdersByMerchant :many
SELECT
    id,
    cart_id,
    merchant_id,
    branch_id,
    actor_id,
    payment_type,
    vat_rate,
    total_amount,
    status,
    delivery_address,
    customer_name,
    customer_phone,
    created_at,
    updated_at
FROM orders
WHERE merchant_id = $1
ORDER BY created_at DESC, id;

-- name: ListOrderItemsByOrder :many
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
WHERE order_id = $1
ORDER BY product_id;

-- name: ListOrderItemAddonsByOrder :many
SELECT
    order_id,
    product_id,
    addon_id,
    addon_name,
    addon_price,
    quantity,
    line_addon_total
FROM order_item_addons
WHERE order_id = $1
ORDER BY product_id, addon_id;
