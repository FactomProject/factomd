package state

import (
	"fmt"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/messages"

	"github.com/FactomProject/factomd/common/interfaces"
)

// toggle to disable old/new for testing
const useNewHolding = true

// This hold a slice of messages dependent on a hash
type HoldingList struct {
	holding    map[[32]byte][]interfaces.IMsg
	s          *State // for debug logging
	dependents map[[32]byte]bool
}

func (l *HoldingList) Init(s *State) {
	l.holding = make(map[[32]byte][]interfaces.IMsg)
	l.s = s
	l.dependents = make(map[[32]byte]bool)

	if ! useNewHolding {
		l.s.LogPrintf("newHolding", "DISABLED")
	}
}

func (l *HoldingList) Messages() map[[32]byte][]interfaces.IMsg {
	return l.holding
}

func (l *HoldingList) GetSize() int {
	return len(l.dependents)
}

func (l *HoldingList) Exists(h [32]byte) bool {
	return l.dependents[h]
}

// Add a message to a dependent holding list
func (l *HoldingList) Add(h [32]byte, msg interfaces.IMsg) bool {

	if l.dependents[msg.GetMsgHash().Fixed()] {
		return false
	}

	if l.holding[h] == nil {
		l.holding[h] = []interfaces.IMsg{msg}
	} else {
		l.holding[h] = append(l.holding[h], msg)
	}

	l.dependents[msg.GetMsgHash().Fixed()] = true
	return true
}

// get and remove the list of dependent message for a hash
func (l *HoldingList) Get(h [32]byte) []interfaces.IMsg {
	rval := l.holding[h]
	delete(l.holding, h)

	for _, msg := range rval {
		l.s.LogMessage("newHolding", "DequeueFromDependantHolding()", msg)
		delete(l.dependents, msg.GetMsgHash().Fixed())
	}
	return rval
}

func (l *HoldingList) ExecuteForNewHeight(ht uint32) {
	l.s.ExecuteFromHolding(HeightToHash(ht))
}

// clean stale messages from holding
func (l *HoldingList) Review() {

	if ! useNewHolding && l.GetSize() > 0 {
		panic("found messages in new-holding while disabled")
	}

	for h := range l.holding {
		dh := l.holding[h]
		if nil == l {
			continue
		}
		for _, msg := range dh {
			if l.isMsgStale(msg) {
				l.Get(h) // remove all from holding
				l.s.LogMessage("newHolding", "RemoveFromDependantHolding()", msg)
				continue
			}
		}
	}
}

func (l *HoldingList) isMsgStale(msg interfaces.IMsg) (res bool) {

	/*
		REVIEW:
		Maybe we should treat the message stream as a votes on the "highest known block" where known servers trump unknown servers who disagree?

		Consider setting HKB and HAB when we complete minute 1 of a block to the current leader height.
		That at least would make us recover from a spoofed ack attack.
	*/

	switch msg.Type() {
	case constants.EOM_MSG, constants.ACK_MSG, constants.DIRECTORY_BLOCK_SIGNATURE_MSG:
		if msg.(*messages.EOM).DBHeight < l.s.GetHighestKnownBlock()-1 {
			res = true
		}
	default:
		l.s.LogMessage("newHolding", "SKIP_DBHT_REVIEW", msg)
	}

	if msg.GetTimestamp().GetTime().UnixNano() < l.s.GetFilterTimeNano() {
		res = true
	}

	if res {
		l.s.LogMessage("newHolding", "EXPIRE", msg)
	} else {
		l.s.LogMessage("newHolding", "NOT_EXPIRED", msg)
	}

	return false
}

func (s *State) HoldForHeight(ht uint32, msg interfaces.IMsg) int {
	if s.GetLLeaderHeight()+1 == ht && s.GetCurrentMinute() >= 9 {
		s.LogMessage("newHolding", fmt.Sprintf("SKIP_HoldForHeight %x", ht), msg)
		return 0 // send to old holding
	}
	s.LogMessage("newHolding", fmt.Sprintf("HoldForHeight %x", ht), msg)
	return s.Add(HeightToHash(ht), msg) // add to new holding
}

// Add a message to a dependent holding list
func (s *State) Add(h [32]byte, msg interfaces.IMsg) int {

	if ! useNewHolding {
		return 0
	}

	if msg == nil {
		panic("Empty Message Added to Holding")
	}

	if h == [32]byte{} {
		panic("Empty Hash Passed to New Holding")
	}

	if s.Hold.Add(h, msg) {
		// return negative value so message is marked invalid and not processed in standard holding
		s.LogMessage("newHolding", fmt.Sprintf("Add %x", h[:6]), msg)
	}

	// mark as invalid for validator loop
	return -2 // ensures message is not sent to hold holding
}

// get and remove the list of dependent message for a hash
func (s *State) Get(h [32]byte) []interfaces.IMsg {
	return s.Hold.Get(h)
}

// Execute a list of messages from holding that are dependent on a hash
// the hash may be a EC address or a CainID or a height (ok heights are not really hashes but we cheat on that)
func (s *State) ExecuteFromHolding(h [32]byte) {

	if ! useNewHolding && s.Hold.GetSize() > 0 {
		panic("found messages in new-holding while disabled")
	}
	// get the list of messages waiting on this hash
	l := s.Get(h)
	if l == nil {
		s.LogPrintf("newHolding", "ExecuteFromDependantHolding(%x) nothing waiting", h[:6])
		return
	}
	s.LogPrintf("newHolding", "ExecuteFromDependantHolding(%d)[%x]", len(l), h[:6])

	for _, m := range l {
		s.LogPrintf("newHolding", "Delete M-%x", m.GetMsgHash().Bytes()[:3])
	}

	go func() {
		// add the messages to the msgQueue so they get executed as space is available
		for _, m := range l {
			s.LogMessage("msgQueue", "enqueue_from_dependent_holding", m)
			s.msgQueue <- m
		}
	}()
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
