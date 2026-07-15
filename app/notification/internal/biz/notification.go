package biz

import (
	"bytes"
	"context"
	"fmt"
	"text/template"
	"time"
)

// PreferenceRepo defines the contract for preference persistence.
type PreferenceRepo interface {
	GetPreferences(ctx context.Context, userID string) (*Preference, error)
	UpsertPreferences(ctx context.Context, prefs *Preference) (*Preference, error)
}

// Preference represents notification preferences for a user.
type Preference struct {
	UserID          string
	PushEnabled     bool
	EmailEnabled    bool
	StreakReminder  bool
	FriendActivity  bool
	Promotions      bool
	QuietHoursStart *string
	QuietHoursEnd   *string
}

// GotifySender sends push notifications via Gotify.
type GotifySender interface {
	Send(msg GotifyMessage) error
}

// GotifyMessage is the data sent to Gotify.
type GotifyMessage struct {
	Title    string
	Message  string
	Priority int
}

// NotificationUsecase handles notification business logic.
type NotificationUsecase struct {
	prefRepo PreferenceRepo
	gotify   GotifySender
}

// NewNotificationUsecase creates a new NotificationUsecase.
func NewNotificationUsecase(prefRepo PreferenceRepo, gotify GotifySender) *NotificationUsecase {
	return &NotificationUsecase{prefRepo: prefRepo, gotify: gotify}
}

// GetPreferences returns the user's notification preferences.
func (uc *NotificationUsecase) GetPreferences(ctx context.Context, userID string) (*Preference, error) {
	return uc.prefRepo.GetPreferences(ctx, userID)
}

// UpdatePreferences updates the user's notification preferences and returns the result.
func (uc *NotificationUsecase) UpdatePreferences(ctx context.Context, prefs *Preference) (*Preference, error) {
	return uc.prefRepo.UpsertPreferences(ctx, prefs)
}

// SendNotification sends a notification to the user via Gotify, respecting their preferences.
func (uc *NotificationUsecase) SendNotification(ctx context.Context, userID, category, templateID string, params map[string]string) (bool, error) {
	prefs, err := uc.prefRepo.GetPreferences(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("get preferences: %w", err)
	}

	if !prefs.PushEnabled {
		return false, nil
	}

	if !isCategoryAllowed(category, prefs) {
		return false, nil
	}

	if inQuietHours(prefs) {
		return false, nil
	}

	title, body := resolveTemplate(templateID, category, params)
	if title == "" || body == "" {
		return false, fmt.Errorf("unknown template: %s", templateID)
	}

	msg := GotifyMessage{
		Title:    title,
		Message:  body,
		Priority: priorityForCategory(category),
	}
	if err := uc.gotify.Send(msg); err != nil {
		return false, fmt.Errorf("send gotify: %w", err)
	}
	return true, nil
}

func isCategoryAllowed(category string, prefs *Preference) bool {
	switch category {
	case "streak_reminder":
		return prefs.StreakReminder
	case "friend_activity":
		return prefs.FriendActivity
	case "promotions":
		return prefs.Promotions
	default:
		return true
	}
}

func inQuietHours(prefs *Preference) bool {
	if prefs.QuietHoursStart == nil || prefs.QuietHoursEnd == nil {
		return false
	}
	now := time.Now()
	start, err := time.Parse("15:04", *prefs.QuietHoursStart)
	if err != nil {
		return false
	}
	end, err := time.Parse("15:04", *prefs.QuietHoursEnd)
	if err != nil {
		return false
	}
	currentMinutes := now.Hour()*60 + now.Minute()
	startMinutes := start.Hour()*60 + start.Minute()
	endMinutes := end.Hour()*60 + end.Minute()

	if startMinutes <= endMinutes {
		return currentMinutes >= startMinutes && currentMinutes < endMinutes
	}
	return currentMinutes >= startMinutes || currentMinutes < endMinutes
}

func priorityForCategory(category string) int {
	switch category {
	case "streak_reminder":
		return 3
	case "achievement":
		return 7
	case "friend_activity":
		return 5
	default:
		return 5
	}
}

func resolveTemplate(templateID, category string, params map[string]string) (string, string) {
	templates := getDefaultTemplates()
	t, ok := templates[templateID]
	if !ok {
		return "", ""
	}
	title := execTemplate(t.Title, params)
	body := execTemplate(t.Body, params)
	return title, body
}

type templateDef struct {
	Title string
	Body  string
}

func getDefaultTemplates() map[string]templateDef {
	return map[string]templateDef{
		"streak_reminder": {
			Title: "Streak Reminder",
			Body:  "You have a streak of {{.Streak}} days! Keep going!",
		},
		"friend_joined": {
			Title: "Friend Joined",
			Body:  "{{.FriendName}} just joined Puchi! Say hello!",
		},
		"achievement_unlocked": {
			Title: "Achievement Unlocked",
			Body:  "Congratulations! You unlocked '{{.AchievementName}}'!",
		},
	}
}

func execTemplate(tmplText string, params map[string]string) string {
	tmpl, err := template.New("").Parse(tmplText)
	if err != nil {
		return tmplText
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, params); err != nil {
		return tmplText
	}
	return buf.String()
}
