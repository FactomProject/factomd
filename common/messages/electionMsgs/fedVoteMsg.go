// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package electionMsgs

import (
	"bytes"
	"fmt"
	"os"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/messages/msgbase"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/elections"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/goleveldb/leveldb/errors"
	log "github.com/sirupsen/logrus"
)

var _ = fmt.Print

// FedVoteMsg
// We vote on the Audit Server to replace a Federated Server that fails
// We vote to move to the next round, if the audit server fails.
// Could make these two messages, but for now we will do it in one.
type FedVoteMsg struct {
	msgbase.MessageBase
	TS       interfaces.Timestamp // Message Timestamp
	TypeMsg  byte                 // Can be either a Volunteer from an Audit Server, or End of round
	DBHeight uint32               // Directory Block Height that owns this ack
	SigType  bool                 // False for dbsig, true for EOM

	// NOT MARSHALED
	Super interfaces.ISignableElectionMsg `json:"-"`
}

func (m *FedVoteMsg) InitFields(elect interfaces.IElections) {
	election := elect.(*elections.Elections)
	m.TS = primitives.NewTimestampNow()
	m.DBHeight = uint32(election.DBHeight)
	m.Minute = byte(election.Minute)
	// You need to init the type
}

func (m *FedVoteMsg) ElectionProcess(is interfaces.IState, elect interfaces.IElections) {
}

var _ interfaces.IMsg = (*FedVoteMsg)(nil)
var _ interfaces.IElectionMsg = (*FedVoteMsg)(nil)

func (a *FedVoteMsg) IsSameAs(msg interfaces.IMsg) bool {
	b, ok := msg.(*FedVoteMsg)
	if !ok {
		return false
	}
	if a.TS.GetTimeMilli() != b.TS.GetTimeMilli() {
		return false
	}
	if a.DBHeight != b.DBHeight {
		return false
	}
	if a.VMIndex != b.VMIndex {
		return false
	}
	if a.Minute != b.Minute {
		return false
	}
	binA, errA := a.MarshalBinary()
	binB, errB := a.MarshalBinary()
	if errA != nil || errB != nil || bytes.Compare(binA, binB) != 0 {
		return false
	}
	return true
}

//func (m *FedVoteMsg) GetServerID() interfaces.IHash {
//	return nil
//}

func (m *FedVoteMsg) ComparisonMinute() int {
	if !m.SigType {
		return -1
	}
	return int(m.GetMinute())
}

func (m *FedVoteMsg) GetTimestamp() interfaces.Timestamp {
	return m.TS
}

func (m *FedVoteMsg) LogFields() log.Fields {
	return log.Fields{"category": "message", "messagetype": "FedVoteMsg", "dbheight": m.DBHeight}
}

func (m *FedVoteMsg) GetRepeatHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *FedVoteMsg) GetHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *FedVoteMsg) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
		data, err := m.MarshalBinary()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

func (m *FedVoteMsg) Type() byte {
	return constants.FEDVOTE_MSG_BASE
}

func (m *FedVoteMsg) ElectionValidate(ie interfaces.IElections) int {
	e := ie.(*elections.Elections)

	// Current height and minute TODO: Also check VMIndex?
	if int(m.DBHeight) == e.DBHeight && e.ComparisonMinute() == m.ComparisonMinute() {
		sm := m.Super
		vol := sm.GetVolunteerMessage().(*FedVoteVolunteerMsg)

		// Protect from index out of bounds
		if int(vol.ServerIdx) >= len(e.Audit) || int(vol.FedIdx) >= len(e.Federated) {
			return -1
		}

		if !vol.ServerID.IsSameAs(e.Audit[vol.ServerIdx].GetChainID()) ||
			!vol.FedID.IsSameAs(e.Federated[vol.FedIdx].GetChainID()) {
			return -1
		}

		// For a different election on this minute
		if int(m.VMIndex) != e.VMIndex {
			return 0
		}

		return 1 // This is our election!
	}

	// Ignore all elections messages from the past
	if int(m.DBHeight) < e.DBHeight || (int(m.DBHeight) == e.DBHeight && m.ComparisonMinute() < e.ComparisonMinute()) {
		e.LogMessage("election", "Message is invalid (past)", m)
		return -1
	}

	// Is from the future, probably from Marty McFly
	if int(m.DBHeight) > e.DBHeight || (int(m.DBHeight) == e.DBHeight && m.ComparisonMinute() > e.ComparisonMinute()) {
		return 0
	}

	panic(errors.New("Thought I covered all the cases"))
}

// ValidateVolunteer validates if the volunteer has their process list at the correct height
// If the volunteer is too low, it is invalid. If it is too high, then we return 0.
func (m *FedVoteMsg) ValidateVolunteer(v FedVoteVolunteerMsg, is interfaces.IState) int {
	s := is.(*state.State)

	pl := s.ProcessLists.Get(v.DBHeight)
	if pl == nil {
		return 0
	}

	if v.VMIndex >= len(pl.VMs) || v.VMIndex < 0 {
		return -1
	}

	vm := pl.VMs[v.VMIndex]

	ack := v.Ack.(*messages.Ack)
	if vm.Height < int(ack.Height) {
		return 0
	} else if vm.Height > int(ack.Height) {
		return -1
	}

	return 1
}

func (m *FedVoteMsg) Validate(is interfaces.IState) int {
	s := is.(*state.State)
	if m.DBHeight < s.GetLeaderHeight() {
		return -1
	}
	sm := m.Super
	vol := sm.GetVolunteerMessage().(*FedVoteVolunteerMsg)

	// Check to make sure the volunteer message can be put in our process list
	if validVolunteer := m.ValidateVolunteer(*vol, is); validVolunteer != 1 {
		if validVolunteer == -1 {
			return -1
		}

		// Volunteer is not valid because the volunteer has a higher process list height
		return 0
	}

	signed, err := sm.MarshalForSignature()
	if err != nil {
		return -1
	}

	valid, err := is.VerifyAuthoritySignature(signed, sm.GetSignature().GetSignature(), m.DBHeight)
	if err != nil || valid < 0 {
		return -1
	}

	return 1
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *FedVoteMsg) ComputeVMIndex(state interfaces.IState) {
}

// Execute the leader functions of the given message
// Leader, follower, do the same thing.
func (m *FedVoteMsg) LeaderExecute(state interfaces.IState) {
	m.FollowerExecute(state)
}

func (m *FedVoteMsg) FollowerExecute(state interfaces.IState) {
	state.ElectionsQueue().Enqueue(m)
}

// Acknowledgements do not go into the process list.
func (e *FedVoteMsg) Process(dbheight uint32, state interfaces.IState) bool {
	panic("Ack object should never have its Process() method called")
}

func (e *FedVoteMsg) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *FedVoteMsg) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (m *FedVoteMsg) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			os.Stderr.WriteString("Error UnmashalBinaryData FedVoteMsg")
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()

	buf := primitives.NewBuffer(data)
	if t, e := buf.PopByte(); e != nil || t != m.Type() {
		return nil, errors.New("Not a Fed Vote Base")
	}
	if m.TS, err = buf.PopTimestamp(); err != nil {
		return nil, err
	}
	if m.DBHeight, err = buf.PopUInt32(); err != nil {
		return nil, err
	}
	if m.VMIndex, err = buf.PopInt(); err != nil {
		return nil, err
	}
	if m.Minute, err = buf.PopByte(); err != nil {
		return nil, err
	}
	if m.SigType, err = buf.PopBool(); err != nil {
		return nil, err
	}
	newData = buf.DeepCopyBytes()
	return
}

func (m *FedVoteMsg) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *FedVoteMsg) MarshalBinary() (data []byte, err error) {
	var buf primitives.Buffer

	if err = buf.PushByte(m.Type()); err != nil {
		return nil, err
	}
	if e := buf.PushTimestamp(m.TS); e != nil {
		return nil, e
	}
	if e := buf.PushUInt32(m.DBHeight); e != nil {
		return nil, e
	}
	if e := buf.PushInt(m.VMIndex); e != nil {
		return nil, e
	}
	if e := buf.PushByte(m.Minute); e != nil {
		return nil, e
	}
	if e := buf.PushBool(m.SigType); e != nil {
		return nil, e
	}
	data = buf.DeepCopyBytes()
	return data, nil
}

func (m *FedVoteMsg) String() string {
	if m.LeaderChainID == nil {
		m.LeaderChainID = primitives.NewZeroHash()
	}
	return fmt.Sprintf("%s DBHeight %d Minute %d", "FedVoteMsg ", m.DBHeight, m.Minute)
}
