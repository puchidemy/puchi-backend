package biz

import "context"

// TransactionManager runs callbacks inside a database transaction.
type TransactionManager interface {
	InTx(ctx context.Context, fn func(GuestRepoInterface, ProgressRepoInterface) error) error
}
