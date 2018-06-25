// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"encoding/binary"
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"

	"github.com/FactomProject/factomd/common/messages/msgbase"
	log "github.com/sirupsen/logrus"
)

var _ = log.Printf

// eLogger is for EOM Messages and extends packageLogger
var eLogger = packageLogger.WithFields(log.Fields{"message": "EOM"})

type EOM struct {
	msgbase.MessageBase
	Timestamp interfaces.Timestamp
	Minute    byte

	DBHeight  uint32
	SysHeight uint32
	SysHash   interfaces.IHash
	ChainID   interfaces.IHash
	Signature interfaces.IFullSignature
	FactoidVM bool

	//Not marshalled
	hash         interfaces.IHash
	MarkerSent   bool // If we have set EOM markers on blocks like Factoid blocks and such.
	marshalCache []byte
}

//var _ interfaces.IConfirmation = (*EOM)(nil)
var _ interfaces.Signable = (*EOM)(nil)
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

// Fix EOM hash to match and not have the sig so duplicates are not generated.
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

	// if this is a EOM for a saved block it's invalid (old)
	if m.DBHeight <= state.GetHighestSavedBlk() {
		return -1
	}
	// if this is a DBSig for a future block it's invalid (to far in the future)
	if m.DBHeight > state.GetHighestKnownBlock() { // (this may need to be +1?)
		return -1
	}

	found, _ := state.GetVirtualServers(m.DBHeight, int(m.Minute), m.ChainID)
	if !found { // Only EOM from federated servers are valid.
		return -1
	}

	// Check signature
	eomSigned, err := m.VerifySignature()
	if err != nil || !eomSigned {
		vlog := func(format string, args ...interface{}) {
			eLogger.WithFields(log.Fields{"func": "Validate", "lheight": state.GetLeaderHeight()}).WithFields(m.LogFields()).Errorf(format, args...)
		}

		if err != nil {
			vlog("[1] Failed to verify signature. Err: %s -- Msg: %s", err.Error(), m.String())
		}
		if !eomSigned {
			vlog("[1] Failed to verify, not signed. Msg: %s", m.String())
		}
		return -1
	}
	// if !eomSigned {
	// 	state.Logf("warning", "[EOM Validate (2)] Failed to verify signature. Msg: %s", err.Error(), m.String())
	// 	return -1
	// }
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
	signature, err := msgbase.SignSignable(m, key)
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
	return msgbase.VerifyMessage(m)
}

func (m *EOM) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling EOM message: %v", r)
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

	m.ChainID = primitives.NewHash(constants.ZERO_HASH)
	newData, err = m.ChainID.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	m.Minute, newData = newData[0], newData[1:]

	if m.Minute < 0 || m.Minute >= 10 {
		return nil, fmt.Errorf("Minute number is out of range")
	}

	m.VMIndex = int(newData[0])
	newData = newData[1:]
	m.FactoidVM = uint8(newData[0]) == 1
	newData = newData[1:]

	m.DBHeight, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]
	m.SysHeight, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]

	m.SysHash = primitives.NewHash(constants.ZERO_HASH)
	newData, err = m.SysHash.UnmarshalBinaryData(newData)

	b, newData := newData[0], newData[1:]
	if b > 0 {
		sig := new(primitives.Signature)
		newData, err = sig.UnmarshalBinaryData(newData)
		if err != nil {
			return nil, err
		}
		m.Signature = sig
	}

	m.marshalCache = append(m.marshalCache, data[:len(data)-len(newData)]...)

	return
}

func (m *EOM) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *EOM) MarshalForSignature() (data []byte, err error) {
	var buf primitives.Buffer
	buf.Write([]byte{m.Type()})
	if d, err := m.Timestamp.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}

	if d, err := m.ChainID.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}

	binary.Write(&buf, binary.BigEndian, m.Minute)
	binary.Write(&buf, binary.BigEndian, uint8(m.VMIndex))
	if m.FactoidVM {
		binary.Write(&buf, binary.BigEndian, uint8(1))
	} else {
		binary.Write(&buf, binary.BigEndian, uint8(0))
	}
	return buf.DeepCopyBytes(), nil
}

func (m *EOM) MarshalBinary() (data []byte, err error) {

	if m.marshalCache != nil {
		return m.marshalCache, nil
	}

	var buf primitives.Buffer
	resp, err := m.MarshalForSignature()
	if err != nil {
		return nil, err
	}
	buf.Write(resp)

	binary.Write(&buf, binary.BigEndian, m.DBHeight)
	binary.Write(&buf, binary.BigEndian, m.SysHeight)

	if m.SysHash == nil {
		m.SysHash = primitives.NewHash(constants.ZERO_HASH)
	}
	if d, err := m.SysHash.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}

	sig := m.GetSignature()
	if sig != nil {
		buf.WriteByte(1)
		sigBytes, err := sig.MarshalBinary()
		if err != nil {
			return nil, err
		}
		buf.Write(sigBytes)
	} else {
		buf.WriteByte(0)
	}

	m.marshalCache = buf.DeepCopyBytes()
	return m.marshalCache, nil
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
	return fmt.Sprintf("%6s-%30s FF %2d %1s-Leader[%x] hash[%x] %s",
		"EOM",
		fmt.Sprintf("DBh/VMh/h %d/%d/-- minute %d", m.DBHeight, m.VMIndex, m.Minute),
		m.SysHeight,
		f,
		m.ChainID.Bytes()[3:6],
		m.GetMsgHash().Bytes()[:3],
		local)
}

func (m *EOM) LogFields() log.Fields {
	return log.Fields{"category": "message", "messagetype": "eom", "dbheight": m.DBHeight, "vm": m.VMIndex,
		"minute": m.Minute, "chainid": m.ChainID.String(), "sysheight": m.SysHeight,
		"hash": m.GetMsgHash().String()}
}
