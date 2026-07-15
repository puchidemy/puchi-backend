-- name: ListAchievementDefs :many
SELECT * FROM core.achievements_def ORDER BY id;

-- name: GetUserAchievement :one
SELECT * FROM core.user_achievements WHERE user_id = $1 AND achievement_id = $2;

-- name: ListUserAchievements :many
SELECT * FROM core.user_achievements WHERE user_id = $1 ORDER BY achievement_id;

-- name: UpsertUserAchievement :one
INSERT INTO core.user_achievements (user_id, achievement_id, progress, unlocked, unlocked_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (user_id, achievement_id)
DO UPDATE SET progress = EXCLUDED.progress, unlocked = EXCLUDED.unlocked, unlocked_at = EXCLUDED.unlocked_at
RETURNING *;
