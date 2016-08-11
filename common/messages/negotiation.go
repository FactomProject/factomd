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
type Negotiation struct {
	MessageBase
	Timestamp interfaces.Timestamp

	// The following 4 fields represent the "Core" of the message
	ServerID interfaces.IHash
	VMIndex  byte
	DBHeight uint32
	Height   uint32

	Signature interfaces.IFullSignature

	//Not marshalled
	hash interfaces.IHash
}

var _ interfaces.IMsg = (*Negotiation)(nil)
var _ Signable = (*Negotiation)(nil)

func (m *Negotiation) Process(uint32, interfaces.IState) bool { return true }

func (m *Negotiation) GetRepeatHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *Negotiation) GetHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *Negotiation) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
		data, err := m.MarshalBinary()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

func (m *Negotiation) GetCoreHash() interfaces.IHash {
	data, err := m.MarshalForSignature()
	if err != nil {
		return nil
	}
	return primitives.Sha(data)
}

func (m *Negotiation) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp
}

func (m *Negotiation) Type() byte {
	return constants.NEGOTIATION_MSG
}

func (m *Negotiation) MarshalForSignature() (data []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error marshalling Negotiation Core: %v", r)
		}
	}()

	var buf primitives.Buffer

	if d, err := m.ServerID.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}

	buf.WriteByte(m.VMIndex)
	binary.Write(&buf, binary.BigEndian, uint32(m.DBHeight))
	binary.Write(&buf, binary.BigEndian, uint32(m.Height))

	return buf.DeepCopyBytes(), nil
}

func (m *Negotiation) PreMarshalBinary() (data []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error marshalling Invalid Negotiation: %v", r)
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

	buf.WriteByte(m.VMIndex)
	binary.Write(&buf, binary.BigEndian, uint32(m.DBHeight))
	binary.Write(&buf, binary.BigEndian, uint32(m.Height))

	return buf.DeepCopyBytes(), nil
}

func (m *Negotiation) MarshalBinary() (data []byte, err error) {
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

func (m *Negotiation) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling With Signatures Invalid Negotiation: %v", r)
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

	if m.ServerID == nil {
		m.ServerID = primitives.NewZeroHash()
	}
	newData, err = m.ServerID.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	m.VMIndex, newData = newData[0], newData[1:]
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

func (m *Negotiation) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *Negotiation) GetSignature() interfaces.IFullSignature {
	return m.Signature
}

func (m *Negotiation) VerifySignature() (bool, error) {
	return VerifyMessage(m)
}

func (m *Negotiation) Sign(key interfaces.Signer) error {
	signature, err := SignSignable(m, key)
	if err != nil {
		return err
	}
	m.Signature = signature
	return nil
}

func (m *Negotiation) String() string {
	return fmt.Sprintf("%6s-VM%3v: (%v) PL:%5d DBHt:%5d -- hash[:3]=%x",
		"Negotiation",
		m.VMIndex,
		m.ServerID.String()[:8],
		m.Height,
		m.DBHeight,
		m.GetHash().Bytes()[:3])
}

func (m *Negotiation) GetDBHeight() uint32 {
	return m.DBHeight
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *Negotiation) Validate(state interfaces.IState) int {
	if m.Signature == nil {
		return -1
	}
	// Check signature
	bytes, err := m.MarshalForSignature()
	if err != nil {
		//fmt.Println("Err is not nil on Negotiation sig check (marshalling): ", err)
		return -1
	}
	sig := m.Signature.GetSignature()
	negSigned, err := state.VerifyFederatedSignature(bytes, sig)
	if err != nil {
		//fmt.Println("Err is not nil on Negotiation sig check (verifying): ", err)
		return -1
	}
	if !negSigned {
		return -1
	}
	return 1 // err == nil and negSigned == true
}

func (m *Negotiation) ComputeVMIndex(state interfaces.IState) {

}

// Execute the leader functions of the given message
func (m *Negotiation) LeaderExecute(state interfaces.IState) {
	m.FollowerExecute(state)
}

func (m *Negotiation) FollowerExecute(state interfaces.IState) {
	state.FollowerExecuteNegotiation(m)
}

func (e *Negotiation) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *Negotiation) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *Negotiation) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (a *Negotiation) IsSameAs(b *Negotiation) bool {
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

func NewNegotiation(timeStamp interfaces.Timestamp, serverID interfaces.IHash, vmIndex int, dbheight uint32, height uint32) *Negotiation {
	nego := new(Negotiation)
	nego.Timestamp = timeStamp
	nego.VMIndex = byte(vmIndex)
	nego.DBHeight = dbheight
	nego.Height = height
	nego.ServerID = serverID
	return nego
}
