-- +goose Down
DELETE FROM auth.permissions;
DELETE FROM auth.roles;
DROP TABLE IF EXISTS auth.user_roles CASCADE;
DROP TABLE IF EXISTS auth.role_permissions CASCADE;
DROP TABLE IF EXISTS auth.permission_version CASCADE;
DROP TABLE IF EXISTS auth.permissions CASCADE;
DROP TABLE IF EXISTS auth.roles CASCADE;
