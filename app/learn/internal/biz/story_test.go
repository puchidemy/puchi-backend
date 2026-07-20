package biz

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/puchidemy/puchi-backend/app/learn/internal/data/sqlc/gen"
)

func TestGuestSceneSoftGate_BlocksFourthSceneStart(t *testing.T) {
	sceneID := "a3000000-0000-4000-8000-000000000004"
	storyID := "a2000000-0000-4000-8000-000000000001"
	guestID := "guest-story-1"

	storyRepo := &mockStoryRepoFlexible{
		getScene: func(_ context.Context, id string) (*gen.LearnScene, error) {
			return &gen.LearnScene{ID: id, StoryID: storyID, Position: 4}, nil
		},
		getStory: func(_ context.Context, id string) (*gen.LearnStory, error) {
			return &gen.LearnStory{ID: id, Status: "published"}, nil
		},
	}
	progress := &mockProgressRepo{
		getLesson: func(_ context.Context, _, _, _ string) (*gen.LearnUserLessonProgress, error) {
			return nil, pgx.ErrNoRows
		},
	}
	// Override scene progress helpers via embedding wrapper.
	progressCounted := &countingSceneProgress{
		mockProgressRepo: progress,
		completed:        3,
		sceneStatus:      map[string]string{},
	}

	uc := NewLearnUsecase(&mockGuestRepo{}, progressCounted, &mockCurriculumRepo{}, storyRepo, &mockAttemptRepo{
		createAttempt: func(_ context.Context, _, _, _ string) (*gen.LearnAttempt, error) {
			t.Fatal("legacy CreateAttempt must not be called")
			return nil, nil
		},
	}, NoOpLessonEventPublisher{}, &mockTxManager{guest: &mockGuestRepo{}, progress: progressCounted})

	_, err := uc.StartActivity(context.Background(), "guest", guestID, sceneID)
	if !errors.Is(err, ErrGuestSoftGate) {
		t.Fatalf("expected ErrGuestSoftGate, got %v", err)
	}
}

func TestGuestSceneSoftGate_AllowsReplayCompletedScene(t *testing.T) {
	sceneID := "a3000000-0000-4000-8000-000000000001"
	storyID := "a2000000-0000-4000-8000-000000000001"
	attemptID := "a5000000-0000-4000-8000-000000000001"

	storyRepo := &mockStoryRepoFlexible{
		getScene: func(_ context.Context, id string) (*gen.LearnScene, error) {
			return &gen.LearnScene{ID: id, StoryID: storyID, Position: 1}, nil
		},
		getStory: func(_ context.Context, id string) (*gen.LearnStory, error) {
			return &gen.LearnStory{ID: id, Status: "published"}, nil
		},
	}
	progress := &countingSceneProgress{
		mockProgressRepo: &mockProgressRepo{},
		completed:        3,
		sceneStatus:      map[string]string{sceneID: "completed"},
	}
	attempts := &mockAttemptRepoFlexible{
		getActive: func(_ context.Context, _, _, _ string) (*gen.LearnActivityAttempt, error) {
			return &gen.LearnActivityAttempt{ID: attemptID, Status: "active", SceneID: sceneID, StoryID: storyID}, nil
		},
	}

	uc := NewLearnUsecase(&mockGuestRepo{}, progress, &mockCurriculumRepo{}, storyRepo, attempts, NoOpLessonEventPublisher{}, &mockTxManager{
		guest:    &mockGuestRepo{},
		progress: progress,
	})

	id, err := uc.StartActivity(context.Background(), "guest", "guest-1", sceneID)
	if err != nil {
		t.Fatalf("StartActivity: %v", err)
	}
	if id.String() != attemptID {
		t.Fatalf("attempt id = %s, want %s", id, attemptID)
	}
}

func TestSubmitActivityAnswer_UsesGradingEngine(t *testing.T) {
	attemptID := "a5000000-0000-4000-8000-000000000002"
	activityID := "a4000000-0000-4000-8000-000000000001"
	sceneID := "a3000000-0000-4000-8000-000000000001"

	storyRepo := &mockStoryRepoFlexible{
		getActivity: func(_ context.Context, id string) (*gen.LearnActivity, error) {
			return &gen.LearnActivity{
				ID:      id,
				SceneID: sceneID,
				Type:    "select",
				Prompt:  json.RawMessage(`{"question":"q","options":["Phở","Pizza"]}`),
				Answer:  json.RawMessage(`{"correct":"Phở"}`),
			}, nil
		},
	}
	var inserted bool
	attempts := &mockAttemptRepoFlexible{
		getByID: func(_ context.Context, id string) (*gen.LearnActivityAttempt, error) {
			return &gen.LearnActivityAttempt{
				ID:        id,
				OwnerType: "user",
				OwnerID:   "u1",
				SceneID:   sceneID,
				Status:    "active",
			}, nil
		},
		insertAnswer: func(_ context.Context, _, _ string, _ json.RawMessage, correct bool) error {
			inserted = true
			if !correct {
				t.Fatal("expected correct=true")
			}
			return nil
		},
	}

	uc := NewLearnUsecase(&mockGuestRepo{}, &mockProgressRepo{}, &mockCurriculumRepo{}, storyRepo, attempts, NoOpLessonEventPublisher{}, &mockTxManager{
		guest:    &mockGuestRepo{},
		progress: &mockProgressRepo{},
	})

	ok, err := uc.SubmitActivityAnswer(context.Background(), "user", "u1", attemptID, activityID, json.RawMessage(`{"answer":"Phở"}`))
	if err != nil {
		t.Fatalf("SubmitActivityAnswer: %v", err)
	}
	if !ok || !inserted {
		t.Fatalf("ok=%v inserted=%v", ok, inserted)
	}
}

type mockStoryRepoFlexible struct {
	mockStoryRepo
	getScene    func(ctx context.Context, id string) (*gen.LearnScene, error)
	getStory    func(ctx context.Context, id string) (*gen.LearnStory, error)
	getActivity func(ctx context.Context, id string) (*gen.LearnActivity, error)
}

func (m *mockStoryRepoFlexible) GetSceneByID(ctx context.Context, id string) (*gen.LearnScene, error) {
	if m.getScene != nil {
		return m.getScene(ctx, id)
	}
	return m.mockStoryRepo.GetSceneByID(ctx, id)
}

func (m *mockStoryRepoFlexible) GetStoryByID(ctx context.Context, id string) (*gen.LearnStory, error) {
	if m.getStory != nil {
		return m.getStory(ctx, id)
	}
	return m.mockStoryRepo.GetStoryByID(ctx, id)
}

func (m *mockStoryRepoFlexible) GetActivityByID(ctx context.Context, id string) (*gen.LearnActivity, error) {
	if m.getActivity != nil {
		return m.getActivity(ctx, id)
	}
	return m.mockStoryRepo.GetActivityByID(ctx, id)
}

type countingSceneProgress struct {
	*mockProgressRepo
	completed   int32
	sceneStatus map[string]string
}

func (p *countingSceneProgress) CountCompletedScenesByOwner(context.Context, string, string) (int32, error) {
	return p.completed, nil
}

func (p *countingSceneProgress) GetSceneProgress(_ context.Context, _, _, sceneID string) (*gen.LearnUserSceneProgress, error) {
	if status, ok := p.sceneStatus[sceneID]; ok {
		return &gen.LearnUserSceneProgress{SceneID: sceneID, Status: status}, nil
	}
	return nil, pgx.ErrNoRows
}

func (p *countingSceneProgress) UpsertSceneProgress(context.Context, string, string, string, string) error {
	return nil
}

func (p *countingSceneProgress) UpsertStoryProgress(context.Context, string, string, string, string, int32) error {
	return nil
}

type mockAttemptRepoFlexible struct {
	mockAttemptRepo
	getActive    func(ctx context.Context, ownerType, ownerID, sceneID string) (*gen.LearnActivityAttempt, error)
	getByID      func(ctx context.Context, id string) (*gen.LearnActivityAttempt, error)
	insertAnswer func(ctx context.Context, attemptID, activityID string, payload json.RawMessage, correct bool) error
}

func (m *mockAttemptRepoFlexible) GetActiveActivityAttemptByOwnerScene(ctx context.Context, ownerType, ownerID, sceneID string) (*gen.LearnActivityAttempt, error) {
	if m.getActive != nil {
		return m.getActive(ctx, ownerType, ownerID, sceneID)
	}
	return m.mockAttemptRepo.GetActiveActivityAttemptByOwnerScene(ctx, ownerType, ownerID, sceneID)
}

func (m *mockAttemptRepoFlexible) GetActivityAttemptByID(ctx context.Context, id string) (*gen.LearnActivityAttempt, error) {
	if m.getByID != nil {
		return m.getByID(ctx, id)
	}
	return m.mockAttemptRepo.GetActivityAttemptByID(ctx, id)
}

func (m *mockAttemptRepoFlexible) InsertActivityAttemptAnswer(ctx context.Context, attemptID, activityID string, payload json.RawMessage, correct bool) error {
	if m.insertAnswer != nil {
		return m.insertAnswer(ctx, attemptID, activityID, payload, correct)
	}
	return m.mockAttemptRepo.InsertActivityAttemptAnswer(ctx, attemptID, activityID, payload, correct)
}
