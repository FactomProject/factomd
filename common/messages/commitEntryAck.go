// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
)

//General acknowledge message
type CommitEntryAck struct {
	Ack
}

var _ interfaces.IMsg = (*CommitEntryAck)(nil)

func (m *CommitEntryAck) Type() int {
	return constants.COMMIT_ENTRY_ACK_MSG
}

func (m *CommitEntryAck) Int() int {
	return -1
}

func (m *CommitEntryAck) String() string {
	return ""
}

func (m *CommitEntryAck) MarshalBinary() (data []byte, err error) {
	return MarshalAck(m)
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *CommitEntryAck) Validate(interfaces.IState) int {
	return 0
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *CommitEntryAck) Leader(state interfaces.IState) bool {
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
func (m *CommitEntryAck) LeaderExecute(state interfaces.IState) error {
	return nil
}

// Returns true if this is a message for this server to execute as a follower
func (m *CommitEntryAck) Follower(interfaces.IState) bool {
	return true
}

func (m *CommitEntryAck) FollowerExecute(interfaces.IState) error {
	return nil
}
