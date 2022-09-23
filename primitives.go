package spry

import (
	"encoding/json"
)

type Identifiers = map[string]any

type IdentifierSet = map[string][]Identifier

type Actor[T any] interface {
	GetIdentifiers() Identifiers
}

type Aggregate[T any] interface {
	GetIdentifiers() Identifiers
	GetIdentifierSet() IdSet
}

type IdSet struct {
	ids IdentifierSet
}

func (set *IdSet) AddIdsFor(actorType string, ids Identifiers) {
	list, ok := set.ids[actorType]
	if !ok {
		list = []Identifiers{ids}
	} else {
		list = append(list, ids)
	}
	set.ids[actorType] = list
}

func (set *IdSet) GetIdsFor(actorType string) []Identifiers {
	list, ok := set.ids[actorType]
	if ok {
		return list
	}
	return []Identifiers{}
}

func (set *IdSet) RemoveIdsFrom(actorType string) bool {
	list, ok := set.ids[actorType]
	if ok {
		list.		
		return false
	}
	return false
}

func (set *IdSet) ToIdentifierSet() IdentifierSet {
	return set.ids
}

func CreateIdSet() IdSet {
	return IdSet{
		ids: IdentifierSet{},
	}
}

type ActorMeta struct {
	// how many events should occur before the next snapshot
	SnapshotFrequency int
	// controls whether snapshots occur during fetch (read)
	SnapshotDuringRead bool
	// controls whether snapshots occur during handle (write)
	SnapshotDuringWrite bool
	// controls whether snapshots can occur during partitions
	// requires a storage adapter for a database that can
	// detect this
	SnapshotDuringPartition bool
}

type HasMeta interface {
	GetActorMeta() ActorMeta
}

var default_meta = ActorMeta{
	SnapshotFrequency:       20,
	SnapshotDuringRead:      false,
	SnapshotDuringWrite:     true,
	SnapshotDuringPartition: true,
}

func getEmpty[T Actor[T]]() T {
	return *new(T)
}

func GetActorMeta[T Actor[T]]() ActorMeta {
	var empty any = getEmpty[T]()
	hasMeta, ok := empty.(HasMeta)
	if ok {
		return hasMeta.GetActorMeta()
	}
	return default_meta
}

type Command interface {
	GetIdentifiers() Identifiers
	Handle(any) ([]Event, []error)
}

type Event interface {
	Apply(any) any
}

type Repository[T Actor[T]] interface {
	Apply(events []Event, actor T) T
	Fetch(ids Identifiers) (T, error)
	Handle(command Command) Results[T]
}

func IdMapToString(ids Identifiers) (string, error) {
	bytes, err := ToJson(ids)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

type Results[T Actor[T]] struct {
	Original T
	Modified T
	Events   []Event
	Errors   []error
}

func ToJson[T any](obj T) ([]byte, error) {
	return json.Marshal(obj)
}

func FromJson[T any](bytes []byte) (T, error) {
	obj := *new(T)
	err := json.Unmarshal(bytes, &obj)
	return obj, err
}
