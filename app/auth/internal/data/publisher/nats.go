package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
)

// Publisher handles outbox-based event publishing to NATS.
type Publisher struct {
	nc     *nats.Conn
	pool   *pgxpool.Pool
	logger *slog.Logger
}

// OutboxMessage represents a row from the auth.outbox table.
type OutboxMessage struct {
	ID          uuid.UUID  `json:"id"`
	Topic       string     `json:"topic"`
	Payload     []byte     `json:"payload"`
	CreatedAt   time.Time  `json:"created_at"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	RetryCount  int        `json:"retry_count"`
}

// New creates a new Publisher, connecting to NATS and using the given DB pool.
// If natsURL is empty, the publisher operates in "noop" mode (Publish is a no-op).
func New(natsURL string, pool *pgxpool.Pool) (*Publisher, error) {
	if natsURL == "" {
		return &Publisher{
			pool:   pool,
			logger: slog.Default(),
		}, nil
	}

	nc, err := nats.Connect(natsURL,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(10),
		nats.ReconnectWait(2*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("connect to NATS: %w", err)
	}
	return &Publisher{
		nc:     nc,
		pool:   pool,
		logger: slog.Default(),
	}, nil
}

// Close closes the NATS connection.
func (p *Publisher) Close() {
	if p.nc != nil {
		p.nc.Close()
	}
}

// Publish writes an event to the outbox table (transactional with the DB operation).
// If NATS is not connected (noop mode), the event is still written to the outbox.
func (p *Publisher) Publish(ctx context.Context, topic string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal outbox payload: %w", err)
	}

	_, err = p.pool.Exec(ctx,
		`INSERT INTO auth.outbox (topic, payload) VALUES ($1, $2)`,
		topic, data)
	if err != nil {
		return fmt.Errorf("insert outbox: %w", err)
	}
	return nil
}

// Start starts a background goroutine that polls the outbox table
// and publishes pending messages to NATS. If NATS is not connected (noop mode),
// this is a no-op.
func (p *Publisher) Start(ctx context.Context) {
	if p.nc == nil {
		return
	}
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				p.logger.Info("outbox publisher stopped")
				return
			case <-ticker.C:
				p.processOutbox(ctx)
			}
		}
	}()
}

func (p *Publisher) processOutbox(ctx context.Context) {
	rows, err := p.pool.Query(ctx,
		`SELECT id, topic, payload, retry_count FROM auth.outbox
		 WHERE published_at IS NULL AND retry_count < 3
		 ORDER BY created_at ASC LIMIT 50`)
	if err != nil {
		p.logger.Error("outbox query", "error", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var msg OutboxMessage
		if err := rows.Scan(&msg.ID, &msg.Topic, &msg.Payload, &msg.RetryCount); err != nil {
			p.logger.Error("outbox scan", "error", err)
			continue
		}

		// Publish to NATS
		if err := p.nc.Publish(msg.Topic, msg.Payload); err != nil {
			p.logger.Error("nats publish", "topic", msg.Topic, "id", msg.ID, "error", err)
			// Increment retry count
			_, _ = p.pool.Exec(ctx,
				`UPDATE auth.outbox SET retry_count = retry_count + 1 WHERE id = $1`,
				msg.ID)
			continue
		}

		// Mark as published
		_, _ = p.pool.Exec(ctx,
			`UPDATE auth.outbox SET published_at = now() WHERE id = $1`,
			msg.ID)
	}
}

// Event subjects/topics
const (
	EventUserCreated    = "auth.user.created"
	EventUserDeleted    = "auth.user.deleted"
	EventUserDisabled   = "auth.user.disabled"
	EventSessionRevoked = "auth.session.revoked"
	EventSocialLinked   = "auth.social.linked"
	EventEmailSend      = "email.send"
)
