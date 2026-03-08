-- name: CreateRole :one
INSERT INTO roles (
    merchant_id,
    role_type,
    description
)
VALUES (
    $1,
    $2,
    $3
)
RETURNING id, merchant_id, role_type, description, created_at;

-- name: GetRole :one
SELECT
    id,
    merchant_id,
    role_type,
    description,
    created_at
FROM roles
WHERE merchant_id = $1 AND id = $2
LIMIT 1;

-- name: UpdateRole :one
UPDATE roles
SET
    role_type = $3,
    description = $4
WHERE merchant_id = $1 AND id = $2
RETURNING id, merchant_id, role_type, description, created_at;
