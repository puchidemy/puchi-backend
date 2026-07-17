package data

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/puchidemy/puchi-backend/app/learn/internal/biz"
	"github.com/puchidemy/puchi-backend/app/learn/internal/conf"
)

const (
	subjectLessonCompleted = "learn.lesson.completed"
	subjectUnitCompleted   = "learn.unit.completed"
)

type lessonCompletedPayload struct {
	UserID      string `json:"user_id"`
	LessonID    string `json:"lesson_id"`
	UnitID      string `json:"unit_id"`
	XP          int32  `json:"xp"`
	CompletedAt string `json:"completed_at"`
}

type unitCompletedPayload struct {
	UserID      string `json:"user_id"`
	LessonID    string `json:"lesson_id"`
	UnitID      string `json:"unit_id"`
	XP          int32  `json:"xp"`
	CompletedAt string `json:"completed_at"`
}

// NATSLessonEventPublisher publishes learn completion events to NATS.
type NATSLessonEventPublisher struct {
	nc  *nats.Conn
	log *slog.Logger
}

// NewNATSLessonEventPublisher connects to NATS when url is set; otherwise no-op.
func NewNATSLessonEventPublisher(cfg *conf.Data, log *slog.Logger) (*NATSLessonEventPublisher, func(), error) {
	p := &NATSLessonEventPublisher{log: log}
	cleanup := func() {}

	url := ""
	if cfg != nil && cfg.GetNats() != nil {
		url = cfg.GetNats().GetUrl()
	}
	if url == "" {
		log.Info("nats disabled (empty url)")
		return p, cleanup, nil
	}

	nc, err := nats.Connect(url,
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2*time.Second),
	)
	if err != nil {
		return nil, nil, err
	}
	p.nc = nc
	cleanup = func() {
		nc.Close()
	}
	log.Info("nats connected", "url", url)
	return p, cleanup, nil
}

func (p *NATSLessonEventPublisher) PublishLessonCompleted(_ context.Context, ev biz.LessonCompletedEvent) error {
	return p.publish(subjectLessonCompleted, lessonCompletedPayload{
		UserID:      ev.UserID,
		LessonID:    ev.LessonID,
		UnitID:      ev.UnitID,
		XP:          ev.XP,
		CompletedAt: formatCompletedAt(ev.CompletedAt),
	})
}

func (p *NATSLessonEventPublisher) PublishUnitCompleted(_ context.Context, ev biz.UnitCompletedEvent) error {
	return p.publish(subjectUnitCompleted, unitCompletedPayload{
		UserID:      ev.UserID,
		UnitID:      ev.UnitID,
		XP:          ev.XP,
		CompletedAt: formatCompletedAt(ev.CompletedAt),
	})
}

func (p *NATSLessonEventPublisher) publish(subject string, payload any) error {
	if p.nc == nil {
		p.log.Debug("nats noop publish", "subject", subject)
		return nil
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal %s: %w", subject, err)
	}
	if err := p.nc.Publish(subject, b); err != nil {
		return fmt.Errorf("publish %s: %w", subject, err)
	}
	return nil
}

func formatCompletedAt(t time.Time) string {
	if t.IsZero() {
		t = time.Now().UTC()
	}
	return t.UTC().Format(time.RFC3339)
}
