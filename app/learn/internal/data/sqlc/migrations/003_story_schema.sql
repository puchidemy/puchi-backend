-- +goose Up
-- Story-first curriculum (additive; legacy units/skills/lessons remain).

CREATE TABLE learn.cities (
  id UUID PRIMARY KEY DEFAULT uuidv7(),
  slug TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  position INT NOT NULL,
  map_x REAL NOT NULL,
  map_y REAL NOT NULL,
  cover_url TEXT,
  blurb TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE learn.stories (
  id UUID PRIMARY KEY DEFAULT uuidv7(),
  city_id UUID NOT NULL REFERENCES learn.cities(id),
  slug TEXT NOT NULL,
  title TEXT NOT NULL,
  summary TEXT NOT NULL DEFAULT '',
  cover_url TEXT,
  cefr TEXT NOT NULL CHECK (cefr IN ('A1','A2','B1','B2','C1','C2')),
  tags TEXT[] NOT NULL DEFAULT '{}',
  audio_url TEXT,
  vocab_focus TEXT[] NOT NULL DEFAULT '{}',
  grammar_focus TEXT[] NOT NULL DEFAULT '{}',
  cultural_fact TEXT,
  est_minutes INT,
  position INT NOT NULL,
  status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','published')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (city_id, slug),
  UNIQUE (city_id, position)
);

CREATE TABLE learn.scenes (
  id UUID PRIMARY KEY DEFAULT uuidv7(),
  story_id UUID NOT NULL REFERENCES learn.stories(id),
  position INT NOT NULL,
  title TEXT,
  narration TEXT NOT NULL,
  dialogue_json JSONB,
  illustration_url TEXT,
  audio_url TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (story_id, position)
);

CREATE TABLE learn.activities (
  id UUID PRIMARY KEY DEFAULT uuidv7(),
  scene_id UUID NOT NULL REFERENCES learn.scenes(id),
  position INT NOT NULL,
  type TEXT NOT NULL CHECK (type IN ('select','match','listen','dictate')),
  prompt JSONB NOT NULL,
  answer JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (scene_id, position)
);

CREATE TABLE learn.user_story_progress (
  owner_type TEXT NOT NULL CHECK (owner_type IN ('user','guest')),
  owner_id TEXT NOT NULL,
  story_id UUID NOT NULL REFERENCES learn.stories(id),
  status TEXT NOT NULL CHECK (status IN ('not_started','in_progress','completed')),
  xp_earned INT NOT NULL DEFAULT 0,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (owner_type, owner_id, story_id)
);

CREATE TABLE learn.user_scene_progress (
  owner_type TEXT NOT NULL CHECK (owner_type IN ('user','guest')),
  owner_id TEXT NOT NULL,
  scene_id UUID NOT NULL REFERENCES learn.scenes(id),
  status TEXT NOT NULL CHECK (status IN ('not_started','in_progress','completed')),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (owner_type, owner_id, scene_id)
);

CREATE TABLE learn.activity_attempts (
  id UUID PRIMARY KEY DEFAULT uuidv7(),
  owner_type TEXT NOT NULL,
  owner_id TEXT NOT NULL,
  story_id UUID NOT NULL REFERENCES learn.stories(id),
  scene_id UUID NOT NULL REFERENCES learn.scenes(id),
  status TEXT NOT NULL DEFAULT 'active',
  session_xp INT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  completed_at TIMESTAMPTZ
);

CREATE TABLE learn.activity_attempt_answers (
  id BIGSERIAL PRIMARY KEY,
  attempt_id UUID NOT NULL REFERENCES learn.activity_attempts(id) ON DELETE CASCADE,
  activity_id UUID NOT NULL REFERENCES learn.activities(id),
  payload JSONB NOT NULL,
  correct BOOLEAN NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX learn_stories_city_id_idx ON learn.stories (city_id);
CREATE INDEX learn_scenes_story_id_idx ON learn.scenes (story_id);
CREATE INDEX learn_activities_scene_id_idx ON learn.activities (scene_id);
CREATE INDEX learn_activity_attempts_owner_scene_idx
  ON learn.activity_attempts (owner_type, owner_id, scene_id);

-- +goose Down
DROP TABLE IF EXISTS learn.activity_attempt_answers;
DROP TABLE IF EXISTS learn.activity_attempts;
DROP TABLE IF EXISTS learn.user_scene_progress;
DROP TABLE IF EXISTS learn.user_story_progress;
DROP TABLE IF EXISTS learn.activities;
DROP TABLE IF EXISTS learn.scenes;
DROP TABLE IF EXISTS learn.stories;
DROP TABLE IF EXISTS learn.cities;
