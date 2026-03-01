BEGIN;

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ======================
-- ACTORS
-- ======================
CREATE TABLE actors (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        email VARCHAR(255) NOT NULL UNIQUE,
        password_hash TEXT NOT NULL,
        first_name VARCHAR(255) NOT NULL,
        last_name VARCHAR(255) NOT NULL,
        is_active BOOLEAN NOT NULL DEFAULT true,
        last_login TIMESTAMPTZ,
        created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
        modified_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ======================
-- ROLES
-- ======================
CREATE TABLE roles (
       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
       name VARCHAR(50) NOT NULL UNIQUE,
       description TEXT,
       created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ======================
-- ACTOR_ROLES
-- ======================
CREATE TABLE actor_roles (
     actor_id UUID NOT NULL,
     role_id UUID NOT NULL,
     assigned_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

     PRIMARY KEY (actor_id, role_id),

     FOREIGN KEY (actor_id)
         REFERENCES actors(id)
         ON DELETE CASCADE,

     FOREIGN KEY (role_id)
         REFERENCES roles(id)
         ON DELETE CASCADE,
);

-- ======================
-- SESSIONS
-- ======================
CREATE TABLE sessions (
          id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
          actor_id UUID NOT NULL,
          refresh_token TEXT NOT NULL,
          user_agent VARCHAR(255) NOT NULL,
          client_ip VARCHAR(45) NOT NULL,
          is_blocked BOOLEAN NOT NULL DEFAULT false,
          expires_at TIMESTAMPTZ NOT NULL,
          created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

          FOREIGN KEY (actor_id)
              REFERENCES actors(id)
              ON DELETE CASCADE
);

-- Index for session lookup
CREATE INDEX idx_sessions_actor_id ON sessions(actor_id);



COMMIT;
