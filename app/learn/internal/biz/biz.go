package biz

import (
	"github.com/google/wire"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewLearnUsecase,
	NewNoOpLessonEventPublisher,
	wire.Bind(new(LessonEventPublisher), new(*NoOpLessonEventPublisher)),
)

// LearnUsecase handles learn business logic.
type LearnUsecase struct {
	guestRepo      GuestRepoInterface
	progressRepo   ProgressRepoInterface
	curriculumRepo CurriculumRepoInterface
	attemptRepo    AttemptRepoInterface
	publisher      LessonEventPublisher
	tx             TransactionManager
}

// NewLearnUsecase creates a new LearnUsecase.
func NewLearnUsecase(
	guestRepo GuestRepoInterface,
	progressRepo ProgressRepoInterface,
	curriculumRepo CurriculumRepoInterface,
	attemptRepo AttemptRepoInterface,
	publisher LessonEventPublisher,
	tx TransactionManager,
) *LearnUsecase {
	return &LearnUsecase{
		guestRepo:      guestRepo,
		progressRepo:   progressRepo,
		curriculumRepo: curriculumRepo,
		attemptRepo:    attemptRepo,
		publisher:      publisher,
		tx:             tx,
	}
}
