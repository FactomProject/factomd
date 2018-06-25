// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"

	"github.com/FactomProject/factomd/common/messages/msgbase"
	log "github.com/sirupsen/logrus"
)

//Requests entry blocks from a range of DBlocks

type MissingEntryBlocks struct {
	msgbase.MessageBase
	Timestamp interfaces.Timestamp

	DBHeightStart uint32 // First block missing
	DBHeightEnd   uint32 // Last block missing.

	//Not signed!
}

var _ interfaces.IMsg = (*MissingEntryBlocks)(nil)

func (a *MissingEntryBlocks) IsSameAs(b *MissingEntryBlocks) bool {
	if b == nil {
		return false
	}
	if a.Timestamp.GetTimeMilli() != b.Timestamp.GetTimeMilli() {
		return false
	}
	if a.DBHeightStart != b.DBHeightStart {
		return false
	}
	if a.DBHeightEnd != b.DBHeightEnd {
		return false
	}

	return true
}

func (m *MissingEntryBlocks) GetRepeatHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *MissingEntryBlocks) GetHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *MissingEntryBlocks) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
		data, err := m.MarshalBinary()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

func (m *MissingEntryBlocks) Type() byte {
	return constants.MISSING_ENTRY_BLOCKS
}

func (m *MissingEntryBlocks) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *MissingEntryBlocks) Validate(state interfaces.IState) int {
	if m.DBHeightStart > m.DBHeightEnd {
		return -1
	}

	// ignore it if it's from the future
	if m.DBHeightStart > state.GetHighestKnownBlock() {
		return -1
	}
	// if they are asking for too many DBStates in one request then toss the request
	if m.DBHeightEnd - m.DBHeightStart > constants.MAX_EB_PER_REQUEST {
		return -1
	}


	return 1
}

func (m *MissingEntryBlocks) ComputeVMIndex(state interfaces.IState) {
}

// Execute the leader functions of the given message
func (m *MissingEntryBlocks) LeaderExecute(state interfaces.IState) {
	m.FollowerExecute(state)
}

func (m *MissingEntryBlocks) FollowerExecute(state interfaces.IState) {
	if state.NetworkOutMsgQueue().Length() > state.NetworkOutMsgQueue().Cap()*99/100 {
		return
	}
	start := m.DBHeightStart
	end := m.DBHeightEnd

	if end-start > constants.MAX_EB_PER_REQUEST {
		end = start + constants.MAX_EB_PER_REQUEST
	}

	if end > state.GetHighestKnownBlock() {
		end = state.GetHighestKnownBlock()
	}
	db := state.GetDB()

	resp := NewEntryBlockResponse(state).(*EntryBlockResponse)

	for i := start; i <= end; i++ {
		dblk, err := db.FetchDBlockByHeight(i)
		if err != nil {
			return
		}
		if dblk == nil {
			return
		}
		for _, v := range dblk.GetDBEntries() {
			if v.GetChainID().IsMinuteMarker() == true {
				continue
			}
			eBlock, err := db.FetchEBlock(v.GetKeyMR())
			if err != nil {
				return
			}
			resp.EBlocks = append(resp.EBlocks, eBlock)
			for _, v := range eBlock.GetBody().GetEBEntries() {
				entry, err := db.FetchEntry(v)
				if err != nil {
					return
				}
				resp.Entries = append(resp.Entries, entry)
			}
		}
	}

	resp.SetOrigin(m.GetOrigin())
	resp.SetNetworkOrigin(m.GetNetworkOrigin())
	resp.SendOut(state, resp)
	state.IncDBStateAnswerCnt()

	return
}

// Acknowledgements do not go into the process list.
func (e *MissingEntryBlocks) Process(dbheight uint32, state interfaces.IState) bool {
	panic("Ack object should never have its Process() method called")
}

func (e *MissingEntryBlocks) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *MissingEntryBlocks) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (m *MissingEntryBlocks) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling Directory Block State Missing Message: %v", r)
		}
	}()
	newData = data
	if newData[0] != m.Type() {
		return nil, fmt.Errorf("Invalid Message type")
	}
	newData = newData[1:]

	m.Peer2Peer = true // This is always a Peer2peer message

	m.Timestamp = new(primitives.Timestamp)
	newData, err = m.Timestamp.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	m.DBHeightStart, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]
	m.DBHeightEnd, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]

	return
}

func (m *MissingEntryBlocks) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *MissingEntryBlocks) MarshalForSignature() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "MissingEntryBlocks.MarshalForSignature err:%v", *pe)
		}
	}(&err)
	var buf primitives.Buffer

	binary.Write(&buf, binary.BigEndian, m.Type())

	t := m.GetTimestamp()
	data, err := t.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	binary.Write(&buf, binary.BigEndian, m.DBHeightStart)
	binary.Write(&buf, binary.BigEndian, m.DBHeightEnd)

	return buf.DeepCopyBytes(), nil
}

func (m *MissingEntryBlocks) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "MissingEntryBlocks.MarshalBinary err:%v", *pe)
		}
	}(&err)
	return m.MarshalForSignature()
}

func (m *MissingEntryBlocks) String() string {
	return fmt.Sprintf("MissingEntryBlocks: %d-%d", m.DBHeightStart, m.DBHeightEnd)
}

func (m *MissingEntryBlocks) LogFields() log.Fields {
	return log.Fields{"category": "message", "messagetype": "missingentryblocks",
		"dbheightstart": m.DBHeightStart,
		"dbheightend":   m.DBHeightEnd}
}

func NewMissingEntryBlocks(state interfaces.IState, dbheightStart uint32, dbheightEnd uint32) interfaces.IMsg {
	msg := new(MissingEntryBlocks)

	msg.Peer2Peer = true // Always a peer2peer request.
	msg.Timestamp = state.GetTimestamp()
	msg.DBHeightStart = dbheightStart
	msg.DBHeightEnd = dbheightEnd

	return msg
}
