// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
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

	Signature interfaces.IFullSignature
}

var _ interfaces.IMsg = (*AddServerMsg)(nil)

func (m *AddServerMsg) IsSameAs(b *AddServerMsg) bool {
	if uint64(m.Timestamp) != uint64(b.Timestamp) {
		return false
	}
	if !m.ServerChainID.IsSameAs(b.ServerChainID) {
		return false
	}
	if m.Signature == nil && b.Signature != nil {
		return false
	}
	if m.Signature != nil {
		if m.Signature.IsSameAs(b.Signature) == false {
			return false
		}
	}
	return true
}

func (m *AddServerMsg) GetHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *AddServerMsg) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
		data, err := m.MarshalForSignature()
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

func (m *AddServerMsg) Validate(state interfaces.IState) int {
	return 1 // Not going to check right now.
	authoritativeKey, _ := hex.DecodeString("cc1985cdfae4e32b5a454dfda8ce5e1361558482684f3367649c3ad852c8e31a")

	if m.GetSignature() == nil || bytes.Compare(m.GetSignature().GetKey(), authoritativeKey) != 0 {
		// the message was not signed with the proper authoritative signing key (from conf file)
		// it is therefore considered invalid
		return -1
	}

	isVer, err := m.VerifySignature()
	if err != nil || !isVer {
		// if there is an error during signature verification
		// or if the signature is invalid
		// the message is considered invalid
		return -1
	}

	return 1
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *AddServerMsg) Leader(state interfaces.IState) bool {
	return state.LeaderFor(m, constants.ADMIN_CHAINID)
}

// Execute the leader functions of the given message
func (m *AddServerMsg) LeaderExecute(state interfaces.IState) error {
	return state.LeaderExecute(m)
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
func (e *AddServerMsg) Process(dbheight uint32, state interfaces.IState) bool {
	if state.GetOut() == true {
		state.Println("Processing to add a Server: ", dbheight)
	}
	return state.ProcessAddServer(dbheight, e)
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

func (m *AddServerMsg) Sign(key interfaces.Signer) error {
	signature, err := SignSignable(m, key)
	if err != nil {
		return err
	}
	m.Signature = signature
	return nil
}

func (m *AddServerMsg) GetSignature() interfaces.IFullSignature {
	return m.Signature
}

func (m *AddServerMsg) VerifySignature() (bool, error) {
	return VerifyMessage(m)
}

func (m *AddServerMsg) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		return
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling Add Server Message: %v", r)
		}
	}()

	newData = data[1:] // Skip our type;  Someone else's problem.

	newData, err = m.Timestamp.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	m.ServerChainID = new(primitives.Hash)
	newData, err = m.ServerChainID.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	if len(newData) > 32 {
		m.Signature = new(primitives.Signature)
		newData, err = m.Signature.UnmarshalBinaryData(newData)
		if err != nil {
			return nil, err
		}
	}
	return
}

func (m *AddServerMsg) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *AddServerMsg) MarshalForSignature() ([]byte, error) {
	var buf primitives.Buffer

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

	return buf.DeepCopyBytes(), nil
}

func (m *AddServerMsg) MarshalBinary() ([]byte, error) {
	var buf primitives.Buffer

	data, err := m.MarshalForSignature()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	if m.Signature != nil {
		data, err = m.Signature.MarshalBinary()
		if err != nil {
			return nil, err
		}
		buf.Write(data)
	}

	return buf.DeepCopyBytes(), nil
}

func (m *AddServerMsg) String() string {
	return fmt.Sprintf("AddServer: ChainID: %s Time: %v", m.ServerChainID.String(), m.Timestamp)
}

func NewAddServerMsg(state interfaces.IState) interfaces.IMsg {
	msg := new(AddServerMsg)
	msg.ServerChainID = state.GetIdentityChainID()
	msg.Timestamp = state.GetTimestamp()

	return msg

}
