package state

import (
	"fmt"
	"sync"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
)

var _ = fmt.Println

// SafeMsgMap is a threadsafe map[[32]byte]interfaces.IMsg
type SafeMsgMap struct {
	msgmap map[[32]byte]interfaces.IMsg
	sync.RWMutex
}

func NewSafeMsgMap() *SafeMsgMap {
	m := new(SafeMsgMap)
	m.msgmap = make(map[[32]byte]interfaces.IMsg)

	return m
}

func (m *SafeMsgMap) Get(key [32]byte) (msg interfaces.IMsg) {
	m.RLock()
	defer m.RUnlock()
	return m.msgmap[key]
}

func (m *SafeMsgMap) Put(key [32]byte, msg interfaces.IMsg) {
	m.Lock()
	m.msgmap[key] = msg
	m.Unlock()
}

func (m *SafeMsgMap) Delete(key [32]byte) (msg interfaces.IMsg, found bool) {
	m.Lock()
	delete(m.msgmap, key)
	m.Unlock()
	return
}

func (m *SafeMsgMap) Len() int {
	m.Lock()
	defer m.Unlock()
	return len(m.msgmap)
}

func (m *SafeMsgMap) Copy() *SafeMsgMap {
	m2 := NewSafeMsgMap()

	m.RLock()
	for k, v := range m.msgmap {
		m2.msgmap[k] = v
	}
	m.RUnlock()

	return m2
}

// Reset will delete all elements
func (m *SafeMsgMap) Reset() {
	m.Lock()
	if len(m.msgmap) > 0 {
		m.msgmap = make(map[[32]byte]interfaces.IMsg)
	}
	m.Unlock()
}

//
// Used if a Commit Map
//

// Cleanup will clean old elements out from the commit map.
func (m *SafeMsgMap) Cleanup(s *State) {
	m.Lock()
	// Time out commits every now and again. Also check for entries that have been revealed
	now := s.GetTimestamp()
	for k, msg := range m.msgmap {
		{
			c, ok := msg.(*messages.CommitChainMsg)
			if ok && !s.NoEntryYet(c.CommitChain.EntryHash, now) {
				delete(m.msgmap, k)
				continue
			}
		}
		c, ok := msg.(*messages.CommitEntryMsg)
		if ok && !s.NoEntryYet(c.CommitEntry.EntryHash, now) {
			delete(m.msgmap, k)
			continue
		}

		_, ok = s.Replay.Valid(constants.TIME_TEST, msg.GetRepeatHash().Fixed(), msg.GetTimestamp(), now)
		if !ok {
			delete(m.msgmap, k)
		}
		ok = s.Replay.IsHashUnique(constants.REVEAL_REPLAY, k)
		if !ok {
			delete(m.msgmap, k)
		}
	}
	m.Unlock()
}

// RemoveExpired is used when treating this as a commit map. Do not
func (m *SafeMsgMap) RemoveExpired(s *State) {
	m.Lock()
	// Time out commits every now and again.
	for k, v := range m.msgmap {
		if v != nil {
			_, ok := s.Replay.Valid(constants.TIME_TEST, v.GetRepeatHash().Fixed(), v.GetTimestamp(), s.GetTimestamp())
			if !ok {
				delete(m.msgmap, k)
			}
		}
	}
	m.Unlock()
}

//
// For tests
//

// GetRaw is used in testing and simcontrol. Do no use this in production
func (m *SafeMsgMap) GetRaw() map[[32]byte]interfaces.IMsg {
	raw := m.Copy()
	return raw.msgmap
}
