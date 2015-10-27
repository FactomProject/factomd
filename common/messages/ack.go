// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"bytes"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

//General acknowledge message
type Ack struct {
	Timestamp
	OriginalHash interfaces.IHash
}

var _ interfaces.IMsg = (*Ack)(nil)

func (m *Ack) Type() int {
	return -1
}

func (m *Ack) Int() int {
	return -1
}

func (m *Ack) Bytes() []byte {
	return m.OriginalHash.Bytes()
}

func (m *Ack) UnmarshalBinaryData(data []byte) (newdata []byte, err error) {
	m.OriginalHash = new(primitives.Hash)
	return m.OriginalHash.UnmarshalBinaryData(data)
}

func (m *Ack) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *Ack) MarshalBinary() (data []byte, err error) {
	return MarshalAck(m)
}

func (m *Ack) String() string {
	return ""
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *Ack) Validate(interfaces.IState) int {
	return 0
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *Ack) Leader(state interfaces.IState) bool {
	switch state.GetNetworkNumber() {
	case 0: // Main Network
		panic("Not implemented yet")
	case 1: // Test Network
		panic("Not implemented yet")
	case 2: // Local Network
		panic("Not implemented yet")
	default:
		panic("Not implemented yet")
	}

}

// Execute the leader functions of the given message
func (m *Ack) LeaderExecute(state interfaces.IState) error {
	return nil
}

// Returns true if this is a message for this server to execute as a follower
func (m *Ack) Follower(interfaces.IState) bool {
	return true
}

func (m *Ack) FollowerExecute(interfaces.IState) error {
	return nil
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

type IAck interface {
	Type() int
	GetTimeByte() ([]byte, error)
	Bytes() []byte
}

func MarshalAck(ack IAck) ([]byte, error) {
	resp := []byte{}
	resp = append(resp, byte(ack.Type()))
	timeByte, err := ack.GetTimeByte()
	if err != nil {
		return nil, err
	}
	resp = append(resp, timeByte...)
	resp = append(resp, ack.Bytes()...)
	return resp, nil
}
