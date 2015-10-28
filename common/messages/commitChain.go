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
type CommitChainMsg struct {
	//ChainID          interfaces.IHash
	ChainIDEntryHash interfaces.IHash
	//EntryHash        interfaces.IHash
}

var _ interfaces.IMsg = (*CommitChainMsg)(nil)

func (m *CommitChainMsg) Type() int {
	return constants.COMMIT_CHAIN_MSG
}

func (m *CommitChainMsg) Int() int {
	return -1
}

func (m *CommitChainMsg) Bytes() []byte {
	return nil
}

func (m *CommitChainMsg) UnmarshalBinaryData(data []byte) (newdata []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()

	return nil, nil
}

func (m *CommitChainMsg) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *CommitChainMsg) MarshalBinary() (data []byte, err error) {
	return nil, nil
}

func (m *CommitChainMsg) String() string {
	return ""
}

func (m *CommitChainMsg) DBHeight() int {
	return 0
}

func (m *CommitChainMsg) ChainID() []byte {
	return nil
}

func (m *CommitChainMsg) ListHeight() int {
	return 0
}

func (m *CommitChainMsg) SerialHash() []byte {
	return nil
}

func (m *CommitChainMsg) Signature() []byte {
	return nil
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *CommitChainMsg) Validate(interfaces.IState) int {
	return 0
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *CommitChainMsg) Leader(state interfaces.IState) bool {
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
func (m *CommitChainMsg) LeaderExecute(state interfaces.IState) error {
	return nil
}

// Returns true if this is a message for this server to execute as a follower
func (m *CommitChainMsg) Follower(interfaces.IState) bool {
	return true
}

func (m *CommitChainMsg) FollowerExecute(interfaces.IState) error {
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
