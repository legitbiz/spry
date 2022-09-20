package tests

import (
	"testing"

	"github.com/arobson/spry/postgres"
	"github.com/arobson/spry/storage"
	"github.com/arobson/spry/tests"
)

func TestHandleCommandSuccessfully(t *testing.T) {
	store := postgres.CreatePostgresStorage(CONNECTION_STRING)
	store.RegisterPrimitives(
		tests.PlayerCreated{},
		tests.PlayerDamaged{},
		tests.PlayerHealed{},
	)

	t.Cleanup(func() {
		_ = TruncateTable("player_commands")
		_ = TruncateTable("player_events")
		_ = TruncateTable("player_id_map")
		_ = TruncateTable("player_snapshots")
	})

	repo := storage.GetRepositoryFor[tests.Player](store)
	results := repo.Handle(tests.CreatePlayer{Name: "Bob"})

	// create player
	expected := tests.PlayerCreated{Name: "Bob"}
	if len(results.Events) == 0 ||
		results.Events[0].(tests.PlayerCreated) != expected {
		t.Error("event was not generated or did not match expected output")
	}
	if results.Original.Name != "" {
		t.Error("original actor instance was modified but should not have been")
	}
	if results.Modified.Name != "Bob" {
		t.Error("modified actor did not contain expected state")
	}

	// damage player
	expected2 := tests.PlayerDamaged{Damage: 40}
	results2 := repo.Handle(tests.DamagePlayer{Name: "Bob", Damage: 40})
	if len(results2.Events) == 0 ||
		results2.Events[0].(tests.PlayerDamaged) != expected2 {
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
	expected3 := tests.PlayerHealed{Health: 10}
	results3 := repo.Handle(tests.HealPlayer{Name: "Bob", Health: 10})
	if len(results3.Events) == 0 ||
		results3.Events[0].(tests.PlayerHealed) != expected3 {
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
