package tests

import (
	"context"
	"fmt"

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
