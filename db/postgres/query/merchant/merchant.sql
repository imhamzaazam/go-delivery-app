-- name: CreateMerchant :one
INSERT INTO merchants (
    name,
    ntn,
    address,
    category,
    contact_number
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING
    id,
    name,
    ntn,
    address,
    logo,
    category,
    contact_number,
    created_at,
    updated_at;

-- name: GetMerchant :one
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
WHERE id = $1
LIMIT 1;

-- name: UpdateMerchant :one
UPDATE merchants
SET
    name = $2,
    ntn = $3,
    address = $4,
    category = $5,
    contact_number = $6,
    updated_at = NOW()
WHERE id = $1
RETURNING id, name, ntn, address, logo, category, contact_number, created_at, updated_at;