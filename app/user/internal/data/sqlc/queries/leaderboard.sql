-- name: GetWeeklyLeaderboard :many
SELECT wl.user_id, u.username, u.avatar_key,
       COALESCE(us.level, 1) AS level, wl.weekly_xp,
       ROW_NUMBER() OVER (ORDER BY wl.weekly_xp DESC)::int AS rank
FROM user_social.weekly_leaderboard wl
JOIN core.users u ON u.id = wl.user_id
LEFT JOIN core.user_stats us ON us.user_id = wl.user_id
WHERE wl.week_start = date_trunc('week', current_date)::date
ORDER BY wl.weekly_xp DESC
LIMIT $1;

-- name: UpsertWeeklyXP :exec
INSERT INTO user_social.weekly_leaderboard (user_id, week_start, weekly_xp)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, week_start)
DO UPDATE SET weekly_xp = user_social.weekly_leaderboard.weekly_xp + $3;

-- name: GetUserWeeklyXP :one
SELECT weekly_xp
FROM user_social.weekly_leaderboard
WHERE user_id = $1 AND week_start = $2;
