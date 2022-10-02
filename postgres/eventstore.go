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

func (store *PostgresEventStore) Add(ctx context.Context, events []storage.EventRecord) error {
	tx := storage.GetTx[pgx.Tx](ctx)
	batch := pgx.Batch{}
	for _, event := range events {
		data, err := spry.ToJson(event)
		if err != nil {
			return err
		}
		query, _ := store.Templates.Execute(
			"insert_event.sql",
			queryData(event.ActorType),
		)
		batch.Queue(
			query,
			event.Id,
			event.ActorId,
			data,
			event.CreatedOn,
			event.CreatedByVersion,
		)
	}

	results := tx.SendBatch(ctx, &batch)
	_, _ = results.Exec()
	err := results.Close()
	return err
}

func (store *InMemoryEventStore) FetchAggregatedSince(
	ctx context.Context,
	ids spry.AggregateIdMap,
	eventUUID uuid.UUID,
	types storage.TypeMap) ([]storage.EventRecord, error) {

	return []storage.EventRecord, nil
}

func (store *PostgresEventStore) FetchSince(
	ctx context.Context,
	actorName string,
	actorId uuid.UUID,
	eventUUID uuid.UUID,
	types storage.TypeMap) ([]storage.EventRecord, error) {
	query, _ := store.Templates.Execute(
		"select_events_since.sql",
		queryData(actorName),
	)
	tx := storage.GetTx[pgx.Tx](ctx)
	rows, err := tx.Query(
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
		record.Data, err = types.AsEvent(record.Type, record.Data)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}
