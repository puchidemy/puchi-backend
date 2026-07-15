package data

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/puchidemy/puchi-backend/app/core/internal/data/sqlc/gen"
)

// AchievementRepo wraps sqlc-generated queries for core.achievements* tables.
type AchievementRepo struct {
	q *gen.Queries
}

// NewAchievementRepo creates a new AchievementRepo.
func NewAchievementRepo(pool *pgxpool.Pool) *AchievementRepo {
	return &AchievementRepo{q: gen.New(pool)}
}

// ListAchievementDefs retrieves all achievement definitions.
func (r *AchievementRepo) ListAchievementDefs(ctx context.Context) ([]gen.CoreAchievementsDef, error) {
	return r.q.ListAchievementDefs(ctx)
}

// ListUserAchievements retrieves all achievement progress for a user.
func (r *AchievementRepo) ListUserAchievements(ctx context.Context, userID string) ([]gen.CoreUserAchievement, error) {
	return r.q.ListUserAchievements(ctx, userID)
}
