package data

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/puchidemy/puchi-backend/app/learn/internal/biz"
	"github.com/puchidemy/puchi-backend/app/learn/internal/data/sqlc/gen"
)

// TransactionManager runs biz callbacks inside a pgx transaction.
type TransactionManager struct {
	pool *pgxpool.Pool
}

// NewTransactionManager creates a TransactionManager.
func NewTransactionManager(pool *pgxpool.Pool) *TransactionManager {
	return &TransactionManager{pool: pool}
}

// InTx begins a transaction, passes tx-scoped repos to fn, and commits on success.
func (t *TransactionManager) InTx(ctx context.Context, fn func(biz.GuestRepoInterface, biz.ProgressRepoInterface) error) error {
	tx, err := t.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := gen.New(tx)
	if err := fn(&GuestRepo{q: q}, &ProgressRepo{q: q}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
