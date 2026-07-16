package biz

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// RBACUsecase handles role-based access control operations.
type RBACUsecase struct {
	roleRepo       RoleRepo
	permissionRepo PermissionRepo
	userRepo       UserRepo
}

// NewRBACUsecase creates a new RBACUsecase.
func NewRBACUsecase(roleRepo RoleRepo, permissionRepo PermissionRepo, userRepo UserRepo) *RBACUsecase {
	return &RBACUsecase{
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
		userRepo:       userRepo,
	}
}

// AssignRole assigns a role to a user by role name.
func (uc *RBACUsecase) AssignRole(ctx context.Context, userID uuid.UUID, roleName string, grantedBy uuid.UUID) error {
	role, err := uc.roleRepo.GetByName(ctx, roleName)
	if err != nil {
		return fmt.Errorf("get role by name: %w", err)
	}

	if err := uc.roleRepo.AssignToUser(ctx, userID, role.ID, grantedBy); err != nil {
		return fmt.Errorf("assign role to user: %w", err)
	}

	if err := uc.permissionRepo.IncrementVersion(ctx); err != nil {
		return fmt.Errorf("increment permission version: %w", err)
	}

	return nil
}

// RemoveRole removes a role assignment from a user by role name.
func (uc *RBACUsecase) RemoveRole(ctx context.Context, userID uuid.UUID, roleName string) error {
	role, err := uc.roleRepo.GetByName(ctx, roleName)
	if err != nil {
		return fmt.Errorf("get role by name: %w", err)
	}

	if err := uc.roleRepo.RemoveFromUser(ctx, userID, role.ID); err != nil {
		return fmt.Errorf("remove role from user: %w", err)
	}

	if err := uc.permissionRepo.IncrementVersion(ctx); err != nil {
		return fmt.Errorf("increment permission version: %w", err)
	}

	return nil
}

// GetUserRoles returns the role names assigned to a user.
func (uc *RBACUsecase) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]string, error) {
	roles, err := uc.roleRepo.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user roles: %w", err)
	}

	names := make([]string, len(roles))
	for i, r := range roles {
		names[i] = r.Name
	}
	return names, nil
}

// GetUserPermissions resolves role → permissions for a user and returns permission names.
func (uc *RBACUsecase) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	roles, err := uc.roleRepo.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user roles: %w", err)
	}

	// Use a set to deduplicate permissions across roles
	permSet := make(map[string]struct{})
	for _, role := range roles {
		perms, err := uc.roleRepo.GetRolePermissions(ctx, role.ID)
		if err != nil {
			return nil, fmt.Errorf("get role permissions: %w", err)
		}
		for _, p := range perms {
			permSet[p.Name] = struct{}{}
		}
	}

	names := make([]string, 0, len(permSet))
	for name := range permSet {
		names = append(names, name)
	}
	return names, nil
}

// GetPermissionVersion returns the current permission version for cache invalidation.
func (uc *RBACUsecase) GetPermissionVersion(ctx context.Context) (int64, error) {
	return uc.permissionRepo.GetVersion(ctx)
}

// ListRoles returns all roles.
func (uc *RBACUsecase) ListRoles(ctx context.Context) ([]*Role, error) {
	return uc.roleRepo.List(ctx)
}

// ListPermissions returns all permissions.
func (uc *RBACUsecase) ListPermissions(ctx context.Context) ([]*Permission, error) {
	return uc.permissionRepo.List(ctx)
}
