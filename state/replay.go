// Copyright 2015 FactomProject Authors. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package state

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"sync"
	"time"
)

const numBuckets = 27

var _ = time.Now()
var _ = fmt.Print

type Replay struct {
	mutex    sync.Mutex
	buckets  [numBuckets]map[[32]byte]byte
	basetime int // hours since 1970
	center   int // Hour of the current time.
	check    map[[32]byte]byte
}

// Remember that Unix time is in seconds since 1970.  This code
// wants to be handed time in seconds.
func hours(unix int64) int {
	return int(unix / 60 / 60)
}

// Returns false if the hash is too old, or is already a
// member of the set.  Timestamp is in seconds.
func (r *Replay) Valid(hash [32]byte, timestamp interfaces.Timestamp, systemtime interfaces.Timestamp) (index int, valid bool) {
	timeSeconds := timestamp.GetTimeSeconds()
	systemTimeSeconds := systemtime.GetTimeSeconds()
	// Check the timestamp to see if within 12 hours of the system time.  That not valid, we are
	// just done without any added concerns.
	if timeSeconds-systemTimeSeconds > 60*60*12 || systemTimeSeconds-timeSeconds > 60*60*12 {
		return -1, false
	}

	_, okc := r.check[hash]

	now := hours(systemTimeSeconds)

	// We don't let the system clock go backwards.  likely an attack if it does.
	if now < r.center {
		now = r.center
	}

	if r.center == 0 {
		r.center = now
		r.basetime = now - (numBuckets / 2)
		r.check = make(map[[32]byte]byte, 0)
	}
	for r.center < now {
		copy(r.buckets[:], r.buckets[1:])
		r.buckets[numBuckets-1] = nil
		r.center++
		r.basetime++
	}

	t := hours(timeSeconds)
	index = t - r.basetime
	if index < 0 || index >= numBuckets {
		fmt.Println("dddd Timestamp false on time:", index)
		return 0, false
	}

	if r.buckets[index] == nil {
		r.buckets[index] = make(map[[32]byte]byte)
	} else {
		_, ok := r.buckets[index][hash]
		if ok {
			if !okc {
				panic(fmt.Sprintf("dddd Replay Failure returns false %x %d", hash, timestamp))
			}
			return index, false
		}
	}
	if okc {
		panic(fmt.Sprintf("dddd Replay Failure returns true %x %d", hash, timestamp))
	}
	return index, true
}

// Checks if the timestamp is valid.  If the timestamp is too old or
// too far into the future, then we don't consider it valid.  Or if we
// have seen this hash before, then it is not valid.  To that end,
// this code remembers hashes tested in the past, and rejects the
// second submission of the same hash.
func (r *Replay) IsTSValid(hash interfaces.IHash, timestamp interfaces.Timestamp) bool {
	return r.IsTSValid_(hash.Fixed(), timestamp, *interfaces.NewTimestampNow())
}

// To make the function testable, the logic accepts the current time
// as a parameter.  This way, the test code can manipulate the clock
// at will.
func (r *Replay) IsTSValid_(hash [32]byte, timestamp interfaces.Timestamp, now interfaces.Timestamp) bool {

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if index, ok := r.Valid(hash, timestamp, now); ok {
		// Mark this hash as seen
		r.buckets[index][hash] = 'x'
		r.check[hash] = 'x'
		return true
	}

	return false
}
