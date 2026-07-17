package nats

import (
	"fmt"
	"log/slog"

	natsio "github.com/nats-io/nats.go"
)

// ConnectOptional connects when url is non-empty; otherwise returns (nil, noop cleanup, nil).
// Callers treat a nil connection as "NATS disabled".
func ConnectOptional(url string, log *slog.Logger) (*natsio.Conn, func(), error) {
	cleanup := func() {}
	if url == "" {
		if log != nil {
			log.Info("nats disabled (empty url)")
		}
		return nil, cleanup, nil
	}

	nc, err := Connect(url)
	if err != nil {
		return nil, nil, fmt.Errorf("nats connect: %w", err)
	}
	if log != nil {
		log.Info("nats connected", "url", url)
	}
	cleanup = func() {
		_ = nc.Drain()
		nc.Close()
	}
	return nc, cleanup, nil
}
