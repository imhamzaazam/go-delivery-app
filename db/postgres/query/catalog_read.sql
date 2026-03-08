-- name: ListProductCategoriesByMerchant :many
SELECT
    id,
    merchant_id,
    name,
    description,
    created_at
FROM product_categories
WHERE merchant_id = $1
ORDER BY created_at DESC, id;

-- name: ListProductsByMerchant :many
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
WHERE merchant_id = $1
ORDER BY created_at DESC, id;

-- name: GetProductDetail :one
SELECT
    p.id,
    p.merchant_id,
    p.category_id,
    p.name,
    p.description,
    p.base_price,
    p.image_url,
    p.track_inventory,
    p.is_active,
    p.created_at,
    p.updated_at,
    COALESCE(pc.name, '') AS category_name
FROM products p
LEFT JOIN product_categories pc
  ON pc.merchant_id = p.merchant_id
 AND pc.id = p.category_id
WHERE p.merchant_id = $1
  AND p.id = $2
LIMIT 1;

-- name: ListProductAddonsByProduct :many
SELECT
    pa.id,
    pa.product_id,
    pa.name,
    pa.price,
    pa.created_at
FROM product_addons pa
WHERE pa.product_id = $1
ORDER BY pa.created_at DESC, pa.id;

-- name: ListInventoryByMerchant :many
SELECT
    p.id AS product_id,
    p.name AS product_name,
    COALESCE(SUM(pi.quantity), 0)::int AS quantity
FROM products p
LEFT JOIN product_inventory pi
  ON pi.product_id = p.id
WHERE p.merchant_id = $1
GROUP BY p.id, p.name
ORDER BY p.name;
