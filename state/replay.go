// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package state

import (
	"fmt"
	"sync"
	"time"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

const Range = 60                // Double this for the period we protect, i.e. 120 means +/- 120 minutes
const numBuckets = Range*2 + 60 // Cover the rage in the future and in the past, with an hour buffer.

var _ = time.Now()
var _ = fmt.Print

type Replay struct {
	Mutex    sync.Mutex
	Buckets  [numBuckets]map[[32]byte]int
	Basetime int // hours since 1970
	Center   int // Hour of the current time.
}

func (r *Replay) Save() *Replay {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	newr := new(Replay)
	for i, b := range r.Buckets {
		if b != nil {
			newr.Buckets[i] = make(map[[32]byte]int, 0)
			for k := range b {
				newr.Buckets[i][k] = b[k]
			}
		}
	}
	newr.Basetime = r.Basetime
	newr.Center = r.Center
	return newr
}

// Remember that Unix time is in seconds since 1970.  This code
// wants to be handed time in seconds.
func Minutes(unix int64) int {
	return int(unix / 60)
}

// Returns false if the hash is too old, or is already a
// member of the set.  Timestamp is in seconds.
func (r *Replay) Valid(mask int, hash [32]byte, timestamp interfaces.Timestamp, systemtime interfaces.Timestamp) (index int, valid bool) {
	now := Minutes(systemtime.GetTimeSeconds())
	t := Minutes(timestamp.GetTimeSeconds())

	diff := now - t
	// Check the timestamp to see if within 12 hours of the system time.  That not valid, we are
	// just done without any added concerns.
	if diff > Range || diff < -Range {
		//fmt.Println("Time in hours, range:", hours(timeSeconds-systemTimeSeconds), HourRange)
		return -1, false
	}

	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	if mask == constants.TIME_TEST {
		return -1, true
	}

	// We don't let the system clock go backwards.  likely an attack if it does.
	// Move the current time up to r.center if it is in the past.
	if now < r.Center {
		now = r.Center
	}

	if r.Center == 0 {
		r.Center = now
		r.Basetime = r.Center - (numBuckets / 2)
	}
	for r.Center < now {
		for k := range r.Buckets[0] {
			delete(r.Buckets[0], k)
		}
		copy(r.Buckets[:], r.Buckets[1:])
		r.Buckets[numBuckets-1] = make(map[[32]byte]int)
		r.Center++
		r.Basetime++
	}

	// Just take the time of the thing in hours less the basetime to get the index.
	index = t - r.Basetime

	if index < 0 || index >= numBuckets {
		return -1, false
	}

	if r.Buckets[index] == nil {
		r.Buckets[index] = make(map[[32]byte]int)
	} else {
		v, _ := r.Buckets[index][hash]
		if v&mask > 0 {
			return index, false
		}
	}
	return index, true
}

// Checks if the timestamp is valid.  If the timestamp is too old or
// too far into the future, then we don't consider it valid.  Or if we
// have seen this hash before, then it is not valid.  To that end,
// this code remembers hashes tested in the past, and rejects the
// second submission of the same hash.
func (r *Replay) IsTSValid(mask int, hash interfaces.IHash, timestamp interfaces.Timestamp) bool {
	return r.IsTSValid_(mask, hash.Fixed(), timestamp, primitives.NewTimestampNow())
}

// To make the function testable, the logic accepts the current time
// as a parameter.  This way, the test code can manipulate the clock
// at will.
func (r *Replay) IsTSValid_(mask int, hash [32]byte, timestamp interfaces.Timestamp, now interfaces.Timestamp) bool {
	if index, ok := r.Valid(mask, hash, timestamp, now); ok {
		r.Mutex.Lock()
		defer r.Mutex.Unlock()
		// Mark this hash as seen
		if mask != constants.TIME_TEST {
			r.Buckets[index][hash] = r.Buckets[index][hash] | mask
		}
		return true
	}

	return false
}

// Returns True if there is no record of this hash in the Replay structures.
// Returns false if we have seen this hash before.
func (r *Replay) IsHashUnique(mask int, hash [32]byte) bool {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	for _, bucket := range r.Buckets {
		if bucket[hash]&mask > 0 {
			return false
		}
	}
	return true
}

func (r *Replay) SetHashNow(mask int, hash [32]byte, now interfaces.Timestamp) {
	if r.IsHashUnique(mask, hash) {
		index := Minutes(now.GetTimeSeconds()) - r.Basetime
		if index < 0 || index >= len(r.Buckets) {
			return
		}

		r.Mutex.Lock()
		defer r.Mutex.Unlock()

		if r.Buckets[index] == nil {
			r.Buckets[index] = make(map[[32]byte]int)
		}
		r.Buckets[index][hash] = mask | r.Buckets[index][hash]
	}
}

func (r *Replay) Clear(mask int, hash [32]byte) {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	for _, bucket := range r.Buckets {
		if bucket != nil {
			if v, ok := bucket[hash]; ok {
				bucket[hash] = v &^ mask
			}
		}
	}
}
