// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"container/list"
	"time"

	"github.com/FactomProject/factomd/common/messages"
)

func (list *DBStateList) Catchup() {
	missing := list.State.StatesMissing
	waiting := list.State.StatesWaiting
	recieved := list.State.StatesReceived

	requestTimeout := list.State.RequestTimeout
	requestLimit := list.State.RequestLimit

	// keep the lists up to date with the saved states.
	go func() {
		for {
			// get the height of the saved blocks
			hs := func() uint32 {
				l := list.State.GetLLeaderHeight()
				if list.State.GetCurrentMinute() == 0 {
					if l < 2 {
						return 0
					}
					return l - 2
				}
				return l - 1
			}()

			// get the hight of the known blocks
			hk := func() uint32 {
				a := list.State.GetHighestAck()
				k := list.State.GetHighestKnownBlock()
				if a > k {
					return a
				}
				return k
			}()

			if recieved.Base() < hs {
				recieved.SetBase(hs)
			}

			// TODO: removing missing and waiting states could be done in parallel.
			// remove any states from the missing list that have been saved.
			for e := missing.List.Front(); e != nil; e = e.Next() {
				s := e.Value.(*MissingState)
				if recieved.Has(s.Height()) {
					missing.Del(s.Height())
				}
			}

			// remove any states from the waiting list that have been saved.
			for e := waiting.List.Front(); e != nil; e = e.Next() {
				s := e.Value.(*WaitingState)
				if recieved.Has(s.Height()) {
					waiting.Del(s.Height())
				}
			}

			// find gaps in the recieved list
			for e := recieved.List.Front(); e != nil; e = e.Next() {
				// if the height of the next recieved state is not equal to the
				// height of the current recieved state plus one then there is a
				// gap in the recieved state list.
				if e.Next() != nil {
					for n := e.Value.(*ReceivedState).Height(); n+1 < e.Next().Value.(*ReceivedState).Height(); n++ {
						missing.Notify <- NewMissingState(n + 1)
					}
				}
			}

			// add all known states after the last recieved to the missing list
			for n := recieved.HeighestRecieved() + 1; n < hk; n++ {
				missing.Notify <- NewMissingState(n)
			}

			time.Sleep(5 * time.Second)
		}
	}()

	// watch the waiting list and move any requests that have timed out back
	// into the missing list.
	go func() {
		for {
			for e := waiting.List.Front(); e != nil; e = e.Next() {
				s := e.Value.(*WaitingState)
				if s.RequestAge() > requestTimeout {
					waiting.Del(s.Height())
					missing.Notify <- NewMissingState(s.Height())
				}
			}

			time.Sleep(1 * time.Second)
		}
	}()

	// manage the state lists
	go func() {
		for {
			select {
			case s := <-missing.Notify:
				if recieved.Get(s.Height()) == nil {
					if !waiting.Has(s.Height()) {
						missing.Add(s.Height())
					}
				}
			case s := <-waiting.Notify:
				if !waiting.Has(s.Height()) {
					waiting.Add(s.Height())
				}
			case m := <-recieved.Notify:
				s := NewReceivedState(m)
				if s != nil {
					missing.Del(s.Height())
					waiting.Del(s.Height())
					recieved.Add(s.Height(), s.Message())
				}
			}
		}
	}()

	// request missing states from the network
	go func() {
		for {
			// TODO: replace waiting.Len with some kind of mechanism that tracks
			// the number of batch requests instead of the number of waiting states
			// e.g. there could be up to the batch limit number of waiting states
			// represented by a single request
			if waiting.Len() < requestLimit {
				// TODO: the batch limit should probably be set by a configuration variable
				b, e := missing.NextConsecutiveMissing(10)

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
				for i := b; i <= e; i++ {
					missing.Del(i)
					waiting.Notify <- NewWaitingState(i)
				}
			} else {
				// if the next missing state is a lower height than the last waiting
				// state prune the waiting list
				m := missing.GetFront()
				w := waiting.GetEnd()
				if m != nil && w != nil {
					if m.Height() < w.Height() {
						waiting.Del(w.Height())
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

// TODO: if StatesMissing takes a long time to seek through the list we should
// replace the iteration with binary search

type StatesMissing struct {
	List   *list.List
	Notify chan *MissingState
}

// NewStatesMissing creates a new list of missing DBStates.
func NewStatesMissing() *StatesMissing {
	l := new(StatesMissing)
	l.List = list.New()
	l.Notify = make(chan *MissingState)
	return l
}

// Add adds a new MissingState to the list.
func (l *StatesMissing) Add(height uint32) {
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

// Del removes a MissingState from the list.
func (l *StatesMissing) Del(height uint32) {
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
	for e := l.List.Front(); e != nil; e = e.Next() {
		s := e.Value.(*MissingState)
		if s.Height() == height {
			return s
		}
	}
	return nil
}

func (l *StatesMissing) GetFront() *MissingState {
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
	f := l.List.Front()
	if f == nil {
		return 0, 0
	}
	beg := f.Value.(*MissingState).Height()
	end := beg
	c := 0
	for e := l.List.Front(); e != nil; e = e.Next() {
		h := e.Value.(*MissingState).Height()
		if h > end+1 {
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
	e := l.List.Front()
	if e != nil {
		s := e.Value.(*MissingState)
		l.Del(s.Height())
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
	List   *list.List
	Notify chan *WaitingState
}

func NewStatesWaiting() *StatesWaiting {
	l := new(StatesWaiting)
	l.List = list.New()
	l.Notify = make(chan *WaitingState)
	return l
}

func (l *StatesWaiting) Add(height uint32) {
	// l.List.PushBack(NewWaitingState(height))
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

func (l *StatesWaiting) Del(height uint32) {
	for e := l.List.Front(); e != nil; e = e.Next() {
		s := e.Value.(*WaitingState)
		if s.Height() == height {
			l.List.Remove(e)
		}
	}
}

func (l *StatesWaiting) Get(height uint32) *WaitingState {
	for e := l.List.Front(); e != nil; e = e.Next() {
		s := e.Value.(*WaitingState)
		if s.Height() == height {
			return s
		}
	}
	return nil
}

func (l *StatesWaiting) GetEnd() *WaitingState {
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

// StatesReceived is the list of DBStates recieved from the network. "base"
// represents the height of known saved states.
type StatesReceived struct {
	List   *list.List
	Notify chan *messages.DBStateMsg
	base   uint32
}

func NewStatesReceived() *StatesReceived {
	l := new(StatesReceived)
	l.List = list.New()
	l.Notify = make(chan *messages.DBStateMsg)
	return l
}

// Base returns the base height of the StatesReceived list
func (l *StatesReceived) Base() uint32 {
	return l.base
}

func (l *StatesReceived) SetBase(height uint32) {
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

// HeighestRecieved returns the height of the last member in StatesReceived
func (l *StatesReceived) HeighestRecieved() uint32 {
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

// Add adds a new recieved state to the list.
func (l *StatesReceived) Add(height uint32, msg *messages.DBStateMsg) {
	if msg == nil {
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

// Del removes a state from the StatesReceived list
func (l *StatesReceived) Del(height uint32) {
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
	for e := l.List.Back(); e != nil; e = e.Prev() {
		s := e.Value.(*ReceivedState)
		if s == nil {
			return nil
		}

		if s.Height() == height {
			return s
		}
	}

	return nil
}

func (l *StatesReceived) Has(height uint32) bool {
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
			l.SetBase(s.Height())
			l.List.Remove(e)
			return s
		}

		if s.Height() <= l.Base() {
			l.List.Remove(e)
		}
	}
	return nil
}
