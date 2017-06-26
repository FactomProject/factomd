// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/log"
)

var _ = log.Printf

type EOM struct {
	MessageBase
	Timestamp interfaces.Timestamp
	Minute    byte

	DBHeight  uint32
	SysHeight uint32
	SysHash   interfaces.IHash
	ChainID   interfaces.IHash
	Signature interfaces.IFullSignature
	FactoidVM bool

	//Not marshalled
	hash       interfaces.IHash
	MarkerSent bool // If we have set EOM markers on blocks like Factoid blocks and such.
}

//var _ interfaces.IConfirmation = (*EOM)(nil)
var _ Signable = (*EOM)(nil)
var _ interfaces.IMsg = (*EOM)(nil)

func (a *EOM) IsSameAs(b *EOM) bool {
	if b == nil {
		return false
	}
	if a.Timestamp.GetTimeMilli() != b.Timestamp.GetTimeMilli() {
		return false
	}
	if a.Minute != b.Minute {
		return false
	}
	if a.DBHeight != b.DBHeight {
		return false
	}

	if a.ChainID == nil && b.ChainID != nil {
		return false
	}
	if a.ChainID != nil {
		if a.ChainID.IsSameAs(b.ChainID) == false {
			return false
		}
	}

	if a.Signature == nil && b.Signature != nil {
		return false
	}
	if a.Signature != nil {
		if a.Signature.IsSameAs(b.Signature) == false {
			return false
		}
	}

	return true
}

func (e *EOM) Process(dbheight uint32, state interfaces.IState) bool {
	return state.ProcessEOM(dbheight, e)
}

func (m *EOM) GetRepeatHash() interfaces.IHash {
	if m.RepeatHash == nil {
		data, err := m.MarshalBinary()
		if err != nil {
			return nil
		}
		m.RepeatHash = primitives.Sha(data)
	}
	return m.RepeatHash
}

func (m *EOM) GetHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *EOM) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
		data, err := m.MarshalForSignature()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

func (m *EOM) GetTimestamp() interfaces.Timestamp {
	if m.Timestamp == nil {
		m.Timestamp = new(primitives.Timestamp)
	}
	return m.Timestamp
}

func (m *EOM) Type() byte {
	return constants.EOM_MSG
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *EOM) Validate(state interfaces.IState) int {
	if m.IsLocal() {
		return 1
	}

	// Ignore old EOM
	if m.DBHeight <= state.GetHighestSavedBlk() {
		return -1
	}

	found, _ := state.GetVirtualServers(m.DBHeight, int(m.Minute), m.ChainID)
	if !found { // Only EOM from federated servers are valid.
		return -1
	}

	// Check signature
	eomSigned, err := m.VerifySignature()
	if err != nil {
		return -1
	}
	if !eomSigned {
		return -1
	}
	return 1
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *EOM) ComputeVMIndex(state interfaces.IState) {
}

// Execute the leader functions of the given message
func (m *EOM) LeaderExecute(state interfaces.IState) {
	state.LeaderExecuteEOM(m)
}

func (m *EOM) FollowerExecute(state interfaces.IState) {
	state.FollowerExecuteEOM(m)
}

func (e *EOM) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *EOM) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (m *EOM) Sign(key interfaces.Signer) error {
	signature, err := SignSignable(m, key)
	if err != nil {
		return err
	}
	m.Signature = signature
	return nil
}

func (m *EOM) GetSignature() interfaces.IFullSignature {
	return m.Signature
}

func (m *EOM) VerifySignature() (bool, error) {
	return VerifyMessage(m)
}

func (m *EOM) UnmarshalBinaryData(data []byte) ([]byte, error) {
	buf := primitives.NewBuffer(data)

	t, err := buf.PopByte()
	if err != nil {
		return nil, err
	}
	if t != m.Type() {
		return nil, fmt.Errorf("Invalid Message type")
	}

	m.Timestamp = new(primitives.Timestamp)
	err = buf.PopBinaryMarshallable(m.Timestamp)
	if err != nil {
		return nil, err
	}

	m.ChainID = primitives.NewZeroHash()
	err = buf.PopBinaryMarshallable(m.ChainID)
	if err != nil {
		return nil, err
	}

	m.Minute, err = buf.PopByte()
	if err != nil {
		return nil, err
	}
	if m.Minute < 0 || m.Minute >= 10 {
		return nil, fmt.Errorf("Minute number is out of range")
	}

	t, err = buf.PopByte()
	if err != nil {
		return nil, err
	}
	m.VMIndex = int(t)
	m.FactoidVM, err = buf.PopBool()
	if err != nil {
		return nil, err
	}

	m.DBHeight, err = buf.PopUInt32()
	if err != nil {
		return nil, err
	}
	m.SysHeight, err = buf.PopUInt32()
	if err != nil {
		return nil, err
	}

	m.SysHash = primitives.NewZeroHash()
	err = buf.PopBinaryMarshallable(m.SysHash)
	if err != nil {
		return nil, err
	}

	if buf.Len() > 0 {
		m.Signature = new(primitives.Signature)
		err = buf.PopBinaryMarshallable(m.Signature)
		if err != nil {
			return nil, err
		}
	}

	return buf.DeepCopyBytes(), nil
}

func (m *EOM) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *EOM) MarshalForSignature() ([]byte, error) {
	buf := primitives.NewBuffer(nil)

	err := buf.PushByte(m.Type())
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(m.Timestamp)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(m.ChainID)
	if err != nil {
		return nil, err
	}
	err = buf.PushByte(m.Minute)
	if err != nil {
		return nil, err
	}
	err = buf.PushByte(byte(m.VMIndex))
	if err != nil {
		return nil, err
	}

	err = buf.PushBool(m.FactoidVM)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

func (m *EOM) MarshalBinary() ([]byte, error) {
	h, err := m.MarshalForSignature()
	if err != nil {
		return nil, err
	}
	buf := primitives.NewBuffer(h)

	err = buf.PushUInt32(m.DBHeight)
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(m.SysHeight)
	if err != nil {
		return nil, err
	}

	if m.SysHash == nil {
		m.SysHash = primitives.NewZeroHash()
	}
	err = buf.PushBinaryMarshallable(m.SysHash)
	if err != nil {
		return nil, err
	}

	sig := m.GetSignature()
	if sig != nil {
		err = buf.PushBinaryMarshallable(sig)
		if err != nil {
			return nil, err
		}
	}

	return buf.DeepCopyBytes(), nil
}

func (m *EOM) String() string {
	local := ""
	if m.IsLocal() {
		local = "local"
	}
	f := "-"
	if m.FactoidVM {
		f = "F"
	}
	return fmt.Sprintf("%6s-VM%3d: Min:%4d DBHt:%5d FF %2d -%1s-Leader[%x] hash[%x] %s",
		"EOM",
		m.VMIndex,
		m.Minute,
		m.DBHeight,
		m.SysHeight,
		f,
		m.ChainID.Bytes()[:4],
		m.GetMsgHash().Bytes()[:3],
		local)
}
