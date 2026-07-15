package data

import (
	"context"

	"github.com/puchidemy/puchi-backend/app/notification/internal/biz"
	"github.com/puchidemy/puchi-backend/app/notification/internal/conf"

	"github.com/google/wire"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewTodoRepo, NewPreferenceRepo, NewGotifyClient, wire.FieldsOf(new(*Data), "Pool"), wire.Bind(new(biz.PreferenceRepo), new(*PreferenceRepo)), wire.Bind(new(biz.GotifySender), new(*GotifyClient)))

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

var _ biz.PreferenceRepo = (*PreferenceRepo)(nil)
