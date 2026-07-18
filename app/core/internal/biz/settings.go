package biz

import (
	"context"

	"github.com/puchidemy/puchi-backend/app/core/internal/data/sqlc/gen"
)

// SettingsRepo defines the repository contract for user settings.
type SettingsRepo interface {
	Get(ctx context.Context, userID string) (*gen.CoreUserSetting, error)
	Upsert(ctx context.Context, params gen.UpsertUserSettingsParams) (*gen.CoreUserSetting, error)
	EnsureDefaults(ctx context.Context, userID string) (*gen.CoreUserSetting, error)
}
