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
	Timestamp   interfaces.Timestamp
	CommitChain *entryCreditBlock.CommitChain

	// Not marshaled... Just used by the leader
	count int
}

var _ interfaces.IMsg = (*CommitChainMsg)(nil)
var _ interfaces.ICounted = (*CommitChainMsg)(nil)

func (m *CommitChainMsg) GetCount() int {
	return m.count
}

func (m *CommitChainMsg) IncCount() {
	m.count += 1
}

func (m *CommitChainMsg) SetCount(cnt int) {
	m.count = cnt
}

func (m *CommitChainMsg) Process(dbheight uint32, state interfaces.IState) {
	ecblk := state.GetEntryCreditBlock(dbheight)
	ecbody := ecblk.GetBody()
	ecbody.AddEntry(m.CommitChain)
	state.GetFactoidState(dbheight).UpdateECTransaction(m.CommitChain)
	state.PutCommits(dbheight, m.GetHash(), m)
}

func (m *CommitChainMsg) GetHash() interfaces.IHash {
	return m.CommitChain.EntryHash
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
func (m *CommitChainMsg) Validate(dbheight uint32, state interfaces.IState) int {
	if !m.CommitChain.IsValid() {
		return -1
	}
	ebal := state.GetFactoidState(dbheight).GetECBalance(*m.CommitChain.ECPubKey)
	if int(m.CommitChain.Credits) > int(ebal) {
		fmt.Println("Not enough Credits")
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
	v := m.Validate(state.GetDBHeight(), state)
	if v <= 0 {
		return fmt.Errorf("Commit Chain no longer valid")
	}
	b := m.GetHash()

	ack, err := NewAck(state, b)
	state.PutCommits(state.GetDBHeight(), m.GetHash(), m)
	if err != nil {
		return err
	}

	state.NetworkOutMsgQueue() <- ack
	state.FollowerInMsgQueue() <- ack // Send the Ack to follower
	state.FollowerInMsgQueue() <- m   // Send factoid trans to follower

	return nil
}

// Returns true if this is a message for this server to execute as a follower
func (m *CommitChainMsg) Follower(state interfaces.IState) bool {
	return true
}

func (m *CommitChainMsg) FollowerExecute(state interfaces.IState) error {
	matched, err := state.MatchAckFollowerExecute(m)
	if err != nil {
		return err
	}
	if matched { // We matched, we must be remembered!
		state.PutCommits(state.GetDBHeight(), m.GetHash(), m)
	}
	return nil
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
	return "CommitChainMsg " + m.Timestamp.String() + " " + m.GetHash().String()
}
