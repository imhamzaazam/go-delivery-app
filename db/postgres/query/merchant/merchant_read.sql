-- name: ListMerchants :many
SELECT
    id,
    name,
    ntn,
    address,
    logo,
    category,
    contact_number,
    created_at,
    updated_at
FROM merchants
ORDER BY created_at DESC, id;

-- name: ListBranchesByMerchant :many
SELECT
    id,
    merchant_id,
    name,
    address,
    contact_number,
    city,
    opening_time_minutes,
    closing_time_minutes,
    created_at,
    updated_at
FROM branches
WHERE merchant_id = $1
ORDER BY created_at DESC, id;

-- name: ListRolesByMerchant :many
SELECT
    id,
    merchant_id,
    role_type,
    description,
    created_at
FROM roles
WHERE merchant_id = $1
ORDER BY created_at DESC, id;

-- name: ListDiscountsByMerchant :many
SELECT
    id,
    merchant_id,
    product_id,
    category_id,
    type,
    value,
    description,
    valid_from,
    valid_to,
    created_at
FROM merchant_discounts
WHERE merchant_id = $1
ORDER BY created_at DESC, id;
