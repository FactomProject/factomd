// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"bytes"
	"fmt"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

//Structure to request missing messages in a node's process list
type MissingMsg struct {
	MessageBase
	DBHeight          int
	ProcessListHeight int
	Timestamp         interfaces.Timestamp

	//Not marshalled
	hash interfaces.IHash
}

var _ interfaces.IMsg = (*MissingMsg)(nil)

func (m *MissingMsg) Process(uint32, interfaces.IState) bool { return true }

func (m *MissingMsg) GetHash() interfaces.IHash {
	if m.hash == nil {
		data, err := m.MarshalForSignature()
		if err != nil {
			panic(fmt.Sprintf("Error in MissingMsg.GetHash(): %s", err.Error()))
		}
		m.hash = primitives.Sha(data)
	}
	return m.hash
}

func (m *MissingMsg) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
		data, err := m.MarshalBinary()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

func (m *MissingMsg) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp
}

func (m *MissingMsg) Type() int {
	return constants.MISSING_MSG
}

func (m *MissingMsg) Int() int {
	return -1
}

func (m *MissingMsg) Bytes() []byte {
	return nil
}

func (m *MissingMsg) UnmarshalBinaryData(data []byte) (newdata []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()

	return nil, nil
}

func (m *MissingMsg) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *MissingMsg) MarshalBinary() (data []byte, err error) {
	return nil, nil
}

func (m *MissingMsg) MarshalForSignature() (data []byte, err error) {
	return nil, nil
}

func (m *MissingMsg) String() string {
	return ""
}

func (m *MissingMsg) ChainID() []byte {
	return nil
}

func (m *MissingMsg) ListHeight() int {
	return 0
}

func (m *MissingMsg) Signature() []byte {
	return nil
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *MissingMsg) Validate(state interfaces.IState) int {
	return 1
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *MissingMsg) Leader(state interfaces.IState) bool {
	return false
}

// Execute the leader functions of the given message
func (m *MissingMsg) LeaderExecute(state interfaces.IState) error {
	return nil
}

// Returns true if this is a message for this server to execute as a follower
func (m *MissingMsg) Follower(interfaces.IState) bool {
	return true
}

func (m *MissingMsg) FollowerExecute(interfaces.IState) error {
	return nil
}

func (e *MissingMsg) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *MissingMsg) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *MissingMsg) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}
