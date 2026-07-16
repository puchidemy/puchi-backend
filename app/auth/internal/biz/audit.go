package biz

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

// AuditUsecase handles audit logging (fire-and-forget in Phase 3).
type AuditUsecase struct {
	auditRepo AuditRepo
	logger    *slog.Logger
}

// NewAuditUsecase creates a new AuditUsecase.
func NewAuditUsecase(auditRepo AuditRepo, logger *slog.Logger) *AuditUsecase {
	return &AuditUsecase{
		auditRepo: auditRepo,
		logger:    logger,
	}
}

// Log creates an audit log entry. Fire-and-forget in Phase 3 so it doesn't
// block the main flow. In production this should use an async worker / event bus.
func (uc *AuditUsecase) Log(ctx context.Context, userID *uuid.UUID, action string, resource string, resourceID string, ip string, ua string, metadata map[string]any) {
	go func() {
		// Use a detached context so the log outlives the request
		logCtx := context.WithoutCancel(ctx)

		var metaRaw json.RawMessage
		if metadata != nil {
			raw, err := json.Marshal(metadata)
			if err != nil {
				uc.logger.WarnContext(logCtx, "failed to marshal audit metadata",
					slog.String("action", action),
					slog.Any("error", err),
				)
				return
			}
			metaRaw = raw
		}

		if err := uc.auditRepo.Create(logCtx, &AuditLog{
			UserID:     userID,
			Action:     action,
			Resource:   resource,
			ResourceID: resourceID,
			IPAddress:  ip,
			UserAgent:  ua,
			Metadata:   metaRaw,
			CreatedAt:  time.Now(),
		}); err != nil {
			uc.logger.WarnContext(logCtx, "failed to create audit log",
				slog.String("action", action),
				slog.Any("error", err),
			)
		}
	}()
}
