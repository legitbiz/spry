package storage

import (
	"context"
	"reflect"
	"time"

	"github.com/arobson/spry"
	"github.com/gofrs/uuid"
)

type Repository[T any] struct {
	ActorType reflect.Type
	ActorName string
	Storage   Storage
	Mapping   TypeMap
}

func getEmpty[T any]() T {
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

func (repository Repository[T]) fetchActor(ctx context.Context, ids spry.Identifiers) (Snapshot, error) {

	// get the latest snapshot or initialize and empty
	snapshot, actorId, err := repository.getLatestSnapshot(ctx, ids)
	if err != nil {
		return snapshot, err
	}

	// check for all events since the latest snapshot
	events, records, err := repository.getEventsSince(ctx, actorId, snapshot)
	if err != nil {
		return snapshot, err
	}

	// apply events to actor instance
	repository.updateActor(events, records, &snapshot)

	// write snapshot
	err = repository.writeSnapshot(ctx, len(events), snapshot)
	return snapshot, err
}

func (repository Repository[T]) fetchAggregate(ctx context.Context, ids spry.Identifiers) (Snapshot, error) {

	// get the latest snapshot or initialize and empty
	snapshot, actorId, err := repository.getLatestSnapshot(ctx, ids)
	if err != nil {
		return snapshot, err
	}

	// check for all events since the latest snapshot
	events, records, err := repository.getAggregatedEventsSince(ctx, actorId, snapshot)
	if err != nil {
		return snapshot, err
	}

	// apply events to actor instance
	repository.updateActor(events, records, &snapshot)

	// write snapshot
	err = repository.writeSnapshot(ctx, len(events), snapshot)
	return snapshot, err
}

func (repository Repository[T]) getLatestSnapshot(ctx context.Context, ids spry.Identifiers) (Snapshot, uuid.UUID, error) {
	// create an empty actor instance and empty snapshot
	empty := getEmpty[T]()
	snapshot, err := NewSnapshot(empty)
	if err != nil {
		return snapshot, uuid.Nil, err
	}

	// fetch the actor id from the identifier
	actorId, err := repository.Storage.FetchId(ctx, repository.ActorName, ids)
	if err != nil {
		return snapshot, actorId, err
	}

	// fetch the latest snapshot from storage or return empty
	if actorId != uuid.Nil {
		latest, err := repository.Storage.FetchLatestSnapshot(ctx, repository.ActorName, actorId)
		if err != nil {
			return snapshot, actorId, err
		}
		if latest.IsValid() {
			snapshot = latest
		} else {
			snapshot.ActorId = actorId
		}
	} else {
		snapshot.ActorId, err = GetId()
		if err != nil {
			return snapshot, actorId, err
		}
	}
	return snapshot, actorId, nil
}

func (repository Repository[T]) getEventsSince(ctx context.Context, actorId uuid.UUID, snapshot Snapshot) ([]spry.Event, []EventRecord, error) {
	records := []EventRecord{}
	var err error
	if actorId != uuid.Nil {
		records, err = repository.Storage.FetchEventsSince(
			ctx,
			repository.ActorName,
			actorId,
			snapshot.LastEventId,
		)
		if err != nil {
			return nil, nil, err
		}
	}

	eventCount := len(records)
	events := make([]spry.Event, eventCount)
	for i, record := range records {
		events[i] = record.Data.(spry.Event)
	}
	return events, records, nil
}

func (repository Repository[T]) getAggregatedEventsSince(ctx context.Context, aggregateId uuid.UUID, snapshot Snapshot) ([]spry.Event, []EventRecord, error) {
	var err error

	// get id map from map store
	if aggregateId != uuid.Nil {
		return nil, nil, err
	}

	idMap, err := repository.Storage.FetchIdMap(ctx, repository.ActorName, aggregateId)
	if err != nil {
		return nil, nil, err
	}

	records, err := repository.Storage.FetchAggregatedEventsSince(
		ctx,
		idMap,
		snapshot.LastEventId,
	)
	if err != nil {
		return nil, nil, err
	}

	eventCount := len(records)
	events := make([]spry.Event, eventCount)
	for i, record := range records {
		events[i] = record.Data.(spry.Event)
	}
	return events, records, nil
}

func (repository Repository[T]) updateActor(events []spry.Event, records []EventRecord, snapshot *Snapshot) {
	actor := snapshot.Data.(T)
	next := repository.Apply(events, actor)
	eventCount := len(events)
	if eventCount > 0 {
		snapshot.EventsApplied += uint64(eventCount)
		last := records[len(records)-1]
		snapshot.LastEventOn = last.CreatedOn
		snapshot.LastEventId = last.Id
		snapshot.Version++
		snapshot.Data = next
	}

	if snapshot.ActorId == uuid.Nil {
		snapshot.ActorId, _ = GetId()
	}
}

func (repository Repository[T]) writeSnapshot(ctx context.Context, eventCount int, snapshot Snapshot) error {
	config := spry.GetActorMeta[T]()
	// do we allow snapshotting during read?
	// if so, have we passed the event threshold?
	var err error = nil
	if config.SnapshotDuringRead &&
		eventCount > config.SnapshotFrequency {
		snapshot.EventSinceSnapshot = 0
		// ignore any error creating snapshots during read
		err = repository.Storage.AddSnapshot(
			ctx,
			repository.ActorName,
			snapshot,
			config.SnapshotDuringPartition,
		)
	}

	return err
}
