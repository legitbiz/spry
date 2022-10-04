package tests

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
)

var CONNECTION_STRING = "postgres://spry:yippyskippy@localhost:5540/sprydb"

func TruncateTables(tableNames ...string) error {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, CONNECTION_STRING)
	if err != nil {
		fmt.Print("uh oh", err)
		return err
	}
	defer conn.Close(ctx)
	for _, tableName := range tableNames {
		_, err = conn.Exec(
			ctx,
			fmt.Sprintf("DELETE FROM %s;", tableName),
		)
		if err != nil {
			fmt.Println("beans :(", err)
		}
	}
	return nil
}
