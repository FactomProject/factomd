// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"encoding/binary"
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"

	log "github.com/FactomProject/logrus"
)

//A placeholder structure for messages
type RevealEntryMsg struct {
	MessageBase
	Timestamp interfaces.Timestamp
	Entry     interfaces.IEntry

	//No signature!

	//Not marshalled
	hash        interfaces.IHash
	chainIDHash interfaces.IHash
	IsEntry     bool
	commitChain *CommitChainMsg
	commitEntry *CommitEntryMsg
}

var _ interfaces.IMsg = (*RevealEntryMsg)(nil)

func (m *RevealEntryMsg) IsSameAs(msg interfaces.IMsg) bool {
	m2, ok := msg.(*RevealEntryMsg)
	if !ok {
		return false
	}
	if !m.GetMsgHash().IsSameAs(m2.GetMsgHash()) {
		return false
	}
	return true
}

func (m *RevealEntryMsg) Process(dbheight uint32, state interfaces.IState) bool {
	return state.ProcessRevealEntry(dbheight, m)
}

func (m *RevealEntryMsg) GetRepeatHash() interfaces.IHash {
	return m.Entry.GetHash()
}

func (m *RevealEntryMsg) GetHash() interfaces.IHash {
	return m.Entry.GetHash()
}

func (m *RevealEntryMsg) GetMsgHash() interfaces.IHash {
	return m.Entry.GetHash()
}

func (m *RevealEntryMsg) GetChainIDHash() interfaces.IHash {
	if m.chainIDHash == nil {
		m.chainIDHash = primitives.Sha(m.Entry.GetChainID().Bytes())
	}
	return m.chainIDHash
}

func (m *RevealEntryMsg) GetTimestamp() interfaces.Timestamp {
	if m.Timestamp == nil {
		m.Timestamp = new(primitives.Timestamp)
	}
	return m.Timestamp
}

func (m *RevealEntryMsg) Type() byte {
	return constants.REVEAL_ENTRY_MSG
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
// Also return the matching commit, if 1 (Don't put it back into the Commit List)
func (m *RevealEntryMsg) Validate(state interfaces.IState) int {
	commit := state.NextCommit(m.Entry.GetHash())

	if commit == nil {
		return 0
	}
	//
	// Make sure one of the two proper commits got us here.
	var okChain, okEntry bool
	m.commitChain, okChain = commit.(*CommitChainMsg)
	m.commitEntry, okEntry = commit.(*CommitEntryMsg)
	if !okChain && !okEntry { // What is this trash doing here?  Not a commit at all!
		return -1
	}

	// Now make sure the proper amount of credits were paid to record the entry.
	// The chain must exist
	if okEntry {
		m.IsEntry = true
		ECs := int(m.commitEntry.CommitEntry.Credits)
		// Any entry over 10240 bytes will be rejected
		if m.Entry.KSize() > 10 {
			return -1
		}

		if m.Entry.KSize() > ECs {
			return 0 // not enough payments on the EC to reveal this entry.  Return 0 to wait on another commit
		}

		// Make sure we have a chain.  If we don't, then bad things happen.
		db := state.GetAndLockDB()
		dbheight := state.GetLeaderHeight()
		eb := state.GetNewEBlocks(dbheight, m.Entry.GetChainID())
		if eb == nil {
			eb_db := state.GetNewEBlocks(dbheight-1, m.Entry.GetChainID())
			eb = eb_db
		}
		if eb == nil {
			eb_db2 := state.GetNewEBlocks(dbheight-2, m.Entry.GetChainID())
			eb = eb_db2
		}
		if eb == nil {
			eb, _ = db.FetchEBlockHead(m.Entry.GetChainID())
		}

		if eb == nil {
			// No chain, we have to leave it be and maybe one will be made.
			return 0
		}
		return 1
	}

	m.IsEntry = false
	ECs := int(m.commitChain.CommitChain.Credits)
	if m.Entry.KSize()+10 > ECs {
		return 0 // Wait for a commit that might fund us properly
	}

	return 1
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *RevealEntryMsg) ComputeVMIndex(state interfaces.IState) {
	m.VMIndex = state.ComputeVMIndex(m.Entry.GetChainID().Bytes())
}

// Execute the leader functions of the given message
func (m *RevealEntryMsg) LeaderExecute(state interfaces.IState) {
	state.LeaderExecuteRevealEntry(m)
}

func (m *RevealEntryMsg) FollowerExecute(state interfaces.IState) {
	state.FollowerExecuteRevealEntry(m)
}

func (e *RevealEntryMsg) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *RevealEntryMsg) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func NewRevealEntryMsg() *RevealEntryMsg {
	return new(RevealEntryMsg)
}

func (m *RevealEntryMsg) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	newData = data
	if newData[0] != m.Type() {
		return nil, fmt.Errorf("%s", "Invalid Message type")
	}
	newData = newData[1:]

	t := new(primitives.Timestamp)
	newData, err = t.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}
	m.Timestamp = t

	e := entryBlock.NewEntry()
	newData, err = e.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}
	m.Entry = e

	return newData, nil
}

func (m *RevealEntryMsg) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *RevealEntryMsg) MarshalBinary() (data []byte, err error) {
	var buf primitives.Buffer

	binary.Write(&buf, binary.BigEndian, m.Type())

	t := m.GetTimestamp()
	data, err = t.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	data, err = m.Entry.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	return buf.DeepCopyBytes(), nil
}

func (m *RevealEntryMsg) String() string {
	if m.GetLeaderChainID() == nil {
		m.SetLeaderChainID(primitives.NewZeroHash())
	}
	str := fmt.Sprintf("%6s-VM%3d: Min:%4d          -- Leader[%x] Entry[%x] ChainID[%x] hash[%x]",
		"REntry",
		m.VMIndex,
		m.Minute,
		m.GetLeaderChainID().Bytes()[:5],
		m.Entry.GetHash().Bytes()[:3],
		m.Entry.GetChainID().Bytes()[:5],
		m.GetHash().Bytes()[:3])

	return str
}

func (m *RevealEntryMsg) LogFields() log.Fields {
	return log.Fields{"category": "message", "messagetype": "revealentry",
		"vm":         m.VMIndex,
		"minute":     m.Minute,
		"leaderid":   m.GetLeaderChainID().String()[4:10],
		"entryhash":  m.Entry.GetHash().String()[:6],
		"entrychain": m.Entry.GetChainID().String()[:6],
		"hash":       m.GetHash().String()[:6]}
}
