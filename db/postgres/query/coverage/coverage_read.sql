-- name: ListAreas :many
SELECT
    id,
    name,
    city,
    created_at
FROM areas
ORDER BY city, name, id;

-- name: ListZonesByArea :many
SELECT
    id,
    area_id,
    name,
    ST_AsText(coordinates) AS coordinates_wkt,
    created_at
FROM zones
WHERE area_id = $1
ORDER BY name, id;

-- name: ListMerchantServiceZonesByMerchant :many
SELECT
    msz.id,
    msz.merchant_id,
    msz.zone_id,
    msz.branch_id,
    msz.created_at,
    z.name AS zone_name,
    ST_AsText(z.coordinates) AS zone_coordinates_wkt,
    a.id AS area_id,
    a.name AS area_name,
    a.city AS area_city,
    COALESCE(b.name, '') AS branch_name
FROM merchant_service_zones msz
JOIN zones z
  ON z.id = msz.zone_id
JOIN areas a
  ON a.id = z.area_id
LEFT JOIN branches b
  ON b.id = msz.branch_id
WHERE msz.merchant_id = $1
ORDER BY msz.created_at DESC, msz.id;
