package storage

import (
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

func (repository Repository[T]) fetch(ids spry.Identifiers) (Snapshot, error) {
	// create an empty actor instance and empty snapshot
	empty := getEmpty[T]()
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
			actorId,
			eventId,
		)
		if err != nil {
			return snapshot, err
		}
	}

	eventCount := len(events)
	es := make([]spry.Event, eventCount)
	for i, record := range events {
		es[i] = record.Data.(spry.Event)
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
		snapshot.ActorId, _ = GetId()
	}

	return snapshot, nil
}

func (repository Repository[T]) Fetch(ids spry.Identifiers) (T, error) {
	snapshot, err := repository.fetch(ids)
	if err != nil {
		return getEmpty[T](), err
	}
	return snapshot.Data.(T), nil
}

func (repository Repository[T]) Handle(command spry.Command) spry.Results[T] {
	identifiers := command.GetIdentifiers()
	baseline, err := repository.fetch(identifiers)
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
	snapshot.EventsApplied += uint64(len(events))
	snapshot.LastCommandId = cmdRecord.Id
	snapshot.LastCommandOn = cmdRecord.HandledOn
	snapshot.LastEventId = lastEventRecord.Id
	snapshot.LastEventOn = lastEventRecord.CreatedOn
	snapshot.EventsApplied += uint64(len(eventRecords))
	snapshot.Version++

	// store id map
	err = repository.Storage.AddMap(identifiers, snapshot.ActorId)
	if err != nil {
		return spry.Results[T]{
			Original: actor,
			Modified: next,
			Events:   events,
			Errors:   []error{err},
		}
	}

	// store events
	err = repository.Storage.AddEvents(eventRecords)
	if err != nil {
		return spry.Results[T]{
			Original: actor,
			Modified: next,
			Events:   events,
			Errors:   []error{err},
		}
	}

	// store snapshot?
	err = repository.Storage.AddSnapshot(snapshot)
	if err != nil {
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

func GetRepositoryFor[T spry.Actor[T]](storage Storage) Repository[T] {
	actorType := reflect.TypeOf(*new(T))
	actorName := actorType.Name()
	return Repository[T]{
		ActorType: actorType,
		ActorName: actorName,
		Storage:   storage,
	}
}
