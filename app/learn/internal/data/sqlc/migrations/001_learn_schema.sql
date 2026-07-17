-- +goose Up
CREATE SCHEMA IF NOT EXISTS learn;

CREATE TABLE learn.courses (
  id UUID PRIMARY KEY DEFAULT uuidv7(),
  slug TEXT NOT NULL UNIQUE,
  title TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE learn.units (
  id UUID PRIMARY KEY DEFAULT uuidv7(),
  course_id UUID NOT NULL REFERENCES learn.courses(id),
  position INT NOT NULL,
  title TEXT NOT NULL,
  UNIQUE (course_id, position)
);

CREATE TABLE learn.skills (
  id UUID PRIMARY KEY DEFAULT uuidv7(),
  unit_id UUID NOT NULL REFERENCES learn.units(id),
  position INT NOT NULL,
  title TEXT NOT NULL,
  UNIQUE (unit_id, position)
);

CREATE TABLE learn.lessons (
  id UUID PRIMARY KEY DEFAULT uuidv7(),
  skill_id UUID NOT NULL REFERENCES learn.skills(id),
  position INT NOT NULL,
  title TEXT NOT NULL,
  xp_reward INT NOT NULL DEFAULT 10,
  required BOOLEAN NOT NULL DEFAULT true,
  UNIQUE (skill_id, position)
);

CREATE TABLE learn.exercises (
  id UUID PRIMARY KEY DEFAULT uuidv7(),
  lesson_id UUID NOT NULL REFERENCES learn.lessons(id),
  position INT NOT NULL,
  type TEXT NOT NULL CHECK (type IN ('select','match','listen','dictate')),
  prompt JSONB NOT NULL,
  answer JSONB NOT NULL,
  UNIQUE (lesson_id, position)
);

CREATE TABLE learn.exercise_assets (
  id BIGSERIAL PRIMARY KEY,
  exercise_id UUID NOT NULL REFERENCES learn.exercises(id) ON DELETE CASCADE,
  media_id BIGINT,
  object_key TEXT,
  role TEXT NOT NULL DEFAULT 'primary'
);

CREATE TABLE learn.guests (
  id UUID PRIMARY KEY,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  claimed_at TIMESTAMPTZ,
  claimed_user_id TEXT
);

CREATE TABLE learn.user_lesson_progress (
  owner_type TEXT NOT NULL CHECK (owner_type IN ('user','guest')),
  owner_id TEXT NOT NULL,
  lesson_id UUID NOT NULL REFERENCES learn.lessons(id),
  status TEXT NOT NULL CHECK (status IN ('not_started','in_progress','completed')),
  xp_earned INT NOT NULL DEFAULT 0,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (owner_type, owner_id, lesson_id)
);

CREATE TABLE learn.user_unit_progress (
  owner_type TEXT NOT NULL CHECK (owner_type IN ('user','guest')),
  owner_id TEXT NOT NULL,
  unit_id UUID NOT NULL REFERENCES learn.units(id),
  status TEXT NOT NULL CHECK (status IN ('not_started','in_progress','completed')),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (owner_type, owner_id, unit_id)
);

CREATE TABLE learn.attempts (
  id UUID PRIMARY KEY DEFAULT uuidv7(),
  owner_type TEXT NOT NULL,
  owner_id TEXT NOT NULL,
  lesson_id UUID NOT NULL REFERENCES learn.lessons(id),
  status TEXT NOT NULL DEFAULT 'active',
  session_xp INT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  completed_at TIMESTAMPTZ
);

CREATE TABLE learn.attempt_answers (
  id BIGSERIAL PRIMARY KEY,
  attempt_id UUID NOT NULL REFERENCES learn.attempts(id) ON DELETE CASCADE,
  exercise_id UUID NOT NULL REFERENCES learn.exercises(id),
  payload JSONB NOT NULL,
  correct BOOLEAN NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS learn.attempt_answers;
DROP TABLE IF EXISTS learn.attempts;
DROP TABLE IF EXISTS learn.user_unit_progress;
DROP TABLE IF EXISTS learn.user_lesson_progress;
DROP TABLE IF EXISTS learn.guests;
DROP TABLE IF EXISTS learn.exercise_assets;
DROP TABLE IF EXISTS learn.exercises;
DROP TABLE IF EXISTS learn.lessons;
DROP TABLE IF EXISTS learn.skills;
DROP TABLE IF EXISTS learn.units;
DROP TABLE IF EXISTS learn.courses;
DROP SCHEMA IF EXISTS learn;
