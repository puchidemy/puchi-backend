package data

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/puchidemy/puchi-backend/app/learn/internal/biz"
	"github.com/puchidemy/puchi-backend/app/learn/internal/data/sqlc/gen"
)

// GuestRepo wraps sqlc guest queries.
type GuestRepo struct {
	q *gen.Queries
}

// NewGuestRepo creates a GuestRepo.
func NewGuestRepo(pool *pgxpool.Pool) *GuestRepo {
	return &GuestRepo{q: gen.New(pool)}
}

// CreateGuest inserts a guest row.
func (r *GuestRepo) CreateGuest(ctx context.Context, id string) error {
	return r.q.CreateGuest(ctx, id)
}

// GetGuestByID returns a guest by ID.
func (r *GuestRepo) GetGuestByID(ctx context.Context, id string) (*gen.LearnGuest, error) {
	row, err := r.q.GetGuestByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// GetGuestByIDForUpdate returns a guest row locked for update within a transaction.
func (r *GuestRepo) GetGuestByIDForUpdate(ctx context.Context, id string) (*gen.LearnGuest, error) {
	row, err := r.q.GetGuestByIDForUpdate(ctx, id)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// ClaimGuest marks a guest as claimed for the given user.
func (r *GuestRepo) ClaimGuest(ctx context.Context, guestID, userID string) error {
	rows, err := r.q.ClaimGuest(ctx, gen.ClaimGuestParams{
		ID:            guestID,
		ClaimedUserID: &userID,
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return biz.ErrGuestAlreadyClaimed
	}
	return nil
}

// mapNoRows converts pgx.ErrNoRows for callers that expect pointer returns.
func mapNoRows[T any](v T, err error) (*T, error) {
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}
	return &v, nil
}
