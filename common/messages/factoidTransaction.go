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

	//Not marshalled
	hash      interfaces.IHash
	processed bool
}

var _ interfaces.IMsg = (*FactoidTransaction)(nil)

func (m *FactoidTransaction) GetHash() interfaces.IHash {
	if m.hash == nil {
		data, err := m.Transaction.MarshalBinarySig()
		if err != nil {
			panic(fmt.Sprintf("Error in CommitChain.GetHash(): %s", err.Error()))
		}
		m.hash = primitives.Sha(data)
	}
	return m.hash
}

func (m *FactoidTransaction) GetTimestamp() interfaces.Timestamp {
	return interfaces.Timestamp(m.Transaction.GetMilliTimestamp())
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

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *FactoidTransaction) Validate(dbheight uint32, state interfaces.IState) int {
	err := state.GetFactoidState(dbheight).Validate(1, m.Transaction)
	if err != nil {
		fmt.Println(err.Error())
		return -1
	}
	return 1
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *FactoidTransaction) Leader(state interfaces.IState) bool {
	return state.LeaderFor(constants.FACTOID_CHAINID)
}

// Execute the leader functions of the given message
func (m *FactoidTransaction) LeaderExecute(state interfaces.IState) error {
	if err := state.GetFactoidState(state.GetDBHeight()).Validate(1, m.Transaction); err != nil {
		return err
	}
	msg, err := NewAck(state, m.GetHash())
	if err != nil {
		return err
	}
	state.NetworkOutMsgQueue() <- msg // Send the Ack to the network
	state.FollowerInMsgQueue() <- msg // Send the Ack to follower
	state.FollowerInMsgQueue() <- m   // Send factoid trans to follower
	return nil
}

// Returns true if this is a message for this server to execute as a follower
func (m *FactoidTransaction) Follower(state interfaces.IState) bool {
	return true
}

func (m *FactoidTransaction) FollowerExecute(state interfaces.IState) error {
	_, err := state.MatchAckFollowerExecute(m)
	return err
}

func (m *FactoidTransaction) Process(dbheight uint32, state interfaces.IState) {

	if m.processed {
		return
	}
	m.processed = true
	fmt.Println("Process Factoid")
	// We can only get a Factoid Transaction once.  Add it, and remove it from the lists.
	state.GetFactoidState(dbheight).AddTransaction(1, m.Transaction)

}

func (m *FactoidTransaction) Int() int {
	return -1
}

func (m *FactoidTransaction) Bytes() []byte {
	return nil
}

func (m *FactoidTransaction) UnmarshalTransData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()

	m.Transaction = new(factoid.Transaction)
	newData, err = m.Transaction.UnmarshalBinaryData(data)

	return newData, err
}

func (m *FactoidTransaction) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()

	newData = data[1:]

	return m.UnmarshalTransData(newData)
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
	return "Factoid Transaction " + m.GetHash().String()
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
