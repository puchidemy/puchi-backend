package data

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/puchidemy/puchi-backend/app/learn/internal/data/sqlc/gen"
)

// ProgressRepo wraps sqlc progress and guest reassignment queries.
type ProgressRepo struct {
	q *gen.Queries
}

// NewProgressRepo creates a ProgressRepo.
func NewProgressRepo(pool *pgxpool.Pool) *ProgressRepo {
	return &ProgressRepo{q: gen.New(pool)}
}

// ListLessonProgressByOwner lists lesson progress rows for an owner.
func (r *ProgressRepo) ListLessonProgressByOwner(ctx context.Context, ownerType, ownerID string) ([]gen.LearnUserLessonProgress, error) {
	return r.q.ListLessonProgressByOwner(ctx, gen.ListLessonProgressByOwnerParams{
		OwnerType: ownerType,
		OwnerID:   ownerID,
	})
}

// ListUnitProgressByOwner lists unit progress rows for an owner.
func (r *ProgressRepo) ListUnitProgressByOwner(ctx context.Context, ownerType, ownerID string) ([]gen.LearnUserUnitProgress, error) {
	return r.q.ListUnitProgressByOwner(ctx, gen.ListUnitProgressByOwnerParams{
		OwnerType: ownerType,
		OwnerID:   ownerID,
	})
}

// GetLessonProgress returns lesson progress for an owner.
func (r *ProgressRepo) GetLessonProgress(ctx context.Context, ownerType, ownerID, lessonID string) (*gen.LearnUserLessonProgress, error) {
	row, err := r.q.GetLessonProgress(ctx, gen.GetLessonProgressParams{
		OwnerType: ownerType,
		OwnerID:   ownerID,
		LessonID:  lessonID,
	})
	return mapNoRows(row, err)
}

// GetUnitProgress returns unit progress for an owner.
func (r *ProgressRepo) GetUnitProgress(ctx context.Context, ownerType, ownerID, unitID string) (*gen.LearnUserUnitProgress, error) {
	row, err := r.q.GetUnitProgress(ctx, gen.GetUnitProgressParams{
		OwnerType: ownerType,
		OwnerID:   ownerID,
		UnitID:    unitID,
	})
	return mapNoRows(row, err)
}

// UpsertLessonProgress upserts lesson progress for an owner.
func (r *ProgressRepo) UpsertLessonProgress(ctx context.Context, ownerType, ownerID, lessonID, status string, xp int32) error {
	return r.q.UpsertLessonProgress(ctx, gen.UpsertLessonProgressParams{
		OwnerType: ownerType,
		OwnerID:   ownerID,
		LessonID:  lessonID,
		Status:    status,
		XpEarned:  xp,
	})
}

// UpsertUnitProgress upserts unit progress for an owner.
func (r *ProgressRepo) UpsertUnitProgress(ctx context.Context, ownerType, ownerID, unitID, status string) error {
	return r.q.UpsertUnitProgress(ctx, gen.UpsertUnitProgressParams{
		OwnerType: ownerType,
		OwnerID:   ownerID,
		UnitID:    unitID,
		Status:    status,
	})
}

// DeleteGuestLessonProgress removes a guest lesson progress row.
func (r *ProgressRepo) DeleteGuestLessonProgress(ctx context.Context, guestID, lessonID string) error {
	return r.q.DeleteGuestLessonProgress(ctx, gen.DeleteGuestLessonProgressParams{
		OwnerID:  guestID,
		LessonID: lessonID,
	})
}

// DeleteGuestUnitProgress removes a guest unit progress row.
func (r *ProgressRepo) DeleteGuestUnitProgress(ctx context.Context, guestID, unitID string) error {
	return r.q.DeleteGuestUnitProgress(ctx, gen.DeleteGuestUnitProgressParams{
		OwnerID: guestID,
		UnitID:  unitID,
	})
}

// ReassignGuestLessonProgress moves remaining guest lesson rows to the user.
func (r *ProgressRepo) ReassignGuestLessonProgress(ctx context.Context, guestID, userID string) error {
	return r.q.ReassignGuestLessonProgress(ctx, gen.ReassignGuestLessonProgressParams{
		OwnerID:   guestID,
		OwnerID_2: userID,
	})
}

// ReassignGuestUnitProgress moves remaining guest unit rows to the user.
func (r *ProgressRepo) ReassignGuestUnitProgress(ctx context.Context, guestID, userID string) error {
	return r.q.ReassignGuestUnitProgress(ctx, gen.ReassignGuestUnitProgressParams{
		OwnerID:   guestID,
		OwnerID_2: userID,
	})
}

// ReassignGuestAttempts moves guest attempt rows to the user.
func (r *ProgressRepo) ReassignGuestAttempts(ctx context.Context, guestID, userID string) error {
	return r.q.ReassignGuestAttempts(ctx, gen.ReassignGuestAttemptsParams{
		OwnerID:   guestID,
		OwnerID_2: userID,
	})
}

// GetStoryProgress returns story progress for an owner.
func (r *ProgressRepo) GetStoryProgress(ctx context.Context, ownerType, ownerID, storyID string) (*gen.LearnUserStoryProgress, error) {
	row, err := r.q.GetStoryProgress(ctx, gen.GetStoryProgressParams{
		OwnerType: ownerType,
		OwnerID:   ownerID,
		StoryID:   storyID,
	})
	return mapNoRows(row, err)
}

// UpsertStoryProgress upserts story progress for an owner.
func (r *ProgressRepo) UpsertStoryProgress(ctx context.Context, ownerType, ownerID, storyID, status string, xp int32) error {
	return r.q.UpsertStoryProgress(ctx, gen.UpsertStoryProgressParams{
		OwnerType: ownerType,
		OwnerID:   ownerID,
		StoryID:   storyID,
		Status:    status,
		XpEarned:  xp,
	})
}

// ListStoryProgressByOwner lists story progress rows for an owner.
func (r *ProgressRepo) ListStoryProgressByOwner(ctx context.Context, ownerType, ownerID string) ([]gen.LearnUserStoryProgress, error) {
	return r.q.ListStoryProgressByOwner(ctx, gen.ListStoryProgressByOwnerParams{
		OwnerType: ownerType,
		OwnerID:   ownerID,
	})
}

// DeleteGuestStoryProgress removes a guest story progress row.
func (r *ProgressRepo) DeleteGuestStoryProgress(ctx context.Context, guestID, storyID string) error {
	return r.q.DeleteGuestStoryProgress(ctx, gen.DeleteGuestStoryProgressParams{
		OwnerID: guestID,
		StoryID: storyID,
	})
}

// ReassignGuestStoryProgress moves remaining guest story rows to the user.
func (r *ProgressRepo) ReassignGuestStoryProgress(ctx context.Context, guestID, userID string) error {
	return r.q.ReassignGuestStoryProgress(ctx, gen.ReassignGuestStoryProgressParams{
		OwnerID:   guestID,
		OwnerID_2: userID,
	})
}

// GetSceneProgress returns scene progress for an owner.
func (r *ProgressRepo) GetSceneProgress(ctx context.Context, ownerType, ownerID, sceneID string) (*gen.LearnUserSceneProgress, error) {
	row, err := r.q.GetSceneProgress(ctx, gen.GetSceneProgressParams{
		OwnerType: ownerType,
		OwnerID:   ownerID,
		SceneID:   sceneID,
	})
	return mapNoRows(row, err)
}

// UpsertSceneProgress upserts scene progress for an owner.
func (r *ProgressRepo) UpsertSceneProgress(ctx context.Context, ownerType, ownerID, sceneID, status string) error {
	return r.q.UpsertSceneProgress(ctx, gen.UpsertSceneProgressParams{
		OwnerType: ownerType,
		OwnerID:   ownerID,
		SceneID:   sceneID,
		Status:    status,
	})
}

// ListSceneProgressByOwner lists scene progress rows for an owner.
func (r *ProgressRepo) ListSceneProgressByOwner(ctx context.Context, ownerType, ownerID string) ([]gen.LearnUserSceneProgress, error) {
	return r.q.ListSceneProgressByOwner(ctx, gen.ListSceneProgressByOwnerParams{
		OwnerType: ownerType,
		OwnerID:   ownerID,
	})
}

// CountCompletedScenesByOwner returns completed scene count for an owner.
func (r *ProgressRepo) CountCompletedScenesByOwner(ctx context.Context, ownerType, ownerID string) (int32, error) {
	return r.q.CountCompletedScenesByOwner(ctx, gen.CountCompletedScenesByOwnerParams{
		OwnerType: ownerType,
		OwnerID:   ownerID,
	})
}

// DeleteGuestSceneProgress removes a guest scene progress row.
func (r *ProgressRepo) DeleteGuestSceneProgress(ctx context.Context, guestID, sceneID string) error {
	return r.q.DeleteGuestSceneProgress(ctx, gen.DeleteGuestSceneProgressParams{
		OwnerID: guestID,
		SceneID: sceneID,
	})
}

// ReassignGuestSceneProgress moves remaining guest scene rows to the user.
func (r *ProgressRepo) ReassignGuestSceneProgress(ctx context.Context, guestID, userID string) error {
	return r.q.ReassignGuestSceneProgress(ctx, gen.ReassignGuestSceneProgressParams{
		OwnerID:   guestID,
		OwnerID_2: userID,
	})
}

// ReassignGuestActivityAttempts moves guest activity attempt rows to the user.
func (r *ProgressRepo) ReassignGuestActivityAttempts(ctx context.Context, guestID, userID string) error {
	return r.q.ReassignGuestActivityAttempts(ctx, gen.ReassignGuestActivityAttemptsParams{
		OwnerID:   guestID,
		OwnerID_2: userID,
	})
}
