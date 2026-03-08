-- name: CreateMerchantServiceZone :one
INSERT INTO merchant_service_zones (
    merchant_id,
    zone_id,
    branch_id
)
VALUES (
    $1,
    $2,
    $3
)
RETURNING id, merchant_id, zone_id, branch_id, created_at;

-- name: GetMerchantServiceZone :one
SELECT
    id,
    merchant_id,
    zone_id,
    branch_id,
    created_at
FROM merchant_service_zones
WHERE id = $1
LIMIT 1;

-- name: UpdateMerchantServiceZone :one
UPDATE merchant_service_zones
SET
    branch_id = $2
WHERE id = $1
RETURNING id, merchant_id, zone_id, branch_id, created_at;
