package auth

import (
	"context"
	"fmt"
	"log/slog"
)

// SyncUserRepo defines the repository operations needed by UserSyncer.
// Return types are interface{} since the concrete user type is service-specific.
type SyncUserRepo interface {
	GetUser(ctx context.Context, id string) (interface{}, error)
	CreateUserFromAuth(ctx context.Context, userID, email string) (interface{}, error)
}

// UserSyncer handles lazy user creation from auth.
// After a JWT is verified, it ensures the corresponding user exists in the DB.
type UserSyncer struct {
	repo SyncUserRepo
}

// NewUserSyncer creates a new UserSyncer.
func NewUserSyncer(repo SyncUserRepo) *UserSyncer {
	return &UserSyncer{repo: repo}
}

// EnsureUserExists checks if user exists in DB; creates if not (lazy creation from auth).
// The email parameter is optional — if empty, the user is created with a placeholder.
func (s *UserSyncer) EnsureUserExists(ctx context.Context, userID string, email string) error {
	_, err := s.repo.GetUser(ctx, userID)
	if err == nil {
		return nil // user already exists
	}

	slog.Debug("auth sync: user not found in DB, creating from auth",
		"user_id", userID,
	)

	if email == "" {
		slog.Warn("auth sync: no email available for user, using empty",
			"user_id", userID,
		)
	}

	_, err = s.repo.CreateUserFromAuth(ctx, userID, email)
	if err != nil {
		return fmt.Errorf("lazy create user from auth: %w", err)
	}

	slog.Info("auth sync: user created from auth",
		"user_id", userID,
		"email", email,
	)
	return nil
}
