-- name: GetRoleByName :one
SELECT * FROM auth.roles WHERE name = $1;

-- name: ListRoles :many
SELECT * FROM auth.roles ORDER BY name;

-- name: ListPermissions :many
SELECT * FROM auth.permissions ORDER BY resource, action;

-- name: GetRolePermissions :many
SELECT p.* FROM auth.permissions p
INNER JOIN auth.role_permissions rp ON p.id = rp.permission_id
WHERE rp.role_id = $1;

-- name: GetUserRoles :many
SELECT r.* FROM auth.roles r
INNER JOIN auth.user_roles ur ON r.id = ur.role_id
WHERE ur.user_id = $1;

-- name: AssignRoleToUser :exec
INSERT INTO auth.user_roles (user_id, role_id, granted_by)
VALUES ($1, $2, $3) ON CONFLICT DO NOTHING;

-- name: RemoveRoleFromUser :exec
DELETE FROM auth.user_roles WHERE user_id = $1 AND role_id = $2;

-- name: IncrementPermissionVersion :exec
UPDATE auth.permission_version SET version = version + 1 WHERE id = 1;

-- name: GetPermissionVersion :one
SELECT version FROM auth.permission_version WHERE id = 1;
