package spry

import (
	"encoding/json"
)

type Identifiers = map[string]any

type Actor[T any] interface {
	GetIdentifiers() Identifiers
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

type Results[T Actor[T]] struct {
	Original T
	Modified T
	Events   []Event
	Errors   []error
}

func IdMapToString(ids Identifiers) (string, error) {
	bytes, err := ToJson(ids)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func ToJson[T any](obj T) ([]byte, error) {
	return json.Marshal(obj)
}

func FromJson[T any](bytes []byte) (T, error) {
	obj := *new(T)
	err := json.Unmarshal(bytes, obj)
	return obj, err
}
