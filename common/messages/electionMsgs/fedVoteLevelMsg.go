// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package electionMsgs

import (
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	log "github.com/sirupsen/logrus"
	//"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/elections"
	"github.com/FactomProject/factomd/state"
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

	messageHash interfaces.IHash
}

var _ interfaces.IMsg = (*FedVoteVolunteerMsg)(nil)
var _ interfaces.IElectionMsg = (*FedVoteVolunteerMsg)(nil)

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
	m.InitiateElectionAdapter(is)

	resp := e.Adapter.Execute(m)
	if resp == nil {
		return
	}

	resp.SendOut(is, resp)
	// We also need to check if we should change our state if the eletion resolved
	if vote, ok := resp.(*FedVoteLevelMsg); ok {
		if vote.Committed {
			vote.SetLocal(true)
			is.InMsgQueue().Enqueue(vote)
		}
	}

	/*_____ End Election Adapter Control  _____*/
}

var _ interfaces.IMsg = (*FedVoteVolunteerMsg)(nil)

func (a *FedVoteLevelMsg) IsSameAs(msg interfaces.IMsg) bool {

	return false
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

// Execute the leader functions of the given message
// Leader, follower, do the same thing.
func (m *FedVoteLevelMsg) LeaderExecute(state interfaces.IState) {
	m.FollowerExecute(state)
}

func (m *FedVoteLevelMsg) FollowerExecute(is interfaces.IState) {
	if !m.Committed {
		is.ElectionsQueue().Enqueue(m)
		return
	}

	s := is.(*state.State)
	pl := s.ProcessLists.Get(m.DBHeight)
	pl.FedServers[m.Volunteer.FedIdx], pl.AuditServers[m.Volunteer.ServerIdx] =
		pl.AuditServers[m.Volunteer.ServerIdx], pl.FedServers[m.Volunteer.FedIdx]

	pl.AddToProcessList(m.Volunteer.Ack.(*messages.Ack), m.Volunteer.Missing)
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

	return
}

func (m *FedVoteLevelMsg) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *FedVoteLevelMsg) MarshalBinary() (data []byte, err error) {
	var buf primitives.Buffer

	data = buf.DeepCopyBytes()
	return data, nil
}

func (m *FedVoteLevelMsg) String() string {
	return ""
}
