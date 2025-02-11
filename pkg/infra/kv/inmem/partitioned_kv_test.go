package inmem

import (
	"math/rand/v2"
	"runtime"
	"sync"
	"testing"
)

const TestsPartitionCount = 6
const BenchmarkIterations = 1000000
const maxStrSize = 40

func TestPartitionedKVStorage_BasicReadWrite(t *testing.T) {
	t.Parallel()
	storage := NewPartitionedKVStorage(TestsPartitionCount)

	testData := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	for k, v := range testData {
		storage.Set(k, v)
	}

	for k, expectedVal := range testData {
		val, ok := storage.Get(k)
		if !ok {
			t.Errorf("Expected key %s to exist", k)
		}
		if val != expectedVal {
			t.Errorf("Expected value %s for key %s, got %s", expectedVal, k, val)
		}
	}
}

func TestPartitionedKVStorage_MissingKey(t *testing.T) {
	t.Parallel()
	storage := NewPartitionedKVStorage(TestsPartitionCount)

	_, ok := storage.Get("non-existent-key")
	if ok {
		t.Errorf("Expected missing key to return false, but got true")
	}
}

func TestPartitionedKVStorage_ConcurrentRead(t *testing.T) {
	t.Parallel()
	storage := NewPartitionedKVStorage(TestsPartitionCount)
	storage.Set("key", "value")

	var wg sync.WaitGroup
	const workers = 1000

	wg.Add(workers)
	for range workers {
		go func() {
			defer wg.Done()
			val, ok := storage.Get("key")
			if !ok || val != "value" {
				t.Errorf("Concurrent read failed: got %s, expected value", val)
			}
		}()
	}
	wg.Wait()
}

func randString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.IntN(len(letters))]
	}
	return string(b)
}

func TestPartitionedKVStorage_ConcurrentWrite(t *testing.T) {
	t.Parallel()
	storage := NewPartitionedKVStorage(TestsPartitionCount)
	var wg sync.WaitGroup

	wg.Add(BenchmarkIterations)
	for range BenchmarkIterations {
		go func() {
			randKey := randString(rand.IntN(maxStrSize))
			randVal := randString(rand.IntN(maxStrSize))
			storage.Set(randKey, randVal)
			wg.Done()
		}()
	}
	wg.Wait()
}

func BenchmarkPartitionedKVStorage(b *testing.B) {
	storage := NewPartitionedKVStorage(runtime.GOMAXPROCS(0) * 2)
	var wg sync.WaitGroup

	wg.Add(BenchmarkIterations)
	b.ResetTimer()
	for range BenchmarkIterations {
		go func() {
			randKey := randString(rand.IntN(maxStrSize))
			if write := rand.IntN(2); write == 1 {
				randVal := randString(rand.IntN(maxStrSize))
				storage.Set(randKey, randVal)
			} else {
				storage.Get(randKey)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func BenchmarkSyncMap(b *testing.B) {
	var m sync.Map
	var wg sync.WaitGroup

	wg.Add(BenchmarkIterations)
	b.ResetTimer()
	for range BenchmarkIterations {
		go func() {
			randKey := randString(rand.IntN(maxStrSize))
			if write := rand.IntN(2); write == 1 {
				randVal := randString(rand.IntN(maxStrSize))
				m.Store(randKey, randVal)
			} else {
				m.Load(randKey)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func BenchmarkMapRWMutex(b *testing.B) {
	m := make(map[string]string)
	var mtx sync.RWMutex
	var wg sync.WaitGroup

	wg.Add(BenchmarkIterations)
	b.ResetTimer()
	for range BenchmarkIterations {
		go func() {
			randKey := randString(rand.IntN(maxStrSize))
			if write := rand.IntN(2); write == 1 {
				randVal := randString(rand.IntN(maxStrSize))
				mtx.Lock()
				m[randKey] = randVal
				mtx.Unlock()
			} else {
				mtx.RLock()
				_ = m[randKey]
				mtx.RUnlock()
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
