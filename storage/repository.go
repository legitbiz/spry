package storage

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/arobson/spry"
	"github.com/gofrs/uuid"
)

type Repository[T spry.Actor[T]] struct {
	ActorType reflect.Type
	ActorName string
	Storage   Storage
	Mapping   TypeMap
}

func getEmpty[T spry.Actor[T]]() T {
	return *new(T)
}

// A side-effect free way of applying events to an actor instance
func (repository Repository[T]) Apply(events []spry.Event, actor T) T {
	var modified T = actor
	for _, event := range events {
		event.Apply(&modified)
	}
	return modified
}

func (repository Repository[T]) fetch(ctx context.Context, ids spry.Identifiers) (Snapshot, error) {
	// create an empty actor instance and empty snapshot
	empty := getEmpty[T]()
	snapshot, err := NewSnapshot(empty)
	events := []EventRecord{}
	if err != nil {
		return snapshot, err
	}

	// fetch the actor id from the identifier
	actorId, err := repository.Storage.FetchId(ctx, repository.ActorName, ids)
	if err != nil {
		return snapshot, err
	}

	// check for the latest snapshot available
	if actorId != uuid.Nil {
		latest, err := repository.Storage.FetchLatestSnapshot(ctx, repository.ActorName, actorId)
		if err != nil {
			return snapshot, err
		}
		if latest.IsValid() {
			snapshot = latest
		} else {
			snapshot.ActorId = actorId
		}
	} else {
		snapshot.ActorId, _ = GetId()
		if err != nil {
			return snapshot, err
		}
	}

	// check for all events since the latest snapshot
	if actorId != uuid.Nil {
		eventId := snapshot.LastEventId
		events, err = repository.Storage.FetchEventsSince(
			ctx,
			repository.ActorName,
			actorId,
			eventId,
		)
		if err != nil {
			return snapshot, err
		}
	}

	// extract events
	eventCount := len(events)
	es := make([]spry.Event, eventCount)
	for i, record := range events {
		es[i] = record.Data.(spry.Event)
	}

	// apply events to actor instance
	actor := snapshot.Data.(T)
	next := repository.Apply(es, actor)

	if eventCount > 0 {
		snapshot.EventsApplied += uint64(eventCount)
		last := events[len(events)-1]
		snapshot.LastEventOn = last.CreatedOn
		snapshot.LastEventId = last.Id
		snapshot.Version++
		snapshot.Data = next
	}

	if snapshot.ActorId == uuid.Nil {
		snapshot.ActorId, _ = GetId()
	}

	config := spry.GetActorMeta[T]()
	// do we allow snapshotting during read?
	// if so, have we passed the event threshold?
	if config.SnapshotDuringRead &&
		eventCount > config.SnapshotFrequency {
		snapshot.EventSinceSnapshot = 0
		// ignore any error creating snapshots during read
		_ = repository.Storage.AddSnapshot(
			ctx,
			repository.ActorName,
			snapshot,
			config.SnapshotDuringPartition,
		)
	}

	return snapshot, nil
}

func (repository Repository[T]) Fetch(ids spry.Identifiers) (T, error) {
	ctx := context.Background()
	ctx, err := repository.Storage.GetContext(ctx)
	if err != nil {
		return getEmpty[T](), err
	}
	snapshot, err := repository.fetch(ctx, ids)
	if err != nil {
		return getEmpty[T](), err
	}

	return snapshot.Data.(T), nil
}

func (repository Repository[T]) handleActorCommand(ctx context.Context, command spry.Command) spry.Results[T] {
	identifiers := command.(spry.Actor[T]).GetIdentifiers()
	baseline, err := repository.fetch(ctx, identifiers)
	if err != nil {
		return spry.Results[T]{
			Errors: []error{err},
		}
	}

	cmdRecord, s, done := repository.createCommandRecord(command, baseline)
	if done {
		return s
	}

	actor := baseline.Data.(T)
	events, errors := command.Handle(actor)
	next := repository.Apply(events, actor)
	eventRecords, s, done := repository.createEventRecords(events, baseline, cmdRecord)
	if done {
		return s
	}

	snapshot, s, done := repository.createSnapshot(next, baseline, cmdRecord, eventRecords)
	if done {
		return s
	}

	// store id map
	err = repository.Storage.AddMap(ctx, repository.ActorName, identifiers, snapshot.ActorId)
	if err != nil {
		return spry.Results[T]{
			Original: actor,
			Modified: next,
			Events:   events,
			Errors:   []error{err},
		}
	}

	// store events
	err = repository.Storage.AddEvents(ctx, eventRecords)
	if err != nil {
		_ = repository.Storage.Rollback(ctx)
		return spry.Results[T]{
			Original: actor,
			Modified: next,
			Events:   events,
			Errors:   []error{err},
		}
	}

	config := spry.GetActorMeta[T]()
	// do we allow snapshotting during read?
	// if so, have we passed the event threshold?
	if config.SnapshotDuringWrite &&
		snapshot.EventSinceSnapshot >= config.SnapshotFrequency {
		snapshot.EventSinceSnapshot = 0
		err = repository.Storage.AddSnapshot(
			ctx,
			repository.ActorName,
			snapshot,
			config.SnapshotDuringPartition,
		)
		if err != nil {
			return spry.Results[T]{
				Original: actor,
				Modified: next,
				Events:   events,
				Errors:   []error{err},
			}
		}
	}

	err = repository.Storage.Commit(ctx)
	if err != nil {
		_ = repository.Storage.Rollback(ctx)
		return spry.Results[T]{
			Original: actor,
			Modified: next,
			Events:   events,
			Errors:   []error{err},
		}
	}

	return spry.Results[T]{
		Original: actor,
		Modified: next,
		Events:   events,
		Errors:   errors,
	}
}

func (repository Repository[T]) createEventRecords(events []spry.Event, baseline Snapshot, cmdRecord CommandRecord) ([]EventRecord, spry.Results[T], bool) {
	eventRecords := make([]EventRecord, len(events))
	for i, event := range events {
		record, err := NewEventRecord(event)
		if err != nil {
			return nil, spry.Results[T]{
				Original: baseline.Data.(T),
				Errors:   []error{err},
			}, true
		}

		record.ActorId = baseline.ActorId
		if record.ActorType == "" {
			record.ActorType = repository.ActorName
		}
		if record.CreatedBy == "" {
			record.CreatedBy = repository.ActorName
		}
		record.CreatedById = baseline.ActorId
		record.CreatedByVersion = baseline.Version
		record.CreatedOn = time.Now()
		record.Id, _ = GetId()
		record.InitiatedBy = cmdRecord.Type
		record.InitiatedById = cmdRecord.Id
		eventRecords[i] = record
	}
	return eventRecords, spry.Results[T]{}, false
}

func (repository Repository[T]) createSnapshot(next T, baseline Snapshot, cmdRecord CommandRecord, events []EventRecord) (Snapshot, spry.Results[T], bool) {
	lastEventRecord := events[len(events)-1]
	snapshot, err := NewSnapshot(next)
	if err != nil {
		return Snapshot{}, spry.Results[T]{
			Original: next,
			Errors:   []error{err},
		}, true
	}
	snapshot.ActorId = baseline.ActorId
	snapshot.LastCommandId = cmdRecord.Id
	snapshot.LastCommandOn = cmdRecord.HandledOn
	snapshot.LastEventId = lastEventRecord.Id
	snapshot.LastEventOn = lastEventRecord.CreatedOn
	snapshot.EventsApplied += uint64(len(events))
	snapshot.EventSinceSnapshot += len(events)
	snapshot.Version++
	return snapshot, spry.Results[T]{}, false
}

func (repository Repository[T]) createCommandRecord(command spry.Command, baseline Snapshot) (CommandRecord, spry.Results[T], bool) {
	cmdRecord, err := NewCommandRecord(command)
	if err != nil {
		return CommandRecord{}, spry.Results[T]{
			Original: baseline.Data.(T),
			Errors:   []error{err},
		}, true
	}
	cmdRecord.HandledBy = baseline.ActorId
	cmdRecord.HandledVersion = baseline.Version
	cmdRecord.HandledOn = time.Now()
	return cmdRecord, spry.Results[T]{}, false
}

func (repository Repository[T]) handleAggregateCommand(ctx context.Context, command spry.Command) spry.Results[T] {
	return spry.Results[T]{}
}

func (repository Repository[T]) Handle(command spry.Command) spry.Results[T] {
	ctx, err := repository.Storage.GetContext(context.Background())
	if err != nil {
		return spry.Results[T]{Errors: []error{err}}
	}
	if _, ok := command.(spry.Aggregate[T]); ok {
		return repository.handleAggregateCommand(ctx, command)
	}
	if _, ok := command.(spry.Actor[T]); ok {
		return repository.handleActorCommand(ctx, command)
	}
	return spry.Results[T]{
		Errors: []error{errors.New("command must implement GetIdentifiers or GetIdentifierSet")},
	}
}

func GetRepositoryFor[T spry.Actor[T]](storage Storage) Repository[T] {
	actorType := reflect.TypeOf(*new(T))
	actorName := actorType.Name()
	return Repository[T]{
		ActorType: actorType,
		ActorName: actorName,
		Storage:   storage,
	}
}
