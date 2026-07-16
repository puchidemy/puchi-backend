-- +goose Up
ALTER TABLE core.users ADD COLUMN age_range TEXT NOT NULL DEFAULT '';
ALTER TABLE core.users ADD COLUMN onboarding_completed BOOLEAN NOT NULL DEFAULT false;

CREATE TABLE core.user_onboarding (
    user_id    TEXT PRIMARY KEY REFERENCES core.users(id) ON DELETE CASCADE,
    how_heard  TEXT NOT NULL DEFAULT '',
    why_learn  TEXT NOT NULL DEFAULT '',
    level      TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
