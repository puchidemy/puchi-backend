package data

import (
	"context"

	"github.com/google/wire"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/puchidemy/puchi-backend/app/core/internal/biz"
	"github.com/puchidemy/puchi-backend/app/core/internal/conf"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewUserRepo, wire.FieldsOf(new(*Data), "Pool"), wire.Bind(new(biz.UserRepoInterface), new(*UserRepo)))

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
