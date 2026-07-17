package biz

import (
	"github.com/google/wire"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(NewLearnUsecase)

// LearnUsecase handles learn business logic.
type LearnUsecase struct {
	guestRepo    GuestRepoInterface
	progressRepo ProgressRepoInterface
}

// NewLearnUsecase creates a new LearnUsecase.
func NewLearnUsecase(guestRepo GuestRepoInterface, progressRepo ProgressRepoInterface) *LearnUsecase {
	return &LearnUsecase{
		guestRepo:    guestRepo,
		progressRepo: progressRepo,
	}
}
