-- name: GetUserStats :one
SELECT * FROM core.user_stats WHERE user_id = $1;

-- name: CreateUserStats :exec
INSERT INTO core.user_stats (user_id) VALUES ($1)
ON CONFLICT (user_id) DO NOTHING;

-- name: UpdateUserStats :one
UPDATE core.user_stats
SET current_xp = $2, total_xp = $3, level = $4, current_streak = $5,
    longest_streak = $6,
    total_lessons = $7, completed_lessons = $8,
    total_minutes = $9, accuracy = $10, words_learned = $11, updated_at = now()
WHERE user_id = $1
RETURNING *;
