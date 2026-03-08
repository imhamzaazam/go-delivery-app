BEGIN;

-- ======================
-- EXTENSIONS
-- ======================
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "postgis";

-- ======================
-- ENUM TYPES
-- ======================
CREATE TYPE merchant_category AS ENUM ('restaurant', 'pharma', 'bakery');
CREATE TYPE payment_type AS ENUM ('card', 'cash');
CREATE TYPE discount_type AS ENUM ('flat', 'percentage');
CREATE TYPE order_status_type AS ENUM ('pending', 'accepted', 'out_for_delivery', 'delivered', 'refunded', 'cancelled');
CREATE TYPE subscription_status_type AS ENUM ('trial', 'active', 'suspended', 'cancelled');
CREATE TYPE role_type AS ENUM ('admin', 'merchant', 'employee', 'customer');
CREATE TYPE city_type AS ENUM ('Karachi', 'Lahore');

-- ======================
-- MERCHANTS
-- ======================
CREATE TABLE merchants (
                           id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                           name VARCHAR(255) NOT NULL,
                           ntn VARCHAR(50) NOT NULL UNIQUE,
                           address TEXT NOT NULL,
                           logo TEXT,
                           category merchant_category NOT NULL,
                           contact_number VARCHAR(14) NOT NULL,
                           created_at TIMESTAMPTZ DEFAULT NOW(),
                           updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- ======================
-- BRANCHES
-- ======================
CREATE TABLE branches (
                          id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                          merchant_id UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
                          name VARCHAR(255) NOT NULL,
                          address TEXT NOT NULL,
                          contact_number VARCHAR(14),
                          city city_type NOT NULL,
                          created_at TIMESTAMPTZ DEFAULT NOW(),
                          updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- ======================
-- ROLES
-- ======================
CREATE TABLE roles (
                       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                       merchant_id UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
                       role_type role_type NOT NULL,
                       description TEXT,
                       created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                       UNIQUE (merchant_id, id),
                       UNIQUE (merchant_id, role_type)
);

-- ======================
-- ACTORS
-- ======================
CREATE TABLE actors (
                        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                        merchant_id UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
                        email VARCHAR(255) NOT NULL,
                        password_hash TEXT NOT NULL,
                        first_name VARCHAR(255) NOT NULL,
                        last_name VARCHAR(255) NOT NULL,
                        is_active BOOLEAN NOT NULL DEFAULT true,
                        last_login TIMESTAMPTZ,
                        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                        modified_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                        UNIQUE (merchant_id, id),
                        UNIQUE (merchant_id, email)
);

-- ======================
-- ACTOR ROLES
-- ======================
CREATE TABLE actor_roles (
                             merchant_id UUID NOT NULL,
                             actor_id UUID NOT NULL,
                             role_id UUID NOT NULL,
                             assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                             PRIMARY KEY (merchant_id, actor_id),
                             FOREIGN KEY (merchant_id, actor_id) REFERENCES actors(merchant_id, id) ON DELETE CASCADE,
                             FOREIGN KEY (merchant_id, role_id) REFERENCES roles(merchant_id, id) ON DELETE CASCADE
);

-- ======================
-- SESSIONS
-- ======================
CREATE TABLE sessions (
                          id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                          merchant_id UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
                          actor_id UUID NOT NULL,
                          refresh_token TEXT NOT NULL UNIQUE,
                          user_agent VARCHAR(255) NOT NULL,
                          client_ip VARCHAR(45) NOT NULL,
                          is_blocked BOOLEAN NOT NULL DEFAULT false,
                          expires_at TIMESTAMPTZ NOT NULL,
                          created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                          FOREIGN KEY (merchant_id, actor_id) REFERENCES actors(merchant_id, id) ON DELETE CASCADE
);

-- ======================
-- PRODUCT CATEGORIES
-- ======================
CREATE TABLE product_categories (
                                    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                    merchant_id UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
                                    name VARCHAR(255) NOT NULL,
                                    description TEXT,
                                    created_at TIMESTAMPTZ DEFAULT NOW(),
                                    UNIQUE (merchant_id, name)
);

-- ======================
-- PRODUCTS
-- ======================
CREATE TABLE products (
                          id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                          merchant_id UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
                          category_id UUID REFERENCES product_categories(id) ON DELETE SET NULL,
                          name VARCHAR(255) NOT NULL,
                          description TEXT,
                          base_price NUMERIC(12,2) NOT NULL,
                          image_url TEXT,
                          track_inventory BOOLEAN NOT NULL DEFAULT true,
                          is_active BOOLEAN NOT NULL DEFAULT true,
                          created_at TIMESTAMPTZ DEFAULT NOW(),
                          updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- ======================
-- PRODUCT INVENTORY (branch-level)
-- ======================
CREATE TABLE product_inventory (
                                   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                   product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
                                   branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
                                   quantity INT NOT NULL DEFAULT 0,
                                   created_at TIMESTAMPTZ DEFAULT NOW(),
                                   updated_at TIMESTAMPTZ DEFAULT NOW(),
                                   UNIQUE (product_id, branch_id)
);

-- ======================
-- PRODUCT ADDONS
-- ======================
CREATE TABLE product_addons (
                                id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
                                name VARCHAR(255) NOT NULL,
                                price NUMERIC(12,2) NOT NULL,
                                created_at TIMESTAMPTZ DEFAULT NOW()
);

-- ======================
-- AREAS AND ZONES
-- ======================
CREATE TABLE areas (
                       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                       name VARCHAR(255) NOT NULL,
                       city city_type NOT NULL,
                       created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE zones (
                       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                       area_id UUID NOT NULL REFERENCES areas(id) ON DELETE CASCADE,
                       name VARCHAR(255) NOT NULL,
                       coordinates GEOMETRY(POLYGON, 4326) NOT NULL,
                       created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE merchant_service_zones (
                                        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                        merchant_id UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
                                        zone_id UUID NOT NULL REFERENCES zones(id) ON DELETE CASCADE,
                                        branch_id UUID REFERENCES branches(id) ON DELETE SET NULL,
                                        created_at TIMESTAMPTZ DEFAULT NOW(),
                                        UNIQUE (merchant_id, zone_id, branch_id)
);

-- ======================
-- VAT / TAX RULES
-- ======================
CREATE TABLE vat_rules (
                           id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                           merchant_id UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
                           payment_type payment_type NOT NULL,
                           rate NUMERIC(5,2) NOT NULL,
                           created_at TIMESTAMPTZ DEFAULT NOW(),
                           UNIQUE (merchant_id, payment_type)
);

-- ======================
-- MERCHANT DISCOUNTS
-- ======================
CREATE TABLE merchant_discounts (
                                    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                    merchant_id UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
                                    type discount_type NOT NULL,
                                    value NUMERIC(12,2) NOT NULL,
                                    description TEXT,
                                    valid_from TIMESTAMPTZ,
                                    valid_to TIMESTAMPTZ,
                                    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- ======================
-- CARTS AND CART ITEMS
-- ======================
CREATE TABLE carts (
                       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                       merchant_id UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
                       branch_id UUID REFERENCES branches(id) ON DELETE SET NULL,
                       actor_id UUID,
                       created_at TIMESTAMPTZ DEFAULT NOW(),
                       updated_at TIMESTAMPTZ DEFAULT NOW(),
                       FOREIGN KEY (merchant_id, actor_id) REFERENCES actors(merchant_id, id) ON DELETE RESTRICT
);

CREATE TABLE cart_items (
                            cart_id UUID NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
                            product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
                            quantity INT NOT NULL DEFAULT 1,
                            addon_ids UUID[],
                            applied_discount_id UUID REFERENCES merchant_discounts(id),
                            applied_discount_amount NUMERIC(12,2),
                            PRIMARY KEY (cart_id, product_id)
);

-- ======================
-- ORDERS AND ORDER ITEMS
-- ======================
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cart_id UUID NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
    merchant_id UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    branch_id UUID REFERENCES branches(id) ON DELETE SET NULL,
    actor_id UUID,
    payment_type payment_type NOT NULL,
    vat_rate NUMERIC(5,2),
    total_amount NUMERIC(12,2) NOT NULL,
    status order_status_type NOT NULL DEFAULT 'pending',
    delivery_address TEXT NOT NULL,
    customer_name VARCHAR(255) NOT NULL,
    customer_phone VARCHAR(14) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    FOREIGN KEY (merchant_id, actor_id) REFERENCES actors(merchant_id, id) ON DELETE RESTRICT
);

CREATE TABLE order_items (
     order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
     product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
     quantity INT NOT NULL,
     price NUMERIC(12,2) NOT NULL,
    base_amount NUMERIC(12,2) NOT NULL,
    addon_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
    discount_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
    tax_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
    line_total NUMERIC(12,2) NOT NULL,
    CHECK (price >= 0),
    CHECK (quantity > 0),
    CHECK (base_amount >= 0),
    CHECK (addon_amount >= 0),
    CHECK (discount_amount >= 0),
    CHECK (tax_amount >= 0),
    CHECK (line_total >= 0),
     PRIMARY KEY (order_id, product_id)
);

CREATE TABLE order_item_addons (
    order_id UUID NOT NULL,
    product_id UUID NOT NULL,
    addon_id UUID NOT NULL REFERENCES product_addons(id) ON DELETE RESTRICT,
    addon_name VARCHAR(255) NOT NULL,
    addon_price NUMERIC(12,2) NOT NULL,
    quantity INT NOT NULL DEFAULT 1,
    line_addon_total NUMERIC(12,2) NOT NULL,
    PRIMARY KEY (order_id, product_id, addon_id),
    FOREIGN KEY (order_id, product_id) REFERENCES order_items(order_id, product_id) ON DELETE CASCADE,
    CHECK (addon_price >= 0),
    CHECK (quantity > 0),
    CHECK (line_addon_total >= 0)
);

-- ======================
-- INDEXES
-- ======================
CREATE INDEX idx_actors_merchant_id ON actors(merchant_id);
CREATE INDEX idx_sessions_merchant_actor ON sessions(merchant_id, actor_id);
CREATE INDEX idx_roles_merchant_id ON roles(merchant_id);
CREATE INDEX idx_products_merchant_id ON products(merchant_id);
CREATE INDEX idx_zones_geom ON zones USING GIST(coordinates);
CREATE INDEX idx_inventory_product_branch ON product_inventory(product_id, branch_id);

COMMIT;
