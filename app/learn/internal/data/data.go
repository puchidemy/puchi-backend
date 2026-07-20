package data

import (
	"context"

	"github.com/google/wire"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/puchidemy/puchi-backend/app/learn/internal/biz"
	"github.com/puchidemy/puchi-backend/app/learn/internal/conf"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData,
	NewNATSLessonEventPublisher,
	NewGuestRepo,
	NewProgressRepo,
	NewCurriculumRepo,
	NewStoryRepo,
	NewAttemptRepo,
	NewTransactionManager,
	wire.FieldsOf(new(*Data), "Pool"),
	wire.Bind(new(biz.LessonEventPublisher), new(*NATSLessonEventPublisher)),
	wire.Bind(new(biz.GuestRepoInterface), new(*GuestRepo)),
	wire.Bind(new(biz.ProgressRepoInterface), new(*ProgressRepo)),
	wire.Bind(new(biz.CurriculumRepoInterface), new(*CurriculumRepo)),
	wire.Bind(new(biz.StoryRepoInterface), new(*StoryRepo)),
	wire.Bind(new(biz.AttemptRepoInterface), new(*AttemptRepo)),
	wire.Bind(new(biz.TransactionManager), new(*TransactionManager)),
)

// Data wraps the database connection pool.
type Data struct {
	Pool *pgxpool.Pool
}

// NewData creates a new Data instance with a pgxpool connection.
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
