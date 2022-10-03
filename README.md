# spry

## `main` build [![CircleCI](https://dl.circleci.com/status-badge/img/gh/arobson/spry/tree/main.svg?style=svg)](https://dl.circleci.com/status-badge/redirect/gh/arobson/spry/tree/main)

An event sourcing library in Go.

> Initial postgres backend and in-memory storage implementation are functional.

## Use

```golang
package main

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
	Name  string
	Score uint32
}

// All Actors must provide a method that identifies an actor uniquely based on
// its state.
// Identifiers are a set of key/value pairs based on model state that uniquely 
// identifies the model.
// A field like Name is a decent choice (assuming you can enforce uniqueness).
// Score would be a mistake to include!

func (p Player) GetIdentifiers() spry.Identifiers {
	return spry.Identifiers{"Name": p.Name}
}

// Actors may optionally implement a method that returns configuration details 
// to the storage engine in spry controlling how often and when spry should 
// take a snapshot.

func (p Player) GetConfiguration() spry.ActorMeta {
	// defaults are shown below
	return spry.ActorMeta{
		snapshotFrequency: 			10,     // produces snapshots often
		snapshotDuringRead: 		false,  // prevents snapshotting during fetch (read)
		snapshotDuringWrite: 		true,   // take snapshots during command handling (write)
		snapshotDuringPartition: 	true,   // if supported by the storage adpater, 
										    // snapshot even if a partition is detected
	}
}

// Commands are VerbNoun named structures that target a specific Actor using 
// the same function signature that Actors have.
// The command must carry enough information to correctly target a specific 
// Actor instance.

type CreatePlayer struct {
	Name string
}

func (c CreatePlayer) GetIdentifiers() spry.Identifiers {
	return spry.Identifiers{"Name": c.Name}
}

// Commands must also provide a Handle method for accepting an Actor instance
// and determining how to handle that type of Actor. The Handle method must 
// never mutate the Actor and should return either events or errors that 
// describe why the command was invalid for a given actor state. Commands 
// passed an actor type that isn't relevant should no-op and return empty 
// arrays.

func (command CreatePlayer) Handle(actor any) ([]spry.Event, []error) {
	var events []spry.Event
	switch actor.(type) {
	case Player:
		events = append(events, PlayerCreated(command))
	}
	return events, []error{}
}

// Events must implement an Apply method, similar to Command's Handle. 
// Events are expected to handle any Actor type passed to them. An Event's 
// job is to mutate a copy of the Actor's state.

type PlayerCreated struct {
	Name string
}

// This shows that the Apply function can dispatch the application to a method 
// on the Event or the Actor the point is that you have flexibility in how you
// model your application as long as each "primitive" implements the required 
// methods.

func (e PlayerCreated) Apply(actor any) any {
	switch a := actor.(type) {
	case *Player:
		e.applyToPlayer(a)
	}
	return actor
}

func (e PlayerCreated) applyToPlayer(player *Player) {
	player.Score = 0
	player.Name = e.Name
}

// spry's API is exposed via a Repository instance.
// Repositories are strongly-typed to a single Actor type. Repository instances
// are created by a Storage implementation. Two implementations currently 
// exist:
// - InMemory (for simple testing)
// - Postgres

// creating a Repository from an InMemoryStore:

func FromMemory[T spry.Actor[T]]() spry.Repository[T] {
	memoryStore := memory.InMemoryStorage()
	return storage.GetRepositoryFor[T](memoryStore)
}

// creating a Repository from a PostgresStore:

func FromPostgres[T spry.Actor[T]](connectionURI string) spry.Repository[T] {
	postgresStore := postgres.CreatePostgresStorage(connectionURI)
	// any disk backed stores will require you to register event types
	// with the storage mechanism so it will know how to correctly deserialize
	// those events on fetch
	postgresStore.RegisterPrimitives({
		PlayerCreated{}
	})
	return storage.GetRepositoryFor[T](postgresStore)
}

// Let's see how all this comes together when using the Repository API:

func main() {
	// let's get our repository:

	// no errors are returned - storage instances will intentionally panic
	// and crash the process when a storage medium cannot be contacted
	players := FromPostgres[Player]("postgres://user:password@localhost:5432/dbname")

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
	myPlayer, err := players.Fetch(spry.Identifiers{"Name": "A Super Unique Name"})
	if err != nil {
		fmt.Println("failed to read player", err)
		os.Exit(1)
	}
	fmt.Printf("loaded player %s successfully", myPlayer.Name)
}
```

## Why?

An initial glance may make this seem like more effort than a more common CRUD/ORM based approach where
models are persisted and read directly to / from schema and generated (or handwritten) queries. While
this approach is more familiar, it has a number of shortcomings that we tend not to consider until
we're faced with the limitations.

  1. All data requires a custom-purpose schema
  1. All persistence and access require tailored queries defined as SQL or ORM API
  1. All writes destroy data from the system - you can't read old information
  1. Errors in application code or queries result in permanent data loss
  1. Performance tuning is always specific to targeted schema or queries
  1. Performance behavior is unpredictable
  1. Integrations take longer and have to be adapted for all schema/model changes

In contrast, an event-sourced approach like this one has the following advantages:

  1. No schema to write or manage: schema is the same for each Actor type
  1. No queries to write or adapt: storage adapters can pre-implement all queries ahead of time
  1. No data loss during writes: events provide a complete log of all historical state
  1. Fix and replay: changes in event application can correct defects in how state was determined
  1. Improvements are global: improvements to a storage adapter impact _all_ Actor's data access
  1. Performance behavior is more predictable with uniform access patterns
  1. Simple integrations: build integrations around streams of events from your system

## Concepts

### Availability and Partition Tolerance

spry provides some configuration behaviors for Actors that allow the application to tune behaviors 
like snapshotting frequency to adapt to specific data access patterns without forcing application 
authors to delve into implementation details. That said, its focus is on providing an approach to 
event sourcing that is very similar to CRDTs in that storage adapters _can_ detect divergent 
replicas (incompatible snapshots) and repair them automatically when possible. This means that 
partitions in your system or storage layer don't have to result in offline modes or degraded 
availability.

### Emphasis on Simplicity

While spry certainly has a few complex routines that are part of its 

### Actors

An actor's role is to provide data and identifiers that uniquely set it apart from every other 
Actor of the same type. They can include methods that handle a Command or apply an event but this 
is a stylistic choice for the application authors to make.

### Commands

A command is how we define change to an Actor's state. Instead of mutating the Actor state 
directly, the command should result in one or more events **or** one or more errors. Events 
define the changes to take place when applied to the model while errors should explain to the 
application why the Actor's logic refuses to handle the command.

### Events

Events are essentially a mutator attached to data. With ordering guarantees, we can always load 
events, apply them in order to a baseline (or new instance) and get the same result. This quality 
is sometimes referred to as _commutation_ because different processes can derive the same state 
independent of one another given access to the event log.

### Ordering Guarantees

Spry makes use of [RFC 4122 v6][1] which provides coordination-free, k-ordered, UUIDs. These ids 
are used as ids across all record types. This should ensure that events and snapshots can always 
sort in the order they were created. 

### Snapshots

Snapshots are a point-in-time capture of an Actor's state. Creating these at regular, configurable 
intervals prevents spry from having to read _every_ event that has occurred for a particular actor 
over its entire history.

## Storage

### Philosophy

When writing a storage adapter for spry, it's important to remember a few guiding principles:

 * Every Actor should receive its own set of tables/buckets/storage
 * All write operations should append only (`INSERT` never `UPDATE`)
 * The UUIDs produced by Spry should never be exposed to the application
 * Spry expects that the Identifiers map will map to a consistent UUID

### CommandStore

### EventStore

### MapStore

### SnapshotStore


[1]: https://datatracker.ietf.org/doc/html/draft-peabody-dispatch-new-uuid-format-03#section-5.1