// Copyright 2017 Factom Foundation
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

	SignatureList SigList

	//Not marshalled
	Sent   interfaces.Timestamp
	IsInDB bool
}

var _ interfaces.IMsg = (*DBStateMsg)(nil)

func (a *DBStateMsg) IsSameAs(b *DBStateMsg) bool {
	defer func() {
		if r := recover(); r != nil {
			return
		}
	}()

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

	if len(a.EBlocks) != len(b.EBlocks) || (len(a.Entries) != len(b.Entries)) {
		return false
	}

	for i := range a.EBlocks {
		ok, err = primitives.AreBinaryMarshallablesEqual(a.EBlocks[i], b.EBlocks[i])
		if err != nil || ok == false {
			return false
		}
	}

	for i := range a.Entries {
		ok, err = primitives.AreBinaryMarshallablesEqual(a.Entries[i], b.Entries[i])
		if err != nil || ok == false {
			return false
		}
	}

	return true
}

func (m *DBStateMsg) GetRepeatHash() interfaces.IHash {
	return m.DirectoryBlock.GetHash()
}

func (m *DBStateMsg) GetHash() interfaces.IHash {

	//	data, _ := m.MarshalBinary()
	//	return primitives.Sha(data)

	// These two calls do the same thing.  It should be standardized.
	return m.GetMsgHash()
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

	// No matter what, a block has to have what a block has to have.
	if m.DirectoryBlock == nil || m.AdminBlock == nil || m.FactoidBlock == nil || m.EntryCreditBlock == nil {
		state.AddStatus(fmt.Sprintf("DBStateMsg.Validate() Fail  Doesn't have all the blocks ht: %d", m.DirectoryBlock.GetHeader().GetDBHeight()))
		//We need the basic block types
		return -1
	}

	if m.IsInDB {
		return 1
	}

	dbheight := m.DirectoryBlock.GetHeader().GetDBHeight()

	// Just accept the genesis block
	if dbheight == 0 {
		return 1
	}

	if state.GetNetworkID() != m.DirectoryBlock.GetHeader().GetNetworkID() {
		state.AddStatus(fmt.Sprintf("DBStateMsg.Validate() Fail  ht: %d Expecting NetworkID %x and found %x",
			dbheight, state.GetNetworkID(), m.DirectoryBlock.GetHeader().GetNetworkID()))
		//Wrong network ID
		return -1
	}

	diff := int(dbheight) - (int(state.GetHighestSavedBlk())) // Difference from the working height (completed+1)

	// Look at saved heights if not too far from what we have saved.
	if diff < -1 {
		state.AddStatus(fmt.Sprintf("DBStateMsg.Validate() Fail dbstate dbht: %d Highest Saved %d diff %d",
			dbheight, state.GetHighestSavedBlk(), diff))
		return -1
	}

	if m.DirectoryBlock.GetHeader().GetNetworkID() == constants.MAIN_NETWORK_ID {
		key := constants.CheckPoints[dbheight]
		if key != "" {
			if key != m.DirectoryBlock.DatabasePrimaryIndex().String() {
				state.AddStatus(fmt.Sprintf("DBStateMsg.Validate() Fail  ht: %d checkpoint failure. Had %s Expected %s",
					dbheight, m.DirectoryBlock.DatabasePrimaryIndex().String(), key))
				//Key does not match checkpoint
				return -1
			}
		}
	}

	return 1
}

func (m *DBStateMsg) SigTally(state interfaces.IState) int {
	dbheight := m.DirectoryBlock.GetHeader().GetDBHeight()

	validSigCount := 0

	data, err := m.DirectoryBlock.GetHeader().MarshalBinary()
	if err != nil {
		state.AddStatus(fmt.Sprint("Debug: DBState Signature Error, Marshal binary errored"))
		return validSigCount
	}

	for _, sig := range m.SignatureList.List {
		//Check signature against the Skeleton key
		authoritativeKey := state.GetNetworkBootStrapKey()
		if authoritativeKey != nil {
			if bytes.Compare(sig.GetKey(), authoritativeKey.Bytes()) == 0 {
				if sig.Verify(data) {
					validSigCount++
				}
			}
		}

		check, err := state.VerifyAuthoritySignature(data, sig.GetSignature(), dbheight)
		if err == nil && check >= 0 {
			validSigCount++
		}
	}

	return validSigCount
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

	eBlockCount, newData := binary.BigEndian.Uint32(newData[0:4]), newData[4:]

	for i := uint32(0); i < eBlockCount; i++ {
		eBlock := entryBlock.NewEBlock()
		newData, err = eBlock.UnmarshalBinaryData(newData)
		if err != nil {
			panic(err.Error())
		}
		m.EBlocks = append(m.EBlocks, eBlock)
	}

	entryCount, newData := binary.BigEndian.Uint32(newData[0:4]), newData[4:]

	for i := uint32(0); i < entryCount; i++ {
		var entrySize uint32
		entrySize, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]
		entry := entryBlock.NewEntry()
		newData, err = newData[int(entrySize):], entry.UnmarshalBinary(newData[:int(entrySize)])
		if err != nil {
			panic(err.Error())
		}
		m.Entries = append(m.Entries, entry)
	}

	newData, err = m.SignatureList.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
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

	eBlockCount := uint32(len(m.EBlocks))
	binary.Write(&buf, binary.BigEndian, eBlockCount)
	for _, eb := range m.EBlocks {
		bin, err := eb.MarshalBinary()
		if err != nil {
			return nil, err
		}
		buf.Write(bin)
	}

	entryCount := uint32(len(m.Entries))
	binary.Write(&buf, binary.BigEndian, entryCount)
	for _, e := range m.Entries {
		bin, err := e.MarshalBinary()
		if err != nil || bin == nil || len(bin) == 0 {
			return nil, err
		}
		entrySize := uint32(len(bin))
		binary.Write(&buf, binary.BigEndian, entrySize)
		buf.Write(bin)
	}

	if d, err := m.SignatureList.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}

	return buf.DeepCopyBytes(), nil
}

func (m *DBStateMsg) String() string {
	data, _ := m.MarshalBinary()
	return fmt.Sprintf("DBState: dbht:%3d [size: %11s] dblock %6x admin %6x fb %6x ec %6x hash %6x",
		m.DirectoryBlock.GetHeader().GetDBHeight(),
		primitives.AddCommas(int64(len(data))),
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
	entries []interfaces.IEBEntry,
	sigList []interfaces.IFullSignature) interfaces.IMsg {

	msg := new(DBStateMsg)
	msg.NoResend = true

	msg.Peer2Peer = true

	msg.Timestamp = timestamp

	msg.DirectoryBlock = d
	msg.AdminBlock = a
	msg.FactoidBlock = f
	msg.EntryCreditBlock = e

	msg.EBlocks = eBlocks
	msg.Entries = entries

	sl := new(SigList)
	sl.Length = uint32(len(sigList))
	sl.List = sigList

	msg.SignatureList = *sl

	return msg
}
