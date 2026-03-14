-- name: CreateArea :one
INSERT INTO areas (
    name,
    city
)
VALUES (
    $1,
    $2
)
RETURNING id, name, city, created_at;

-- name: GetArea :one
SELECT
    id,
    name,
    city,
    created_at
FROM areas
WHERE id = $1
LIMIT 1;

-- name: CreateZone :one
INSERT INTO zones (
    area_id,
    name,
    coordinates
)
VALUES (
    $1,
    $2,
    ST_GeomFromText($3, 4326)
)
RETURNING id, area_id, name, ST_AsText(coordinates) AS coordinates_wkt, created_at;

-- name: GetZone :one
SELECT
    id,
    area_id,
    name,
    ST_AsText(coordinates) AS coordinates_wkt,
    created_at
FROM zones
WHERE id = $1
LIMIT 1;

-- name: UpdateArea :one
UPDATE areas
SET
    name = $2,
    city = $3
WHERE id = $1
RETURNING id, name, city, created_at;

-- name: UpdateZone :one
UPDATE zones
SET
    area_id = $2,
    name = $3,
    coordinates = ST_GeomFromText($4, 4326)
WHERE id = $1
RETURNING id, area_id, name, ST_AsText(coordinates) AS coordinates_wkt, created_at;
