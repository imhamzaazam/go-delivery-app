-- name: ListCartItemsByCart :many
SELECT
    cart_id,
    product_id,
    quantity,
    addon_ids,
    applied_discount_id,
    applied_discount_amount
FROM cart_items
WHERE cart_id = $1
ORDER BY product_id;
