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

const HourRange = 1                    // Double this for the period we protect, i.e. 4 means +/- 4 hours
const numBuckets = HourRange*2*60 + 60 // cover an hour each way, and an hour in the middle.

var _ = time.Now()
var _ = fmt.Print

type Replay struct {
	Mutex    sync.Mutex
	Buckets  [numBuckets]map[[32]byte]int
	Basetime int // hours since 1970
	Center   int // Hour of the current time.
}

// Remember that Unix time is in seconds since 1970.  This code
// wants to be handed time in seconds.
func Minutes(unix int64) int {
	return int(unix / 60)
}

// Returns false if the hash is too old, or is already a
// member of the set.  Timestamp is in seconds.
func (r *Replay) Valid(mask int, hash [32]byte, timestamp interfaces.Timestamp, systemtime interfaces.Timestamp) (index int, valid bool) {
	timeSeconds := timestamp.GetTimeSeconds()
	systemTimeSeconds := systemtime.GetTimeSeconds()
	diff := Minutes(systemTimeSeconds) - Minutes(timeSeconds)
	// Check the timestamp to see if within 12 hours of the system time.  That not valid, we are
	// just done without any added concerns.
	if diff >= Minutes(HourRange*60*60) || diff <= -Minutes(HourRange*60*60) {
		//fmt.Println("Time in hours, range:", hours(timeSeconds-systemTimeSeconds), HourRange)
		return -1, false
	}

	now := Minutes(systemTimeSeconds)
	t := Minutes(timeSeconds)

	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	// We don't let the system clock go backwards.  likely an attack if it does.
	// Move the current time up to r.center if it is in the past.
	if now < r.Center {
		now = r.Center
	}

	if r.Center == 0 {
		r.Center = now
		r.Basetime = now - (numBuckets / 2) + 1
	}
	for r.Center < now {
		copy(r.Buckets[:], r.Buckets[1:])
		r.Buckets[numBuckets-1] = nil
		r.Center++
		r.Basetime++
	}

	// Just take the time of the thing in hours less the basetime to get the index.
	index = t - r.Basetime

	if index < 0 || index >= numBuckets {
		return 0, false
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
		r.Buckets[index][hash] = r.Buckets[index][hash] | mask
		return true
	}

	return false
}
