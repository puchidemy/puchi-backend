-- name: GetUserStats :one
SELECT * FROM core.user_stats WHERE user_id = $1;

-- name: CreateUserStats :exec
INSERT INTO core.user_stats (user_id) VALUES ($1);

-- name: UpdateUserStats :one
UPDATE core.user_stats
SET current_xp = $2, total_xp = $3, level = $4, current_streak = $5,
    total_lessons = $6, completed_lessons = $7,
    total_minutes = $8, accuracy = $9, words_learned = $10, updated_at = now()
WHERE user_id = $1
RETURNING *;
