package biz

import "context"

// StatsTxManagerInterface runs stats repo callbacks inside a database transaction.
type StatsTxManagerInterface interface {
	InTx(ctx context.Context, fn func(StatsRepoInterface) error) error
}
