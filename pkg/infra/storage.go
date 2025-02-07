package infra

import (
	"hash/fnv"
	"sync"
)

type KVStorage interface {
	Set(key, val string)
	Get(key string) (val string, ok bool)
}

type PartitionedStorage struct {
	partitions    []*Partition
	numPartitions int
}

func NewPartitionedStorage(numPartitions int) *PartitionedStorage {
	partitions := make([]*Partition, numPartitions)
	for i := 0; i < numPartitions; i++ {
		partitions[i] = NewPartition()
	}
	return &PartitionedStorage{
		partitions:    partitions,
		numPartitions: numPartitions,
	}
}

func (ps *PartitionedStorage) getPartition(key string) *Partition {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	idx := int(h.Sum32()) % ps.numPartitions
	return ps.partitions[idx]
}

func (ps *PartitionedStorage) Set(key, val string) {
	partition := ps.getPartition(key)
	partition.Set(key, val)
}

func (ps *PartitionedStorage) Get(key string) (val string, ok bool) {
	partition := ps.getPartition(key)
	return partition.Get(key)
}

type Partition struct {
	bucket map[string]string
	m      *sync.RWMutex
}

func NewPartition() *Partition {
	return &Partition{
		bucket: make(map[string]string),
	}
}

func (p *Partition) Set(key, val string) {
	p.m.Lock()
	defer p.m.Unlock()
	p.bucket[key] = val
}

func (p *Partition) Get(key string) (string, bool) {
	p.m.RLock()
	defer p.m.Unlock()
	val, ok := p.bucket[key]
	return val, ok
}
