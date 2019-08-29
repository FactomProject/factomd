// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package state

import (
	"fmt"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"
)

const Range = 60                // Double this for the period we protect, i.e. 120 means +/- 120 minutes
const numBuckets = Range*2 + 60 // Cover the range in the future and in the past, with an hour buffer.

var _ = time.Now()
var _ = fmt.Print

type Replay struct {
	Mutex    sync.Mutex
	Buckets  [numBuckets]map[[32]byte]int
	Basetime int // hours since 1970
	Center   int // Hour of the current time.
	// debug
	s    *State
	name string
}

var _ interfaces.BinaryMarshallable = (*Replay)(nil)

func RandomReplay() *Replay {
	r := new(Replay)

	for i := 0; i < numBuckets; i++ {
		l2 := random.RandIntBetween(0, 50)
		m := map[[32]byte]int{}
		for j := 0; j < l2; j++ {
			h := primitives.RandomHash()
			m[h.Fixed()] = random.RandInt()
		}
		r.Buckets[i] = m
	}

	r.Basetime = random.RandInt()
	r.Center = random.RandInt()

	return r
}

func (r *Replay) Init() {
	for i := range r.Buckets {
		if r.Buckets[i] == nil {
			r.Buckets[i] = map[[32]byte]int{}
		}
	}
}

func (a *Replay) IsSameAs(b *Replay) bool {
	if a == nil || b == nil {
		if a == nil && b == nil {
			return true
		}
		return false
	}
	a.Init()
	b.Init()

	if len(a.Buckets) != len(b.Buckets) {
		return false
	}
	for i := range a.Buckets {
		if len(a.Buckets[i]) != len(b.Buckets[i]) {
			return false
		}
		for k := range a.Buckets[i] {
			if a.Buckets[i][k] != b.Buckets[i][k] {
				return false
			}
		}
	}

	if a.Basetime != b.Basetime {
		return false
	}
	if a.Center != b.Center {
		return false
	}
	return true
}

func (r *Replay) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "Replay.MarshalBinary err:%v", *pe)
		}
	}(&err)
	r.Init()
	b := primitives.NewBuffer(nil)

	for _, v := range r.Buckets {
		err := PushBucketMap(b, v)
		if err != nil {
			return nil, err
		}
	}

	err = b.PushInt(r.Basetime)
	if err != nil {
		return nil, err
	}
	err = b.PushInt(r.Center)
	if err != nil {
		return nil, err
	}

	return b.DeepCopyBytes(), nil
}

func (r *Replay) UnmarshalBinaryData(p []byte) ([]byte, error) {
	r.Init()
	b := primitives.NewBuffer(p)

	for i := 0; i < numBuckets; i++ {
		m, err := PopBucketMap(b)
		if err != nil {
			return nil, err
		}
		r.Buckets[i] = m
	}
	var err error
	r.Basetime, err = b.PopInt()
	if err != nil {
		return nil, err
	}
	r.Center, err = b.PopInt()
	if err != nil {
		return nil, err
	}

	return b.DeepCopyBytes(), nil
}

func (r *Replay) UnmarshalBinary(p []byte) error {
	_, err := r.UnmarshalBinaryData(p)
	return err
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

func maskToString(mask int) (rval string) {
	if mask&constants.INTERNAL_REPLAY != 0 {
		rval += "|INTERNAL_REPLAY"
	}

	if mask&constants.NETWORK_REPLAY != 0 {
		rval += "|NETWORK_REPLAY"
	}
	if mask&constants.TIME_TEST != 0 {
		rval += "|TIME_TEST"
	}
	if mask&constants.REVEAL_REPLAY != 0 {
		rval += "|REVEAL_REPLAY"
	}
	if mask&constants.BLOCK_REPLAY != 0 {
		rval += "|BLOCK_REPLAY"
	}
	if mask > 31 {
		panic("Unknown replay mask")
	}
	if mask == 0 {
		rval = "|Empty" // extra character to be removed
	}
	return rval[1:]
}

// This move the center time of the replay filter to the current systemtime.
// Really it should track the time for the current block.
func (r *Replay) Recenter(systemtime interfaces.Timestamp) {
	now := Minutes(systemtime.GetTimeSeconds())

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

}

// Returns false if the hash is too old, or is already a
// member of the set.  Timestamp is in seconds.
// Does not add the hash to the buckets!
// Assume locking is done by caller
func (r *Replay) validate(mask int, hash [32]byte, timestamp interfaces.Timestamp, systemtime interfaces.Timestamp) (index int, valid bool) {
	now := Minutes(systemtime.GetTimeSeconds())
	t := Minutes(timestamp.GetTimeSeconds())

	r.Recenter(systemtime) // Move the center of our replay list to the current time.

	diff := now - t
	// Check the timestamp to see if within 12 hours of the system time.  That not valid, we are
	// just done without any added concerns.
	if diff > Range || diff < -Range {
		//fmt.Println("Time in hours, range:", hours(timeSeconds-systemTimeSeconds), HourRange)
		return -1, false
	}

	if mask == constants.TIME_TEST {
		return -1, true
	}
	// Just take the time of the thing in hours less the basetime to get the index.
	index = t - r.Basetime

	if index < 0 || index >= numBuckets {
		return -1, false
	}

	if r.Buckets[index] == nil {
		r.Buckets[index] = make(map[[32]byte]int)
	} else {
		v, ok := r.Buckets[index][hash]
		_ = ok
		if v&mask > 0 {
			return index, false
		}
	}
	return index, true
}

// Returns false if the hash is too old, or is already a
// member of the set.  Timestamp is in seconds.
// Does not add the hash to the buckets!
func (r *Replay) Valid(mask int, hash [32]byte, timestamp interfaces.Timestamp, systemtime interfaces.Timestamp) (index int, valid bool) {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	return r.validate(mask, hash, timestamp, systemtime)
}

// Checks if the timestamp is valid.  If the timestamp is too old or
// too far into the future, then we don't consider it valid.  Or if we
// have seen this hash before, then it is not valid.  To that end,
// this code remembers hashes tested in the past, and rejects the
// second submission of the same hash.
func (r *Replay) IsTSValid(mask int, hash interfaces.IHash, timestamp interfaces.Timestamp) bool {
	return r.IsTSValidAndUpdateState(mask, hash.Fixed(), timestamp, primitives.NewTimestampNow())
}

// To make the function testable, the logic accepts the current time
// as a parameter.  This way, the test code can manipulate the clock
// at will.
func (r *Replay) IsTSValidAndUpdateState(mask int, hash [32]byte, timestamp interfaces.Timestamp, systemtime interfaces.Timestamp) (rval bool) {
	r.Mutex.Lock()
	// Fixed to avoid a race with unlocking after finding the index and then using it in a later when it could be stale

	index, rval := r.validate(mask, hash, timestamp, systemtime)
	if rval {
		// Mark this hash as seen
		if mask != constants.TIME_TEST {
			r.Buckets[index][hash] = r.Buckets[index][hash] | mask
			r.Mutex.Unlock()
			//r.s.LogPrintf("replay", "Add %x (%s) to %s from %s", hash[:3], maskToString(mask), r.Name, atomic.WhereAmIString(1))
			return rval // true
		}
		r.Mutex.Unlock()
		return rval // true
	}

	r.Mutex.Unlock()
	return rval // false
}

// Returns True if there is no record of this hash in the Replay structures.
// Returns false if we have seen this hash before.
// Assume locking is done by caller
func (r *Replay) IsHashUnique_internal(mask int, hash [32]byte) bool {

	for _, bucket := range r.Buckets {
		if bucket[hash]&mask > 0 {
			return false
		}
	}
	return true
}

// Returns True if there is no record of this hash in the Replay structures.
// Returns false if we have seen this hash before.
func (r *Replay) IsHashUnique(mask int, hash [32]byte) bool {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	return r.IsHashUnique_internal(mask, hash)
}

// Fixed conflicting lock path
// Add a hash/mask at a specific time if it did not exist prior.
func (r *Replay) SetHashNow(mask int, hash [32]byte, now interfaces.Timestamp) bool {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	if r.IsHashUnique_internal(mask, hash) {
		index := Minutes(now.GetTimeSeconds()) - r.Basetime
		if index < 0 || index >= len(r.Buckets) {
			return false
		}

		if r.Buckets[index] == nil {
			r.Buckets[index] = make(map[[32]byte]int)
		}
		r.Buckets[index][hash] = mask | r.Buckets[index][hash]
		//r.s.LogPrintf("replay", "Add2 %x (%s) to %s from %s", hash[:3], maskToString(mask), r.Name, atomic.WhereAmIString(1))
		return true
	}
	return false
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

func PushBucketMap(b *primitives.Buffer, m map[[32]byte]int) error {
	l := len(m)
	err := b.PushVarInt(uint64(l))
	if err != nil {
		return err
	}

	keys := [][32]byte{}
	for k := range m {
		keys = append(keys, k)
	}
	sort.Sort(ByKey(keys))

	for _, k := range keys {
		err = b.Push(k[:])
		if err != nil {
			return err
		}
		err = b.PushInt(m[k])
		if err != nil {
			return err
		}
	}
	return nil
}

func PopBucketMap(buf *primitives.Buffer) (map[[32]byte]int, error) {
	m := map[[32]byte]int{}
	k := make([]byte, 32)
	l, err := buf.PopVarInt()
	if err != nil {
		return nil, err
	}
	for i := 0; i < int(l); i++ {
		var b [32]byte
		err = buf.Pop(k)
		if err != nil {
			return nil, err
		}
		copy(b[:], k)
		v, err := buf.PopInt()
		if err != nil {
			return nil, err
		}
		m[b] = v
	}
	return m, nil
}
