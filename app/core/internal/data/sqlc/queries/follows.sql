-- name: Follow :exec
INSERT INTO user_social.follows (follower_id, following_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: Unfollow :exec
DELETE FROM user_social.follows
WHERE follower_id = $1 AND following_id = $2;

-- name: IsFollowing :one
SELECT EXISTS(
    SELECT 1 FROM user_social.follows
    WHERE follower_id = $1 AND following_id = $2
) AS following;

-- name: ListFollowers :many
SELECT u.id, u.username, u.first_name, u.last_name, u.avatar_key,
       COALESCE(us.level, 1) AS level,
       COALESCE(us.current_streak, 0) AS streak,
       EXISTS(
           SELECT 1 FROM user_social.follows f2
           WHERE f2.follower_id = $2 AND f2.following_id = u.id
       ) AS is_following
FROM core.users u
JOIN user_social.follows f ON f.follower_id = u.id
LEFT JOIN core.user_stats us ON us.user_id = u.id
WHERE f.following_id = $1
ORDER BY f.created_at DESC;

-- name: ListFollowing :many
SELECT u.id, u.username, u.first_name, u.last_name, u.avatar_key,
       COALESCE(us.level, 1) AS level,
       COALESCE(us.current_streak, 0) AS streak,
       true AS is_following
FROM core.users u
JOIN user_social.follows f ON f.following_id = u.id
LEFT JOIN core.user_stats us ON us.user_id = u.id
WHERE f.follower_id = $1
ORDER BY f.created_at DESC;

-- name: SearchUsers :many
SELECT u.id, u.username, u.first_name, u.last_name, u.avatar_key,
       COALESCE(us.level, 1) AS level,
       COALESCE(us.current_streak, 0) AS streak,
       EXISTS(
           SELECT 1 FROM user_social.follows f2
           WHERE f2.follower_id = $2 AND f2.following_id = u.id
       ) AS is_following
FROM core.users u
LEFT JOIN core.user_stats us ON us.user_id = u.id
WHERE u.username ILIKE '%' || $1::text || '%'
   OR u.first_name ILIKE '%' || $1::text || '%'
   OR u.last_name ILIKE '%' || $1::text || '%'
ORDER BY u.username ASC
LIMIT $3;

-- name: GetFollowCounts :one
WITH counts AS (
    SELECT
        (SELECT COUNT(*) FROM user_social.follows f1 WHERE f1.follower_id = $1) AS following_count,
        (SELECT COUNT(*) FROM user_social.follows f2 WHERE f2.following_id = $1) AS followers_count
)
SELECT following_count::bigint, followers_count::bigint FROM counts;
