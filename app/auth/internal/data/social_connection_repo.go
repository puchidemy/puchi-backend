package data

import (
	"context"

	"github.com/google/uuid"

	"github.com/puchidemy/puchi-backend/app/auth/internal/biz"
	"github.com/puchidemy/puchi-backend/app/auth/internal/data/sqlc/gen"
)

// SocialConnectionRepo wraps sqlc-generated queries for auth.social_connections.
type SocialConnectionRepo struct {
	q *gen.Queries
}

// NewSocialConnectionRepo creates a new SocialConnectionRepo.
func NewSocialConnectionRepo(d *Data) *SocialConnectionRepo {
	return &SocialConnectionRepo{q: gen.New(d.Pool)}
}

// Create inserts a new social connection and populates the connection's ID and LinkedAt.
func (r *SocialConnectionRepo) Create(ctx context.Context, conn *biz.SocialConnection) error {
	row, err := r.q.CreateSocialConnection(ctx, gen.CreateSocialConnectionParams{
		UserID:         conn.UserID.String(),
		Provider:       conn.Provider,
		ProviderUserID: conn.ProviderUserID,
		ProviderEmail:  stringPtr(conn.ProviderEmail),
		AvatarUrl:      stringPtr(conn.AvatarURL),
	})
	if err != nil {
		return err
	}
	conn.ID = uuid.MustParse(row.ID)
	conn.LinkedAt = row.LinkedAt
	return nil
}

// GetByProvider retrieves a social connection by provider and provider user ID.
func (r *SocialConnectionRepo) GetByProvider(ctx context.Context, provider, providerUserID string) (*biz.SocialConnection, error) {
	row, err := r.q.GetSocialConnection(ctx, gen.GetSocialConnectionParams{
		Provider:       provider,
		ProviderUserID: providerUserID,
	})
	if err != nil {
		return nil, err
	}
	return toBizSocialConn(row), nil
}

// ListByUser returns all social connections for a given user.
func (r *SocialConnectionRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]*biz.SocialConnection, error) {
	rows, err := r.q.ListSocialConnectionsByUser(ctx, userID.String())
	if err != nil {
		return nil, err
	}
	result := make([]*biz.SocialConnection, len(rows))
	for i, row := range rows {
		result[i] = toBizSocialConn(row)
	}
	return result, nil
}

// Delete removes a social connection by ID, scoped to the user.
func (r *SocialConnectionRepo) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return r.q.DeleteSocialConnection(ctx, gen.DeleteSocialConnectionParams{
		ID:     id.String(),
		UserID: userID.String(),
	})
}

// toBizSocialConn converts a gen.AuthSocialConnection (sqlc-generated) to a biz.SocialConnection.
func toBizSocialConn(row gen.AuthSocialConnection) *biz.SocialConnection {
	conn := &biz.SocialConnection{
		ID:             uuid.MustParse(row.ID),
		UserID:         uuid.MustParse(row.UserID),
		Provider:       row.Provider,
		ProviderUserID: row.ProviderUserID,
		LinkedAt:       row.LinkedAt,
	}
	if row.ProviderEmail != nil {
		conn.ProviderEmail = *row.ProviderEmail
	}
	if row.AvatarUrl != nil {
		conn.AvatarURL = *row.AvatarUrl
	}
	return conn
}
