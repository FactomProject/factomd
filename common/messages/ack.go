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
	ServerIndex byte // Server index (signature could be one of several)
	Timestamp   interfaces.Timestamp
	MessageHash interfaces.IHash

	DBHeight uint32 // Directory Block Height that owns this ack
	Height   uint32 // Height of this ack in this process list

	SerialHash interfaces.IHash

	Signature interfaces.IFullSignature

	//Not marshalled
	hash interfaces.IHash
}

var _ interfaces.IMsg = (*Ack)(nil)
var _ Signable = (*Ack)(nil)

func (a *Ack) IsSameAs(b *Ack) bool {
	if b == nil {
		return false
	}
	if a.ServerIndex != b.ServerIndex {
		return false
	}
	if a.DBHeight != b.DBHeight {
		return false
	}
	if a.Height != b.Height {
		return false
	}
	if a.Timestamp != b.Timestamp {
		return false
	}

	if a.MessageHash == nil && b.MessageHash != nil {
		return false
	}
	if a.MessageHash.IsSameAs(b.MessageHash) == false {
		return false
	}

	if a.SerialHash == nil && b.SerialHash != nil {
		return false
	}
	if a.SerialHash.IsSameAs(b.SerialHash) == false {
		return false
	}

	if a.Signature == nil && b.Signature != nil {
		return false
	}
	if a.Signature.IsSameAs(b.Signature) == false {
		return false
	}

	return true
}

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

func (m *Ack) Type() int {
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
	return 1
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *Ack) Leader(state interfaces.IState) bool {
	return false
}

// Execute the leader functions of the given message
func (m *Ack) LeaderExecute(state interfaces.IState) error {
	return fmt.Errorf("Should never execute an Acknowledgement in the Leader")
}

// Returns true if this is a message for this server to execute as a follower
func (m *Ack) Follower(interfaces.IState) bool {
	return true
}

func (m *Ack) FollowerExecute(state interfaces.IState) error {
	_, err := state.FollowerExecuteAck(m)
	return err
}

// Acknowledgements do not go into the process list.
func (e *Ack) Process(dbheight uint32, state interfaces.IState) {
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

	newData = data[1:]
	m.ServerIndex, newData = newData[0], newData[1:]

	newData, err = m.Timestamp.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	m.MessageHash = new(primitives.Hash)
	newData, err = m.MessageHash.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	m.DBHeight, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]
	m.Height, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]

	if m.SerialHash == nil {
		m.SerialHash = primitives.NewHash(constants.ZERO_HASH)
	}
	newData, err = m.SerialHash.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	if len(newData) > 0 {
		sig := new(primitives.Signature)
		newData, err = sig.UnmarshalBinaryData(newData)
		if err != nil {
			return nil, err
		}
		m.Signature = sig
	}
	return
}

func (m *Ack) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *Ack) MarshalForSignature() ([]byte, error) {
	var buf bytes.Buffer

	binary.Write(&buf, binary.BigEndian, byte(m.Type()))
	binary.Write(&buf, binary.BigEndian, m.ServerIndex)

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

	binary.Write(&buf, binary.BigEndian, m.DBHeight)
	binary.Write(&buf, binary.BigEndian, m.Height)

	data, err = m.SerialHash.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	return buf.Bytes(), nil
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
	return fmt.Sprintf("%6s-%3d: db/pl %2d/%2d,   -- hash[:10]=%x",
		"ACK",
		m.ServerIndex,
		m.DBHeight,
		m.Height,
		m.GetHash().Bytes()[:10])
}
