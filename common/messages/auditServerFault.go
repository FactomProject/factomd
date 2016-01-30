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
	Timestamp interfaces.Timestamp

	//Not marshalled
	hash interfaces.IHash
}

var _ interfaces.IMsg = (*AuditServerFault)(nil)

func (e *AuditServerFault) Process(uint32, interfaces.IState) {
	panic("AuditServerFault object should never have its Process() method called")
}

func (m *AuditServerFault) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp
}

func (m *AuditServerFault) GetHash() interfaces.IHash {

	return nil
}

func (m *AuditServerFault) Type() int {
	return constants.AUDIT_SERVER_FAULT_MSG
}

func (m *AuditServerFault) Int() int {
	return -1
}

func (m *AuditServerFault) Bytes() []byte {
	return nil
}

func (m *AuditServerFault) UnmarshalBinaryData(data []byte) (newdata []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()

	return nil, nil
}

func (m *AuditServerFault) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *AuditServerFault) MarshalBinary() (data []byte, err error) {
	return nil, nil
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
func (m *AuditServerFault) Validate(dbheight uint32, state interfaces.IState) int {
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
