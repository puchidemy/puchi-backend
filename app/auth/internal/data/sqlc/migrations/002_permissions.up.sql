-- +goose Up

-- Role-based access control (RBAC)

-- Roles
CREATE TABLE auth.roles (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL UNIQUE,
    description TEXT,
    is_system   BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Permissions
CREATE TABLE auth.permissions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL UNIQUE,
    resource    TEXT NOT NULL,
    action      TEXT NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Role → Permission mapping
CREATE TABLE auth.role_permissions (
    role_id       UUID NOT NULL REFERENCES auth.roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES auth.permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- User → Role assignment
CREATE TABLE auth.user_roles (
    user_id    UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    role_id    UUID NOT NULL REFERENCES auth.roles(id) ON DELETE CASCADE,
    granted_by UUID REFERENCES auth.users(id),
    granted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, role_id)
);

-- Permission version singleton (for cache invalidation)
CREATE TABLE auth.permission_version (
    id      INTEGER PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    version BIGINT NOT NULL DEFAULT 1
);
INSERT INTO auth.permission_version (id, version) VALUES (1, 1);

-- Seed default roles
INSERT INTO auth.roles (name, description, is_system) VALUES
    ('admin', 'Full system access', true),
    ('teacher', 'Can create and manage content', true),
    ('student', 'Default user role', true),
    ('user', 'Basic authenticated user', true);

-- Seed default permissions
INSERT INTO auth.permissions (name, resource, action, description) VALUES
    ('lesson:read', 'lesson', 'read', 'View lessons'),
    ('lesson:write', 'lesson', 'write', 'Create and edit lessons'),
    ('lesson:delete', 'lesson', 'delete', 'Delete lessons'),
    ('profile:read', 'profile', 'read', 'View profiles'),
    ('profile:write', 'profile', 'write', 'Edit own profile'),
    ('admin:users', 'admin', 'manage', 'Manage users'),
    ('admin:roles', 'admin', 'manage', 'Manage roles and permissions'),
    ('content:manage', 'content', 'manage', 'Manage all content');
