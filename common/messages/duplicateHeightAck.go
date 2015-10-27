// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
)

//General acknowledge message
type DuplicateHeightAck struct {
	Ack
}

var _ interfaces.IMsg = (*DuplicateHeightAck)(nil)

func (m *DuplicateHeightAck) Type() int {
	return constants.DUPLICATE_HEIGHT_ACK_MSG
}

func (m *DuplicateHeightAck) Int() int {
	return -1
}

func (m *DuplicateHeightAck) String() string {
	return ""
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *DuplicateHeightAck) Validate(interfaces.IState) int {
	return 0
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *DuplicateHeightAck) Leader(state interfaces.IState) bool {
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
func (m *DuplicateHeightAck) LeaderExecute(state interfaces.IState) error {
	return nil
}

// Returns true if this is a message for this server to execute as a follower
func (m *DuplicateHeightAck) Follower(interfaces.IState) bool {
	return true
}

func (m *DuplicateHeightAck) FollowerExecute(interfaces.IState) error {
	return nil
}
