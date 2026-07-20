package biz

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/puchidemy/puchi-backend/app/learn/internal/data/sqlc/gen"
)

var (
	ErrAttemptNotFound   = errors.New("attempt not found")
	ErrAttemptForbidden  = errors.New("attempt forbidden")
	ErrAttemptNotActive  = errors.New("attempt not active")
	ErrExerciseNotFound  = errors.New("exercise not found")
	ErrExerciseForbidden = errors.New("exercise forbidden")
)

// AttemptRepoInterface persists lesson/activity attempts and answers.
type AttemptRepoInterface interface {
	CreateAttempt(ctx context.Context, ownerType, ownerID, lessonID string) (*gen.LearnAttempt, error)
	GetAttemptByID(ctx context.Context, id string) (*gen.LearnAttempt, error)
	GetActiveAttemptByOwnerLesson(ctx context.Context, ownerType, ownerID, lessonID string) (*gen.LearnAttempt, error)
	InsertAttemptAnswer(ctx context.Context, attemptID, exerciseID string, payload json.RawMessage, correct bool) error
	CompleteAttempt(ctx context.Context, attemptID string, sessionXP int32) error
	ListAttemptAnswersByAttemptID(ctx context.Context, attemptID string) ([]gen.LearnAttemptAnswer, error)

	CreateActivityAttempt(ctx context.Context, ownerType, ownerID, storyID, sceneID string) (*gen.LearnActivityAttempt, error)
	GetActivityAttemptByID(ctx context.Context, id string) (*gen.LearnActivityAttempt, error)
	GetActiveActivityAttemptByOwnerScene(ctx context.Context, ownerType, ownerID, sceneID string) (*gen.LearnActivityAttempt, error)
	InsertActivityAttemptAnswer(ctx context.Context, attemptID, activityID string, payload json.RawMessage, correct bool) error
	CompleteActivityAttempt(ctx context.Context, attemptID string, sessionXP int32) error
	ListActivityAttemptAnswersByAttemptID(ctx context.Context, attemptID string) ([]gen.LearnActivityAttemptAnswer, error)
}

func (uc *LearnUsecase) loadLessonSkill(ctx context.Context, lessonID string) (*gen.LearnLesson, *gen.LearnSkill, error) {
	lesson, err := uc.curriculumRepo.GetLessonByID(ctx, lessonID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, ErrCurriculumNotFound
		}
		return nil, nil, err
	}
	skill, err := uc.curriculumRepo.GetSkillByID(ctx, lesson.SkillID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, ErrCurriculumNotFound
		}
		return nil, nil, err
	}
	return lesson, skill, nil
}

func (uc *LearnUsecase) assertAttemptOwner(attempt *gen.LearnAttempt, ownerType, ownerID string) error {
	if attempt.OwnerType != ownerType || attempt.OwnerID != ownerID {
		return ErrAttemptForbidden
	}
	return nil
}

// StartLesson creates an attempt and marks lesson progress in_progress.
func (uc *LearnUsecase) StartLesson(ctx context.Context, ownerType, ownerID, lessonID, trialUnitID string) (uuid.UUID, error) {
	if _, _, err := uc.loadLessonSkill(ctx, lessonID); err != nil {
		return uuid.Nil, err
	}
	if err := uc.assertGuestSoftGate(ctx, ownerType, ownerID, lessonID, trialUnitID); err != nil {
		return uuid.Nil, err
	}

	attempt, err := uc.attemptRepo.CreateAttempt(ctx, ownerType, ownerID, lessonID)
	if err != nil {
		return uuid.Nil, err
	}

	if err := uc.progressRepo.UpsertLessonProgress(ctx, ownerType, ownerID, lessonID, "in_progress", 0); err != nil {
		return uuid.Nil, err
	}

	id, err := uuid.Parse(attempt.ID)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

// SubmitAnswer grades a payload and stores the attempt answer.
// Soft-gate is not applied here — guests may finish an in-flight attempt.
func (uc *LearnUsecase) SubmitAnswer(ctx context.Context, ownerType, ownerID, attemptID, exerciseID string, payload json.RawMessage, _ string) (bool, error) {
	attempt, err := uc.attemptRepo.GetAttemptByID(ctx, attemptID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, ErrAttemptNotFound
		}
		return false, err
	}
	if err := uc.assertAttemptOwner(attempt, ownerType, ownerID); err != nil {
		return false, err
	}
	if attempt.Status != "active" {
		return false, ErrAttemptNotActive
	}

	if _, _, err := uc.loadLessonSkill(ctx, attempt.LessonID); err != nil {
		return false, err
	}

	exercise, err := uc.curriculumRepo.GetExerciseByID(ctx, exerciseID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, ErrExerciseNotFound
		}
		return false, err
	}
	if exercise.LessonID != attempt.LessonID {
		return false, ErrExerciseForbidden
	}

	correct, err := Grade(exercise.Type, exercise.Prompt, exercise.Answer, payload)
	if err != nil {
		return false, err
	}
	if err := uc.attemptRepo.InsertAttemptAnswer(ctx, attemptID, exerciseID, payload, correct); err != nil {
		return false, err
	}
	return correct, nil
}

// CompleteLesson finalizes the active attempt, updates progress, and publishes for users.
func (uc *LearnUsecase) CompleteLesson(ctx context.Context, ownerType, ownerID, lessonID, trialUnitID string) (int32, bool, error) {
	lesson, skill, err := uc.loadLessonSkill(ctx, lessonID)
	if err != nil {
		return 0, false, err
	}
	if err := uc.assertGuestSoftGate(ctx, ownerType, ownerID, lessonID, trialUnitID); err != nil {
		return 0, false, err
	}

	attempt, err := uc.attemptRepo.GetActiveAttemptByOwnerLesson(ctx, ownerType, ownerID, lessonID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, false, ErrAttemptNotFound
		}
		return 0, false, err
	}

	answers, err := uc.attemptRepo.ListAttemptAnswersByAttemptID(ctx, attempt.ID)
	if err != nil {
		return 0, false, err
	}
	exercises, err := uc.curriculumRepo.ListExercisesByLessonID(ctx, lessonID)
	if err != nil {
		return 0, false, err
	}

	sessionXP := computeSessionXP(lesson.XpReward, len(exercises), countCorrect(answers))
	if err := uc.attemptRepo.CompleteAttempt(ctx, attempt.ID, sessionXP); err != nil {
		return 0, false, err
	}

	prev, _ := uc.progressRepo.GetLessonProgress(ctx, ownerType, ownerID, lessonID)
	prevXP := int32(0)
	if prev != nil {
		prevXP = prev.XpEarned
	}
	mergedXP := maxInt32(prevXP, sessionXP)
	if err := uc.progressRepo.UpsertLessonProgress(ctx, ownerType, ownerID, lessonID, "completed", mergedXP); err != nil {
		return 0, false, err
	}

	unitCompleted, err := uc.markUnitProgressIfComplete(ctx, ownerType, ownerID, skill.UnitID)
	if err != nil {
		return 0, false, err
	}

	if ownerType == "user" {
		completedAt := time.Now().UTC()
		if err := uc.publisher.PublishLessonCompleted(ctx, LessonCompletedEvent{
			UserID:      ownerID,
			LessonID:    lessonID,
			UnitID:      skill.UnitID,
			XP:          sessionXP,
			CompletedAt: completedAt,
		}); err != nil {
			return 0, false, err
		}
		if unitCompleted {
			if err := uc.publisher.PublishUnitCompleted(ctx, UnitCompletedEvent{
				UserID:      ownerID,
				UnitID:      skill.UnitID,
				XP:          sessionXP,
				CompletedAt: completedAt,
			}); err != nil {
				return 0, false, err
			}
		}
	}

	return sessionXP, unitCompleted, nil
}

func countCorrect(answers []gen.LearnAttemptAnswer) int {
	n := 0
	for _, a := range answers {
		if a.Correct {
			n++
		}
	}
	return n
}

func computeSessionXP(reward int32, totalExercises, correctCount int) int32 {
	if totalExercises == 0 {
		return 0
	}
	if correctCount >= totalExercises {
		return reward
	}
	return int32(int64(reward) * int64(correctCount) / int64(totalExercises))
}

func (uc *LearnUsecase) markUnitProgressIfComplete(ctx context.Context, ownerType, ownerID, unitID string) (bool, error) {
	skills, err := uc.curriculumRepo.ListSkillsByUnitID(ctx, unitID)
	if err != nil {
		return false, err
	}

	requiredLessons := 0
	completedRequired := 0
	for _, skill := range skills {
		lessons, err := uc.curriculumRepo.ListLessonsBySkillID(ctx, skill.ID)
		if err != nil {
			return false, err
		}
		for _, lesson := range lessons {
			if !lesson.Required {
				continue
			}
			requiredLessons++
			progress, err := uc.progressRepo.GetLessonProgress(ctx, ownerType, ownerID, lesson.ID)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					continue
				}
				return false, err
			}
			if progress.Status == "completed" {
				completedRequired++
			}
		}
	}

	if requiredLessons == 0 || completedRequired < requiredLessons {
		return false, nil
	}

	if err := uc.progressRepo.UpsertUnitProgress(ctx, ownerType, ownerID, unitID, "completed"); err != nil {
		return false, err
	}
	return true, nil
}
