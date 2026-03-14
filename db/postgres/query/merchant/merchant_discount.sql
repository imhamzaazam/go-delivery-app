-- name: CreateMerchantDiscount :one
INSERT INTO merchant_discounts (
    merchant_id,
    product_id,
    category_id,
    type,
    value,
    description,
    valid_from,
    valid_to
)
VALUES (
    sqlc.arg(merchant_id),
    NULLIF(sqlc.arg(product_id)::uuid, '00000000-0000-0000-0000-000000000000'::uuid),
    NULLIF(sqlc.arg(category_id)::uuid, '00000000-0000-0000-0000-000000000000'::uuid),
    sqlc.arg(type),
    sqlc.arg(value),
    sqlc.arg(description),
    sqlc.arg(valid_from),
    sqlc.arg(valid_to)
)
RETURNING id, merchant_id, product_id, category_id, type, value, description, valid_from, valid_to, created_at;

-- name: GetMerchantDiscount :one
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
WHERE merchant_id = $1 AND id = $2
LIMIT 1;

-- name: UpdateMerchantDiscount :one
UPDATE merchant_discounts
SET
    product_id = NULLIF(sqlc.arg(product_id)::uuid, '00000000-0000-0000-0000-000000000000'::uuid),
    category_id = NULLIF(sqlc.arg(category_id)::uuid, '00000000-0000-0000-0000-000000000000'::uuid),
    type = sqlc.arg(type),
    value = sqlc.arg(value),
    description = sqlc.arg(description),
    valid_from = sqlc.arg(valid_from),
    valid_to = sqlc.arg(valid_to)
WHERE merchant_id = sqlc.arg(merchant_id) AND id = sqlc.arg(id)
RETURNING id, merchant_id, product_id, category_id, type, value, description, valid_from, valid_to, created_at;
