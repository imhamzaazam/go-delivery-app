-- name: CreateBranch :one
INSERT INTO branches (
    merchant_id,
    name,
    address,
    contact_number,
    city,
    opening_time_minutes,
    closing_time_minutes
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
RETURNING id, merchant_id, name, address, contact_number, city, opening_time_minutes, closing_time_minutes, created_at, updated_at;

-- name: GetBranch :one
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
WHERE merchant_id = $1 AND id = $2
LIMIT 1;

-- name: UpdateBranch :one
UPDATE branches
SET
    name = $3,
    address = $4,
    contact_number = $5,
    city = $6,
    opening_time_minutes = $7,
    closing_time_minutes = $8,
    updated_at = NOW()
WHERE merchant_id = $1 AND id = $2
RETURNING id, merchant_id, name, address, contact_number, city, opening_time_minutes, closing_time_minutes, created_at, updated_at;
