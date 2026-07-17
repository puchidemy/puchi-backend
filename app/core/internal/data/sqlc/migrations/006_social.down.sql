-- +goose Down
DROP TABLE IF EXISTS user_social.weekly_leaderboard;
DROP TABLE IF EXISTS user_social.follows;
DROP SCHEMA IF EXISTS user_social;
