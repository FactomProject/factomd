// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
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
	return m.GetMsgHash()
}

func (m *RevealEntryMsg) GetHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *RevealEntryMsg) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
		data, err := m.MarshalBinary()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
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
func (m *RevealEntryMsg) ValidateRTN(state interfaces.IState) (interfaces.IMsg, int) {
	commit := state.NextCommit(m.Entry.GetHash())

	if commit == nil {
		return nil, 0
	}
	//
	// Make sure one of the two proper commits got us here.
	var okChain, okEntry bool
	m.commitChain, okChain = commit.(*CommitChainMsg)
	m.commitEntry, okEntry = commit.(*CommitEntryMsg)
	if !okChain && !okEntry { // Discard any invalid entries in the map.  Should never happen.
		fmt.Println("dddd Bad EB Commit", state.GetFactomNodeName())
		return m.ValidateRTN(state)
	}

	// Now make sure the proper amount of credits were paid to record the entry.
	// The chain must exist
	if okEntry {
		m.IsEntry = true
		ECs := int(m.commitEntry.CommitEntry.Credits)
		if m.Entry.KSize() > ECs {
			fmt.Println("dddd EB Commit is short", state.GetFactomNodeName())
			return m.ValidateRTN(state) // Discard commits that are not funded properly.
		}

		// Make sure we have a chain.  If we don't, then bad things happen.
		db := state.GetAndLockDB()
		dbheight := state.GetLeaderHeight()
		eb := state.GetNewEBlocks(dbheight, m.Entry.GetChainID())
		eb_db := state.GetNewEBlocks(dbheight-1, m.Entry.GetChainID())
		if eb_db == nil {
			eb_db, _ = db.FetchEBlockHead(m.Entry.GetChainID())
		}

		if eb_db == nil && eb == nil {
			// If we don't have a chain, put the commit back.  Don't want to lose it.
			state.PutCommit(m.Entry.GetHash(), commit)
			return nil, 0
		}
	} else {
		m.IsEntry = false
		ECs := int(m.commitChain.CommitChain.Credits)
		if m.Entry.KSize()+10 > ECs { // Discard commits that are not funded properly
			return m.ValidateRTN(state)
		}
	}

	return commit, 1
}

func (m *RevealEntryMsg) Validate(state interfaces.IState) int {
	commit, rtn := m.ValidateRTN(state)
	if rtn == 1 {
		// Don't lose the commit that validates the entry
		state.PutCommit(m.Entry.GetHash(), commit)
	}
	return rtn
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

func (e *RevealEntryMsg) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
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
