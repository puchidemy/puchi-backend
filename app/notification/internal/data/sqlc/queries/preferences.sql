-- name: GetPreferences :one
SELECT * FROM notification.preferences WHERE user_id = $1;

-- name: UpsertPreferences :one
INSERT INTO notification.preferences (
    user_id, push_enabled, email_enabled,
    streak_reminder, friend_activity, promotions,
    quiet_hours_start, quiet_hours_end
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (user_id) DO UPDATE SET
    push_enabled      = COALESCE($2, notification.preferences.push_enabled),
    email_enabled     = COALESCE($3, notification.preferences.email_enabled),
    streak_reminder   = COALESCE($4, notification.preferences.streak_reminder),
    friend_activity   = COALESCE($5, notification.preferences.friend_activity),
    promotions        = COALESCE($6, notification.preferences.promotions),
    quiet_hours_start = COALESCE($7, notification.preferences.quiet_hours_start),
    quiet_hours_end   = COALESCE($8, notification.preferences.quiet_hours_end)
RETURNING *;
