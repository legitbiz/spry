package spry

import (
	"reflect"
	"time"

	"github.com/gofrs/uuid"
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
	maps MapStore,
	commands CommandStore,
	events EventStore,
	snapshots SnapshotStore) Storage {
	return Storage{
		Events:    events,
		Commands:  commands,
		Maps:      maps,
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

func (maps *InMemoryMapStore) Add(ids Identifiers, uid uuid.UUID) error {
	if maps.IdMap == nil {
		maps.IdMap = map[string]uuid.UUID{}
	}
	key, _ := IdMapToString(ids)
	maps.IdMap[key] = uid
	return nil
}

func (maps *InMemoryMapStore) GetId(ids Identifiers) (uuid.UUID, error) {
	if maps.IdMap == nil {
		maps.IdMap = map[string]uuid.UUID{}
	}
	key, _ := IdMapToString(ids)
	uid := maps.IdMap[key]
	if uid == uuid.Nil {
		return uuid.Nil, nil
	}
	return uid, nil
}

type InMemoryEventStore struct {
	Events map[uuid.UUID][]EventRecord
}

func (store *InMemoryEventStore) Add(events []EventRecord) error {
	if store.Events == nil {
		store.Events = map[uuid.UUID][]EventRecord{}
	}
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
	if store.Events == nil {
		store.Events = map[uuid.UUID][]EventRecord{}
	}
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
	if store.Commands == nil {
		store.Commands = map[uuid.UUID][]CommandRecord{}
	}
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
	if store.Snapshots == nil {
		store.Snapshots = map[uuid.UUID][]Snapshot{}
	}
	actorId := snapshot.ActorId
	if stored, ok := store.Snapshots[actorId]; ok {
		store.Snapshots[actorId] = append(stored, snapshot)
	} else {
		store.Snapshots[actorId] = []Snapshot{snapshot}
	}
	return nil
}

func (store *InMemorySnapshotStore) Fetch(actorId uuid.UUID) (Snapshot, error) {
	if store.Snapshots == nil {
		store.Snapshots = map[uuid.UUID][]Snapshot{}
	}
	if stored, ok := store.Snapshots[actorId]; ok {
		return stored[len(stored)-1], nil
	}
	return Snapshot{}, nil
}

func InMemoryStorage() Storage {
	return NewStorage(
		&InMemoryMapStore{},
		&InMemoryCommandStore{},
		&InMemoryEventStore{},
		&InMemorySnapshotStore{},
	)
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

func (repository Repository[T]) fetch(ids Identifiers) (Snapshot, error) {
	// create an empty actor instance and empty snapshot
	empty := repository.getEmpty()
	snapshot, err := NewSnapshot(empty)
	events := []EventRecord{}
	if err != nil {
		return snapshot, err
	}

	// fetch the actor id from the identifier
	actorId, err := repository.Storage.FetchId(ids)
	if err != nil {
		return snapshot, err
	}

	// check for the latest snapshot available
	if actorId != uuid.Nil {
		latest, err := repository.Storage.FetchLatestSnapshot(actorId)
		if err != nil {
			return snapshot, err
		}
		if latest.IsValid() {
			snapshot = latest
		}
	} else {
		snapshot.ActorId, err = GetId()
		if err != nil {
			return snapshot, err
		}
	}

	// check for all events since the latest snapshot
	if actorId != uuid.Nil {
		eventId := snapshot.LastEventId
		events, err = repository.Storage.FetchEventsSince(
			actorId,
			eventId,
		)
		if err != nil {
			return snapshot, err
		}
	}

	eventCount := len(events)
	es := make([]Event, eventCount)
	for i, record := range events {
		es[i] = record.Data.(Event)
	}
	actor := snapshot.Data.(T)
	// apply events to snapshot
	next := repository.Apply(es, actor)

	// update snapshot record
	if eventCount > 0 {
		snapshot.EventsApplied += uint64(eventCount)
		last := events[len(events)-1]
		snapshot.LastEventOn = last.CreatedOn
		snapshot.LastEventId = last.Id
		snapshot.Version++
		snapshot.Data = next
	}

	if snapshot.ActorId == uuid.Nil {

	}

	return snapshot, nil
}

func (repository Repository[T]) Fetch(ids Identifiers) (T, error) {
	snapshot, err := repository.fetch(ids)
	if err != nil {
		return repository.getEmpty(), err
	}
	return snapshot.Data.(T), nil
}

func (repository Repository[T]) Handle(command Command) Results[T] {
	identifiers := command.GetIdentifiers()
	baseline, err := repository.fetch(identifiers)
	if err != nil {
		return Results[T]{
			Errors: []error{err},
		}
	}

	cmdRecord, err := NewCommandRecord(command)
	if err != nil {
		return Results[T]{
			Original: baseline.Data.(T),
			Errors:   []error{err},
		}
	}
	cmdRecord.HandledBy = baseline.ActorId
	cmdRecord.HandledOn = time.Now()

	actor := baseline.Data.(T)
	events, errors := command.Handle(actor)
	next := repository.Apply(events, actor)
	eventRecords := make([]EventRecord, len(events))
	for i, event := range events {
		record, err := NewEventRecord(event)
		if err != nil {
			return Results[T]{
				Original: baseline.Data.(T),
				Errors:   []error{err},
			}
		}

		record.ActorId = baseline.ActorId
		record.ActorType = repository.ActorName
		record.CreatedBy = repository.ActorName
		record.CreatedById = baseline.ActorId
		record.CreatedByVersion = baseline.Version
		record.CreatedOn = time.Now()
		record.Id, _ = GetId()
		record.InitiatedBy = cmdRecord.Type
		record.InitiatedById = cmdRecord.Id
		eventRecords[i] = record
	}
	lastEventRecord := eventRecords[len(eventRecords)-1]

	snapshot, err := NewSnapshot(next)
	if err != nil {
		return Results[T]{
			Original: next,
			Errors:   []error{err},
		}
	}
	snapshot.ActorId = baseline.ActorId
	snapshot.EventsApplied += uint64(len(events))
	snapshot.LastCommandId = cmdRecord.Id
	snapshot.LastCommandOn = cmdRecord.HandledOn
	snapshot.LastEventId = lastEventRecord.Id
	snapshot.LastEventOn = lastEventRecord.CreatedOn
	snapshot.EventsApplied += uint64(len(eventRecords))
	snapshot.Version++

	// store id map
	err = repository.Storage.Maps.Add(identifiers, snapshot.ActorId)
	if err != nil {
		return Results[T]{
			Original: actor,
			Modified: next,
			Events:   events,
			Errors:   []error{err},
		}
	}

	// store events
	err = repository.Storage.Events.Add(eventRecords)
	if err != nil {
		return Results[T]{
			Original: actor,
			Modified: next,
			Events:   events,
			Errors:   []error{err},
		}
	}

	// store snapshot?
	err = repository.Storage.AddSnapshot(snapshot)
	if err != nil {
		return Results[T]{
			Original: actor,
			Modified: next,
			Events:   events,
			Errors:   []error{err},
		}
	}

	return Results[T]{
		Original: actor,
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
