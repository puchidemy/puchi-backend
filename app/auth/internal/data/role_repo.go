package data

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/puchidemy/puchi-backend/app/auth/internal/biz"
	"github.com/puchidemy/puchi-backend/app/auth/internal/data/sqlc/gen"
)

// RoleRepo wraps sqlc-generated queries for auth.roles.
type RoleRepo struct {
	q *gen.Queries
}

// NewRoleRepo creates a new RoleRepo.
func NewRoleRepo(d *Data) *RoleRepo {
	return &RoleRepo{q: gen.New(d.Pool)}
}

// GetByName retrieves a role by its name.
func (r *RoleRepo) GetByName(ctx context.Context, name string) (*biz.Role, error) {
	row, err := r.q.GetRoleByName(ctx, name)
	if err != nil {
		return nil, err
	}
	return toRole(row), nil
}

// List returns all roles ordered by name.
func (r *RoleRepo) List(ctx context.Context) ([]*biz.Role, error) {
	rows, err := r.q.ListRoles(ctx)
	if err != nil {
		return nil, err
	}
	roles := make([]*biz.Role, len(rows))
	for i, row := range rows {
		roles[i] = toRole(row)
	}
	return roles, nil
}

// GetRolePermissions returns all permissions assigned to a role.
func (r *RoleRepo) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*biz.Permission, error) {
	rows, err := r.q.GetRolePermissions(ctx, roleID.String())
	if err != nil {
		return nil, err
	}
	perms := make([]*biz.Permission, len(rows))
	for i, row := range rows {
		perms[i] = toPermission(row)
	}
	return perms, nil
}

// GetUserRoles returns all roles assigned to a user.
func (r *RoleRepo) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*biz.Role, error) {
	rows, err := r.q.GetUserRoles(ctx, userID.String())
	if err != nil {
		return nil, err
	}
	roles := make([]*biz.Role, len(rows))
	for i, row := range rows {
		roles[i] = toRole(row)
	}
	return roles, nil
}

// AssignToUser assigns a role to a user.
func (r *RoleRepo) AssignToUser(ctx context.Context, userID uuid.UUID, roleID uuid.UUID, grantedBy uuid.UUID) error {
	return r.q.AssignRoleToUser(ctx, gen.AssignRoleToUserParams{
		UserID:    userID.String(),
		RoleID:    roleID.String(),
		GrantedBy: pgtype.UUID{Bytes: grantedBy, Valid: grantedBy != uuid.Nil},
	})
}

// RemoveFromUser removes a role assignment from a user.
func (r *RoleRepo) RemoveFromUser(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) error {
	return r.q.RemoveRoleFromUser(ctx, gen.RemoveRoleFromUserParams{
		UserID: userID.String(),
		RoleID: roleID.String(),
	})
}

// toRole converts a gen.AuthRole to a biz.Role.
func toRole(row gen.AuthRole) *biz.Role {
	return &biz.Role{
		ID:          uuid.MustParse(row.ID),
		Name:        row.Name,
		Description: derefString(row.Description),
		IsSystem:    row.IsSystem,
		CreatedAt:   row.CreatedAt,
	}
}

// derefString safely dereferences a string pointer, returning empty string for nil.
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
