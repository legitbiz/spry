package storage

import (
	"github.com/arobson/spry"
	"github.com/gofrs/uuid"
)

type CommandStore interface {
	Add(string, CommandRecord) error
}

type EventStore interface {
	Add(string, []EventRecord) error
	FetchSince(string, uuid.UUID, uuid.UUID) ([]EventRecord, error)
}

type MapStore interface {
	Add(string, spry.Identifiers, uuid.UUID) error
	GetId(string, spry.Identifiers) (uuid.UUID, error)
}

type SnapshotStore interface {
	Add(string, Snapshot) error
	Fetch(string, uuid.UUID) (Snapshot, error)
}

type Storage interface {
	AddCommand(string, CommandRecord) error
	AddEvents(string, []EventRecord) error
	AddMap(string, spry.Identifiers, uuid.UUID) error
	AddSnapshot(string, Snapshot) error
	FetchEventsSince(string, uuid.UUID, uuid.UUID) ([]EventRecord, error)
	FetchId(string, spry.Identifiers) (uuid.UUID, error)
	FetchLatestSnapshot(string, uuid.UUID) (Snapshot, error)
}

type Stores struct {
	Commands  CommandStore
	Events    EventStore
	Maps      MapStore
	Snapshots SnapshotStore
}

func (storage Stores) AddCommand(actorName string, command CommandRecord) error {
	return storage.Commands.Add(actorName, command)
}

func (storage Stores) AddEvents(actorName string, events []EventRecord) error {
	return storage.Events.Add(actorName, events)
}

func (storage Stores) AddMap(actorName string, identifiers spry.Identifiers, uid uuid.UUID) error {
	return storage.Maps.Add(actorName, identifiers, uid)
}

func (storage Stores) AddSnapshot(actorName string, snapshot Snapshot) error {
	return storage.Snapshots.Add(actorName, snapshot)
}

func (storage Stores) FetchEventsSince(actorName string, actorId uuid.UUID, eventId uuid.UUID) ([]EventRecord, error) {
	return storage.Events.FetchSince(actorName, actorId, eventId)
}

func (storage Stores) FetchId(actorName string, identifiers spry.Identifiers) (uuid.UUID, error) {
	return storage.Maps.GetId(actorName, identifiers)
}

func (storage Stores) FetchLatestSnapshot(actorName string, actorId uuid.UUID) (Snapshot, error) {
	return storage.Snapshots.Fetch(actorName, actorId)
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
