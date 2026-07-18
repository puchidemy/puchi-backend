package data

import (
	"context"

	"github.com/google/wire"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/puchidemy/puchi-backend/app/core/internal/biz"
	"github.com/puchidemy/puchi-backend/app/core/internal/conf"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData,
	NewUserRepo,
	NewStatsRepo,
	NewStatsTxManager,
	NewAchievementRepo,
	NewSocialRepo,
	NewSettingsRepo,
	wire.FieldsOf(new(*Data), "Pool"),
	wire.Bind(new(biz.UserRepoInterface), new(*UserRepo)),
	wire.Bind(new(biz.StatsRepoInterface), new(*StatsRepo)),
	wire.Bind(new(biz.StatsTxManagerInterface), new(*StatsTxManager)),
	wire.Bind(new(biz.AchievementRepoInterface), new(*AchievementRepo)),
	wire.Bind(new(biz.SocialRepoInterface), new(*SocialRepo)),
	wire.Bind(new(biz.SettingsRepo), new(*SettingsRepo)),
)

// Data .
type Data struct {
	Pool *pgxpool.Pool
}

// NewData .
func NewData(cfg *conf.Data) (*Data, func(), error) {
	pool, err := pgxpool.New(context.Background(), cfg.Database.Source)
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() {
		pool.Close()
	}
	return &Data{Pool: pool}, cleanup, nil
}
