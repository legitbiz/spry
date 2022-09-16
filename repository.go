package spry

import (
	"reflect"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
)

func GetId() (uuid.UUID, error) {
	UUID, err := uuid.NewV6()
	if err != nil {
		return uuid.Nil, err
	}
	return UUID, nil
}

type Results[T Actor[T]] struct {
	Original T
	Modified T
	Events   []Event
	Errors   []error
}

type CommandStore interface {
	Add(CommandRecord) error
}

type EventStore interface {
	Add([]EventRecord) error
	FetchSince(uuid.UUID, uuid.UUID) ([]EventRecord, error)
}

type MapStore interface {
	Add(Identifiers, uuid.UUID) error
	GetId(Identifiers) (uuid.UUID, error)
}

type SnapshotStore interface {
	Add(Snapshot) error
	Fetch(uuid.UUID) (Snapshot, error)
}

type Storage struct {
	Commands  CommandStore
	Events    EventStore
	Maps      MapStore
	Snapshots SnapshotStore
}

func NewStorage(
	idMaps MapStore,
	commands CommandStore,
	events EventStore,
	snapshots SnapshotStore) Storage {
	return Storage{
		Events:    events,
		Commands:  commands,
		Snapshots: snapshots,
	}
}

func (storage Storage) AddCommand(command CommandRecord) error {
	return storage.Commands.Add(command)
}

func (storage Storage) AddEvents(events []EventRecord) error {
	return storage.Events.Add(events)
}

func (storage Storage) AddMap(identifiers Identifiers, uid uuid.UUID) error {
	return storage.Maps.Add(identifiers, uid)
}

func (storage Storage) AddSnapshot(snapshot Snapshot) error {
	return storage.Snapshots.Add(snapshot)
}

func (storage Storage) FetchEventsSince(actorId uuid.UUID, eventId uuid.UUID) ([]EventRecord, error) {
	return storage.Events.FetchSince(actorId, eventId)
}

func (storage Storage) FetchId(identifiers Identifiers) (uuid.UUID, error) {
	return storage.Maps.GetId(identifiers)
}

func (storage Storage) FetchLatestSnapshot(actorId uuid.UUID) (Snapshot, error) {
	return storage.Snapshots.Fetch(actorId)
}

func IdMapToString(ids Identifiers) (string, error) {
	bytes, err := ToJson(ids)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

type InMemoryMapStore struct {
	IdMap map[string]uuid.UUID
}

func (maps InMemoryMapStore) Add(ids Identifiers, uid uuid.UUID) error {
	key, _ := IdMapToString(ids)
	maps.IdMap[key] = uid
	return nil
}

func (maps InMemoryMapStore) GetId(ids Identifiers) (uuid.UUID, error) {
	key, _ := IdMapToString(ids)
	uid := maps.IdMap[key]
	if uid != uuid.Nil {
		return uuid.Nil, errors.Errorf("no uuid exists for map %s", ids)
	}
	return uid, nil
}

type InMemoryEventStore struct {
	Events map[uuid.UUID][]EventRecord
}

func (store *InMemoryEventStore) Add(events []EventRecord) error {
	for _, event := range events {
		actorId := event.CreatedById
		if stored, ok := store.Events[actorId]; ok {
			store.Events[actorId] = append(stored, event)
		} else {
			store.Events[actorId] = []EventRecord{event}
		}
	}
	return nil
}

func (store *InMemoryEventStore) FetchSince(actorId uuid.UUID, eventUUID uuid.UUID) ([]EventRecord, error) {
	if stored, ok := store.Events[actorId]; ok {
		return stored, nil
	} else {
		return []EventRecord{}, nil
	}
}

type InMemoryCommandStore struct {
	Commands map[uuid.UUID][]CommandRecord
}

func (store *InMemoryCommandStore) Add(command CommandRecord) error {
	actorId := command.HandledBy
	if stored, ok := store.Commands[actorId]; ok {
		store.Commands[actorId] = append(stored, command)
	} else {
		store.Commands[actorId] = []CommandRecord{command}
	}
	return nil
}

type InMemorySnapshotStore struct {
	Snapshots map[uuid.UUID][]Snapshot
}

func (store *InMemorySnapshotStore) Add(snapshot Snapshot) error {
	actorId := snapshot.Id
	if stored, ok := store.Snapshots[actorId]; ok {
		store.Snapshots[actorId] = append(stored, snapshot)
	} else {
		store.Snapshots[actorId] = []Snapshot{snapshot}
	}
	return nil
}

func (store *InMemorySnapshotStore) Fetch(actorId uuid.UUID) (Snapshot, error) {
	if stored, ok := store.Snapshots[actorId]; ok {
		return stored[len(stored)-1]
	}
	return Snapshot{}
}

type Repository[T Actor[T]] struct {
	ActorType reflect.Type
	ActorName string
	Storage   Storage
}

func (repository Repository[T]) getEmpty() T {
	return *new(T)
}

// A side-effect free way of applying events to an actor instance
func (repository Repository[T]) Apply(events []Event, actor T) T {
	var modified T = actor
	for _, event := range events {
		event.Apply(&modified)
	}
	return modified
}

func (repository Repository[T]) Fetch(ids Identifiers) (T, error) {
	// check for a registered actor id for the identifiers
	empty := repository.getEmpty()
	baseline, err := NewSnapshot(empty)
	events := []EventRecord{}
	if err != nil {
		return empty, err
	}
	actorId, err := repository.Storage.FetchId(ids)
	if err != nil {
		return empty, err
	}

	// check for the latest snapshot available
	if actorId != uuid.Nil {
		baseline, err = repository.Storage.FetchLatestSnapshot(actorId)
		if err != nil {
			return empty, err
		}
	}

	// check for all events since the latest snapshot
	if actorId != uuid.Nil {
		eventId := baseline.LastEventId
		events, err = repository.Storage.FetchEventsSince(
			actorId,
			eventId,
		)
		if err != nil {
			return empty, err
		}
	}

	es := make([]Event, len(events))
	for i, record := range events {
		es[i] = record.Data.(Event)
	}
	actor := baseline.Data.(T)
	// apply events to snapshot
	next := repository.Apply(es, actor)

	// update snapshot record

	// return actor
	return next, nil
}

func (repository Repository[T]) Handle(command Command) Results[T] {
	identifiers := command.GetIdentifiers()
	empty, err := repository.Fetch(identifiers)
	if err != nil {
		return Results[T]{
			Errors: []error{err},
		}
	}

	events, errors := command.Handle(empty)
	next := repository.Apply(events, empty)
	// next := empty.Apply(events)
	return Results[T]{
		Original: empty,
		Modified: next,
		Events:   events,
		Errors:   errors,
	}
}

func GetRepositoryFor[T Actor[T]](storage Storage) Repository[T] {
	actorType := reflect.TypeOf(*new(T))
	actorName := actorType.Name()
	return Repository[T]{
		ActorType: actorType,
		ActorName: actorName,
		Storage:   storage,
	}
}
