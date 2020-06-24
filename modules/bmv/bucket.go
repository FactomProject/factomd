package bmv

import (
	"sync"
	"time"
)

type bucket struct {
	time time.Time
	mtx  sync.RWMutex
	data map[[32]byte]time.Time
}

func newBucket() *bucket {
	b := new(bucket)
	b.data = make(map[[32]byte]time.Time)
	return b
}

func (b *bucket) Set(key [32]byte, time time.Time) {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	b.data[key] = time
}

func (b *bucket) Get(key [32]byte) (time.Time, bool) {
	b.mtx.RLock()
	defer b.mtx.RUnlock()
	t, ok := b.data[key]
	return t, ok
}

// Time returns the cutoff time for this bucket
func (b *bucket) Time() time.Time {
	b.mtx.RLock()
	defer b.mtx.RUnlock()
	return b.time
}
func (b *bucket) SetTime(t time.Time) {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	b.time = t
}

// Transfer takes another bucket and transfers all items from the other
// bucket to b that happened before its time
func (b *bucket) Transfer(other *bucket) {
	b.mtx.Lock()
	other.mtx.Lock()
	defer b.mtx.Unlock()
	defer other.mtx.Unlock()

	// move items from other to us that belong here
	for k, v := range other.data {
		if v.Before(other.time) {
			b.data[k] = v
			delete(other.data, k)
		}
	}
}
