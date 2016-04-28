// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"bytes"
	//	"encoding/binary"
	"encoding/binary"
	"fmt"

	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// Communicate a Directory Block State

type DBStateMsg struct {
	MessageBase
	Timestamp interfaces.Timestamp

	//TODO: handle misformed DBStates!
	DirectoryBlock   interfaces.IDirectoryBlock
	AdminBlock       interfaces.IAdminBlock
	FactoidBlock     interfaces.IFBlock
	EntryCreditBlock interfaces.IEntryCreditBlock
	EntryBlocks      []interfaces.IEntryBlock

	//Not signed!
}

var _ interfaces.IMsg = (*DBStateMsg)(nil)

func (a *DBStateMsg) IsSameAs(b *DBStateMsg) bool {
	if a.Timestamp != b.Timestamp {
		return false
	}

	ok, err := primitives.AreBinaryMarshallablesEqual(a.DirectoryBlock, b.DirectoryBlock)
	if err != nil || ok == false {
		return false
	}

	ok, err = primitives.AreBinaryMarshallablesEqual(a.AdminBlock, b.AdminBlock)
	if err != nil || ok == false {
		return false
	}

	ok, err = primitives.AreBinaryMarshallablesEqual(a.FactoidBlock, b.FactoidBlock)
	if err != nil || ok == false {
		return false
	}

	ok, err = primitives.AreBinaryMarshallablesEqual(a.EntryCreditBlock, b.EntryCreditBlock)
	if err != nil || ok == false {
		return false
	}

	if a.EntryBlocks == nil && b.EntryBlocks != nil {
		return false
	}
	if a.EntryBlocks != nil {
		if len(a.EntryBlocks) != len(b.EntryBlocks) {
			return false
		}
		for i := range a.EntryBlocks {
			ok, err = primitives.AreBinaryMarshallablesEqual(a.EntryBlocks[i], b.EntryBlocks[i])
			if err != nil || ok == false {
				return false
			}
		}
	}

	return true
}

func (m *DBStateMsg) GetHash() interfaces.IHash {
	return nil
}

func (m *DBStateMsg) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
		data, err := m.MarshalBinary()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

func (m *DBStateMsg) Type() byte {
	return constants.DBSTATE_MSG
}

func (m *DBStateMsg) Int() int {
	return -1
}

func (m *DBStateMsg) Bytes() []byte {
	return nil
}

func (m *DBStateMsg) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *DBStateMsg) Validate(state interfaces.IState) int {

	return 1
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *DBStateMsg) Leader(state interfaces.IState) bool {
	return false
}

// Execute the leader functions of the given message
func (m *DBStateMsg) LeaderExecute(state interfaces.IState) error {
	return fmt.Errorf("Should never execute a DBState in the Leader")
}

// Returns true if this is a message for this server to execute as a follower
func (m *DBStateMsg) Follower(interfaces.IState) bool {
	return true
}

func (m *DBStateMsg) FollowerExecute(state interfaces.IState) error {
	return state.FollowerExecuteDBState(m)
}

// Acknowledgements do not go into the process list.
func (e *DBStateMsg) Process(dbheight uint32, state interfaces.IState) bool {
	panic("Ack object should never have its Process() method called")
}

func (e *DBStateMsg) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *DBStateMsg) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *DBStateMsg) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (m *DBStateMsg) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		return
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling Directory Block State Message: %v", r)
		}
	}()
	newData = data
	if newData[0] != m.Type() {
		return nil, fmt.Errorf("Invalid Message type")
	}
	newData = newData[1:]

	m.Peer2peer = true

	newData, err = m.Timestamp.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	m.DirectoryBlock = new(directoryBlock.DirectoryBlock)
	newData, err = m.DirectoryBlock.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	m.AdminBlock = new(adminBlock.AdminBlock)
	newData, err = m.AdminBlock.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	m.FactoidBlock = new(factoid.FBlock)
	newData, err = m.FactoidBlock.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	m.EntryCreditBlock = entryCreditBlock.NewECBlock()
	newData, err = m.EntryCreditBlock.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}
	for len(newData) > 0 {
		entry, data, err := entryBlock.UnmarshalEBlockData(newData)
		if err != nil {
			return nil, err
		}
		m.EntryBlocks = append(m.EntryBlocks, entry)
		newData = data
	}

	return
}

func (m *DBStateMsg) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *DBStateMsg) MarshalBinary() ([]byte, error) {
	var buf primitives.Buffer

	binary.Write(&buf, binary.BigEndian, m.Type())

	t := m.GetTimestamp()
	data, err := t.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	data, err = m.DirectoryBlock.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	data, err = m.AdminBlock.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	data, err = m.FactoidBlock.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	data, err = m.EntryCreditBlock.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)
	for i := range m.EntryBlocks {
		data, err = m.EntryBlocks[i].MarshalBinary()
		if err != nil {
			return nil, err
		}
		buf.Write(data)
	}

	return buf.DeepCopyBytes(), nil
}

func (m *DBStateMsg) String() string {
	return fmt.Sprintf("DBState: %d dblock %x admin %x fb %x ec %x",
		m.DirectoryBlock.GetHeader().GetDBHeight(),
		m.DirectoryBlock.GetKeyMR().Bytes()[:5],
		m.AdminBlock.GetHash().Bytes()[:5],
		m.FactoidBlock.GetHash().Bytes()[:5],
		m.EntryCreditBlock.GetHash().Bytes()[:5])
}

func NewDBStateMsg(timestamp interfaces.Timestamp,
	d interfaces.IDirectoryBlock,
	a interfaces.IAdminBlock,
	f interfaces.IFBlock,
	e interfaces.IEntryCreditBlock,
	ebs []interfaces.IEntryBlock) interfaces.IMsg {

	msg := new(DBStateMsg)

	msg.Peer2peer = true

	msg.Timestamp = timestamp

	msg.DirectoryBlock = d
	msg.AdminBlock = a
	msg.FactoidBlock = f
	msg.EntryCreditBlock = e
	msg.EntryBlocks = ebs

	return msg
}
