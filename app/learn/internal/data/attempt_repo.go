package data

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/puchidemy/puchi-backend/app/learn/internal/data/sqlc/gen"
)

// AttemptRepo wraps sqlc attempt queries.
type AttemptRepo struct {
	q *gen.Queries
}

// NewAttemptRepo creates an AttemptRepo.
func NewAttemptRepo(pool *pgxpool.Pool) *AttemptRepo {
	return &AttemptRepo{q: gen.New(pool)}
}

// CreateAttempt inserts a new active attempt.
func (r *AttemptRepo) CreateAttempt(ctx context.Context, ownerType, ownerID, lessonID string) (*gen.LearnAttempt, error) {
	row, err := r.q.CreateAttempt(ctx, gen.CreateAttemptParams{
		OwnerType: ownerType,
		OwnerID:   ownerID,
		LessonID:  lessonID,
	})
	return mapNoRows(row, err)
}

// GetAttemptByID returns an attempt by ID.
func (r *AttemptRepo) GetAttemptByID(ctx context.Context, id string) (*gen.LearnAttempt, error) {
	row, err := r.q.GetAttemptByID(ctx, id)
	return mapNoRows(row, err)
}

// GetActiveAttemptByOwnerLesson returns the latest active attempt for an owner and lesson.
func (r *AttemptRepo) GetActiveAttemptByOwnerLesson(ctx context.Context, ownerType, ownerID, lessonID string) (*gen.LearnAttempt, error) {
	row, err := r.q.GetActiveAttemptByOwnerLesson(ctx, gen.GetActiveAttemptByOwnerLessonParams{
		OwnerType: ownerType,
		OwnerID:   ownerID,
		LessonID:  lessonID,
	})
	return mapNoRows(row, err)
}

// InsertAttemptAnswer records a graded answer for an attempt.
func (r *AttemptRepo) InsertAttemptAnswer(ctx context.Context, attemptID, exerciseID string, payload json.RawMessage, correct bool) error {
	_, err := r.q.InsertAttemptAnswer(ctx, gen.InsertAttemptAnswerParams{
		AttemptID:  attemptID,
		ExerciseID: exerciseID,
		Payload:    payload,
		Correct:    correct,
	})
	return err
}

// CompleteAttempt marks an attempt completed with session XP.
func (r *AttemptRepo) CompleteAttempt(ctx context.Context, attemptID string, sessionXP int32) error {
	return r.q.CompleteAttempt(ctx, gen.CompleteAttemptParams{
		ID:        attemptID,
		SessionXp: sessionXP,
	})
}

// ListAttemptAnswersByAttemptID lists answers for an attempt.
func (r *AttemptRepo) ListAttemptAnswersByAttemptID(ctx context.Context, attemptID string) ([]gen.LearnAttemptAnswer, error) {
	return r.q.ListAttemptAnswersByAttemptID(ctx, attemptID)
}

// CreateActivityAttempt inserts a new active activity attempt for a scene.
func (r *AttemptRepo) CreateActivityAttempt(ctx context.Context, ownerType, ownerID, storyID, sceneID string) (*gen.LearnActivityAttempt, error) {
	row, err := r.q.CreateActivityAttempt(ctx, gen.CreateActivityAttemptParams{
		OwnerType: ownerType,
		OwnerID:   ownerID,
		StoryID:   storyID,
		SceneID:   sceneID,
	})
	return mapNoRows(row, err)
}

// GetActivityAttemptByID returns an activity attempt by ID.
func (r *AttemptRepo) GetActivityAttemptByID(ctx context.Context, id string) (*gen.LearnActivityAttempt, error) {
	row, err := r.q.GetActivityAttemptByID(ctx, id)
	return mapNoRows(row, err)
}

// GetActiveActivityAttemptByOwnerScene returns the latest active activity attempt.
func (r *AttemptRepo) GetActiveActivityAttemptByOwnerScene(ctx context.Context, ownerType, ownerID, sceneID string) (*gen.LearnActivityAttempt, error) {
	row, err := r.q.GetActiveActivityAttemptByOwnerScene(ctx, gen.GetActiveActivityAttemptByOwnerSceneParams{
		OwnerType: ownerType,
		OwnerID:   ownerID,
		SceneID:   sceneID,
	})
	return mapNoRows(row, err)
}

// InsertActivityAttemptAnswer records a graded activity answer.
func (r *AttemptRepo) InsertActivityAttemptAnswer(ctx context.Context, attemptID, activityID string, payload json.RawMessage, correct bool) error {
	_, err := r.q.InsertActivityAttemptAnswer(ctx, gen.InsertActivityAttemptAnswerParams{
		AttemptID:  attemptID,
		ActivityID: activityID,
		Payload:    payload,
		Correct:    correct,
	})
	return err
}

// CompleteActivityAttempt marks an activity attempt completed with session XP.
func (r *AttemptRepo) CompleteActivityAttempt(ctx context.Context, attemptID string, sessionXP int32) error {
	return r.q.CompleteActivityAttempt(ctx, gen.CompleteActivityAttemptParams{
		ID:        attemptID,
		SessionXp: sessionXP,
	})
}

// ListActivityAttemptAnswersByAttemptID lists answers for an activity attempt.
func (r *AttemptRepo) ListActivityAttemptAnswersByAttemptID(ctx context.Context, attemptID string) ([]gen.LearnActivityAttemptAnswer, error) {
	return r.q.ListActivityAttemptAnswersByAttemptID(ctx, attemptID)
}
