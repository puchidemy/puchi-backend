-- name: CreateMagicLink :one
INSERT INTO auth.magic_links (email, user_id, token, redirect_to, expires_at)
VALUES ($1, $2, $3, $4, now() + interval '15 minutes') RETURNING *;

-- name: GetMagicLinkByToken :one
SELECT * FROM auth.magic_links WHERE token = $1 AND used_at IS NULL AND expires_at > now();

-- name: MarkMagicLinkUsed :exec
UPDATE auth.magic_links SET used_at = now() WHERE id = $1;
