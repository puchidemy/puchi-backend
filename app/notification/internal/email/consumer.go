package email

import (
	"fmt"
	"log/slog"
	"os"

	pnats "github.com/puchidemy/puchi-backend/pkg/nats"
)

// StartConsumer connects to NATS and subscribes to email.send.
// Returns a cleanup function that unsubscribes and closes the connection.
func StartConsumer(logger *slog.Logger) (func(), error) {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}

	nc, cleanup, err := pnats.ConnectOptional(natsURL, logger)
	if err != nil {
		return nil, err
	}
	if nc == nil {
		return cleanup, nil
	}

	smtpSender := NewFromEnv(logger)
	if !smtpSender.IsConfigured() {
		logger.Warn("SMTP not fully configured — emails will be logged instead of sent")
	} else {
		logger.Info("SMTP configured", "host", os.Getenv("SMTP_HOST"), "port", os.Getenv("SMTP_PORT"))
	}

	h := NewHandler(smtpSender, logger)
	sub, err := nc.Subscribe(pnats.SubjectEmailSend, h.HandleMessage)
	if err != nil {
		cleanup()
		return nil, fmt.Errorf("subscribe %s: %w", pnats.SubjectEmailSend, err)
	}
	logger.Info("subscribed to " + pnats.SubjectEmailSend)

	return func() {
		_ = sub.Unsubscribe()
		cleanup()
		logger.Info("email consumer stopped")
	}, nil
}
