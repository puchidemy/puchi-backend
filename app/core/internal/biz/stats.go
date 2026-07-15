package biz

import (
	"context"
	"errors"
	"math"

	"github.com/puchidemy/puchi-backend/app/core/internal/data/sqlc/gen"
)

// StatsRepoInterface defines the repo contract for stats (dependency inversion).
type StatsRepoInterface interface {
	GetUserStats(ctx context.Context, userID string) (*gen.CoreUserStat, error)
	UpdateStats(ctx context.Context, arg gen.UpdateUserStatsParams) (*gen.CoreUserStat, error)
}

// StatsUsecase handles gamification stats operations.
type StatsUsecase struct {
	repo StatsRepoInterface
}

// NewStatsUsecase creates a new StatsUsecase.
func NewStatsUsecase(repo StatsRepoInterface) *StatsUsecase {
	return &StatsUsecase{repo: repo}
}

// GetStats returns the user's gamification stats.
func (uc *StatsUsecase) GetStats(ctx context.Context, userID string) (*gen.CoreUserStat, error) {
	stats, err := uc.repo.GetUserStats(ctx, userID)
	if err != nil {
		return nil, errors.New("stats not found")
	}
	return stats, nil
}

// GetXPToNextLevel calculates the XP required to reach the next level.
// Formula: ceil(level * 60 * 1.5)
func (uc *StatsUsecase) GetXPToNextLevel(level int32) int32 {
	return int32(math.Ceil(float64(level) * 60 * 1.5))
}
