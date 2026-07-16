package data

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/puchidemy/puchi-backend/app/auth/internal/biz"
	"github.com/puchidemy/puchi-backend/app/auth/internal/data/sqlc/gen"
)

// MagicLinkRepo wraps sqlc-generated queries for auth.magic_links.
type MagicLinkRepo struct {
	q *gen.Queries
}

// NewMagicLinkRepo creates a new MagicLinkRepo.
func NewMagicLinkRepo(d *Data) *MagicLinkRepo {
	return &MagicLinkRepo{q: gen.New(d.Pool)}
}

// Create inserts a new magic link and populates the link's ID and timestamps.
func (r *MagicLinkRepo) Create(ctx context.Context, ml *biz.MagicLink) error {
	var userID pgtype.UUID
	if ml.UserID != uuid.Nil {
		if err := userID.Scan(ml.UserID.String()); err != nil {
			return err
		}
		userID.Valid = true
	}

	var redirectTo *string
	if ml.RedirectTo != "" {
		redirectTo = &ml.RedirectTo
	}

	row, err := r.q.CreateMagicLink(ctx, gen.CreateMagicLinkParams{
		Email:      ml.Email,
		UserID:     userID,
		Token:      ml.Token,
		RedirectTo: redirectTo,
	})
	if err != nil {
		return err
	}
	ml.ID = uuid.MustParse(row.ID)
	ml.ExpiresAt = row.ExpiresAt
	ml.CreatedAt = row.CreatedAt
	return nil
}

// GetByToken retrieves a magic link by its token.
func (r *MagicLinkRepo) GetByToken(ctx context.Context, token string) (*biz.MagicLink, error) {
	row, err := r.q.GetMagicLinkByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	return toMagicLink(row), nil
}

// MarkUsed marks a magic link as used.
func (r *MagicLinkRepo) MarkUsed(ctx context.Context, id uuid.UUID) error {
	return r.q.MarkMagicLinkUsed(ctx, id.String())
}

// toMagicLink converts a gen.AuthMagicLink to a biz.MagicLink.
func toMagicLink(row gen.AuthMagicLink) *biz.MagicLink {
	ml := &biz.MagicLink{
		ID:         uuid.MustParse(row.ID),
		Email:      row.Email,
		Token:      row.Token,
		ExpiresAt:  row.ExpiresAt,
		CreatedAt:  row.CreatedAt,
	}
	if row.RedirectTo != nil {
		ml.RedirectTo = *row.RedirectTo
	}
	if row.UserID.Valid {
		ml.UserID = uuid.UUID(row.UserID.Bytes)
	}
	if row.UsedAt.Valid {
		t := row.UsedAt.Time
		ml.UsedAt = &t
	}
	return ml
}
