package storage

import (
	"context"
	"errors"
	"reflect"

	"github.com/arobson/spry"
)

type AggregateRepository[T spry.Aggregate[T]] struct {
	Repository[T]
}

func (repository AggregateRepository[T]) handleAggregateCommand(ctx context.Context, command spry.Command) spry.Results[T] {
	idSet := command.(spry.Aggregate[T]).GetIdentifierSet()
	aggregateId := idSet.GetIdsFor(repository.ActorName)[0]
	baseline, err := repository.fetch(ctx, aggregateId)
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

func (repository AggregateRepository[T]) Handle(command spry.Command) spry.Results[T] {
	ctx, err := repository.Storage.GetContext(context.Background())
	if err != nil {
		return spry.Results[T]{Errors: []error{err}}
	}
	if _, ok := command.(spry.Aggregate[T]); ok {
		return repository.handleAggregateCommand(ctx, command)
	}
	return spry.Results[T]{
		Errors: []error{errors.New("command must implement GetIdentifierSet")},
	}
}

func GetAggregateRepositoryFor[T spry.Aggregate[T]](storage Storage) AggregateRepository[T] {
	actorType := reflect.TypeOf(*new(T))
	actorName := actorType.Name()
	return AggregateRepository[T]{
		Repository: Repository[T]{
			ActorType: actorType,
			ActorName: actorName,
			Storage:   storage,
		},
	}
}
