-- name: GetSocialConnection :one
SELECT * FROM auth.social_connections WHERE provider = $1 AND provider_user_id = $2;

-- name: ListSocialConnectionsByUser :many
SELECT * FROM auth.social_connections WHERE user_id = $1;

-- name: CreateSocialConnection :one
INSERT INTO auth.social_connections (user_id, provider, provider_user_id, provider_email, avatar_url)
VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: DeleteSocialConnection :exec
DELETE FROM auth.social_connections WHERE id = $1 AND user_id = $2;
