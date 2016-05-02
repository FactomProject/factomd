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
type AuditServerFault struct {
	MessageBase
	Timestamp interfaces.Timestamp

	//Not marshalled
	hash interfaces.IHash
}

var _ interfaces.IMsg = (*AuditServerFault)(nil)

func (a *AuditServerFault) IsSameAs(b *AuditServerFault) bool {
	if b == nil {
		return false
	}
	if a.Timestamp != b.Timestamp {
		return false
	}

	//TODO: expand

	return true
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

func (m *AuditServerFault) Int() int {
	return -1
}

func (m *AuditServerFault) Bytes() []byte {
	return nil
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

	newData, err = m.Timestamp.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	//TODO: expand

	return newData, nil
}

func (m *AuditServerFault) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *AuditServerFault) MarshalBinary() (data []byte, err error) {
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

func (m *AuditServerFault) String() string {
	return ""
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

func (m *AuditServerFault) Signature() []byte {
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
func (m *AuditServerFault) Leader(state interfaces.IState) bool {
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
func (m *AuditServerFault) LeaderExecute(state interfaces.IState) error {
	return nil
}

// Returns true if this is a message for this server to execute as a follower
func (m *AuditServerFault) Follower(interfaces.IState) bool {
	return true
}

func (m *AuditServerFault) FollowerExecute(interfaces.IState) error {
	return nil
}

func (e *AuditServerFault) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *AuditServerFault) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *AuditServerFault) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}
