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
type SignatureTimeout struct {
	Timestamp interfaces.Timestamp

	//Not marshalled
	hash interfaces.IHash
}

var _ interfaces.IMsg = (*SignatureTimeout)(nil)

func (m *SignatureTimeout) Process(uint32, interfaces.IState) {}

func (m *SignatureTimeout) GetHash() interfaces.IHash {
	if m.hash == nil {
		data, err := m.MarshalForSignature()
		if err != nil {
			panic(fmt.Sprintf("Error in CommitChain.GetHash(): %s", err.Error()))
		}
		m.hash = primitives.Sha(data)
	}
	return m.hash
}

func (m *SignatureTimeout) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp
}

func (m *SignatureTimeout) Type() int {
	return constants.SIGNATURE_TIMEOUT_MSG
}

func (m *SignatureTimeout) Int() int {
	return -1
}

func (m *SignatureTimeout) Bytes() []byte {
	return nil
}

func (m *SignatureTimeout) UnmarshalBinaryData(data []byte) (newdata []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()

	return nil, nil
}

func (m *SignatureTimeout) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *SignatureTimeout) MarshalForSignature() (data []byte, err error) {
	return nil, nil
}

func (m *SignatureTimeout) MarshalBinary() (data []byte, err error) {
	return nil, nil
}

func (m *SignatureTimeout) String() string {
	return ""
}

func (m *SignatureTimeout) DBHeight() int {
	return 0
}

func (m *SignatureTimeout) ChainID() []byte {
	return nil
}

func (m *SignatureTimeout) ListHeight() int {
	return 0
}

func (m *SignatureTimeout) SerialHash() []byte {
	return nil
}

func (m *SignatureTimeout) Signature() []byte {
	return nil
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *SignatureTimeout) Validate(dbheight uint32, state interfaces.IState) int {
	return 0
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *SignatureTimeout) Leader(state interfaces.IState) bool {
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
func (m *SignatureTimeout) LeaderExecute(state interfaces.IState) error {
	return nil
}

// Returns true if this is a message for this server to execute as a follower
func (m *SignatureTimeout) Follower(interfaces.IState) bool {
	return true
}

func (m *SignatureTimeout) FollowerExecute(interfaces.IState) error {
	return nil
}

func (e *SignatureTimeout) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *SignatureTimeout) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *SignatureTimeout) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}
