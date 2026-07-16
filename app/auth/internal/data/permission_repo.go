package data

import (
	"context"

	"github.com/google/uuid"

	"github.com/puchidemy/puchi-backend/app/auth/internal/biz"
	"github.com/puchidemy/puchi-backend/app/auth/internal/data/sqlc/gen"
)

// PermissionRepo wraps sqlc-generated queries for auth.permissions.
type PermissionRepo struct {
	q *gen.Queries
}

// NewPermissionRepo creates a new PermissionRepo.
func NewPermissionRepo(d *Data) *PermissionRepo {
	return &PermissionRepo{q: gen.New(d.Pool)}
}

// List returns all permissions ordered by resource, action.
func (r *PermissionRepo) List(ctx context.Context) ([]*biz.Permission, error) {
	rows, err := r.q.ListPermissions(ctx)
	if err != nil {
		return nil, err
	}
	perms := make([]*biz.Permission, len(rows))
	for i, row := range rows {
		perms[i] = toPermission(row)
	}
	return perms, nil
}

// IncrementVersion increments the permission version for cache invalidation.
func (r *PermissionRepo) IncrementVersion(ctx context.Context) error {
	return r.q.IncrementPermissionVersion(ctx)
}

// GetVersion returns the current permission version.
func (r *PermissionRepo) GetVersion(ctx context.Context) (int64, error) {
	return r.q.GetPermissionVersion(ctx)
}

// toPermission converts a gen.AuthPermission to a biz.Permission.
func toPermission(row gen.AuthPermission) *biz.Permission {
	return &biz.Permission{
		ID:          uuid.MustParse(row.ID),
		Name:        row.Name,
		Resource:    row.Resource,
		Action:      row.Action,
		Description: derefString(row.Description),
		CreatedAt:   row.CreatedAt,
	}
}
