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
	MessageBase
	Timestamp interfaces.Timestamp

	Signature interfaces.IFullSignature
}

var _ interfaces.IMsg = (*EOMTimeout)(nil)
var _ Signable = (*EOMTimeout)(nil)

func (a *EOMTimeout) IsSameAs(b *EOMTimeout) bool {
	if b == nil {
		return false
	}
	if a.Timestamp != b.Timestamp {
		return false
	}

	//TODO: expand

	if a.Signature == nil && b.Signature != nil {
		return false
	}
	if a.Signature != nil {
		if a.Signature.IsSameAs(b.Signature) == false {
			return false
		}
	}

	return true
}

func (m *EOMTimeout) Sign(key interfaces.Signer) error {
	signature, err := SignSignable(m, key)
	if err != nil {
		return err
	}
	m.Signature = signature
	return nil
}

func (m *EOMTimeout) GetSignature() interfaces.IFullSignature {
	return m.Signature
}

func (m *EOMTimeout) VerifySignature() (bool, error) {
	return VerifyMessage(m)
}

func (e *EOMTimeout) Process(uint32, interfaces.IState) bool {
	panic("EOMTimeout is not implemented.")
}

func (m *EOMTimeout) GetHash() interfaces.IHash {
	return nil
}

func (m *EOMTimeout) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
		data, err := m.MarshalBinary()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

func (m *EOMTimeout) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp
}

func (m *EOMTimeout) Type() byte {
	return constants.EOM_TIMEOUT_MSG
}

func (m *EOMTimeout) Int() int {
	return -1
}

func (m *EOMTimeout) Bytes() []byte {
	return nil
}

func (m *EOMTimeout) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling Eom Timeout: %v", r)
		}
	}()
	newData = data
	if newData[0] != m.Type() {
		return nil, fmt.Errorf("Invalid Message type")
	}
	newData = newData[1:]

	newData, err = m.Timestamp.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	//TODO: expand

	if len(newData) > 0 {
		m.Signature = new(primitives.Signature)
		newData, err = m.Signature.UnmarshalBinaryData(newData)
		if err != nil {
			return nil, err
		}
	}

	return newData, nil
}

func (m *EOMTimeout) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *EOMTimeout) MarshalForSignature() (data []byte, err error) {
	var buf primitives.Buffer
	buf.Write([]byte{m.Type()})
	if d, err := m.Timestamp.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}

	//TODO: expand

	return buf.DeepCopyBytes(), nil
}

func (m *EOMTimeout) MarshalBinary() (data []byte, err error) {
	resp, err := m.MarshalForSignature()
	if err != nil {
		return nil, err
	}
	sig := m.GetSignature()

	if sig != nil {
		sigBytes, err := sig.MarshalBinary()
		if err != nil {
			return nil, err
		}
		return append(resp, sigBytes...), nil
	}
	return resp, nil
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

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *EOMTimeout) Validate(state interfaces.IState) int {
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
