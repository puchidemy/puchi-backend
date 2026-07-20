package nats

import (
	"time"

	natsio "github.com/nats-io/nats.go"
)

// Subject names shared across services. Keep in sync with publishers/consumers.
const (
	SubjectUserCreated     = "auth.user.created"
	SubjectEmailSend       = "email.send"
	SubjectLessonCompleted = "learn.lesson.completed"
	SubjectUnitCompleted   = "learn.unit.completed"
	SubjectSceneCompleted  = "learn.scene.completed"
	SubjectStoryCompleted  = "learn.story.completed"
)

// Queue groups for competing consumers.
const (
	QueueCoreLearn = "core-learn"
)

// DefaultReconnectWait between reconnect attempts.
const DefaultReconnectWait = 2 * time.Second

// Connect opens a NATS connection with Puchi defaults (infinite reconnect).
func Connect(url string) (*natsio.Conn, error) {
	return natsio.Connect(url, DefaultOptions()...)
}

// DefaultOptions are shared connection options.
func DefaultOptions() []natsio.Option {
	return []natsio.Option{
		natsio.MaxReconnects(-1),
		natsio.ReconnectWait(DefaultReconnectWait),
		natsio.RetryOnFailedConnect(true),
	}
}
