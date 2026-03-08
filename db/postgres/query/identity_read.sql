-- name: GetActorProfileByMerchantAndEmail :one
SELECT
    a.merchant_id,
    a.id AS uid,
    a.email,
    trim(concat(a.first_name, ' ', a.last_name)) AS full_name,
    a.is_active,
    a.last_login,
    a.created_at,
    a.modified_at
FROM actors a
WHERE a.merchant_id = $1
  AND a.email = $2
LIMIT 1;

-- name: ListActorsByMerchant :many
SELECT
    a.merchant_id,
    a.id AS uid,
    a.email,
    trim(concat(a.first_name, ' ', a.last_name)) AS full_name,
    a.is_active,
    a.last_login,
    a.created_at,
    a.modified_at
FROM actors a
WHERE a.merchant_id = $1
ORDER BY a.created_at DESC, a.id;

-- name: ListEmployeesByMerchant :many
SELECT DISTINCT
    a.merchant_id,
    a.id AS uid,
    a.email,
    trim(concat(a.first_name, ' ', a.last_name)) AS full_name,
    a.is_active,
    a.last_login,
    a.created_at,
    a.modified_at
FROM actors a
JOIN actor_roles ar
  ON ar.merchant_id = a.merchant_id
 AND ar.actor_id = a.id
JOIN roles r
  ON r.merchant_id = ar.merchant_id
 AND r.id = ar.role_id
WHERE a.merchant_id = $1
  AND r.role_type = 'employee'
ORDER BY a.created_at DESC, a.id;
