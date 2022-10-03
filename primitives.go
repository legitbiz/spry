package spry

import (
	"encoding/json"
)

type Identifiers = map[string]any

type IdentifierSet = map[string][]Identifiers

type Actor[T any] interface {
	GetIdentifiers() Identifiers
}

type Aggregate[T any] interface {
	GetIdentifierSet() IdentifierSet
}

type IdSet struct {
	ids IdentifierSet
}

func (set *IdSet) AddIdsFor(actorName string, ids ...Identifiers) {
	list, ok := set.ids[actorName]
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
	set.ids[actorName] = list
}

func (set *IdSet) GetIdsFor(actorName string) []Identifiers {
	if list, ok := set.ids[actorName]; ok {
		return list
	}
	return []Identifiers{}
}

func (set *IdSet) RemoveIdsFrom(actorName string, ids ...Identifiers) bool {
	list, ok := set.ids[actorName]
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
		set.ids[actorName] = list
		return true
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

func IdSetFromIdentifierSet(ids IdentifierSet) IdSet {
	return IdSet{ids: ids}
}

func getEmpty[T any]() T {
	return *new(T)
}

type Command interface {
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
