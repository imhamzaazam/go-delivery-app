BEGIN;

-- ======================
-- DROP TABLES (reverse dependency order)
-- ======================
DROP INDEX IF EXISTS idx_inventory_product_branch;
DROP INDEX IF EXISTS idx_zones_geom;
DROP INDEX IF EXISTS idx_products_merchant_id;
DROP INDEX IF EXISTS idx_roles_merchant_id;
DROP INDEX IF EXISTS idx_sessions_merchant_actor;
DROP INDEX IF EXISTS idx_actors_merchant_id;

DROP TABLE IF EXISTS order_item_addons CASCADE;
DROP TABLE IF EXISTS order_items CASCADE;
DROP TABLE IF EXISTS orders CASCADE;

DROP TABLE IF EXISTS cart_items CASCADE;
DROP TABLE IF EXISTS carts CASCADE;

DROP TABLE IF EXISTS merchant_discounts CASCADE;
DROP TABLE IF EXISTS vat_rules CASCADE;

DROP TABLE IF EXISTS merchant_service_zones CASCADE;
DROP TABLE IF EXISTS zones CASCADE;
DROP TABLE IF EXISTS areas CASCADE;

DROP TABLE IF EXISTS product_addons CASCADE;
DROP TABLE IF EXISTS product_inventory CASCADE;
DROP TABLE IF EXISTS products CASCADE;
DROP TABLE IF EXISTS product_categories CASCADE;

DROP TABLE IF EXISTS sessions CASCADE;
DROP TABLE IF EXISTS actor_roles CASCADE;
DROP TABLE IF EXISTS actors CASCADE;
DROP TABLE IF EXISTS roles CASCADE;

DROP TABLE IF EXISTS branches CASCADE;
DROP TABLE IF EXISTS merchants CASCADE;

-- ======================
-- DROP ENUM TYPES
-- ======================
DROP TYPE IF EXISTS city_type;
DROP TYPE IF EXISTS role_type;
DROP TYPE IF EXISTS subscription_status_type;
DROP TYPE IF EXISTS order_status_type;
DROP TYPE IF EXISTS discount_type;
DROP TYPE IF EXISTS payment_type;
DROP TYPE IF EXISTS merchant_category;

-- ======================
-- DROP EXTENSIONS
-- ======================
DROP EXTENSION IF EXISTS postgis;
DROP EXTENSION IF EXISTS pgcrypto;

COMMIT;
