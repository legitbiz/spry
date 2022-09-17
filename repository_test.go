package spry

import (
	"context"
	"fmt"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
)

func TestFetchingIds(t *testing.T) {
	ids := make([]uuid.UUID, 1000)
	for i := range ids {
		id, err := GetId()
		if err != nil {
			t.Error("failed to generate an id")
		}
		ids[i] = id
	}
	for i := 1; i < len(ids); i++ {
		p, n := ids[i-1], ids[i]
		if p == n {
			t.Error("duplicate ids :@")
		} else if p.String() > n.String() {
			t.Errorf("ids are not in increasing order %s <= %s",
				n, p)
		} else {
			fmt.Printf("%s\n", p)
		}
	}
}

func TestSqlOrdering(t *testing.T) {
	connectionString := "postgres://spry:yippyskippy@localhost:5540/sprydb"
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, connectionString)
	if err != nil {
		t.Error("could not connect to the databass", err)
	}

	ids := make([]uuid.UUID, 10000)
	for i := range ids {
		id, err := GetId()
		if err != nil {
			t.Error("failed to generate an id")
		}

		ids[i] = id
		tag, err := conn.Exec(
			ctx,
			"INSERT INTO id_sort_test (id, ordering) VALUES ($1, $2)",
			id,
			i,
		)
		if err != nil || tag.RowsAffected() < 1 {
			t.Error("failed to insert row", err)
		}
	}
}

func TestGetRepositoryFor(t *testing.T) {
	storage := InMemoryStorage()
	repo := GetRepositoryFor[Player](storage)
	if repo.ActorName != "Player" {
		t.Error("actor name was not the expected type")
	}
}

func TestHandleCommandSuccessfully(t *testing.T) {
	storage := InMemoryStorage()
	repo := GetRepositoryFor[Player](storage)
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

	expected2 := PlayerDamaged{Damage: 10}
	results2 := repo.Handle(DamagePlayer{Name: "Bob", Damage: 10})
	if len(results2.Events) == 0 ||
		results2.Events[0].(PlayerDamaged) != expected2 {
		t.Error("event was not generated or did not match expected output")
	}
	if results2.Original.HitPoints != 100 {
		t.Error("original actor instance was modified but should not have been")
	}
	if results2.Modified.HitPoints != 90 {
		t.Error("modified actor did not contain expected state")
	}
	if results2.Modified.Name != "Bob" {
		t.Error("incorrect creation of new actor occurred")
	}
}
