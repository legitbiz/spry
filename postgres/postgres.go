package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/arobson/spry"
	"github.com/arobson/spry/storage"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type QueryData struct {
	ActorName string
}

func queryData(name string) QueryData {
	return QueryData{ActorName: name}
}

type PostgresEventStore struct {
	Pool      pgxpool.Pool
	Templates storage.StringTemplate
}

func (store *PostgresEventStore) Add(actorName string, events []storage.EventRecord) error {
	ctx := context.Background()
	query, _ := store.Templates.Execute(
		"insert_event.sql",
		queryData(actorName),
	)
	batch := pgx.Batch{}
	for _, event := range events {
		data, err := spry.ToJson(event)
		if err != nil {
			return err
		}
		batch.Queue(
			query,
			event.Id,
			event.ActorId,
			data,
			event.CreatedOn,
			event.CreatedByVersion,
		)
	}
	results := store.Pool.SendBatch(ctx, &batch)
	_, _ = results.Exec()
	_, _ = results.Exec()
	err := results.Close()
	return err
}

func (store *PostgresEventStore) FetchSince(actorName string, actorId uuid.UUID, eventUUID uuid.UUID) ([]storage.EventRecord, error) {
	ctx := context.Background()
	query, _ := store.Templates.Execute(
		"select_events_since.sql",
		queryData(actorName),
	)
	rows, err := store.Pool.Query(
		ctx,
		query,
		actorId,
		eventUUID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	records := []storage.EventRecord{}
	for rows.Next() {
		buffer := []byte{}
		rows.Scan(nil, nil, nil, &buffer, nil)
		record, err := spry.FromJson[storage.EventRecord](buffer)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

type PostgresCommandStore struct {
	Pool      pgxpool.Pool
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

type PostgresSnapshotStore struct {
	Pool      pgxpool.Pool
	Templates storage.StringTemplate
}

func (store *PostgresSnapshotStore) Add(actorName string, snapshot storage.Snapshot) error {
	ctx := context.Background()
	query, _ := store.Templates.Execute(
		"insert_snapshot.sql",
		queryData(actorName),
	)
	err := store.Pool.BeginTxFunc(
		ctx,
		pgx.TxOptions{},
		func(t pgx.Tx) error {
			data, err := spry.ToJson(snapshot)
			_, err = t.Exec(
				ctx,
				query,
				snapshot.Id,
				snapshot.ActorId,
				data,
				snapshot.LastCommandId,
				snapshot.LastCommandOn,
				snapshot.LastEventId,
				snapshot.LastEventOn,
				snapshot.Version,
			)
			return err
		},
	)
	return err
}

func (store *PostgresSnapshotStore) Fetch(actorName string, actorId uuid.UUID) (storage.Snapshot, error) {
	ctx := context.Background()
	query, _ := store.Templates.Execute(
		"select_latest_snapshot.sql",
		queryData(actorName),
	)
	rows, err := store.Pool.Query(
		ctx,
		query,
		actorId,
	)
	if err != nil {
		return storage.Snapshot{}, err
	}
	defer rows.Close()
	record := storage.Snapshot{}
	if rows.Next() {
		buffer := []byte{}
		rows.Scan(nil, &buffer, nil, nil, nil, nil, nil)
		record, err = spry.FromJson[storage.Snapshot](buffer)
		if err != nil {
			return record, err
		}
	}
	return record, nil
}

func PostgresStorage() storage.Storage {
	return storage.NewStorage(
		&PostgresMapStore{},
		&PostgresCommandStore{},
		&PostgresEventStore{},
		&PostgresSnapshotStore{},
	)
}

type PostgresMapStore struct {
	Pool      pgxpool.Pool
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

func CreatePostgresStorage(connectionURI string) storage.Storage {
	pool, err := pgxpool.Connect(context.Background(), connectionURI)
	if err != nil {
		fmt.Println("failed to connect to the backing store")
		panic("oh no")
	}

	// load templates
	templates, _ := storage.CreateTemplateFrom(
		"./sql/insert_command.sql",
		"./sql/insert_event.sql",
		"./sql/insert_map.sql",
		"./sql/insert_snapshot.sql",
		"./sql/select_events_since.sql",
		"./sql/select_latest_snapshot.sql",
		"./sql/select_id_by_map.sql",
	)

	return storage.Stores{
		Commands:  &PostgresCommandStore{Templates: *templates, Pool: *pool},
		Events:    &PostgresEventStore{Templates: *templates, Pool: *pool},
		Maps:      &PostgresMapStore{Templates: *templates, Pool: *pool},
		Snapshots: &PostgresSnapshotStore{Templates: *templates, Pool: *pool},
	}
}
