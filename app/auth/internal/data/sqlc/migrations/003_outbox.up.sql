-- +goose Up

-- Outbox table for transactional event publishing to NATS.
CREATE TABLE auth.outbox (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    topic         TEXT NOT NULL,
    payload       BYTEA NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at  TIMESTAMPTZ,
    retry_count   INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX idx_outbox_unpublished ON auth.outbox(published_at, created_at ASC)
    WHERE published_at IS NULL AND retry_count < 3;
