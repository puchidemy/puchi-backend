package data

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/puchidemy/puchi-backend/app/learn/internal/data/sqlc/gen"
)

// StoryRepo wraps sqlc story-first curriculum queries.
type StoryRepo struct {
	q *gen.Queries
}

// NewStoryRepo creates a StoryRepo.
func NewStoryRepo(pool *pgxpool.Pool) *StoryRepo {
	return &StoryRepo{q: gen.New(pool)}
}

func (r *StoryRepo) ListCities(ctx context.Context) ([]gen.LearnCity, error) {
	return r.q.ListCities(ctx)
}

func (r *StoryRepo) GetCityBySlug(ctx context.Context, slug string) (*gen.LearnCity, error) {
	return mapNoRows(r.q.GetCityBySlug(ctx, slug))
}

func (r *StoryRepo) GetCityByID(ctx context.Context, id string) (*gen.LearnCity, error) {
	return mapNoRows(r.q.GetCityByID(ctx, id))
}

func (r *StoryRepo) CountPublishedStoriesByCity(ctx context.Context, cityID string) (int32, error) {
	return r.q.CountPublishedStoriesByCity(ctx, cityID)
}

func (r *StoryRepo) CountCompletedStoriesByOwnerCity(ctx context.Context, ownerType, ownerID, cityID string) (int32, error) {
	return r.q.CountCompletedStoriesByOwnerCity(ctx, gen.CountCompletedStoriesByOwnerCityParams{
		OwnerType: ownerType,
		OwnerID:   ownerID,
		CityID:    cityID,
	})
}

func (r *StoryRepo) ListPublishedStoriesByCity(ctx context.Context, cityID string) ([]gen.LearnStory, error) {
	return r.q.ListPublishedStoriesByCity(ctx, cityID)
}

func (r *StoryRepo) GetStoryByID(ctx context.Context, id string) (*gen.LearnStory, error) {
	return mapNoRows(r.q.GetStoryByID(ctx, id))
}

func (r *StoryRepo) ListScenesByStoryID(ctx context.Context, storyID string) ([]gen.LearnScene, error) {
	return r.q.ListScenesByStoryID(ctx, storyID)
}

func (r *StoryRepo) GetSceneByID(ctx context.Context, id string) (*gen.LearnScene, error) {
	return mapNoRows(r.q.GetSceneByID(ctx, id))
}

func (r *StoryRepo) ListActivitiesBySceneID(ctx context.Context, sceneID string) ([]gen.LearnActivity, error) {
	return r.q.ListActivitiesBySceneID(ctx, sceneID)
}

func (r *StoryRepo) GetActivityByID(ctx context.Context, id string) (*gen.LearnActivity, error) {
	return mapNoRows(r.q.GetActivityByID(ctx, id))
}

func (r *StoryRepo) ListActivitiesByStoryID(ctx context.Context, storyID string) ([]gen.LearnActivity, error) {
	return r.q.ListActivitiesByStoryID(ctx, storyID)
}
