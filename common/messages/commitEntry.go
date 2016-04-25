// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

//A placeholder structure for messages
type CommitEntryMsg struct {
	MessageBase
	Timestamp   interfaces.Timestamp
	CommitEntry *entryCreditBlock.CommitEntry

	//Not marshalled
	hash interfaces.IHash

	// Not marshaled... Just used by the leader
	count int
}

var _ interfaces.IMsg = (*CommitEntryMsg)(nil)
var _ interfaces.ICounted = (*CommitEntryMsg)(nil)

func (m *CommitEntryMsg) GetCount() int {
	return m.count
}

func (m *CommitEntryMsg) IncCount() {
	m.count += 1
}

func (m *CommitEntryMsg) SetCount(cnt int) {
	m.count = cnt
}

func (m *CommitEntryMsg) Process(dbheight uint32, state interfaces.IState) bool {
	return state.ProcessCommitEntry(dbheight, m)
}

func (m *CommitEntryMsg) GetHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *CommitEntryMsg) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
		data, err := m.MarshalBinary()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

func (m *CommitEntryMsg) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp
}

func (m *CommitEntryMsg) Type() int {
	return constants.COMMIT_ENTRY_MSG
}

func (m *CommitEntryMsg) Int() int {
	return -1
}

func (m *CommitEntryMsg) Bytes() []byte {
	return nil
}

func (m *CommitEntryMsg) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling Commit entry Message: %v", r)
		}
	}()
	newData = data[1:]

	t := new(interfaces.Timestamp)
	newData, err = t.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}
	m.Timestamp = *t

	ce := entryCreditBlock.NewCommitEntry()
	newData, err = ce.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}
	m.CommitEntry = ce
	return newData, nil
}

func (m *CommitEntryMsg) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *CommitEntryMsg) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	binary.Write(&buf, binary.BigEndian, byte(m.Type()))

	t := m.GetTimestamp()
	data, err = t.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	data, err = m.CommitEntry.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	return buf.Bytes(), nil
}

func (m *CommitEntryMsg) String() string {
	return ""
}

func (m *CommitEntryMsg) DBHeight() int {
	return 0
}

func (m *CommitEntryMsg) ChainID() []byte {
	return nil
}

func (m *CommitEntryMsg) ListHeight() int {
	return 0
}

func (m *CommitEntryMsg) SerialHash() []byte {
	return nil
}

func (m *CommitEntryMsg) Signature() []byte {
	return nil
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *CommitEntryMsg) Validate(state interfaces.IState) int {
	//TODO: implement properly, check EC balance
	return 1
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *CommitEntryMsg) Leader(state interfaces.IState) bool {
	//TODO: implement properly
		return state.LeaderFor(m, constants.EC_CHAINID)
	/*
		switch state.GetNetworkNumber() {
		case 0: // Main Network
			panic("Not implemented yet")
		case 1: // Test Network
			panic("Not implemented yet")
		case 2: // Local Network
			panic("Not implemented yet")
		default:
			panic("Not implemented yet")
		}*/

}

// Execute the leader functions of the given message
func (m *CommitEntryMsg) LeaderExecute(state interfaces.IState) error {
	return state.LeaderExecute(m)
}

// Returns true if this is a message for this server to execute as a follower
func (m *CommitEntryMsg) Follower(interfaces.IState) bool {
	return true
}

func (m *CommitEntryMsg) FollowerExecute(state interfaces.IState) error {
	_, err := state.FollowerExecuteMsg(m)
	return err
}

func (e *CommitEntryMsg) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *CommitEntryMsg) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *CommitEntryMsg) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func NewCommitEntryMsg() *CommitEntryMsg {
	return new(CommitEntryMsg)
}
