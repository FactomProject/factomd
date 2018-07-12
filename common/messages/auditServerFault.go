// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"fmt"
	"reflect"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"

	"github.com/FactomProject/factomd/common/messages/msgbase"
	log "github.com/sirupsen/logrus"
)

//A placeholder structure for messages
type AuditServerFault struct {
	msgbase.MessageBase
	Timestamp interfaces.Timestamp

	Signature interfaces.IFullSignature

	//Not marshalled
	hash interfaces.IHash
}

var _ interfaces.IMsg = (*AuditServerFault)(nil)
var _ interfaces.Signable = (*AuditServerFault)(nil)

func (m *AuditServerFault) GetRepeatHash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("AuditServerFault.GetRepeatHash() saw an interface that was nil")
		}
	}()

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
	signature, err := msgbase.SignSignable(m, key)
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
	return msgbase.VerifyMessage(m)
}

func (e *AuditServerFault) Process(uint32, interfaces.IState) bool {
	panic("AuditServerFault object should never have its Process() method called")
}

func (m *AuditServerFault) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp
}

func (m *AuditServerFault) GetHash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("AuditServerFault.GetHash() saw an interface that was nil")
		}
	}()

	return nil
}

func (m *AuditServerFault) GetMsgHash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("AuditServerFault.GetMsgHash() saw an interface that was nil")
		}
	}()

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

func (m *AuditServerFault) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling AuditServerFault: %v", r)
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

func (m *AuditServerFault) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *AuditServerFault) MarshalForSignature() (data []byte, err error) {
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

func (m *AuditServerFault) MarshalBinary() (data []byte, err error) {
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

func (m *AuditServerFault) String() string {
	return "AuditFault"
}

func (m *AuditServerFault) LogFields() log.Fields {
	return log.Fields{}
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
