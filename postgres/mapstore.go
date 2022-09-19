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

func (store *PostgresMapStore) Add(actorName string, ids spry.Identifiers, uid uuid.UUID) error {
	ctx := context.Background()
	query, _ := store.Templates.Execute(
		"insert_map.sql",
		queryData(actorName),
	)
	id, _ := storage.GetId()
	err := store.Pool.BeginTxFunc(
		ctx,
		pgx.TxOptions{},
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

func (store *PostgresMapStore) GetId(actorName string, ids spry.Identifiers) (uuid.UUID, error) {
	ctx := context.Background()
	query, _ := store.Templates.Execute(
		"select_id_by_map.sql",
		queryData(actorName),
	)
	data, err := spry.ToJson(ids)
	if err != nil {
		return uuid.Nil, err
	}
	rows, err := store.Pool.Query(
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
