package data

import (
	"context"
	"encoding/json"
	"net/netip"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/puchidemy/puchi-backend/app/auth/internal/biz"
	"github.com/puchidemy/puchi-backend/app/auth/internal/data/sqlc/gen"
)

// AuditRepo wraps sqlc-generated queries for auth.audit_logs.
type AuditRepo struct {
	q *gen.Queries
}

// NewAuditRepo creates a new AuditRepo.
func NewAuditRepo(d *Data) *AuditRepo {
	return &AuditRepo{q: gen.New(d.Pool)}
}

// Create inserts a new audit log entry.
func (r *AuditRepo) Create(ctx context.Context, log *biz.AuditLog) error {
	var userID pgtype.UUID
	if log.UserID != nil {
		if err := userID.Scan(log.UserID.String()); err != nil {
			return err
		}
	}

	var ipAddr *netip.Addr
	if log.IPAddress != "" {
		addr, err := netip.ParseAddr(log.IPAddress)
		if err == nil {
			ipAddr = &addr
		}
	}

	var metadata []byte
	if log.Metadata != nil {
		var err error
		metadata, err = json.Marshal(log.Metadata)
		if err != nil {
			return err
		}
	}

	_, err := r.q.CreateAuditLog(ctx, gen.CreateAuditLogParams{
		UserID:    userID,
		Action:    log.Action,
		Resource:  stringPtr(log.Resource),
		ResourceID: stringPtr(log.ResourceID),
		IpAddress: ipAddr,
		UserAgent: stringPtr(log.UserAgent),
		Metadata:  metadata,
	})
	return err
}
