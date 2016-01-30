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

//A placeholder structure for messages
type Heartbeat struct {
	Timestamp       interfaces.Timestamp
	DBlockHash      interfaces.IHash //Hash of last Directory Block
	IdentityChainID interfaces.IHash //Identity Chain ID

	Signature interfaces.IFullSignature

	//Not marshalled
	hash interfaces.IHash
}

var _ interfaces.IMsg = (*Heartbeat)(nil)
var _ Signable = (*Heartbeat)(nil)

func (m *Heartbeat) Process(uint32, interfaces.IState) {}

func (m *Heartbeat) GetHash() interfaces.IHash {
	if m.hash == nil {
		data, err := m.MarshalForSignature()
		if err != nil {
			panic(fmt.Sprintf("Error in CommitChain.GetHash(): %s", err.Error()))
		}
		m.hash = primitives.Sha(data)
	}
	return m.hash
}

func (m *Heartbeat) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp
}

func (m *Heartbeat) Type() int {
	return constants.HEARTBEAT_MSG
}

func (m *Heartbeat) Int() int {
	return -1
}

func (m *Heartbeat) Bytes() []byte {
	return nil
}

func (m *Heartbeat) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()

	data = data[1:] // skip type

	newData, err = m.Timestamp.UnmarshalBinaryData(data)
	if err != nil {
		return nil, err
	}

	hash := new(primitives.Hash)

	newData, err = hash.UnmarshalBinaryData(data)
	if err != nil {
		return nil, err
	}
	m.DBlockHash = hash

	hash = new(primitives.Hash)
	newData, err = hash.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}
	m.IdentityChainID = hash

	if len(newData) > 0 {
		sig := new(primitives.Signature)
		newData, err = sig.UnmarshalBinaryData(newData)
		if err != nil {
			return nil, err
		}
		m.Signature = sig
	}

	return nil, nil
}

func (m *Heartbeat) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *Heartbeat) MarshalForSignature() (data []byte, err error) {
	if m.DBlockHash == nil || m.IdentityChainID == nil {
		return nil, fmt.Errorf("Message is incomplete")
	}

	answer := []byte{}

	answer = append(answer, byte(m.Type()))

	ts, err := m.Timestamp.MarshalBinary()
	if err != nil {
		return nil, err
	}
	answer = append(answer, ts...)

	hash, err := m.DBlockHash.MarshalBinary()
	if err != nil {
		return nil, err
	}
	answer = append(answer, hash...)

	hash2, err := m.IdentityChainID.MarshalBinary()
	if err != nil {
		return nil, err
	}
	answer = append(answer, hash2...)
	return answer, nil
}

func (m *Heartbeat) MarshalBinary() (data []byte, err error) {
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

func (m *Heartbeat) String() string {
	return ""
}

func (m *Heartbeat) DBHeight() int {
	return 0
}

func (m *Heartbeat) ChainID() []byte {
	return nil
}

func (m *Heartbeat) ListHeight() int {
	return 0
}

func (m *Heartbeat) SerialHash() []byte {
	return nil
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *Heartbeat) Validate(dbheight uint32, state interfaces.IState) int {
	return 0
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *Heartbeat) Leader(state interfaces.IState) bool {
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
func (m *Heartbeat) LeaderExecute(state interfaces.IState) error {
	return nil
}

// Returns true if this is a message for this server to execute as a follower
func (m *Heartbeat) Follower(interfaces.IState) bool {
	return true
}

func (m *Heartbeat) FollowerExecute(interfaces.IState) error {
	return nil
}

func (e *Heartbeat) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *Heartbeat) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *Heartbeat) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (m *Heartbeat) Sign(key interfaces.Signer) error {
	signature, err := SignSignable(m, key)
	if err != nil {
		return err
	}
	m.Signature = signature
	return nil
}

func (m *Heartbeat) GetSignature() interfaces.IFullSignature {
	return m.Signature
}

func (m *Heartbeat) VerifySignature() (bool, error) {
	return VerifyMessage(m)
}
