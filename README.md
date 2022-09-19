# spry

An event sourcing library in Go.

> A simplistic postgres storage and in-memory storage implementation are working.

## Use

```golang
package demo

import (
	"fmt"
	"os"

	"github.com/arobson/spry"
	"github.com/arobson/spry/memory"
	"github.com/arobson/spry/postgres"
	"github.com/arobson/spry/storage"
)

// Actors are how you model state in spry.

type Player struct {
	Name String
	Score uint32
}

// All Actors must provide a method that identifies
// an actor uniquely based on it's state.
// Identifiers are a set of key/value pairs based on
// model state that uniquely identifiers the model.
// A field like Name is a decent choice (assuming you enforce uniqueness).
// Score would be a mistake to include!

func (p Player) GetIdentifiers() spry.Identifiers {
	return Identifiers{"Name": p.Name}
}

// Commands are Verb-Noun named structures that target a specific
// Actor using the same function signature that Actors have.
// The command must carry enough information to correctly
// target a specific Actor instance.

type CreatePlayer struct {
	Name string
}

func (c CreatePlayer) GetIdentifiers() spry.Identifiers {
	return Identifiers{"Name": c.Name}
}

// Commands must also provide a Handle method for accepting
// an Actor instance and determining how to handle that type
// of Actor. The Handle method must never mutate the Actor
// and should return either events or errors that describe
// why the command was invalid for a given actor state.
// Commands passed an actor type that isn't relevant should
// no-op and return empty arrays.

func (command CreatePlayer) Handle(actor any) ([]spry.Event, []error) {
	var events []spry.Event
	switch actor.(type) {
	case Player:
		events = append(events, PlayerCreated(command))
	}
	return events, []error{}
}

// Events must implement an Apply method, that similar to
// the Command's Handle, are expected to handle any Actor
// type passed to them. An Event's job is to mutate a copy
// of the Actor's state.

type PlayerCreated struct {
	Name string
}

// This shows that the Apply function can dispatch
// the application to a method on the Event or the Actor
// the point is that you have flexibility in how you
// model your application as long as each "primitive"
// implements the required methods.
func (e PlayerCreated) Apply(actor any) any {
	switch a := actor.(type) {
	case *Player:
		e.applyToPlayer(a)
	}
	return actor
}

func (e PlayerCreated) applyToPlayer(player *Player) {
	player.HitPoints = 100
	player.Name = e.Name
}

// spry's API is exposed via a Repository instance.
// Repository's are strongly-typed to a single Actor type.
// Repository instances are created by a Storage implementation.
// Two implementations exist presently - InMemory (for simple testing)
// and Postgres.

// creating a Repository from an InMemoryStore
func FromMemory[T Actor[T]]() Repository[T] {
	memoryStore := memory.InMemoryStorage()
	return storage.GetRepositoryFor[Player](store)
}

// creating a Repository from a PostgresStore
func FromPostgres[T Actor[T]](connectionURI) Repository[T] {
	postgresStore := postgres.CreatePostgresStorage(connectionURI string)
	return storage.GetRepositoryFor[Player](store)
}

// Let's see how all this comes together when using the Repository API
func main() {
	// let's get our repository:

	// no errors are returned - storage instances will intentionally panic
	// and crash the process when a storage medium cannot be contacted
	players := FromPostgres("postgres://user:password@localhost:5432/dbname")

	// let's create Player
	results := players.Handle(CreatePlayer{Name: "A Super Unique Name"})

	// the results struct has 4 properties:
	// * Original - the initial state of the Actor instance
	// * Modified - the new state of the Actor (if any) that occurs after the 
	//				Events produced by the Command (if any) are applied to the
	//				Original state of the Actor.
	// * Events	- the events that were produced by handling the
	//			  Command on the Actor given it's original state.
	// * Errors - the errors produced by by attempting to handle
	//			  the Command for the Actor in its Original state.
	//			  Errors are how your model should indicate that
	//			  a Command is not valid in the Actor's given state.
	fmt.Printf("Player created: %s\n", results.Modified.Name)

	// To get the present state of your Actor, you can call Fetch:
	myPlayer, err := players.Fetch(Identifiers{"Name": "A Super Unique Name"})
	if err != nil {
		fmt.Println("failed to read player", err)
		os.Exit(1)
	}
	fmt.Printf("loaded player %s successfully", myPlayer.Name)
}
```

## Why?

An initial glance may make this seem like more effort than a more common CRUD/ORM based approach where
models are read and persisted directly to / from schema and generated (or hand-written) queries. While
this approach is more familiar, it has a number of short-comings that we tend not to consider until
we're faced with the limitations.

  1. All data requires a custom-purpose schema
  1. All persistence and access require tailored queries defined as SQL or ORM API
  1. All writes destory data from the system - you can't read old information
  1. Errors in application code or queries result in permanant data loss
  1. Performance tuning is always specific to targeted schema or queries
  1. Performance behavior is unpredictable
  1. Integrations take longer and have to be adpated for all schema/model changes

In contrast, an event sourced approach like this one has has the following advantages:

  1. No schema to write or manage: schema is the same for each Actor type
  1. No queries to write or adapt: storage adapters can pre-implement all queries ahead of time
  1. No data loss during writes: events provide a complete log of all historical state
  1. Fix and replay: changes in event application can correct defects in how state was determined
  1. Improvements are global: improvements to a storage adapter impact _all_ Actor's data access
  1. Performance behavior is more predictable with uniform access patterns
  1. Simple integrations: build integrations around streams of events from your system

## Concepts

## Storage