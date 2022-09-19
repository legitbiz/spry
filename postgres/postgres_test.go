package postgres

import (
	"context"
	"fmt"
	"testing"

	"github.com/arobson/spry/storage"
	"github.com/arobson/spry/tests"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
)

var CONNECTION_STRING = "postgres://spry:yippyskippy@localhost:5540/sprydb"

func TruncateTable(tableName string) error {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, CONNECTION_STRING)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)
	_, err = conn.Exec(
		ctx,
		fmt.Sprintf("TRUNCATE TABLE %s;", tableName),
	)
	if err != nil {
		return err
	}
	return nil
}

func TestEventStorage(t *testing.T) {
	store := CreatePostgresStorage(
		CONNECTION_STRING,
	)

	aid1, _ := storage.GetId()

	e1 := tests.PlayerCreated{Name: "Bill"}
	er1, _ := storage.NewEventRecord(e1)
	er1.ActorId = aid1
	er1.ActorType = "Player"
	er1.CreatedBy = "Player"
	er1.CreatedById = aid1
	er1.Id, _ = storage.GetId()
	er1.Data = e1
	er1.Type = "PlayerCreated"

	e2 := tests.PlayerDamaged{Damage: 10}
	er2, _ := storage.NewEventRecord(e2)
	er2.ActorId = aid1
	er2.ActorType = "Player"
	er2.CreatedBy = "Player"
	er2.CreatedById = aid1
	er2.Id, _ = storage.GetId()
	er2.Data = e1
	er2.Type = "PlayerCreated"

	err := store.AddEvents("Player", []storage.EventRecord{
		er1,
		er2,
	})

	if err != nil {
		t.Error(err)
	}

	records, err := store.FetchEventsSince("player", aid1, uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 {
		t.Fatalf("expected %d records but got %d instead", 2, len(records))
	}

	err = TruncateTable("player_events")
	if err != nil {
		t.Error(err)
	}
}
