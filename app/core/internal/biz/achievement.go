package biz

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/puchidemy/puchi-backend/app/core/internal/data/sqlc/gen"
)

// AchievementRepoInterface defines the repo contract for achievements.
type AchievementRepoInterface interface {
	ListAchievementDefs(ctx context.Context) ([]gen.CoreAchievementsDef, error)
	ListUserAchievements(ctx context.Context, userID string) ([]gen.CoreUserAchievement, error)
}

// AchievementUsecase handles achievement operations.
type AchievementUsecase struct {
	repo AchievementRepoInterface
}

// NewAchievementUsecase creates a new AchievementUsecase.
func NewAchievementUsecase(repo AchievementRepoInterface) *AchievementUsecase {
	return &AchievementUsecase{repo: repo}
}

// AchievementItem represents an achievement with the user's progress.
type AchievementItem struct {
	ID            string
	Title         string
	Description   string
	Icon          string
	Color         string
	Progress      int32
	ProgressLabel string
	Unlocked      bool
	UnlockedAt    pgtype.Timestamptz
}

// ListAchievements returns all achievement definitions with the user's progress.
func (uc *AchievementUsecase) ListAchievements(ctx context.Context, userID string) ([]AchievementItem, error) {
	defs, err := uc.repo.ListAchievementDefs(ctx)
	if err != nil {
		return nil, fmt.Errorf("list achievement defs: %w", err)
	}

	userAchievements, err := uc.repo.ListUserAchievements(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list user achievements: %w", err)
	}

	uaMap := make(map[string]gen.CoreUserAchievement, len(userAchievements))
	for _, ua := range userAchievements {
		uaMap[ua.AchievementID] = ua
	}

	items := make([]AchievementItem, 0, len(defs))
	for _, def := range defs {
		item := AchievementItem{
			ID:          def.ID,
			Title:       def.Title,
			Description: def.Description,
			Icon:        def.Icon,
			Color:       def.Color,
		}
		if ua, ok := uaMap[def.ID]; ok {
			item.Progress = ua.Progress
			item.Unlocked = ua.Unlocked
			item.UnlockedAt = ua.UnlockedAt
		}
		item.ProgressLabel = fmt.Sprintf("%d/%d", item.Progress, def.RequirementValue)
		items = append(items, item)
	}
	return items, nil
}
