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

// Communicate a Directory Block State

type DBStateMissing struct {
	msgbase.MessageBase
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

func (m *DBStateMissing) GetRepeatHash() interfaces.IHash {
	return m.GetMsgHash()
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

// Only send the same block again after 15 seconds.
func (m *DBStateMissing) send(dbheight uint32, state interfaces.IState) (msglen int) {
	send := true

	now := state.GetTimestamp()
	sents := state.GetDBStatesSent()
	var keeps []*interfaces.DBStateSent

	for _, v := range sents {
		if now.GetTimeSeconds()-v.Sent.GetTimeSeconds() < 10 {
			if v.DBHeight == dbheight {
				send = false
			}
			keeps = append(keeps, v)
		}
	}
	if send {
		msg, err := state.LoadDBState(dbheight)
		if err != nil {
			return
		}
		if msg == nil {
			return
		}

		dbstatemsg := msg.(*DBStateMsg)
		dbstatemsg.IsInDB = false // else validateSignatures would approve it automatically
		if dbstatemsg.ValidateSignatures(state) != 1 {
			return // the last DBState we have saved may not have any or all the signatures so we can't share
		}

		b, err := msg.MarshalBinary()
		if err != nil {
			return
		}
		msglen = len(b)
		msg.SetOrigin(m.GetOrigin())
		msg.SetNetworkOrigin(m.GetNetworkOrigin())
		msg.SetNoResend(false)
		msg.SendOut(state, msg)
		state.IncDBStateAnswerCnt()
		v := new(interfaces.DBStateSent)
		v.DBHeight = dbheight
		v.Sent = now
		keeps = append(keeps, v)

		state.SetDBStatesSent(keeps)
	}
	return
}

func NewEnd(inLen int, start uint32, end uint32) (s uint32, e uint32) {
	switch {
	case inLen > constants.INMSGQUEUE_HIGH:
		return 0, 0
	case inLen > constants.INMSGQUEUE_MED && end-start > constants.DBSTATE_REQUEST_LIM_MED:
		end = start + constants.DBSTATE_REQUEST_LIM_MED
	case end-start > constants.DBSTATE_REQUEST_LIM_HIGH:
		end = start + constants.DBSTATE_REQUEST_LIM_HIGH
	}
	return start, end
}

func (m *DBStateMissing) FollowerExecute(state interfaces.IState) {
	if state.NetworkOutMsgQueue().Length() > state.NetworkOutMsgQueue().Cap()*99/100 {
		return
	}
	// TODO: Likely need to consider a limit on how many blocks we reply with.  For now,
	// just give them what they ask for.
	start := m.DBHeightStart
	end := m.DBHeightEnd

	if end == 0 {
		return
	}

	// Look at our backlog of messages from the network.  If we are really behind, ignore completely.
	// Otherwise, dial back our response, or give them as  much as we can.  In any event, limit to
	// just a bit over 1 MB
	start, end = NewEnd(state.InMsgQueue().Length(), start, end)

	sent := 0
	for dbs := start; dbs <= end && sent < 1024*1024; dbs++ {
		sent += m.send(dbs, state)
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

func (m *DBStateMissing) MarshalForSignature() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "DBStateMissing.MarshalForSignature err:%v", *pe)
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

func (m *DBStateMissing) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "DBStateMissing.MarshalBinary err:%v", *pe)
		}
	}(&err)
	return m.MarshalForSignature()
}

func (m *DBStateMissing) String() string {
	return fmt.Sprintf("DBStateMissing: %d-%d", m.DBHeightStart, m.DBHeightEnd)
}

func (m *DBStateMissing) LogFields() log.Fields {
	return log.Fields{"category": "message", "messagetype": "dbstatemissing",
		"dbheightstart": m.DBHeightStart,
		"dbheightend":   m.DBHeightEnd}
}

func NewDBStateMissing(state interfaces.IState, dbheightStart uint32, dbheightEnd uint32) interfaces.IMsg {
	msg := new(DBStateMissing)

	msg.Peer2Peer = true // Always a peer2peer request.
	msg.Timestamp = state.GetTimestamp()
	msg.DBHeightStart = dbheightStart
	msg.DBHeightEnd = dbheightEnd

	return msg
}
