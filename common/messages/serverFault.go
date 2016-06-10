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

//A placeholder structure for messages
type ServerFault struct {
	MessageBase
	Timestamp interfaces.Timestamp

	ServerID interfaces.IHash
	VMIndex  int
	DBHeight uint32
	Height   uint32

	Signature interfaces.IFullSignature

	//Not marshalled
	hash interfaces.IHash
}

var _ interfaces.IMsg = (*ServerFault)(nil)
var _ Signable = (*ServerFault)(nil)

func (m *ServerFault) Process(uint32, interfaces.IState) bool { return true }

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

func (m *ServerFault) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp
}

func (m *ServerFault) Type() byte {
	return constants.FED_SERVER_FAULT_MSG
}

func (m *ServerFault) MarshalForSignature() (data []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling Invalid Server Fault: %v", r)
		}
	}()

	var buf primitives.Buffer

	buf.Write([]byte{m.Type()})
	if d, err := m.Timestamp.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}
	if d, err := m.ServerID.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}

	buf.WriteByte(byte(m.VMIndex))
	binary.Write(&buf, binary.BigEndian, uint32(m.DBHeight))
	binary.Write(&buf, binary.BigEndian, uint32(m.Height))

	return buf.DeepCopyBytes(), nil
}

func (m *ServerFault) MarshalBinary() (data []byte, err error) {
	resp, err := m.MarshalForSignature()
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

	newData, err = m.Timestamp.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	m.VMIndex, newData = int(newData[0]), newData[1:]
	m.DBHeight, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]
	m.Height, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]

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
	return VerifyMessage(m)
}

func (m *ServerFault) Sign(key interfaces.Signer) error {
	signature, err := SignSignable(m, key)
	if err != nil {
		return err
	}
	m.Signature = signature
	return nil
}

func (m *ServerFault) String() string {
	return fmt.Sprintf("%6s-VM%3d: PL:%5d DBHt:%5d -- hash[:3]=%x",
		"SFault",
		m.VMIndex,
		m.Height,
		m.DBHeight,
		m.GetHash().Bytes()[:3])
}

func (m *ServerFault) GetDBHeight() uint32 {
	return m.DBHeight
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *ServerFault) Validate(state interfaces.IState) int {
	return 1 //ToDo:  Need to Validate the sigature against known federated servers
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

func (e *ServerFault) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (a *ServerFault) IsSameAs(b *ServerFault) bool {
	if b == nil {
		return false
	}
	if a.Timestamp != b.Timestamp {
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

func NewServerFault(timeStamp interfaces.Timestamp, serverID interfaces.IHash, vmIndex int, dbheight uint32, height uint32) *ServerFault {
	sf := new(ServerFault)
	sf.Timestamp = timeStamp
	sf.VMIndex = vmIndex
	sf.DBHeight = dbheight
	sf.Height = height
	return sf
}
