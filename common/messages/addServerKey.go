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

// Communicate a Admin Block Change
/*
	TYPE_MINUTE_NUM         uint8 = iota // 0
	TYPE_DB_SIGNATURE                    // 1
	TYPE_REVEAL_MATRYOSHKA               // 2
	TYPE_ADD_MATRYOSHKA                  // 3
	TYPE_ADD_SERVER_COUNT                // 4
	TYPE_ADD_FED_SERVER                  // 5
	TYPE_ADD_AUDIT_SERVER                // 6
	TYPE_REMOVE_FED_SERVER               // 7
	TYPE_ADD_FED_SERVER_KEY              // 8
	TYPE_ADD_BTC_ANCHOR_KEY              // 9
*/

type AddServerKeyMsg struct {
	MessageBase
	Timestamp        interfaces.Timestamp // Message Timestamp
	IdentityChainID  interfaces.IHash     // ChainID of new server
	AdminBlockChange byte
	KeyType          byte
	KeyPriority      byte
	Key              interfaces.IHash

	Signature interfaces.IFullSignature
}

var _ interfaces.IMsg = (*AddServerKeyMsg)(nil)
var _ Signable = (*AddServerKeyMsg)(nil)

func (m *AddServerKeyMsg) GetHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *AddServerKeyMsg) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
		data, err := m.MarshalForSignature()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

func (m *AddServerKeyMsg) Type() byte {
	return constants.ADDSERVER_KEY_MSG
}

func (m *AddServerKeyMsg) Int() int {
	return -1
}

func (m *AddServerKeyMsg) Bytes() []byte {
	return nil
}

func (m *AddServerKeyMsg) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp
}

func (m *AddServerKeyMsg) Validate(state interfaces.IState) int {
	return 1
	// TODO: Check Signiture

	// Check to see if identity exists and is audit or fed server
	if !state.VerifyIdentityAdminInfo(m.IdentityChainID) {
		return -1
	}

	// Should only be 20 bytes in the hash
	if m.AdminBlockChange == constants.TYPE_ADD_BTC_ANCHOR_KEY {
		for _, b := range m.Key.Bytes()[21:] {
			if b != 0 {
				return -1
			}
		}
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
func (m *AddServerKeyMsg) ComputeVMIndex(state interfaces.IState) {
	m.VMIndex = state.ComputeVMIndex(constants.ADMIN_CHAINID)
}

// Execute the leader functions of the given message
func (m *AddServerKeyMsg) LeaderExecute(state interfaces.IState) {
	state.LeaderExecute(m)
}

func (m *AddServerKeyMsg) FollowerExecute(state interfaces.IState) {
	state.FollowerExecuteMsg(m)
}

// Acknowledgements do not go into the process list.
func (e *AddServerKeyMsg) Process(dbheight uint32, state interfaces.IState) bool {
	return state.ProcessAddServerKey(dbheight, e)
}

func (e *AddServerKeyMsg) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *AddServerKeyMsg) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *AddServerKeyMsg) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (m *AddServerKeyMsg) Sign(key interfaces.Signer) error {
	signature, err := SignSignable(m, key)
	if err != nil {
		return err
	}
	m.Signature = signature
	return nil
}

func (m *AddServerKeyMsg) GetSignature() interfaces.IFullSignature {
	return m.Signature
}

func (m *AddServerKeyMsg) VerifySignature() (bool, error) {
	return VerifyMessage(m)
}

func (m *AddServerKeyMsg) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		return
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling Add Server Message: %v", r)
		}
	}()
	newData = data
	if newData[0] != m.Type() {
		return nil, fmt.Errorf("Invalid Message type")
	}
	newData = newData[1:]

	newData, err = m.Timestamp.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	m.IdentityChainID = new(primitives.Hash)
	newData, err = m.IdentityChainID.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	m.AdminBlockChange = newData[0]
	newData = newData[1:]

	m.KeyType = newData[0]
	newData = newData[1:]

	m.KeyPriority = newData[0]
	newData = newData[1:]

	m.Key = new(primitives.Hash)
	newData, err = m.Key.UnmarshalBinaryData(newData)
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

func (m *AddServerKeyMsg) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *AddServerKeyMsg) MarshalForSignature() ([]byte, error) {
	var buf primitives.Buffer

	binary.Write(&buf, binary.BigEndian, m.Type())

	t := m.GetTimestamp()
	data, err := t.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	data, err = m.IdentityChainID.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	binary.Write(&buf, binary.BigEndian, uint8(m.AdminBlockChange))
	binary.Write(&buf, binary.BigEndian, uint8(m.KeyType))
	binary.Write(&buf, binary.BigEndian, uint8(m.KeyPriority))

	data, err = m.Key.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	return buf.DeepCopyBytes(), nil
}

func (m *AddServerKeyMsg) MarshalBinary() ([]byte, error) {
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

func (m *AddServerKeyMsg) String() string {
	var mtype string
	if m.AdminBlockChange == constants.TYPE_ADD_MATRYOSHKA {
		mtype = "MHash"
	} else if m.AdminBlockChange == constants.TYPE_ADD_FED_SERVER_KEY {
		mtype = "Signing Key"
	} else if m.AdminBlockChange == constants.TYPE_ADD_BTC_ANCHOR_KEY {
		mtype = "BTC Key"
	} else {
		mtype = "other"
	}
	return fmt.Sprintf("AddServerKey (%s): ChainID: %x Time: %x  Key: %x Msg Hash %x ",
		mtype,
		m.IdentityChainID.Bytes()[:3],
		m.Timestamp,
		m.Key.Bytes()[:3],
		m.GetMsgHash().Bytes()[:3])

}

func (m *AddServerKeyMsg) IsSameAs(b *AddServerKeyMsg) bool {
	if b == nil {
		return false
	}
	if uint64(m.Timestamp) != uint64(b.Timestamp) {
		return false
	}
	if !m.IdentityChainID.IsSameAs(b.IdentityChainID) {
		return false
	}
	if m.AdminBlockChange != b.AdminBlockChange {
		return false
	}
	if m.KeyType != b.KeyType {
		return false
	}
	if m.KeyPriority != b.KeyPriority {
		return false
	}
	if !m.Key.IsSameAs(b.Key) {
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

func NewAddServerKeyMsg(state interfaces.IState, identityChain interfaces.IHash, adminChange byte, keyPriority byte, keyType byte, key interfaces.IHash) interfaces.IMsg {
	msg := new(AddServerKeyMsg)
	msg.IdentityChainID = identityChain
	msg.AdminBlockChange = adminChange
	msg.KeyType = keyType
	msg.KeyPriority = keyPriority
	msg.Key = key
	msg.Timestamp = state.GetTimestamp()

	return msg

}
