package postgres

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type PostgresTxProvider struct {
	Pool *pgxpool.Pool
}

func (txp PostgresTxProvider) GetTransaction(ctx context.Context) (pgx.Tx, error) {
	tx, err := txp.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	return tx, nil
}
