// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"bytes"
//	"encoding/binary"
	"fmt"
	"encoding/binary"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	
)

// Communicate a Directory Block State

type DBStateMissing struct {
	MessageBase
	Timestamp   		interfaces.Timestamp
		
	DBHeightStart		uint32	// First block missing
	DBHeightEnd			uint32	// Last block missing.
	
}

var _ interfaces.IMsg = (*DBStateMissing)(nil)

func (m *DBStateMissing) IsSameAs(b *DBStateMissing) bool {
	return true
}

func (m *DBStateMissing) GetHash() interfaces.IHash {
	return nil
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

func (m *DBStateMissing) Type() int {
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
func (m *DBStateMissing) Validate(dbheight uint32, state interfaces.IState) int {
	return 1
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *DBStateMissing) Leader(state interfaces.IState) bool {
	return false
}

// Execute the leader functions of the given message
func (m *DBStateMissing) LeaderExecute(state interfaces.IState) error {
	return fmt.Errorf("Should never execute a DBState in the Leader")
}

// Returns true if this is a message for this server to execute as a follower
func (m *DBStateMissing) Follower(interfaces.IState) bool {
	return true
}

func (m *DBStateMissing) FollowerExecute(state interfaces.IState) error {
	
	end := m.DBHeightStart+100		// Process 100 at a time.
	if end > m.DBHeightEnd {
		end = m.DBHeightEnd
	}
	
	for dbs := m.DBHeightStart; dbs <= end; dbs++ {
		msg,_ := state.LoadDBState(dbs)
		msg.SetOrigin(m.GetOrigin())
		state.NetworkOutMsgQueue() <- msg
	}
	
	return nil
}

// Acknowledgements do not go into the process list.
func (e *DBStateMissing) Process(dbheight uint32, state interfaces.IState) {
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
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	
	m.Peer2peer = true			// This is always a Peer2peer message
	
	newData = data[1:]			// Skip our type;  Someone else's problem.
	
	newData, err = m.Timestamp.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	m.DBHeightStart, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]
	m.DBHeightEnd,   newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]
	
	return
}

func (m *DBStateMissing) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *DBStateMissing) MarshalForSignature() ([]byte, error) {

	var buf bytes.Buffer
	
	binary.Write(&buf, binary.BigEndian, byte(m.Type()))
	
	t := m.GetTimestamp()
	data, err := t.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	binary.Write(&buf, binary.BigEndian, m.DBHeightStart)
	binary.Write(&buf, binary.BigEndian, m.DBHeightEnd)
		
	return buf.Bytes(), nil
}

func (m *DBStateMissing) MarshalBinary() ([]byte, error) {
	return m.MarshalForSignature()
}

func (m *DBStateMissing) String() string {
	return fmt.Sprintf("DBStateMissing: %d-%d",m.DBHeightStart,m.DBHeightEnd)
}

func NewDBStateMissing(state interfaces.IState, dbheightStart uint32, dbheightEnd uint32 ) interfaces.IMsg {

	msg := new(DBStateMissing)
	
	msg.Peer2peer = true					// Always a peer2peer request.
	msg.Timestamp = state.GetTimestamp()
	msg.DBHeightStart = dbheightStart
	msg.DBHeightEnd = dbheightEnd
	
	return msg
}