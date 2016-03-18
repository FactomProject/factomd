// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// Communicate a Directory Block State

type AddServerMsg struct {
	MessageBase
	Timestamp     interfaces.Timestamp // Message Timestamp
	ServerChainID interfaces.IHash     // ChainID of new server
}

var _ interfaces.IMsg = (*AddServerMsg)(nil)

func (m *AddServerMsg) IsSameAs(b *AddServerMsg) bool {
	return true
}

func (m *AddServerMsg) GetHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *AddServerMsg) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
		data, err := m.MarshalBinary()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

func (m *AddServerMsg) Type() int {
	return constants.ADDSERVER_MSG
}

func (m *AddServerMsg) Int() int {
	return -1
}

func (m *AddServerMsg) Bytes() []byte {
	return nil
}

func (m *AddServerMsg) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp
}

// Validate the message, TBD
func (m *AddServerMsg) Validate(state interfaces.IState) int {
	return 1
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *AddServerMsg) Leader(state interfaces.IState) bool {
	return true
}

// Execute the leader functions of the given message
func (m *AddServerMsg) LeaderExecute(state interfaces.IState) error {
	return state.LeaderExecuteAddServer(m)
}

// Returns true if this is a message for this server to execute as a follower
func (m *AddServerMsg) Follower(interfaces.IState) bool {
	return true
}

func (m *AddServerMsg) FollowerExecute(state interfaces.IState) error {
	_, err := state.FollowerExecuteMsg(m)
	return err
}

// Acknowledgements do not go into the process list.
func (e *AddServerMsg) Process(dbheight uint32, state interfaces.IState) {
	state.ProcessAddServer(dbheight, e)
}

func (e *AddServerMsg) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *AddServerMsg) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *AddServerMsg) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (m *AddServerMsg) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()

	newData = data[1:] // Skip our type;  Someone else's problem.

	newData, err = m.Timestamp.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	newData, err = m.ServerChainID.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	return
}

func (m *AddServerMsg) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *AddServerMsg) MarshalForSignature() ([]byte, error) {

	var buf bytes.Buffer

	binary.Write(&buf, binary.BigEndian, byte(m.Type()))

	t := m.GetTimestamp()
	data, err := t.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	data, err = m.ServerChainID.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	return buf.Bytes(), nil
}

func (m *AddServerMsg) MarshalBinary() ([]byte, error) {
	return m.MarshalForSignature()
}

func (m *AddServerMsg) String() string {
	return fmt.Sprintf("AddServer: ChainID: %s", m.ServerChainID.String())
}

func NewAddServerMsg(state interfaces.IState) interfaces.IMsg {

	msg := new(AddServerMsg)
	msg.ServerChainID = state.GetIdentityChainID()
	msg.Timestamp = state.GetTimestamp()

	return msg
}
