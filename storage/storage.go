package storage

import (
	"context"
	"reflect"

	"github.com/arobson/spry"
	"github.com/gofrs/uuid"
)

type NoOpTx struct{}

func (tx NoOpTx) Rollback() error {
	return nil
}
func (tx NoOpTx) Commit() error {
	return nil
}

var tx_key = reflect.TypeOf(NoOpTx{})

func GetTx[T any](ctx context.Context) T {
	return ctx.Value(tx_key).(T)
}

type CommandStore interface {
	Add(context.Context, string, CommandRecord) error
}

type EventStore interface {
	Add(context.Context, string, []EventRecord) error
	FetchSince(context.Context, string, uuid.UUID, uuid.UUID) ([]EventRecord, error)
}

type MapStore interface {
	Add(context.Context, string, spry.Identifiers, uuid.UUID) error
	GetId(context.Context, string, spry.Identifiers) (uuid.UUID, error)
}

type SnapshotStore interface {
	Add(context.Context, string, Snapshot) error
	Fetch(context.Context, string, uuid.UUID) (Snapshot, error)
}

type TxProvider[T any] interface {
	GetTransaction(ctx context.Context) (T, error)
}

type Storage interface {
	AddCommand(context.Context, string, CommandRecord) error
	AddEvents(context.Context, string, []EventRecord) error
	AddMap(context.Context, string, spry.Identifiers, uuid.UUID) error
	AddSnapshot(context.Context, string, Snapshot) error
	FetchEventsSince(context.Context, string, uuid.UUID, uuid.UUID) ([]EventRecord, error)
	FetchId(context.Context, string, spry.Identifiers) (uuid.UUID, error)
	FetchLatestSnapshot(context.Context, string, uuid.UUID) (Snapshot, error)
	GetContext(context.Context) (context.Context, error)
}

type Stores[Tx any] struct {
	Commands     CommandStore
	Events       EventStore
	Maps         MapStore
	Snapshots    SnapshotStore
	Transactions TxProvider[Tx]
}

func (storage Stores[Tx]) AddCommand(ctx context.Context, actorName string, command CommandRecord) error {
	return storage.Commands.Add(ctx, actorName, command)
}

func (storage Stores[Tx]) AddEvents(ctx context.Context, actorName string, events []EventRecord) error {
	return storage.Events.Add(ctx, actorName, events)
}

func (storage Stores[Tx]) AddMap(ctx context.Context, actorName string, identifiers spry.Identifiers, uid uuid.UUID) error {
	return storage.Maps.Add(ctx, actorName, identifiers, uid)
}

func (storage Stores[Tx]) AddSnapshot(ctx context.Context, actorName string, snapshot Snapshot) error {
	return storage.Snapshots.Add(ctx, actorName, snapshot)
}

func (storage Stores[Tx]) FetchEventsSince(ctx context.Context, actorName string, actorId uuid.UUID, eventId uuid.UUID) ([]EventRecord, error) {
	return storage.Events.FetchSince(ctx, actorName, actorId, eventId)
}

func (storage Stores[Tx]) FetchId(ctx context.Context, actorName string, identifiers spry.Identifiers) (uuid.UUID, error) {
	return storage.Maps.GetId(ctx, actorName, identifiers)
}

func (storage Stores[Tx]) FetchLatestSnapshot(ctx context.Context, actorName string, actorId uuid.UUID) (Snapshot, error) {
	return storage.Snapshots.Fetch(ctx, actorName, actorId)
}

func (storage Stores[Tx]) GetContext(ctx context.Context) (context.Context, error) {
	newTx, err := storage.Transactions.GetTransaction(ctx)
	if err != nil {
		return nil, err
	}
	return context.WithValue(ctx, tx_key, newTx), nil
}

func NewStorage[Tx any](
	maps MapStore,
	commands CommandStore,
	events EventStore,
	snapshots SnapshotStore,
	txs TxProvider[Tx]) Storage {
	return Stores[Tx]{
		Events:       events,
		Commands:     commands,
		Maps:         maps,
		Snapshots:    snapshots,
		Transactions: txs,
	}
}
