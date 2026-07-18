-- +goose Up
CREATE TABLE IF NOT EXISTS core.user_settings (
  user_id TEXT PRIMARY KEY,
  sound_effects BOOLEAN NOT NULL DEFAULT true,
  animations BOOLEAN NOT NULL DEFAULT true,
  motivational_messages BOOLEAN NOT NULL DEFAULT true,
  listening_exercises BOOLEAN NOT NULL DEFAULT true,
  theme TEXT NOT NULL DEFAULT 'system',
  locale TEXT NOT NULL DEFAULT 'en',
  privacy_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
