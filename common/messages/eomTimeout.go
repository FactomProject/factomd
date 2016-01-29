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
type EOMTimeout struct {
	Timestamp interfaces.Timestamp
}

var _ interfaces.IMsg = (*EOMTimeout)(nil)

func (e *EOMTimeout) Process(uint32, interfaces.IState) {
	panic("EOMTimeout is not implemented.")
}

func (m *EOMTimeout) GetHash() interfaces.IHash {
	return nil
}

func (m *EOMTimeout) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp
}

func (m *EOMTimeout) Type() int {
	return constants.EOM_TIMEOUT_MSG
}

func (m *EOMTimeout) Int() int {
	return -1
}

func (m *EOMTimeout) Bytes() []byte {
	return nil
}

func (m *EOMTimeout) UnmarshalBinaryData(data []byte) (newdata []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()

	return nil, nil
}

func (m *EOMTimeout) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *EOMTimeout) MarshalBinary() (data []byte, err error) {
	return nil, nil
}

func (m *EOMTimeout) String() string {
	return ""
}

func (m *EOMTimeout) DBHeight() int {
	return 0
}

func (m *EOMTimeout) ChainID() []byte {
	return nil
}

func (m *EOMTimeout) ListHeight() int {
	return 0
}

func (m *EOMTimeout) SerialHash() []byte {
	return nil
}

func (m *EOMTimeout) Signature() []byte {
	return nil
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *EOMTimeout) Validate(dbheight uint32, state interfaces.IState) int {
	return 0
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *EOMTimeout) Leader(state interfaces.IState) bool {
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
func (m *EOMTimeout) LeaderExecute(state interfaces.IState) error {
	return nil
}

// Returns true if this is a message for this server to execute as a follower
func (m *EOMTimeout) Follower(interfaces.IState) bool {
	return true
}

func (m *EOMTimeout) FollowerExecute(interfaces.IState) error {
	return nil
}

func (e *EOMTimeout) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *EOMTimeout) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *EOMTimeout) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}
