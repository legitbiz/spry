package tests

import (
	"testing"
	"time"

	"github.com/arobson/spry"
	"github.com/arobson/spry/storage"
)

func TestEventRecordSerialization(t *testing.T) {
	uid, err := storage.GetId()
	if err != nil {
		t.Fatal("could not create UUID")
	}
	buffer, err := spry.ToJson(storage.EventRecord{
		Id:               uid,
		Type:             "ThingHappened",
		ActorId:          uid,
		ActorType:        "Test",
		CreatedOn:        time.Now(),
		CreatedByVersion: 0,
		CreatedBy:        "Test",
		CreatedById:      uid,
	})
	if err != nil {
		t.Fatal("failed to serialize record", err)
	}

	record, err := spry.FromJson[storage.EventRecord](buffer)
	if err != nil {
		t.Fatal("failed to deserialize record from json", err)
	}
	if record.Id != uid ||
		record.Type != "ThingHappened" ||
		record.ActorId != uid ||
		record.ActorType != "Test" ||
		record.CreatedByVersion != 0 ||
		record.CreatedBy != "Test" ||
		record.CreatedById != uid {
		t.Fatal("deserialization created invalid object", err)
	}
}
