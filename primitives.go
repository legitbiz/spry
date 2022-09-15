package spry

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Identifiers = map[string]any

type Actor[T any] interface {
	GetIdentifiers() Identifiers
}

type Command interface {
	Handle(any) ([]Event, []error)
}

type Event interface {
	Apply(any) any
}

func ToJson[T any](obj T) ([]byte, error) {
	return json.Marshal(obj)
}

func FromJson[T any](bytes []byte) (T, error) {
	obj := *new(T)
	err := json.Unmarshal(bytes, obj)
	return obj, err
}

type Snapshot struct {
	// a generated uuid (system id) for the actor instance
	Id uuid.UUID `json:"id"`
	// the type name of the actor
	Type string `json:"type"`
	// a serialized causal tracker
	Vector string `json:"vector"`
	// a numeric version of the model
	Version uint64 `json:"version"`
	// the causal tracker of the preceding snapshot
	Ancestor string `json:"ancestor"`
	// the number of events applied to reach the present state
	EventsApplied uint64 `json:"eventsApplied"`
	// the UUID of the last event played against the instance
	LastEventId uuid.UUID `json:"lastEventId"`
	// the UUID of the last command handled
	LastCommandId uuid.UUID `json:"lastCommandId"`
	// the wall clock at the time of the last command
	LastCommandOn time.Time `json:"lastCommandOn"`
	// the wall clock at the time of the last event
	LastEventOn time.Time `json:"lastEventOn"`
	// the contents of the snapshot
	Data any
}

type EventRecord struct {
	// a generated uuid for this event
	Id uuid.UUID `json:"id"`
	// the type name of the event
	Type string `json:"type"`
	// inferred from the actor emitting the event
	ActorNamespace string `json:"namespace"`
	// the type of the model the event was generated for
	ActorType string `json:"actor"`
	// this is the addressable identity of the owning model
	SnapshotId uuid.UUID `json:"snapshotId"`
	// UTC ISO date time string when event was created
	CreatedOn time.Time `json:"createdOn"`
	// the type of the actor instantiating the event
	CreatedBy string `json:"createdBy"`
	// the id of the snapshot instantiating the event
	CreatedById uuid.UUID `json:"createdById"`
	// the vector of the snapshot instantiating the event
	CreatedByVector string `json:"createdByVector"`
	// the version of the snapshot instantiating the event
	CreatedByVersion uint64 `json:"createdByVersion"`
	// the command type/topic that triggered the event
	InitiatedBy string `json:"initiatedBy"`
	// the id of the message that triggered the event
	InitiatedById uuid.UUID `json:"initiatedById"`
	// the contents of the event
	Data any
}

type CommandRecord struct {
	// a generated uuid for this event
	Id uuid.UUID `json:"id"`
	// the type name of the command
	Type string `json:"type"`
	// namespace for the command
	Namespace string `json:"namespace"`
	// the time the command was handled
	ReceivedOn time.Time `json:"receivedOn"`
	// the time the command was handled
	HandledOn time.Time `json:"handledOn"`
	// the id of the recipient actor
	HandledBy uuid.UUID `json:"handledBy"`
	// the contents of the event
	Data any
}
