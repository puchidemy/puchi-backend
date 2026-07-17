-- name: GetDailyActivity :one
SELECT * FROM core.daily_activities WHERE user_id = $1 AND activity_date = $2;

-- name: UpsertDailyActivity :one
INSERT INTO core.daily_activities (user_id, activity_date, lessons_completed, xp_earned)
VALUES ($1, $2, 1, $3)
ON CONFLICT (user_id, activity_date)
DO UPDATE SET
    lessons_completed = core.daily_activities.lessons_completed + 1,
    xp_earned = core.daily_activities.xp_earned + EXCLUDED.xp_earned
RETURNING *;

-- name: GetLatestActivityDateBefore :one
SELECT activity_date FROM core.daily_activities
WHERE user_id = $1 AND activity_date < $2 AND lessons_completed > 0
ORDER BY activity_date DESC
LIMIT 1;
