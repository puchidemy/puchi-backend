package biz

import (
	"github.com/google/wire"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(NewLearnUsecase)

// LearnUsecase is an empty scaffold usecase. Curriculum and guest-progress
// business logic lands here in later tasks of the learn-service reorg.
type LearnUsecase struct {
	pool *pgxpool.Pool
}

// NewLearnUsecase creates a new LearnUsecase.
func NewLearnUsecase(pool *pgxpool.Pool) *LearnUsecase {
	return &LearnUsecase{pool: pool}
}
