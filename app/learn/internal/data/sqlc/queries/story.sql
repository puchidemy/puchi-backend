-- name: ListCities :many
SELECT * FROM learn.cities ORDER BY position;

-- name: GetCityBySlug :one
SELECT * FROM learn.cities WHERE slug = $1;

-- name: GetCityByID :one
SELECT * FROM learn.cities WHERE id = $1;

-- name: CountPublishedStoriesByCity :one
SELECT COUNT(*)::int AS count
FROM learn.stories
WHERE city_id = $1 AND status = 'published';

-- name: CountCompletedStoriesByOwnerCity :one
SELECT COUNT(*)::int AS count
FROM learn.user_story_progress usp
JOIN learn.stories s ON s.id = usp.story_id
WHERE usp.owner_type = $1
  AND usp.owner_id = $2
  AND s.city_id = $3
  AND usp.status = 'completed';

-- name: ListPublishedStoriesByCity :many
SELECT * FROM learn.stories
WHERE city_id = $1 AND status = 'published'
ORDER BY position;

-- name: GetStoryByID :one
SELECT * FROM learn.stories WHERE id = $1;

-- name: ListScenesByStoryID :many
SELECT * FROM learn.scenes WHERE story_id = $1 ORDER BY position;

-- name: GetSceneByID :one
SELECT * FROM learn.scenes WHERE id = $1;

-- name: ListActivitiesBySceneID :many
SELECT * FROM learn.activities WHERE scene_id = $1 ORDER BY position;

-- name: GetActivityByID :one
SELECT * FROM learn.activities WHERE id = $1;

-- name: ListActivitiesByStoryID :many
SELECT a.*
FROM learn.activities a
JOIN learn.scenes sc ON sc.id = a.scene_id
WHERE sc.story_id = $1
ORDER BY sc.position, a.position;
