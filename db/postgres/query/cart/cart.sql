-- name: CreateCart :one
INSERT INTO carts (
    id,
    merchant_id,
    branch_id,
    actor_id
)
VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING id, merchant_id, branch_id, actor_id, created_at, updated_at;

-- name: CreateGuestCart :one
INSERT INTO carts (
    id,
    merchant_id,
    branch_id,
    actor_id
)
VALUES (
    $1,
    $2,
    $3,
    NULL
)
RETURNING id, merchant_id, branch_id, actor_id, created_at, updated_at;

-- name: GetCart :one
SELECT
    id,
    merchant_id,
    branch_id,
    actor_id,
    created_at,
    updated_at
FROM carts
WHERE merchant_id = $1 AND id = $2
LIMIT 1;

-- name: UpdateCart :one
UPDATE carts
SET
    branch_id = $3,
    actor_id = $4,
    updated_at = NOW()
WHERE merchant_id = $1 AND id = $2
RETURNING id, merchant_id, branch_id, actor_id, created_at, updated_at;
