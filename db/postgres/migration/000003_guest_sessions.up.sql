-- Allow guest checkout: sessions and carts can exist without an actor.
-- actor_id NULL means guest session/cart (no login required).
ALTER TABLE sessions ALTER COLUMN actor_id DROP NOT NULL;
