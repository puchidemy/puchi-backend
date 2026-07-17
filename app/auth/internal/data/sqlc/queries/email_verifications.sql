-- name: CreateEmailVerification :one
INSERT INTO auth.email_verifications (user_id, token, email, expires_at)
VALUES ($1, $2, $3, now() + interval '15 minutes') RETURNING *;

-- name: GetEmailVerificationByUserAndHash :one
SELECT * FROM auth.email_verifications WHERE user_id = $1 AND token = $2 AND used_at IS NULL AND expires_at > now();

-- name: MarkEmailVerificationUsed :exec
UPDATE auth.email_verifications SET used_at = now() WHERE id = $1 AND used_at IS NULL;
