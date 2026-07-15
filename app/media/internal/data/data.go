package data

import (
	"context"

	"github.com/google/wire"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/puchidemy/puchi-backend/app/media/internal/biz"
	"github.com/puchidemy/puchi-backend/app/media/internal/conf"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewMediaRepo, NewStorageProvider, wire.FieldsOf(new(*Data), "Pool"), wire.Bind(new(biz.MediaRepo), new(*MediaRepo)), wire.Bind(new(biz.StorageProvider), new(*MockStorage)))

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

// NewStorageProvider returns a MockStorage for now.
// Replace with a real MinIO/Garage client when ready.
func NewStorageProvider() *MockStorage {
	return &MockStorage{}
}
