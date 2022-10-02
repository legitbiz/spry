package memory

import (
	"context"
	"sort"

	"github.com/arobson/spry"
	"github.com/arobson/spry/storage"
	"github.com/gofrs/uuid"
)

type IdLinks map[string]map[uuid.UUID]spry.AggregatedIds

type InMemoryCommandStore struct {
	Commands map[uuid.UUID][]storage.CommandRecord
}

func (store *InMemoryCommandStore) Add(
	ctx context.Context,
	actorType string,
	command storage.CommandRecord) error {
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

func (store *InMemoryEventStore) Add(ctx context.Context, events []storage.EventRecord) error {
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

func (store *InMemoryEventStore) FetchAggregatedSince(
	ctx context.Context,
	ids spry.AggregateIdMap,
	eventUUID uuid.UUID,
	types storage.TypeMap) ([]storage.EventRecord, error) {

	if store.Events == nil {
		store.Events = map[uuid.UUID][]storage.EventRecord{}
	}

	ok := false
	var records []storage.EventRecord
	if records, ok = store.Events[ids.ActorId]; !ok {
		records = []storage.EventRecord{}
	}

	for _, l := range ids.Aggregated {
		for _, i := range l {
			if e, ok := store.Events[i]; ok {
				records = append(records, e...)
			}
		}
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].Id.String() < records[j].Id.String()
	})

	return records, nil
}

func (store *InMemoryEventStore) FetchSince(
	ctx context.Context,
	actorType string,
	actorId uuid.UUID,
	eventUUID uuid.UUID,
	types storage.TypeMap) ([]storage.EventRecord, error) {
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
	IdMap   map[string]uuid.UUID
	LinkMap IdLinks
}

func (maps *InMemoryMapStore) AddId(ctx context.Context, actorType string, ids spry.Identifiers, uid uuid.UUID) error {
	if maps.IdMap == nil {
		maps.IdMap = map[string]uuid.UUID{}
	}
	key, _ := spry.IdentifiersToString(ids)
	maps.IdMap[key] = uid
	return nil
}

func (maps *InMemoryMapStore) AddLink(ctx context.Context, parentType string, parentId uuid.UUID, childType string, childId uuid.UUID) error {
	if maps.LinkMap == nil {
		maps.LinkMap = IdLinks{}
	}
	if maps.LinkMap[parentType] == nil {
		maps.LinkMap[parentType] = map[uuid.UUID]spry.AggregatedIds{}
	}
	if maps.LinkMap[parentType][parentId] == nil {
		maps.LinkMap[parentType][parentId] = spry.AggregatedIds{}
	}
	if maps.LinkMap[parentType][parentId][childType] == nil {
		maps.LinkMap[parentType][parentId][childType] = []uuid.UUID{childId}
	} else {
		maps.LinkMap[parentType][parentId][childType] = append(maps.LinkMap[parentType][parentId][childType], childId)
	}
	return nil
}

func (maps *InMemoryMapStore) GetId(ctx context.Context, actorType string, ids spry.Identifiers) (uuid.UUID, error) {
	if maps.IdMap == nil {
		maps.IdMap = map[string]uuid.UUID{}
	}
	key, _ := spry.IdentifiersToString(ids)
	uid := maps.IdMap[key]
	if uid == uuid.Nil {
		return uuid.Nil, nil
	}
	return uid, nil
}

func (maps *InMemoryMapStore) GetIdMap(ctx context.Context, actorType string, uid uuid.UUID) (spry.AggregateIdMap, error) {
	if maps.IdMap == nil {
		maps.IdMap = map[string]uuid.UUID{}
	}

	ok := false
	var aggregates map[uuid.UUID]spry.AggregatedIds
	if aggregates, ok = maps.LinkMap[actorType]; !ok {
		aggregates = map[uuid.UUID]spry.AggregatedIds{}
		maps.LinkMap[actorType] = aggregates
	}

	idMap := spry.CreateAggregateIdMap(actorType, uid)
	if actors, ok := aggregates[uid]; ok {
		for k, v := range actors {
			idMap.AddIdsFor(k, v...)
		}
	}

	return idMap, nil
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

func (provider InMemoryTxProvider) Commit(ctx context.Context) error {
	return nil
}

func (provider InMemoryTxProvider) GetTransaction(ctx context.Context) (storage.NoOpTx, error) {
	return storage.NoOpTx{}, nil
}

func (provider InMemoryTxProvider) Rollback(ctx context.Context) error {
	return nil
}

func InMemoryStorage() storage.Storage {
	return storage.NewStorage[storage.NoOpTx](
		&InMemoryCommandStore{},
		&InMemoryEventStore{},
		&InMemoryMapStore{},
		&InMemorySnapshotStore{},
		&InMemoryTxProvider{},
	)
}
