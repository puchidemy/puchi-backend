-- +goose Up

-- UUID v7 extension (native PG 18) — no extension needed, uuidv7() built-in

CREATE SCHEMA IF NOT EXISTS core;

-- Users: ID từ Supertokens (string UUID v4)
CREATE TABLE core.users (
    id             TEXT PRIMARY KEY,
    username       TEXT UNIQUE NOT NULL,
    first_name     TEXT NOT NULL DEFAULT '',
    last_name      TEXT NOT NULL DEFAULT '',
    email          TEXT UNIQUE NOT NULL,
    avatar_key     TEXT,
    bio            TEXT DEFAULT '',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    st_sign_up_at            TIMESTAMPTZ,
    st_third_party_provider  TEXT,
    st_third_party_user_id   TEXT
);

CREATE INDEX idx_users_username ON core.users(username);
CREATE INDEX idx_users_email ON core.users(email);

-- User stats: 1:1 with users
CREATE TABLE core.user_stats (
    user_id          TEXT PRIMARY KEY REFERENCES core.users(id) ON DELETE CASCADE,
    current_xp       INT NOT NULL DEFAULT 0,
    total_xp         INT NOT NULL DEFAULT 0,
    level            INT NOT NULL DEFAULT 1,
    current_streak   INT NOT NULL DEFAULT 0,
    longest_streak   INT NOT NULL DEFAULT 0,
    streak_freezes   INT NOT NULL DEFAULT 0,
    crowns           INT NOT NULL DEFAULT 0,
    gems             INT NOT NULL DEFAULT 0,
    total_lessons    INT NOT NULL DEFAULT 0,
    total_minutes    INT NOT NULL DEFAULT 0,
    accuracy         REAL NOT NULL DEFAULT 0,
    words_learned    INT NOT NULL DEFAULT 0,
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Daily activities with UUID v7 PK (time-ordered, B-tree friendly)
CREATE TABLE core.daily_activities (
    id                uuid PRIMARY KEY DEFAULT uuidv7(),
    user_id           TEXT NOT NULL REFERENCES core.users(id) ON DELETE CASCADE,
    activity_date     DATE NOT NULL,
    lessons_completed INT NOT NULL DEFAULT 0,
    xp_earned         INT NOT NULL DEFAULT 0,
    minutes_spent     INT NOT NULL DEFAULT 0,
    UNIQUE (user_id, activity_date)
);

CREATE INDEX idx_daily_activities_user_date ON core.daily_activities(user_id, activity_date);

-- Weekly XP history with UUID v7 PK
CREATE TABLE core.xp_history (
    id         uuid PRIMARY KEY DEFAULT uuidv7(),
    user_id    TEXT NOT NULL REFERENCES core.users(id) ON DELETE CASCADE,
    week_start DATE NOT NULL,
    xp_earned  INT NOT NULL DEFAULT 0,
    UNIQUE (user_id, week_start)
);

CREATE INDEX idx_xp_history_user_week ON core.xp_history(user_id, week_start);

-- Achievement definitions
CREATE TABLE core.achievements_def (
    id               TEXT PRIMARY KEY,
    title            TEXT NOT NULL,
    description      TEXT NOT NULL,
    icon             TEXT NOT NULL,
    color            TEXT NOT NULL,
    category         TEXT NOT NULL,
    requirement_type TEXT NOT NULL,
    requirement_value INT NOT NULL
);

-- User achievement progress
CREATE TABLE core.user_achievements (
    user_id        TEXT NOT NULL REFERENCES core.users(id) ON DELETE CASCADE,
    achievement_id TEXT NOT NULL REFERENCES core.achievements_def(id),
    progress       INT NOT NULL DEFAULT 0,
    unlocked       BOOLEAN NOT NULL DEFAULT false,
    unlocked_at    TIMESTAMPTZ,
    PRIMARY KEY (user_id, achievement_id)
);

-- Level thresholds
CREATE TABLE core.level_thresholds (
    level       INT PRIMARY KEY,
    xp_required INT NOT NULL
);

-- Seed level data
INSERT INTO core.level_thresholds (level, xp_required) VALUES
(1, 0), (2, 60), (3, 120), (4, 200), (5, 300),
(6, 450), (7, 650), (8, 900), (9, 1200), (10, 1500);

-- +goose Down
DROP TABLE IF EXISTS core.level_thresholds;
DROP TABLE IF EXISTS core.user_achievements;
DROP TABLE IF EXISTS core.achievements_def;
DROP TABLE IF EXISTS core.xp_history;
DROP TABLE IF EXISTS core.daily_activities;
DROP TABLE IF EXISTS core.user_stats;
DROP TABLE IF EXISTS core.users;
DROP SCHEMA IF EXISTS core;
