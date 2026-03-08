-- Revert: sessions must have actor_id (no guest sessions).
-- First delete any sessions with NULL actor_id.
DELETE FROM sessions WHERE actor_id IS NULL;
ALTER TABLE sessions ALTER COLUMN actor_id SET NOT NULL;
