package data

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/puchidemy/puchi-backend/app/core/internal/biz"
	"github.com/puchidemy/puchi-backend/app/core/internal/data/sqlc/gen"
)

// StatsTxManager runs stats callbacks inside a pgx transaction.
type StatsTxManager struct {
	pool *pgxpool.Pool
}

// NewStatsTxManager creates a StatsTxManager.
func NewStatsTxManager(pool *pgxpool.Pool) *StatsTxManager {
	return &StatsTxManager{pool: pool}
}

// InTx begins a transaction, passes a tx-scoped stats repo to fn, and commits on success.
func (t *StatsTxManager) InTx(ctx context.Context, fn func(biz.StatsRepoInterface) error) error {
	tx, err := t.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	repo := &StatsRepo{q: gen.New(tx)}
	if err := fn(repo); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
