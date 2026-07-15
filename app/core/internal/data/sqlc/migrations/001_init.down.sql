-- +goose Down
DROP TABLE IF EXISTS core.level_thresholds;
DROP TABLE IF EXISTS core.user_achievements;
DROP TABLE IF EXISTS core.achievements_def;
DROP TABLE IF EXISTS core.xp_history;
DROP TABLE IF EXISTS core.daily_activities;
DROP TABLE IF EXISTS core.user_stats;
DROP TABLE IF EXISTS core.users;
DROP SCHEMA IF EXISTS core;
