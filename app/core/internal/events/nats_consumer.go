package events

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/puchidemy/puchi-backend/app/core/internal/biz"
	"github.com/puchidemy/puchi-backend/app/core/internal/conf"
	pnats "github.com/puchidemy/puchi-backend/pkg/nats"

	"github.com/google/wire"
)

// ProviderSet is events providers.
var ProviderSet = wire.NewSet(NewLearnConsumer)

type lessonCompletedPayload struct {
	UserID      string `json:"user_id"`
	LessonID    string `json:"lesson_id"`
	UnitID      string `json:"unit_id"`
	XP          int32  `json:"xp"`
	CompletedAt string `json:"completed_at"`
}

type unitCompletedPayload struct {
	UserID      string `json:"user_id"`
	UnitID      string `json:"unit_id"`
	XP          int32  `json:"xp"`
	CompletedAt string `json:"completed_at"`
}

// LearnConsumer subscribes to learn completion NATS subjects.
type LearnConsumer struct {
	nc   *nats.Conn
	subs []*nats.Subscription
	log  *slog.Logger
}

// NewLearnConsumer connects and subscribes when NATS URL is set.
func NewLearnConsumer(cfg *conf.Data, uc *biz.StatsUsecase, log *slog.Logger) (*LearnConsumer, func(), error) {
	c := &LearnConsumer{log: log}

	url := ""
	if cfg != nil && cfg.GetNats() != nil {
		url = cfg.GetNats().GetUrl()
	}

	nc, baseCleanup, err := pnats.ConnectOptional(url, log)
	if err != nil {
		return nil, nil, err
	}
	if nc == nil {
		return c, baseCleanup, nil
	}
	c.nc = nc

	cleanup := func() {
		for _, sub := range c.subs {
			_ = sub.Unsubscribe()
		}
		baseCleanup()
	}

	lessonSub, err := nc.QueueSubscribe(pnats.SubjectLessonCompleted, pnats.QueueCoreLearn, func(msg *nats.Msg) {
		c.handleLessonCompleted(uc, msg)
	})
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	unitSub, err := nc.QueueSubscribe(pnats.SubjectUnitCompleted, pnats.QueueCoreLearn, func(msg *nats.Msg) {
		c.handleUnitCompleted(uc, msg)
	})
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	c.subs = []*nats.Subscription{lessonSub, unitSub}
	log.Info("nats learn consumer subscribed", "subjects", []string{pnats.SubjectLessonCompleted, pnats.SubjectUnitCompleted})
	return c, cleanup, nil
}

func (c *LearnConsumer) handleLessonCompleted(uc *biz.StatsUsecase, msg *nats.Msg) {
	var payload lessonCompletedPayload
	if err := json.Unmarshal(msg.Data, &payload); err != nil {
		c.log.Error("decode learn.lesson.completed", "err", err)
		return
	}
	completedAt := parseCompletedAt(payload.CompletedAt)
	if err := uc.OnLessonCompleted(context.Background(), biz.LessonCompletedEvent{
		UserID:      payload.UserID,
		LessonID:    payload.LessonID,
		UnitID:      payload.UnitID,
		XP:          payload.XP,
		CompletedAt: completedAt,
	}); err != nil {
		c.log.Error("apply learn.lesson.completed", "user_id", payload.UserID, "lesson_id", payload.LessonID, "err", err)
	}
}

func (c *LearnConsumer) handleUnitCompleted(uc *biz.StatsUsecase, msg *nats.Msg) {
	var payload unitCompletedPayload
	if err := json.Unmarshal(msg.Data, &payload); err != nil {
		c.log.Error("decode learn.unit.completed", "err", err)
		return
	}
	completedAt := parseCompletedAt(payload.CompletedAt)
	if err := uc.OnUnitCompleted(context.Background(), biz.UnitCompletedEvent{
		UserID:      payload.UserID,
		UnitID:      payload.UnitID,
		XP:          payload.XP,
		CompletedAt: completedAt,
	}); err != nil {
		c.log.Error("apply learn.unit.completed", "user_id", payload.UserID, "unit_id", payload.UnitID, "err", err)
	}
}

func parseCompletedAt(raw string) time.Time {
	if raw == "" {
		return time.Now().UTC()
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Now().UTC()
	}
	return t.UTC()
}
