-- name: UpsertVatRule :one
INSERT INTO vat_rules (
    merchant_id,
    payment_type,
    rate
)
VALUES (
    $1,
    $2,
    $3
)
ON CONFLICT (merchant_id, payment_type)
DO UPDATE
SET rate = EXCLUDED.rate
RETURNING id, merchant_id, payment_type, rate, created_at;

-- name: GetVatRule :one
SELECT
    id,
    merchant_id,
    payment_type,
    rate,
    created_at
FROM vat_rules
WHERE merchant_id = $1 AND payment_type = $2
LIMIT 1;

-- name: UpdateVatRuleByID :one
UPDATE vat_rules
SET
    payment_type = $3,
    rate = $4
WHERE merchant_id = $1 AND id = $2
RETURNING id, merchant_id, payment_type, rate, created_at;
