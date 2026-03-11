-- name: GetMonthlySalesReport :one
SELECT
    COALESCE(SUM(oi.line_total), 0)::numeric AS total_sales,
    COALESCE(SUM(oi.tax_amount), 0)::numeric AS total_tax,
    COALESCE(SUM(oi.discount_amount), 0)::numeric AS total_discount,
    COALESCE(SUM(oi.base_amount + oi.addon_amount - oi.discount_amount), 0)::numeric AS profit_estimate
FROM orders o
JOIN order_items oi
  ON oi.order_id = o.id
WHERE o.merchant_id = sqlc.arg(merchant_id)
  AND o.status IN ('accepted', 'out_for_delivery', 'delivered')
  AND EXTRACT(MONTH FROM o.created_at)::int = sqlc.arg(month)::int
  AND EXTRACT(YEAR FROM o.created_at)::int = sqlc.arg(year)::int;
