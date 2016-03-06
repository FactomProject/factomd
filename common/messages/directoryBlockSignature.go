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
type DirectoryBlockSignature struct {
	MessageBase
	Timestamp             interfaces.Timestamp
	DBHeight  			  uint32
	DirectoryBlockKeyMR   interfaces.IHash
	ServerIdentityChainID interfaces.IHash

	Signature interfaces.IFullSignature

	//Not marshalled
	hash interfaces.IHash
}

var _ interfaces.IMsg = (*DirectoryBlockSignature)(nil)
var _ Signable = (*DirectoryBlockSignature)(nil)

func (e *DirectoryBlockSignature) Process(dbheight uint32, state interfaces.IState) {
	state.ProcessSignPL(dbheight, e)
}

func (m *DirectoryBlockSignature) GetHash() interfaces.IHash {
	if m.hash == nil {
		data, err := m.MarshalForSignature()
		if err != nil {
			panic(fmt.Sprintf("Error in CommitChain.GetHash(): %s", err.Error()))
		}
		m.hash = primitives.Sha(data)
	}
	return m.hash
}

func (m *DirectoryBlockSignature) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
		data, _ := m.MarshalForSignature()
		if data == nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

func (m *DirectoryBlockSignature) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp
}

func (m *DirectoryBlockSignature) Type() int {
	return constants.DIRECTORY_BLOCK_SIGNATURE_MSG
}

func (m *DirectoryBlockSignature) Int() int {
	return -1
}

func (m *DirectoryBlockSignature) Bytes() []byte {
	return nil
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *DirectoryBlockSignature) Validate(dbheight uint32, state interfaces.IState) int {
	return 1
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *DirectoryBlockSignature) Leader(state interfaces.IState) bool {
	return true
}

// Execute the leader functions of the given message
func (m *DirectoryBlockSignature) LeaderExecute(state interfaces.IState) error {
	return state.LeaderExecuteDBSig(m)
}

// Returns true if this is a message for this server to execute as a follower
func (m *DirectoryBlockSignature) Follower(interfaces.IState) bool {
	return true
}

func (m *DirectoryBlockSignature) FollowerExecute(state interfaces.IState) error {
	_, err := state.FollowerExecuteMsg(m)
	return err
}

func (m *DirectoryBlockSignature) Sign(key interfaces.Signer) error {
	signature, err := SignSignable(m, key)
	if err != nil {
		return err
	}
	m.Signature = signature
	return nil
}

func (m *DirectoryBlockSignature) GetSignature() interfaces.IFullSignature {
	return m.Signature
}

func (m *DirectoryBlockSignature) VerifySignature() (bool, error) {
	return VerifyMessage(m)
}

func (m *DirectoryBlockSignature) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()

	// Type byte:  Someone else's problem.
	newData = data[1:]

	// TimeStamp
	newData, err = m.Timestamp.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	m.DBHeight, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]

	hash := new(primitives.Hash)
	newData, err = hash.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}
	m.DirectoryBlockKeyMR = hash

	hash = new(primitives.Hash)
	newData, err = hash.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}
	m.ServerIdentityChainID = hash

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

func (m *DirectoryBlockSignature) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *DirectoryBlockSignature) MarshalForSignature() ([]byte, error) {

	if m.DirectoryBlockKeyMR == nil || m.ServerIdentityChainID == nil {
		return nil, fmt.Errorf("Message is incomplete")
	}

	var buf bytes.Buffer
	buf.Write([]byte{byte(m.Type())})

	t := m.GetTimestamp()
	data, err := t.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	binary.Write(&buf, binary.BigEndian, m.DBHeight)

	hash, err := m.DirectoryBlockKeyMR.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(hash)

	hash, err = m.ServerIdentityChainID.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(hash)

	return buf.Bytes(), nil
}

func (m *DirectoryBlockSignature) MarshalBinary() (data []byte, err error) {
	
	var sig interfaces.IFullSignature
	resp, err := m.MarshalForSignature()
	if err == nil {
		sig = m.GetSignature()
	}
	
	if sig != nil {
		sigBytes, err := sig.MarshalBinary()
		if err != nil {
			return resp, nil
		}
		return append(resp, sigBytes...), nil
	}
	return resp, nil
}

func (m *DirectoryBlockSignature) String() string {
	return fmt.Sprintf("DB Sig %s", m.GetHash().String())
}

func (e *DirectoryBlockSignature) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *DirectoryBlockSignature) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *DirectoryBlockSignature) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

/*******************************************************************
 * Support
 *******************************************************************/

func NewDirectoryBlockSignature(dbHeight uint32) *DirectoryBlockSignature {
	dbm := new(DirectoryBlockSignature)
	dbm.DBHeight = dbHeight
	dbm.DirectoryBlockKeyMR = primitives.NewHash(constants.ZERO_HASH)
	dbm.ServerIdentityChainID = primitives.NewHash(constants.ZERO_HASH)
	return dbm
}
