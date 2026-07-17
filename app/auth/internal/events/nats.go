package events

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/nats-io/nats.go"
	pnats "github.com/puchidemy/puchi-backend/pkg/nats"
)

// Publisher publishes auth domain events. No-op when NATS URL is empty.
type Publisher struct {
	nc  *nats.Conn
	log *slog.Logger
	mu  sync.Mutex
}

func New(url string, log *slog.Logger) (*Publisher, error) {
	p := &Publisher{log: log}
	nc, cleanup, err := pnats.ConnectOptional(url, log)
	if err != nil {
		return nil, err
	}
	_ = cleanup // Close() owns lifecycle
	p.nc = nc
	return p, nil
}

func (p *Publisher) Close() {
	if p.nc != nil {
		_ = p.nc.Drain()
		p.nc.Close()
	}
}

func (p *Publisher) Publish(ctx context.Context, subject string, payload any) {
	if p.nc == nil {
		p.log.Debug("nats noop publish", "subject", subject)
		return
	}
	b, err := json.Marshal(payload)
	if err != nil {
		p.log.Error("marshal event", "subject", subject, "err", err)
		return
	}
	if err := p.nc.Publish(subject, b); err != nil {
		p.log.Error("publish event", "subject", subject, "err", err)
	}
}

func (p *Publisher) UserCreated(userID, email, username string) {
	go p.Publish(context.Background(), pnats.SubjectUserCreated, map[string]any{
		"user_id":  userID,
		"email":    email,
		"username": username,
	})
}

func (p *Publisher) SendEmail(to, template string, data map[string]any) {
	go p.Publish(context.Background(), pnats.SubjectEmailSend, map[string]any{
		"to":       to,
		"template": template,
		"data":     data,
	})
}
