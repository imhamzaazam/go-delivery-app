-- Quick verification queries for Suijing demo tenant

SELECT id, name, ntn, category, contact_number, created_at
FROM merchants
WHERE id = '0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11';

SELECT b.id, b.name, b.city, b.contact_number
FROM branches b
WHERE b.merchant_id = '0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11'
ORDER BY b.name;

SELECT a.id, a.email, a.first_name, a.last_name, r.role_type
FROM actors a
LEFT JOIN actor_roles ar ON ar.actor_id = a.id AND ar.merchant_id = a.merchant_id
LEFT JOIN roles r ON r.id = ar.role_id AND r.merchant_id = a.merchant_id
WHERE a.merchant_id = '0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11'
ORDER BY a.email;

SELECT p.id, p.name, p.base_price, c.name AS category, p.is_active
FROM products p
LEFT JOIN product_categories c ON c.id = p.category_id
WHERE p.merchant_id = '0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11'
ORDER BY p.name;

SELECT o.id, o.status, o.payment_type, o.total_amount, o.customer_name, o.created_at
FROM orders o
WHERE o.merchant_id = '0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11'
ORDER BY o.created_at DESC;
