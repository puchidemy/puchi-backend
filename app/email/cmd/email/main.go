package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/puchidemy/puchi-backend/app/email/internal/handler"
	"github.com/puchidemy/puchi-backend/app/email/internal/sender"

	"github.com/nats-io/nats.go"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	}))

	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}

	// Connect to NATS
	nc, err := nats.Connect(natsURL,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(10),
		nats.ReconnectWait(2),
	)
	if err != nil {
		logger.Error("connect to NATS", "error", err)
		os.Exit(1)
	}
	defer nc.Close()
	logger.Info("connected to NATS", "url", natsURL)

	// Create SMTP sender
	smtpSender := sender.NewFromEnv(logger)
	if !smtpSender.IsConfigured() {
		logger.Warn("SMTP not fully configured — emails will be logged instead of sent")
	}

	// Create NATS subscription handler
	emailHandler := handler.New(smtpSender, logger)

	// Subscribe to email.send topic
	sub, err := nc.Subscribe("email.send", emailHandler.HandleMessage)
	if err != nil {
		logger.Error("subscribe to email.send", "error", err)
		os.Exit(1)
	}
	logger.Info("subscribed to email.send")

	// Wait for shutdown signal
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	<-ctx.Done()
	logger.Info("shutting down")

	sub.Unsubscribe()
	nc.Flush()
	logger.Info("email service stopped")
}
