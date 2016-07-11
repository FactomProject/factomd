// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"bytes"
	//	"encoding/binary"
	"encoding/binary"
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// Communicate a Directory Block State

type DBStateMissing struct {
	MessageBase
	Timestamp interfaces.Timestamp

	DBHeightStart uint32 // First block missing
	DBHeightEnd   uint32 // Last block missing.

	//Not signed!
}

var _ interfaces.IMsg = (*DBStateMissing)(nil)

func (a *DBStateMissing) IsSameAs(b *DBStateMissing) bool {
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

func (m *DBStateMissing) GetHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *DBStateMissing) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
		data, err := m.MarshalBinary()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

func (m *DBStateMissing) Type() byte {
	return constants.DBSTATE_MISSING_MSG
}

func (m *DBStateMissing) Int() int {
	return -1
}

func (m *DBStateMissing) Bytes() []byte {
	return nil
}

func (m *DBStateMissing) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *DBStateMissing) Validate(state interfaces.IState) int {
	if m.DBHeightStart > m.DBHeightEnd {
		return -1
	}
	return 1
}

func (m *DBStateMissing) ComputeVMIndex(state interfaces.IState) {

}

// Execute the leader functions of the given message
func (m *DBStateMissing) LeaderExecute(state interfaces.IState) {
	m.FollowerExecute(state)
}

func (m *DBStateMissing) FollowerExecute(state interfaces.IState) {
	if len(state.NetworkOutMsgQueue()) > 1000 {
		return
	}

	// TODO: Likely need to consider a limit on how many blocks we reply with.  For now,
	// just give them what they ask for.
	start := m.DBHeightStart
	end := m.DBHeightEnd
	if end-start > 10 {
		end = start + 10
	}
	for dbs := start; dbs <= end; dbs++ {
		msg, err := state.LoadDBState(dbs)
		if msg != nil && err == nil { // If I don't have this block, ignore.
			msg.SetOrigin(m.GetOrigin())
			state.NetworkOutMsgQueue() <- msg
		}
	}

	return
}

// Acknowledgements do not go into the process list.
func (e *DBStateMissing) Process(dbheight uint32, state interfaces.IState) bool {
	panic("Ack object should never have its Process() method called")
}

func (e *DBStateMissing) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *DBStateMissing) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *DBStateMissing) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (m *DBStateMissing) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
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

func (m *DBStateMissing) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *DBStateMissing) MarshalForSignature() ([]byte, error) {
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

func (m *DBStateMissing) MarshalBinary() ([]byte, error) {
	return m.MarshalForSignature()
}

func (m *DBStateMissing) String() string {
	return fmt.Sprintf("DBStateMissing: %d-%d", m.DBHeightStart, m.DBHeightEnd)
}

func NewDBStateMissing(state interfaces.IState, dbheightStart uint32, dbheightEnd uint32) interfaces.IMsg {
	msg := new(DBStateMissing)

	msg.Peer2Peer = true // Always a peer2peer request.
	msg.Timestamp = state.GetTimestamp()
	msg.DBHeightStart = dbheightStart
	msg.DBHeightEnd = dbheightEnd

	return msg
}
