-- name: CreateAttempt :one
INSERT INTO learn.attempts (owner_type, owner_id, lesson_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetAttemptByID :one
SELECT * FROM learn.attempts WHERE id = $1;

-- name: InsertAttemptAnswer :one
INSERT INTO learn.attempt_answers (attempt_id, exercise_id, payload, correct)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: CompleteAttempt :exec
UPDATE learn.attempts
SET status = 'completed', completed_at = now(), session_xp = $2
WHERE id = $1;

-- name: ReassignGuestAttempts :exec
UPDATE learn.attempts SET owner_type = 'user', owner_id = $2
WHERE owner_type = 'guest' AND owner_id = $1;
