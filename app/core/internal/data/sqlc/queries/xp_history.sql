-- name: UpsertWeeklyXP :exec
INSERT INTO core.xp_history (user_id, week_start, xp_earned)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, week_start)
DO UPDATE SET xp_earned = core.xp_history.xp_earned + EXCLUDED.xp_earned;

-- name: ListWeeklyXPHistory :many
SELECT * FROM core.xp_history
WHERE user_id = $1 AND week_start >= $2
ORDER BY week_start ASC;
