package data

import (
	"context"
	"net/netip"

	"github.com/google/uuid"

	"github.com/puchidemy/puchi-backend/app/auth/internal/biz"
	"github.com/puchidemy/puchi-backend/app/auth/internal/data/sqlc/gen"
)

// SessionRepo wraps sqlc-generated queries for auth.sessions.
type SessionRepo struct {
	q *gen.Queries
}

// NewSessionRepo creates a new SessionRepo.
func NewSessionRepo(d *Data) *SessionRepo {
	return &SessionRepo{q: gen.New(d.Pool)}
}

// Create inserts a new session and populates the session's ID and timestamps.
func (r *SessionRepo) Create(ctx context.Context, session *biz.Session) error {
	var ipAddr *netip.Addr
	if session.IPAddress != "" {
		addr, err := netip.ParseAddr(session.IPAddress)
		if err == nil {
			ipAddr = &addr
		}
	}

	row, err := r.q.CreateSession(ctx, gen.CreateSessionParams{
		UserID:           session.UserID.String(),
		RefreshTokenHash: session.RefreshTokenHash,
		TokenFamily:      session.TokenFamily.String(),
		ChildNumber:      int32(session.ChildNumber),
		IpAddress:        ipAddr,
		UserAgent:        stringPtr(session.UserAgent),
		DeviceName:       stringPtr(session.DeviceName),
		DeviceType:       stringPtr(session.DeviceType),
		Os:               stringPtr(session.OS),
		ExpiresAt:        session.ExpiresAt,
	})
	if err != nil {
		return err
	}
	session.ID = uuid.MustParse(row.ID)
	session.CreatedAt = row.CreatedAt
	return nil
}

// GetByRefreshTokenHash retrieves a session by its refresh token hash.
func (r *SessionRepo) GetByRefreshTokenHash(ctx context.Context, hash string) (*biz.Session, error) {
	row, err := r.q.GetSessionByHash(ctx, hash)
	if err != nil {
		return nil, err
	}
	return toSession(row), nil
}

// Revoke marks a session as inactive.
func (r *SessionRepo) Revoke(ctx context.Context, id uuid.UUID) error {
	return r.q.RevokeSession(ctx, id.String())
}

// RevokeFamily revokes all active sessions in a token family.
func (r *SessionRepo) RevokeFamily(ctx context.Context, family uuid.UUID) error {
	return r.q.RevokeFamily(ctx, family.String())
}

// RevokeAllForUser revokes all active sessions for a user.
func (r *SessionRepo) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	return r.q.RevokeAllForUser(ctx, userID.String())
}

// ListByUser returns all active sessions for a user.
func (r *SessionRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]*biz.Session, error) {
	rows, err := r.q.ListSessionsByUser(ctx, userID.String())
	if err != nil {
		return nil, err
	}
	sessions := make([]*biz.Session, len(rows))
	for i, row := range rows {
		sessions[i] = toSession(row)
	}
	return sessions, nil
}

// HasActiveInFamily checks if there are active sessions in the given token family,
// excluding the specified session ID.
func (r *SessionRepo) HasActiveInFamily(ctx context.Context, family uuid.UUID, excludeSessionID uuid.UUID) (bool, error) {
	return r.q.HasActiveSessionsInFamily(ctx, gen.HasActiveSessionsInFamilyParams{
		TokenFamily: family.String(),
		ID:          excludeSessionID.String(),
	})
}

// toSession converts a gen.AuthSession (sqlc-generated) to a biz.Session.
func toSession(row gen.AuthSession) *biz.Session {
	s := &biz.Session{
		ID:               uuid.MustParse(row.ID),
		UserID:           uuid.MustParse(row.UserID),
		RefreshTokenHash: row.RefreshTokenHash,
		TokenFamily:      uuid.MustParse(row.TokenFamily),
		ChildNumber:      int(row.ChildNumber),
		IsActive:         row.IsActive,
		ExpiresAt:        row.ExpiresAt,
		LastUsedAt:       row.LastUsedAt,
		CreatedAt:        row.CreatedAt,
	}
	if row.IpAddress != nil {
		s.IPAddress = row.IpAddress.String()
	}
	if row.UserAgent != nil {
		s.UserAgent = *row.UserAgent
	}
	if row.DeviceName != nil {
		s.DeviceName = *row.DeviceName
	}
	if row.DeviceType != nil {
		s.DeviceType = *row.DeviceType
	}
	if row.Os != nil {
		s.OS = *row.Os
	}
	if row.Location != nil {
		s.Location = *row.Location
	}
	if row.RevokedAt.Valid {
		t := row.RevokedAt.Time
		s.RevokedAt = &t
	}
	return s
}

// stringPtr returns a pointer to s if non-empty, nil otherwise.
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
