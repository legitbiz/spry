package tests

import (
	"testing"

	"github.com/arobson/spry/memory"
	"github.com/arobson/spry/storage"
)

func TestGetRepositoryFor(t *testing.T) {
	store := memory.InMemoryStorage()
	repo := storage.GetRepositoryFor[Player](store)
	if repo.ActorName != "Player" {
		t.Error("actor name was not the expected type")
	}
}

func TestActorHandlesCommandSuccessfully(t *testing.T) {
	store := memory.InMemoryStorage()
	repo := storage.GetRepositoryFor[Player](store)
	results := repo.Handle(CreatePlayer{Name: "Bob"})

	// create player
	expected := PlayerCreated{Name: "Bob"}
	if len(results.Events) == 0 ||
		results.Events[0].(PlayerCreated) != expected {
		t.Error("event was not generated or did not match expected output")
	}
	if results.Original.Name != "" {
		t.Error("original actor instance was modified but should not have been")
	}
	if results.Modified.Name != "Bob" {
		t.Error("modified actor did not contain expected state")
	}

	// damage player
	expected2 := PlayerDamaged{Damage: 40}
	results2 := repo.Handle(DamagePlayer{Name: "Bob", Damage: 40})
	if len(results2.Events) == 0 ||
		results2.Events[0].(PlayerDamaged) != expected2 {
		t.Error("event was not generated or did not match expected output")
	}
	if results2.Original.HitPoints != 100 {
		t.Error("original actor instance was modified but should not have been")
	}
	if results2.Modified.HitPoints != 60 {
		t.Error("modified actor did not contain expected state")
	}
	if results2.Modified.Name != "Bob" {
		t.Error("incorrect creation of new actor occurred")
	}

	// heal player
	expected3 := PlayerHealed{Health: 10}
	results3 := repo.Handle(HealPlayer{Name: "Bob", Health: 10})
	if len(results3.Events) == 0 ||
		results3.Events[0].(PlayerHealed) != expected3 {
		t.Error("event was not generated or did not match expected output")
	}
	if results3.Original.HitPoints != 60 {
		t.Error("original actor instance was modified but should not have been")
	}
	if results3.Modified.HitPoints != 70 {
		t.Errorf("expected player health to = %d but was %d", 70, results.Modified.HitPoints)
	}
	if results3.Modified.Name != "Bob" {
		t.Error("incorrect creation of new actor occurred")
	}
}

//func TestAggregateHandlesCommandSuccessfully(t *testing.T) {
//	store := memory.InMemoryStorage()
//	motorists := storage.GetRepositoryFor[Motorist](store)
//	vehicles := storage.GetRepositoryFor[Vehicle](store)
//
//	r1 := motorists.Handle()
//
//	r2 := motorists.Handle(RegisterVehicle{
//		VIN:          "abc123",
//		Type:         "Moped",
//		Make:         "Hyundai",
//		Model:        "Scootchum",
//		Color:        "Blurple",
//		OwnerLicense: "001020304",
//		OwnerState:   "KS",
//	})
//
//}
