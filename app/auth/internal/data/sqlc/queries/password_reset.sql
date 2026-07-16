-- name: CreatePasswordResetToken :one
INSERT INTO auth.password_reset_tokens (user_id, token, expires_at)
VALUES ($1, $2, now() + interval '15 minutes') RETURNING *;

-- name: GetPasswordResetToken :one
SELECT * FROM auth.password_reset_tokens WHERE token = $1 AND used_at IS NULL AND expires_at > now();

-- name: MarkPasswordResetTokenUsed :exec
UPDATE auth.password_reset_tokens SET used_at = now() WHERE id = $1 AND used_at IS NULL;
