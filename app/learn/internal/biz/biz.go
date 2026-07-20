package biz

import (
	"github.com/google/wire"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewLearnUsecase,
)

// LearnUsecase handles learn business logic.
type LearnUsecase struct {
	guestRepo      GuestRepoInterface
	progressRepo   ProgressRepoInterface
	curriculumRepo CurriculumRepoInterface
	storyRepo      StoryRepoInterface
	attemptRepo    AttemptRepoInterface
	publisher      LessonEventPublisher
	tx             TransactionManager
}

// NewLearnUsecase creates a new LearnUsecase.
func NewLearnUsecase(
	guestRepo GuestRepoInterface,
	progressRepo ProgressRepoInterface,
	curriculumRepo CurriculumRepoInterface,
	storyRepo StoryRepoInterface,
	attemptRepo AttemptRepoInterface,
	publisher LessonEventPublisher,
	tx TransactionManager,
) *LearnUsecase {
	return &LearnUsecase{
		guestRepo:      guestRepo,
		progressRepo:   progressRepo,
		curriculumRepo: curriculumRepo,
		storyRepo:      storyRepo,
		attemptRepo:    attemptRepo,
		publisher:      publisher,
		tx:             tx,
	}
}
