package biz

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/puchidemy/puchi-backend/app/learn/internal/data/sqlc/gen"
)

var (
	ErrGuestNotFound       = errors.New("guest not found")
	ErrGuestAlreadyClaimed = errors.New("guest already claimed")
)

// GuestRepoInterface persists guest session rows.
type GuestRepoInterface interface {
	CreateGuest(ctx context.Context, id string) error
	GetGuestByID(ctx context.Context, id string) (*gen.LearnGuest, error)
	ClaimGuest(ctx context.Context, guestID, userID string) error
}

// ProgressRepoInterface persists lesson/unit progress and guest reassignment.
type ProgressRepoInterface interface {
	ListLessonProgressByOwner(ctx context.Context, ownerType, ownerID string) ([]gen.LearnUserLessonProgress, error)
	ListUnitProgressByOwner(ctx context.Context, ownerType, ownerID string) ([]gen.LearnUserUnitProgress, error)
	GetLessonProgress(ctx context.Context, ownerType, ownerID, lessonID string) (*gen.LearnUserLessonProgress, error)
	GetUnitProgress(ctx context.Context, ownerType, ownerID, unitID string) (*gen.LearnUserUnitProgress, error)
	UpsertLessonProgress(ctx context.Context, ownerType, ownerID, lessonID, status string, xp int32) error
	UpsertUnitProgress(ctx context.Context, ownerType, ownerID, unitID, status string) error
	DeleteGuestLessonProgress(ctx context.Context, guestID, lessonID string) error
	DeleteGuestUnitProgress(ctx context.Context, guestID, unitID string) error
	ReassignGuestLessonProgress(ctx context.Context, guestID, userID string) error
	ReassignGuestUnitProgress(ctx context.Context, guestID, userID string) error
	ReassignGuestAttempts(ctx context.Context, guestID, userID string) error
}

var statusRank = map[string]int{
	"not_started": 0,
	"in_progress": 1,
	"completed":   2,
}

func higherStatus(a, b string) string {
	if statusRank[a] >= statusRank[b] {
		return a
	}
	return b
}

func maxInt32(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

// CreateGuestSession inserts a new guest row and returns its ID.
func (uc *LearnUsecase) CreateGuestSession(ctx context.Context) (uuid.UUID, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, fmt.Errorf("generate guest id: %w", err)
	}
	if err := uc.guestRepo.CreateGuest(ctx, id.String()); err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

// ClaimGuest merges guest progress into the authenticated user and marks the guest claimed.
func (uc *LearnUsecase) ClaimGuest(ctx context.Context, userID string, guestID uuid.UUID) (int32, error) {
	guest, err := uc.guestRepo.GetGuestByID(ctx, guestID.String())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrGuestNotFound
		}
		return 0, err
	}
	if guest.ClaimedAt.Valid {
		return 0, ErrGuestAlreadyClaimed
	}

	guestIDStr := guestID.String()
	lessonsMerged := int32(0)

	guestLessons, err := uc.progressRepo.ListLessonProgressByOwner(ctx, "guest", guestIDStr)
	if err != nil {
		return 0, err
	}
	for _, gp := range guestLessons {
		userLesson, err := uc.progressRepo.GetLessonProgress(ctx, "user", userID, gp.LessonID)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return 0, err
			}
			continue
		}
		mergedStatus := higherStatus(gp.Status, userLesson.Status)
		mergedXP := maxInt32(gp.XpEarned, userLesson.XpEarned)
		if err := uc.progressRepo.UpsertLessonProgress(ctx, "user", userID, gp.LessonID, mergedStatus, mergedXP); err != nil {
			return 0, err
		}
		if err := uc.progressRepo.DeleteGuestLessonProgress(ctx, guestIDStr, gp.LessonID); err != nil {
			return 0, err
		}
		lessonsMerged++
	}

	guestUnits, err := uc.progressRepo.ListUnitProgressByOwner(ctx, "guest", guestIDStr)
	if err != nil {
		return 0, err
	}
	for _, gp := range guestUnits {
		userUnit, err := uc.progressRepo.GetUnitProgress(ctx, "user", userID, gp.UnitID)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return 0, err
			}
			continue
		}
		mergedStatus := higherStatus(gp.Status, userUnit.Status)
		if err := uc.progressRepo.UpsertUnitProgress(ctx, "user", userID, gp.UnitID, mergedStatus); err != nil {
			return 0, err
		}
		if err := uc.progressRepo.DeleteGuestUnitProgress(ctx, guestIDStr, gp.UnitID); err != nil {
			return 0, err
		}
	}

	if err := uc.progressRepo.ReassignGuestLessonProgress(ctx, guestIDStr, userID); err != nil {
		return 0, err
	}
	if err := uc.progressRepo.ReassignGuestUnitProgress(ctx, guestIDStr, userID); err != nil {
		return 0, err
	}
	if err := uc.progressRepo.ReassignGuestAttempts(ctx, guestIDStr, userID); err != nil {
		return 0, err
	}
	if err := uc.guestRepo.ClaimGuest(ctx, guestIDStr, userID); err != nil {
		return 0, err
	}
	return lessonsMerged, nil
}
