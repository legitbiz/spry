package postgres

import (
	"context"
	"embed"
	"fmt"

	"github.com/arobson/spry/storage"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type QueryData struct {
	ActorName string
}

//go:embed sql
var sqlFiles embed.FS

func queryData(name string) QueryData {
	return QueryData{ActorName: name}
}

func CreatePostgresStorage(connectionURI string) storage.Storage {
	pool, err := pgxpool.Connect(context.Background(), connectionURI)
	if err != nil {
		fmt.Println("failed to connect to the backing store")
		panic("oh no")
	}

	// load templates
	templates, err := storage.CreateTemplateFromFS(
		sqlFiles,
		"sql/insert_command.sql",
		"sql/insert_event.sql",
		"sql/insert_map.sql",
		"sql/insert_snapshot.sql",
		"sql/select_events_since.sql",
		"sql/select_latest_snapshot.sql",
		"sql/select_id_by_map.sql",
	)

	if err != nil {
		fmt.Println("failed to read sql templates")
		panic("oh no")
	}

	return storage.Stores[pgx.Tx]{
		Commands:     &PostgresCommandStore{Templates: *templates, Pool: pool},
		Events:       &PostgresEventStore{Templates: *templates, Pool: pool},
		Maps:         &PostgresMapStore{Templates: *templates, Pool: pool},
		Snapshots:    &PostgresSnapshotStore{Templates: *templates, Pool: pool},
		Transactions: &PostgresTxProvider{Pool: pool},
	}
}
