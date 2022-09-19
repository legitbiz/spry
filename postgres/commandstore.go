package postgres

import (
	"context"

	"github.com/arobson/spry"
	"github.com/arobson/spry/storage"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type PostgresCommandStore struct {
	Pool      *pgxpool.Pool
	Templates storage.StringTemplate
}

func (store *PostgresCommandStore) Add(actorName string, command storage.CommandRecord) error {
	ctx := context.Background()
	query, _ := store.Templates.Execute(
		"insert_command.sql",
		queryData(actorName),
	)
	err := store.Pool.BeginTxFunc(
		ctx,
		pgx.TxOptions{},
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
