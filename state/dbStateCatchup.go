// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"container/list"
	"reflect"
	"sync"
	"time"

	"github.com/FactomProject/factomd/common/messages"
)

type GenericListItem interface {
	Height() uint32
}

func getHeightSafe(i GenericListItem) int {
	if i == nil || reflect.ValueOf(i).IsNil() {
		return -1
	}
	return int(i.Height())
}

func waitForLoaded(s *State) {
	// Don't start until the db is finished loading.
	for !s.DBFinished {
		time.Sleep(1 * time.Second)
	}
	if s.highestKnown < s.DBHeightAtBoot {
		s.highestKnown = s.DBHeightAtBoot + 1 // Make sure we ask for the next block after the database at startup.
	}
}

// TODO: Redesign Catchup. Some assumptions were made that made this more
// TODO: complex than it needed to be.
func (list *DBStateList) Catchup() {
	missing := list.State.StatesMissing
	waiting := list.State.StatesWaiting
	received := list.State.StatesReceived

	factomSecond := list.State.FactomSecond()

	requestTimeout := time.Duration(list.State.RequestTimeout) * factomSecond
	requestLimit := list.State.RequestLimit

	// Wait for db to be loaded
	waitForLoaded(list.State)

	// keep the lists up to date with the saved states.
	go func() {
		// Notify missing will add the height to the missing
		// if it is not received and not already requested.
		notifyMissing := func(n uint32) bool {
			if !waiting.Has(n) {
				list.State.LogPrintf("dbstatecatchup", "{actual} notify missing %d", n)
				missing.Add(n)
				return true
			}
			return false
		}

		var hs, hk uint32
		hsf := func() (rval uint32) {
			defer func() {
				if hs != rval {
					list.State.LogPrintf("dbstatecatchup", "HS = %d", rval)
				}
			}()
			// Sets the floor for what we will be requesting
			// AKA : What we have. In reality the receivedlist should
			// indicate that we have it, however, because a dbstate
			// is not fully validated before we get it, we cannot
			// assume that.
			floor := uint32(0)
			// Once it is in the db, we can assume it's all good.
			if d, err := list.State.DB.FetchDBlockHead(); err == nil && d != nil {
				floor = d.GetDatabaseHeight() // If it is in our db, let's make sure to stop asking
			}

			list.State.LogPrintf("dbstatecatchup", "Floor diff %d / %d", list.State.GetHighestSavedBlk(), floor)

			// get the hightest block in the database at boot
			b := list.State.GetDBHeightAtBoot()

			// don't request states that are in the database at boot time
			if b > floor {
				return b
			}
			return floor
		}

		// get the height of the known blocks
		hkf := func() (rval uint32) {
			a := list.State.GetHighestAck()
			k := list.State.GetHighestKnownBlock()
			defer func() {
				if hk != rval {
					list.State.LogPrintf("dbstatecatchup", "HK = %d", rval)
				}
			}()
			// check that known is more than 2 ahead of acknowledged to make
			// sure not to ask for blocks that haven't finished
			if k > a+2 {
				return k - 2
			}
			if a < 2 {
				return a
			}
			return a - 2 // Acks are for height + 1 (sometimes +2 in min 0)
		}
		hs = hsf()
		hk = hkf()
		list.State.LogPrintf("dbstatecatchup", "Start with hs = %d hk = %d", hs, hk)

		for {
			start := time.Now()
			// get the height of the saved blocks
			hs = hsf()
			hk = hkf()
			// The base means anything below we can toss
			base := received.Base()
			if base < hs {
				list.State.LogPrintf("dbstatecatchup", "Received base set to %d", hs)
				received.SetBase(hs)
				base = hs
			}

			receivedSlice := received.ListAsSlice()

			// When we pull the slice, we might be able to trim the receivedSlice for the next loop
			sliceKeep := 0
			// TODO: Rewrite to stop redundant looping over missing/waiting list
			// TODO: for each delete. It shouldn't be too bad atm, as most things are in order.
			for i, h := range receivedSlice {
				list.State.LogPrintf("dbstatecatchup", "missing & waiting delete %d", h)
				// remove any states from the missing list that have been saved.
				missing.LockAndDelete(h)
				// remove any states from the waiting list that have been saved.
				waiting.LockAndDelete(h)
				// Clean our our received list as well.
				if h <= base {
					sliceKeep = i
					received.LockAndDelete(h)
				}
			}

			// find gaps in the received list
			// we can start at `sliceKeep` because everything below it was removed
			for i := sliceKeep; i < len(receivedSlice)-1; i++ {
				h := receivedSlice[i]
				// if the height of the next received state is not equal to the
				// height of the current received state plus one then there is a
				// gap in the received state list.
				for n := h; n+1 < receivedSlice[i+1]; n++ {
					// missing.Notify <- NewMissingState(n + 1)
					r := notifyMissing(n + 1)
					list.State.LogPrintf("dbstatecatchup", "{gf} notify missing %d [%t]", n, r)
				}
			}

			// TODO: Better limit the number of asks based on what we already asked for.
			// TODO: If we implement that, ensure that we don't drop anything, as this covers any holes
			// TODO:	that might be made
			max := 3000 // Limit the number of new asks we will add for each iteration
			// add all known states after the last received to the missing list
			for n := received.Heighestreceived() + 1; n <= hk && max > 0; n++ {
				max--
				// missing.Notify <- NewMissingState(n)
				r := notifyMissing(n)
				list.State.LogPrintf("dbstatecatchup", "{hf (%d, %d)} notify missing %d [%t]", hk, max, n, r)
			}

			list.State.LogPrintf("dbstatecatchup", "height update took %s. SubBase:%d/%d/%d, Miss[v%d, ^_, T%d], Wait [v_, ^%d, T%d], Rec[v%d, ^%d, T%d]",
				time.Since(start),
				received.Base(), hs, list.State.GetDBHeightAtBoot(),
				getHeightSafe(missing.GetFront()), missing.Len(),
				getHeightSafe(waiting.GetEnd()), waiting.Len(),
				received.Base(), received.Heighestreceived(), received.List.Len())
			time.Sleep(factomSecond)
		}
	}()

	// watch the waiting list and move any requests that have timed out back
	// into the missing list.
	go func() {
		for {
			base := received.Base()
			waitingSlice := waiting.ListAsSlice()
			//for e := waiting.List.Front(); e != nil; e = e.Next() {
			for _, s := range waitingSlice {
				// Instead of choosing if to ask for it, just remove it
				if s.Height() <= base {
					waiting.LockAndDelete(s.Height())
					continue
				}
				if s.RequestAge() > requestTimeout {
					waiting.LockAndDelete(s.Height())
					if received.Get(s.Height()) == nil {
						list.State.LogPrintf("dbstatecatchup", "request timeout : waiting -> missing %d", s.Height())
						missing.Add(s.Height())
					}
				}
			}

			time.Sleep(requestTimeout)
		}
	}()

	// manage received dbstates
	go func() {
		for {
			select {
			case m := <-received.Notify:
				s := NewReceivedState(m)
				if s != nil {
					list.State.LogPrintf("dbstatecatchup", "dbstate received : missing & waiting delete, received add %d", s.Height())
					missing.LockAndDelete(s.Height())
					waiting.LockAndDelete(s.Height())
					received.Add(s.Height(), s.Message())
				}
			}
		}
	}()

	// request missing states from the network
	go func() {
		for {
			if waiting.Len() < requestLimit {
				// TODO: the batch limit should probably be set by a configuration variable
				b, e := missing.NextConsecutiveMissing(10)
				list.State.LogPrintf("dbstatecatchup", "dbstate requesting from %d to %d", b, e)

				if b == 0 && e == 0 {
					time.Sleep(1 * time.Second)
					continue
				}

				// make sure the end doesn't come before the beginning
				if e < b {
					e = b
				}

				msg := messages.NewDBStateMissing(list.State, b, e)
				msg.SendOut(list.State, msg)
				list.State.DBStateAskCnt += 1 // Total number of dbstates requests
				for i := b; i <= e; i++ {
					list.State.LogPrintf("dbstatecatchup", "dbstate requested : missing -> waiting %d", i)
					missing.LockAndDelete(i)
					waiting.Add(i)
				}
			} else {
				// if the next missing state is a lower height than the last waiting
				// state prune the waiting list
				m := missing.GetFront()
				w := waiting.GetEnd()
				if m != nil && w != nil {
					if m.Height() < w.Height() {
						list.State.LogPrintf("dbstatecatchup", "waiting delete, cleanup %d", w.Height())
						waiting.LockAndDelete(w.Height())
					}
				}

				time.Sleep(50 * time.Millisecond)
			}
		}
	}()
}

// MissingState is information about a DBState that is known to exist but is not
// available on the current node.
type MissingState struct {
	height uint32
}

// NewMissingState creates a new MissingState for the DBState at a specific
// height.
func NewMissingState(height uint32) *MissingState {
	s := new(MissingState)
	s.height = height
	return s
}

func (s *MissingState) Height() uint32 {
	return s.height
}

type StatesMissing struct {
	List *list.List
	// Notify chan *MissingState
	lock *sync.Mutex
}

// NewStatesMissing creates a new list of missing DBStates.
func NewStatesMissing() *StatesMissing {
	l := new(StatesMissing)
	l.List = list.New()
	// l.Notify = make(chan *MissingState)
	l.lock = new(sync.Mutex)
	return l
}

// Add adds a new MissingState to the list.
func (l *StatesMissing) Add(height uint32) {
	l.lock.Lock()
	defer l.lock.Unlock()

	for e := l.List.Back(); e != nil; e = e.Prev() {
		s := e.Value.(*MissingState)
		if height > s.Height() {
			l.List.InsertAfter(NewMissingState(height), e)
			return
		} else if height == s.Height() {
			return
		}
	}
	l.List.PushFront(NewMissingState(height))
}

// LockAndDelete removes a MissingState from the list.
func (l *StatesMissing) LockAndDelete(height uint32) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.DeleteLockless(height)
}

func (l *StatesMissing) DeleteLockless(height uint32) {
	// DeleteLockless does not lock the mutex, if called from another top level func
	if l == nil {
		return
	}

	for e := l.List.Front(); e != nil; e = e.Next() {
		if e.Value.(*MissingState).Height() == height {
			l.List.Remove(e)
			break
		}
	}
}

func (l *StatesMissing) Get(height uint32) *MissingState {
	// We want to lock here, as something can be deleted/added as we are iterating
	// and mess up our for loop
	l.lock.Lock()
	defer l.lock.Unlock()

	for e := l.List.Front(); e != nil; e = e.Next() {
		s := e.Value.(*MissingState)
		if s.Height() == height {
			return s
		}
	}
	return nil
}

func (l *StatesMissing) GetFront() *MissingState {
	// We want to lock here, as we first check the length, then grab the root.
	// the root could be deleted after we checked the len.
	l.lock.Lock()
	defer l.lock.Unlock()

	e := l.List.Front()
	if e != nil {
		s := e.Value.(*MissingState)
		if s != nil {
			return s
		}
	}
	return nil
}

func (l *StatesMissing) Len() int {
	return l.List.Len()
}

// NextConsecutiveMissing returns the heights of the the next n or fewer
// consecutive missing states
func (l *StatesMissing) NextConsecutiveMissing(n int) (uint32, uint32) {
	// We want to lock here, as something can be deleted/added as we are iterating
	// and mess up our for loop
	l.lock.Lock()
	defer l.lock.Unlock()

	f := l.List.Front()
	if f == nil {
		return 0, 0
	}
	beg := f.Value.(*MissingState).Height()
	end := beg
	c := 0
	for e := f.Next(); e != nil; e = e.Next() {
		h := e.Value.(*MissingState).Height()
		// We are looking to see if the consecutive height
		// sequence is broken. Easy to check if h != the next one
		// we are expecting.
		if h != end+1 {
			break
		}
		end++
		c++
		// TODO: the batch limit should probably be set as a configuration variable
		if c == n {
			break
		}
	}
	return beg, end
}

// GetNext pops the next MissingState from the list.
func (l *StatesMissing) GetNext() *MissingState {
	// We want to lock here, as we first check the length, then grab the root.
	// the root could be deleted after we checked the len.
	l.lock.Lock()
	defer l.lock.Unlock()

	e := l.List.Front()
	if e != nil {
		s := e.Value.(*MissingState)
		l.DeleteLockless(s.Height())
		return s
	}
	return nil
}

type WaitingState struct {
	height        uint32
	requestedTime time.Time
}

func NewWaitingState(height uint32) *WaitingState {
	s := new(WaitingState)
	s.height = height
	s.requestedTime = time.Now()
	return s
}

func (s *WaitingState) Height() uint32 {
	return s.height
}

func (s *WaitingState) RequestAge() time.Duration {
	return time.Since(s.requestedTime)
}

func (s *WaitingState) ResetRequestAge() {
	s.requestedTime = time.Now()
}

type StatesWaiting struct {
	List *list.List
	// Notify chan *WaitingState
	lock *sync.Mutex
}

func NewStatesWaiting() *StatesWaiting {
	l := new(StatesWaiting)
	l.List = list.New()
	// l.Notify = make(chan *WaitingState)
	l.lock = new(sync.Mutex)
	return l
}

func (l *StatesWaiting) ListAsSlice() []*WaitingState {
	// Lock as we are iterating
	l.lock.Lock()
	defer l.lock.Unlock()

	slice := make([]*WaitingState, l.List.Len())
	i := 0
	for e := l.List.Front(); e != nil; e = e.Next() {
		slice[i] = e.Value.(*WaitingState)
		i++
	}
	return slice

}

func (l *StatesWaiting) Add(height uint32) {
	l.lock.Lock()
	defer l.lock.Unlock()

	for e := l.List.Back(); e != nil; e = e.Prev() {
		s := e.Value.(*WaitingState)
		if s == nil {
			n := NewWaitingState(height)
			l.List.InsertAfter(n, e)
			return
		} else if height > s.Height() {
			n := NewWaitingState(height)
			l.List.InsertAfter(n, e)
			return
		} else if height == s.Height() {
			return
		}
	}
	l.List.PushFront(NewWaitingState(height))
}

func (l *StatesWaiting) LockAndDelete(height uint32) {
	l.lock.Lock()
	defer l.lock.Unlock()

	for e := l.List.Front(); e != nil; e = e.Next() {
		s := e.Value.(*WaitingState)
		if s.Height() == height {
			l.List.Remove(e)
			break
		}
	}
}

func (l *StatesWaiting) Get(height uint32) *WaitingState {
	// We want to lock here, as something can be deleted/added as we are iterating
	// and mess up our for loop
	l.lock.Lock()
	defer l.lock.Unlock()

	for e := l.List.Front(); e != nil; e = e.Next() {
		s := e.Value.(*WaitingState)
		if s.Height() == height {
			return s
		}
	}
	return nil
}

func (l *StatesWaiting) GetEnd() *WaitingState {
	// We want to lock here, as check the length then grab the root.
	// The root could be deleted after we checked for the length
	l.lock.Lock()
	defer l.lock.Unlock()

	e := l.List.Back()
	if e != nil {
		s := e.Value.(*WaitingState)
		if s != nil {
			return s
		}
	}
	return nil
}

func (l *StatesWaiting) Has(height uint32) bool {
	// We want to lock here, as something can be deleted/added as we are iterating
	// and mess up our for loop
	l.lock.Lock()
	defer l.lock.Unlock()

	for e := l.List.Front(); e != nil; e = e.Next() {
		s := e.Value.(*WaitingState)
		if s.Height() == height {
			return true
		}
	}
	return false
}

func (l *StatesWaiting) Len() int {
	return l.List.Len()
}

// ReceivedState represents a DBStateMsg received from the network
type ReceivedState struct {
	height uint32
	msg    *messages.DBStateMsg
}

// NewReceivedState creates a new member for the StatesReceived list
func NewReceivedState(msg *messages.DBStateMsg) *ReceivedState {
	if msg == nil {
		return nil
	}
	s := new(ReceivedState)
	s.height = msg.DirectoryBlock.GetHeader().GetDBHeight()
	s.msg = msg
	return s
}

// Height returns the block height of the received state
func (s *ReceivedState) Height() uint32 {
	return s.height
}

// Message returns the DBStateMsg received from the network.
func (s *ReceivedState) Message() *messages.DBStateMsg {
	return s.msg
}

// StatesReceived is the list of DBStates received from the network. "base"
// represents the height of known saved states.
type StatesReceived struct {
	List   *list.List
	Notify chan *messages.DBStateMsg
	base   uint32
	lock   *sync.Mutex
}

func NewStatesReceived() *StatesReceived {
	l := new(StatesReceived)
	l.List = list.New()
	l.Notify = make(chan *messages.DBStateMsg)
	l.lock = new(sync.Mutex)
	return l
}

// SubBase returns the base height of the StatesReceived list
func (l *StatesReceived) Base() uint32 {
	return l.base
}

func (l *StatesReceived) SetBase(height uint32) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.SetBaseLockless(height)
}

func (l *StatesReceived) SetBaseLockless(height uint32) {
	l.base = height

	for e := l.List.Front(); e != nil; e = e.Next() {
		switch v := e.Value.(*ReceivedState).Height(); {
		case v < l.base:
			l.List.Remove(e)
		case v == l.base:
			l.List.Remove(e)
			break
		case v > l.base:
			break
		}
	}
}

// Heighestreceived returns the height of the last member in StatesReceived
func (l *StatesReceived) Heighestreceived() uint32 {
	// We want to lock here, as we first check the length, then grab the root.
	// the root could be deleted after we checked the len.
	l.lock.Lock()
	defer l.lock.Unlock()

	height := uint32(0)
	s := l.List.Back()
	if s != nil {
		height = s.Value.(*ReceivedState).Height()
	}
	if l.Base() > height {
		return l.Base()
	}
	return height
}

// ListAsSlice will return the list as a slice
// to be iterated over in a threadsafe manner.
func (l *StatesReceived) ListAsSlice() []uint32 {
	// Lock as we are iterating
	l.lock.Lock()
	defer l.lock.Unlock()

	slice := make([]uint32, l.List.Len())
	i := 0
	for e := l.List.Front(); e != nil; e = e.Next() {
		slice[i] = e.Value.(*ReceivedState).Height()
		i++
	}
	return slice
}

// Add adds a new received state to the list.
func (l *StatesReceived) Add(height uint32, msg *messages.DBStateMsg) {
	if msg == nil {
		return
	}

	l.lock.Lock()
	defer l.lock.Unlock()

	if height < l.base {
		// We already know we had this height
		// This should really never happen
		return
	}

	for e := l.List.Back(); e != nil; e = e.Prev() {
		s := e.Value.(*ReceivedState)
		if s == nil {
			n := NewReceivedState(msg)
			l.List.InsertAfter(n, e)
			return
		} else if height > s.Height() {
			n := NewReceivedState(msg)
			l.List.InsertAfter(n, e)
			return
		} else if height == s.Height() {
			return
		}
	}
	l.List.PushFront(NewReceivedState(msg))
}

// LockAndDelete removes a state from the StatesReceived list
func (l *StatesReceived) LockAndDelete(height uint32) {
	l.lock.Lock()
	defer l.lock.Unlock()

	for e := l.List.Back(); e != nil; e = e.Prev() {
		s := e.Value.(*ReceivedState)
		if s == nil {
			break
		} else if s.Height() == height {
			l.List.Remove(e)
			break
		}
	}
}

// Get returns a member from the StatesReceived list
func (l *StatesReceived) Get(height uint32) *ReceivedState {
	// We want to lock here, as something can be deleted/added as we are iterating
	// and mess up our for loop
	l.lock.Lock()
	defer l.lock.Unlock()

	for e := l.List.Back(); e != nil; e = e.Prev() {
		s := e.Value.(*ReceivedState)
		if height > s.Height() {

		}
		if s.Height() == height {
			return s
		}
	}

	return nil
}

func (l *StatesReceived) Has(height uint32) bool {
	// We want to lock here, as something can be deleted/added as we are iterating
	// and mess up our for loop
	l.lock.Lock()
	defer l.lock.Unlock()

	if height <= l.Base() {
		return true
	}

	for e := l.List.Front(); e != nil; e = e.Next() {
		s := e.Value.(*ReceivedState)
		if s == nil {
			return false
		}
		if s.Height() == height {
			return true
		}
	}
	return false
}

func (l *StatesReceived) GetNext() *ReceivedState {
	l.lock.Lock()
	defer l.lock.Unlock()

	if l.List.Len() == 0 {
		return nil
	}
	e := l.List.Front()
	if e != nil {
		s := e.Value.(*ReceivedState)

		if s == nil {
			l.List.Remove(e)
			return nil
		}

		if s.Height() == l.Base()+1 {
			l.SetBaseLockless(s.Height())
			l.List.Remove(e)
			return s
		}

		if s.Height() <= l.Base() {
			l.List.Remove(e)
		}
	}
	return nil
}
