-- name: CreateGuest :exec
INSERT INTO learn.guests (id) VALUES ($1);

-- name: GetGuestByID :one
SELECT * FROM learn.guests WHERE id = $1;

-- name: GetGuestByIDForUpdate :one
SELECT * FROM learn.guests WHERE id = $1 FOR UPDATE;

-- name: ClaimGuest :execrows
UPDATE learn.guests
SET claimed_at = now(), claimed_user_id = $2
WHERE id = $1 AND claimed_at IS NULL;

-- name: TouchGuest :exec
UPDATE learn.guests SET last_seen_at = now() WHERE id = $1;
