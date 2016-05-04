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
	MessageBase
	Timestamp   interfaces.Timestamp
	Transaction interfaces.ITransaction

	//No signature!

	//Not marshalled
	hash      interfaces.IHash
	processed bool
}

var _ interfaces.IMsg = (*FactoidTransaction)(nil)

func (a *FactoidTransaction) IsSameAs(b *FactoidTransaction) bool {
	if b == nil {
		return false
	}
	if a.Timestamp != b.Timestamp {
		return false
	}

	ok, err := primitives.AreBinaryMarshallablesEqual(a.Transaction, b.Transaction)
	if err != nil || ok == false {
		return false
	}

	return true
}

func (m *FactoidTransaction) GetHash() interfaces.IHash {

		data, err := m.Transaction.MarshalBinarySig()
		if err != nil {
			panic(fmt.Sprintf("Error in CommitChain.GetHash(): %s", err.Error()))
		}
		m.hash = primitives.Sha(data)

	return m.hash
}

func (m *FactoidTransaction) GetMsgHash() interfaces.IHash {

		data, err := m.MarshalBinary()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)

	return m.MsgHash
}

func (m *FactoidTransaction) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp
}

func (m *FactoidTransaction) GetTransaction() interfaces.ITransaction {
	return m.Transaction
}

func (m *FactoidTransaction) SetTransaction(transaction interfaces.ITransaction) {
	m.Transaction = transaction
}

func (m *FactoidTransaction) Type() byte {
	return constants.FACTOID_TRANSACTION_MSG
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *FactoidTransaction) Validate(state interfaces.IState) int {
	err := state.GetFactoidState().Validate(1, m.Transaction)
	if err != nil {
		return -1
	}
	return 1
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *FactoidTransaction) Leader(state interfaces.IState) bool {
	return state.LeaderFor(m, constants.FACTOID_CHAINID)
}

// Execute the leader functions of the given message
func (m *FactoidTransaction) LeaderExecute(state interfaces.IState) error {
	return state.LeaderExecute(m)
}

// Returns true if this is a message for this server to execute as a follower
func (m *FactoidTransaction) Follower(state interfaces.IState) bool {
	return true
}

func (m *FactoidTransaction) FollowerExecute(state interfaces.IState) error {
	_, err := state.FollowerExecuteMsg(m)
	return err
}

func (m *FactoidTransaction) Process(dbheight uint32, state interfaces.IState) bool {
	if m.processed {
		return true
	}
	m.processed = true
	err := state.GetFactoidState().AddTransaction(1, m.Transaction)
	if err != nil {
		// Need to do something here if the transaction sent from the leader
		// does not validate.  Maybe the follower ignores, but leaders should fault
		// the offending leader...   For now we print and ignore.
		// TODO
		fmt.Println(err)
	}

	return true

}

func (m *FactoidTransaction) Int() int {
	return -1
}

func (m *FactoidTransaction) Bytes() []byte {
	return nil
}

func (m *FactoidTransaction) UnmarshalTransData(datax []byte) (newData []byte, err error) {
	newData =datax
	defer func() {
		return
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling Transaction Factoid: %v", r)
		}
	}()

	m.Transaction = new(factoid.Transaction)
	newData, err = m.Transaction.UnmarshalBinaryData(newData)

	return newData, err
}

func (m *FactoidTransaction) UnmarshalBinaryData(data []byte) (newData []byte, err error) {

	newData = data

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling Factoid: %v", r)
		}
	}()
	if newData[0] != m.Type() {
		return nil, fmt.Errorf("Invalid Message type")
	}
	newData = newData[1:]

	newData, err = m.Timestamp.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	m.Transaction = new(factoid.Transaction)
	newData, err = m.Transaction.UnmarshalBinaryData(newData)
	return newData, err
}

func (m *FactoidTransaction) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *FactoidTransaction) MarshalBinary() (data []byte, err error) {
	var buf primitives.Buffer
	buf.Write([]byte{m.Type()})

	if d, err := m.Timestamp.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}

	if d, err := m.Transaction.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}

	return buf.DeepCopyBytes(), nil
}

func (m *FactoidTransaction) String() string {
	return fmt.Sprintf("Factoid Transaction %x VM %d", m.GetHash().Bytes()[:3], m.VMIndex)
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
