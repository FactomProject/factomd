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
	MessageBase
	Timestamp interfaces.Timestamp

	Signature interfaces.IFullSignature

	//Not marshalled
	hash interfaces.IHash
}

var _ interfaces.IMsg = (*InvalidAck)(nil)
var _ Signable = (*InvalidAck)(nil)

func (a *InvalidAck) IsSameAs(b *InvalidAck) bool {
	if b == nil {
		return false
	}
	if a.Timestamp != b.Timestamp {
		return false
	}

	if a.Signature == nil && b.Signature != nil {
		return false
	}
	if a.Signature != nil {
		if a.Signature.IsSameAs(b.Signature) == false {
			return false
		}
	}
	//TODO: expand

	return true
}

func (m *InvalidAck) Process(uint32, interfaces.IState) bool { return true }

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

func (m *InvalidAck) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
		data, err := m.MarshalBinary()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

func (m *InvalidAck) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp
}

func (m *InvalidAck) Type() byte {
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
			err = fmt.Errorf("Error unmarshalling Invalid Ack: %v", r)
		}
	}()

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

func (m *InvalidAck) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling With Signatures Invalid Ack: %v", r)
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

func (m *InvalidAck) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *InvalidAck) MarshalBinary() (data []byte, err error) {
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

func (m *InvalidAck) GetSignature() interfaces.IFullSignature {
	return m.Signature
}

func (m *InvalidAck) VerifySignature() (bool, error) {
	return VerifyMessage(m)
}

func (m *InvalidAck) Sign(key interfaces.Signer) error {
	signature, err := SignSignable(m, key)
	if err != nil {
		return err
	}
	m.Signature = signature
	return nil
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

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *InvalidAck) Validate(state interfaces.IState) int {
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
