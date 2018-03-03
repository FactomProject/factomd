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
	Minute   byte                 // Minute (-1 for dbsig)
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
	return constants.INVALID_MSG
}

// InitiateElectionAdapter will create a new election adapter if needed for the election message
func (m *FedVoteMsg) InitiateElectionAdapter(st interfaces.IState) {
	s := st.(*state.State)
	e := s.Elections.(*elections.Elections)

	if e.Adapter == nil || e.Adapter.GetDBHeight() < int(m.DBHeight) || e.Adapter.GetMinute() < int(m.Minute) {
		// TODO: Is cancelling an old election ALWAYS the best way? Should we have some cleanup? Maybe validate
		// TODO: the new election is valid and the old one has concluded
		e.Adapter = NewElectionAdapter(e)
		e.Adapter.SetObserver(!s.IsLeader())
	}
}

func (m *FedVoteMsg) Validate(st interfaces.IState) int {
	s := st.(*state.State)
	e := s.Elections.(*elections.Elections)

	// Ignore all elections messages from the past
	if int(m.DBHeight) < e.DBHeight || int(m.Minute) < e.Minute {
		return -1
	}

	// Is from the future, probably from Marty McFly
	if int(m.DBHeight) > e.DBHeight || int(m.Minute) < e.Minute {
		return 0
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
	if t, e := buf.PopByte(); e != nil || t != constants.VOLUNTEERAUDIT {
		return nil, errors.New("Not a Volunteer Audit type")
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
	newData = buf.DeepCopyBytes()
	return
}

func (m *FedVoteMsg) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *FedVoteMsg) MarshalBinary() (data []byte, err error) {
	var buf primitives.Buffer

	if err = buf.PushByte(constants.VOLUNTEERAUDIT); err != nil {
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
	data = buf.DeepCopyBytes()
	return data, nil
}

func (m *FedVoteMsg) String() string {
	if m.LeaderChainID == nil {
		m.LeaderChainID = primitives.NewZeroHash()
	}
	return fmt.Sprintf("%s DBHeight %d Minute %d", "FedVoteMsg ", m.DBHeight, m.Minute)
}
