-- +goose Up
CREATE SCHEMA IF NOT EXISTS notification;

CREATE TABLE notification.preferences (
    user_id           TEXT PRIMARY KEY,
    push_enabled      BOOLEAN NOT NULL DEFAULT true,
    email_enabled     BOOLEAN NOT NULL DEFAULT false,
    streak_reminder   BOOLEAN NOT NULL DEFAULT true,
    friend_activity   BOOLEAN NOT NULL DEFAULT true,
    promotions        BOOLEAN NOT NULL DEFAULT false,
    quiet_hours_start TIME,
    quiet_hours_end   TIME
);

CREATE TABLE notification.templates (
    id       TEXT PRIMARY KEY,
    title    TEXT NOT NULL,
    body     TEXT NOT NULL,
    category TEXT NOT NULL
);
