// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package electionMsgs

import (
	//"github.com/PaulSnow/factom2d/state"
	"bytes"
	"errors"
	"fmt"
	"os"

	"github.com/PaulSnow/factom2d/activations"

	"github.com/PaulSnow/factom2d/common/constants"
	"github.com/PaulSnow/factom2d/common/interfaces"
	"github.com/PaulSnow/factom2d/common/messages"
	"github.com/PaulSnow/factom2d/common/primitives"
	"github.com/PaulSnow/factom2d/state"
	log "github.com/sirupsen/logrus"

	"github.com/PaulSnow/factom2d/common/messages/msgbase"
	"github.com/PaulSnow/factom2d/elections"
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
	EOMFrom   interfaces.IHash

	// Information about the vote for comparing
	Level uint32
	Rank  uint32

	Signature interfaces.IFullSignature

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
		m.Signer.Bytes()[3:6],
		m.Volunteer.ServerName,
		m.Committed,
		m.Level,
		m.Rank)
}

func NewFedVoteLevelMessage(signer interfaces.IHash, vol FedVoteVolunteerMsg) *FedVoteLevelMsg {
	f := new(FedVoteLevelMsg)
	f.SetFullBroadcast(true)
	f.Volunteer = vol
	f.Signer = signer
	f.EOMFrom = new(primitives.Hash)
	f.VMIndex = vol.VMIndex
	f.SigType = vol.SigType

	return f
}

func (m *FedVoteLevelMsg) ElectionProcess(is interfaces.IState, elect interfaces.IElections) {
	e := elect.(*elections.Elections)

	elections.CheckAuthSetsMatch("FedVoteLevelMsg.ElectionProcess()", e, e.State.(*state.State))

	// TODO: determine if we need to check here too, or if checking before every election is fine
	if !e.IsSafeToReplaceFed(e.FedID) {
		if is.IsActive(activations.AUTHRORITY_SET_MAX_DELTA) {
			e.LogPrintf("election", "FedVoteLevelMsg.ElectionProcess(): cannot remove more than half of the block's starting feds")
			return
		} else {
			e.LogPrintf("election", "FedVoteLevelMsg.ElectionProcess() WARN: removing more than half of the block's starting feds")
		}
	}

	/******  Election Adapter Control   ******/
	/**	Controlling the inner election state**/
	m.processIfCommitted(is, elect) // This will end the election if it's over

	resp := e.Adapter.Execute(m)
	if resp == nil {
		return
	}

	resp.SendOut(is, resp)

	// We also need to check if we should change our state if the election resolved
	if vote, ok := resp.(*FedVoteLevelMsg); ok {
		vote.processIfCommitted(is, elect)
	}

	/*_____ End Election Adapter Control  _____*/
}

// processCommitted will process a message that has it's committed flag. It will only
// process 1 commit message for 1 election. If you give it another, it will just toss it
func (m *FedVoteLevelMsg) processIfCommitted(is interfaces.IState, elect interfaces.IElections) {
	if !m.Committed {
		return
	}
	e := elect.(*elections.Elections)

	elections.CheckAuthSetsMatch("processIfCommitted()", e, e.State.(*state.State))

	// This block of code is only called ONCE per election
	if !e.Adapter.IsElectionProcessed() {
		// Used for printouts before we swap them!
		idx := e.LeaderIndex(m.Volunteer.FedID)
		fmt.Printf("**** FedVoteLevelMsg %12s Swaping Fed: %d(%x) Audit: %d(%x)\n",
			is.GetFactomNodeName(),
			m.Volunteer.FedIdx, m.Volunteer.FedID.Bytes()[3:6],
			m.Volunteer.ServerIdx, m.Volunteer.ServerID.Bytes()[3:6])

		e.LogPrintf("election", "**** FedVoteLevelMsg %12s Swapping Fed: %d(%x) Audit: %d(%x)",

			is.GetFactomNodeName(),
			m.Volunteer.FedIdx, m.Volunteer.FedID.Bytes()[3:6],
			m.Volunteer.ServerIdx, m.Volunteer.ServerID.Bytes()[3:6])

		e.LogPrintf("election", "LeaderSwapState %d/%d/%d", m.DBHeight, m.VMIndex, m.Minute)
		e.LogPrintf("election", "Demote  %x[%d]", e.Federated[m.Volunteer.FedIdx].GetChainID().Bytes()[3:6], m.Volunteer.FedIdx)
		e.LogPrintf("election", "Promote %x[%d]", e.Audit[m.Volunteer.ServerIdx].GetChainID().Bytes()[3:6], m.Volunteer.ServerIdx)

		e.Federated[m.Volunteer.FedIdx], e.Audit[m.Volunteer.ServerIdx] =
			e.Audit[m.Volunteer.ServerIdx], e.Federated[m.Volunteer.FedIdx]
		e.Adapter.SetElectionProcessed(true)
		m.ProcessInState = true
		m.SetValid()

		// Ensure we don't start another election for this server
		se := new(EomSigInternal)
		se.SigType = m.Volunteer.SigType
		se.NName = m.Volunteer.Name
		se.DBHeight = m.Volunteer.DBHeight
		se.Minute = m.Volunteer.Minute
		se.VMIndex = m.Volunteer.VMIndex
		se.ServerID = m.Volunteer.ServerID

		e.Msgs = append(e.Msgs, se)

		// Send for the state to do the swap. It will only be sent with this
		// flag ONCE
		is.LogMessage("InMsgQueue", "enqueue_FedVoteLevelMsg", m)
		is.InMsgQueue().Enqueue(m)
		// End the election by setting this to '-1'
		e.Electing = -1
		e.LogPrintf("election", "**** Election is over. Elected %d[%x] ****", m.Volunteer.ServerIdx, m.Volunteer.ServerID.Bytes()[3:6])

		e.LogPrintf("faulting", "**** Election is over. Elected %d[%x] ****", m.Volunteer.ServerIdx, m.Volunteer.ServerID.Bytes()[3:6])
		e.LogPrintf("faulting", e.Adapter.MessageLists())
		e.LogPrintf("faulting", e.Adapter.Status())

		// Add some string feedback for prints
		t := "EOM"
		if !m.SigType {
			t = "DBSig"
		}
		s := is.(*state.State)
		//								   T   N    mH  mM  mV  eH  eM  eV
		s.Election1 = fmt.Sprintf("%6s %10s %5d %5d %5d %5d %5d %5d  ",
			t,
			s.FactomNodeName,
			m.DBHeight,
			m.Minute,
			m.VMIndex,
			e.DBHeight,
			e.Minute,
			e.VMIndex)

		if idx != -1 {
			s.Election2 = e.FeedBackStr("m", true, idx)
		}
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
	e := s.Elections.(*elections.Elections)
	if pl == nil || e.Adapter == nil {
		//s.Holding[m.GetMsgHash().Fixed()] = m
		s.AddToHolding(m.GetMsgHash().Fixed(), m) // FedVoteLevelMsg.FollowerExecute

		return
	}

	// Committed should only be processed once.
	//		ProcessInState is not marshalled, so only we can pass this to ourselves
	//		allowing the election adapter to ensure only once behavior
	if m.ProcessInState {
		elections.CheckAuthSetsMatch("FedVoteLevelMsg.FollowerExecute", e, s)

		fmt.Println("LeaderSwapState", s.GetFactomNodeName(), m.DBHeight, m.Minute)

		s.LogPrintf("election", "**** FedVoteLevelMsg %12s Swapping Fed: %d(%x) Audit: %d(%x)\n",
			s.GetFactomNodeName(),
			m.Volunteer.FedIdx, m.Volunteer.FedID.Bytes()[3:6],
			m.Volunteer.ServerIdx, m.Volunteer.ServerID.Bytes()[3:6])

		s.LogPrintf("executeMsg", "Pre  Election s.Leader=%v s.LeaderVMIndex to %v", s.Leader, s.LeaderVMIndex)
		s.LogPrintf("executeMsg", "LeaderSwapState %d/%d/%d", m.DBHeight, m.VMIndex, m.Minute)
		s.LogPrintf("executeMsg", "Demote  %x[%d]", pl.FedServers[m.Volunteer.FedIdx].GetChainID().Bytes()[3:6], m.Volunteer.FedIdx)
		s.LogPrintf("executeMsg", "Promote %x[%d]", pl.AuditServers[m.Volunteer.ServerIdx].GetChainID().Bytes()[3:6], m.Volunteer.ServerIdx)

		pl.FedServers[m.Volunteer.FedIdx], pl.AuditServers[m.Volunteer.ServerIdx] =
			pl.AuditServers[m.Volunteer.ServerIdx], pl.FedServers[m.Volunteer.FedIdx]

		s.LogPrintf("executeMsg", "Pre  Election s.Leader=%v s.LeaderVMIndex to %v", s.Leader, s.LeaderVMIndex)

		// reset my leader variables, cause maybe we changed...
		Leader, LeaderVMIndex := s.LeaderPL.GetVirtualServers(int(s.CurrentMinute), s.IdentityChainID)
		{ // debug
			if s.Leader != Leader {
				s.LogPrintf("executeMsg", "FedVoteLevelMsg.FollowerExecute() changed s.Leader to %v", Leader)
				s.Leader = Leader
			}
			if s.LeaderVMIndex != LeaderVMIndex {
				s.LogPrintf("executeMsg", "FedVoteLevelMsg.FollowerExecute() changed s.LeaderVMIndex to %v", LeaderVMIndex)
				s.LeaderVMIndex = LeaderVMIndex
			}
		}
		s.LogPrintf("executeMsg", "Post Election s.Leader=%v s.LeaderVMIndex to %v", s.Leader, s.LeaderVMIndex)

		// Trim the processlist for all messages above this height. They are signed by the old leader, and have
		// not yet been processed.
		pl.TrimVMList(m.Volunteer.Ack.(*messages.Ack).Height, m.VMIndex)

		// Add to the process list (which will get immediately processed)
		is.LogMessage("executeMsg", "add to pl", m.Volunteer.Ack)
		pl.AddToProcessList(pl.State, m.Volunteer.Ack.(*messages.Ack), m.Volunteer.Missing)
	} else {
		is.ElectionsQueue().Enqueue(m)
	}
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

	if !a.EOMFrom.IsSameAs(b.EOMFrom) {
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

	if !a.Signature.IsSameAs(b.Signature) {
		return false
	}

	return true
}

func (m *FedVoteLevelMsg) GetServerID() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "FedVoteLevelMsg.GetServerID") }()

	return m.Signer
}

func (m *FedVoteLevelMsg) LogFields() log.Fields {
	return log.Fields{"category": "message", "messagetype": "FedVoteVolunteerMsg", "dbheight": m.DBHeight, "newleader": m.Signer.String()[4:12]}
}

func (m *FedVoteLevelMsg) GetRepeatHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "FedVoteLevelMsg.GetRepeatHash") }()

	return m.GetMsgHash()
}

// We have to return the hash of the underlying message.

func (m *FedVoteLevelMsg) GetHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "FedVoteLevelMsg.GetHash") }()

	return m.GetMsgHash()
}

func (m *FedVoteLevelMsg) GetTimestamp() interfaces.Timestamp {
	return m.TS.Clone()
}

func (m *FedVoteLevelMsg) GetMsgHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "FedVoteLevelMsg.GetMsgHash") }()

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

func (m *FedVoteLevelMsg) GetVolunteerMessage() interfaces.ISignableElectionMsg {
	return &m.Volunteer
}

func (m *FedVoteLevelMsg) ElectionValidate(ie interfaces.IElections) int {
	// Set the super and let the base validate
	m.FedVoteMsg.Super = m
	return m.FedVoteMsg.ElectionValidate(ie)
}

func (m *FedVoteLevelMsg) Validate(is interfaces.IState) int {
	if m.IsValid() {
		return 1
	}

	// Set the super and let the base validate
	m.FedVoteMsg.Super = m
	return m.FedVoteMsg.Validate(is)
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

	if t, e := buf.PopByte(); e != nil || t != m.Type() {
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

	m.EOMFrom, err = buf.PopIHash()
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

	m.Signature = new(primitives.Signature)
	err = buf.PopBinaryMarshallable(m.Signature)
	if err != nil {
		return nil, err
	}

	data = buf.Bytes()
	return data, nil
}

func (m *FedVoteLevelMsg) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *FedVoteLevelMsg) Sign(key interfaces.Signer) error {
	signature, err := msgbase.SignSignable(m, key)
	if err != nil {
		return err
	}
	m.Signature = signature
	return nil
}

func (m *FedVoteLevelMsg) GetSignature() interfaces.IFullSignature {
	return m.Signature
}

func (m *FedVoteLevelMsg) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "FedVoteLevelMsg.MarshalBinary err:%v", *pe)
		}
	}(&err)
	var buf primitives.Buffer

	data, err := m.MarshalForSignature()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	err = buf.PushBinaryMarshallableMsgArray(m.Justification)
	if err != nil {
		return nil, err
	}

	if m.Signature != nil {
		err = buf.PushBinaryMarshallable(m.Signature)
		if err != nil {
			return nil, err
		}
	}

	return buf.DeepCopyBytes(), nil
}

func (m *FedVoteLevelMsg) MarshalForSignature() (data []byte, err error) {
	var buf primitives.Buffer

	if err = buf.PushByte(m.Type()); err != nil {
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

	err = buf.PushIHash(m.EOMFrom)
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

	data = buf.DeepCopyBytes()
	return data, nil
}
