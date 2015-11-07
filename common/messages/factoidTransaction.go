// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"bytes"
	"fmt"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

//A placeholder structure for messages
type FactoidTransaction struct {
	Transaction interfaces.ITransaction
}

var _ interfaces.IMsg = (*FactoidTransaction)(nil)

func (m *FactoidTransaction) GetHash() interfaces.IHash {
	if m.hash == nil {
		data,err := m.Transaction.MarshalBinarySig()
		if err != nil {
			panic(fmt.Sprintf("Error in CommitChain.GetHash(): %s",err.Error()))
		}
		m.hash = primitives.Sha(data)
	}
	return m.hash
}

func (m *FactoidTransaction) GetTimestamp() interfaces.Timestamp {
	return Timestamp(m.Transaction.GetMilliTimestamp())
}

func (m *FactoidTransaction) GetTransaction() interfaces.ITransaction {
	return m.Transaction
}

func (m *FactoidTransaction) SetTransaction(transaction interfaces.ITransaction) {
	m.Transaction = transaction
}

func (m *FactoidTransaction) Type() int {
	return constants.FACTOID_TRANSACTION_MSG
}

func (m *FactoidTransaction) Int() int {
	return -1
}

func (m *FactoidTransaction) Bytes() []byte {
	return nil
}

func (m *FactoidTransaction) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	newData = data[1:]

	m.Transaction = new(factoid.Transaction)
	newData, err = m.Transaction.UnmarshalBinaryData(newData)

	return newData, err
}

func (m *FactoidTransaction) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *FactoidTransaction) MarshalBinary() (data []byte, err error) {
	data, err = m.Transaction.MarshalBinary()
	if err != nil {
		return nil, err
	}
	data = append([]byte{byte(m.Type())}, data...)
	return data, nil
}

func (m *FactoidTransaction) String() string {
	return ""
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *FactoidTransaction) Validate(interfaces.IState) int {
	return 0
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *FactoidTransaction) Leader(state interfaces.IState) bool {
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
func (m *FactoidTransaction) LeaderExecute(state interfaces.IState) error {
	return nil
}

// Returns true if this is a message for this server to execute as a follower
func (m *FactoidTransaction) Follower(interfaces.IState) bool {
	return true
}

func (m *FactoidTransaction) FollowerExecute(interfaces.IState) error {
	return nil
}

func (e *FactoidTransaction) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *FactoidTransaction) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *FactoidTransaction) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}
