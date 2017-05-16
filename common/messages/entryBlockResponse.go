// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

//Requests entry blocks from a range of DBlocks

type EntryBlockResponse struct {
	MessageBase
	Timestamp interfaces.Timestamp

	EBlockCount uint32
	EBlocks     []interfaces.IEntryBlock
	EntryCount  uint32
	Entries     []interfaces.IEBEntry

	//Not signed!
}

var _ interfaces.IMsg = (*EntryBlockResponse)(nil)

func (a *EntryBlockResponse) IsSameAs(b *EntryBlockResponse) bool {
	if b == nil {
		return false
	}
	if a.Timestamp.GetTimeMilli() != b.Timestamp.GetTimeMilli() {
		return false
	}
	if a.EBlockCount != b.EBlockCount {
		return false
	}
	if a.EntryCount != b.EntryCount {
		return false
	}

	if len(a.EBlocks) != len(b.EBlocks) {
		return false
	}

	if len(a.Entries) != len(b.Entries) {
		return false
	}

	//TODO: check blocks and entries

	return true
}

func (m *EntryBlockResponse) GetRepeatHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *EntryBlockResponse) GetHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *EntryBlockResponse) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
		data, err := m.MarshalBinary()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

func (m *EntryBlockResponse) Type() byte {
	return constants.ENTRY_BLOCK_RESPONSE
}

func (m *EntryBlockResponse) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *EntryBlockResponse) Validate(state interfaces.IState) int {
	if m.EBlockCount != uint32(len(m.EBlocks)) {
		return -1
	}
	if m.EntryCount != uint32(len(m.Entries)) {
		return -1
	}

	return 1
}

func (m *EntryBlockResponse) ComputeVMIndex(state interfaces.IState) {
}

// Execute the leader functions of the given message
func (m *EntryBlockResponse) LeaderExecute(state interfaces.IState) {
	m.FollowerExecute(state)
}

func (m *EntryBlockResponse) FollowerExecute(state interfaces.IState) {
	if state.NetworkOutMsgQueue().Length() > 1000 {
		return
	}

	db := state.GetAndLockDB()
	defer state.UnlockDB()

	db.StartMultiBatch()
	for _, v := range m.EBlocks {
		err := db.ProcessEBlockMultiBatchWithoutHead(v, true)
		if err != nil {
			panic(err)
		}
	}
	for _, v := range m.Entries {
		err := db.InsertEntryMultiBatch(v)
		if err != nil {
			panic(err)
		}
	}
	err := db.ExecuteMultiBatch()
	if err != nil {
		panic(err)
	}

	return
}

// Acknowledgements do not go into the process list.
func (e *EntryBlockResponse) Process(dbheight uint32, state interfaces.IState) bool {
	panic("Ack object should never have its Process() method called")
}

func (e *EntryBlockResponse) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *EntryBlockResponse) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (m *EntryBlockResponse) UnmarshalBinaryData(data []byte) ([]byte, error) {
	buf := primitives.NewBuffer(data)

	t, err := buf.PopByte()
	if err != nil {
		return nil, err
	}
	if t != m.Type() {
		return nil, fmt.Errorf("Invalid Message type")
	}

	m.Peer2Peer = true // This is always a Peer2peer message

	m.Timestamp = new(primitives.Timestamp)
	err = buf.PopBinaryMarshallable(m.Timestamp)
	if err != nil {
		return nil, err
	}

	m.EBlockCount, err = buf.PopUInt32()

	for i := 0; i < int(m.EBlockCount); i++ {
		eBlock := entryBlock.NewEBlock()
		err = buf.PopBinaryMarshallable(eBlock)
		if err != nil {
			return nil, err
		}
		m.EBlocks = append(m.EBlocks, eBlock)
	}

	m.EntryCount, err = buf.PopUInt32()
	if err != nil {
		return nil, err
	}

	for i := 0; i < int(m.EntryCount); i++ {
		entry := entryBlock.NewEntry()
		err = buf.PopBinaryMarshallable(entry)
		if err != nil {
			return nil, err
		}
		m.Entries = append(m.Entries, entry)
	}

	return buf.DeepCopyBytes(), nil
}

func (m *EntryBlockResponse) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *EntryBlockResponse) MarshalForSignature() ([]byte, error) {
	buf := primitives.NewBuffer(nil)

	err := buf.PushByte(m.Type())
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(m.GetTimestamp())
	if err != nil {
		return nil, err
	}

	m.EBlockCount = uint32(len(m.EBlocks))
	err = buf.PushUInt32(m.EBlockCount)
	if err != nil {
		return nil, err
	}
	for _, eb := range m.EBlocks {
		err = buf.PushBinaryMarshallable(eb)
		if err != nil {
			return nil, err
		}
	}

	m.EntryCount = uint32(len(m.Entries))
	err = buf.PushUInt32(m.EntryCount)
	for _, e := range m.Entries {
		err = buf.PushBinaryMarshallable(e)
		if err != nil {
			return nil, err
		}
	}

	return buf.DeepCopyBytes(), nil
}

func (m *EntryBlockResponse) MarshalBinary() ([]byte, error) {
	return m.MarshalForSignature()
}

func (m *EntryBlockResponse) String() string {
	str, _ := m.JSONString()
	return str
}

func NewEntryBlockResponse(state interfaces.IState) interfaces.IMsg {
	msg := new(EntryBlockResponse)

	msg.Peer2Peer = true // Always a peer2peer request.
	msg.Timestamp = state.GetTimestamp()

	return msg
}
