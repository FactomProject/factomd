package p2p

import (
	"fmt"
	"math/rand"
	"sync"
)

// golang's rand.Rand instances are not thread safe
// this wraps the unsafe functions in a mutex
type lockSource struct {
	src rand.Source64
	mtx sync.Mutex
}

func newLockSource(seed int64) (*lockSource, error) {
	src, ok := rand.NewSource(seed).(rand.Source64)
	if !ok {
		return nil, fmt.Errorf("golang version incompatibility, expected math/ran.NewSource to be Source64")
	}
	ls := new(lockSource)
	ls.src = src
	return ls, nil
}

func (r *lockSource) Int63() (n int64) {
	r.mtx.Lock()
	n = r.src.Int63()
	r.mtx.Unlock()
	return
}

func (r *lockSource) Uint64() (n uint64) {
	r.mtx.Lock()
	n = r.src.Uint64()
	r.mtx.Unlock()
	return
}

func (r *lockSource) Seed(seed int64) {
	r.mtx.Lock()
	r.src.Seed(seed)
	r.mtx.Unlock()
}
