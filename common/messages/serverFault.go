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

//A placeholder structure for messages
type ServerFault struct {
	msgbase.MessageBase

	// The following 5 fields represent the "Core" of the message
	// This should match the Core of FullServerFault messages
	ServerID      interfaces.IHash
	AuditServerID interfaces.IHash
	VMIndex       byte
	DBHeight      uint32
	Height        uint32
	SystemHeight  uint32
	Timestamp     interfaces.Timestamp

	Signature interfaces.IFullSignature

	//Not marshalled
	hash interfaces.IHash
}

var _ interfaces.IMsg = (*ServerFault)(nil)
var _ interfaces.Signable = (*ServerFault)(nil)

func (m *ServerFault) Process(uint32, interfaces.IState) bool { return true }

func (m *ServerFault) GetRepeatHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *ServerFault) GetHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *ServerFault) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
		data, err := m.MarshalBinary()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

func (m *ServerFault) GetCoreHash() interfaces.IHash {
	data, err := m.MarshalForSignature()
	if err != nil {
		return nil
	}
	return primitives.Sha(data)
}

func (m *ServerFault) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp
}

func (m *ServerFault) Type() byte {
	return constants.FED_SERVER_FAULT_MSG
}

func (m *ServerFault) MarshalForSignature() (data []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error marshalling Server Fault Core: %v", r)
		}
	}()

	var buf primitives.Buffer

	if d, err := m.ServerID.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}
	if d, err := m.AuditServerID.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}

	buf.WriteByte(m.VMIndex)
	binary.Write(&buf, binary.BigEndian, uint32(m.DBHeight))
	binary.Write(&buf, binary.BigEndian, uint32(m.Height))
	binary.Write(&buf, binary.BigEndian, uint32(m.SystemHeight))

	if d, err := m.Timestamp.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}

	return buf.DeepCopyBytes(), nil
}

func (m *ServerFault) PreMarshalBinary() (data []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error marshalling Invalid Server Fault: %v", r)
		}
	}()

	var buf primitives.Buffer

	buf.Write([]byte{m.Type()})
	if d, err := m.ServerID.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}
	if d, err := m.AuditServerID.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}

	buf.WriteByte(m.VMIndex)
	binary.Write(&buf, binary.BigEndian, uint32(m.DBHeight))
	binary.Write(&buf, binary.BigEndian, uint32(m.Height))
	binary.Write(&buf, binary.BigEndian, uint32(m.SystemHeight))
	if d, err := m.Timestamp.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}

	return buf.DeepCopyBytes(), nil
}

func (m *ServerFault) MarshalBinary() (data []byte, err error) {
	resp, err := m.PreMarshalBinary()
	if err != nil {
		return nil, err
	}
	sig := m.GetSignature()

	if sig != nil {
		sigBytes, err := sig.MarshalBinary()
		if err != nil {
			return nil, err
		}
		return append(resp, sigBytes...), nil
	}

	return resp, nil
}

func (m *ServerFault) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling With Signatures Invalid Server Fault: %v", r)
		}
	}()
	newData = data
	if newData[0] != m.Type() {
		return nil, fmt.Errorf("Invalid Message type")
	}
	newData = newData[1:]

	if m.ServerID == nil {
		m.ServerID = primitives.NewZeroHash()
	}
	newData, err = m.ServerID.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}
	if m.AuditServerID == nil {
		m.AuditServerID = primitives.NewZeroHash()
	}
	newData, err = m.AuditServerID.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	m.VMIndex, newData = newData[0], newData[1:]
	m.DBHeight, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]
	m.Height, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]
	m.SystemHeight, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]

	m.Timestamp = new(primitives.Timestamp)
	newData, err = m.Timestamp.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	if len(newData) > 0 {
		m.Signature = new(primitives.Signature)
		newData, err = m.Signature.UnmarshalBinaryData(newData)
		if err != nil {
			return nil, err
		}
	}

	return newData, nil
}

func (m *ServerFault) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *ServerFault) GetSignature() interfaces.IFullSignature {
	return m.Signature
}

func (m *ServerFault) VerifySignature() (bool, error) {
	return msgbase.VerifyMessage(m)
}

func (m *ServerFault) Sign(key interfaces.Signer) error {
	signature, err := msgbase.SignSignable(m, key)
	if err != nil {
		return err
	}
	m.Signature = signature
	return nil
}

func (m *ServerFault) String() string {
	var sig [3]byte

	if m.Signature != nil {
		copy(sig[:], m.Signature.Bytes()[:3])
	}

	return fmt.Sprintf("%6s %v VM%3d: (%v) AuditID: %v PL:%5d DBHt:%5d SysHt:%3d sig[:3]=%x hash[:3]=%x",
		"SFault",
		m.GetCoreHash().String()[:10],
		m.VMIndex,
		m.ServerID.String()[:10],
		m.AuditServerID.String()[:10],
		m.Height,
		m.DBHeight,
		m.SystemHeight,
		sig[:3],
		m.GetHash().Bytes()[:3])
}

func (m *ServerFault) LogFields() log.Fields {
	return log.Fields{"category": "message", "messagetype": "serverfault",
		"vm":        m.VMIndex,
		"dbheight":  m.DBHeight,
		"leaderid":  m.ServerID.String(),
		"auditid":   m.AuditServerID.String(),
		"sysheight": m.SystemHeight,
		"signature": string(m.Signature.Bytes()),
		"corehash":  m.GetCoreHash().String(),
		"hash":      m.GetHash().String()}
}

func (m *ServerFault) GetDBHeight() uint32 {
	return m.DBHeight
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *ServerFault) Validate(state interfaces.IState) int {
	if m.Signature == nil {
		return -1
	}
	if m.ServerID == nil || m.ServerID.IsZero() {
		return -1
	}
	if m.AuditServerID == nil || m.AuditServerID.IsZero() {
		return -1
	}

	return 1
}

func (m *ServerFault) ComputeVMIndex(state interfaces.IState) {
}

// Execute the leader functions of the given message
func (m *ServerFault) LeaderExecute(state interfaces.IState) {
	m.FollowerExecute(state)
}

func (m *ServerFault) FollowerExecute(state interfaces.IState) {
	state.FollowerExecuteSFault(m)
}

func (e *ServerFault) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *ServerFault) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (a *ServerFault) IsSameAs(b *ServerFault) bool {
	if b == nil {
		return false
	}
	if a.Timestamp.GetTimeMilli() != b.Timestamp.GetTimeMilli() {
		return false
	}

	if a.Signature == nil && b.Signature != nil {
		return false
	}
	if a.Signature != nil {
		if a.Signature.IsSameAs(b.Signature) == false {
			return false
		}
	}
	//TODO: expand

	return true
}

//*******************************************************************************
// Support Functions
//*******************************************************************************

func NewServerFault(serverID interfaces.IHash, auditServerID interfaces.IHash, vmIndex int, dbheight uint32, height uint32, systemHeight int, timeStamp interfaces.Timestamp) *ServerFault {
	sf := new(ServerFault)
	sf.VMIndex = byte(vmIndex)
	sf.DBHeight = dbheight
	sf.Height = height
	sf.ServerID = serverID
	sf.AuditServerID = auditServerID
	sf.SystemHeight = uint32(systemHeight)
	sf.Timestamp = timeStamp
	return sf
}
