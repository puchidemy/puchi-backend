package events

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

const (
	SubjectUserCreated = "auth.user.created"
	SubjectEmailSend   = "email.send"
)

// Publisher publishes auth domain events. No-op when NATS URL is empty.
type Publisher struct {
	nc  *nats.Conn
	log *slog.Logger
	mu  sync.Mutex
}

func New(url string, log *slog.Logger) (*Publisher, error) {
	p := &Publisher{log: log}
	if url == "" {
		log.Info("nats disabled (empty url)")
		return p, nil
	}
	nc, err := nats.Connect(url,
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2*time.Second),
	)
	if err != nil {
		return nil, err
	}
	p.nc = nc
	log.Info("nats connected", "url", url)
	return p, nil
}

func (p *Publisher) Close() {
	if p.nc != nil {
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
	go p.Publish(context.Background(), SubjectUserCreated, map[string]any{
		"user_id":  userID,
		"email":    email,
		"username": username,
	})
}

func (p *Publisher) SendEmail(to, template string, data map[string]any) {
	go p.Publish(context.Background(), SubjectEmailSend, map[string]any{
		"to":       to,
		"template": template,
		"data":     data,
	})
}
