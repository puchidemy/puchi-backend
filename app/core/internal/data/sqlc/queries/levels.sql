-- name: GetLevelThreshold :one
SELECT xp_required FROM core.level_thresholds WHERE level = $1;

-- name: GetNextLevelThreshold :one
SELECT xp_required FROM core.level_thresholds WHERE level = $1;
