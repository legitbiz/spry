package tests

import (
	"strings"
	"testing"

	"github.com/arobson/spry/storage"
)

func TestCreateTemplate(t *testing.T) {
	type Data struct {
		ActorName string
	}
	result, err := storage.CreateFromTemplate(
		"create_actor_schema.sql",
		"../postgres/sql/create_actor_schema.sql",
		Data{ActorName: strings.ToLower("TesT")},
	)
	if err != nil {
		t.Error("failed to create schema from template", err)
	}
	if result == "" {
		t.Error(result)
	}
}
