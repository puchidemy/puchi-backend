package biz

import (
	"context"
)

// LessonCompletedEvent is emitted when a user completes a lesson.
type LessonCompletedEvent struct {
	UserID   string
	LessonID string
	UnitID   string
	XP       int32
}

// UnitCompletedEvent is emitted when a user completes all required lessons in a unit.
type UnitCompletedEvent struct {
	UserID string
	UnitID string
	XP     int32
}

// LessonEventPublisher publishes learn completion events (user-only at call sites).
type LessonEventPublisher interface {
	PublishLessonCompleted(ctx context.Context, ev LessonCompletedEvent) error
	PublishUnitCompleted(ctx context.Context, ev UnitCompletedEvent) error
}

// NoOpLessonEventPublisher discards events (Task 6 wires real NATS).
type NoOpLessonEventPublisher struct{}

func (NoOpLessonEventPublisher) PublishLessonCompleted(_ context.Context, _ LessonCompletedEvent) error {
	return nil
}

func (NoOpLessonEventPublisher) PublishUnitCompleted(_ context.Context, _ UnitCompletedEvent) error {
	return nil
}

// NewNoOpLessonEventPublisher returns a no-op publisher for wire.
func NewNoOpLessonEventPublisher() *NoOpLessonEventPublisher {
	return &NoOpLessonEventPublisher{}
}
