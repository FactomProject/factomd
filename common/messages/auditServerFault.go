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
type AuditServerFault struct {
	MessageBase
	Timestamp interfaces.Timestamp

	Signature interfaces.IFullSignature

	//Not marshalled
	hash interfaces.IHash
}

var _ interfaces.IMsg = (*AuditServerFault)(nil)
var _ Signable = (*AuditServerFault)(nil)

func (m *AuditServerFault) GetRepeatHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (a *AuditServerFault) IsSameAs(b *AuditServerFault) bool {
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

func (m *AuditServerFault) Sign(key interfaces.Signer) error {
	signature, err := SignSignable(m, key)
	if err != nil {
		return err
	}
	m.Signature = signature
	return nil
}

func (m *AuditServerFault) GetSignature() interfaces.IFullSignature {
	return m.Signature
}

func (m *AuditServerFault) VerifySignature() (bool, error) {
	return VerifyMessage(m)
}

func (e *AuditServerFault) Process(uint32, interfaces.IState) bool {
	panic("AuditServerFault object should never have its Process() method called")
}

func (m *AuditServerFault) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp
}

func (m *AuditServerFault) GetHash() interfaces.IHash {
	return nil
}

func (m *AuditServerFault) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
		data, err := m.MarshalBinary()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

func (m *AuditServerFault) Type() byte {
	return constants.AUDIT_SERVER_FAULT_MSG
}

func (m *AuditServerFault) UnmarshalBinaryData(data []byte) ([]byte, error) {
	buf := primitives.NewBuffer(data)

	t, err := buf.PopByte()
	if err != nil {
		return nil, err
	}
	if t != m.Type() {
		return nil, fmt.Errorf("Invalid Message type")
	}

	m.Timestamp = new(primitives.Timestamp)
	err = buf.PopBinaryMarshallable(m.Timestamp)
	if err != nil {
		return nil, err
	}

	//TODO: expand

	if buf.Len() > 0 {
		m.Signature = new(primitives.Signature)
		err = buf.PopBinaryMarshallable(m.Signature)
		if err != nil {
			return nil, err
		}
	}

	return buf.DeepCopyBytes(), nil
}

func (m *AuditServerFault) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *AuditServerFault) MarshalForSignature() ([]byte, error) {
	buf := primitives.NewBuffer(nil)
	err := buf.PushByte(m.Type())
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(m.Timestamp)
	if err != nil {
		return nil, err
	}

	//TODO: expand

	return buf.DeepCopyBytes(), nil
}

func (m *AuditServerFault) MarshalBinary() ([]byte, error) {
	h, err := m.MarshalForSignature()
	if err != nil {
		return nil, err
	}
	buf := primitives.NewBuffer(h)

	sig := m.GetSignature()
	if sig != nil {
		err := buf.PushBinaryMarshallable(sig)
		if err != nil {
			return nil, err
		}
	}
	return buf.DeepCopyBytes(), nil
}

func (m *AuditServerFault) String() string {
	return "AuditFault"
}

func (m *AuditServerFault) DBHeight() int {
	return 0
}

func (m *AuditServerFault) ChainID() []byte {
	return nil
}

func (m *AuditServerFault) ListHeight() int {
	return 0
}

func (m *AuditServerFault) SerialHash() []byte {
	return nil
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *AuditServerFault) Validate(state interfaces.IState) int {
	return 0
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *AuditServerFault) ComputeVMIndex(state interfaces.IState) {
}

// Execute the leader functions of the given message
func (m *AuditServerFault) LeaderExecute(state interfaces.IState) {
}

func (m *AuditServerFault) FollowerExecute(interfaces.IState) {
}

func (e *AuditServerFault) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *AuditServerFault) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}
