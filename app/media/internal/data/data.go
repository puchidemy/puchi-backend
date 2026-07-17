package data

import (
	"context"
	"fmt"

	"github.com/google/wire"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/puchidemy/puchi-backend/app/media/internal/biz"
	"github.com/puchidemy/puchi-backend/app/media/internal/conf"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewMediaRepo, NewStorageProvider, wire.FieldsOf(new(*Data), "Pool"), wire.Bind(new(biz.MediaRepo), new(*MediaRepo)))

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

// NewStorageProvider returns S3Storage when R2/S3 is configured, otherwise MockStorage for local dev.
func NewStorageProvider(media *conf.Media) (biz.StorageProvider, error) {
	if media == nil || media.GetStorage() == nil || media.GetStorage().GetEndpoint() == "" {
		return &MockStorage{}, nil
	}

	accessKey, secretKey := storageCredentials(media.GetStorage())
	if accessKey == "" || secretKey == "" {
		return nil, fmt.Errorf("storage credentials are required when endpoint is configured")
	}

	storage, err := NewS3StorageFromConfig(media.GetStorage(), media.GetUpload())
	if err != nil {
		return nil, err
	}
	return storage, nil
}
