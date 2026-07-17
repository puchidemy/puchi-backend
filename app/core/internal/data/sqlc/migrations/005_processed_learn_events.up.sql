-- +goose Up
CREATE TABLE core.processed_learn_events (
    event_type   TEXT NOT NULL CHECK (event_type IN ('lesson', 'unit')),
    user_id      TEXT NOT NULL REFERENCES core.users(id) ON DELETE CASCADE,
    source_id    TEXT NOT NULL,
    xp_applied   INT NOT NULL DEFAULT 0,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (event_type, user_id, source_id)
);

CREATE INDEX idx_processed_learn_events_user ON core.processed_learn_events(user_id);

-- +goose Down
DROP TABLE IF EXISTS core.processed_learn_events;
