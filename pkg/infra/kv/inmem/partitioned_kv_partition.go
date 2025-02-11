package inmem

import "sync"

type Partition struct {
	bucket        map[string]string
	reverseBucket map[string]string
	m             sync.RWMutex
}

func NewPartition() *Partition {
	return &Partition{
		bucket:        make(map[string]string),
		reverseBucket: make(map[string]string),
	}
}

func (p *Partition) Set(key, val string) {
	p.m.Lock()

	if oldVal, exists := p.bucket[key]; exists {
		delete(p.reverseBucket, oldVal)
	}

	p.bucket[key] = val
	p.reverseBucket[val] = key
	p.m.Unlock()
}

func (p *Partition) Get(key string) (string, bool) {
	p.m.RLock()
	val, ok := p.bucket[key]
	p.m.RUnlock()
	return val, ok
}
