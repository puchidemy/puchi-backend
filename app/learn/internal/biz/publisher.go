package biz

import (
	"context"
	"time"
)

// LessonCompletedEvent is emitted when a user completes a lesson.
type LessonCompletedEvent struct {
	UserID      string
	LessonID    string
	UnitID      string
	XP          int32
	CompletedAt time.Time
}

// UnitCompletedEvent is emitted when a user completes all required lessons in a unit.
type UnitCompletedEvent struct {
	UserID      string
	UnitID      string
	XP          int32
	CompletedAt time.Time
}

// SceneCompletedEvent is emitted when a user completes a story scene.
type SceneCompletedEvent struct {
	UserID      string
	SceneID     string
	StoryID     string
	CompletedAt time.Time
}

// StoryCompletedEvent is emitted when a user completes a story.
type StoryCompletedEvent struct {
	UserID      string
	StoryID     string
	CityID      string
	XP          int32
	CompletedAt time.Time
}

// LessonEventPublisher publishes learn completion events (user-only at call sites).
type LessonEventPublisher interface {
	PublishLessonCompleted(ctx context.Context, ev LessonCompletedEvent) error
	PublishUnitCompleted(ctx context.Context, ev UnitCompletedEvent) error
	PublishSceneCompleted(ctx context.Context, ev SceneCompletedEvent) error
	PublishStoryCompleted(ctx context.Context, ev StoryCompletedEvent) error
}

// NoOpLessonEventPublisher discards events.
type NoOpLessonEventPublisher struct{}

func (NoOpLessonEventPublisher) PublishLessonCompleted(_ context.Context, _ LessonCompletedEvent) error {
	return nil
}

func (NoOpLessonEventPublisher) PublishUnitCompleted(_ context.Context, _ UnitCompletedEvent) error {
	return nil
}

func (NoOpLessonEventPublisher) PublishSceneCompleted(_ context.Context, _ SceneCompletedEvent) error {
	return nil
}

func (NoOpLessonEventPublisher) PublishStoryCompleted(_ context.Context, _ StoryCompletedEvent) error {
	return nil
}

// NewNoOpLessonEventPublisher returns a no-op publisher for wire.
func NewNoOpLessonEventPublisher() *NoOpLessonEventPublisher {
	return &NoOpLessonEventPublisher{}
}
