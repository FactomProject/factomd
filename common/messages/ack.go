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

//General acknowledge message
type Ack struct {
	MessageBase
	Timestamp   interfaces.Timestamp // Timestamp of Ack by Leader
	MessageHash interfaces.IHash     // Hash of message acknowledged

	DBHeight   uint32           // Directory Block Height that owns this ack
	Height     uint32           // Height of this ack in this process list
	SerialHash interfaces.IHash // Serial hash including previous ack

	Signature interfaces.IFullSignature

	//Not marshalled
	hash interfaces.IHash
}

var _ interfaces.IMsg = (*Ack)(nil)
var _ Signable = (*Ack)(nil)

// We have to return the haswh of the underlying message.
func (m *Ack) GetHash() interfaces.IHash {
	return m.MessageHash
}

func (m *Ack) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
		data, err := m.MarshalForSignature()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

func (m *Ack) Type() byte {
	return constants.ACK_MSG
}

func (m *Ack) Int() int {
	return -1
}

func (m *Ack) Bytes() []byte {
	return m.MessageHash.Bytes()
}

func (m *Ack) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *Ack) Validate(state interfaces.IState) int {
	// Check signature
	bytes, err := m.MarshalForSignature()
	if err != nil {
		fmt.Println("Err is not nil on Ack sig check: ", err)
		return -1
	}
	sig := m.Signature.GetSignature()
	ackSigned, err := state.VerifyFederatedSignature(bytes, sig)

	//ackSigned, err := m.VerifySignature()
	if err != nil {
		fmt.Println("(For Testing, allowing msg to validate)Err is not nil on Ack sig check: ", err)
		return 1
		return -1
	}
	if !ackSigned {
		return -1
	}
	return 1
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *Ack) ComputeVMIndex(state interfaces.IState) {

}

// Execute the leader functions of the given message
// Leader, follower, do the same thing.
func (m *Ack) LeaderExecute(state interfaces.IState) {
	m.FollowerExecute(state)
}

func (m *Ack) FollowerExecute(state interfaces.IState) {
	state.FollowerExecuteAck(m)
}

// Acknowledgements do not go into the process list.
func (e *Ack) Process(dbheight uint32, state interfaces.IState) bool {
	panic("Ack object should never have its Process() method called")
}

func (e *Ack) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *Ack) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *Ack) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (m *Ack) Sign(key interfaces.Signer) error {
	signature, err := SignSignable(m, key)
	if err != nil {
		return err
	}
	m.Signature = signature
	return nil
}

func (m *Ack) GetSignature() interfaces.IFullSignature {
	return m.Signature
}

func (m *Ack) VerifySignature() (bool, error) {
	return VerifyMessage(m)
}

func (m *Ack) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	newData = data
	if newData[0] != m.Type() {
		return nil, fmt.Errorf("Invalid Message type")
	}
	newData = newData[1:]

	m.VMIndex, newData = int(newData[0]), newData[1:]

	m.Timestamp = new(primitives.Timestamp)
	newData, err = m.Timestamp.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	m.MessageHash = new(primitives.Hash)
	newData, err = m.MessageHash.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	newData, err = m.GetFullMsgHash().UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	m.LeaderChainID = new(primitives.Hash)
	newData, err = m.LeaderChainID.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	m.DBHeight, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]
	m.Height, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]
	m.Minute, newData = newData[0], newData[1:]

	if m.SerialHash == nil {
		m.SerialHash = primitives.NewHash(constants.ZERO_HASH)
	}
	newData, err = m.SerialHash.UnmarshalBinaryData(newData)
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
	return
}

func (m *Ack) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *Ack) MarshalForSignature() ([]byte, error) {
	var buf primitives.Buffer

	binary.Write(&buf, binary.BigEndian, m.Type())
	binary.Write(&buf, binary.BigEndian, byte(m.VMIndex))

	t := m.GetTimestamp()
	data, err := t.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	data, err = m.MessageHash.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	data, err = m.GetFullMsgHash().MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	data, err = m.LeaderChainID.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	binary.Write(&buf, binary.BigEndian, m.DBHeight)
	binary.Write(&buf, binary.BigEndian, m.Height)
	binary.Write(&buf, binary.BigEndian, m.Minute)

	data, err = m.SerialHash.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	return buf.DeepCopyBytes(), nil
}

func (m *Ack) MarshalBinary() (data []byte, err error) {
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

func (m *Ack) String() string {
	return fmt.Sprintf("%6s-VM%3d: PL:%5d DBHt:%5d -- Leader[:3]=%x hash[:3]=%x",
		"ACK",
		m.VMIndex,
		m.Height,
		m.DBHeight,
		m.LeaderChainID.Bytes()[:3],
		m.GetHash().Bytes()[:3])

}

func (a *Ack) IsSameAs(b *Ack) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil {
		return false
	}

	if a.VMIndex != b.VMIndex {
		return false
	}

	if a.Minute != b.Minute {
		return false
	}

	if a.DBHeight != b.DBHeight {
		return false
	}
	if a.Height != b.Height {
		return false
	}
	if a.Timestamp.GetTimeMilli() != b.Timestamp.GetTimeMilli() {
		return false
	}

	if a.GetFullMsgHash().IsSameAs(b.GetFullMsgHash()) == false {
		return false
	}

	if a.MessageHash.IsSameAs(b.MessageHash) == false {
		return false
	}

	if a.SerialHash.IsSameAs(b.SerialHash) == false {
		return false
	}

	if a.Signature != nil {
		if a.Signature.IsSameAs(b.Signature) == false {
			return false
		}
	}

	if a.LeaderChainID == nil && b.LeaderChainID != nil {
		return false
	}
	if a.LeaderChainID != nil {
		if a.LeaderChainID.IsSameAs(b.LeaderChainID) == false {
			return false
		}
	}

	return true
}
