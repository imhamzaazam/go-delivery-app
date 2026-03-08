-- Checkout query: move cart -> order snapshot (header + items + addons)
--
-- Usage (psql example):
-- psql "$DB_SOURCE" \
--   -v cart_id="c978a88d-127b-4a9f-bca9-93f88d9eb001" \
--   -v payment_type="card" \
--   -v delivery_address="Apartment 4B, Seaview Residency, Karachi" \
--   -v customer_name="Sara Ahmed" \
--   -v customer_phone="03001234567" \
--   -v cleanup_cart="true" \
--   -f scripts/sql/checkout_from_cart.sql
--
-- Notes:
-- - VAT rate is pulled from vat_rules by merchant + payment_type.
-- - If no vat_rules row exists, VAT defaults to 0.
-- - Uses one transaction and locks cart row FOR UPDATE.

BEGIN;

WITH locked_cart AS (
  SELECT c.*
  FROM carts c
  WHERE c.id = :'cart_id'::uuid
  FOR UPDATE
),
vat_ctx AS (
  SELECT
    lc.id AS cart_id,
    lc.merchant_id,
    COALESCE(vr.rate, 0)::numeric(5,2) AS vat_rate
  FROM locked_cart lc
  LEFT JOIN vat_rules vr
    ON vr.merchant_id = lc.merchant_id
   AND vr.payment_type = :'payment_type'::payment_type
),
cart_lines AS (
  SELECT
    lc.id AS cart_id,
    ci.product_id,
    ci.quantity,
    p.base_price AS unit_price,
    (ci.quantity * p.base_price)::numeric(12,2) AS base_amount,
    COALESCE(ci.applied_discount_amount, 0)::numeric(12,2) AS discount_amount,
    COALESCE(ci.addon_ids, ARRAY[]::uuid[]) AS addon_ids
  FROM locked_cart lc
  JOIN cart_items ci ON ci.cart_id = lc.id
  JOIN products p ON p.id = ci.product_id
),
addon_expanded AS (
  SELECT
    cl.cart_id,
    cl.product_id,
    pa.id AS addon_id,
    pa.name AS addon_name,
    pa.price AS addon_price,
    1::int AS addon_qty,
    pa.price::numeric(12,2) AS line_addon_total
  FROM cart_lines cl
  JOIN LATERAL unnest(cl.addon_ids) a(addon_id) ON true
  JOIN product_addons pa ON pa.id = a.addon_id
),
addon_totals AS (
  SELECT
    ae.cart_id,
    ae.product_id,
    COALESCE(SUM(ae.line_addon_total), 0)::numeric(12,2) AS addon_amount
  FROM addon_expanded ae
  GROUP BY ae.cart_id, ae.product_id
),
line_calc AS (
  SELECT
    cl.cart_id,
    cl.product_id,
    cl.quantity,
    cl.unit_price,
    cl.base_amount,
    COALESCE(at.addon_amount, 0)::numeric(12,2) AS addon_amount,
    cl.discount_amount,
    vc.vat_rate,
    ROUND(((cl.base_amount + COALESCE(at.addon_amount, 0) - cl.discount_amount) * vc.vat_rate) / 100.0, 2)::numeric(12,2) AS tax_amount,
    ROUND((cl.base_amount + COALESCE(at.addon_amount, 0) - cl.discount_amount) + (((cl.base_amount + COALESCE(at.addon_amount, 0) - cl.discount_amount) * vc.vat_rate) / 100.0), 2)::numeric(12,2) AS line_total
  FROM cart_lines cl
  LEFT JOIN addon_totals at
    ON at.cart_id = cl.cart_id
   AND at.product_id = cl.product_id
  JOIN vat_ctx vc ON vc.cart_id = cl.cart_id
),
order_totals AS (
  SELECT
    lc.id AS cart_id,
    COALESCE(SUM(lc2.line_total), 0)::numeric(12,2) AS order_total,
    COALESCE(MAX(lc2.vat_rate), 0)::numeric(5,2) AS vat_rate
  FROM line_calc lc2
  JOIN locked_cart lc ON lc.id = lc2.cart_id
  GROUP BY lc.id
),
new_order AS (
  INSERT INTO orders (
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
    customer_phone
  )
  SELECT
    gen_random_uuid(),
    lc.id,
    lc.merchant_id,
    lc.branch_id,
    lc.actor_id,
    :'payment_type'::payment_type,
    ot.vat_rate,
    ot.order_total,
    'pending'::order_status_type,
    :'delivery_address',
    :'customer_name',
    :'customer_phone'
  FROM locked_cart lc
  JOIN order_totals ot ON ot.cart_id = lc.id
  RETURNING id, cart_id
),
insert_order_items AS (
  INSERT INTO order_items (
    order_id,
    product_id,
    quantity,
    price,
    base_amount,
    addon_amount,
    discount_amount,
    tax_amount,
    line_total
  )
  SELECT
    no.id,
    lc.product_id,
    lc.quantity,
    lc.unit_price,
    lc.base_amount,
    lc.addon_amount,
    lc.discount_amount,
    lc.tax_amount,
    lc.line_total
  FROM line_calc lc
  JOIN new_order no ON no.cart_id = lc.cart_id
  RETURNING order_id, product_id
),
insert_order_item_addons AS (
  INSERT INTO order_item_addons (
    order_id,
    product_id,
    addon_id,
    addon_name,
    addon_price,
    quantity,
    line_addon_total
  )
  SELECT
    no.id,
    ae.product_id,
    ae.addon_id,
    ae.addon_name,
    ae.addon_price,
    ae.addon_qty,
    ae.line_addon_total
  FROM addon_expanded ae
  JOIN new_order no ON no.cart_id = ae.cart_id
  RETURNING order_id, product_id, addon_id
),
cleanup AS (
  DELETE FROM cart_items ci
  WHERE ci.cart_id = :'cart_id'::uuid
    AND COALESCE(:'cleanup_cart', 'false')::boolean = true
  RETURNING ci.cart_id
)
SELECT
  no.id AS new_order_id,
  no.cart_id,
  (SELECT COUNT(*) FROM insert_order_items) AS inserted_items,
  (SELECT COUNT(*) FROM insert_order_item_addons) AS inserted_addons
FROM new_order no;

COMMIT;
