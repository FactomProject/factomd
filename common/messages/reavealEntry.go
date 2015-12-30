// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"bytes"
	"fmt"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

//A placeholder structure for messages
type RevealEntryMsg struct {
	Timestamp interfaces.Timestamp
	Entry     *entryBlock.Entry

	//Not marshalled
	hash interfaces.IHash
}

var _ interfaces.IMsg = (*RevealEntryMsg)(nil)

func (m *RevealEntryMsg) Process(interfaces.IState) {}

func (m *RevealEntryMsg) GetHash() interfaces.IHash {
	if m.hash == nil {
		m.hash = m.Entry.GetHash()
	}
	return m.hash
}

func (m *RevealEntryMsg) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp
}

func (m *RevealEntryMsg) Type() int {
	return constants.REVEAL_ENTRY_MSG
}

func (m *RevealEntryMsg) Int() int {
	return -1
}

func (m *RevealEntryMsg) Bytes() []byte {
	return nil
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *RevealEntryMsg) Validate(interfaces.IState) int {
	return 1	// We should validate the size of the reveal and so forth.  But it is
	            // the follower that will choose to relay the reveal to other nodes.
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *RevealEntryMsg) Leader(state interfaces.IState) bool {
	return state.LeaderFor(m.Entry.GetHash().Bytes())
}

// Execute the leader functions of the given message
func (m *RevealEntryMsg) LeaderExecute(state interfaces.IState) error {
	v := m.Validate(state)
	if v <= 0 {
		return fmt.Errorf("Reveal is no longer valid")
	}
	b := m.GetHash()
	
	msg, err := NewAck(state, b)
	
	if err != nil {
		return err
	}
	
	state.NetworkOutMsgQueue() <- msg
	state.FollowerInMsgQueue() <- m   // Send factoid trans to follower
	state.FollowerInMsgQueue() <- msg // Send the Ack to follower
	
	return nil
}

// Returns true if this is a message for this server to execute as a follower
func (m *RevealEntryMsg) Follower(interfaces.IState) bool {
	return true
}

func (m *RevealEntryMsg) FollowerExecute(interfaces.IState) error {
	return nil
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
	newData = data[1:]
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
	data, err = m.Entry.MarshalBinary()
	if err != nil {
		return nil, err
	}
	data = append([]byte{byte(m.Type())}, data...)
	return data, nil
}

func (m *RevealEntryMsg) String() string {
	return "RevealEntryMsg "+m.Timestamp.String()+" "+m.Entry.GetHash().String()
}
