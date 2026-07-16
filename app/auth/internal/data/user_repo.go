package data

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/puchidemy/puchi-backend/app/auth/internal/biz"
	"github.com/puchidemy/puchi-backend/app/auth/internal/data/sqlc/gen"
)

const pgUniqueViolation = "23505"

// UserRepo wraps sqlc-generated queries for auth.users.
type UserRepo struct {
	q *gen.Queries
}

// NewUserRepo creates a new UserRepo.
func NewUserRepo(d *Data) *UserRepo {
	return &UserRepo{q: gen.New(d.Pool)}
}

// Create inserts a new user and populates the user's ID and timestamps.
// Returns biz.ErrEmailAlreadyExists if the email is already taken.
func (r *UserRepo) Create(ctx context.Context, user *biz.User) error {
	row, err := r.q.CreateUser(ctx, gen.CreateUserParams{
		Email:           user.Email,
		EmailNormalized: user.EmailNormalized,
		PasswordHash:    user.PasswordHash,
		DisplayName:     user.DisplayName,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return biz.ErrEmailAlreadyExists
		}
		return err
	}
	user.ID = uuid.MustParse(row.ID)
	user.CreatedAt = row.CreatedAt
	user.UpdatedAt = row.UpdatedAt
	return nil
}

// GetByEmail retrieves a user by normalized email.
func (r *UserRepo) GetByEmail(ctx context.Context, emailNormalized string) (*biz.User, error) {
	row, err := r.q.GetUserByEmail(ctx, emailNormalized)
	if err != nil {
		return nil, err
	}
	return toUser(row), nil
}

// GetByID retrieves a user by ID.
func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*biz.User, error) {
	row, err := r.q.GetUserByID(ctx, id.String())
	if err != nil {
		return nil, err
	}
	return toUser(row), nil
}

// UpdateLastLogin sets last_login_at to now for the given user.
func (r *UserRepo) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	return r.q.UpdateUserLastLogin(ctx, id.String())
}

// UpdatePassword updates the password hash for the given user.
func (r *UserRepo) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	return r.q.UpdateUserPassword(ctx, gen.UpdateUserPasswordParams{
		ID:           id.String(),
		PasswordHash: &passwordHash,
	})
}

// SetEmailVerified marks the user's email as verified.
func (r *UserRepo) SetEmailVerified(ctx context.Context, id uuid.UUID) error {
	return r.q.SetEmailVerified(ctx, id.String())
}

// SetActive sets the user's active status.
func (r *UserRepo) SetActive(ctx context.Context, id uuid.UUID, active bool) error {
	return r.q.SetUserActive(ctx, gen.SetUserActiveParams{
		ID:       id.String(),
		IsActive: active,
	})
}

// toUser converts a gen.AuthUser (sqlc-generated) to a biz.User.
func toUser(row gen.AuthUser) *biz.User {
	return &biz.User{
		ID:              uuid.MustParse(row.ID),
		Email:           row.Email,
		EmailNormalized: row.EmailNormalized,
		EmailVerified:   row.EmailVerified,
		PasswordHash:    row.PasswordHash,
		DisplayName:     row.DisplayName,
		Locale:          row.Locale,
		IsActive:        row.IsActive,
		IsSuperAdmin:    row.IsSuperAdmin,
		LastLoginAt:     pgTimestamptzToTimePtr(row.LastLoginAt),
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}

// pgTimestamptzToTimePtr converts a pgtype.Timestamptz to *time.Time.
func pgTimestamptzToTimePtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}
