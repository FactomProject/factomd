// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package electionMsgs

import (
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/messages/msgbase"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/goleveldb/leveldb/errors"
	log "github.com/FactomProject/logrus"
)

var _ = fmt.Print

// We send this message out twice.  The first time, with Majority == false.  Because
// we only see the volunteer message, and have not seen a majority of the leaders respond
// that they have seen the volunteer message.
//
// Then we send this message out again, once we have seen a majority of the leaders claim
// that they have seen the volunteer message, and accept the volunteer as a replacement.
//
// Once we have seen a majority of the leaders claiming to have seen a majority of the
// the leaders accept the volunteer, then replace the faulted leader with the new leader.

type LeaderAck struct {
	msgbase.MessageBase
	TS          interfaces.Timestamp // Message Timestamp
	Majority    bool                 // Indicates a majority of Leaders agree
	Name        string               // Server name
	ServerIdx   uint32               // Index of Server replacing
	ServerID    interfaces.IHash     // Volunteer Server ChainID
	Weight      interfaces.IHash     // Computed Weight at this DBHeight, Minute, Round
	DBHeight    uint32               // Directory Block Height that owns this ack
	Minute      byte                 // Minute (-1 for dbsig)
	Round       int                  // Voting Round
	messageHash interfaces.IHash
}

func (m *LeaderAck) ElectionProcess(is interfaces.IState, elect interfaces.IElections) {
	fmt.Printf("eee %10s %s\n", is.GetFactomNodeName(), m.String())
}

var _ interfaces.IMsg = (*LeaderAck)(nil)

func (a *LeaderAck) IsSameAs(msg interfaces.IMsg) bool {
	b, ok := msg.(*LeaderAck)
	if !ok {
		return false
	}
	if a.TS.GetTimeMilli() != b.TS.GetTimeMilli() {
		return false
	}
	if a.Name != b.Name {
		return false
	}
	if a.Majority != b.Majority {
		return false
	}
	if a.ServerIdx != b.ServerIdx {
		return false
	}
	if a.ServerID.Fixed() != b.ServerID.Fixed() {
		return false
	}
	if a.Weight.Fixed() != b.Weight.Fixed() {
		return false
	}
	if a.DBHeight != b.DBHeight {
		return false
	}
	if a.VMIndex != b.VMIndex {
		return false
	}
	if a.Round != b.Round {
		return false
	}
	if a.Minute != b.Minute {
		return false
	}
	return true
}

func (m *LeaderAck) GetServerID() interfaces.IHash {
	return m.ServerID
}

func (m *LeaderAck) LogFields() log.Fields {
	return log.Fields{"category": "message", "messagetype": "VolunteerAudit", "dbheight": m.DBHeight, "newleader": m.ServerID.String()[4:12]}
}

func (m *LeaderAck) GetRepeatHash() interfaces.IHash {
	return m.GetMsgHash()
}

// We have to return the haswh of the underlying message.
func (m *LeaderAck) GetHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *LeaderAck) GetTimestamp() interfaces.Timestamp {
	return m.TS
}

func (m *LeaderAck) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
		data, err := m.MarshalBinary()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

func (m *LeaderAck) Type() byte {
	return constants.LEADER_ACK_VOLUNTEER
}

func (m *LeaderAck) Validate(state interfaces.IState) int {
	return 1
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *LeaderAck) ComputeVMIndex(state interfaces.IState) {
}

// Execute the leader functions of the given message
// Leader, follower, do the same thing.
func (m *LeaderAck) LeaderExecute(state interfaces.IState) {
	m.FollowerExecute(state)
}

func (m *LeaderAck) FollowerExecute(is interfaces.IState) {
	fmt.Printf("eee %10s %20s %s\n",
		is.GetFactomNodeName(),
		"Leader Ack Volunteer Message",
		m.String())
	s := is.(*state.State)

	eom := messages.General.CreateMsg(constants.EOM_MSG)
	eom, _ = s.CreateEOM(eom, m.VMIndex)

	va := new(VolunteerAudit)
	va.VMIndex = m.VMIndex
	va.TS = primitives.NewTimestampNow()
	va.Name = m.Name
	va.ServerIdx = uint32(m.ServerIdx)
	va.ServerID = m.ServerID
	va.Weight = m.Weight
	va.DBHeight = m.DBHeight
	va.Minute = m.Minute
	va.Round = m.Round
	fmt.Printf("eee %10s %20s %s\n", is.GetFactomNodeName(), "I'm an Audit Server Volunteer!", va.String())
	va.SendOut(is, va)
	is.ElectionsQueue().Enqueue(va)
}

// Acknowledgements do not go into the process list.
func (e *LeaderAck) Process(dbheight uint32, state interfaces.IState) bool {
	panic("Ack object should never have its Process() method called")
}

func (e *LeaderAck) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *LeaderAck) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (m *LeaderAck) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		return
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()

	buf := primitives.NewBuffer(data)

	if t, e := buf.PopByte(); e != nil || t != constants.LEADER_ACK_VOLUNTEER {
		return nil, errors.New("Not a Leader Ack Message type")
	}
	if m.TS, err = buf.PopTimestamp(); err != nil {
		return nil, err
	}
	if m.Majority, err = buf.PopBool(); err != nil {
		return nil, err
	}
	if m.Name, err = buf.PopString(); err != nil {
		return nil, err
	}
	if m.ServerIdx, err = buf.PopUInt32(); err != nil {
		return nil, err
	}
	if m.ServerID, err = buf.PopIHash(); err != nil {
		return nil, err
	}
	if m.Weight, err = buf.PopIHash(); err != nil {
		return nil, err
	}
	if m.DBHeight, err = buf.PopUInt32(); err != nil {
		return nil, err
	}
	if m.VMIndex, err = buf.PopInt(); err != nil {
		return nil, err
	}
	if m.Round, err = buf.PopInt(); err != nil {
		return nil, err
	}
	if m.Minute, err = buf.PopByte(); err != nil {
		return nil, err
	}
	return buf.PopBytes()
}

func (m *LeaderAck) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *LeaderAck) MarshalBinary() (data []byte, err error) {

	var buf primitives.Buffer

	if e := buf.PushByte(constants.LEADER_ACK_VOLUNTEER); e != nil {
		return nil, e
	}
	if e := buf.PushTimestamp(m.TS); e != nil {
		return nil, e
	}
	if e := buf.PushBool(m.Majority); e != nil {
		return nil, e
	}
	if e := buf.PushString(m.Name); e != nil {
		return nil, e
	}
	if e := buf.PushUInt32(m.ServerIdx); e != nil {
		return nil, e
	}
	if e := buf.PushIHash(m.ServerID); e != nil {
		return nil, e
	}
	if e := buf.PushIHash(m.Weight); e != nil {
		return nil, e
	}
	if e := buf.PushUInt32(m.DBHeight); e != nil {
		return nil, e
	}
	if e := buf.PushInt(m.VMIndex); e != nil {
		return nil, e
	}
	if e := buf.PushInt(m.Round); e != nil {
		return nil, e
	}
	if e := buf.PushByte(m.Minute); e != nil {
		return nil, e
	}
	return buf.DeepCopyBytes(), nil
}

func (m *LeaderAck) String() string {
	if m.LeaderChainID == nil {
		m.LeaderChainID = primitives.NewZeroHash()
	}
	return fmt.Sprintf("%s %10s ID: %x WT: %x serverIdx: %d vmIdx: %d round: %d dbheight: %d minute: %d ",
		"Leader Ack",
		m.Name,
		m.ServerID.Bytes()[2:5],
		m.Weight.Bytes()[2:5],
		m.ServerIdx,
		m.VMIndex,
		m.Round,
		m.DBHeight,
		m.Minute)
}
