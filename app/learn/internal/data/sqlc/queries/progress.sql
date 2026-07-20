-- name: GetLessonProgress :one
SELECT * FROM learn.user_lesson_progress
WHERE owner_type = $1 AND owner_id = $2 AND lesson_id = $3;

-- name: UpsertLessonProgress :exec
INSERT INTO learn.user_lesson_progress (owner_type, owner_id, lesson_id, status, xp_earned)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (owner_type, owner_id, lesson_id) DO UPDATE SET
  status = EXCLUDED.status,
  xp_earned = GREATEST(learn.user_lesson_progress.xp_earned, EXCLUDED.xp_earned),
  updated_at = now();

-- name: ListLessonProgressByOwner :many
SELECT * FROM learn.user_lesson_progress
WHERE owner_type = $1 AND owner_id = $2;

-- name: ListUnitProgressByOwner :many
SELECT * FROM learn.user_unit_progress
WHERE owner_type = $1 AND owner_id = $2;

-- name: DeleteGuestLessonProgress :exec
DELETE FROM learn.user_lesson_progress
WHERE owner_type = 'guest' AND owner_id = $1 AND lesson_id = $2;

-- name: DeleteGuestUnitProgress :exec
DELETE FROM learn.user_unit_progress
WHERE owner_type = 'guest' AND owner_id = $1 AND unit_id = $2;

-- name: GetUnitProgress :one
SELECT * FROM learn.user_unit_progress
WHERE owner_type = $1 AND owner_id = $2 AND unit_id = $3;

-- name: UpsertUnitProgress :exec
INSERT INTO learn.user_unit_progress (owner_type, owner_id, unit_id, status)
VALUES ($1, $2, $3, $4)
ON CONFLICT (owner_type, owner_id, unit_id) DO UPDATE SET
  status = EXCLUDED.status,
  updated_at = now();

-- name: ReassignGuestLessonProgress :exec
UPDATE learn.user_lesson_progress SET owner_type = 'user', owner_id = $2
WHERE owner_type = 'guest' AND owner_id = $1;

-- name: ReassignGuestUnitProgress :exec
UPDATE learn.user_unit_progress SET owner_type = 'user', owner_id = $2
WHERE owner_type = 'guest' AND owner_id = $1;

-- name: GetStoryProgress :one
SELECT * FROM learn.user_story_progress
WHERE owner_type = $1 AND owner_id = $2 AND story_id = $3;

-- name: UpsertStoryProgress :exec
INSERT INTO learn.user_story_progress (owner_type, owner_id, story_id, status, xp_earned)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (owner_type, owner_id, story_id) DO UPDATE SET
  status = EXCLUDED.status,
  xp_earned = GREATEST(learn.user_story_progress.xp_earned, EXCLUDED.xp_earned),
  updated_at = now();

-- name: ListStoryProgressByOwner :many
SELECT * FROM learn.user_story_progress
WHERE owner_type = $1 AND owner_id = $2;

-- name: DeleteGuestStoryProgress :exec
DELETE FROM learn.user_story_progress
WHERE owner_type = 'guest' AND owner_id = $1 AND story_id = $2;

-- name: ReassignGuestStoryProgress :exec
UPDATE learn.user_story_progress SET owner_type = 'user', owner_id = $2
WHERE owner_type = 'guest' AND owner_id = $1;

-- name: GetSceneProgress :one
SELECT * FROM learn.user_scene_progress
WHERE owner_type = $1 AND owner_id = $2 AND scene_id = $3;

-- name: UpsertSceneProgress :exec
INSERT INTO learn.user_scene_progress (owner_type, owner_id, scene_id, status)
VALUES ($1, $2, $3, $4)
ON CONFLICT (owner_type, owner_id, scene_id) DO UPDATE SET
  status = EXCLUDED.status,
  updated_at = now();

-- name: ListSceneProgressByOwner :many
SELECT * FROM learn.user_scene_progress
WHERE owner_type = $1 AND owner_id = $2;

-- name: CountCompletedScenesByOwner :one
SELECT COUNT(*)::int AS count
FROM learn.user_scene_progress
WHERE owner_type = $1 AND owner_id = $2 AND status = 'completed';

-- name: DeleteGuestSceneProgress :exec
DELETE FROM learn.user_scene_progress
WHERE owner_type = 'guest' AND owner_id = $1 AND scene_id = $2;

-- name: ReassignGuestSceneProgress :exec
UPDATE learn.user_scene_progress SET owner_type = 'user', owner_id = $2
WHERE owner_type = 'guest' AND owner_id = $1;

-- name: ReassignGuestActivityAttempts :exec
UPDATE learn.activity_attempts SET owner_type = 'user', owner_id = $2
WHERE owner_type = 'guest' AND owner_id = $1;
