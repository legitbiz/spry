# spry

An event sourcing library in Go.

> This only works as an in-memory proof of concept and is missing some features.

## Use

```go
package yourApp

import (
    "github.com/arobson/spry"
)



// An actor is a simple struct with one
// required method - GetIdentifiers()
type Player struct {
    spry.Identifiers
    Name string
    HitPoints uint
}

// required by spry to identify the model as
// a unique instance
func (p Player) GetIdentifiers() Identifiers {
	return map[string]any{"name": p.Name}
}

// Event definitions consist of a struct and 
// method for applying the struct to each
// model type they suport.
type PlayerCreated struct {
    Name string
}

func (event PlayerCreated) applyToPlayer(player *Player) {
	player.HitPoints = 100
	player.Name = event.Name
}

// the public Apply call must diferentiate model
// based on its type but can dispatch to a specific
// private method for application.

// The Apply method must never mutate the incoming
// actor instance but should return the changed
// copy.
func (event PlayerCreated) Apply(actor any) any {
	switch a := actor.(type) {
	case *Player:
		event.applyToPlayer(a)
	}
	return actor
}

type PlayerDamaged struct {
	Damage int
}

func (event PlayerDamaged) Apply(actor any) any {
	switch a := actor.(type) {
	case *Player:
		a.HitPoints -= event.Damage
	}
	return actor
}

func (event PlayerHealed) Apply(actor any) any {
	switch a := actor.(type) {
	case Player:
		a.HitPoints -= event.Health
	}
	return actor
}

type PlayerDied struct {
	Message string
}

func (event PlayerDied) Apply(actor any) any {
	switch a := actor.(type) {
	case World:
		a.PlayerCount--
	}
	return actor
}

// Commands are like Events in that they
// require a method for dispatching the command
// to model diferentiated by type.
// Commands also require a GetIdentifiers method
// to specify the model instance that they target.
// Finally, it is expected that a Command Handle call
// return either a list of events OR a list of errors
// but never both.

type CreatePlayer struct {
	Name string
}

func (command CreatePlayer) GetIdentifiers() Identifiers {
	return Identifiers{"name": command.Name}
}

func (command CreatePlayer) Handle(actor any) ([]Event, []error) {
	var events []Event
	switch actor.(type) {
	case Player:
		events = append(events, PlayerCreated(command))
	}
	return events, []error{}
}

type DamagePlayer struct {
	Name   string
	Damage int
}

func (command DamagePlayer) GetIdentifiers() Identifiers {
	return Identifiers{"name": command.Name}
}

func (command DamagePlayer) Handle(actor any) ([]Event, []error) {
	var events []Event
	switch a := actor.(type) {
	case Player:
		if a.HitPoints <= command.Damage {
			events = append(events, PlayerDied{Message: "you died"})
		}
		events = append(events, PlayerDamaged{Damage: command.Damage})
	}
	return events, []error{}
}

type HealPlayer struct {
	Name   string
	Health int
}

func (command HealPlayer) GetIdentifiers() Identifiers {
	return Identifiers{"name": command.Name}
}

func (command HealPlayer) Handle(actor any) ([]Event, []error) {
	var events []Event
	switch actor.(type) {
	case Player:
		events = append(events, PlayerHealed{Health: command.Health})
	}
	return events, []error{}
}

```