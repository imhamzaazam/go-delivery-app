BEGIN;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM actor_roles
        GROUP BY merchant_id, actor_id
        HAVING COUNT(*) > 1
    ) THEN
        RAISE EXCEPTION 'actor_roles contains actors with multiple roles; resolve duplicates before applying 1:1 constraint';
    END IF;
END $$;

ALTER TABLE actor_roles DROP CONSTRAINT IF EXISTS actor_roles_pkey;
ALTER TABLE actor_roles ADD CONSTRAINT actor_roles_pkey PRIMARY KEY (merchant_id, actor_id);

COMMIT;
