package spry

import (
	"encoding/json"

	"github.com/gofrs/uuid"
)

type Identifiers = map[string]any

type IdentifierSet = map[string][]Identifiers

type AggregatedIds = map[string][]uuid.UUID

type AggregateIdMap struct {
	ActorName  string
	ActorId    uuid.UUID
	Aggregated AggregatedIds
}

func (idMap *AggregateIdMap) AddIdsFor(child string, id ...uuid.UUID) {
	ids := idMap.Aggregated
	if list, ok := ids[child]; ok {
		ids[child] = append(list, id...)
	} else {
		ids[child] = id
	}
}

type Actor[T any] interface {
	GetIdentifiers() Identifiers
}

type Aggregate[T any] interface {
	GetIdentifierSet() IdentifierSet
}

type IdSet struct {
	ids IdentifierSet
}

func (set *IdSet) AddIdsFor(actorType string, ids ...Identifiers) {
	list, ok := set.ids[actorType]
	if !ok {
		list = make([]Identifiers, len(ids))
	}
	for _, id := range ids {
		if !ok {
			list = []Identifiers{id}
		} else {
			list = append(list, id)
		}
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

func (set *IdSet) RemoveIdsFrom(actorType string, ids ...Identifiers) bool {
	list, ok := set.ids[actorType]
	lookup := map[string]int{}
	for i, id := range list {
		s, _ := IdentifiersToString(id)
		lookup[s] = i
	}
	if ok {
		index := make([]int, len(ids))
		for i, id := range ids {
			s, _ := IdentifiersToString(id)
			if idx, ok := lookup[s]; ok {
				index[i] = idx
			}
		}
		for i := len(index) - 1; i >= 0; i-- {
			idx := index[i]
			copy(list[idx:], list[idx+1:])
			list[len(list)-1] = nil
			list = list[:len(list)-1]
		}
		set.ids[actorType] = list
		return true
	}
	return false
}

func (set *IdSet) ToIdentifierSet() IdentifierSet {
	return set.ids
}

func CreateAggregateIdMap(actorName string, actorId uuid.UUID) AggregateIdMap {
	return AggregateIdMap{
		ActorName:  actorName,
		ActorId:    actorId,
		Aggregated: AggregatedIds{},
	}
}

func emptyAggregateIdMap() AggregateIdMap {
	return AggregateIdMap{}
}

func CreateIdSet() IdSet {
	return IdSet{
		ids: IdentifierSet{},
	}
}

func IdSetFromIdentifierSet(ids IdentifierSet) IdSet {
	return IdSet{ids: ids}
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

func getEmpty[T any]() T {
	return *new(T)
}

func GetActorMeta[T any]() ActorMeta {
	var empty any = getEmpty[T]()
	hasMeta, ok := empty.(HasMeta)
	if ok {
		return hasMeta.GetActorMeta()
	}
	return default_meta
}

type Command interface {
	Handle(any) ([]Event, []error)
}

type Event interface {
	Apply(any) any
}

type EventMetadata struct {
	CreatedBy  string
	CreatedFor string
}

func (e EventMetadata) GetEventMeta() EventMetadata {
	return e
}

type Namespaced interface {
	GetEventMeta() EventMetadata
}

type Repository[T Actor[T]] interface {
	Apply(events []Event, actor T) T
	Fetch(ids Identifiers) (T, error)
	Handle(command Command) Results[T]
}

func IdentifiersToString(ids Identifiers) (string, error) {
	bytes, err := ToJson(ids)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

type Results[T any] struct {
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
