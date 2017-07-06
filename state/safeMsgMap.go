package state

import (
	"sync"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
)

// CommitMap is a threadsafe map[[32]byte]interfaces.IMsg
type CommitMap struct {
	msgmap map[[32]byte]interfaces.IMsg
	sync.RWMutex
}

func NewCommitMap() *CommitMap {
	m := new(CommitMap)
	m.msgmap = make(map[[32]byte]interfaces.IMsg)

	return m
}

func (m *CommitMap) Get(key [32]byte) (msg interfaces.IMsg) {
	m.RLock()
	defer m.RUnlock()
	return m.msgmap[key]
}

func (m *CommitMap) Put(key [32]byte, msg interfaces.IMsg) {
	m.Lock()
	m.msgmap[key] = msg
	m.Unlock()
}

func (m *CommitMap) Delete(key [32]byte) (msg interfaces.IMsg, found bool) {
	m.Lock()
	delete(m.msgmap, key)
	m.Unlock()
	return
}

func (m *CommitMap) Len() int {
	return len(m.msgmap)
}

func (m *CommitMap) Copy() *CommitMap {
	m2 := NewCommitMap()

	m.RLock()
	for k, v := range m.msgmap {
		m2.msgmap[k] = v
	}
	m.RUnlock()

	return m2
}

// GetRaw is used in testing and simcontrol. Do no use this in production
func (m *CommitMap) GetRaw() map[[32]byte]interfaces.IMsg {
	raw := m.Copy()
	return raw.msgmap
}

// Cleanup will clean old elements out from the commit map.
func (m *CommitMap) Cleanup(s *State) {
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
	}
	m.Unlock()
}

func (m *CommitMap) RemoveExpired(s *State) {
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
