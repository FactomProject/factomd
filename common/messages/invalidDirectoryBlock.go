// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

//A placeholder structure for messages
type InvalidDirectoryBlock struct {
	MessageBase
	Timestamp interfaces.Timestamp

	Signature interfaces.IFullSignature

	//Not marshalled
	hash interfaces.IHash
}

var _ interfaces.IMsg = (*InvalidDirectoryBlock)(nil)
var _ Signable = (*InvalidDirectoryBlock)(nil)

func (a *InvalidDirectoryBlock) IsSameAs(b *InvalidDirectoryBlock) bool {
	if b == nil {
		return false
	}
	if a.Timestamp.GetTimeMilli() != b.Timestamp.GetTimeMilli() {
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

func (m *InvalidDirectoryBlock) Sign(key interfaces.Signer) error {
	signature, err := SignSignable(m, key)
	if err != nil {
		return err
	}
	m.Signature = signature
	return nil
}

func (m *InvalidDirectoryBlock) GetSignature() interfaces.IFullSignature {
	return m.Signature
}

func (m *InvalidDirectoryBlock) VerifySignature() (bool, error) {
	return VerifyMessage(m)
}

func (m *InvalidDirectoryBlock) Process(uint32, interfaces.IState) bool { return true }

func (m *InvalidDirectoryBlock) GetRepeatHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *InvalidDirectoryBlock) GetHash() interfaces.IHash {
	if m.hash == nil {
		data, err := m.MarshalForSignature()
		if err != nil {
			panic(fmt.Sprintf("Error in CommitChain.GetHash(): %s", err.Error()))
		}
		m.hash = primitives.Sha(data)
	}
	return m.hash
}

func (m *InvalidDirectoryBlock) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
		data, err := m.MarshalBinary()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

func (m *InvalidDirectoryBlock) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp
}

func (m *InvalidDirectoryBlock) Type() byte {
	return constants.INVALID_DIRECTORY_BLOCK_MSG
}

func (m *InvalidDirectoryBlock) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	newData = data
	if newData[0] != m.Type() {
		return nil, fmt.Errorf("Invalid Message type")
	}
	newData = newData[1:]

	m.Timestamp = new(primitives.Timestamp)
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

func (m *InvalidDirectoryBlock) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *InvalidDirectoryBlock) MarshalBinary() (data []byte, err error) {
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

func (m *InvalidDirectoryBlock) MarshalForSignature() (data []byte, err error) {
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

func (m *InvalidDirectoryBlock) String() string {
	return "Invalid Directory Block"
}

func (m *InvalidDirectoryBlock) DBHeight() int {
	return 0
}

func (m *InvalidDirectoryBlock) ChainID() []byte {
	return nil
}

func (m *InvalidDirectoryBlock) ListHeight() int {
	return 0
}

func (m *InvalidDirectoryBlock) SerialHash() []byte {
	return nil
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *InvalidDirectoryBlock) Validate(state interfaces.IState) int {
	return 0
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *InvalidDirectoryBlock) ComputeVMIndex(state interfaces.IState) {
}

// Execute the leader functions of the given message
func (m *InvalidDirectoryBlock) LeaderExecute(state interfaces.IState) {
}

func (m *InvalidDirectoryBlock) FollowerExecute(interfaces.IState) {
}

func (e *InvalidDirectoryBlock) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *InvalidDirectoryBlock) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}
