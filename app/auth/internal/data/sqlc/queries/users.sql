-- name: CreateUser :one
INSERT INTO auth.users (email, email_normalized, password_hash, display_name)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM auth.users WHERE email_normalized = $1;

-- name: GetUserByID :one
SELECT * FROM auth.users WHERE id = $1;

-- name: UpdateUserLastLogin :exec
UPDATE auth.users SET last_login_at = now(), updated_at = now() WHERE id = $1;

-- name: UpdateUserPassword :exec
UPDATE auth.users SET password_hash = $2, updated_at = now() WHERE id = $1;

-- name: SetEmailVerified :exec
UPDATE auth.users SET email_verified = true, updated_at = now() WHERE id = $1;

-- name: SetUserActive :exec
UPDATE auth.users SET is_active = $2, updated_at = now() WHERE id = $1;
