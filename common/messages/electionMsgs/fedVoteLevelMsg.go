// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package electionMsgs

import (
	"errors"
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/state"
	log "github.com/sirupsen/logrus"
	//"github.com/FactomProject/factomd/state"
	"bytes"

	"github.com/FactomProject/factomd/elections"
)

var _ = fmt.Print

// FedVoteLevelMsg
// We can vote on a majority of votes seen and issuing a vote at a particular
// level. Only 1 vote per level can be issued per leader.
type FedVoteLevelMsg struct {
	FedVoteMsg

	// Signer of the message
	Signer interfaces.IHash

	// Volunteer fields
	Volunteer FedVoteVolunteerMsg

	Committed bool
	// Information about the vote for comparing
	Level uint32
	Rank  uint32

	// Need a majority of these to justify our vote
	Justification []interfaces.IMsg

	// Tells the state to process it
	ProcessInState bool

	messageHash interfaces.IHash
}

var _ interfaces.IMsg = (*FedVoteVolunteerMsg)(nil)
var _ interfaces.IElectionMsg = (*FedVoteVolunteerMsg)(nil)

func (m *FedVoteLevelMsg) String() string {
	return fmt.Sprintf("%s DBHeight %d Minute %d Signer %x Volunteer %s Committed %v Level %d Rank %d",
		"Fed VoteLevelMsg",
		m.DBHeight,
		m.Minute,
		m.Signer.Bytes()[3:9],
		m.Volunteer.ServerName,
		m.Committed,
		m.Level,
		m.Rank)
}

func NewFedVoteLevelMessage(signer interfaces.IHash, vol FedVoteVolunteerMsg) *FedVoteLevelMsg {
	f := new(FedVoteLevelMsg)
	f.Volunteer = vol
	f.Signer = signer

	return f
}

func (m *FedVoteLevelMsg) ElectionProcess(is interfaces.IState, elect interfaces.IElections) {
	e := elect.(*elections.Elections)
	/******  Election Adapter Control   ******/
	/**	Controlling the inner election state**/
	m.processIfCommitted(is, elect)
	m.InitiateElectionAdapter(is)

	resp := e.Adapter.Execute(m)
	if resp == nil {
		return
	}

	resp.SendOut(is, resp)
	// We also need to check if we should change our state if the eletion resolved
	if vote, ok := resp.(*FedVoteLevelMsg); ok {
		vote.processIfCommitted(is, elect)
	}

	/*_____ End Election Adapter Control  _____*/
}

// processCommitted will process a message that has it's committed flag
func (m *FedVoteLevelMsg) processIfCommitted(is interfaces.IState, elect interfaces.IElections) {
	if !m.Committed {
		return
	}
	e := elect.(*elections.Elections)

	// do not do it twice
	if !e.Adapter.IsElectionProcessed() {
		e.Federated[m.Volunteer.FedIdx], e.Audit[m.Volunteer.ServerIdx] =
			e.Audit[m.Volunteer.ServerIdx], e.Federated[m.Volunteer.FedIdx]
		e.Adapter.SetElectionProcessed(true)
		m.ProcessInState = true
		m.SetValid()
		// Send for the state to do the swap
		is.InMsgQueue().Enqueue(m)
	}
}

// Execute the leader functions of the given message
// Leader, follower, do the same thing.
func (m *FedVoteLevelMsg) LeaderExecute(state interfaces.IState) {
	m.FollowerExecute(state)
}

func (m *FedVoteLevelMsg) FollowerExecute(is interfaces.IState) {
	s := is.(*state.State)
	pl := s.ProcessLists.Get(m.DBHeight)
	if pl == nil {
		s.Holding[m.GetMsgHash().Fixed()] = m
	}

	// Committed should only be processed once.
	//		ProcessInState is not marshalled, so only we can pass this to ourselves
	//		allowing the election adapter to ensure only once behavior
	if m.ProcessInState {
		//fmt.Println("LeaderSwap", s.GetFactomNodeName(), m.DBHeight, m.Minute)
		pl.FedServers[m.Volunteer.FedIdx], pl.AuditServers[m.Volunteer.ServerIdx] =
			pl.AuditServers[m.Volunteer.ServerIdx], pl.FedServers[m.Volunteer.FedIdx]

		pl.AddToProcessList(m.Volunteer.Ack.(*messages.Ack), m.Volunteer.Missing)
	} else {
		is.ElectionsQueue().Enqueue(m)
	}

	// reset my leader variables, cause maybe we changed...
	s.Leader, s.LeaderVMIndex = s.LeaderPL.GetVirtualServers(int(m.Minute), s.IdentityChainID)
}

var _ interfaces.IMsg = (*FedVoteVolunteerMsg)(nil)

func (a *FedVoteLevelMsg) IsSameAs(msg interfaces.IMsg) bool {
	b, ok := msg.(*FedVoteLevelMsg)
	if !ok {
		return false
	}

	if !a.FedVoteMsg.IsSameAs(&b.FedVoteMsg) {
		return false
	}

	if !a.Signer.IsSameAs(b.Signer) {
		return false
	}

	if a.Committed != b.Committed {
		return false
	}

	if a.Level != b.Level {
		return false
	}

	if a.Rank != b.Rank {
		return false
	}

	if !a.Volunteer.IsSameAs(&b.Volunteer) {
		return false
	}

	if len(a.Justification) != len(b.Justification) {
		return false
	}

	for i := range a.Justification {
		data, err := a.Justification[i].MarshalBinary()
		if err != nil {
			return false
		}

		datab, err := b.Justification[i].MarshalBinary()
		if err != nil {
			return false
		}

		if bytes.Compare(data, datab) != 0 {
			return false
		}
	}

	return true
}

func (m *FedVoteLevelMsg) GetServerID() interfaces.IHash {
	return m.Signer
}

func (m *FedVoteLevelMsg) LogFields() log.Fields {
	return log.Fields{"category": "message", "messagetype": "FedVoteVolunteerMsg", "dbheight": m.DBHeight, "newleader": m.Signer.String()[4:12]}
}

func (m *FedVoteLevelMsg) GetRepeatHash() interfaces.IHash {
	return m.GetMsgHash()
}

// We have to return the haswh of the underlying message.

func (m *FedVoteLevelMsg) GetHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *FedVoteLevelMsg) GetTimestamp() interfaces.Timestamp {
	return m.TS
}

func (m *FedVoteLevelMsg) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
		data, err := m.MarshalBinary()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

func (m *FedVoteLevelMsg) Type() byte {
	return constants.VOLUNTEERLEVELVOTE
}

func (m *FedVoteLevelMsg) Validate(state interfaces.IState) int {
	baseMsg := m.FedVoteMsg.Validate(state)
	if baseMsg == -1 {
		return -1
	}
	return 1
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *FedVoteLevelMsg) ComputeVMIndex(state interfaces.IState) {
}

// Acknowledgements do not go into the process list.
func (e *FedVoteLevelMsg) Process(dbheight uint32, state interfaces.IState) bool {
	panic("Ack object should never have its Process() method called")
}

func (e *FedVoteLevelMsg) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *FedVoteLevelMsg) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (m *FedVoteLevelMsg) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	var buf primitives.Buffer
	buf.Write(data)

	if t, e := buf.PopByte(); e != nil || t != constants.VOLUNTEERLEVELVOTE {
		return nil, errors.New("Not a Volunteer Level type")
	}

	err = buf.PopBinaryMarshallable(&m.FedVoteMsg)
	if err != nil {
		return
	}

	m.Signer, err = buf.PopIHash()
	if err != nil {
		return
	}

	err = buf.PopBinaryMarshallable(&m.Volunteer)
	if err != nil {
		return
	}

	m.Committed, err = buf.PopBool()
	if err != nil {
		return
	}

	m.Level, err = buf.PopUInt32()
	if err != nil {
		return
	}

	m.Rank, err = buf.PopUInt32()
	if err != nil {
		return
	}

	m.Justification, err = buf.PopBinaryMarshallableMsgArray()
	if err != nil {
		return
	}

	data = buf.Bytes()
	return data, nil
}

func (m *FedVoteLevelMsg) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *FedVoteLevelMsg) MarshalBinary() (data []byte, err error) {
	var buf primitives.Buffer

	if err = buf.PushByte(constants.VOLUNTEERLEVELVOTE); err != nil {
		return nil, err
	}

	err = buf.PushBinaryMarshallable(&m.FedVoteMsg)
	if err != nil {
		return
	}

	err = buf.PushIHash(m.Signer)
	if err != nil {
		return
	}

	err = buf.PushBinaryMarshallable(&m.Volunteer)
	if err != nil {
		return
	}

	err = buf.PushBool(m.Committed)
	if err != nil {
		return
	}

	err = buf.PushUInt32(m.Level)
	if err != nil {
		return
	}

	err = buf.PushUInt32(m.Rank)
	if err != nil {
		return
	}

	err = buf.PushBinaryMarshallableMsgArray(m.Justification)
	if err != nil {
		return
	}

	data = buf.DeepCopyBytes()
	return data, nil
}
