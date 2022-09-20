package memory

import (
	"context"

	"github.com/arobson/spry"
	"github.com/arobson/spry/storage"
	"github.com/gofrs/uuid"
)

type InMemoryCommandStore struct {
	Commands map[uuid.UUID][]storage.CommandRecord
}

func (store *InMemoryCommandStore) Add(ctx context.Context, actorType string, command storage.CommandRecord) error {
	if store.Commands == nil {
		store.Commands = map[uuid.UUID][]storage.CommandRecord{}
	}
	actorId := command.HandledBy
	if stored, ok := store.Commands[actorId]; ok {
		store.Commands[actorId] = append(stored, command)
	} else {
		store.Commands[actorId] = []storage.CommandRecord{command}
	}
	return nil
}

type InMemoryEventStore struct {
	Events map[uuid.UUID][]storage.EventRecord
}

func (store *InMemoryEventStore) Add(ctx context.Context, actorType string, events []storage.EventRecord) error {
	if store.Events == nil {
		store.Events = map[uuid.UUID][]storage.EventRecord{}
	}
	for _, event := range events {
		actorId := event.ActorId
		if stored, ok := store.Events[actorId]; ok {
			store.Events[actorId] = append(stored, event)
		} else {
			store.Events[actorId] = []storage.EventRecord{event}
		}
	}
	return nil
}

func (store *InMemoryEventStore) FetchSince(ctx context.Context, actorType string, actorId uuid.UUID, eventUUID uuid.UUID) ([]storage.EventRecord, error) {
	if store.Events == nil {
		store.Events = map[uuid.UUID][]storage.EventRecord{}
	}
	if stored, ok := store.Events[actorId]; ok {
		return stored, nil
	} else {
		return []storage.EventRecord{}, nil
	}
}

type InMemoryMapStore struct {
	IdMap map[string]uuid.UUID
}

func (maps *InMemoryMapStore) Add(ctx context.Context, actorType string, ids spry.Identifiers, uid uuid.UUID) error {
	if maps.IdMap == nil {
		maps.IdMap = map[string]uuid.UUID{}
	}
	key, _ := spry.IdMapToString(ids)
	maps.IdMap[key] = uid
	return nil
}

func (maps *InMemoryMapStore) GetId(ctx context.Context, actorType string, ids spry.Identifiers) (uuid.UUID, error) {
	if maps.IdMap == nil {
		maps.IdMap = map[string]uuid.UUID{}
	}
	key, _ := spry.IdMapToString(ids)
	uid := maps.IdMap[key]
	if uid == uuid.Nil {
		return uuid.Nil, nil
	}
	return uid, nil
}

type InMemorySnapshotStore struct {
	Snapshots map[uuid.UUID][]storage.Snapshot
}

func (store *InMemorySnapshotStore) Add(ctx context.Context, actorType string, snapshot storage.Snapshot, allowPartition bool) error {
	if store.Snapshots == nil {
		store.Snapshots = map[uuid.UUID][]storage.Snapshot{}
	}
	actorId := snapshot.ActorId
	if stored, ok := store.Snapshots[actorId]; ok {
		store.Snapshots[actorId] = append(stored, snapshot)
	} else {
		store.Snapshots[actorId] = []storage.Snapshot{snapshot}
	}
	return nil
}

func (store *InMemorySnapshotStore) Fetch(ctx context.Context, actorType string, actorId uuid.UUID) (storage.Snapshot, error) {
	if store.Snapshots == nil {
		store.Snapshots = map[uuid.UUID][]storage.Snapshot{}
	}
	if stored, ok := store.Snapshots[actorId]; ok {
		return stored[len(stored)-1], nil
	}
	return storage.Snapshot{}, nil
}

type InMemoryTxProvider struct {
}

func (provider InMemoryTxProvider) GetTransaction(ctx context.Context) (storage.NoOpTx, error) {
	return storage.NoOpTx{}, nil
}

func InMemoryStorage() storage.Storage {
	return storage.NewStorage[storage.NoOpTx](
		&InMemoryMapStore{},
		&InMemoryCommandStore{},
		&InMemoryEventStore{},
		&InMemorySnapshotStore{},
		&InMemoryTxProvider{},
	)
}
