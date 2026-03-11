-- name: CreateOrder :one
INSERT INTO orders (
    cart_id,
    merchant_id,
    branch_id,
    actor_id,
    payment_type,
    vat_rate,
    total_amount,
    status,
    delivery_address,
    customer_name,
    customer_phone
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10,
    $11
)
RETURNING id, cart_id, merchant_id, branch_id, actor_id, payment_type, vat_rate, total_amount, status, delivery_address, customer_name, customer_phone, created_at, updated_at;

-- name: CreateOrderGuest :one
INSERT INTO orders (
    cart_id,
    merchant_id,
    branch_id,
    actor_id,
    payment_type,
    vat_rate,
    total_amount,
    status,
    delivery_address,
    customer_name,
    customer_phone
)
VALUES (
    $1,
    $2,
    $3,
    NULL,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10
)
RETURNING id, cart_id, merchant_id, branch_id, actor_id, payment_type, vat_rate, total_amount, status, delivery_address, customer_name, customer_phone, created_at, updated_at;

-- name: GetOrder :one
SELECT
    id,
    cart_id,
    merchant_id,
    branch_id,
    actor_id,
    payment_type,
    vat_rate,
    total_amount,
    status,
    delivery_address,
    customer_name,
    customer_phone,
    created_at,
    updated_at
FROM orders
WHERE merchant_id = $1 AND id = $2
LIMIT 1;

-- name: UpdateOrder :one
UPDATE orders
SET
    branch_id = $3,
    actor_id = $4,
    payment_type = $5,
    vat_rate = $6,
    total_amount = $7,
    status = $8,
    delivery_address = $9,
    customer_name = $10,
    customer_phone = $11,
    updated_at = NOW()
WHERE merchant_id = $1 AND id = $2
RETURNING id, cart_id, merchant_id, branch_id, actor_id, payment_type, vat_rate, total_amount, status, delivery_address, customer_name, customer_phone, created_at, updated_at;
