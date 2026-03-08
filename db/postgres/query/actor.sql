-- name: CreateActor :one
INSERT INTO actors (
    merchant_id,
    email,
    password_hash,
    first_name,
    last_name,
    is_active,
    last_login
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
RETURNING
    id AS uid,
    email,
    trim(concat(first_name, ' ', last_name)) AS full_name,
    created_at,
    modified_at;

-- name: GetActor :one
SELECT
    a.merchant_id,
    a.id AS uid,
    a.email,
    a.password_hash AS password,
    trim(concat(a.first_name, ' ', a.last_name)) AS full_name,
    a.is_active,
    a.last_login,
    a.created_at,
    a.modified_at
FROM actors a
WHERE a.merchant_id = $1
  AND a.email = $2
LIMIT 1;

-- name: GetActorByUID :one
SELECT
    a.merchant_id,
    a.id AS uid,
    a.email,
    a.password_hash AS password,
    trim(concat(a.first_name, ' ', a.last_name)) AS full_name,
    a.is_active,
    a.last_login,
    a.created_at,
    a.modified_at
FROM actors a
WHERE a.merchant_id = $1
  AND a.id = $2
LIMIT 1;

-- name: UpdateActor :one
UPDATE actors
SET
        first_name = $3,
        last_name = $4,
        is_active = $5,
        last_login = $6,
        modified_at = NOW()
WHERE merchant_id = $1
    AND id = $2
RETURNING
        merchant_id,
        id AS uid,
        email,
        password_hash AS password,
        trim(concat(first_name, ' ', last_name)) AS full_name,
        is_active,
        last_login,
        created_at,
        modified_at;
