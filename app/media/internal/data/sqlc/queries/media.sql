-- name: CreateMediaObject :one
INSERT INTO media.objects (user_id, object_key, bucket, content_type, category, size_bytes, status)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetMediaObject :one
SELECT * FROM media.objects WHERE id = $1;

-- name: GetMediaObjectByKey :one
SELECT * FROM media.objects WHERE object_key = $1;

-- name: ListUserMedia :many
SELECT * FROM media.objects WHERE user_id = $1 ORDER BY created_at DESC;

-- name: UpdateMediaStatus :one
UPDATE media.objects SET status = $2 WHERE id = $1 RETURNING *;

-- name: DeleteMediaObject :exec
DELETE FROM media.objects WHERE id = $1;

-- name: DeleteMediaObjectByKey :exec
DELETE FROM media.objects WHERE object_key = $1;
