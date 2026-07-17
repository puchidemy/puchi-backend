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
