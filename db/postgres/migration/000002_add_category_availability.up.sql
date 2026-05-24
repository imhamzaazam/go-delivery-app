BEGIN;

ALTER TABLE product_categories
ADD COLUMN is_available BOOLEAN NOT NULL DEFAULT true;

COMMIT;
