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

const numBuckets = 24

var _ = time.Now()
var _ = fmt.Print

type Replay struct {
	mutex    sync.Mutex
	buckets  []map[[32]byte]int64
	lasttime int64 // hours since 1970
}

// Remember that Unix time is in seconds since 1970.  This code
// wants to be handed time in seconds.

func hours(unix int64) int64 {
	return unix / 60 / 60
}

// Checks if the timestamp is valid.  If the timestamp is too old or
// too far into the future, then we don't consider it valid.  Or if we
// have seen this hash before, then it is not valid.  To that end,
// this code remembers hashes tested in the past, and rejects the
// second submission of the same hash.
func (r *Replay) IsTSValid(hash interfaces.IHash, timestamp int64) bool {
	return r.IsTSValid_(hash.Fixed(), timestamp, time.Now().Unix())
}

// To make the function testable, the logic accepts the current time
// as a parameter.  This way, the test code can manipulate the clock
// at will.
func (r *Replay) IsTSValid_(hash [32]byte, timestamp int64, now int64) bool {

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if len(r.buckets) < numBuckets {
		r.buckets = make([]map[[32]byte]int64, numBuckets, numBuckets)
	}

	now = hours(now)

	// If we have no buckets, or more than 24 hours has passed,
	// toss all the buckets. We do this by setting lasttime 24 hours
	// in the past.
	if now-r.lasttime > int64(numBuckets) {
		r.lasttime = now - int64(numBuckets)
	}

	// for every hour that has passed, toss one bucket by shifting
	// them all down a slot, and allocating a new bucket.
	for r.lasttime < now {
		r.buckets = append(r.buckets, make(map[[32]byte]int64))
		r.lasttime++
	}

	t := hours(timestamp)
	index := int(t - now + int64(numBuckets)/2)
	if index < 0 || index >= numBuckets {
		return false
	}

	if r.buckets[index] == nil {
		r.buckets[index] = make(map[[32]byte]int64)
	} else {
		_, ok := r.buckets[index][hash]
		if ok {
			return false
		}
	}
	r.buckets[index][hash] = t

	return true
}
