// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"bytes"
	"fmt"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/log"
)

//General acknowledge message
type Ack struct {
	Timestamp    Timestamp
	OriginalHash interfaces.IHash

	Signature *primitives.Signature
}

var _ interfaces.IMsg = (*Ack)(nil)

func (m *Ack) Type() int {
	return constants.ACK_MSG
}

func (m *Ack) Int() int {
	return -1
}

func (m *Ack) Bytes() []byte {
	return m.OriginalHash.Bytes()
}

func (m *Ack) GetTimestamp() *Timestamp {
	return &m.Timestamp
}

func (m *Ack) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	newData = data[1:]
	newData, err = m.Timestamp.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}
	m.OriginalHash = new(primitives.Hash)
	newData, err = m.OriginalHash.UnmarshalBinaryData(newData)
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

func (m *Ack) MarshalForSignature() (data []byte, err error) {
	return MarshalAckForSignature(m)
}

func (m *Ack) MarshalBinary() (data []byte, err error) {
	return MarshalAck(m)
}

func (m *Ack) String() string {
	str, _ := m.JSONString()
	log.Printf("str - %v", str)
	return str
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

func (m *Ack) Sign(key primitives.Signer) error {
	signature, err := SignSignable(m, key)
	if err != nil {
		return err
	}
	m.Signature = signature
	return nil
}

func (m *Ack) GetSignature() *primitives.Signature {
	return m.Signature
}

func (m *Ack) VerifySignature() (bool, error) {
	return VerifyMessage(m)
}

type IAck interface {
	Type() int
	GetTimestamp() *Timestamp
	Bytes() []byte
	GetSignature() *primitives.Signature
}

func MarshalAckForSignature(ack IAck) ([]byte, error) {
	resp := []byte{}
	resp = append(resp, byte(ack.Type()))
	timeByte, err := ack.GetTimestamp().MarshalBinary()
	if err != nil {
		return nil, err
	}
	resp = append(resp, timeByte...)
	resp = append(resp, ack.Bytes()...)
	return resp, nil
}

func MarshalAck(ack IAck) ([]byte, error) {
	resp, err := MarshalAckForSignature(ack)
	if err != nil {
		return nil, err
	}
	sig := ack.GetSignature()

	if sig != nil {
		sigBytes, err := sig.MarshalBinary()
		if err != nil {
			return nil, err
		}
		return append(resp, sigBytes...), nil
	}
	return resp, nil
}
