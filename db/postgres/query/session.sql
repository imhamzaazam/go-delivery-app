-- name: CreateSession :one
WITH inserted AS (
	INSERT INTO sessions (
		id,
		merchant_id,
		actor_id,
		refresh_token,
		user_agent,
		client_ip,
		is_blocked,
		expires_at
	)
	VALUES (
		$1,
		$2,
		$3,
		$4,
		$5,
		$6,
		$7,
		$8
	)
	RETURNING id, merchant_id, actor_id, refresh_token, user_agent, client_ip, is_blocked, expires_at, created_at
)
SELECT
	inserted.id AS id,
	inserted.id AS uid,
	a.email AS actor_email,
	inserted.refresh_token,
	inserted.user_agent,
	inserted.client_ip,
	inserted.is_blocked,
	inserted.expires_at,
	inserted.created_at
FROM inserted
JOIN actors a
  ON a.merchant_id = inserted.merchant_id
 AND a.id = inserted.actor_id;

-- name: GetSession :one
SELECT
	s.id AS id,
	s.id AS uid,
	COALESCE(a.email, '') AS actor_email,
	s.refresh_token,
	s.user_agent,
	s.client_ip,
	s.is_blocked,
	s.expires_at,
	s.created_at
FROM sessions s
LEFT JOIN actors a ON a.merchant_id = s.merchant_id AND a.id = s.actor_id
WHERE s.id = $1 LIMIT 1;

-- name: CreateGuestSession :one
INSERT INTO sessions (
	id,
	merchant_id,
	actor_id,
	refresh_token,
	user_agent,
	client_ip,
	is_blocked,
	expires_at
)
VALUES (
	$1,
	$2,
	NULL,
	$3,
	$4,
	$5,
	$6,
	$7
)
RETURNING id, merchant_id, actor_id, refresh_token, user_agent, client_ip, is_blocked, expires_at, created_at;

-- name: UpdateSessionRefresh :one
UPDATE sessions
SET
	refresh_token = $2,
	user_agent = $3,
	client_ip = $4,
	expires_at = $5
WHERE id = $1
RETURNING id, merchant_id, actor_id, refresh_token, user_agent, client_ip, is_blocked, expires_at, created_at;

-- name: UpdateSessionBlockStatus :one
UPDATE sessions
SET
	is_blocked = $2
WHERE id = $1
RETURNING id, merchant_id, actor_id, refresh_token, user_agent, client_ip, is_blocked, expires_at, created_at;