// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// Communicate a Admin Block Change

type ChangeServerKeyMsg struct {
	MessageBase
	Timestamp        interfaces.Timestamp // Message Timestamp
	IdentityChainID  interfaces.IHash     // ChainID of new server
	AdminBlockChange byte
	KeyType          byte
	KeyPriority      byte
	Key              interfaces.IHash

	Signature interfaces.IFullSignature
}

var _ interfaces.IMsg = (*ChangeServerKeyMsg)(nil)
var _ Signable = (*ChangeServerKeyMsg)(nil)

func (m *ChangeServerKeyMsg) GetRepeatHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *ChangeServerKeyMsg) GetHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *ChangeServerKeyMsg) GetMsgHash() interfaces.IHash {
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
	return m.Timestamp
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
	sig := m.Signature.GetSignature()
	authSigned, err := state.VerifyAuthoritySignature(bytes, sig, state.GetLeaderHeight())
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
	signature, err := SignSignable(m, key)
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
	return VerifyMessage(m)
}

func (m *ChangeServerKeyMsg) UnmarshalBinaryData(data []byte) ([]byte, error) {
	buf := primitives.NewBuffer(data)
	t, err := buf.PopByte()
	if t != m.Type() {
		return nil, fmt.Errorf("Invalid Message type")
	}

	m.Timestamp = new(primitives.Timestamp)
	err = buf.PopBinaryMarshallable(m.Timestamp)
	if err != nil {
		return nil, err
	}

	m.IdentityChainID = primitives.NewZeroHash()
	err = buf.PopBinaryMarshallable(m.IdentityChainID)
	if err != nil {
		return nil, err
	}

	m.AdminBlockChange, err = buf.PopByte()
	if err != nil {
		return nil, err
	}
	m.KeyType, err = buf.PopByte()
	if err != nil {
		return nil, err
	}
	m.KeyPriority, err = buf.PopByte()
	if err != nil {
		return nil, err
	}

	m.Key = primitives.NewZeroHash()
	err = buf.PopBinaryMarshallable(m.Key)
	if err != nil {
		return nil, err
	}

	if buf.Len() > 32 {
		m.Signature = new(primitives.Signature)
		err = buf.PopBinaryMarshallable(m.Signature)
		if err != nil {
			return nil, err
		}
	}

	return buf.DeepCopyBytes(), nil
}

func (m *ChangeServerKeyMsg) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *ChangeServerKeyMsg) MarshalForSignature() ([]byte, error) {
	buf := primitives.NewBuffer(nil)

	err := buf.PushByte(m.Type())
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(m.GetTimestamp())
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(m.IdentityChainID)
	if err != nil {
		return nil, err
	}

	err = buf.PushByte(m.AdminBlockChange)
	if err != nil {
		return nil, err
	}
	err = buf.PushByte(m.KeyType)
	if err != nil {
		return nil, err
	}
	err = buf.PushByte(m.KeyPriority)
	if err != nil {
		return nil, err
	}

	err = buf.PushBinaryMarshallable(m.Key)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

func (m *ChangeServerKeyMsg) MarshalBinary() ([]byte, error) {
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
