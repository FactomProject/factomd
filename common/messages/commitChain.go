// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"bytes"
	"fmt"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

//A placeholder structure for messages
type CommitChainMsg struct {
	CommitChain *entryCreditBlock.CommitChain
	Timestamp   interfaces.Timestamp

	//Not marshalled
	hash interfaces.IHash
}

var _ interfaces.IMsg = (*CommitChainMsg)(nil)

func (m *CommitChainMsg) Process(state interfaces.IState) {
	ecblk := state.GetCurrentEntryCreditBlock()
	ecbody := ecblk.GetBody()
	ecbody.AddEntry(m.CommitChain)
	state.GetFactoidState().UpdateECTransaction(m.CommitChain)
}

func (m *CommitChainMsg) GetHash() interfaces.IHash {
	if m.hash == nil {
		data, err := m.CommitChain.MarshalBinary()
		if err != nil {
			panic(fmt.Sprintf("Error in CommitChain.GetHash(): %s", err.Error()))
		}
		m.hash = primitives.Sha(data)
	}
	return m.hash
}

func (m *CommitChainMsg) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp
}

func (m *CommitChainMsg) Type() int {
	return constants.COMMIT_CHAIN_MSG
}

func (m *CommitChainMsg) Int() int {
	return -1
}

func (m *CommitChainMsg) Bytes() []byte {
	return nil
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *CommitChainMsg) Validate(state interfaces.IState) int {
	if !m.CommitChain.IsValid() {
		return -1
	}
	ebal := state.GetFactoidState().GetECBalance(*m.CommitChain.ECPubKey)
	if int(m.CommitChain.Credits) < int(ebal) {
		return 0
	}
	return 1

}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *CommitChainMsg) Leader(state interfaces.IState) bool {
	return state.LeaderFor(constants.EC_CHAINID)
}

// Execute the leader functions of the given message
func (m *CommitChainMsg) LeaderExecute(state interfaces.IState) error {
	v := m.Validate(state)
	if v <= 0 {
		return fmt.Errorf("Commit Chain no longer valid")
	}
	b := m.CommitChain.Hash().Bytes()

	msg, err := NewAck(state, b)

	if err != nil {
		return err
	}

	state.NetworkOutMsgQueue() <- msg
	state.FollowerInMsgQueue() <- m   // Send factoid trans to follower
	state.FollowerInMsgQueue() <- msg // Send the Ack to follower

	return nil
}

// Returns true if this is a message for this server to execute as a follower
func (m *CommitChainMsg) Follower(state interfaces.IState) bool {
	return true
}

func (m *CommitChainMsg) FollowerExecute(state interfaces.IState) error {
	return state.MatchAckFollowerExecute(m) 
}

func (e *CommitChainMsg) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *CommitChainMsg) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *CommitChainMsg) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (m *CommitChainMsg) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	newData = data[1:]
	cc := entryCreditBlock.NewCommitChain()
	newData, err = cc.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}
	m.CommitChain = cc
	return newData, nil
}

func (m *CommitChainMsg) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *CommitChainMsg) MarshalBinary() (data []byte, err error) {
	data, err = m.CommitChain.MarshalBinary()
	if err != nil {
		return nil, err
	}
	data = append([]byte{byte(m.Type())}, data...)
	return data, nil
}

func (m *CommitChainMsg) String() string {
	return ""
}
