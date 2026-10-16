package tx

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionalManager struct {
	pool *pgxpool.Pool
}

func NewTransactionalManager(pool *pgxpool.Pool) *TransactionalManager {
	return &TransactionalManager{pool: pool}
}

func (m *TransactionalManager) WithTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	return pgx.BeginTxFunc(ctx, m.pool, pgx.TxOptions{}, fn)
}
