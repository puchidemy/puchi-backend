package data

import (
	"context"

	"github.com/google/uuid"

	"github.com/puchidemy/puchi-backend/app/auth/internal/biz"
	"github.com/puchidemy/puchi-backend/app/auth/internal/data/sqlc/gen"
)

// EmailVerificationRepo wraps sqlc-generated queries for auth.email_verifications.
type EmailVerificationRepo struct {
	q *gen.Queries
}

// NewEmailVerificationRepo creates a new EmailVerificationRepo.
func NewEmailVerificationRepo(d *Data) *EmailVerificationRepo {
	return &EmailVerificationRepo{q: gen.New(d.Pool)}
}

// Create inserts a new email verification record.
func (r *EmailVerificationRepo) Create(ctx context.Context, ev *biz.EmailVerification) error {
	row, err := r.q.CreateEmailVerification(ctx, gen.CreateEmailVerificationParams{
		UserID: ev.UserID.String(),
		Token:  ev.CodeHash,
		Email:  ev.Email,
	})
	if err != nil {
		return err
	}
	ev.ID = uuid.MustParse(row.ID)
	ev.ExpiresAt = row.ExpiresAt
	ev.CreatedAt = row.CreatedAt
	return nil
}

// GetByUserAndHash retrieves a valid (unused, not expired) email verification
// record for the given user and code hash.
func (r *EmailVerificationRepo) GetByUserAndHash(ctx context.Context, userID uuid.UUID, codeHash string) (*biz.EmailVerification, error) {
	row, err := r.q.GetEmailVerificationByUserAndHash(ctx, gen.GetEmailVerificationByUserAndHashParams{
		UserID: userID.String(),
		Token:  codeHash,
	})
	if err != nil {
		return nil, err
	}
	return toEmailVerification(row), nil
}

// MarkVerified marks an email verification record as used.
func (r *EmailVerificationRepo) MarkVerified(ctx context.Context, id uuid.UUID) error {
	return r.q.MarkEmailVerificationUsed(ctx, id.String())
}

// toEmailVerification converts a gen.AuthEmailVerification to a biz.EmailVerification.
func toEmailVerification(row gen.AuthEmailVerification) *biz.EmailVerification {
	ev := &biz.EmailVerification{
		ID:        uuid.MustParse(row.ID),
		UserID:    uuid.MustParse(row.UserID),
		CodeHash:  row.Token,
		Email:     row.Email,
		ExpiresAt: row.ExpiresAt,
		CreatedAt: row.CreatedAt,
	}
	if row.UsedAt.Valid {
		t := row.UsedAt.Time
		ev.VerifiedAt = &t
	}
	return ev
}

var _ biz.EmailVerificationRepo = (*EmailVerificationRepo)(nil)
