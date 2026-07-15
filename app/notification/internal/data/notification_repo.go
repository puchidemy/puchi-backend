package data

import (
	"context"
	"fmt"
	"time"

	"github.com/puchidemy/puchi-backend/app/notification/internal/biz"
	"github.com/puchidemy/puchi-backend/app/notification/internal/data/sqlc/gen"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PreferenceRepo implements biz.PreferenceRepo using PostgreSQL.
type PreferenceRepo struct {
	pool *pgxpool.Pool
	q    *gen.Queries
}

// NewPreferenceRepo creates a new PreferenceRepo.
func NewPreferenceRepo(pool *pgxpool.Pool) *PreferenceRepo {
	return &PreferenceRepo{
		pool: pool,
		q:    gen.New(pool),
	}
}

// GetPreferences returns the notification preferences for a user.
func (r *PreferenceRepo) GetPreferences(ctx context.Context, userID string) (*biz.Preference, error) {
	row, err := r.q.GetPreferences(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get preferences: %w", err)
	}
	return preferenceFromGen(row), nil
}

// UpsertPreferences creates or updates notification preferences for a user.
func (r *PreferenceRepo) UpsertPreferences(ctx context.Context, prefs *biz.Preference) (*biz.Preference, error) {
	row, err := r.q.UpsertPreferences(ctx, gen.UpsertPreferencesParams{
		UserID:          prefs.UserID,
		PushEnabled:     prefs.PushEnabled,
		EmailEnabled:    prefs.EmailEnabled,
		StreakReminder:  prefs.StreakReminder,
		FriendActivity:  prefs.FriendActivity,
		Promotions:      prefs.Promotions,
		QuietHoursStart: stringToPgtTime(prefs.QuietHoursStart),
		QuietHoursEnd:   stringToPgtTime(prefs.QuietHoursEnd),
	})
	if err != nil {
		return nil, fmt.Errorf("upsert preferences: %w", err)
	}
	return preferenceFromGen(row), nil
}

func preferenceFromGen(row gen.NotificationPreference) *biz.Preference {
	return &biz.Preference{
		UserID:          row.UserID,
		PushEnabled:     row.PushEnabled,
		EmailEnabled:    row.EmailEnabled,
		StreakReminder:  row.StreakReminder,
		FriendActivity:  row.FriendActivity,
		Promotions:      row.Promotions,
		QuietHoursStart: pgtTimeToString(row.QuietHoursStart),
		QuietHoursEnd:   pgtTimeToString(row.QuietHoursEnd),
	}
}

func stringToPgtTime(s *string) pgtype.Time {
	if s == nil || *s == "" {
		return pgtype.Time{Valid: false}
	}
	t, err := time.Parse("15:04", *s)
	if err != nil {
		return pgtype.Time{Valid: false}
	}
	micros := int64(t.Hour())*3600000000 + int64(t.Minute())*60000000
	return pgtype.Time{
		Microseconds: micros,
		Valid:        true,
	}
}

func pgtTimeToString(t pgtype.Time) *string {
	if !t.Valid {
		return nil
	}
	hours := t.Microseconds / 3600000000
	minutes := (t.Microseconds % 3600000000) / 60000000
	s := fmt.Sprintf("%02d:%02d", hours, minutes)
	return &s
}
