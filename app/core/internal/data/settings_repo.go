package data

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/puchidemy/puchi-backend/app/core/internal/data/sqlc/gen"
)

// SettingsRepo wraps sqlc-generated queries for core.user_settings.
type SettingsRepo struct {
	q *gen.Queries
}

// NewSettingsRepo creates a new SettingsRepo.
func NewSettingsRepo(pool *pgxpool.Pool) *SettingsRepo {
	return &SettingsRepo{q: gen.New(pool)}
}

// Get retrieves settings for a user.
func (r *SettingsRepo) Get(ctx context.Context, userID string) (*gen.CoreUserSetting, error) {
	row, err := r.q.GetUserSettings(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// Upsert inserts or updates settings for a user.
func (r *SettingsRepo) Upsert(ctx context.Context, params gen.UpsertUserSettingsParams) (*gen.CoreUserSetting, error) {
	row, err := r.q.UpsertUserSettings(ctx, params)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// EnsureDefaults returns existing settings, or inserts a defaults row if missing.
func (r *SettingsRepo) EnsureDefaults(ctx context.Context, userID string) (*gen.CoreUserSetting, error) {
	row, err := r.q.GetUserSettings(ctx, userID)
	if err == nil {
		return &row, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	row, err = r.q.CreateUserSettingsDefaults(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &row, nil
}
