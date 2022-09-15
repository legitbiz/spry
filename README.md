# spry

An event sourcing library in Go.

> Just some ideas at the moment, certainly nothing functional yet.

## Use

> No, really this won't work right now. Just an example.

```go
package yourApp

import (
    "github.com/arobson/spry"
)

// your actor
type Player struct {
    spry.Identifiers
    Name string
    HitPoints uint
}

// events

type PlayerCreated struct {
    Name string
}

type PlayerDamaged struct {
    Damage int
}

type PlayerHealed struct {
    Health int
}

type PlayerDied struct {
    Message string
}

// commands

type DamagePlayer struct {
    Damage int
}

type HealPlayer struct {
    Health int
}

// event application



// command handlers



```