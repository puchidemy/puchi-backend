package biz

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/puchidemy/puchi-backend/app/core/internal/data/sqlc/gen"
)

const (
	learnEventLesson = "lesson"
	learnEventUnit   = "unit"
)

// LessonCompletedEvent is consumed from learn.lesson.completed.
type LessonCompletedEvent struct {
	UserID      string
	LessonID    string
	UnitID      string
	XP          int32
	CompletedAt time.Time
}

// UnitCompletedEvent is consumed from learn.unit.completed.
type UnitCompletedEvent struct {
	UserID      string
	UnitID      string
	XP          int32
	CompletedAt time.Time
}

// OnLessonCompleted applies XP, daily activity, and streak idempotently per user+lesson.
func (uc *StatsUsecase) OnLessonCompleted(ctx context.Context, evt LessonCompletedEvent) error {
	if evt.UserID == "" || evt.LessonID == "" {
		return errors.New("invalid lesson completed event")
	}
	if evt.CompletedAt.IsZero() {
		evt.CompletedAt = time.Now().UTC()
	}

	claimed, err := uc.repo.ClaimLearnEvent(ctx, learnEventLesson, evt.UserID, evt.LessonID, evt.XP)
	if err != nil {
		return err
	}
	if !claimed {
		return nil
	}

	stats, err := uc.ensureStats(ctx, evt.UserID)
	if err != nil {
		return err
	}

	activityDate := toPgDate(evt.CompletedAt)
	hadToday, err := uc.hadActivityToday(ctx, evt.UserID, activityDate)
	if err != nil {
		return err
	}

	var prevDate *time.Time
	if !hadToday {
		pd, err := uc.repo.GetLatestActivityDateBefore(ctx, evt.UserID, activityDate)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		if err == nil && pd.Valid {
			t := pd.Time
			prevDate = &t
		}
	}

	newStreak := nextStreak(stats.CurrentStreak, evt.CompletedAt, hadToday, prevDate)
	longestStreak := stats.LongestStreak
	if newStreak > longestStreak {
		longestStreak = newStreak
	}

	if err := uc.repo.UpsertDailyActivity(ctx, evt.UserID, activityDate, evt.XP); err != nil {
		return err
	}
	if err := uc.repo.UpsertWeeklyXP(ctx, evt.UserID, toPgDate(weekStartUTC(evt.CompletedAt)), evt.XP); err != nil {
		return err
	}

	newTotalXP := stats.TotalXp + evt.XP
	level, currentXP := uc.calcLevelFromTotalXP(ctx, newTotalXP)

	_, err = uc.repo.UpdateStats(ctx, gen.UpdateUserStatsParams{
		UserID:           evt.UserID,
		CurrentXp:        currentXP,
		TotalXp:          newTotalXP,
		Level:            level,
		CurrentStreak:    newStreak,
		LongestStreak:    longestStreak,
		TotalLessons:     stats.TotalLessons + 1,
		CompletedLessons: stats.CompletedLessons + 1,
		TotalMinutes:     stats.TotalMinutes,
		Accuracy:         stats.Accuracy,
		WordsLearned:     stats.WordsLearned,
	})
	return err
}

// OnUnitCompleted applies bonus XP idempotently per user+unit.
func (uc *StatsUsecase) OnUnitCompleted(ctx context.Context, evt UnitCompletedEvent) error {
	if evt.UserID == "" || evt.UnitID == "" {
		return errors.New("invalid unit completed event")
	}
	if evt.CompletedAt.IsZero() {
		evt.CompletedAt = time.Now().UTC()
	}

	claimed, err := uc.repo.ClaimLearnEvent(ctx, learnEventUnit, evt.UserID, evt.UnitID, evt.XP)
	if err != nil {
		return err
	}
	if !claimed {
		return nil
	}
	if evt.XP == 0 {
		return nil
	}

	stats, err := uc.ensureStats(ctx, evt.UserID)
	if err != nil {
		return err
	}

	if err := uc.repo.UpsertWeeklyXP(ctx, evt.UserID, toPgDate(weekStartUTC(evt.CompletedAt)), evt.XP); err != nil {
		return err
	}

	newTotalXP := stats.TotalXp + evt.XP
	level, currentXP := uc.calcLevelFromTotalXP(ctx, newTotalXP)

	_, err = uc.repo.UpdateStats(ctx, gen.UpdateUserStatsParams{
		UserID:           evt.UserID,
		CurrentXp:        currentXP,
		TotalXp:          newTotalXP,
		Level:            level,
		CurrentStreak:    stats.CurrentStreak,
		LongestStreak:    stats.LongestStreak,
		TotalLessons:     stats.TotalLessons,
		CompletedLessons: stats.CompletedLessons,
		TotalMinutes:     stats.TotalMinutes,
		Accuracy:         stats.Accuracy,
		WordsLearned:     stats.WordsLearned,
	})
	return err
}

func (uc *StatsUsecase) ensureStats(ctx context.Context, userID string) (*gen.CoreUserStat, error) {
	stats, err := uc.repo.GetUserStats(ctx, userID)
	if err == nil {
		return stats, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	if err := uc.repo.UpsertStats(ctx, userID); err != nil {
		return nil, err
	}
	stats, err = uc.repo.GetUserStats(ctx, userID)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

func (uc *StatsUsecase) hadActivityToday(ctx context.Context, userID string, activityDate pgtype.Date) (bool, error) {
	row, err := uc.repo.GetDailyActivity(ctx, userID, activityDate)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return row.LessonsCompleted > 0, nil
}

func (uc *StatsUsecase) calcLevelFromTotalXP(ctx context.Context, totalXP int32) (level, currentXP int32) {
	level = 1
	baseXP := int32(0)
	for l := int32(1); l <= 10; l++ {
		req, err := uc.repo.GetLevelThreshold(ctx, l)
		if err != nil {
			break
		}
		if totalXP >= req {
			level = l
			baseXP = req
		}
	}
	return level, totalXP - baseXP
}

// nextStreak returns the streak after a lesson on activityDate.
func nextStreak(current int32, activityDate time.Time, hadActivityToday bool, lastActive *time.Time) int32 {
	if hadActivityToday {
		return current
	}
	today := dateOnlyUTC(activityDate)
	if lastActive == nil {
		return 1
	}
	last := dateOnlyUTC(*lastActive)
	diffDays := int(today.Sub(last).Hours() / 24)
	switch {
	case diffDays <= 0:
		return current
	case diffDays == 1:
		if current <= 0 {
			return 1
		}
		return current + 1
	default:
		return 1
	}
}

func dateOnlyUTC(t time.Time) time.Time {
	t = t.UTC()
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func weekStartUTC(t time.Time) time.Time {
	t = dateOnlyUTC(t)
	wd := int(t.Weekday())
	if wd == 0 {
		wd = 7
	}
	return t.AddDate(0, 0, -(wd - 1))
}

func toPgDate(t time.Time) pgtype.Date {
	t = dateOnlyUTC(t)
	return pgtype.Date{Time: t, Valid: true}
}
