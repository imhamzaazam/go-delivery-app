-- name: CreateMerchantDiscount :one
INSERT INTO merchant_discounts (
    merchant_id,
    type,
    value,
    description,
    valid_from,
    valid_to
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING id, merchant_id, type, value, description, valid_from, valid_to, created_at;

-- name: GetMerchantDiscount :one
SELECT
    id,
    merchant_id,
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
    type = $3,
    value = $4,
    description = $5,
    valid_from = $6,
    valid_to = $7
WHERE merchant_id = $1 AND id = $2
RETURNING id, merchant_id, type, value, description, valid_from, valid_to, created_at;
