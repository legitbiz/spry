package storage

import (
	"github.com/arobson/spry"
	"github.com/gofrs/uuid"
)

type CommandStore interface {
	Add(CommandRecord) error
}

type EventStore interface {
	Add([]EventRecord) error
	FetchSince(uuid.UUID, uuid.UUID) ([]EventRecord, error)
}

type MapStore interface {
	Add(spry.Identifiers, uuid.UUID) error
	GetId(spry.Identifiers) (uuid.UUID, error)
}

type SnapshotStore interface {
	Add(Snapshot) error
	Fetch(uuid.UUID) (Snapshot, error)
}

type Storage interface {
	AddCommand(CommandRecord) error
	AddEvents([]EventRecord) error
	AddMap(spry.Identifiers, uuid.UUID) error
	AddSnapshot(Snapshot) error
	FetchEventsSince(uuid.UUID, uuid.UUID) ([]EventRecord, error)
	FetchId(spry.Identifiers) (uuid.UUID, error)
	FetchLatestSnapshot(actorId uuid.UUID) (Snapshot, error)
}

type Stores struct {
	Commands  CommandStore
	Events    EventStore
	Maps      MapStore
	Snapshots SnapshotStore
}

func NewStorage(
	maps MapStore,
	commands CommandStore,
	events EventStore,
	snapshots SnapshotStore) Storage {
	return Stores{
		Events:    events,
		Commands:  commands,
		Maps:      maps,
		Snapshots: snapshots,
	}
}

func (storage Stores) AddCommand(command CommandRecord) error {
	return storage.Commands.Add(command)
}

func (storage Stores) AddEvents(events []EventRecord) error {
	return storage.Events.Add(events)
}

func (storage Stores) AddMap(identifiers spry.Identifiers, uid uuid.UUID) error {
	return storage.Maps.Add(identifiers, uid)
}

func (storage Stores) AddSnapshot(snapshot Snapshot) error {
	return storage.Snapshots.Add(snapshot)
}

func (storage Stores) FetchEventsSince(actorId uuid.UUID, eventId uuid.UUID) ([]EventRecord, error) {
	return storage.Events.FetchSince(actorId, eventId)
}

func (storage Stores) FetchId(identifiers spry.Identifiers) (uuid.UUID, error) {
	return storage.Maps.GetId(identifiers)
}

func (storage Stores) FetchLatestSnapshot(actorId uuid.UUID) (Snapshot, error) {
	return storage.Snapshots.Fetch(actorId)
}
