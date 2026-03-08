BEGIN;

-- ======================================================
-- Seed: suijing.com.pk (Restaurant demo tenant)
-- Idempotent inserts via ON CONFLICT DO NOTHING
-- ======================================================

-- Merchant
INSERT INTO merchants (
  id, name, ntn, address, logo, category, contact_number
) VALUES (
  '0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11',
  'Suijing Restaurant',
  'NTN-SUIJING-78601',
  'Plot 14-C, Khayaban-e-Ittehad, DHA Phase VI, Karachi',
  'https://suijing.com.pk/assets/logo.png',
  'restaurant',
  '03123456789012'
)
ON CONFLICT (id) DO NOTHING;

-- Branches
INSERT INTO branches (
  id, merchant_id, name, address, contact_number, city
) VALUES
(
  '3a957f6d-6ee1-47fd-95e7-f96bc66c0c01',
  '0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11',
  'Suijing DHA Branch',
  '14-C Khayaban-e-Ittehad, DHA Phase VI, Karachi',
  '02134567890123',
  'Karachi'
),
(
  '3a957f6d-6ee1-47fd-95e7-f96bc66c0c02',
  '0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11',
  'Suijing Clifton Branch',
  'Boat Basin, Block 5, Clifton, Karachi',
  '02134567890124',
  'Karachi'
)
ON CONFLICT (id) DO NOTHING;

-- Roles
INSERT INTO roles (
  id, merchant_id, role_type, description
) VALUES
(
  '4bc78b3c-c99f-4267-a4b3-47c8e51a8a01',
  '0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11',
  'merchant',
  'Merchant owner role with full tenant control'
),
(
  '4bc78b3c-c99f-4267-a4b3-47c8e51a8a02',
  '0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11',
  'employee',
  'Branch operations staff'
),
(
  '4bc78b3c-c99f-4267-a4b3-47c8e51a8a03',
  '0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11',
  'customer',
  'End customer role'
)
ON CONFLICT (merchant_id, role_type) DO NOTHING;

-- Actors (password hashes are placeholders)
INSERT INTO actors (
  id, merchant_id, email, password_hash, first_name, last_name, is_active, last_login
) VALUES
(
  '52ea5f5f-d993-4052-8308-0c5bc6f27801',
  '0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11',
  'owner@suijing.com.pk',
  '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
  'Ammar',
  'Khan',
  true,
  NOW()
),
(
  '52ea5f5f-d993-4052-8308-0c5bc6f27802',
  '0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11',
  'manager.dha@suijing.com.pk',
  '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
  'Hina',
  'Ali',
  true,
  NOW()
),
(
  '52ea5f5f-d993-4052-8308-0c5bc6f27803',
  '0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11',
  'customer1@gmail.com',
  '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
  'Sara',
  'Ahmed',
  true,
  NOW()
)
ON CONFLICT (merchant_id, email) DO NOTHING;

-- Actor role mapping
INSERT INTO actor_roles (
  merchant_id, actor_id, role_id
) VALUES
(
  '0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11',
  '52ea5f5f-d993-4052-8308-0c5bc6f27801',
  '4bc78b3c-c99f-4267-a4b3-47c8e51a8a01'
),
(
  '0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11',
  '52ea5f5f-d993-4052-8308-0c5bc6f27802',
  '4bc78b3c-c99f-4267-a4b3-47c8e51a8a02'
),
(
  '0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11',
  '52ea5f5f-d993-4052-8308-0c5bc6f27803',
  '4bc78b3c-c99f-4267-a4b3-47c8e51a8a03'
)
ON CONFLICT (merchant_id, actor_id) DO UPDATE SET
  role_id = EXCLUDED.role_id,
  assigned_at = NOW();

-- Session sample
INSERT INTO sessions (
  id, merchant_id, actor_id, refresh_token, user_agent, client_ip, is_blocked, expires_at
) VALUES
(
  'feec7cb7-17d7-4e49-9dc8-992389727701',
  '0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11',
  '52ea5f5f-d993-4052-8308-0c5bc6f27803',
  'rt_suijing_demo_001',
  'Mozilla/5.0',
  '103.120.14.22',
  false,
  NOW() + INTERVAL '7 days'
)
ON CONFLICT (refresh_token) DO NOTHING;

-- Product categories
INSERT INTO product_categories (
  id, merchant_id, name, description
) VALUES
(
  '84d25368-0c95-4e77-98d6-44e4970c9301',
  '0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11',
  'Sushi Rolls',
  'Signature sushi roll selection'
),
(
  '84d25368-0c95-4e77-98d6-44e4970c9302',
  '0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11',
  'Ramen Bowls',
  'Japanese ramen variants'
),
(
  '84d25368-0c95-4e77-98d6-44e4970c9303',
  '0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11',
  'Beverages',
  'Drinks and tea'
)
ON CONFLICT (merchant_id, name) DO NOTHING;

-- Products
INSERT INTO products (
  id, merchant_id, category_id, name, description, base_price, image_url, track_inventory, is_active
) VALUES
(
  'f5adf2e2-8b67-4f07-a365-9fefac3f0b01',
  '0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11',
  '84d25368-0c95-4e77-98d6-44e4970c9301',
  'Dragon Roll',
  'Prawn tempura roll with avocado and eel sauce',
  1690.00,
  'https://suijing.com.pk/menu/dragon-roll.jpg',
  true,
  true
),
(
  'f5adf2e2-8b67-4f07-a365-9fefac3f0b02',
  '0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11',
  '84d25368-0c95-4e77-98d6-44e4970c9302',
  'Chicken Miso Ramen',
  'Miso broth with grilled chicken and soft egg',
  1450.00,
  'https://suijing.com.pk/menu/chicken-miso-ramen.jpg',
  true,
  true
),
(
  'f5adf2e2-8b67-4f07-a365-9fefac3f0b03',
  '0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11',
  '84d25368-0c95-4e77-98d6-44e4970c9303',
  'Iced Matcha Latte',
  'Cold matcha latte with milk',
  650.00,
  'https://suijing.com.pk/menu/iced-matcha-latte.jpg',
  true,
  true
)
ON CONFLICT (id) DO NOTHING;

-- Branch inventory
INSERT INTO product_inventory (
  id, product_id, branch_id, quantity
) VALUES
(
  'a53616b8-1207-4bb7-9d25-30c0c1f74201',
  'f5adf2e2-8b67-4f07-a365-9fefac3f0b01',
  '3a957f6d-6ee1-47fd-95e7-f96bc66c0c01',
  120
),
(
  'a53616b8-1207-4bb7-9d25-30c0c1f74202',
  'f5adf2e2-8b67-4f07-a365-9fefac3f0b02',
  '3a957f6d-6ee1-47fd-95e7-f96bc66c0c01',
  80
),
(
  'a53616b8-1207-4bb7-9d25-30c0c1f74203',
  'f5adf2e2-8b67-4f07-a365-9fefac3f0b01',
  '3a957f6d-6ee1-47fd-95e7-f96bc66c0c02',
  90
)
ON CONFLICT (product_id, branch_id) DO UPDATE SET
  quantity = EXCLUDED.quantity,
  updated_at = NOW();

-- Product addons
INSERT INTO product_addons (
  id, product_id, name, price
) VALUES
(
  'e7f6f82f-8dc2-40f6-9c9f-80abf5eec001',
  'f5adf2e2-8b67-4f07-a365-9fefac3f0b01',
  'Extra Wasabi',
  120.00
),
(
  'e7f6f82f-8dc2-40f6-9c9f-80abf5eec002',
  'f5adf2e2-8b67-4f07-a365-9fefac3f0b02',
  'Soft Egg',
  180.00
)
ON CONFLICT (id) DO NOTHING;

-- Areas and zones
INSERT INTO areas (
  id, name, city
) VALUES
(
  '81fd966a-cda7-4ef0-bf1f-3edf9ff8a001',
  'DHA',
  'Karachi'
),
(
  '81fd966a-cda7-4ef0-bf1f-3edf9ff8a002',
  'Clifton',
  'Karachi'
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO zones (
  id, area_id, name, coordinates
) VALUES
(
  '72be7300-2f4f-4e4d-9f1d-505df84d7001',
  '81fd966a-cda7-4ef0-bf1f-3edf9ff8a001',
  'DHA Phase VI Zone',
  ST_GeomFromText('POLYGON((67.0500 24.7900, 67.0900 24.7900, 67.0900 24.8200, 67.0500 24.8200, 67.0500 24.7900))', 4326)
),
(
  '72be7300-2f4f-4e4d-9f1d-505df84d7002',
  '81fd966a-cda7-4ef0-bf1f-3edf9ff8a002',
  'Clifton Zone',
  ST_GeomFromText('POLYGON((67.0100 24.8000, 67.0500 24.8000, 67.0500 24.8400, 67.0100 24.8400, 67.0100 24.8000))', 4326)
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO merchant_service_zones (
  id, merchant_id, zone_id, branch_id
) VALUES
(
  '9f6f3d23-af12-4a6f-af6f-37db76f35001',
  '0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11',
  '72be7300-2f4f-4e4d-9f1d-505df84d7001',
  '3a957f6d-6ee1-47fd-95e7-f96bc66c0c01'
),
(
  '9f6f3d23-af12-4a6f-af6f-37db76f35002',
  '0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11',
  '72be7300-2f4f-4e4d-9f1d-505df84d7002',
  '3a957f6d-6ee1-47fd-95e7-f96bc66c0c02'
)
ON CONFLICT (merchant_id, zone_id, branch_id) DO NOTHING;

-- VAT / tax rules
INSERT INTO vat_rules (
  id, merchant_id, payment_type, rate
) VALUES
(
  'de713163-8384-4ed5-b63f-dfbf7371f001',
  '0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11',
  'card',
  5.00
),
(
  'de713163-8384-4ed5-b63f-dfbf7371f002',
  '0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11',
  'cash',
  0.00
)
ON CONFLICT (merchant_id, payment_type) DO UPDATE SET
  rate = EXCLUDED.rate;

-- Merchant discounts
INSERT INTO merchant_discounts (
  id, merchant_id, type, value, description, valid_from, valid_to
) VALUES
(
  '7d5682a8-279b-46d4-aae0-eec42916d001',
  '0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11',
  'percentage',
  10.00,
  'Weekend Discount',
  NOW() - INTERVAL '1 day',
  NOW() + INTERVAL '30 day'
)
ON CONFLICT (id) DO NOTHING;

-- Cart + items
INSERT INTO carts (
  id, merchant_id, branch_id, actor_id
) VALUES
(
  'c978a88d-127b-4a9f-bca9-93f88d9eb001',
  '0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11',
  '3a957f6d-6ee1-47fd-95e7-f96bc66c0c01',
  '52ea5f5f-d993-4052-8308-0c5bc6f27803'
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO cart_items (
  cart_id, product_id, quantity, addon_ids, applied_discount_id, applied_discount_amount
) VALUES
(
  'c978a88d-127b-4a9f-bca9-93f88d9eb001',
  'f5adf2e2-8b67-4f07-a365-9fefac3f0b01',
  2,
  ARRAY['e7f6f82f-8dc2-40f6-9c9f-80abf5eec001'::UUID],
  '7d5682a8-279b-46d4-aae0-eec42916d001',
  338.00
)
ON CONFLICT (cart_id, product_id) DO UPDATE SET
  quantity = EXCLUDED.quantity,
  addon_ids = EXCLUDED.addon_ids,
  applied_discount_id = EXCLUDED.applied_discount_id,
  applied_discount_amount = EXCLUDED.applied_discount_amount;

-- Order + items
-- Billing math for this seed:
-- base amount = (2 * 1690.00) = 3380.00
-- addon amount = 120.00
-- discount = 338.00
-- taxable subtotal = 3162.00
-- VAT 5% = 158.10
-- order total = 3320.10
INSERT INTO orders (
  id, cart_id, merchant_id, branch_id, actor_id, payment_type, vat_rate, total_amount, status,
  delivery_address, customer_name, customer_phone
) VALUES
(
  '67be2db6-3530-4f37-b6ef-95c1a8322a01',
  'c978a88d-127b-4a9f-bca9-93f88d9eb001',
  '0f9f93c2-4fbe-4de8-a5f2-985f3f7f3a11',
  '3a957f6d-6ee1-47fd-95e7-f96bc66c0c01',
  '52ea5f5f-d993-4052-8308-0c5bc6f27803',
  'card',
  5.00,
  3320.10,
  'pending',
  'Apartment 4B, Seaview Residency, Karachi',
  'Sara Ahmed',
  '03001234567'
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO order_items (
  order_id, product_id, quantity, price, base_amount, addon_amount, discount_amount, tax_amount, line_total
) VALUES
(
  '67be2db6-3530-4f37-b6ef-95c1a8322a01',
  'f5adf2e2-8b67-4f07-a365-9fefac3f0b01',
  2,
  1690.00,
  3380.00,
  120.00,
  338.00,
  158.10,
  3320.10
)
ON CONFLICT (order_id, product_id) DO UPDATE SET
  quantity = EXCLUDED.quantity,
  price = EXCLUDED.price,
  base_amount = EXCLUDED.base_amount,
  addon_amount = EXCLUDED.addon_amount,
  discount_amount = EXCLUDED.discount_amount,
  tax_amount = EXCLUDED.tax_amount,
  line_total = EXCLUDED.line_total;

INSERT INTO order_item_addons (
  order_id, product_id, addon_id, addon_name, addon_price, quantity, line_addon_total
) VALUES
(
  '67be2db6-3530-4f37-b6ef-95c1a8322a01',
  'f5adf2e2-8b67-4f07-a365-9fefac3f0b01',
  'e7f6f82f-8dc2-40f6-9c9f-80abf5eec001',
  'Extra Wasabi',
  120.00,
  1,
  120.00
)
ON CONFLICT (order_id, product_id, addon_id) DO UPDATE SET
  addon_name = EXCLUDED.addon_name,
  addon_price = EXCLUDED.addon_price,
  quantity = EXCLUDED.quantity,
  line_addon_total = EXCLUDED.line_addon_total;

COMMIT;
