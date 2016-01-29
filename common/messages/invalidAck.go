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

//A placeholder structure for messages
type InvalidAck struct {
	Timestamp interfaces.Timestamp

	//Not marshalled
	hash interfaces.IHash
}

var _ interfaces.IMsg = (*InvalidAck)(nil)

func (m *InvalidAck) Process(uint32, interfaces.IState) {}

func (m *InvalidAck) GetHash() interfaces.IHash {
	if m.hash == nil {
		data, err := m.MarshalForSignature()
		if err != nil {
			panic(fmt.Sprintf("Error in CommitChain.GetHash(): %s", err.Error()))
		}
		m.hash = primitives.Sha(data)
	}
	return m.hash
}

func (m *InvalidAck) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp
}

func (m *InvalidAck) Type() int {
	return constants.INVALID_ACK_MSG
}

func (m *InvalidAck) Int() int {
	return -1
}

func (m *InvalidAck) Bytes() []byte {
	return nil
}

func (m *InvalidAck) MarshalForSignature() (data []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()

	return nil, nil
}

func (m *InvalidAck) UnmarshalBinaryData(data []byte) (newdata []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()

	return nil, nil
}

func (m *InvalidAck) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *InvalidAck) MarshalBinary() (data []byte, err error) {
	return nil, nil
}

func (m *InvalidAck) String() string {
	return ""
}

func (m *InvalidAck) DBHeight() int {
	return 0
}

func (m *InvalidAck) ChainID() []byte {
	return nil
}

func (m *InvalidAck) ListHeight() int {
	return 0
}

func (m *InvalidAck) SerialHash() []byte {
	return nil
}

func (m *InvalidAck) Signature() []byte {
	return nil
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *InvalidAck) Validate(dbheight uint32, state interfaces.IState) int {
	return 0
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *InvalidAck) Leader(state interfaces.IState) bool {
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
func (m *InvalidAck) LeaderExecute(state interfaces.IState) error {
	return nil
}

// Returns true if this is a message for this server to execute as a follower
func (m *InvalidAck) Follower(interfaces.IState) bool {
	return true
}

func (m *InvalidAck) FollowerExecute(interfaces.IState) error {
	return nil
}

func (e *InvalidAck) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *InvalidAck) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *InvalidAck) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}
