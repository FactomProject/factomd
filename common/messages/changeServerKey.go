// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"encoding/binary"
	"fmt"
	"os"
	"reflect"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"

	"github.com/FactomProject/factomd/common/messages/msgbase"
	log "github.com/sirupsen/logrus"
)

// Communicate a Admin Block Change

type ChangeServerKeyMsg struct {
	msgbase.MessageBase
	Timestamp        interfaces.Timestamp // Message Timestamp
	IdentityChainID  interfaces.IHash     // ChainID of new server
	AdminBlockChange byte
	KeyType          byte
	KeyPriority      byte
	Key              interfaces.IHash

	Signature interfaces.IFullSignature
}

var _ interfaces.IMsg = (*ChangeServerKeyMsg)(nil)
var _ interfaces.Signable = (*ChangeServerKeyMsg)(nil)

func (m *ChangeServerKeyMsg) GetRepeatHash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("ChangeServerKeyMsg.GetRepeatHash() saw an interface that was nil")
		}
	}()

	return m.GetMsgHash()
}

func (m *ChangeServerKeyMsg) GetHash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("ChangeServerKeyMsg.GetHash() saw an interface that was nil")
		}
	}()

	return m.GetMsgHash()
}

func (m *ChangeServerKeyMsg) GetMsgHash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("ChangeServerKeyMsg.GetMsgHash() saw an interface that was nil")
		}
	}()

	if m.MsgHash == nil {
		data, err := m.MarshalForSignature()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

func (m *ChangeServerKeyMsg) Type() byte {
	return constants.CHANGESERVER_KEY_MSG
}

func (m *ChangeServerKeyMsg) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp.Clone()
}

func (m *ChangeServerKeyMsg) Validate(state interfaces.IState) int {
	// Check to see if identity exists and is audit or fed server
	if !state.VerifyIsAuthority(m.IdentityChainID) {
		fmt.Println("ChangeServerKey Error. Server is not an authority")
		return -1
	}

	// Should only be 20 bytes in the hash if btc key add
	if m.AdminBlockChange == constants.TYPE_ADD_BTC_ANCHOR_KEY {
		for _, b := range m.Key.Bytes()[21:] {
			if b != 0 {
				fmt.Println("ChangeServerKey Error. Newkey is invalid length")
				return -1
			}
		}
	}

	// Check signatures
	bytes, err := m.MarshalForSignature()
	if err != nil || m.Signature == nil {
		return -1
	}
	authSigned, err := state.FastVerifyAuthoritySignature(bytes, m.Signature, state.GetLeaderHeight())
	if err != nil || authSigned != 1 { // authSigned = 1 for fed signed
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
func (m *ChangeServerKeyMsg) ComputeVMIndex(state interfaces.IState) {
	m.VMIndex = state.ComputeVMIndex(constants.ADMIN_CHAINID)
}

// Execute the leader functions of the given message
func (m *ChangeServerKeyMsg) LeaderExecute(state interfaces.IState) {
	state.LeaderExecute(m)
}

func (m *ChangeServerKeyMsg) FollowerExecute(state interfaces.IState) {
	state.FollowerExecuteMsg(m)
}

// Acknowledgements do not go into the process list.
func (e *ChangeServerKeyMsg) Process(dbheight uint32, state interfaces.IState) bool {
	return state.ProcessChangeServerKey(dbheight, e)
}

func (e *ChangeServerKeyMsg) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *ChangeServerKeyMsg) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (m *ChangeServerKeyMsg) Sign(key interfaces.Signer) error {
	signature, err := msgbase.SignSignable(m, key)
	if err != nil {
		return err
	}
	m.Signature = signature
	return nil
}

func (m *ChangeServerKeyMsg) GetSignature() interfaces.IFullSignature {
	return m.Signature
}

func (m *ChangeServerKeyMsg) VerifySignature() (bool, error) {
	return msgbase.VerifyMessage(m)
}

func (m *ChangeServerKeyMsg) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling Add Server Message: %v", r)
		}
	}()
	newData = data
	if newData[0] != m.Type() {
		return nil, fmt.Errorf("Invalid Message type")
	}
	newData = newData[1:]

	m.Timestamp = new(primitives.Timestamp)
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

func (m *ChangeServerKeyMsg) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *ChangeServerKeyMsg) MarshalForSignature() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "ChangeServerKeyMsg.MarshalForSignature err:%v", *pe)
		}
	}(&err)
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

func (m *ChangeServerKeyMsg) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "ChangeServerKeyMsg.MarshalBinary err:%v", *pe)
		}
	}(&err)
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

func (m *ChangeServerKeyMsg) String() string {
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
	return fmt.Sprintf("ChangeServerKey (%s): ChainID: %x Time: %x  Key: %x Msg Hash %x ",
		mtype,
		m.IdentityChainID.Bytes()[:3],
		&m.Timestamp,
		m.Key.Bytes()[:3],
		m.GetMsgHash().Bytes()[:3])

}

func (m *ChangeServerKeyMsg) LogFields() log.Fields {
	return log.Fields{"category": "message", "messagetype": "changeserverkey",
		"server": m.IdentityChainID.String(), "hash": m.GetHash().String()}
}

func (m *ChangeServerKeyMsg) IsSameAs(b *ChangeServerKeyMsg) bool {
	if b == nil {
		return false
	}
	if m.Timestamp.GetTimeMilli() != b.Timestamp.GetTimeMilli() {
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

func NewChangeServerKeyMsg(state interfaces.IState, identityChain interfaces.IHash, adminChange byte, keyPriority byte, keyType byte, key interfaces.IHash) interfaces.IMsg {
	msg := new(ChangeServerKeyMsg)
	msg.IdentityChainID = identityChain
	msg.AdminBlockChange = adminChange
	msg.KeyType = keyType
	msg.KeyPriority = keyPriority
	msg.Key = key
	msg.Timestamp = state.GetTimestamp()

	return msg

}
