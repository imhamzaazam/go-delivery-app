-- Invoice breakdown for a specific order
-- Replace this UUID with your target order ID
-- Example: '67be2db6-3530-4f37-b6ef-95c1a8322a01'

WITH target_order AS (
  SELECT
    o.id,
    o.merchant_id,
    m.name AS merchant_name,
    o.payment_type,
    o.status,
    o.vat_rate,
    o.total_amount,
    o.delivery_address,
    o.customer_name,
    o.customer_phone,
    o.created_at
  FROM orders o
  JOIN merchants m ON m.id = o.merchant_id
  WHERE o.id = '67be2db6-3530-4f37-b6ef-95c1a8322a01'::uuid
),
item_lines AS (
  SELECT
    oi.order_id,
    oi.product_id,
    p.name AS product_name,
    oi.quantity,
    oi.price AS unit_price,
    oi.base_amount,
    oi.addon_amount,
    oi.discount_amount,
    oi.tax_amount,
    oi.line_total
  FROM order_items oi
  JOIN products p ON p.id = oi.product_id
  JOIN target_order t ON t.id = oi.order_id
),
item_addons AS (
  SELECT
    oia.order_id,
    oia.product_id,
    oia.addon_id,
    oia.addon_name,
    oia.addon_price,
    oia.quantity AS addon_qty,
    oia.line_addon_total
  FROM order_item_addons oia
  JOIN target_order t ON t.id = oia.order_id
),
order_totals AS (
  SELECT
    il.order_id,
    ROUND(SUM(il.base_amount), 2) AS base_total,
    ROUND(SUM(il.addon_amount), 2) AS addon_total,
    ROUND(SUM(il.discount_amount), 2) AS discount_total,
    ROUND(SUM(il.tax_amount), 2) AS tax_total,
    ROUND(SUM(il.line_total), 2) AS lines_grand_total
  FROM item_lines il
  GROUP BY il.order_id
)
SELECT
  t.id AS order_id,
  t.merchant_id,
  t.merchant_name,
  t.payment_type,
  t.status,
  t.customer_name,
  t.customer_phone,
  t.delivery_address,
  t.vat_rate,
  t.total_amount AS order_header_total,
  ot.base_total,
  ot.addon_total,
  ot.discount_total,
  ot.tax_total,
  ot.lines_grand_total,
  (t.total_amount - COALESCE(ot.lines_grand_total, 0)) AS header_vs_lines_diff,
  t.created_at
FROM target_order t
LEFT JOIN order_totals ot ON ot.order_id = t.id;

-- Per-item breakdown
SELECT
  il.order_id,
  il.product_id,
  il.product_name,
  il.quantity,
  il.unit_price,
  il.base_amount,
  il.addon_amount,
  il.discount_amount,
  il.tax_amount,
  il.line_total
FROM item_lines il
ORDER BY il.product_name;

-- Per-item addon breakdown
SELECT
  ia.order_id,
  ia.product_id,
  p.name AS product_name,
  ia.addon_id,
  ia.addon_name,
  ia.addon_price,
  ia.addon_qty,
  ia.line_addon_total
FROM item_addons ia
JOIN products p ON p.id = ia.product_id
ORDER BY p.name, ia.addon_name;
