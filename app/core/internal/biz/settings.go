package biz

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/puchidemy/puchi-backend/app/core/internal/data/sqlc/gen"
)

var (
	ErrInvalidTheme            = errors.New("invalid theme")
	ErrInvalidLocale           = errors.New("invalid locale")
	ErrGuestSettingsRequired   = errors.New("guest settings required")
)

var (
	supportedLocales = map[string]struct{}{
		"en": {}, "zh": {}, "de": {}, "es": {}, "fr": {},
		"it": {}, "ja": {}, "ko": {}, "ru": {},
	}
	localePattern = regexp.MustCompile(`^[a-z]{2}(-[a-zA-Z0-9]{2,8})?$`)
)

func validateTheme(theme string) error {
	switch theme {
	case "system", "light", "dark":
		return nil
	default:
		return ErrInvalidTheme
	}
}

func validateLocale(locale string) error {
	if locale == "" {
		return ErrInvalidLocale
	}
	if len(locale) < 2 || len(locale) > 16 {
		return ErrInvalidLocale
	}
	if _, ok := supportedLocales[locale]; ok {
		return nil
	}
	if localePattern.MatchString(locale) {
		return nil
	}
	return ErrInvalidLocale
}

func validateSettingsValues(v SettingsValues) error {
	if err := validateTheme(v.Theme); err != nil {
		return err
	}
	return validateLocale(v.Locale)
}

// SettingsRepo defines the repository contract for user settings.
type SettingsRepo interface {
	Get(ctx context.Context, userID string) (*gen.CoreUserSetting, error)
	Upsert(ctx context.Context, params gen.UpsertUserSettingsParams) (*gen.CoreUserSetting, error)
	EnsureDefaults(ctx context.Context, userID string) (*gen.CoreUserSetting, error)
}

// SettingsValues is the domain shape for settings fields (merge / update).
type SettingsValues struct {
	SoundEffects         bool
	Animations           bool
	MotivationalMessages bool
	ListeningExercises   bool
	Theme                string
	Locale               string
	PrivacyJSON          string
}

// UpdateSettingsInput holds partial settings updates (nil = leave unchanged).
type UpdateSettingsInput struct {
	SoundEffects         *bool
	Animations           *bool
	MotivationalMessages *bool
	ListeningExercises   *bool
	Theme                *string
	Locale               *string
	PrivacyJSON          *string
}

// MergeSettingsResult is the outcome of merging guest settings into server settings.
type MergeSettingsResult struct {
	Settings     *gen.CoreUserSetting
	FieldsMerged []string
}

// Product defaults for user_settings (must match DB defaults / FE defaults).
func productDefaults() SettingsValues {
	return SettingsValues{
		SoundEffects:         true,
		Animations:           true,
		MotivationalMessages: true,
		ListeningExercises:   true,
		Theme:                "system",
		Locale:               "en",
		PrivacyJSON:          "{}",
	}
}

func productDefaultSettings(userID string) *gen.CoreUserSetting {
	d := productDefaults()
	return &gen.CoreUserSetting{
		UserID:               userID,
		SoundEffects:         d.SoundEffects,
		Animations:           d.Animations,
		MotivationalMessages: d.MotivationalMessages,
		ListeningExercises:   d.ListeningExercises,
		Theme:                d.Theme,
		Locale:               d.Locale,
		PrivacyJson:          []byte(d.PrivacyJSON),
	}
}

func settingsFromRow(row *gen.CoreUserSetting) SettingsValues {
	privacy := "{}"
	if len(row.PrivacyJson) > 0 {
		privacy = string(row.PrivacyJson)
	}
	return SettingsValues{
		SoundEffects:         row.SoundEffects,
		Animations:           row.Animations,
		MotivationalMessages: row.MotivationalMessages,
		ListeningExercises:   row.ListeningExercises,
		Theme:                row.Theme,
		Locale:               row.Locale,
		PrivacyJSON:          privacy,
	}
}

func privacyEqual(a, b string) bool {
	if a == "" {
		a = "{}"
	}
	if b == "" {
		b = "{}"
	}
	return bytes.Equal(bytes.TrimSpace([]byte(a)), bytes.TrimSpace([]byte(b)))
}

// MergeSettingsValues applies the guest-wins-only-vs-defaults rule.
// Guest wins a field only if guest ≠ product default AND server == default.
func MergeSettingsValues(server, guest SettingsValues) (SettingsValues, []string) {
	def := productDefaults()
	out := server
	var merged []string

	if guest.SoundEffects != def.SoundEffects && server.SoundEffects == def.SoundEffects {
		out.SoundEffects = guest.SoundEffects
		merged = append(merged, "sound_effects")
	}
	if guest.Animations != def.Animations && server.Animations == def.Animations {
		out.Animations = guest.Animations
		merged = append(merged, "animations")
	}
	if guest.MotivationalMessages != def.MotivationalMessages && server.MotivationalMessages == def.MotivationalMessages {
		out.MotivationalMessages = guest.MotivationalMessages
		merged = append(merged, "motivational_messages")
	}
	if guest.ListeningExercises != def.ListeningExercises && server.ListeningExercises == def.ListeningExercises {
		out.ListeningExercises = guest.ListeningExercises
		merged = append(merged, "listening_exercises")
	}
	if guest.Theme != def.Theme && server.Theme == def.Theme {
		out.Theme = guest.Theme
		merged = append(merged, "theme")
	}
	if guest.Locale != def.Locale && server.Locale == def.Locale {
		out.Locale = guest.Locale
		merged = append(merged, "locale")
	}
	if !privacyEqual(guest.PrivacyJSON, def.PrivacyJSON) && privacyEqual(server.PrivacyJSON, def.PrivacyJSON) {
		out.PrivacyJSON = guest.PrivacyJSON
		merged = append(merged, "privacy_json")
	}

	return out, merged
}

func toUpsertParams(userID string, v SettingsValues) gen.UpsertUserSettingsParams {
	privacy := []byte(v.PrivacyJSON)
	if len(privacy) == 0 {
		privacy = []byte("{}")
	}
	return gen.UpsertUserSettingsParams{
		UserID:               userID,
		SoundEffects:         v.SoundEffects,
		Animations:           v.Animations,
		MotivationalMessages: v.MotivationalMessages,
		ListeningExercises:   v.ListeningExercises,
		Theme:                v.Theme,
		Locale:               v.Locale,
		PrivacyJson:          privacy,
	}
}

// GetSettings returns settings for the user, inserting product defaults if missing.
func (uc *ProfileUsecase) GetSettings(ctx context.Context, userID string) (*gen.CoreUserSetting, error) {
	row, err := uc.settings.EnsureDefaults(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get settings: %w", err)
	}
	return row, nil
}

// UpdateSettings applies a partial update and returns the full settings row.
func (uc *ProfileUsecase) UpdateSettings(ctx context.Context, userID string, input UpdateSettingsInput) (*gen.CoreUserSetting, error) {
	current, err := uc.settings.EnsureDefaults(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("load settings: %w", err)
	}
	v := settingsFromRow(current)
	if input.SoundEffects != nil {
		v.SoundEffects = *input.SoundEffects
	}
	if input.Animations != nil {
		v.Animations = *input.Animations
	}
	if input.MotivationalMessages != nil {
		v.MotivationalMessages = *input.MotivationalMessages
	}
	if input.ListeningExercises != nil {
		v.ListeningExercises = *input.ListeningExercises
	}
	if input.Theme != nil {
		v.Theme = *input.Theme
	}
	if input.Locale != nil {
		v.Locale = *input.Locale
	}
	if input.PrivacyJSON != nil {
		v.PrivacyJSON = *input.PrivacyJSON
	}
	if input.Theme != nil {
		if err := validateTheme(*input.Theme); err != nil {
			return nil, err
		}
	}
	if input.Locale != nil {
		if err := validateLocale(*input.Locale); err != nil {
			return nil, err
		}
	}

	row, err := uc.settings.Upsert(ctx, toUpsertParams(userID, v))
	if err != nil {
		return nil, fmt.Errorf("update settings: %w", err)
	}
	return row, nil
}

// MergeSettings merges guest settings into the authenticated user's settings.
func (uc *ProfileUsecase) MergeSettings(ctx context.Context, userID string, guest *SettingsValues) (*MergeSettingsResult, error) {
	if guest == nil {
		return nil, ErrGuestSettingsRequired
	}
	if err := validateSettingsValues(*guest); err != nil {
		return nil, err
	}

	current, err := uc.settings.EnsureDefaults(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("load settings: %w", err)
	}
	merged, fields := MergeSettingsValues(settingsFromRow(current), *guest)
	row, err := uc.settings.Upsert(ctx, toUpsertParams(userID, merged))
	if err != nil {
		return nil, fmt.Errorf("merge settings: %w", err)
	}
	return &MergeSettingsResult{Settings: row, FieldsMerged: fields}, nil
}
