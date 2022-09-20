package storage

import (
	"context"
	"reflect"
	"time"

	"github.com/arobson/spry"
	"github.com/gofrs/uuid"
)

type Repository[T spry.Actor[T]] struct {
	ActorType reflect.Type
	ActorName string
	Storage   Storage
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

func (repository Repository[T]) Handle(command spry.Command) spry.Results[T] {
	ctx, err := repository.Storage.GetContext(context.Background())
	if err != nil {
		return spry.Results[T]{Errors: []error{err}}
	}
	identifiers := command.GetIdentifiers()
	baseline, err := repository.fetch(ctx, identifiers)
	if err != nil {
		return spry.Results[T]{
			Errors: []error{err},
		}
	}

	cmdRecord, err := NewCommandRecord(command)
	if err != nil {
		return spry.Results[T]{
			Original: baseline.Data.(T),
			Errors:   []error{err},
		}
	}
	cmdRecord.HandledBy = baseline.ActorId
	cmdRecord.HandledVersion = baseline.Version
	cmdRecord.HandledOn = time.Now()

	actor := baseline.Data.(T)
	events, errors := command.Handle(actor)
	next := repository.Apply(events, actor)
	eventRecords := make([]EventRecord, len(events))
	for i, event := range events {
		record, err := NewEventRecord(event)
		if err != nil {
			return spry.Results[T]{
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
		return spry.Results[T]{
			Original: next,
			Errors:   []error{err},
		}
	}
	snapshot.ActorId = baseline.ActorId
	snapshot.LastCommandId = cmdRecord.Id
	snapshot.LastCommandOn = cmdRecord.HandledOn
	snapshot.LastEventId = lastEventRecord.Id
	snapshot.LastEventOn = lastEventRecord.CreatedOn
	snapshot.EventsApplied += uint64(len(events))
	snapshot.EventSinceSnapshot += len(events)
	snapshot.Version++

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
	err = repository.Storage.AddEvents(ctx, repository.ActorName, eventRecords)
	if err != nil {
		return spry.Results[T]{
			Original: actor,
			Modified: next,
			Events:   events,
			Errors:   []error{err},
		}
	}

	// store snapshot?

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

	return spry.Results[T]{
		Original: actor,
		Modified: next,
		Events:   events,
		Errors:   errors,
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
