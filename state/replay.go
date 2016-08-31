// Copyright 2015 FactomProject Authors. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package state

import (
	"fmt"
	"sync"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

const HourRange = 4                // Double this for the period we protect, i.e. 4 means +/- 4 hours
const numBuckets = HourRange*2 + 3 // cover an hour each way, and an hour in the middle.

var _ = time.Now()
var _ = fmt.Print

type Replay struct {
	mutex    sync.Mutex
	buckets  [numBuckets]map[[32]byte]int
	basetime int // hours since 1970
	center   int // Hour of the current time.
}

// Remember that Unix time is in seconds since 1970.  This code
// wants to be handed time in seconds.
func hours(unix int64) int {
	return int(unix / 60 / 60)
}

// Returns false if the hash is too old, or is already a
// member of the set.  Timestamp is in seconds.
func (r *Replay) Valid(mask int, hash [32]byte, timestamp interfaces.Timestamp, systemtime interfaces.Timestamp) (index int, valid bool) {
	timeSeconds := timestamp.GetTimeSeconds()
	systemTimeSeconds := systemtime.GetTimeSeconds()
	// Check the timestamp to see if within 12 hours of the system time.  That not valid, we are
	// just done without any added concerns.
	if hours(timeSeconds-systemTimeSeconds) > HourRange || hours(systemTimeSeconds-timeSeconds) > HourRange {
		//fmt.Println("Time in hours, range:", hours(timeSeconds-systemTimeSeconds), HourRange)
		return -1, false
	}

	now := hours(systemTimeSeconds)
	t := hours(timeSeconds)

	r.mutex.Lock()
	defer r.mutex.Unlock()

	// We don't let the system clock go backwards.  likely an attack if it does.
	if now < r.center {
		now = r.center
	}

	if r.center == 0 {
		r.center = now
		r.basetime = now - (numBuckets / 2) + 1
	}
	for r.center < now {
		copy(r.buckets[:], r.buckets[1:])
		r.buckets[numBuckets-1] = nil
		r.center++
		r.basetime++
	}

	// Just take the time of the thing in hours less the basetime to get the index.
	index = t - r.basetime

	if index < 0 || index >= numBuckets {
		return 0, false
	}

	if r.buckets[index] == nil {
		r.buckets[index] = make(map[[32]byte]int)
	} else {
		v, _ := r.buckets[index][hash]
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
		r.mutex.Lock()
		defer r.mutex.Unlock()
		// Mark this hash as seen
		r.buckets[index][hash] = r.buckets[index][hash] | mask
		return true
	}

	return false
}
