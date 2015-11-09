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
type RevealEntry struct {
	Timestamp interfaces.Timestamp
	Entry *entryBlock.Entry
	hash interfaces.IHash
}

var _ interfaces.IMsg = (*RevealEntry)(nil)

func (m *RevealEntry) GetHash() interfaces.IHash {
	if m.hash == nil {
		data,err := m.Entry.MarshalBinary()
		if err != nil {
			panic(fmt.Sprintf("Error in CommitChain.GetHash(): %s",err.Error()))
		}
		m.hash = primitives.Sha(data)
	}
	return m.hash
}

func (m *RevealEntry) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp
}

func (m *RevealEntry) Type() int {
	return constants.REVEAL_ENTRY_MSG
}

func (m *RevealEntry) Int() int {
	return -1
}

func (m *RevealEntry) Bytes() []byte {
	return nil
}

func (m *RevealEntry) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
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

func (m *RevealEntry) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *RevealEntry) MarshalBinary() (data []byte, err error) {
	data, err = m.Entry.MarshalBinary()
	if err != nil {
		return nil, err
	}
	data = append([]byte{byte(m.Type())}, data...)
	return data, nil
}


func (m *RevealEntry) String() string {
	return ""
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *RevealEntry) Validate(interfaces.IState) int {
	return 0
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *RevealEntry) Leader(state interfaces.IState) bool {
	switch state.GetNetworkNumber() {
	case 0: // Main Network
		panic("Not implemented yet")
	case 1: // Test Network
		panic("Not implemented yet")
	case 2: // Local Network
		panic("Not implemented yet")
	default:
		panic("Not implemented yet")
	}

}

// Execute the leader functions of the given message
func (m *RevealEntry) LeaderExecute(state interfaces.IState) error {
	return nil
}

// Returns true if this is a message for this server to execute as a follower
func (m *RevealEntry) Follower(interfaces.IState) bool {
	return true
}

func (m *RevealEntry) FollowerExecute(interfaces.IState) error {
	return nil
}

func (e *RevealEntry) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *RevealEntry) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *RevealEntry) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}
