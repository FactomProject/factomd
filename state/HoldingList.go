package state

import (
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/telemetry"
)

type heldMessage struct {
	dependentHash [32]byte
	offset        int
}

// This hold a slice of messages dependent on a hash
type HoldingList struct {
	holding    map[[32]byte][]interfaces.IMsg
	s          *State                   // for debug logging
	dependents map[[32]byte]heldMessage // used to avoid duplicate entries & track position in holding
}

// access gauge w/ proper labels
func (l *HoldingList) metric(msg interfaces.IMsg) telemetry.Gauge {
	return telemetry.MapSize.WithLabelValues("state", "DependentHolding", msg.Label())
}

func (l *HoldingList) Init(s *State) {
	l.holding = make(map[[32]byte][]interfaces.IMsg)
	l.s = s
	l.dependents = make(map[[32]byte]heldMessage)
}

func (l *HoldingList) Messages() map[[32]byte][]interfaces.IMsg {
	return l.holding
}

func (l *HoldingList) GetSize() int {
	return len(l.dependents)
}

// Get a single msg from dependent holding
func (l *HoldingList) GetDependentMsg(h [32]byte) interfaces.IMsg {
	d, ok := l.dependents[h]
	if !ok {
		return nil
	}
	m := l.holding[d.dependentHash][d.offset]
	return m
}

// remove a single msg from  dependent holding (done when we add it to the process list).
func (l *HoldingList) RemoveDependentMsg(h [32]byte, reason string) {
	d, ok := l.dependents[h]
	if !ok {
		return
	}
	msg := l.holding[d.dependentHash][d.offset]
	l.holding[d.dependentHash][d.offset] = nil
	delete(l.dependents, h)
	l.s.LogMessage("DependentHolding", "delete "+reason, msg)
	return
}

// Add a message to a dependent holding list
func (l *HoldingList) Add(h [32]byte, msg interfaces.IMsg) bool {
	_, found := l.dependents[msg.GetMsgHash().Fixed()]
	if found {
		return false
	}
	l.metric(msg).Inc()
	l.s.LogMessage("DependentHolding", fmt.Sprintf("add[%x]", h[:6]), msg)

	if l.holding[h] == nil {
		l.holding[h] = []interfaces.IMsg{msg}
	} else {
		l.holding[h] = append(l.holding[h], msg)
	}

	l.dependents[msg.GetMsgHash().Fixed()] = heldMessage{h, len(l.holding[h]) - 1}
	return true
}

// get and remove the list of dependent message for a hash
func (l *HoldingList) Get(h [32]byte) []interfaces.IMsg {
	rval := l.holding[h]
	delete(l.holding, h)

	// delete all the individual messages from the list
	for _, msg := range rval {
		if msg == nil {
			continue
		} else {
			l.s.LogMessage("DependentHolding", fmt.Sprintf("delete[%x]", h[:6]), msg)
			l.metric(msg).Dec()
			delete(l.dependents, msg.GetMsgHash().Fixed())
		}
	}
	return rval
}

func (l *HoldingList) ExecuteForNewHeight(ht uint32, minute int) {
	l.s.ExecuteFromHolding(HeightToHash(ht, minute))
}

// clean stale messages from holding
func (l *HoldingList) Review() {

	for h := range l.holding {
		dh := l.holding[h]
		if nil == dh {
			continue
		}

		inUse := false
		for i, msg := range dh {

			if msg != nil && l.isMsgStale(msg) {
				l.holding[h][i] = nil // nil out the held message
				delete(l.dependents, msg.GetMsgHash().Fixed())
				continue
			}

			if msg != nil {
				inUse = true
			}
		}

		if !inUse {
			delete(l.holding, h)
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
	case constants.EOM_MSG:
		if uint32(msg.(*messages.EOM).DBHeight)*10+uint32(msg.(*messages.EOM).Minute) < l.s.GetLLeaderHeight()*10+uint32(l.s.CurrentMinute) {
			res = true
		}
	case constants.ACK_MSG:
		if msg.(*messages.Ack).DBHeight < l.s.GetLLeaderHeight() {
			res = true
		}
	case constants.DIRECTORY_BLOCK_SIGNATURE_MSG:
		if msg.(*messages.DirectoryBlockSignature).DBHeight < l.s.GetLLeaderHeight() {
			res = true
		}
	default:
		//		l.s.LogMessage("DependentHolding", "SKIP_DBHT_REVIEW", msg)
	}

	if msg.GetTimestamp().GetTime().UnixNano() < l.s.GetFilterTimeNano() {
		res = true
	}

	if res {
		l.s.LogMessage("DependentHolding", "EXPIRE", msg)
	} else {
		//		l.s.LogMessage("DependentHolding", "NOT_EXPIRED", msg)
	}

	return res
}

func (s *State) HoldForHeight(ht uint32, minute int, msg interfaces.IMsg) int {
	//	s.LogMessage("DependentHolding", fmt.Sprintf("HoldForHeight %x", ht), msg)
	return s.Add(HeightToHash(ht, minute), msg) // add to new holding
}

// Add a message to a dependent holding list
func (s *State) Add(h [32]byte, msg interfaces.IMsg) int {

	if msg == nil { // REVIEW: consider removing paranoid check
		panic("Empty Message Added to Holding")
	}

	if h == [32]byte{} { // REVIEW: consider removing paranoid check
		panic("Empty Hash Passed to New Holding")
	}

	s.Hold.Add(h, msg)

	// mark as invalid for validator loop
	return -2 // ensures message is not sent to old holding
}

// Execute a list of messages from holding that are dependent on a hash
// the hash may be a EC address or a ChainID or a height (ok heights are not really hashes but we cheat on that)
func (s *State) ExecuteFromHolding(h [32]byte) {

	// get and delete the list of messages waiting on this hash
	l := s.Hold.Get(h)
	if l == nil {
		//		s.LogPrintf("DependentHolding", "ExecuteFromDependentHolding(%x) nothing waiting", h[:6])
		return
	}
	s.LogPrintf("DependentHolding", "ExecuteFromDependentHolding(%d)[%x]", len(l), h[:6])

	for _, m := range l {
		if m == nil {
			continue
		}
		s.LogPrintf("DependentHolding", "delete M-%x", m.GetMsgHash().Bytes()[:3])
	}

	go func() {
		// add the messages to the msgQueue so they get executed as space is available
		for _, m := range l {
			if m == nil {
				continue
			}
			s.LogMessage("msgQueue", "enqueue_from_dependent_holding", m)
			s.msgQueue <- m
		}
	}()
}

// put a height in the first 5 bytes of a hash so we can use it to look up dependent message in holding
func HeightToHash(height uint32, minute int) [32]byte {
	var h [32]byte
	h[0] = byte((height >> 24) & 0xFF)
	h[1] = byte((height >> 16) & 0xFF)
	h[2] = byte((height >> 8) & 0xFF)
	h[3] = byte((height >> 0) & 0xFF)
	h[4] = byte(minute)
	return h
}
