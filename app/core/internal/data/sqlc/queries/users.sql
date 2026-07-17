-- name: CreateUser :one
INSERT INTO core.users (id, username, email, first_name, last_name)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUser :one
SELECT * FROM core.users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM core.users WHERE email = $1;

-- name: UpdateUser :one
UPDATE core.users
SET first_name = $2, last_name = $3, username = $4, bio = $5, avatar_key = $6, age_range = $7, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UpdateAvatarKey :one
UPDATE core.users
SET avatar_key = $2, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UsernameExists :one
SELECT EXISTS(SELECT 1 FROM core.users WHERE username = $1) AS exists;

-- name: DeleteUser :exec
DELETE FROM core.users WHERE id = $1;

-- name: GetUserByUsername :one
SELECT * FROM core.users WHERE username = $1;

-- name: UpdateOnboardingInfo :one
UPDATE core.users
SET first_name = $2, last_name = $3, age_range = $4, username = $5,
    onboarding_completed = true, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UpsertUserOnboarding :one
INSERT INTO core.user_onboarding (user_id, how_heard, why_learn, level)
VALUES ($1, $2, $3, $4)
ON CONFLICT (user_id) 
DO UPDATE SET how_heard = EXCLUDED.how_heard, 
              why_learn = EXCLUDED.why_learn, 
              level = EXCLUDED.level,
              updated_at = now()
RETURNING *;
