package data

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/puchidemy/puchi-backend/app/learn/internal/data/sqlc/gen"
)

// CurriculumRepo wraps sqlc curriculum queries.
type CurriculumRepo struct {
	q *gen.Queries
}

// NewCurriculumRepo creates a CurriculumRepo.
func NewCurriculumRepo(pool *pgxpool.Pool) *CurriculumRepo {
	return &CurriculumRepo{q: gen.New(pool)}
}

// GetUnitByID returns a unit by ID.
func (r *CurriculumRepo) GetUnitByID(ctx context.Context, id string) (*gen.LearnUnit, error) {
	return mapNoRows(r.q.GetUnitByID(ctx, id))
}

// GetSkillByID returns a skill by ID.
func (r *CurriculumRepo) GetSkillByID(ctx context.Context, id string) (*gen.LearnSkill, error) {
	return mapNoRows(r.q.GetSkillByID(ctx, id))
}

// ListSkillsByUnitID returns skills for a unit ordered by position.
func (r *CurriculumRepo) ListSkillsByUnitID(ctx context.Context, unitID string) ([]gen.LearnSkill, error) {
	return r.q.ListSkillsByUnitID(ctx, unitID)
}

// ListLessonsBySkillID returns lessons for a skill ordered by position.
func (r *CurriculumRepo) ListLessonsBySkillID(ctx context.Context, skillID string) ([]gen.LearnLesson, error) {
	return r.q.ListLessonsBySkillID(ctx, skillID)
}

// GetLessonByID returns a lesson by ID.
func (r *CurriculumRepo) GetLessonByID(ctx context.Context, id string) (*gen.LearnLesson, error) {
	return mapNoRows(r.q.GetLessonByID(ctx, id))
}

// ListExercisesByLessonID returns exercises for a lesson ordered by position.
func (r *CurriculumRepo) ListExercisesByLessonID(ctx context.Context, lessonID string) ([]gen.LearnExercise, error) {
	return r.q.ListExercisesByLessonID(ctx, lessonID)
}
