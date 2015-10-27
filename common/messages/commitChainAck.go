// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
)

//General acknowledge message
type CommitChainAck struct {
	Ack
}

var _ interfaces.IMsg = (*CommitChainAck)(nil)

func (m *CommitChainAck) Type() int {
	return constants.COMMIT_CHAIN_ACK_MSG
}

func (m *CommitChainAck) Int() int {
	return -1
}

func (m *CommitChainAck) String() string {
	return ""
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *CommitChainAck) Validate(interfaces.IState) int {
	return 0
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *CommitChainAck) Leader(state interfaces.IState) bool {
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
func (m *CommitChainAck) LeaderExecute(state interfaces.IState) error {
	return nil
}

// Returns true if this is a message for this server to execute as a follower
func (m *CommitChainAck) Follower(interfaces.IState) bool {
	return true
}

func (m *CommitChainAck) FollowerExecute(interfaces.IState) error {
	return nil
}
