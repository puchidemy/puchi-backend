package biz

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/puchidemy/puchi-backend/app/learn/internal/data/sqlc/gen"
)

var (
	ErrTrialLimit          = errors.New("trial limit")
	ErrCurriculumNotFound  = errors.New("curriculum not found")
)

// CurriculumRepoInterface reads curriculum rows.
type CurriculumRepoInterface interface {
	GetUnitByID(ctx context.Context, id string) (*gen.LearnUnit, error)
	GetSkillByID(ctx context.Context, id string) (*gen.LearnSkill, error)
	ListSkillsByUnitID(ctx context.Context, unitID string) ([]gen.LearnSkill, error)
	ListLessonsBySkillID(ctx context.Context, skillID string) ([]gen.LearnLesson, error)
	GetLessonByID(ctx context.Context, id string) (*gen.LearnLesson, error)
	ListExercisesByLessonID(ctx context.Context, lessonID string) ([]gen.LearnExercise, error)
}

// SkillWithLessons groups a skill and its lessons.
type SkillWithLessons struct {
	Skill   gen.LearnSkill
	Lessons []gen.LearnLesson
}

// UnitDetail is a unit with nested skills and lessons.
type UnitDetail struct {
	Unit   gen.LearnUnit
	Skills []SkillWithLessons
}

// LessonDetail is a lesson with exercises.
type LessonDetail struct {
	Lesson    gen.LearnLesson
	Exercises []gen.LearnExercise
}

func (uc *LearnUsecase) assertGuestTrialScope(ownerType, resourceUnitID, trialUnitID string) error {
	if ownerType == "guest" && resourceUnitID != trialUnitID {
		return ErrTrialLimit
	}
	return nil
}

// GetUnit returns a unit with skills and lessons, enforcing guest trial scope.
func (uc *LearnUsecase) GetUnit(ctx context.Context, ownerType, ownerID, unitID, trialUnitID string) (*UnitDetail, error) {
	if err := uc.assertGuestTrialScope(ownerType, unitID, trialUnitID); err != nil {
		return nil, err
	}

	unit, err := uc.curriculumRepo.GetUnitByID(ctx, unitID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCurriculumNotFound
		}
		return nil, err
	}

	skills, err := uc.curriculumRepo.ListSkillsByUnitID(ctx, unitID)
	if err != nil {
		return nil, err
	}

	out := &UnitDetail{Unit: *unit}
	for _, skill := range skills {
		lessons, err := uc.curriculumRepo.ListLessonsBySkillID(ctx, skill.ID)
		if err != nil {
			return nil, err
		}
		out.Skills = append(out.Skills, SkillWithLessons{
			Skill:   skill,
			Lessons: lessons,
		})
	}
	return out, nil
}

// GetLesson returns a lesson with exercises, enforcing guest trial scope via skill unit.
func (uc *LearnUsecase) GetLesson(ctx context.Context, ownerType, ownerID, lessonID, trialUnitID string) (*LessonDetail, error) {
	lesson, err := uc.curriculumRepo.GetLessonByID(ctx, lessonID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCurriculumNotFound
		}
		return nil, err
	}

	skill, err := uc.curriculumRepo.GetSkillByID(ctx, lesson.SkillID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCurriculumNotFound
		}
		return nil, err
	}

	if err := uc.assertGuestTrialScope(ownerType, skill.UnitID, trialUnitID); err != nil {
		return nil, err
	}

	exercises, err := uc.curriculumRepo.ListExercisesByLessonID(ctx, lessonID)
	if err != nil {
		return nil, err
	}

	return &LessonDetail{
		Lesson:    *lesson,
		Exercises: exercises,
	}, nil
}
