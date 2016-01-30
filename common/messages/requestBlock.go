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
type RequestBlock struct {
	Timestamp interfaces.Timestamp

	//Not marshalled
	hash interfaces.IHash
}

var _ interfaces.IMsg = (*RequestBlock)(nil)

func (m *RequestBlock) Process(uint32, interfaces.IState) {}

func (m *RequestBlock) GetHash() interfaces.IHash {
	if m.hash == nil {
		data, err := m.MarshalForSignature()
		if err != nil {
			panic(fmt.Sprintf("Error in CommitChain.GetHash(): %s", err.Error()))
		}
		m.hash = primitives.Sha(data)
	}
	return m.hash
}

func (m *RequestBlock) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp
}

func (m *RequestBlock) Type() int {
	return constants.REQUEST_BLOCK_MSG
}

func (m *RequestBlock) Int() int {
	return -1
}

func (m *RequestBlock) Bytes() []byte {
	return nil
}

func (m *RequestBlock) UnmarshalBinaryData(data []byte) (newdata []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()

	return nil, nil
}

func (m *RequestBlock) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *RequestBlock) MarshalForSignature() (data []byte, err error) {
	return nil, nil
}

func (m *RequestBlock) MarshalBinary() (data []byte, err error) {
	return nil, nil
}

func (m *RequestBlock) String() string {
	return ""
}

func (m *RequestBlock) DBHeight() int {
	return 0
}

func (m *RequestBlock) ChainID() []byte {
	return nil
}

func (m *RequestBlock) ListHeight() int {
	return 0
}

func (m *RequestBlock) SerialHash() []byte {
	return nil
}

func (m *RequestBlock) Signature() []byte {
	return nil
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *RequestBlock) Validate(dbheight uint32, state interfaces.IState) int {
	return 0
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *RequestBlock) Leader(state interfaces.IState) bool {
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
func (m *RequestBlock) LeaderExecute(state interfaces.IState) error {
	return nil
}

// Returns true if this is a message for this server to execute as a follower
func (m *RequestBlock) Follower(interfaces.IState) bool {
	return true
}

func (m *RequestBlock) FollowerExecute(interfaces.IState) error {
	return nil
}

func (e *RequestBlock) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *RequestBlock) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *RequestBlock) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}
