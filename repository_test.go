package spry

import (
	"testing"
)

func TestGetRepositoryFor(t *testing.T) {
	repo := GetRepositoryFor[Player]()
	if repo.ActorName != "Player" {
		t.Error("actor name was not the expected type")
	}
}

func TestHandleCommandSuccessfully(t *testing.T) {
	repo := GetRepositoryFor[Player]()
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

	expected2 := PlayerDamaged{Damage: -10}
	results2 := repo.Handle(DamagePlayer{Damage: -10})
	if len(results2.Events) == 0 ||
		results2.Events[0].(PlayerDamaged) != expected2 {
		t.Error("event was not generated or did not match expected output")
	}
	if results2.Original.HitPoints != 0 {
		t.Error("original actor instance was modified but should not have been")
	}
	if results2.Modified.HitPoints != 10 {
		t.Error("modified actor did not contain expected state")
	}
}
