package p2p

import (
	"crypto/sha1"
	"sync"
	"time"
)

// PeerHashCache keeps a list of the last X hashes for a specified interval.
// In this package, it is used to keep track of which messages were sent by a specific peer.
// Each peer has their own instance of PeerHashCache.
type PeerHashCache struct {
	mtx     sync.RWMutex
	buckets []*phcBucket
	pc      int
	stopper chan interface{}
}

// bucket that holds the hashes for a single interval
type phcBucket struct {
	mtx  sync.RWMutex
	data map[[sha1.Size]byte]bool
}

// Has returns true if the bucket holds the given hash
func (b *phcBucket) Has(hash [sha1.Size]byte) bool {
	b.mtx.RLock()
	defer b.mtx.RUnlock()
	return b.data[hash]
}

// Add a hash to the bucket
func (b *phcBucket) Add(hash [sha1.Size]byte) {
	b.mtx.Lock()
	b.data[hash] = true
	b.mtx.Unlock()
}

// NewPeerHashCache creates a new cache covering the time range of buckets * interval.
func NewPeerHashCache(buckets int, interval time.Duration) *PeerHashCache {
	pr := new(PeerHashCache)
	pr.buckets = make([]*phcBucket, buckets)
	for i := 0; i < buckets; i++ {
		pr.buckets[i] = newphcBucket()
	}
	pr.stopper = make(chan interface{}, 1)
	go pr.cleanup(interval)
	return pr
}

func newphcBucket() *phcBucket {
	b := new(phcBucket)
	b.data = make(map[[sha1.Size]byte]bool)
	return b
}

// Has returns true if the specified hash was seen.
func (pr *PeerHashCache) Has(hash [sha1.Size]byte) bool {
	pr.mtx.RLock()
	defer pr.mtx.RUnlock()

	for i := 0; i < len(pr.buckets); i++ {
		if pr.buckets[pr.ptr(i)].Has(hash) {
			return true
		}
	}
	return false
}

// Add a hash to the cache.
func (pr *PeerHashCache) Add(hash [sha1.Size]byte) {
	pr.mtx.RLock()
	pr.buckets[pr.pc].Add(hash)
	pr.mtx.RUnlock()
}

// Stop the PeerHashCache.
// Panics if called twice.
func (pr *PeerHashCache) Stop() {
	close(pr.stopper)
}

// drops oldest bucket at the end of every interval
func (pr *PeerHashCache) cleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-pr.stopper:
			return
		case <-ticker.C:
			pr.dropOldestBucket()
		}
	}
}

func (pr *PeerHashCache) dropOldestBucket() {
	pr.mtx.Lock()
	pr.pc--
	if pr.pc < 0 {
		pr.pc += len(pr.buckets)
	}
	pr.buckets[pr.pc] = newphcBucket()
	pr.mtx.Unlock()
}

// pointer to the current bucket + offset.
// not thread safe
func (pr *PeerHashCache) ptr(offset int) int {
	return (pr.pc + offset) % len(pr.buckets)
}
