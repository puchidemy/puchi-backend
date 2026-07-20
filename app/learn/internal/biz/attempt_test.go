package biz

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/puchidemy/puchi-backend/app/learn/internal/data/sqlc/gen"
)

type mockAttemptRepo struct {
	createAttempt      func(ctx context.Context, ownerType, ownerID, lessonID string) (*gen.LearnAttempt, error)
	getAttempt         func(ctx context.Context, id string) (*gen.LearnAttempt, error)
	getActiveAttempt   func(ctx context.Context, ownerType, ownerID, lessonID string) (*gen.LearnAttempt, error)
	insertAnswer       func(ctx context.Context, attemptID, exerciseID string, payload json.RawMessage, correct bool) error
	completeAttempt    func(ctx context.Context, attemptID string, sessionXP int32) error
	listAnswers        func(ctx context.Context, attemptID string) ([]gen.LearnAttemptAnswer, error)
}

func (m *mockAttemptRepo) CreateAttempt(ctx context.Context, ownerType, ownerID, lessonID string) (*gen.LearnAttempt, error) {
	return m.createAttempt(ctx, ownerType, ownerID, lessonID)
}

func (m *mockAttemptRepo) GetAttemptByID(ctx context.Context, id string) (*gen.LearnAttempt, error) {
	return m.getAttempt(ctx, id)
}

func (m *mockAttemptRepo) GetActiveAttemptByOwnerLesson(ctx context.Context, ownerType, ownerID, lessonID string) (*gen.LearnAttempt, error) {
	return m.getActiveAttempt(ctx, ownerType, ownerID, lessonID)
}

func (m *mockAttemptRepo) InsertAttemptAnswer(ctx context.Context, attemptID, exerciseID string, payload json.RawMessage, correct bool) error {
	return m.insertAnswer(ctx, attemptID, exerciseID, payload, correct)
}

func (m *mockAttemptRepo) CompleteAttempt(ctx context.Context, attemptID string, sessionXP int32) error {
	return m.completeAttempt(ctx, attemptID, sessionXP)
}

func (m *mockAttemptRepo) ListAttemptAnswersByAttemptID(ctx context.Context, attemptID string) ([]gen.LearnAttemptAnswer, error) {
	return m.listAnswers(ctx, attemptID)
}

func (m *mockAttemptRepo) CreateActivityAttempt(context.Context, string, string, string, string) (*gen.LearnActivityAttempt, error) {
	return nil, pgx.ErrNoRows
}
func (m *mockAttemptRepo) GetActivityAttemptByID(context.Context, string) (*gen.LearnActivityAttempt, error) {
	return nil, pgx.ErrNoRows
}
func (m *mockAttemptRepo) GetActiveActivityAttemptByOwnerScene(context.Context, string, string, string) (*gen.LearnActivityAttempt, error) {
	return nil, pgx.ErrNoRows
}
func (m *mockAttemptRepo) InsertActivityAttemptAnswer(context.Context, string, string, json.RawMessage, bool) error {
	return nil
}
func (m *mockAttemptRepo) CompleteActivityAttempt(context.Context, string, int32) error {
	return nil
}
func (m *mockAttemptRepo) ListActivityAttemptAnswersByAttemptID(context.Context, string) ([]gen.LearnActivityAttemptAnswer, error) {
	return nil, nil
}

type mockPublisher struct {
	lessonCompleted func(ctx context.Context, ev LessonCompletedEvent) error
	unitCompleted   func(ctx context.Context, ev UnitCompletedEvent) error
}

func (m *mockPublisher) PublishLessonCompleted(ctx context.Context, ev LessonCompletedEvent) error {
	if m.lessonCompleted != nil {
		return m.lessonCompleted(ctx, ev)
	}
	return nil
}

func (m *mockPublisher) PublishUnitCompleted(ctx context.Context, ev UnitCompletedEvent) error {
	if m.unitCompleted != nil {
		return m.unitCompleted(ctx, ev)
	}
	return nil
}

func (m *mockPublisher) PublishSceneCompleted(context.Context, SceneCompletedEvent) error {
	return nil
}
func (m *mockPublisher) PublishStoryCompleted(context.Context, StoryCompletedEvent) error {
	return nil
}

func newAttemptTestUsecase(curriculum CurriculumRepoInterface, progress ProgressRepoInterface, attempts AttemptRepoInterface, publisher LessonEventPublisher) *LearnUsecase {
	if publisher == nil {
		publisher = NoOpLessonEventPublisher{}
	}
	return NewLearnUsecase(&mockGuestRepo{}, progress, curriculum, &mockStoryRepo{}, attempts, publisher, &mockTxManager{
		guest:    &mockGuestRepo{},
		progress: progress,
	})
}

func TestGuestCompletesOneLesson(t *testing.T) {
	lessonID := "33333333-3333-3333-3333-333333333331"
	skillID := "22222222-2222-2222-2222-222222222222"
	exerciseID := "44444444-4444-4444-4444-444444444441"
	attemptID := "55555555-5555-5555-5555-555555555551"
	guestID := "guest-1"

	var lessonStatus string
	var lessonXP int32
	attemptCompleted := false

	curriculum := &mockCurriculumRepo{
		getLesson: func(_ context.Context, id string) (*gen.LearnLesson, error) {
			return &gen.LearnLesson{ID: id, SkillID: skillID, XpReward: 10, Required: true}, nil
		},
		getSkill: func(_ context.Context, id string) (*gen.LearnSkill, error) {
			return &gen.LearnSkill{ID: id, UnitID: testTrialUnitID}, nil
		},
		listSkills: func(_ context.Context, unitID string) ([]gen.LearnSkill, error) {
			return []gen.LearnSkill{{ID: skillID, UnitID: unitID}}, nil
		},
		listLessons: func(_ context.Context, sid string) ([]gen.LearnLesson, error) {
			return []gen.LearnLesson{{ID: lessonID, SkillID: sid, Required: true}}, nil
		},
		getExercise: func(_ context.Context, id string) (*gen.LearnExercise, error) {
			return &gen.LearnExercise{
				ID:       id,
				LessonID: lessonID,
				Type:     "select",
				Prompt:   json.RawMessage(`{"options":["Hello","Bye"]}`),
				Answer:   json.RawMessage(`{"correct":"Hello"}`),
			}, nil
		},
		listExercises: func(_ context.Context, lid string) ([]gen.LearnExercise, error) {
			return []gen.LearnExercise{{ID: exerciseID, LessonID: lid}}, nil
		},
	}

	progress := &mockProgressRepo{
		listLessons: func(_ context.Context, ownerType, ownerID string) ([]gen.LearnUserLessonProgress, error) {
			if lessonStatus == "completed" {
				return []gen.LearnUserLessonProgress{{
					OwnerType: ownerType,
					OwnerID:   ownerID,
					LessonID:  lessonID,
					Status:    "completed",
				}}, nil
			}
			return nil, nil
		},
		getLesson: func(_ context.Context, ownerType, ownerID, lid string) (*gen.LearnUserLessonProgress, error) {
			if ownerType == "guest" && ownerID == guestID && lid == lessonID && lessonStatus != "" {
				return &gen.LearnUserLessonProgress{Status: lessonStatus, XpEarned: lessonXP}, nil
			}
			return nil, pgx.ErrNoRows
		},
		upsertLesson: func(_ context.Context, ownerType, ownerID, lid, status string, xp int32) error {
			if ownerType != "guest" || ownerID != guestID || lid != lessonID {
				t.Fatalf("unexpected upsert lesson %s/%s/%s", ownerType, ownerID, lid)
			}
			lessonStatus = status
			lessonXP = xp
			return nil
		},
		getUnit: func(_ context.Context, _, _, _ string) (*gen.LearnUserUnitProgress, error) {
			return nil, pgx.ErrNoRows
		},
		upsertUnit: func(_ context.Context, ownerType, ownerID, unitID, status string) error {
			if ownerType != "guest" || ownerID != guestID || unitID != testTrialUnitID || status != "completed" {
				t.Fatalf("unexpected upsert unit %s/%s/%s status=%s", ownerType, ownerID, unitID, status)
			}
			return nil
		},
	}

	attempts := &mockAttemptRepo{
		createAttempt: func(_ context.Context, ownerType, ownerID, lid string) (*gen.LearnAttempt, error) {
			return &gen.LearnAttempt{ID: attemptID, OwnerType: ownerType, OwnerID: ownerID, LessonID: lid, Status: "active"}, nil
		},
		getAttempt: func(_ context.Context, id string) (*gen.LearnAttempt, error) {
			return &gen.LearnAttempt{ID: id, OwnerType: "guest", OwnerID: guestID, LessonID: lessonID, Status: "active"}, nil
		},
		getActiveAttempt: func(_ context.Context, ownerType, ownerID, lid string) (*gen.LearnAttempt, error) {
			return &gen.LearnAttempt{ID: attemptID, OwnerType: ownerType, OwnerID: ownerID, LessonID: lid, Status: "active"}, nil
		},
		insertAnswer: func(_ context.Context, aid, eid string, _ json.RawMessage, correct bool) error {
			if aid != attemptID || eid != exerciseID || !correct {
				t.Fatalf("unexpected insert answer aid=%s eid=%s correct=%v", aid, eid, correct)
			}
			return nil
		},
		completeAttempt: func(_ context.Context, id string, sessionXP int32) error {
			if id != attemptID || sessionXP != 10 {
				t.Fatalf("complete attempt id=%s xp=%d", id, sessionXP)
			}
			attemptCompleted = true
			return nil
		},
		listAnswers: func(_ context.Context, _ string) ([]gen.LearnAttemptAnswer, error) {
			return []gen.LearnAttemptAnswer{{Correct: true}}, nil
		},
	}

	publisher := &mockPublisher{
		lessonCompleted: func(_ context.Context, _ LessonCompletedEvent) error {
			t.Fatal("guest complete must not publish lesson event")
			return nil
		},
	}

	uc := newAttemptTestUsecase(curriculum, progress, attempts, publisher)
	ctx := context.Background()

	gotAttempt, err := uc.StartLesson(ctx, "guest", guestID, lessonID, testTrialUnitID)
	if err != nil {
		t.Fatalf("StartLesson: %v", err)
	}
	if gotAttempt != uuid.MustParse(attemptID) {
		t.Fatalf("attempt id = %v", gotAttempt)
	}
	if lessonStatus != "in_progress" {
		t.Fatalf("lesson status after start = %q", lessonStatus)
	}

	correct, err := uc.SubmitAnswer(ctx, "guest", guestID, attemptID, exerciseID, json.RawMessage(`{"answer":"Hello"}`), testTrialUnitID)
	if err != nil {
		t.Fatalf("SubmitAnswer: %v", err)
	}
	if !correct {
		t.Fatal("expected correct answer")
	}

	xp, unitCompleted, err := uc.CompleteLesson(ctx, "guest", guestID, lessonID, testTrialUnitID)
	if err != nil {
		t.Fatalf("CompleteLesson: %v", err)
	}
	if xp != 10 {
		t.Fatalf("xp = %d, want 10", xp)
	}
	if !unitCompleted {
		t.Fatal("expected unit completed")
	}
	if lessonStatus != "completed" {
		t.Fatalf("lesson status after complete = %q", lessonStatus)
	}
	if !attemptCompleted {
		t.Fatal("expected attempt completed")
	}
}

func threeCompletedLessons(ownerType, ownerID string) []gen.LearnUserLessonProgress {
	return []gen.LearnUserLessonProgress{
		{OwnerType: ownerType, OwnerID: ownerID, LessonID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1", Status: "completed"},
		{OwnerType: ownerType, OwnerID: ownerID, LessonID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa2", Status: "completed"},
		{OwnerType: ownerType, OwnerID: ownerID, LessonID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa3", Status: "completed"},
	}
}

func TestStartLesson_GuestSoftGate_BlocksWhenThreeCompleted(t *testing.T) {
	lessonID := "33333333-3333-3333-3333-333333333399"
	skillID := "22222222-2222-2222-2222-222222222299"
	guestID := "guest-1"

	curriculum := &mockCurriculumRepo{
		getLesson: func(_ context.Context, id string) (*gen.LearnLesson, error) {
			return &gen.LearnLesson{ID: id, SkillID: skillID}, nil
		},
		getSkill: func(_ context.Context, id string) (*gen.LearnSkill, error) {
			return &gen.LearnSkill{ID: id, UnitID: "99999999-9999-9999-9999-999999999999"}, nil
		},
	}
	progress := &mockProgressRepo{
		listLessons: func(_ context.Context, ownerType, ownerID string) ([]gen.LearnUserLessonProgress, error) {
			return threeCompletedLessons(ownerType, ownerID), nil
		},
		getLesson: func(_ context.Context, _, _, _ string) (*gen.LearnUserLessonProgress, error) {
			return nil, pgx.ErrNoRows
		},
	}
	attempts := &mockAttemptRepo{
		createAttempt: func(_ context.Context, _, _, _ string) (*gen.LearnAttempt, error) {
			t.Fatal("CreateAttempt must not be called when soft-gate blocks")
			return nil, nil
		},
	}

	uc := newAttemptTestUsecase(curriculum, progress, attempts, nil)
	_, err := uc.StartLesson(context.Background(), "guest", guestID, lessonID, testTrialUnitID)
	if !errors.Is(err, ErrGuestSoftGate) {
		t.Fatalf("expected ErrGuestSoftGate, got %v", err)
	}
}

func TestStartLesson_GuestSoftGate_AllowsAlreadyCompleted(t *testing.T) {
	lessonID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1"
	skillID := "22222222-2222-2222-2222-222222222222"
	attemptID := "55555555-5555-5555-5555-555555555599"
	guestID := "guest-1"

	curriculum := &mockCurriculumRepo{
		getLesson: func(_ context.Context, id string) (*gen.LearnLesson, error) {
			return &gen.LearnLesson{ID: id, SkillID: skillID}, nil
		},
		getSkill: func(_ context.Context, id string) (*gen.LearnSkill, error) {
			return &gen.LearnSkill{ID: id, UnitID: "99999999-9999-9999-9999-999999999999"}, nil
		},
	}
	progress := &mockProgressRepo{
		listLessons: func(_ context.Context, ownerType, ownerID string) ([]gen.LearnUserLessonProgress, error) {
			return threeCompletedLessons(ownerType, ownerID), nil
		},
		getLesson: func(_ context.Context, _, _, lid string) (*gen.LearnUserLessonProgress, error) {
			if lid == lessonID {
				return &gen.LearnUserLessonProgress{LessonID: lid, Status: "completed"}, nil
			}
			return nil, pgx.ErrNoRows
		},
		upsertLesson: func(_ context.Context, _, _, _, _ string, _ int32) error {
			return nil
		},
	}
	attempts := &mockAttemptRepo{
		createAttempt: func(_ context.Context, ownerType, ownerID, lid string) (*gen.LearnAttempt, error) {
			return &gen.LearnAttempt{ID: attemptID, OwnerType: ownerType, OwnerID: ownerID, LessonID: lid, Status: "active"}, nil
		},
	}

	uc := newAttemptTestUsecase(curriculum, progress, attempts, nil)
	got, err := uc.StartLesson(context.Background(), "guest", guestID, lessonID, testTrialUnitID)
	if err != nil {
		t.Fatalf("StartLesson: %v", err)
	}
	if got != uuid.MustParse(attemptID) {
		t.Fatalf("attempt id = %v", got)
	}
}

func TestStartLesson_GuestSoftGate_AllowsWhenUnderThree(t *testing.T) {
	lessonID := "33333333-3333-3333-3333-333333333399"
	skillID := "22222222-2222-2222-2222-222222222299"
	attemptID := "55555555-5555-5555-5555-555555555598"
	guestID := "guest-1"

	curriculum := &mockCurriculumRepo{
		getLesson: func(_ context.Context, id string) (*gen.LearnLesson, error) {
			return &gen.LearnLesson{ID: id, SkillID: skillID}, nil
		},
		getSkill: func(_ context.Context, id string) (*gen.LearnSkill, error) {
			return &gen.LearnSkill{ID: id, UnitID: "99999999-9999-9999-9999-999999999999"}, nil
		},
	}
	progress := &mockProgressRepo{
		listLessons: func(_ context.Context, ownerType, ownerID string) ([]gen.LearnUserLessonProgress, error) {
			return []gen.LearnUserLessonProgress{
				{OwnerType: ownerType, OwnerID: ownerID, LessonID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1", Status: "completed"},
				{OwnerType: ownerType, OwnerID: ownerID, LessonID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa2", Status: "completed"},
			}, nil
		},
		getLesson: func(_ context.Context, _, _, _ string) (*gen.LearnUserLessonProgress, error) {
			return nil, pgx.ErrNoRows
		},
		upsertLesson: func(_ context.Context, _, _, _, _ string, _ int32) error {
			return nil
		},
	}
	attempts := &mockAttemptRepo{
		createAttempt: func(_ context.Context, ownerType, ownerID, lid string) (*gen.LearnAttempt, error) {
			return &gen.LearnAttempt{ID: attemptID, OwnerType: ownerType, OwnerID: ownerID, LessonID: lid, Status: "active"}, nil
		},
	}

	uc := newAttemptTestUsecase(curriculum, progress, attempts, nil)
	got, err := uc.StartLesson(context.Background(), "guest", guestID, lessonID, testTrialUnitID)
	if err != nil {
		t.Fatalf("StartLesson: %v", err)
	}
	if got != uuid.MustParse(attemptID) {
		t.Fatalf("attempt id = %v", got)
	}
}

func TestCompleteLesson_GuestSoftGate_BlocksWhenThreeCompleted(t *testing.T) {
	lessonID := "33333333-3333-3333-3333-333333333399"
	skillID := "22222222-2222-2222-2222-222222222299"
	guestID := "guest-1"

	curriculum := &mockCurriculumRepo{
		getLesson: func(_ context.Context, id string) (*gen.LearnLesson, error) {
			return &gen.LearnLesson{ID: id, SkillID: skillID, XpReward: 10}, nil
		},
		getSkill: func(_ context.Context, id string) (*gen.LearnSkill, error) {
			return &gen.LearnSkill{ID: id, UnitID: "99999999-9999-9999-9999-999999999999"}, nil
		},
	}
	progress := &mockProgressRepo{
		listLessons: func(_ context.Context, ownerType, ownerID string) ([]gen.LearnUserLessonProgress, error) {
			return threeCompletedLessons(ownerType, ownerID), nil
		},
		getLesson: func(_ context.Context, _, _, _ string) (*gen.LearnUserLessonProgress, error) {
			return nil, pgx.ErrNoRows
		},
	}

	uc := newAttemptTestUsecase(curriculum, progress, &mockAttemptRepo{}, nil)
	_, _, err := uc.CompleteLesson(context.Background(), "guest", guestID, lessonID, testTrialUnitID)
	if !errors.Is(err, ErrGuestSoftGate) {
		t.Fatalf("expected ErrGuestSoftGate, got %v", err)
	}
}
