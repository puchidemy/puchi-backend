package biz

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/puchidemy/puchi-backend/app/core/internal/data/sqlc/gen"
)

type memStatsRepo struct {
	mu       sync.Mutex
	stats    map[string]*gen.CoreUserStat
	claimed  map[string]struct{}
	daily    map[string]*gen.CoreDailyActivity
	weeklyXP map[string]int32
}

func newMemStatsRepo() *memStatsRepo {
	return &memStatsRepo{
		stats:    make(map[string]*gen.CoreUserStat),
		claimed:  make(map[string]struct{}),
		daily:    make(map[string]*gen.CoreDailyActivity),
		weeklyXP: make(map[string]int32),
	}
}

func (m *memStatsRepo) claimKey(eventType, userID, sourceID string) string {
	return eventType + ":" + userID + ":" + sourceID
}

func (m *memStatsRepo) dailyKey(userID string, d pgtype.Date) string {
	return userID + ":" + d.Time.Format("2006-01-02")
}

func (m *memStatsRepo) GetUserStats(_ context.Context, userID string) (*gen.CoreUserStat, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.stats[userID]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	cp := *s
	return &cp, nil
}

func (m *memStatsRepo) UpsertStats(_ context.Context, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.stats[userID]; !ok {
		m.stats[userID] = &gen.CoreUserStat{UserID: userID, Level: 1}
	}
	return nil
}

func (m *memStatsRepo) UpdateStats(_ context.Context, arg gen.UpdateUserStatsParams) (*gen.CoreUserStat, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s := m.stats[arg.UserID]
	if s == nil {
		return nil, pgx.ErrNoRows
	}
	s.CurrentXp = arg.CurrentXp
	s.TotalXp = arg.TotalXp
	s.Level = arg.Level
	s.CurrentStreak = arg.CurrentStreak
	s.LongestStreak = arg.LongestStreak
	s.TotalLessons = arg.TotalLessons
	s.CompletedLessons = arg.CompletedLessons
	cp := *s
	return &cp, nil
}

func (m *memStatsRepo) GetLevelThreshold(_ context.Context, level int32) (int32, error) {
	thresholds := map[int32]int32{
		1: 0, 2: 60, 3: 120, 4: 200, 5: 300,
		6: 450, 7: 650, 8: 900, 9: 1200, 10: 1500,
	}
	v, ok := thresholds[level]
	if !ok {
		return 0, errors.New("level not found")
	}
	return v, nil
}

func (m *memStatsRepo) GetNextLevelThreshold(ctx context.Context, level int32) (int32, error) {
	return m.GetLevelThreshold(ctx, level+1)
}

func (m *memStatsRepo) ClaimLearnEvent(_ context.Context, eventType, userID, sourceID string, _ int32) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := m.claimKey(eventType, userID, sourceID)
	if _, ok := m.claimed[key]; ok {
		return false, nil
	}
	m.claimed[key] = struct{}{}
	return true, nil
}

func (m *memStatsRepo) GetDailyActivity(_ context.Context, userID string, activityDate pgtype.Date) (*gen.CoreDailyActivity, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	row, ok := m.daily[m.dailyKey(userID, activityDate)]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	cp := *row
	return &cp, nil
}

func (m *memStatsRepo) GetLatestActivityDateBefore(_ context.Context, userID string, before pgtype.Date) (pgtype.Date, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var latest pgtype.Date
	found := false
	for key, row := range m.daily {
		if len(key) <= len(userID)+1 || key[:len(userID)] != userID {
			continue
		}
		if !row.ActivityDate.Valid || !before.Valid {
			continue
		}
		if row.LessonsCompleted <= 0 {
			continue
		}
		if row.ActivityDate.Time.Before(before.Time) {
			if !found || row.ActivityDate.Time.After(latest.Time) {
				latest = row.ActivityDate
				found = true
			}
		}
	}
	if !found {
		return pgtype.Date{}, pgx.ErrNoRows
	}
	return latest, nil
}

func (m *memStatsRepo) UpsertDailyActivity(_ context.Context, userID string, activityDate pgtype.Date, xp int32) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := m.dailyKey(userID, activityDate)
	row, ok := m.daily[key]
	if !ok {
		m.daily[key] = &gen.CoreDailyActivity{
			UserID:           userID,
			ActivityDate:     activityDate,
			LessonsCompleted: 1,
			XpEarned:         xp,
		}
		return nil
	}
	row.LessonsCompleted++
	row.XpEarned += xp
	return nil
}

func (m *memStatsRepo) UpsertWeeklyXP(_ context.Context, userID string, weekStart pgtype.Date, xp int32) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := userID + ":" + weekStart.Time.Format("2006-01-02")
	m.weeklyXP[key] += xp
	return nil
}

func TestOnLessonCompleted_XPAppliedOnce(t *testing.T) {
	repo := newMemStatsRepo()
	repo.stats["user-1"] = &gen.CoreUserStat{UserID: "user-1", Level: 1}
	uc := NewStatsUsecase(repo)

	evt := LessonCompletedEvent{
		UserID:      "user-1",
		LessonID:    "lesson-1",
		XP:          20,
		CompletedAt: time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC),
	}

	if err := uc.OnLessonCompleted(context.Background(), evt); err != nil {
		t.Fatalf("first apply: %v", err)
	}
	stats, err := uc.GetStats(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("get stats: %v", err)
	}
	if stats.TotalXp != 20 {
		t.Fatalf("total xp = %d, want 20", stats.TotalXp)
	}
	if stats.CompletedLessons != 1 {
		t.Fatalf("completed lessons = %d, want 1", stats.CompletedLessons)
	}
	if stats.CurrentStreak != 1 {
		t.Fatalf("streak = %d, want 1", stats.CurrentStreak)
	}

	if err := uc.OnLessonCompleted(context.Background(), evt); err != nil {
		t.Fatalf("duplicate apply: %v", err)
	}
	stats, err = uc.GetStats(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("get stats after duplicate: %v", err)
	}
	if stats.TotalXp != 20 {
		t.Fatalf("duplicate total xp = %d, want 20", stats.TotalXp)
	}
	if stats.CompletedLessons != 1 {
		t.Fatalf("duplicate completed lessons = %d, want 1", stats.CompletedLessons)
	}
}

func TestNextStreak(t *testing.T) {
	today := time.Date(2026, 7, 17, 15, 0, 0, 0, time.UTC)
	yesterday := today.AddDate(0, 0, -1)

	if got := nextStreak(3, today, true, nil); got != 3 {
		t.Fatalf("same day second lesson: got %d want 3", got)
	}
	if got := nextStreak(0, today, false, nil); got != 1 {
		t.Fatalf("first lesson: got %d want 1", got)
	}
	if got := nextStreak(5, today, false, &yesterday); got != 6 {
		t.Fatalf("consecutive day: got %d want 6", got)
	}
	gap := today.AddDate(0, 0, -3)
	if got := nextStreak(5, today, false, &gap); got != 1 {
		t.Fatalf("broken streak: got %d want 1", got)
	}
}
