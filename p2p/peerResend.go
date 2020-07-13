package p2p

import (
	"crypto/sha1"
	"sync"
	"time"
)

type PeerResend struct {
	mtx      sync.RWMutex
	buckets  []*PRBucket
	pc       int
	interval time.Duration
	stopper  chan interface{}
}

type PRBucket struct {
	mtx  sync.RWMutex
	data map[[sha1.Size]byte]bool
}

func (b *PRBucket) Has(hash [sha1.Size]byte) bool {
	b.mtx.RLock()
	defer b.mtx.RUnlock()
	return b.data[hash]
}

func (b *PRBucket) Add(hash [sha1.Size]byte) {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	b.data[hash] = true
}

func NewPeerResend(buckets int, interval time.Duration) *PeerResend {
	pr := new(PeerResend)
	pr.interval = interval
	pr.buckets = make([]*PRBucket, buckets)
	for i := 0; i < buckets; i++ {
		pr.buckets[i] = newPRBucket()
	}
	pr.stopper = make(chan interface{}, 1)
	go pr.Cleanup()
	return pr
}

func newPRBucket() *PRBucket {
	b := new(PRBucket)
	b.data = make(map[[sha1.Size]byte]bool)
	return b
}

func (pr *PeerResend) Has(hash [sha1.Size]byte) bool {
	pr.mtx.RLock()
	defer pr.mtx.RUnlock()

	for i := 0; i < len(pr.buckets); i++ {
		if pr.buckets[pr.ptr(i)].Has(hash) {
			return true
		}
	}
	return false
}

func (pr *PeerResend) Add(hash [sha1.Size]byte) {
	pr.mtx.RLock()
	defer pr.mtx.RUnlock()
	pr.buckets[pr.pc].Add(hash)
}

func (pr *PeerResend) Stop() {
	close(pr.stopper)
}

func (pr *PeerResend) Cleanup() {
	ticker := time.NewTicker(pr.interval)
	for {
		select {
		case <-pr.stopper:
			return
		case <-ticker.C:
			pr.dropOldestBucket()
		}
	}
}

func (pr *PeerResend) dropOldestBucket() {
	pr.mtx.Lock()
	pr.pc--
	if pr.pc < 0 {
		pr.pc += len(pr.buckets)
	}
	pr.buckets[pr.pc] = newPRBucket()
	pr.mtx.Unlock()
}

// not thread safe
func (pr *PeerResend) ptr(i int) int {
	return (pr.pc + i) % len(pr.buckets)
}
