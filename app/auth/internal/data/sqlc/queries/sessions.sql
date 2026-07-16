-- name: CreateSession :one
INSERT INTO auth.sessions (user_id, refresh_token_hash, token_family, child_number,
    ip_address, user_agent, device_name, device_type, os, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: GetSessionByHash :one
SELECT * FROM auth.sessions WHERE refresh_token_hash = $1;

-- name: RevokeSession :exec
UPDATE auth.sessions SET is_active = false, revoked_at = now() WHERE id = $1;

-- name: RevokeFamily :exec
UPDATE auth.sessions SET is_active = false, revoked_at = now()
WHERE token_family = $1 AND is_active = true;

-- name: RevokeAllForUser :exec
UPDATE auth.sessions SET is_active = false, revoked_at = now()
WHERE user_id = $1 AND is_active = true;

-- name: ListSessionsByUser :many
SELECT * FROM auth.sessions WHERE user_id = $1 AND is_active = true
ORDER BY created_at DESC;

-- name: UpdateSessionLastUsed :exec
UPDATE auth.sessions SET last_used_at = now() WHERE id = $1;
