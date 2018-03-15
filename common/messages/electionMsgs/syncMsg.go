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
	log "github.com/sirupsen/logrus"
)

var _ = fmt.Print

// SyncMsg
// If we have timed out on a message, then we sync with the other leaders on 1) the Federated Server that we
// need to replace, and 2) the audit server that we will replace it with.
type SyncMsg struct {
	msgbase.MessageBase
	TS   interfaces.Timestamp // Message Timestamp
	EOM  bool                 // True if EOM message, false if DBSig
	Name string               // Server name

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
	if a.EOM != b.EOM {
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

func (m *SyncMsg) GetServerID() interfaces.IHash {
	return m.ServerID
}

func (m *SyncMsg) LogFields() log.Fields {
	return log.Fields{"category": "message", "messagetype": "FedVoteMsg", "dbheight": m.DBHeight, "newleader": m.ServerID.String()[4:12]}
}

func (m *SyncMsg) GetRepeatHash() interfaces.IHash {
	return m.GetMsgHash()
}

// We have to return the hash of the underlying message.
func (m *SyncMsg) GetHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *SyncMsg) GetTimestamp() interfaces.Timestamp {
	return m.TS
}

func (m *SyncMsg) GetMsgHash() interfaces.IHash {
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

	eom := messages.General.CreateMsg(constants.EOM_MSG)
	eom, ack := s.CreateEOM(true, eom, m.VMIndex)

	if eom == nil { // TODO: What does this mean? -- clay
		is.(*state.State).Holding[m.GetMsgHash().Fixed()] = m
		return // Maybe we are not yet prepared to create an EOM...
	}
	va := new(FedVoteVolunteerMsg)
	va.Missing = eom
	va.Ack = ack

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
	va.EOM = m.EOM

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
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()

	buf := primitives.NewBuffer(data)

	if t, e := buf.PopByte(); e != nil || t != constants.SYNC_MSG {
		return nil, errors.New("Not a Sync Message Audit type")
	}
	if m.TS, err = buf.PopTimestamp(); err != nil {
		return nil, err
	}
	if m.EOM, err = buf.PopBool(); err != nil {
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

func (m *SyncMsg) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *SyncMsg) MarshalBinary() (data []byte, err error) {

	var buf primitives.Buffer

	if e := buf.PushByte(constants.SYNC_MSG); e != nil {
		return nil, e
	}
	if e := buf.PushTimestamp(m.TS); e != nil {
		return nil, e
	}
	if e := buf.PushBool(m.EOM); e != nil {
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
