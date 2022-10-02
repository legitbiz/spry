package tests

import (
	"testing"

	"github.com/arobson/spry"
	"github.com/arobson/spry/storage"
	"github.com/gofrs/uuid"
)

func TestAggregateIds(t *testing.T) {

	ids := spry.CreateAggregateIdMap("parent", uuid.Nil)

	for i := 0; i < 3; i++ {
		id, _ := storage.GetId()
		ids.AddIdsFor("one", id)
	}

	for i := 0; i < 3; i++ {
		id, _ := storage.GetId()
		ids.AddIdsFor("two", id)
	}

	if len(ids.Aggregated["one"]) < 3 {
		t.Error("id set one should have had 3 ids")
	}

	if len(ids.Aggregated["two"]) < 3 {
		t.Error("id set one should have had 3 ids")
	}
}

func TestLastIdMapMerge(t *testing.T) {

}
