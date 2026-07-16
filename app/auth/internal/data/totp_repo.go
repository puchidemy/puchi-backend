package data

import (
	"context"

	"github.com/google/uuid"

	"github.com/puchidemy/puchi-backend/app/auth/internal/biz"
	"github.com/puchidemy/puchi-backend/app/auth/internal/data/sqlc/gen"
)

type TOTPRepo struct {
	q *gen.Queries
}

func NewTOTPRepo(d *Data) *TOTPRepo {
	return &TOTPRepo{q: gen.New(d.Pool)}
}

func (r *TOTPRepo) Upsert(ctx context.Context, totp *biz.TOTPSecret) error {
	row, err := r.q.UpsertTOTPSecret(ctx, gen.UpsertTOTPSecretParams{
		UserID:          totp.UserID.String(),
		EncryptedSecret: totp.EncryptedSecret,
		EncryptedCodes:  totp.EncryptedCodes,
		IsEnabled:       totp.IsEnabled,
	})
	if err != nil {
		return err
	}
	totp.ID = uuid.MustParse(row.ID)
	totp.CreatedAt = row.CreatedAt
	totp.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *TOTPRepo) GetByUser(ctx context.Context, userID uuid.UUID) (*biz.TOTPSecret, error) {
	row, err := r.q.GetTOTPSecret(ctx, userID.String())
	if err != nil {
		return nil, err
	}
	return &biz.TOTPSecret{
		ID:              uuid.MustParse(row.ID),
		UserID:          uuid.MustParse(row.UserID),
		EncryptedSecret: row.EncryptedSecret,
		EncryptedCodes:  row.EncryptedCodes,
		IsEnabled:       row.IsEnabled,
		VerifiedAt:      pgTimestamptzToTimePtr(row.VerifiedAt),
		LastUsedAt:      pgTimestamptzToTimePtr(row.LastUsedAt),
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}, nil
}

func (r *TOTPRepo) Enable(ctx context.Context, userID uuid.UUID) error {
	return r.q.EnableTOTP(ctx, userID.String())
}

func (r *TOTPRepo) Disable(ctx context.Context, userID uuid.UUID) error {
	return r.q.DisableTOTP(ctx, userID.String())
}

