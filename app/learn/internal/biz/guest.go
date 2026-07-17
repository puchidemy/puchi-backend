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
	GetGuestByIDForUpdate(ctx context.Context, id string) (*gen.LearnGuest, error)
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
	guestIDStr := guestID.String()
	var lessonsMerged int32

	err := uc.tx.InTx(ctx, func(guestRepo GuestRepoInterface, progressRepo ProgressRepoInterface) error {
		guest, err := guestRepo.GetGuestByIDForUpdate(ctx, guestIDStr)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrGuestNotFound
			}
			return err
		}
		if guest.ClaimedAt.Valid {
			return ErrGuestAlreadyClaimed
		}

		guestLessons, err := progressRepo.ListLessonProgressByOwner(ctx, "guest", guestIDStr)
		if err != nil {
			return err
		}
		for _, gp := range guestLessons {
			userLesson, err := progressRepo.GetLessonProgress(ctx, "user", userID, gp.LessonID)
			if err != nil {
				if !errors.Is(err, pgx.ErrNoRows) {
					return err
				}
				continue
			}
			mergedStatus := higherStatus(gp.Status, userLesson.Status)
			mergedXP := maxInt32(gp.XpEarned, userLesson.XpEarned)
			if err := progressRepo.UpsertLessonProgress(ctx, "user", userID, gp.LessonID, mergedStatus, mergedXP); err != nil {
				return err
			}
			if err := progressRepo.DeleteGuestLessonProgress(ctx, guestIDStr, gp.LessonID); err != nil {
				return err
			}
			lessonsMerged++
		}

		guestUnits, err := progressRepo.ListUnitProgressByOwner(ctx, "guest", guestIDStr)
		if err != nil {
			return err
		}
		for _, gp := range guestUnits {
			userUnit, err := progressRepo.GetUnitProgress(ctx, "user", userID, gp.UnitID)
			if err != nil {
				if !errors.Is(err, pgx.ErrNoRows) {
					return err
				}
				continue
			}
			mergedStatus := higherStatus(gp.Status, userUnit.Status)
			if err := progressRepo.UpsertUnitProgress(ctx, "user", userID, gp.UnitID, mergedStatus); err != nil {
				return err
			}
			if err := progressRepo.DeleteGuestUnitProgress(ctx, guestIDStr, gp.UnitID); err != nil {
				return err
			}
		}

		if err := progressRepo.ReassignGuestLessonProgress(ctx, guestIDStr, userID); err != nil {
			return err
		}
		if err := progressRepo.ReassignGuestUnitProgress(ctx, guestIDStr, userID); err != nil {
			return err
		}
		if err := progressRepo.ReassignGuestAttempts(ctx, guestIDStr, userID); err != nil {
			return err
		}
		return guestRepo.ClaimGuest(ctx, guestIDStr, userID)
	})
	if err != nil {
		return 0, err
	}
	return lessonsMerged, nil
}
