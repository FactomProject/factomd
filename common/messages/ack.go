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
)

//General acknowledge message
type Ack struct {
	ServerIndex  int 					// Server index (signature could be one of several)
	Timestamp    interfaces.Timestamp
	MessageHash  interfaces.IHash

	Height       int
	SerialHash   interfaces.IHash
	
	Signature interfaces.IFullSignature

	//Not marshalled
	hash interfaces.IHash
}

var _ interfaces.IMsg = (*Ack)(nil)
var _ Signable = (*Ack)(nil)

func NewAck(state interfaces.IState, msg []byte) (iack interfaces.IMsg, err error) {
	var last *Ack
	if state.GetLastAck() != nil {
		last = state.GetLastAck().(*Ack)
	}
	ack := new(Ack)
	ack.Timestamp    = state.GetTimestamp()
	ack.MessageHash  = primitives.Sha(msg)
	if last == nil {
		ack.Height = 0
		ack.SerialHash = ack.MessageHash
	}else{
		ack.Height = last.Height+1
		ack.SerialHash,err = primitives.CreateHash(last.MessageHash,ack.MessageHash)
		if err != nil {
			return nil, err
		}
	}

	last = ack

// TODO:  Add the signature.
	
	return ack, nil
}


func (m *Ack) GetHash() interfaces.IHash {
	if m.hash == nil {
		data, err := m.MarshalForSignature()
		if err != nil {
			panic(fmt.Sprintf("Error in Ack.GetHash(): %s", err.Error()))
		}
		m.hash = primitives.Sha(data)
	}
	return m.hash
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

func (m *Ack) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	newData = data[1:]
	m.ServerIndex = (int) (newData[0])
	newData = newData[1:]
	newData, err = m.Timestamp.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}
	m.MessageHash = new(primitives.Hash)
	newData, err = m.MessageHash.UnmarshalBinaryData(newData)
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
	resp := []byte{}
	resp = append(resp, byte(m.Type()))
	resp = append(resp, byte(m.ServerIndex))
	t := m.GetTimestamp()
	timeByte, err := t.MarshalBinary()
	if err != nil {
		return nil, err
	}
	resp = append(resp, timeByte...)
	resp = append(resp, m.Bytes()...)
	return resp, nil
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
	str, _ := m.JSONString()
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
	acks := state.GetAcks()
	holding := state.GetHolding()
	msg := holding[m.MessageHash.Fixed()]
	if msg == nil {
		acks[m.GetHash().Fixed()]=m
	}else{
		processlist := state.GetProcessList()[m.ServerIndex]
		for len(processlist)< m.Height+1 {
			processlist = append(processlist,nil)
		}
		processlist[m.Height]=msg
		state.GetProcessList()[m.ServerIndex]=processlist
		delete(acks,m.GetHash().Fixed())
	}
	
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

func (m *Ack) GetSignature() interfaces.IFullSignature {
	return m.Signature
}

func (m *Ack) VerifySignature() (bool, error) {
	return VerifyMessage(m)
}
