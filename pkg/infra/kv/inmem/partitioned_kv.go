package inmem

import (
	"hash/fnv"
	"ozon_task/pkg/infra/kv"
)

type PartitionedKVStorage struct {
	partitions    []*Partition
	numPartitions int
}

func NewPartitionedKVStorage(numPartitions int) kv.Storage {
	partitions := make([]*Partition, numPartitions)
	for i := range numPartitions {
		partitions[i] = NewPartition()
	}

	return &PartitionedKVStorage{
		partitions:    partitions,
		numPartitions: numPartitions,
	}
}

func (ps *PartitionedKVStorage) getPartition(key string) *Partition {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	idx := h.Sum32() % uint32(ps.numPartitions)
	return ps.partitions[idx]
}

func (ps *PartitionedKVStorage) Set(key, val string) {
	partition := ps.getPartition(key)
	partition.Set(key, val)
}

func (ps *PartitionedKVStorage) Get(key string) (val string, ok bool) {
	partition := ps.getPartition(key)
	return partition.Get(key)
}
