-- name: UpsertWeeklyXP :exec
INSERT INTO core.xp_history (user_id, week_start, xp_earned)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, week_start)
DO UPDATE SET xp_earned = core.xp_history.xp_earned + EXCLUDED.xp_earned;
