-- Revert track_inventory default to true.
ALTER TABLE products ALTER COLUMN track_inventory SET DEFAULT true;
