package state

import (
	"fmt"
	"sync"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/util/atomic"
)

var _ = fmt.Println

// SafeMsgMap is a threadsafe map[[32]byte]interfaces.IMsg
type SafeMsgMap struct {
	msgmap map[[32]byte]interfaces.IMsg
	sync.RWMutex
	name string
	s    *State
}

func NewSafeMsgMap(name string, s *State) *SafeMsgMap {
	m := new(SafeMsgMap)
	m.msgmap = make(map[[32]byte]interfaces.IMsg)
	m.name = name
	m.s = s
	return m
}

func (m *SafeMsgMap) Get(key [32]byte) (msg interfaces.IMsg) {
	m.RLock()
	defer m.RUnlock()
	return m.msgmap[key]
}

func (m *SafeMsgMap) Put(key [32]byte, msg interfaces.IMsg) {
	m.Lock()
	_, ok := m.msgmap[key]
	if !ok {
		defer m.s.LogMessage(m.name, "put", msg)
	}
	m.msgmap[key] = msg
	m.Unlock()
}

func (m *SafeMsgMap) Delete(key [32]byte) (msg interfaces.IMsg, found bool) {
	m.Lock()
	msg, ok := m.msgmap[key] // return the message being deleted
	if ok {
		defer m.s.LogMessage(m.name, fmt.Sprintf("delete from %s", atomic.WhereAmIString(1)), msg)
		delete(m.msgmap, key)
	} else {
		defer m.s.LogPrintf(m.name, "nodelete from %s M-%x", atomic.WhereAmIString(1), key[:3])
	}
	m.Unlock()
	return
}

func (m *SafeMsgMap) Len() int {
	m.RLock()
	defer m.RUnlock()
	return len(m.msgmap)
}

func (m *SafeMsgMap) Copy() *SafeMsgMap {
	m2 := NewSafeMsgMap("copyOf"+m.name, m.s)
	m2.s = m.s // for debug logging
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
	m.s.LogPrintf(m.name, "reset")
}

//
// Used if a Commit Map
//

// Cleanup will clean old elements out from the commit map.
func (m *SafeMsgMap) Cleanup(s *State) {
	m.Lock()
	// Time out commits every leaderTimestamp and again. Also check for entries that have been revealed
	leaderTimestamp := s.GetLeaderTimestamp()
	for k, msg := range m.msgmap {
		_, ok := s.Replay.Valid(constants.TIME_TEST, msg.GetRepeatHash().Fixed(), msg.GetTimestamp(), leaderTimestamp)
		if !ok {
			msg, ok := m.msgmap[k]
			if ok {
				defer m.s.LogMessage(m.name, "cleanup_timeout", msg)
			}
			delete(m.msgmap, k)
			continue
		}
		ok = s.Replay.IsHashUnique(constants.REVEAL_REPLAY, k)
		if !ok {
			msg, ok := m.msgmap[k]
			if ok {
				defer m.s.LogMessage(m.name, "cleanup_replay", msg)
			}
			delete(m.msgmap, k)
			continue
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
			_, ok := s.Replay.Valid(constants.TIME_TEST, v.GetRepeatHash().Fixed(), v.GetTimestamp(), s.GetLeaderTimestamp())
			if !ok {
				defer m.s.LogMessage(m.name, "RemoveExpired", v)
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
