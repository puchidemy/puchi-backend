package biz

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/puchidemy/puchi-backend/app/core/internal/data/sqlc/gen"
)

// StatsRepoInterface defines the repo contract for stats (dependency inversion).
type StatsRepoInterface interface {
	GetUserStats(ctx context.Context, userID string) (*gen.CoreUserStat, error)
	UpsertStats(ctx context.Context, userID string) error
	UpdateStats(ctx context.Context, arg gen.UpdateUserStatsParams) (*gen.CoreUserStat, error)
	GetLevelThreshold(ctx context.Context, level int32) (int32, error)
	GetNextLevelThreshold(ctx context.Context, level int32) (int32, error)
	ClaimLearnEvent(ctx context.Context, eventType, userID, sourceID string, xp int32) (bool, error)
	GetDailyActivity(ctx context.Context, userID string, activityDate pgtype.Date) (*gen.CoreDailyActivity, error)
	GetLatestActivityDateBefore(ctx context.Context, userID string, before pgtype.Date) (pgtype.Date, error)
	UpsertDailyActivity(ctx context.Context, userID string, activityDate pgtype.Date, xp int32) error
	UpsertWeeklyXP(ctx context.Context, userID string, weekStart pgtype.Date, xp int32) error
	ListDailyActivityRange(ctx context.Context, userID string, from, to pgtype.Date) ([]gen.CoreDailyActivity, error)
	ListWeeklyXPHistory(ctx context.Context, userID string, fromWeek pgtype.Date) ([]gen.CoreXpHistory, error)
}

// StatsUsecase handles gamification stats operations.
type StatsUsecase struct {
	repo StatsRepoInterface
	tx   StatsTxManagerInterface
}

// NewStatsUsecase creates a new StatsUsecase.
func NewStatsUsecase(repo StatsRepoInterface, tx StatsTxManagerInterface) *StatsUsecase {
	return &StatsUsecase{repo: repo, tx: tx}
}

// GetStats returns the user's gamification stats, creating a zero row if missing.
func (uc *StatsUsecase) GetStats(ctx context.Context, userID string) (*gen.CoreUserStat, error) {
	return ensureStats(ctx, uc.repo, userID)
}

// GetXPToNextLevel returns the XP required to reach the next level.
// It queries the level_thresholds table for the correct value.
// Falls back to the formula-based calculation if not found in DB.
func (uc *StatsUsecase) GetXPToNextLevel(ctx context.Context, level int32) int32 {
	xp, err := uc.repo.GetNextLevelThreshold(ctx, level+1)
	if err == nil {
		return xp
	}
	return int32(math.Ceil(float64(level) * 60 * 1.5))
}

// ListDailyActivity returns daily activity rows in [from, to] (inclusive, UTC dates).
// Empty from/to use defaults: last 90 days through today.
func (uc *StatsUsecase) ListDailyActivity(ctx context.Context, userID string, from, to time.Time) ([]gen.CoreDailyActivity, error) {
	now := dateOnlyUTC(time.Now().UTC())
	if to.IsZero() {
		to = now
	} else {
		to = dateOnlyUTC(to)
	}
	if from.IsZero() {
		from = to.AddDate(0, 0, -90)
	} else {
		from = dateOnlyUTC(from)
	}
	if from.After(to) {
		return nil, fmt.Errorf("from must be on or before to")
	}

	rows, err := uc.repo.ListDailyActivityRange(ctx, userID, toPgDate(from), toPgDate(to))
	if err != nil {
		return nil, fmt.Errorf("list daily activity: %w", err)
	}
	return rows, nil
}

// WeeklyXPItem is a weekly XP history entry for API responses.
type WeeklyXPItem struct {
	WeekStart time.Time
	WeekLabel string
	XP        int32
}

// ListWeeklyXP returns weekly XP history for the last N weeks (default 12).
// Weeks with no history are filled with xp=0 so charts stay continuous.
func (uc *StatsUsecase) ListWeeklyXP(ctx context.Context, userID string, weeks int) ([]WeeklyXPItem, error) {
	if weeks <= 0 {
		weeks = 12
	}
	if weeks > 52 {
		weeks = 52
	}

	now := time.Now().UTC()
	currentWeek := weekStartUTC(now)
	fromWeek := currentWeek.AddDate(0, 0, -7*(weeks-1))

	rows, err := uc.repo.ListWeeklyXPHistory(ctx, userID, toPgDate(fromWeek))
	if err != nil {
		return nil, fmt.Errorf("list weekly xp: %w", err)
	}

	byWeek := make(map[string]int32, len(rows))
	for _, row := range rows {
		if !row.WeekStart.Valid {
			continue
		}
		byWeek[row.WeekStart.Time.Format("2006-01-02")] = row.XpEarned
	}

	out := make([]WeeklyXPItem, 0, weeks)
	for i := 0; i < weeks; i++ {
		ws := fromWeek.AddDate(0, 0, 7*i)
		key := ws.Format("2006-01-02")
		out = append(out, WeeklyXPItem{
			WeekStart: ws,
			WeekLabel: formatWeekLabel(ws),
			XP:        byWeek[key],
		})
	}
	return out, nil
}

func formatWeekLabel(weekStart time.Time) string {
	weekEnd := weekStart.AddDate(0, 0, 6)
	if weekStart.Month() == weekEnd.Month() {
		return fmt.Sprintf("%s %d-%d", weekStart.Format("Jan"), weekStart.Day(), weekEnd.Day())
	}
	return fmt.Sprintf("%s %d-%s %d", weekStart.Format("Jan"), weekStart.Day(), weekEnd.Format("Jan"), weekEnd.Day())
}
