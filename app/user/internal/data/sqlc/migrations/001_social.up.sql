-- +goose Up
CREATE SCHEMA IF NOT EXISTS user_social;

CREATE TABLE user_social.follows (
    follower_id  TEXT NOT NULL,
    following_id TEXT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (follower_id, following_id),
    CHECK (follower_id != following_id)
);

CREATE INDEX idx_follows_follower ON user_social.follows(follower_id);
CREATE INDEX idx_follows_following ON user_social.follows(following_id);

CREATE TABLE user_social.weekly_leaderboard (
    id         uuid PRIMARY KEY DEFAULT uuidv7(),
    user_id    TEXT NOT NULL,
    week_start DATE NOT NULL,
    weekly_xp  INT NOT NULL DEFAULT 0,
    rank       INT,
    UNIQUE (user_id, week_start)
);

CREATE INDEX idx_leaderboard_week ON user_social.weekly_leaderboard(week_start, rank);
