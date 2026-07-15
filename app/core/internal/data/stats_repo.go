package data

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/puchidemy/puchi-backend/app/core/internal/data/sqlc/gen"
)

// StatsRepo wraps sqlc-generated queries for core.user_stats.
type StatsRepo struct {
	q *gen.Queries
}

// NewStatsRepo creates a new StatsRepo.
func NewStatsRepo(pool *pgxpool.Pool) *StatsRepo {
	return &StatsRepo{q: gen.New(pool)}
}

// GetUserStats retrieves stats for a user.
func (r *StatsRepo) GetUserStats(ctx context.Context, userID string) (*gen.CoreUserStat, error) {
	row, err := r.q.GetUserStats(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// UpsertStats creates initial stats row for a user if not exists.
func (r *StatsRepo) UpsertStats(ctx context.Context, userID string) error {
	return r.q.CreateUserStats(ctx, userID)
}

// UpdateStats updates stats and returns the updated row.
func (r *StatsRepo) UpdateStats(ctx context.Context, arg gen.UpdateUserStatsParams) (*gen.CoreUserStat, error) {
	row, err := r.q.UpdateUserStats(ctx, arg)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// GetLevelThreshold returns the XP required for a given level.
func (r *StatsRepo) GetLevelThreshold(ctx context.Context, level int32) (int32, error) {
	return r.q.GetLevelThreshold(ctx, level)
}

// GetNextLevelThreshold returns the XP required for the next level.
func (r *StatsRepo) GetNextLevelThreshold(ctx context.Context, level int32) (int32, error) {
	return r.q.GetNextLevelThreshold(ctx, level)
}
