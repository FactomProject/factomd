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

	EBlocks []interfaces.IEntryBlock
	Entries []interfaces.IEBEntry

	//Not signed!
}

var _ interfaces.IMsg = (*DBStateMsg)(nil)

func (a *DBStateMsg) IsSameAs(b *DBStateMsg) bool {
	if b == nil {
		return false
	}
	if a.Timestamp.GetTimeMilli() != b.Timestamp.GetTimeMilli() {
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

	//TODO: compare more

	return true
}

func (m *DBStateMsg) GetRepeatHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *DBStateMsg) GetHash() interfaces.IHash {
	data, _ := m.MarshalBinary()
	return primitives.Sha(data)
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

func (m *DBStateMsg) ComputeVMIndex(state interfaces.IState) {}

// Execute the leader functions of the given message
func (m *DBStateMsg) LeaderExecute(state interfaces.IState) {
	m.FollowerExecute(state)
}

func (m *DBStateMsg) FollowerExecute(state interfaces.IState) {
	state.FollowerExecuteDBState(m)
}

// Acknowledgements do not go into the process list.
func (e *DBStateMsg) Process(dbheight uint32, state interfaces.IState) bool {
	panic("DBStatemsg should never have its Process() method called")
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
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling Directory Block State Message: %v", r)
		}
	}()
	newData = data
	if newData[0] != m.Type() {
		return nil, fmt.Errorf("Invalid Message type")
	}
	newData = newData[1:]

	m.Peer2Peer = true

	m.Timestamp = new(primitives.Timestamp)
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

	EBlockCount, newData := binary.BigEndian.Uint32(newData[0:4]), newData[4:]

	for i := 0; i < int(EBlockCount); i++ {
		eBlock := entryBlock.NewEBlock()
		newData, err = eBlock.UnmarshalBinaryData(newData)
		if err != nil {
			return nil, err
		}
		m.EBlocks = append(m.EBlocks, eBlock)
	}

	EntryCount, newData := binary.BigEndian.Uint32(newData[0:4]), newData[4:]

	for i := 0; i < int(EntryCount); i++ {
		entry := entryBlock.NewEntry()
		newData, err = entry.UnmarshalBinaryData(newData)
		if err != nil {
			return nil, err
		}
		m.Entries = append(m.Entries, entry)
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

	EBlockCount := uint32(len(m.EBlocks))
	binary.Write(&buf, binary.BigEndian, EBlockCount)
	for _, eb := range m.EBlocks {
		bin, err := eb.MarshalBinary()
		if err != nil {
			return nil, err
		}
		buf.Write(bin)
	}

	EntryCount := uint32(len(m.Entries))
	binary.Write(&buf, binary.BigEndian, EntryCount)
	for _, e := range m.Entries {
		bin, err := e.MarshalBinary()
		if err != nil {
			return nil, err
		}
		buf.Write(bin)
	}

	return buf.DeepCopyBytes(), nil
}

func (m *DBStateMsg) String() string {
	return fmt.Sprintf("DBState: ht:%3d dblock %6x admin %6x fb %6x ec %6x hash %6x",
		m.DirectoryBlock.GetHeader().GetDBHeight(),
		m.DirectoryBlock.GetKeyMR().Bytes()[:3],
		m.AdminBlock.GetHash().Bytes()[:3],
		m.FactoidBlock.GetHash().Bytes()[:3],
		m.EntryCreditBlock.GetHash().Bytes()[:3],
		m.GetHash().Bytes()[:3])
}

func NewDBStateMsg(timestamp interfaces.Timestamp,
	d interfaces.IDirectoryBlock,
	a interfaces.IAdminBlock,
	f interfaces.IFBlock,
	e interfaces.IEntryCreditBlock,
	eBlocks []interfaces.IEntryBlock,
	entries []interfaces.IEBEntry) interfaces.IMsg {

	msg := new(DBStateMsg)

	msg.Peer2Peer = true

	msg.Timestamp = timestamp

	msg.DirectoryBlock = d
	msg.AdminBlock = a
	msg.FactoidBlock = f
	msg.EntryCreditBlock = e

	msg.EBlocks = eBlocks
	msg.Entries = entries

	return msg
}
