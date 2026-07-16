package auth

import (
	"context"

	"github.com/puchidemy/puchi-backend/app/core/internal/biz"
	authpkg "github.com/puchidemy/puchi-backend/pkg/auth"
)

// UserSyncer handles lazy user creation from auth.
type UserSyncer = authpkg.UserSyncer

// NewUserSyncer creates a new UserSyncer.
func NewUserSyncer(repo authpkg.SyncUserRepo) *UserSyncer {
	return authpkg.NewUserSyncer(repo)
}

// profileUsecaseAdapter adapts *biz.ProfileUsecase to authpkg.SyncUserRepo.
type profileUsecaseAdapter struct {
	uc *biz.ProfileUsecase
}

func (a *profileUsecaseAdapter) GetUser(ctx context.Context, id string) (interface{}, error) {
	return a.uc.GetProfile(ctx, id)
}

func (a *profileUsecaseAdapter) CreateUserFromAuth(ctx context.Context, userID, email string) (interface{}, error) {
	return a.uc.CreateUserFromAuth(ctx, userID, email)
}

// NewUserSyncerFromUsecase creates a UserSyncer from a ProfileUsecase.
func NewUserSyncerFromUsecase(uc *biz.ProfileUsecase) *UserSyncer {
	return authpkg.NewUserSyncer(&profileUsecaseAdapter{uc: uc})
}
