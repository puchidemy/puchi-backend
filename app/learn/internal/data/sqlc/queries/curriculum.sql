-- name: GetUnitByID :one
SELECT * FROM learn.units WHERE id = $1;

-- name: ListSkillsByUnitID :many
SELECT * FROM learn.skills WHERE unit_id = $1 ORDER BY position;

-- name: GetLessonByID :one
SELECT * FROM learn.lessons WHERE id = $1;

-- name: ListLessonsBySkillID :many
SELECT * FROM learn.lessons WHERE skill_id = $1 ORDER BY position;

-- name: ListExercisesByLessonID :many
SELECT * FROM learn.exercises WHERE lesson_id = $1 ORDER BY position;
