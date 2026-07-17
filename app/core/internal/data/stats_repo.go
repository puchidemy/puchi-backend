package data

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/puchidemy/puchi-backend/app/core/internal/data/sqlc/gen"
)

// StatsRepo wraps sqlc-generated queries for core.user_stats.
type StatsRepo struct {
	q *gen.Queries
}

// NewStatsRepo creates a new StatsRepo.
func NewStatsRepo(pool *pgxpool.Pool) *StatsRepo {
	return &StatsRepo{q: gen.New(pool)}
}

// GetUserStats retrieves stats for a user.
func (r *StatsRepo) GetUserStats(ctx context.Context, userID string) (*gen.CoreUserStat, error) {
	row, err := r.q.GetUserStats(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// UpsertStats creates initial stats row for a user if not exists.
func (r *StatsRepo) UpsertStats(ctx context.Context, userID string) error {
	return r.q.CreateUserStats(ctx, userID)
}

// UpdateStats updates stats and returns the updated row.
func (r *StatsRepo) UpdateStats(ctx context.Context, arg gen.UpdateUserStatsParams) (*gen.CoreUserStat, error) {
	row, err := r.q.UpdateUserStats(ctx, arg)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// GetLevelThreshold returns the XP required for a given level.
func (r *StatsRepo) GetLevelThreshold(ctx context.Context, level int32) (int32, error) {
	return r.q.GetLevelThreshold(ctx, level)
}

// GetNextLevelThreshold returns the XP required for the next level.
func (r *StatsRepo) GetNextLevelThreshold(ctx context.Context, level int32) (int32, error) {
	return r.q.GetNextLevelThreshold(ctx, level)
}

// ClaimLearnEvent inserts an idempotency row; false when already processed.
func (r *StatsRepo) ClaimLearnEvent(ctx context.Context, eventType, userID, sourceID string, xp int32) (bool, error) {
	n, err := r.q.InsertProcessedLearnEvent(ctx, gen.InsertProcessedLearnEventParams{
		EventType: eventType,
		UserID:    userID,
		SourceID:  sourceID,
		XpApplied: xp,
	})
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// GetDailyActivity returns daily activity for a user on a date.
func (r *StatsRepo) GetDailyActivity(ctx context.Context, userID string, activityDate pgtype.Date) (*gen.CoreDailyActivity, error) {
	row, err := r.q.GetDailyActivity(ctx, gen.GetDailyActivityParams{
		UserID:       userID,
		ActivityDate: activityDate,
	})
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// GetLatestActivityDateBefore returns the most recent activity date before the given date.
func (r *StatsRepo) GetLatestActivityDateBefore(ctx context.Context, userID string, before pgtype.Date) (pgtype.Date, error) {
	return r.q.GetLatestActivityDateBefore(ctx, gen.GetLatestActivityDateBeforeParams{
		UserID:       userID,
		ActivityDate: before,
	})
}

// UpsertDailyActivity increments lessons/xp for the activity date.
func (r *StatsRepo) UpsertDailyActivity(ctx context.Context, userID string, activityDate pgtype.Date, xp int32) error {
	_, err := r.q.UpsertDailyActivity(ctx, gen.UpsertDailyActivityParams{
		UserID:       userID,
		ActivityDate: activityDate,
		XpEarned:     xp,
	})
	return err
}

// UpsertWeeklyXP adds XP to the user's weekly history bucket.
func (r *StatsRepo) UpsertWeeklyXP(ctx context.Context, userID string, weekStart pgtype.Date, xp int32) error {
	return r.q.UpsertWeeklyXP(ctx, gen.UpsertWeeklyXPParams{
		UserID:    userID,
		WeekStart: weekStart,
		XpEarned:  xp,
	})
}
