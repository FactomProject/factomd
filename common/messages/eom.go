// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/log"
)

type EOM struct {
	Minute byte

	DirectoryBlockHeight uint32
	IdentityChainID      interfaces.IHash

	Signature *primitives.Signature
}

//var _ interfaces.IConfirmation = (*EOM)(nil)
var _ Signable = (*EOM)(nil)

func (m *EOM) Int() int {
	return int(m.Minute)
}

func (m *EOM) Bytes() []byte {
	var ret []byte
	return append(ret, m.Minute)
}

func (m *EOM) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	newData = data[1:]
	m.Minute, newData = newData[0], newData[1:]

	if m.Minute < 0 || m.Minute >= 10 {
		return nil, fmt.Errorf("Minute number is out of range")
	}

	m.DirectoryBlockHeight, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]

	hash := new(primitives.Hash)
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

	return data, nil
}

func (m *EOM) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *EOM) MarshalForSignature() (data []byte, err error) {
	var buf bytes.Buffer
	buf.Write([]byte{byte(m.Type())})
	binary.Write(&buf, binary.BigEndian, m.Minute)
	binary.Write(&buf, binary.BigEndian, m.DirectoryBlockHeight)
	hash, err := m.IdentityChainID.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(hash)

	return buf.Bytes(), nil
}

func (m *EOM) MarshalBinary() (data []byte, err error) {
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

func (m *EOM) String() string {
	return fmt.Sprintf("EOM(%d)", m.Minute+1)
}

func (m *EOM) Type() int {
	return constants.EOM_MSG
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *EOM) Validate(interfaces.IState) int {
	return 1
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *EOM) Leader(state interfaces.IState) bool {
	return false
}

// Execute the leader functions of the given message
func (m *EOM) LeaderExecute(state interfaces.IState) error {
	olddb := state.GetCurrentDirectoryBlock()
	state.GetFactoidState().ProcessEndOfBlock(state)
	
	
	db, err := directoryblock.CreateDBlock(uint32(state.GetDBHeight()), olddb, 10)
	
	state.SetDBHeight(state.GetDBHeight() + 1)
	if err != nil {
		panic(err.Error())
	}
	state.SetCurrentDirectoryBlock(db)
	db.AddEntry(primitives.NewHash(constants.ADMIN_CHAINID), primitives.NewZeroHash())   // AdminBlock
	db.AddEntry(primitives.NewHash(constants.EC_CHAINID), primitives.NewZeroHash())      // AdminBlock
	db.AddEntry(primitives.NewHash(constants.FACTOID_CHAINID), primitives.NewZeroHash()) // AdminBlock

	if olddb != nil {
		bodyMR, err := olddb.BuildBodyMR()
		if err != nil {
			return err
		}
		olddb.GetHeader().SetBodyMR(bodyMR)
		err = state.GetDB().Put([]byte(constants.DB_DIRECTORY_BLOCKS), olddb.GetKeyMR().Bytes(), olddb)
		if err != nil {
			return err
		}
		err = state.GetDB().Put([]byte(constants.DB_DIRECTORY_BLOCKS), constants.D_CHAINID, olddb)
		if err != nil {
			return err
		}
		log.Printfln("%v", olddb)
	} else {
		log.Println("No old db")
	}
	return nil
}

// Returns true if this is a message for this server to execute as a follower
func (m *EOM) Follower(interfaces.IState) bool {
	return true
}

func (m *EOM) FollowerExecute(state interfaces.IState) error {
	
	state.GetFactoidState().EndOfPeriod(int(m.Minute))
	
	switch state.GetNetworkNumber() {
		case 0: // Main Network
			panic("Not implemented yet")
		case 1: // Test Network
			panic("Not implemented yet")
		case 2: // Local Network
			
		default:
			panic("Not implemented yet")
	}
	
	return nil
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

func (m *EOM) Sign(key primitives.Signer) error {
	signature, err := SignSignable(m, key)
	if err != nil {
		return err
	}
	m.Signature = signature
	return nil
}

func (m *EOM) GetSignature() *primitives.Signature {
	return m.Signature
}

func (m *EOM) VerifySignature() (bool, error) {
	return VerifyMessage(m)
}

/**********************************************************************
 * Support
 **********************************************************************/

func NewEOM(state interfaces.IState, minute int) interfaces.IMsg {
	// The construction of the EOM message needs information from the state of 
	// the server to create the proper serial hashes and such.  Right now
	// I am ignoring all of that.
	eom := new(EOM)
	eom.Minute = byte(minute)
	return eom
}
