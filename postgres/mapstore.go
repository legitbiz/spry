package postgres

import (
	"context"
	"time"

	"github.com/arobson/spry"
	"github.com/arobson/spry/storage"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type PostgresMapStore struct {
	Pool      *pgxpool.Pool
	Templates storage.StringTemplate
}

func (store *PostgresMapStore) Add(ctx context.Context, actorName string, ids spry.Identifiers, uid uuid.UUID) error {
	query, _ := store.Templates.Execute(
		"insert_map.sql",
		queryData(actorName),
	)
	tx := storage.GetTx[pgx.Tx](ctx)
	id, _ := storage.GetId()
	err := tx.BeginFunc(
		ctx,
		func(t pgx.Tx) error {
			data, err := spry.ToJson(ids)
			if err != nil {
				return err
			}
			_, err = t.Exec(
				ctx,
				query,
				id,
				data,
				uid,
				time.Now(),
			)
			return err
		},
	)
	return err
}

func (store *PostgresMapStore) GetId(ctx context.Context, actorName string, ids spry.Identifiers) (uuid.UUID, error) {
	query, _ := store.Templates.Execute(
		"select_id_by_map.sql",
		queryData(actorName),
	)
	data, err := spry.ToJson(ids)
	if err != nil {
		return uuid.Nil, err
	}

	tx := storage.GetTx[pgx.Tx](ctx)
	rows, err := tx.Query(
		ctx,
		query,
		data,
	)
	if err != nil {
		return uuid.Nil, err
	}
	defer rows.Close()
	uid := uuid.Nil
	if rows.Next() {
		err = rows.Scan(nil, nil, &uid, nil)
		if err != nil {
			return uid, err
		}
	}
	return uid, nil
}
