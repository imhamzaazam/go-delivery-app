-- Change track_inventory default from true to false.
-- Merchants opt-in to inventory tracking only when needed.
ALTER TABLE products ALTER COLUMN track_inventory SET DEFAULT false;
