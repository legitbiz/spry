package postgres

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/legitbiz/spry/storage"
)

type PostgresTxProvider struct {
	Pool *pgxpool.Pool
}

func (txp PostgresTxProvider) Commit(ctx context.Context) error {
	tx := storage.GetTx[pgx.Tx](ctx)
	return tx.Commit(ctx)
}

func (txp PostgresTxProvider) GetTransaction(ctx context.Context) (pgx.Tx, error) {
	tx, err := txp.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (txp PostgresTxProvider) Rollback(ctx context.Context) error {
	tx := storage.GetTx[pgx.Tx](ctx)
	return tx.Rollback(ctx)
}
