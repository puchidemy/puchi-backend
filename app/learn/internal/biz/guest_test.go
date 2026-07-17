package biz

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/puchidemy/puchi-backend/app/learn/internal/data/sqlc/gen"
)

type mockGuestRepo struct {
	createGuest       func(ctx context.Context, id string) error
	getGuest          func(ctx context.Context, id string) (*gen.LearnGuest, error)
	getGuestForUpdate func(ctx context.Context, id string) (*gen.LearnGuest, error)
	claimGuest        func(ctx context.Context, guestID, userID string) error
}

func (m *mockGuestRepo) CreateGuest(ctx context.Context, id string) error {
	return m.createGuest(ctx, id)
}

func (m *mockGuestRepo) GetGuestByID(ctx context.Context, id string) (*gen.LearnGuest, error) {
	return m.getGuest(ctx, id)
}

func (m *mockGuestRepo) GetGuestByIDForUpdate(ctx context.Context, id string) (*gen.LearnGuest, error) {
	if m.getGuestForUpdate != nil {
		return m.getGuestForUpdate(ctx, id)
	}
	return m.getGuest(ctx, id)
}

func (m *mockGuestRepo) ClaimGuest(ctx context.Context, guestID, userID string) error {
	return m.claimGuest(ctx, guestID, userID)
}

type mockTxManager struct {
	guest    GuestRepoInterface
	progress ProgressRepoInterface
}

func (m *mockTxManager) InTx(_ context.Context, fn func(GuestRepoInterface, ProgressRepoInterface) error) error {
	return fn(m.guest, m.progress)
}

func newTestUsecase(guest GuestRepoInterface, progress ProgressRepoInterface) *LearnUsecase {
	return NewLearnUsecase(guest, progress, &mockCurriculumRepo{}, &mockAttemptRepo{}, NoOpLessonEventPublisher{}, &mockTxManager{guest: guest, progress: progress})
}

type mockProgressRepo struct {
	listLessons            func(ctx context.Context, ownerType, ownerID string) ([]gen.LearnUserLessonProgress, error)
	listUnits              func(ctx context.Context, ownerType, ownerID string) ([]gen.LearnUserUnitProgress, error)
	getLesson              func(ctx context.Context, ownerType, ownerID, lessonID string) (*gen.LearnUserLessonProgress, error)
	getUnit                func(ctx context.Context, ownerType, ownerID, unitID string) (*gen.LearnUserUnitProgress, error)
	upsertLesson           func(ctx context.Context, ownerType, ownerID, lessonID, status string, xp int32) error
	upsertUnit             func(ctx context.Context, ownerType, ownerID, unitID, status string) error
	deleteGuestLesson      func(ctx context.Context, guestID, lessonID string) error
	deleteGuestUnit        func(ctx context.Context, guestID, unitID string) error
	reassignGuestLessons   func(ctx context.Context, guestID, userID string) error
	reassignGuestUnits     func(ctx context.Context, guestID, userID string) error
	reassignGuestAttempts  func(ctx context.Context, guestID, userID string) error
}

func (m *mockProgressRepo) ListLessonProgressByOwner(ctx context.Context, ownerType, ownerID string) ([]gen.LearnUserLessonProgress, error) {
	return m.listLessons(ctx, ownerType, ownerID)
}

func (m *mockProgressRepo) ListUnitProgressByOwner(ctx context.Context, ownerType, ownerID string) ([]gen.LearnUserUnitProgress, error) {
	return m.listUnits(ctx, ownerType, ownerID)
}

func (m *mockProgressRepo) GetLessonProgress(ctx context.Context, ownerType, ownerID, lessonID string) (*gen.LearnUserLessonProgress, error) {
	return m.getLesson(ctx, ownerType, ownerID, lessonID)
}

func (m *mockProgressRepo) GetUnitProgress(ctx context.Context, ownerType, ownerID, unitID string) (*gen.LearnUserUnitProgress, error) {
	return m.getUnit(ctx, ownerType, ownerID, unitID)
}

func (m *mockProgressRepo) UpsertLessonProgress(ctx context.Context, ownerType, ownerID, lessonID, status string, xp int32) error {
	return m.upsertLesson(ctx, ownerType, ownerID, lessonID, status, xp)
}

func (m *mockProgressRepo) UpsertUnitProgress(ctx context.Context, ownerType, ownerID, unitID, status string) error {
	return m.upsertUnit(ctx, ownerType, ownerID, unitID, status)
}

func (m *mockProgressRepo) DeleteGuestLessonProgress(ctx context.Context, guestID, lessonID string) error {
	return m.deleteGuestLesson(ctx, guestID, lessonID)
}

func (m *mockProgressRepo) DeleteGuestUnitProgress(ctx context.Context, guestID, unitID string) error {
	return m.deleteGuestUnit(ctx, guestID, unitID)
}

func (m *mockProgressRepo) ReassignGuestLessonProgress(ctx context.Context, guestID, userID string) error {
	return m.reassignGuestLessons(ctx, guestID, userID)
}

func (m *mockProgressRepo) ReassignGuestUnitProgress(ctx context.Context, guestID, userID string) error {
	return m.reassignGuestUnits(ctx, guestID, userID)
}

func (m *mockProgressRepo) ReassignGuestAttempts(ctx context.Context, guestID, userID string) error {
	return m.reassignGuestAttempts(ctx, guestID, userID)
}

func TestCreateGuestSession_InsertsGuest(t *testing.T) {
	var insertedID string
	uc := newTestUsecase(&mockGuestRepo{
		createGuest: func(_ context.Context, id string) error {
			insertedID = id
			return nil
		},
	}, &mockProgressRepo{})

	guestID, err := uc.CreateGuestSession(context.Background())
	if err != nil {
		t.Fatalf("CreateGuestSession: %v", err)
	}
	if guestID == uuid.Nil {
		t.Fatal("expected non-nil guest ID")
	}
	if insertedID != guestID.String() {
		t.Fatalf("CreateGuest called with %q, want %q", insertedID, guestID.String())
	}
}

func TestClaimGuest_MergesProgressAndMarksClaimed(t *testing.T) {
	guestID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	userID := "user-123"
	lessonID := "33333333-3333-3333-3333-333333333331"

	var upsertedStatus string
	var upsertedXP int32
	claimed := false

	uc := newTestUsecase(&mockGuestRepo{
		getGuest: func(_ context.Context, id string) (*gen.LearnGuest, error) {
			if id != guestID.String() {
				t.Fatalf("unexpected guest id %q", id)
			}
			return &gen.LearnGuest{ID: id}, nil
		},
		claimGuest: func(_ context.Context, gid, uid string) error {
			if gid != guestID.String() || uid != userID {
				t.Fatalf("ClaimGuest(%q, %q)", gid, uid)
			}
			claimed = true
			return nil
		},
	}, &mockProgressRepo{
		listLessons: func(_ context.Context, ownerType, ownerID string) ([]gen.LearnUserLessonProgress, error) {
			if ownerType != "guest" || ownerID != guestID.String() {
				t.Fatalf("list lessons owner %s/%s", ownerType, ownerID)
			}
			return []gen.LearnUserLessonProgress{{
				OwnerType: "guest",
				OwnerID:   guestID.String(),
				LessonID:  lessonID,
				Status:    "completed",
				XpEarned:  10,
			}}, nil
		},
		listUnits: func(_ context.Context, _, _ string) ([]gen.LearnUserUnitProgress, error) {
			return nil, nil
		},
		getLesson: func(_ context.Context, ownerType, ownerID, lid string) (*gen.LearnUserLessonProgress, error) {
			if ownerType != "user" || ownerID != userID || lid != lessonID {
				t.Fatalf("get lesson %s/%s/%s", ownerType, ownerID, lid)
			}
			return &gen.LearnUserLessonProgress{
				OwnerType: "user",
				OwnerID:   userID,
				LessonID:  lessonID,
				Status:    "in_progress",
				XpEarned:  5,
			}, nil
		},
		upsertLesson: func(_ context.Context, ownerType, ownerID, lid, status string, xp int32) error {
			if ownerType != "user" || ownerID != userID || lid != lessonID {
				t.Fatalf("upsert lesson %s/%s/%s", ownerType, ownerID, lid)
			}
			upsertedStatus = status
			upsertedXP = xp
			return nil
		},
		deleteGuestLesson: func(_ context.Context, gid, lid string) error {
			if gid != guestID.String() || lid != lessonID {
				t.Fatalf("delete guest lesson %s/%s", gid, lid)
			}
			return nil
		},
		reassignGuestLessons: func(_ context.Context, gid, uid string) error {
			if gid != guestID.String() || uid != userID {
				t.Fatalf("reassign lessons %s/%s", gid, uid)
			}
			return nil
		},
		reassignGuestUnits: func(_ context.Context, gid, uid string) error {
			if gid != guestID.String() || uid != userID {
				t.Fatalf("reassign units %s/%s", gid, uid)
			}
			return nil
		},
		reassignGuestAttempts: func(_ context.Context, gid, uid string) error {
			if gid != guestID.String() || uid != userID {
				t.Fatalf("reassign attempts %s/%s", gid, uid)
			}
			return nil
		},
	})

	merged, err := uc.ClaimGuest(context.Background(), userID, guestID)
	if err != nil {
		t.Fatalf("ClaimGuest: %v", err)
	}
	if merged != 1 {
		t.Fatalf("lessons merged = %d, want 1", merged)
	}
	if upsertedStatus != "completed" {
		t.Fatalf("merged status = %q, want completed", upsertedStatus)
	}
	if upsertedXP != 10 {
		t.Fatalf("merged xp = %d, want 10", upsertedXP)
	}
	if !claimed {
		t.Fatal("expected guest marked claimed")
	}
}

func TestClaimGuest_RejectsAlreadyClaimed(t *testing.T) {
	guestID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	uc := newTestUsecase(&mockGuestRepo{
		getGuest: func(_ context.Context, id string) (*gen.LearnGuest, error) {
			return &gen.LearnGuest{
				ID:        id,
				ClaimedAt: pgtype.Timestamptz{Valid: true},
			}, nil
		},
	}, &mockProgressRepo{})

	_, err := uc.ClaimGuest(context.Background(), "user-123", guestID)
	if !errors.Is(err, ErrGuestAlreadyClaimed) {
		t.Fatalf("expected ErrGuestAlreadyClaimed, got %v", err)
	}
}
