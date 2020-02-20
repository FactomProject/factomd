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
	log "github.com/sirupsen/logrus"
)

var _ = fmt.Print

// SyncMsg
// If we have timed out on a message, then we sync with the other leaders on 1) the Federated Server that we
// need to replace, and 2) the audit server that we will replace it with.
type SyncMsg struct {
	msgbase.MessageBase
	TS      interfaces.Timestamp // Message Timestamp
	SigType bool                 // True if SigType message, false if DBSig
	Name    string               // Server name

	// Server that is faulting
	FedIdx uint32           // Server faulting
	FedID  interfaces.IHash // Server faulting

	// Audit server to replace faulting server
	ServerIdx  uint32           // Index of Server replacing
	ServerID   interfaces.IHash // Volunteer Server ChainID
	ServerName string           // Name of the Volunteer

	Weight      interfaces.IHash // Computed Weight at this DBHeight, Minute, Round
	DBHeight    uint32           // Directory Block Height that owns this ack
	Round       int              // Voting Round
	messageHash interfaces.IHash
}

func (m *SyncMsg) ElectionProcess(is interfaces.IState, elect interfaces.IElections) {
}

var _ interfaces.IMsg = (*SyncMsg)(nil)
var _ interfaces.IElectionMsg = (*SyncMsg)(nil)

func (a *SyncMsg) IsSameAs(msg interfaces.IMsg) bool {
	b, ok := msg.(*SyncMsg)
	if !ok {
		return false
	}
	if a.TS.GetTimeMilli() != b.TS.GetTimeMilli() {
		return false
	}
	if a.Name != b.Name {
		return false
	}
	if a.SigType != b.SigType {
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

func (m *SyncMsg) GetServerID() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "SyncMsg.GetServerID") }()

	return m.ServerID
}

func (m *SyncMsg) LogFields() log.Fields {
	return log.Fields{"category": "message", "messagetype": "FedVoteMsg", "dbheight": m.DBHeight, "newleader": m.ServerID.String()[4:12]}
}

func (m *SyncMsg) GetRepeatHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "SyncMsg.GetRepeatHash") }()

	return m.GetMsgHash()
}

// We have to return the hash of the underlying message.
func (m *SyncMsg) GetHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "SyncMsg.GetHash") }()

	return m.GetMsgHash()
}

func (m *SyncMsg) GetTimestamp() interfaces.Timestamp {
	return m.TS.Clone()
}

func (m *SyncMsg) GetMsgHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "SyncMsg.GetMsgHash") }()

	if m.MsgHash == nil {
		data, err := m.MarshalBinary()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

func (m *SyncMsg) Type() byte {
	return constants.SYNC_MSG
}

func (m *SyncMsg) Validate(state interfaces.IState) int {
	if !m.IsLocal() { // FD-886, only accept local messages
		return -1
	}
	//TODO: Must be validated
	return 1
}

func (m *SyncMsg) ElectionValidate(ie interfaces.IElections) int {
	//TODO: Must be validated
	return 1
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *SyncMsg) ComputeVMIndex(state interfaces.IState) {
}

// Execute the leader functions of the given message
// Leader, follower, do the same thing.
func (m *SyncMsg) LeaderExecute(state interfaces.IState) {
	m.FollowerExecute(state)
}

func (m *SyncMsg) FollowerExecute(is interfaces.IState) {
	s := is.(*state.State)

	var msg interfaces.IMsg
	var ack interfaces.IMsg
	if m.SigType {
		msg = messages.General.CreateMsg(constants.EOM_MSG)
		msg, ack = s.CreateEOM(true, msg, m.VMIndex)
	} else {
		msg, ack = s.CreateDBSig(m.DBHeight, m.VMIndex)
	}

	if msg == nil { // TODO: What does this mean? -- clay
		//s.Holding[m.GetMsgHash().Fixed()] = m
		s.AddToHolding(m.GetMsgHash().Fixed(), m) // SyncMsg.FollowerExecute
		return                                    // Maybe we are not yet prepared to create an SigType...
	}

	va := new(FedVoteVolunteerMsg)
	va.Missing = msg
	va.Ack = ack
	va.SetFullBroadcast(true)
	va.FedIdx = m.FedIdx
	va.FedID = m.FedID

	va.ServerIdx = uint32(m.ServerIdx)
	va.ServerID = m.ServerID
	va.ServerName = m.ServerName

	va.VMIndex = m.VMIndex
	va.TS = primitives.NewTimestampNow()
	va.Name = m.Name
	va.Weight = m.Weight
	va.DBHeight = m.DBHeight
	va.Minute = m.Minute
	va.Round = m.Round
	va.SigType = m.SigType

	va.Sign(is)

	va.SendOut(is, va)
	va.FollowerExecute(is)
}

// Acknowledgements do not go into the process list.
func (e *SyncMsg) Process(dbheight uint32, state interfaces.IState) bool {
	panic("Ack object should never have its Process() method called")
}

func (e *SyncMsg) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *SyncMsg) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (m *SyncMsg) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	err = fmt.Errorf("SyncMsg is an internal message only")
	return
}

func (m *SyncMsg) UnmarshalBinary(data []byte) error {
	return fmt.Errorf("SyncMsg is an internal message only")
}

func (m *SyncMsg) MarshalBinary() (data []byte, err error) {

	var buf primitives.Buffer

	if e := buf.PushByte(constants.SYNC_MSG); e != nil {
		return nil, e
	}
	if e := buf.PushTimestamp(m.TS); e != nil {
		return nil, e
	}
	if e := buf.PushBool(m.SigType); e != nil {
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
	return buf.Bytes(), nil
}

func (m *SyncMsg) String() string {
	if m.LeaderChainID == nil {
		m.LeaderChainID = primitives.NewZeroHash()
	}
	return fmt.Sprintf("%s %10s ID: %x WT: %x serverIdx: %d vmIdx: %d round: %d dbheight: %d minute: %d ",
		"Sync Message",
		m.Name,
		m.ServerID.Bytes()[2:5],
		m.Weight.Bytes()[2:5],
		m.ServerIdx,
		m.VMIndex,
		m.Round,
		m.DBHeight,
		m.Minute)
}
