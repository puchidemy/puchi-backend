package data

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/puchidemy/puchi-backend/app/core/internal/data/sqlc/gen"
)

// UserRepo wraps sqlc-generated queries for core.users.
type UserRepo struct {
	q *gen.Queries
}

// NewUserRepo creates a new UserRepo.
func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{q: gen.New(pool)}
}

// CreateUser inserts a new user and returns it.
func (r *UserRepo) CreateUser(ctx context.Context, id, username, email, firstName, lastName string) (*gen.CoreUser, error) {
	row, err := r.q.CreateUser(ctx, gen.CreateUserParams{
		ID:        id,
		Username:  username,
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
	})
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// GetUser retrieves a user by ID.
func (r *UserRepo) GetUser(ctx context.Context, id string) (*gen.CoreUser, error) {
	row, err := r.q.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// GetUserByEmail retrieves a user by email.
func (r *UserRepo) GetUserByEmail(ctx context.Context, email string) (*gen.CoreUser, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// UpdateUser updates a user and returns the updated row.
func (r *UserRepo) UpdateUser(ctx context.Context, id, firstName, lastName, username string, bio, avatarKey *string, ageRange string) (*gen.CoreUser, error) {
	row, err := r.q.UpdateUser(ctx, gen.UpdateUserParams{
		ID:        id,
		FirstName: firstName,
		LastName:  lastName,
		Username:  username,
		Bio:       bio,
		AvatarKey: avatarKey,
		AgeRange:  ageRange,
	})
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// UsernameExists checks whether a username is already taken.
func (r *UserRepo) UsernameExists(ctx context.Context, username string) (bool, error) {
	row, err := r.q.UsernameExists(ctx, username)
	if err != nil {
		return false, err
	}
	return row, nil
}

// GetUserByUsername retrieves a user by username.
func (r *UserRepo) GetUserByUsername(ctx context.Context, username string) (*gen.CoreUser, error) {
	row, err := r.q.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// UpdateOnboardingInfo updates user's first_name, last_name, age_range and sets onboarding_completed=true.
func (r *UserRepo) UpdateOnboardingInfo(ctx context.Context, id, firstName, lastName, ageRange string) (*gen.CoreUser, error) {
	row, err := r.q.UpdateOnboardingInfo(ctx, gen.UpdateOnboardingInfoParams{
		ID:        id,
		FirstName: firstName,
		LastName:  lastName,
		AgeRange:  ageRange,
	})
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// UpsertUserOnboarding inserts or updates onboarding answers.
func (r *UserRepo) UpsertUserOnboarding(ctx context.Context, userID, howHeard, whyLearn, level string) error {
	_, err := r.q.UpsertUserOnboarding(ctx, gen.UpsertUserOnboardingParams{
		UserID:   userID,
		HowHeard: howHeard,
		WhyLearn: whyLearn,
		Level:    level,
	})
	return err
}
