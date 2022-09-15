package spry

import (
	"reflect"

	"github.com/google/uuid"
)

type Results[T Actor[T]] struct {
	Original T
	Modified T
	Events   []any
	Errors   []error
}

type EventStore interface {
	Add(event []Event) error
	FetchSince(actorId Identifiers, eventUUID uuid.UUID) ([]Event, error)
}

type SnapshotStore interface {
	Add(snapshot Snapshot) error
	Fetch(Identifiers) Snapshot
}

type CommandStore interface {
	Add(command Command) error
}

type Storage struct {
	Commands  CommandStore
	Events    EventStore
	Snapshots SnapshotStore
}

type Repository[T Actor[T]] struct {
	ActorType reflect.Type
	ActorName string
}

func (repository Repository[T]) Apply(events []any, actor T) (T, error) {
	return repository.getEmpty(), nil
}

func (repository Repository[T]) Fetch() (T, error) {
	return repository.getEmpty(), nil
}

func (repository Repository[T]) getEmpty() T {
	return *new(T)
}

func (repository Repository[T]) Handle(command any) Results[T] {
	// needs to eventually load instance from storage
	empty := repository.getEmpty()
	events, err := empty.Handle(command)
	next := empty.Apply(events)
	return Results[T]{
		Original: empty,
		Modified: next,
		Events:   events,
		Errors:   []error{err},
	}
}

func GetRepositoryFor[T Actor[T]]() Repository[T] {
	actorType := reflect.TypeOf(*new(T))
	actorName := actorType.Name()
	return Repository[T]{
		ActorType: actorType,
		ActorName: actorName,
	}
}
