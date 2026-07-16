package data

import (
	"context"

	"github.com/puchidemy/puchi-backend/app/auth/internal/biz"
	"github.com/puchidemy/puchi-backend/app/auth/internal/conf"
	"github.com/google/wire"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData,
	NewUserRepo,
	NewSessionRepo,
	wire.Bind(new(biz.UserRepo), new(*UserRepo)),
	wire.Bind(new(biz.SessionRepo), new(*SessionRepo)),
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
