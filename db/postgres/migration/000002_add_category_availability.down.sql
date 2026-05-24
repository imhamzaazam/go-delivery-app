BEGIN;

ALTER TABLE product_categories
DROP COLUMN is_available;

COMMIT;
