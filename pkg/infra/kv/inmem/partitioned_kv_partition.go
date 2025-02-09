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
	defer p.m.Unlock()

	if oldVal, exists := p.bucket[key]; exists {
		delete(p.reverseBucket, oldVal)
	}

	p.bucket[key] = val
	p.reverseBucket[val] = key
}

func (p *Partition) Get(key string) (string, bool) {
	p.m.RLock()
	defer p.m.RUnlock()
	val, ok := p.bucket[key]
	return val, ok
}

func (p *Partition) GetByValue(val string) (string, bool) {
	p.m.RLock()
	defer p.m.RUnlock()
	key, ok := p.reverseBucket[val]
	return key, ok
}
