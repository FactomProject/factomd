package state

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
)

// This hold a slice of messages dependent on a hash
type HoldingList struct {
	holding map[[32]byte][]interfaces.IMsg
	s       *State // for debug logging
	size    int
}

func (l *HoldingList) Init(s *State) {
	l.holding = make(map[[32]byte][]interfaces.IMsg)
	l.s = s
	l.size = 0
}

func (l *HoldingList) GetSize() int {
	return l.size
}

// Add a messsage to a dependent holding list
func (l *HoldingList) Add(h [32]byte, msg interfaces.IMsg) {

	if l.holding[h] == nil {
		l.holding[h] = []interfaces.IMsg{msg}
	} else {
		l.holding[h] = append(l.holding[h], msg)
	}
	l.size++
}

// get and remove the list of dependent message for a hash
func (l *HoldingList) Get(h [32]byte) []interfaces.IMsg {
	rval := l.holding[h]
	l.size -= len(rval)
	l.holding[h] = nil
	return rval
}

// clean stale messages from holding
func (hl *HoldingList) Review() {
	for h := range hl.holding {
		l := hl.holding[h]
		if nil == l {
			continue
		}
		for _, msg := range l {
			if hl.s.IsMsgStale(msg) < 0 {
				hl.Get(h) // remove from holding
				hl.s.LogMessage("newHolding", "RemoveFromDependantHolding()", msg)
				continue
			}
		}
	}
}

// Add a message to a dependent holding list
func (s *State) Add(h [32]byte, msg interfaces.IMsg) {
	s.LogMessage("newHolding", fmt.Sprintf("AddToDependantHolding(%x)", h[:4]), msg)
	s.Hold.Add(h, msg)
}

// get and remove the list of dependent message for a hash
func (s *State) Get(h [32]byte) []interfaces.IMsg {
	return s.Hold.Get(h)
}

// Execute a list of messages from holding that are dependant on a hash
// the hash may be a EC address or a CainID or a height (ok heights are not really hashes but we cheat on that)
func (s *State) ExecuteFromHolding(h [32]byte) {
	// get the list of messages waiting on this hash
	l := s.Get(h)
	if l != nil {
		s.LogPrintf("newHolding", "ExecuteFromDependantHolding(%x)[%d]", len(l), h[:4])
		// add the messages to the msgQueue so they get executed as space is available
		func() {
			for _, m := range l {
				s.LogMessage("msgQueue", "enqueue_from_dependant_holding", m)
				s.msgQueue <- m
			}
		}()
	} else {
		s.LogPrintf("newHolding", "ExecuteFromDependantHolding(%x) nothing waiting", h[:4])
	}
}

// put a height in the first 4 bytes of a hash so we can use it to look up dependent message in holding
func HeightToHash(height uint32) [32]byte {
	var h [32]byte
	h[0] = byte((height >> 24) & 0xFF)
	h[1] = byte((height >> 16) & 0xFF)
	h[2] = byte((height >> 8) & 0xFF)
	h[3] = byte((height >> 0) & 0xFF)
	return h
}
