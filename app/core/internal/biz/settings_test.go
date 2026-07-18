package biz

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/puchidemy/puchi-backend/app/core/internal/data/sqlc/gen"
)

type memSettingsRepo struct {
	mu   sync.Mutex
	rows map[string]*gen.CoreUserSetting
}

func newMemSettingsRepo() *memSettingsRepo {
	return &memSettingsRepo{rows: make(map[string]*gen.CoreUserSetting)}
}

func (m *memSettingsRepo) Get(_ context.Context, userID string) (*gen.CoreUserSetting, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	row, ok := m.rows[userID]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	cp := *row
	return &cp, nil
}

func (m *memSettingsRepo) Upsert(_ context.Context, params gen.UpsertUserSettingsParams) (*gen.CoreUserSetting, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	row := &gen.CoreUserSetting{
		UserID:               params.UserID,
		SoundEffects:         params.SoundEffects,
		Animations:           params.Animations,
		MotivationalMessages: params.MotivationalMessages,
		ListeningExercises:   params.ListeningExercises,
		Theme:                params.Theme,
		Locale:               params.Locale,
		PrivacyJson:          append([]byte(nil), params.PrivacyJson...),
	}
	m.rows[params.UserID] = row
	cp := *row
	return &cp, nil
}

func (m *memSettingsRepo) EnsureDefaults(_ context.Context, userID string) (*gen.CoreUserSetting, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if row, ok := m.rows[userID]; ok {
		cp := *row
		return &cp, nil
	}
	d := productDefaultSettings(userID)
	m.rows[userID] = d
	cp := *d
	return &cp, nil
}

func TestMergeSettings_GuestWinsOnlyVsDefaults(t *testing.T) {
	users := newMemUserRepo()
	stats := newMemStatsRepo()
	settings := newMemSettingsRepo()
	uc := NewProfileUsecase(users, stats, settings)

	userID := "u-merge"
	// server: sound=false (custom), animations=true (default)
	_, err := settings.Upsert(context.Background(), gen.UpsertUserSettingsParams{
		UserID:               userID,
		SoundEffects:         false,
		Animations:           true,
		MotivationalMessages: true,
		ListeningExercises:   true,
		Theme:                "system",
		Locale:               "en",
		PrivacyJson:          []byte("{}"),
	})
	if err != nil {
		t.Fatalf("seed server settings: %v", err)
	}

	// guest: sound=true, animations=false
	guest := SettingsValues{
		SoundEffects:         true,
		Animations:           false,
		MotivationalMessages: true,
		ListeningExercises:   true,
		Theme:                "system",
		Locale:               "en",
		PrivacyJSON:          "{}",
	}

	result, err := uc.MergeSettings(context.Background(), userID, &guest)
	if err != nil {
		t.Fatalf("MergeSettings: %v", err)
	}

	// expect: sound=false (keep server), animations=false (guest)
	if result.Settings.SoundEffects {
		t.Fatalf("sound_effects = true, want false (keep server custom)")
	}
	if result.Settings.Animations {
		t.Fatalf("animations = true, want false (guest wins vs default)")
	}

	if len(result.FieldsMerged) != 1 || result.FieldsMerged[0] != "animations" {
		t.Fatalf("fields_merged = %v, want [animations]", result.FieldsMerged)
	}
}

func TestUpdateSettings_RejectsInvalidTheme(t *testing.T) {
	users := newMemUserRepo()
	stats := newMemStatsRepo()
	settings := newMemSettingsRepo()
	uc := NewProfileUsecase(users, stats, settings)

	userID := "u-theme"
	invalid := "neon"
	_, err := uc.UpdateSettings(context.Background(), userID, UpdateSettingsInput{
		Theme: &invalid,
	})
	if !errors.Is(err, ErrInvalidTheme) {
		t.Fatalf("want ErrInvalidTheme, got %v", err)
	}
}
