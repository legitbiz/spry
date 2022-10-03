package spry

type ActorMeta struct {
	// how many events should occur before the next snapshot
	SnapshotFrequency int
	// controls whether snapshots occur during fetch (read)
	SnapshotDuringRead bool
	// controls whether snapshots occur during handle (write)
	SnapshotDuringWrite bool
	// controls whether snapshots can occur during partitions
	// requires a storage adapter for a database that can
	// detect this
	SnapshotDuringPartition bool
}

type HasMeta interface {
	GetActorMeta() ActorMeta
}

var default_meta = ActorMeta{
	SnapshotFrequency:       20,
	SnapshotDuringRead:      false,
	SnapshotDuringWrite:     true,
	SnapshotDuringPartition: true,
}

func GetActorMeta[T any]() ActorMeta {
	var empty any = getEmpty[T]()
	hasMeta, ok := empty.(HasMeta)
	if ok {
		return hasMeta.GetActorMeta()
	}
	return default_meta
}

type EventMetadata struct {
	CreatedBy  string
	CreatedFor string
}

func (e EventMetadata) GetEventMeta() EventMetadata {
	return e
}

func GetEventMeta(event any) EventMetadata {
	hasMeta, ok := event.(Namespaced)
	if ok {
		return hasMeta.GetEventMeta()
	}
	return EventMetadata{}
}

type Namespaced interface {
	GetEventMeta() EventMetadata
}
