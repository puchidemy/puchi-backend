package data

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/puchidemy/puchi-backend/app/media/internal/data/sqlc/gen"
)

// MediaRepo wraps sqlc-generated queries for media.objects.
type MediaRepo struct {
	q *gen.Queries
}

// NewMediaRepo creates a new MediaRepo.
func NewMediaRepo(pool *pgxpool.Pool) *MediaRepo {
	return &MediaRepo{q: gen.New(pool)}
}

// CreateMediaObject inserts a new media object record.
func (r *MediaRepo) CreateMediaObject(ctx context.Context, arg gen.CreateMediaObjectParams) (*gen.MediaObject, error) {
	row, err := r.q.CreateMediaObject(ctx, arg)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// GetMediaObject retrieves a media object by ID.
func (r *MediaRepo) GetMediaObject(ctx context.Context, id int64) (*gen.MediaObject, error) {
	row, err := r.q.GetMediaObject(ctx, id)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// GetMediaObjectByKey retrieves a media object by object key.
func (r *MediaRepo) GetMediaObjectByKey(ctx context.Context, objectKey string) (*gen.MediaObject, error) {
	row, err := r.q.GetMediaObjectByKey(ctx, objectKey)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// ListUserMedia lists all media objects for a user.
func (r *MediaRepo) ListUserMedia(ctx context.Context, userID string) ([]*gen.MediaObject, error) {
	rows, err := r.q.ListUserMedia(ctx, userID)
	if err != nil {
		return nil, err
	}
	items := make([]*gen.MediaObject, len(rows))
	for i := range rows {
		items[i] = &rows[i]
	}
	return items, nil
}

// UpdateMediaStatus updates the status of a media object.
func (r *MediaRepo) UpdateMediaStatus(ctx context.Context, id int64, status string) (*gen.MediaObject, error) {
	row, err := r.q.UpdateMediaStatus(ctx, gen.UpdateMediaStatusParams{
		ID:     id,
		Status: status,
	})
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// DeleteMediaObject deletes a media object by ID.
func (r *MediaRepo) DeleteMediaObject(ctx context.Context, id int64) error {
	return r.q.DeleteMediaObject(ctx, id)
}
