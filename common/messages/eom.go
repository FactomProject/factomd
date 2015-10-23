// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"fmt"
	"bytes"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/directoryBlock"
)

type EOM struct {
	minute 		int
	
	dbHeight 	uint32
	chainID     interfaces.IHash
	listHeight  uint32
	serialHash  interfaces.IHash
	signature   []byte
}

var _ interfaces.IConfirmation = (*EOM)(nil)

func (m *EOM) Int() int {
	return m.minute
}

func (m *EOM) Bytes() []byte {
	var ret []byte
	return append(ret, byte(m.minute))
}

func (m *EOM) UnmarshalBinaryData(data [] byte) (newdata []byte, err error){
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	m.minute, data = int(data[0]), data[1:]

	if m.minute < 0 || m.minute >= 10 {
		return nil, fmt.Errorf("Minute number is out of range")
	}
	
	//m.dbHeight, data = binary.BigEndian.Uint32(data[:]), data[4:]

	return data, nil
}

func (m *EOM) UnmarshalBinary(data []byte) error{
	_,err := m.UnmarshalBinaryData(data)
	return err
}

func (m *EOM) MarshalBinary() (data []byte, err error) {
	return m.Bytes(), nil
}

func (m *EOM) String() string {
	return fmt.Sprintf("EOM(%d)",m.minute+1)
}
	

func (m *EOM) Type() int {
	return constants.EOM_MSG
}
func (m *EOM) DBHeight() int	{
	return 0
}
func (m *EOM) ChainID() []byte {
	return nil
}
func (m *EOM) ListHeight() int {
	return 0
}

func (m *EOM) SerialHash() []byte {
	return nil
}
func (m *EOM) Signature() []byte {
	return nil
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
	switch state.GetNetworkNumber() {
		case 0 : // Main Network
			panic("Not implemented yet")
		case 1 : // Test Network
			panic("Not implemented yet")
		case 2 : // Local Network
			
			// Note!  We should validate that we are the server for this network
			// by checking keys!
			if state.GetServerState() == 1 {
				return true
			}else{
				return false
			}
		default :
			panic("Not implemented yet")
	}
	
}
// Execute the leader functions of the given message
func (m *EOM) LeaderExecute(state interfaces.IState) error {
	if state.GetServerState() == 1 {
		if m.minute == 9 {
			olddb := state.GetCurrentDirectoryBlock()
			db, err := directoryblock.CreateDBlock(uint32(state.GetDBHeight()),olddb,10)
			state.SetDBHeight(state.GetDBHeight()+1)
			if err != nil {
				panic(err.Error())
			}
			state.SetCurrentDirectoryBlock(db)
			db.AddEntry(primitives.NewHash(constants.ADMIN_CHAINID),   primitives.NewZeroHash()) // AdminBlock
			db.AddEntry(primitives.NewHash(constants.EC_CHAINID),      primitives.NewZeroHash()) // AdminBlock
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
				fmt.Println(olddb)
			}else{
				fmt.Println("No old db")
			}
		}
	}
	return nil
}

// Returns true if this is a message for this server to execute as a follower
func (m *EOM) Follower(interfaces.IState) bool {
	return true
}

func (m *EOM) FollowerExecute(interfaces.IState) error {
	return nil
}

func (m *EOM) JSONByte() ([]byte, error) {
	return nil, nil
}
func (m *EOM) JSONString() (string, error) {
	return "", nil
}
func (m *EOM) JSONBuffer(b *bytes.Buffer) error {
	return nil
}

/**********************************************************************
 * Support
 **********************************************************************/

func NewEOM(state interfaces.IState, minute int) interfaces.IMsg {
	eom := new(EOM)
	eom.minute = minute
	eom.dbHeight = state.GetDBHeight()
	return eom
}