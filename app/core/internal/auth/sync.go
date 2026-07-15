package auth

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/puchidemy/puchi-backend/app/core/internal/biz"
	"github.com/puchidemy/puchi-backend/app/core/internal/data/sqlc/gen"
	"github.com/supertokens/supertokens-golang/recipe/emailpassword"
	"github.com/supertokens/supertokens-golang/recipe/thirdparty"
)

// SyncUserRepo defines the repository operations needed by UserSyncer.
type SyncUserRepo interface {
	GetUser(ctx context.Context, id string) (*gen.CoreUser, error)
	CreateUserFromAuth(ctx context.Context, userID, email string) (*gen.CoreUser, error)
}

// UserSyncer handles lazy user creation from Supertokens auth.
// After a session is verified, it ensures the corresponding user exists in the DB.
type UserSyncer struct {
	repo SyncUserRepo
}

// NewUserSyncer creates a new UserSyncer.
func NewUserSyncer(repo SyncUserRepo) *UserSyncer {
	return &UserSyncer{repo: repo}
}

// profileUsecaseAdapter adapts *biz.ProfileUsecase to SyncUserRepo.
type profileUsecaseAdapter struct {
	uc *biz.ProfileUsecase
}

func (a *profileUsecaseAdapter) GetUser(ctx context.Context, id string) (*gen.CoreUser, error) {
	return a.uc.GetProfile(ctx, id)
}

func (a *profileUsecaseAdapter) CreateUserFromAuth(ctx context.Context, userID, email string) (*gen.CoreUser, error) {
	return a.uc.CreateUserFromAuth(ctx, userID, email)
}

// NewUserSyncerFromUsecase creates a UserSyncer from a ProfileUsecase.
func NewUserSyncerFromUsecase(uc *biz.ProfileUsecase) *UserSyncer {
	return &UserSyncer{repo: &profileUsecaseAdapter{uc: uc}}
}

// fetchEmailFromSupertokens retrieves the user's email from the Supertokens core
// by trying all applicable recipe types (email/password and third party).
func fetchEmailFromSupertokens(userID string) string {
	if user, err := emailpassword.GetUserByID(userID); err == nil && user != nil {
		return user.Email
	}
	if user, err := thirdparty.GetUserByID(userID); err == nil && user != nil {
		return user.Email
	}
	return ""
}

// EnsureUserExists checks if user exists in DB; creates if not (lazy creation from Supertokens).
func (s *UserSyncer) EnsureUserExists(ctx context.Context, userID string) error {
	_, err := s.repo.GetUser(ctx, userID)
	if err == nil {
		return nil // user already exists
	}

	slog.Debug("auth sync: user not found in DB, creating from Supertokens",
		"user_id", userID,
	)

	email := fetchEmailFromSupertokens(userID)
	if email == "" {
		slog.Warn("auth sync: could not fetch email from Supertokens, using empty",
			"user_id", userID,
		)
	}

	_, err = s.repo.CreateUserFromAuth(ctx, userID, email)
	if err != nil {
		return fmt.Errorf("lazy create user from auth: %w", err)
	}

	slog.Info("auth sync: user created from Supertokens",
		"user_id", userID,
		"email", email,
	)
	return nil
}
