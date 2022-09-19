package postgres

import (
	"context"

	"github.com/arobson/spry"
	"github.com/arobson/spry/storage"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type PostgresEventStore struct {
	Pool      *pgxpool.Pool
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
		err = rows.Scan(nil, nil, nil, &buffer, nil)
		if err != nil {
			return nil, err
		}
		record, err := spry.FromJson[storage.EventRecord](buffer)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}
