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
	"github.com/FactomProject/factomd/log"
)

var _ = log.Printf

type EOM struct {
	MessageBase
	Timestamp interfaces.Timestamp
	Minute    byte

	DirectoryBlockHeight uint32
	ServerIndex          int
	ChainID              interfaces.IHash
	Signature            interfaces.IFullSignature

	//Not marshalled
	hash      interfaces.IHash
	MarkerSet bool // Set if we have Processed EOM markers, so we don't repeat.
}

//var _ interfaces.IConfirmation = (*EOM)(nil)
var _ Signable = (*EOM)(nil)

func (e *EOM) Process(dbheight uint32, state interfaces.IState) bool {
	return state.ProcessEOM(dbheight, e)
}

func (m *EOM) GetHash() interfaces.IHash {
	if m.hash == nil {
		data, err := m.MarshalForSignature()
		if err != nil {
			panic(fmt.Sprintf("Error in EOM.GetHash(): %s", err.Error()))
		}
		m.hash = primitives.Sha(data)
	}
	return m.hash
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
	return m.Timestamp
}

func (m *EOM) Int() int {
	return int(m.Minute)
}

func (m *EOM) Bytes() []byte {
	var ret []byte
	return append(ret, m.Minute)
}

func (m *EOM) Type() int {
	return constants.EOM_MSG
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *EOM) Validate(state interfaces.IState) int {

	return 1

	// TODO:  Need to check that the EOM came from a server
	found, _ := state.GetFedServerIndexFor(m.DirectoryBlockHeight, m.ChainID)
	if found { // Only EOM from federated servers are valid.
		return 1
	} else {
		return -1
	}
	//TODO: Check signatures here.
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *EOM) Leader(state interfaces.IState) bool {
	return state.LeaderFor(m.GetHash().Bytes()) // TODO: This has to be fixed!
}

// Execute the leader functions of the given message
func (m *EOM) LeaderExecute(state interfaces.IState) error {
	return state.LeaderExecuteEOM(m)
}

// Returns true if this is a message for this server to execute as a follower
func (m *EOM) Follower(interfaces.IState) bool {
	return true
}

func (m *EOM) FollowerExecute(state interfaces.IState) error {
	_, err := state.FollowerExecuteMsg(m)
	return err
}

func (e *EOM) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *EOM) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *EOM) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (m *EOM) Sign(key interfaces.Signer) error {
	signature, err := SignSignable(m, key)
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
	return VerifyMessage(m)
}

func (m *EOM) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
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

	m.ChainID = primitives.NewHash(constants.ZERO_HASH)
	newData, err = m.ChainID.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	m.Minute, newData = newData[0], newData[1:]

	if m.Minute < 0 || m.Minute >= 10 {
		return nil, fmt.Errorf("Minute number is out of range")
	}

	m.ServerIndex = int(newData[0])
	newData = newData[1:]

	m.DirectoryBlockHeight, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]

	if len(newData) > 0 {
		sig := new(primitives.Signature)
		newData, err = sig.UnmarshalBinaryData(newData)
		if err != nil {
			return nil, err
		}
		m.Signature = sig
	}

	return data, nil
}

func (m *EOM) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *EOM) MarshalForSignature() (data []byte, err error) {
	var buf bytes.Buffer
	buf.Write([]byte{byte(m.Type())})
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
	binary.Write(&buf, binary.BigEndian, uint8(m.ServerIndex))
	return buf.Bytes(), nil
}

func (m *EOM) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer
	resp, err := m.MarshalForSignature()
	if err != nil {
		return nil, err
	}
	buf.Write(resp)

	binary.Write(&buf, binary.BigEndian, m.DirectoryBlockHeight)

	sig := m.GetSignature()
	if sig != nil {
		sigBytes, err := sig.MarshalBinary()
		if err != nil {
			return nil, err
		}
		buf.Write(sigBytes)
	}
	return buf.Bytes(), nil
}

func (m *EOM) String() string {
	return fmt.Sprintf("%6s-%3d: Min:%4d Ht:%5d -- chainID[:5]=%x hash[:5]=%x",
		"EOM",
		m.ServerIndex,
		m.Minute,
		m.DirectoryBlockHeight,
		m.ChainID.Bytes()[:5],
		m.GetMsgHash().Bytes()[:5])

}



