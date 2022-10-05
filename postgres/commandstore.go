package postgres

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/legitbiz/spry"
	"github.com/legitbiz/spry/storage"
)

type PostgresCommandStore struct {
	Pool      *pgxpool.Pool
	Templates storage.StringTemplate
}

func (store *PostgresCommandStore) Add(ctx context.Context, actorName string, command storage.CommandRecord) error {
	query, _ := store.Templates.Execute(
		"insert_command.sql",
		queryData(actorName),
	)

	tx := storage.GetTx[pgx.Tx](ctx)

	err := tx.BeginFunc(
		ctx,
		func(t pgx.Tx) error {
			data, err := spry.ToJson(command)
			if err != nil {
				return err
			}
			_, err = t.Exec(
				ctx,
				query,
				command.Id,
				command.HandledBy,
				data,
				command.CreatedOn,
				command.HandledVersion,
			)
			return err
		},
	)
	return err
}
