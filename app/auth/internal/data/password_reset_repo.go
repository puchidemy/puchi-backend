package data

import (
	"context"

	"github.com/google/uuid"

	"github.com/puchidemy/puchi-backend/app/auth/internal/biz"
	"github.com/puchidemy/puchi-backend/app/auth/internal/data/sqlc/gen"
)

// PasswordResetTokenRepo wraps sqlc-generated queries for auth.password_reset_tokens.
type PasswordResetTokenRepo struct {
	q *gen.Queries
}

// NewPasswordResetTokenRepo creates a new PasswordResetTokenRepo.
func NewPasswordResetTokenRepo(d *Data) *PasswordResetTokenRepo {
	return &PasswordResetTokenRepo{q: gen.New(d.Pool)}
}

// Create inserts a new password reset token.
func (r *PasswordResetTokenRepo) Create(ctx context.Context, token *biz.PasswordResetToken) error {
	row, err := r.q.CreatePasswordResetToken(ctx, gen.CreatePasswordResetTokenParams{
		UserID: token.UserID.String(),
		Token:  token.Token,
	})
	if err != nil {
		return err
	}
	token.ID = uuid.MustParse(row.ID)
	token.ExpiresAt = row.ExpiresAt
	token.CreatedAt = row.CreatedAt
	return nil
}

// GetByToken retrieves a valid (unused, not expired) password reset token.
func (r *PasswordResetTokenRepo) GetByToken(ctx context.Context, tokenHash string) (*biz.PasswordResetToken, error) {
	row, err := r.q.GetPasswordResetToken(ctx, tokenHash)
	if err != nil {
		return nil, err
	}
	return toPasswordResetToken(row), nil
}

// MarkUsed marks a password reset token as used.
func (r *PasswordResetTokenRepo) MarkUsed(ctx context.Context, id uuid.UUID) error {
	return r.q.MarkPasswordResetTokenUsed(ctx, id.String())
}

// toPasswordResetToken converts a gen.AuthPasswordResetToken to a biz.PasswordResetToken.
func toPasswordResetToken(row gen.AuthPasswordResetToken) *biz.PasswordResetToken {
	t := &biz.PasswordResetToken{
		ID:        uuid.MustParse(row.ID),
		UserID:    uuid.MustParse(row.UserID),
		Token:     row.Token,
		ExpiresAt: row.ExpiresAt,
		CreatedAt: row.CreatedAt,
	}
	if row.UsedAt.Valid {
		ut := row.UsedAt.Time
		t.UsedAt = &ut
	}
	return t
}

var _ biz.PasswordResetTokenRepo = (*PasswordResetTokenRepo)(nil)
