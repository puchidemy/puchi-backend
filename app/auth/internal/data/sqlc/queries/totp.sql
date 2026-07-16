-- name: UpsertTOTPSecret :one
INSERT INTO auth.totp_secrets (user_id, encrypted_secret, encrypted_codes, is_enabled)
VALUES ($1, $2, $3, $4) ON CONFLICT (user_id) DO UPDATE
SET encrypted_secret = $2, encrypted_codes = $3, is_enabled = $4, updated_at = now()
RETURNING *;

-- name: GetTOTPSecret :one
SELECT * FROM auth.totp_secrets WHERE user_id = $1;

-- name: EnableTOTP :exec
UPDATE auth.totp_secrets SET is_enabled = true, verified_at = now(), updated_at = now() WHERE user_id = $1;

-- name: DisableTOTP :exec
UPDATE auth.totp_secrets SET is_enabled = false, updated_at = now() WHERE user_id = $1;
