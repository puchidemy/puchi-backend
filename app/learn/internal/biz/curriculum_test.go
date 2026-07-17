package biz

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/puchidemy/puchi-backend/app/learn/internal/data/sqlc/gen"
)

const testTrialUnitID = "11111111-1111-1111-1111-111111111111"

type mockCurriculumRepo struct {
	getUnit    func(ctx context.Context, id string) (*gen.LearnUnit, error)
	getSkill   func(ctx context.Context, id string) (*gen.LearnSkill, error)
	listSkills func(ctx context.Context, unitID string) ([]gen.LearnSkill, error)
	listLessons func(ctx context.Context, skillID string) ([]gen.LearnLesson, error)
	getLesson  func(ctx context.Context, id string) (*gen.LearnLesson, error)
	listExercises func(ctx context.Context, lessonID string) ([]gen.LearnExercise, error)
	getExercise func(ctx context.Context, id string) (*gen.LearnExercise, error)
}

func (m *mockCurriculumRepo) GetUnitByID(ctx context.Context, id string) (*gen.LearnUnit, error) {
	return m.getUnit(ctx, id)
}

func (m *mockCurriculumRepo) GetSkillByID(ctx context.Context, id string) (*gen.LearnSkill, error) {
	return m.getSkill(ctx, id)
}

func (m *mockCurriculumRepo) ListSkillsByUnitID(ctx context.Context, unitID string) ([]gen.LearnSkill, error) {
	return m.listSkills(ctx, unitID)
}

func (m *mockCurriculumRepo) ListLessonsBySkillID(ctx context.Context, skillID string) ([]gen.LearnLesson, error) {
	return m.listLessons(ctx, skillID)
}

func (m *mockCurriculumRepo) GetLessonByID(ctx context.Context, id string) (*gen.LearnLesson, error) {
	return m.getLesson(ctx, id)
}

func (m *mockCurriculumRepo) ListExercisesByLessonID(ctx context.Context, lessonID string) ([]gen.LearnExercise, error) {
	return m.listExercises(ctx, lessonID)
}

func (m *mockCurriculumRepo) GetExerciseByID(ctx context.Context, id string) (*gen.LearnExercise, error) {
	if m.getExercise != nil {
		return m.getExercise(ctx, id)
	}
	return nil, pgx.ErrNoRows
}

func newCurriculumTestUsecase(curriculum CurriculumRepoInterface) *LearnUsecase {
	return NewLearnUsecase(&mockGuestRepo{}, &mockProgressRepo{}, curriculum, &mockAttemptRepo{}, NoOpLessonEventPublisher{}, &mockTxManager{
		guest:    &mockGuestRepo{},
		progress: &mockProgressRepo{},
	})
}

func TestGetUnit_GuestNonTrial_ReturnsTrialLimit(t *testing.T) {
	uc := newCurriculumTestUsecase(&mockCurriculumRepo{})

	_, err := uc.GetUnit(context.Background(), "guest", "guest-1", "22222222-2222-2222-2222-222222222222", testTrialUnitID)
	if !errors.Is(err, ErrTrialLimit) {
		t.Fatalf("expected ErrTrialLimit, got %v", err)
	}
}

func TestGetUnit_GuestTrial_ReturnsUnitWithLessons(t *testing.T) {
	skillID := "22222222-2222-2222-2222-222222222222"
	lessonID := "33333333-3333-3333-3333-333333333331"

	uc := newCurriculumTestUsecase(&mockCurriculumRepo{
		getUnit: func(_ context.Context, id string) (*gen.LearnUnit, error) {
			if id != testTrialUnitID {
				t.Fatalf("unexpected unit id %q", id)
			}
			return &gen.LearnUnit{
				ID:       testTrialUnitID,
				CourseID: "00000000-0000-0000-0000-000000000001",
				Position: 1,
				Title:    "Trial Unit",
			}, nil
		},
		listSkills: func(_ context.Context, unitID string) ([]gen.LearnSkill, error) {
			if unitID != testTrialUnitID {
				t.Fatalf("unexpected unit id %q", unitID)
			}
			return []gen.LearnSkill{{
				ID:       skillID,
				UnitID:   testTrialUnitID,
				Position: 1,
				Title:    "Greetings",
			}}, nil
		},
		listLessons: func(_ context.Context, sid string) ([]gen.LearnLesson, error) {
			if sid != skillID {
				t.Fatalf("unexpected skill id %q", sid)
			}
			return []gen.LearnLesson{{
				ID:       lessonID,
				SkillID:  skillID,
				Position: 1,
				Title:    "Say Hello",
				XpReward: 10,
			}}, nil
		},
	})

	unit, err := uc.GetUnit(context.Background(), "guest", "guest-1", testTrialUnitID, testTrialUnitID)
	if err != nil {
		t.Fatalf("GetUnit: %v", err)
	}
	if unit.Unit.Title != "Trial Unit" {
		t.Fatalf("unit title = %q, want Trial Unit", unit.Unit.Title)
	}
	if len(unit.Skills) != 1 || len(unit.Skills[0].Lessons) != 1 {
		t.Fatalf("expected 1 skill with 1 lesson, got %+v", unit.Skills)
	}
	if unit.Skills[0].Lessons[0].ID != lessonID {
		t.Fatalf("lesson id = %q, want %q", unit.Skills[0].Lessons[0].ID, lessonID)
	}
}

func TestGetUnit_User_AnyUnitAllowed(t *testing.T) {
	unitID := "99999999-9999-9999-9999-999999999999"

	uc := newCurriculumTestUsecase(&mockCurriculumRepo{
		getUnit: func(_ context.Context, id string) (*gen.LearnUnit, error) {
			if id != unitID {
				t.Fatalf("unexpected unit id %q", id)
			}
			return &gen.LearnUnit{ID: unitID, Title: "Premium Unit"}, nil
		},
		listSkills: func(_ context.Context, _ string) ([]gen.LearnSkill, error) {
			return nil, nil
		},
	})

	unit, err := uc.GetUnit(context.Background(), "user", "user-1", unitID, testTrialUnitID)
	if err != nil {
		t.Fatalf("GetUnit: %v", err)
	}
	if unit.Unit.ID != unitID {
		t.Fatalf("unit id = %q, want %q", unit.Unit.ID, unitID)
	}
}

func TestGetLesson_GuestNonTrial_ReturnsTrialLimit(t *testing.T) {
	lessonID := "33333333-3333-3333-3333-333333333399"
	skillID := "22222222-2222-2222-2222-222222222299"

	uc := newCurriculumTestUsecase(&mockCurriculumRepo{
		getLesson: func(_ context.Context, id string) (*gen.LearnLesson, error) {
			return &gen.LearnLesson{ID: id, SkillID: skillID}, nil
		},
		getSkill: func(_ context.Context, id string) (*gen.LearnSkill, error) {
			return &gen.LearnSkill{ID: id, UnitID: "99999999-9999-9999-9999-999999999999"}, nil
		},
	})

	_, err := uc.GetLesson(context.Background(), "guest", "guest-1", lessonID, testTrialUnitID)
	if !errors.Is(err, ErrTrialLimit) {
		t.Fatalf("expected ErrTrialLimit, got %v", err)
	}
}

func TestGetLesson_GuestTrial_ReturnsLessonWithExercises(t *testing.T) {
	lessonID := "33333333-3333-3333-3333-333333333331"
	skillID := "22222222-2222-2222-2222-222222222222"

	uc := newCurriculumTestUsecase(&mockCurriculumRepo{
		getLesson: func(_ context.Context, id string) (*gen.LearnLesson, error) {
			return &gen.LearnLesson{
				ID:       id,
				SkillID:  skillID,
				Title:    "Say Hello",
				XpReward: 10,
			}, nil
		},
		getSkill: func(_ context.Context, id string) (*gen.LearnSkill, error) {
			return &gen.LearnSkill{ID: id, UnitID: testTrialUnitID}, nil
		},
		listExercises: func(_ context.Context, lid string) ([]gen.LearnExercise, error) {
			if lid != lessonID {
				t.Fatalf("unexpected lesson id %q", lid)
			}
			return []gen.LearnExercise{{
				ID:       "44444444-4444-4444-4444-444444444441",
				LessonID: lessonID,
				Position: 1,
				Type:     "select",
			}}, nil
		},
	})

	lesson, err := uc.GetLesson(context.Background(), "guest", "guest-1", lessonID, testTrialUnitID)
	if err != nil {
		t.Fatalf("GetLesson: %v", err)
	}
	if len(lesson.Exercises) != 1 {
		t.Fatalf("expected 1 exercise, got %d", len(lesson.Exercises))
	}
}

func TestGetUnit_NotFound_ReturnsErrNotFound(t *testing.T) {
	uc := newCurriculumTestUsecase(&mockCurriculumRepo{
		getUnit: func(_ context.Context, _ string) (*gen.LearnUnit, error) {
			return nil, pgx.ErrNoRows
		},
	})

	_, err := uc.GetUnit(context.Background(), "user", "user-1", testTrialUnitID, testTrialUnitID)
	if !errors.Is(err, ErrCurriculumNotFound) {
		t.Fatalf("expected ErrCurriculumNotFound, got %v", err)
	}
}
