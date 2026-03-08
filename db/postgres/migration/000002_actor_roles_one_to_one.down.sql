BEGIN;

ALTER TABLE actor_roles DROP CONSTRAINT IF EXISTS actor_roles_pkey;
ALTER TABLE actor_roles ADD CONSTRAINT actor_roles_pkey PRIMARY KEY (merchant_id, actor_id, role_id);

COMMIT;
