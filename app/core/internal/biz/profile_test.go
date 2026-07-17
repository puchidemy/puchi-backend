package biz

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/puchidemy/puchi-backend/app/core/internal/data/sqlc/gen"
)

type memUserRepo struct {
	mu    sync.Mutex
	users map[string]*gen.CoreUser
}

func newMemUserRepo() *memUserRepo {
	return &memUserRepo{users: make(map[string]*gen.CoreUser)}
}

func (m *memUserRepo) CreateUser(_ context.Context, id, username, email, firstName, lastName string) (*gen.CoreUser, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u := &gen.CoreUser{
		ID:        id,
		Username:  username,
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	m.users[id] = u
	cp := *u
	return &cp, nil
}

func (m *memUserRepo) GetUser(_ context.Context, id string) (*gen.CoreUser, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[id]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	cp := *u
	return &cp, nil
}

func (m *memUserRepo) GetUserByEmail(_ context.Context, email string) (*gen.CoreUser, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, u := range m.users {
		if u.Email == email {
			cp := *u
			return &cp, nil
		}
	}
	return nil, pgx.ErrNoRows
}

func (m *memUserRepo) GetUserByUsername(_ context.Context, username string) (*gen.CoreUser, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, u := range m.users {
		if u.Username == username {
			cp := *u
			return &cp, nil
		}
	}
	return nil, pgx.ErrNoRows
}

func (m *memUserRepo) UpdateUser(_ context.Context, id, firstName, lastName, username string, bio, avatarKey *string, ageRange string) (*gen.CoreUser, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[id]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	u.FirstName = firstName
	u.LastName = lastName
	u.Username = username
	u.AgeRange = ageRange
	if bio != nil {
		u.Bio = bio
	}
	if avatarKey != nil {
		u.AvatarKey = avatarKey
	}
	cp := *u
	return &cp, nil
}

func (m *memUserRepo) UpdateAvatarKey(_ context.Context, id, avatarKey string) (*gen.CoreUser, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[id]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	key := avatarKey
	u.AvatarKey = &key
	cp := *u
	return &cp, nil
}

func (m *memUserRepo) UpdateOnboardingInfo(_ context.Context, id, firstName, lastName, ageRange, username string) (*gen.CoreUser, error) {
	return nil, errors.New("not implemented")
}

func (m *memUserRepo) UpsertUserOnboarding(_ context.Context, _, _, _, _ string) error {
	return nil
}

func (m *memUserRepo) UsernameExists(_ context.Context, username string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, u := range m.users {
		if u.Username == username {
			return true, nil
		}
	}
	return false, nil
}

func TestUpdateAvatar_ValidatesPrefix(t *testing.T) {
	users := newMemUserRepo()
	stats := newMemStatsRepo()
	uc := NewProfileUsecase(users, stats)

	_, _ = users.CreateUser(context.Background(), "u1", "alice", "a@x.com", "", "")

	_, err := uc.UpdateAvatar(context.Background(), "u1", "lesson_image/u1/x.jpg")
	if !errors.Is(err, ErrInvalidAvatarKey) {
		t.Fatalf("want ErrInvalidAvatarKey, got %v", err)
	}

	user, err := uc.UpdateAvatar(context.Background(), "u1", "avatar/u1/abc.jpg")
	if err != nil {
		t.Fatalf("UpdateAvatar: %v", err)
	}
	if user.AvatarKey == nil || *user.AvatarKey != "avatar/u1/abc.jpg" {
		t.Fatalf("avatar_key = %v", user.AvatarKey)
	}
}

func TestGetOrCreateProfile_CreatesStats(t *testing.T) {
	users := newMemUserRepo()
	stats := newMemStatsRepo()
	uc := NewProfileUsecase(users, stats)

	user, err := uc.GetOrCreateProfile(context.Background(), "u-new", "newbie@puchi.io.vn")
	if err != nil {
		t.Fatalf("GetOrCreateProfile: %v", err)
	}
	if user.Username == "" {
		t.Fatal("expected generated username")
	}

	st, err := stats.GetUserStats(context.Background(), "u-new")
	if err != nil {
		t.Fatalf("stats missing after lazy create: %v", err)
	}
	if st.Level != 1 {
		t.Fatalf("level = %d, want 1", st.Level)
	}
}

func TestGetStats_CreatesZerosWhenMissing(t *testing.T) {
	repo := newMemStatsRepo()
	uc := NewStatsUsecase(repo, passthroughStatsTx{repo: repo})

	stats, err := uc.GetStats(context.Background(), "ghost")
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}
	if stats.TotalXp != 0 || stats.Level != 1 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
}

func TestListWeeklyXP_FillsMissingWeeks(t *testing.T) {
	repo := newMemStatsRepo()
	uc := NewStatsUsecase(repo, passthroughStatsTx{repo: repo})

	now := time.Now().UTC()
	ws := weekStartUTC(now)
	_ = repo.UpsertWeeklyXP(context.Background(), "u1", toPgDate(ws), 42)

	items, err := uc.ListWeeklyXP(context.Background(), "u1", 4)
	if err != nil {
		t.Fatalf("ListWeeklyXP: %v", err)
	}
	if len(items) != 4 {
		t.Fatalf("len = %d, want 4", len(items))
	}
	if items[len(items)-1].XP != 42 {
		t.Fatalf("current week xp = %d, want 42", items[len(items)-1].XP)
	}
	if items[0].WeekLabel == "" {
		t.Fatal("expected week label")
	}
}
