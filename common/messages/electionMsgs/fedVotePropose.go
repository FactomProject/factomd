// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package electionMsgs

import (
	"errors"
	"fmt"
	"os"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages/msgbase"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/elections"
	"github.com/FactomProject/factomd/state"

	llog "github.com/FactomProject/factomd/log"
	log "github.com/sirupsen/logrus"
)

var _ = fmt.Print

// FedVoteProposalMsg is not a majority, it is just proposing a volunteer
type FedVoteProposalMsg struct {
	FedVoteMsg

	// Signer of the message
	Signer interfaces.IHash

	// Volunteer fields
	Volunteer FedVoteVolunteerMsg

	Signature interfaces.IFullSignature

	messageHash interfaces.IHash
}

var _ interfaces.IMsg = (*FedVoteVolunteerMsg)(nil)
var _ interfaces.IElectionMsg = (*FedVoteVolunteerMsg)(nil)

func NewFedProposalMsg(signer interfaces.IHash, vol FedVoteVolunteerMsg) *FedVoteProposalMsg {
	p := new(FedVoteProposalMsg)
	p.SetFullBroadcast(true)
	p.Volunteer = vol
	p.Signer = signer
	p.FedVoteMsg.TS = primitives.NewTimestampNow()
	p.VMIndex = vol.VMIndex
	p.SigType = vol.SigType

	return p
}

func (m *FedVoteProposalMsg) ElectionProcess(is interfaces.IState, elect interfaces.IElections) {
	e := elect.(*elections.Elections)
	/******  Election Adapter Control   ******/
	/**	Controlling the inner election state**/

	// When we get a propose, we should first execute the volunteer msg. Then execute the
	// propose. This is because the embedded information may be new to us.
	m.Volunteer.ElectionProcess(is, elect)

	// Leaders will respond with a message,
	// followers will respond with nil
	resp := e.Adapter.Execute(m)
	if resp == nil {
		return
	}
	resp.SendOut(is, resp)
	/*_____ End Election Adapter Control  _____*/
}

var _ interfaces.IMsg = (*FedVoteVolunteerMsg)(nil)

func (a *FedVoteProposalMsg) IsSameAs(msg interfaces.IMsg) bool {
	if b, ok := msg.(*FedVoteProposalMsg); ok {
		if !a.FedVoteMsg.IsSameAs(&b.FedVoteMsg) {
			return false
		}
		if !a.Signer.IsSameAs(b.Signer) {
			return false
		}
		if !a.Volunteer.IsSameAs(&b.Volunteer) {
			return false
		}
		return true
	}
	return false
}

func (m *FedVoteProposalMsg) GetServerID() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "FedVoteProposalMsg.GetServerID") }()

	return m.Signer
}

func (m *FedVoteProposalMsg) LogFields() log.Fields {
	return log.Fields{"category": "message", "messagetype": "FedVoteVolunteerMsg", "dbheight": m.DBHeight, "newleader": m.Signer.String()[4:12]}
}

func (m *FedVoteProposalMsg) GetRepeatHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "FedVoteProposalMsg.GetRepeatHash") }()

	return m.GetMsgHash()
}

// We have to return the hash of the underlying message.

func (m *FedVoteProposalMsg) GetHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "FedVoteProposalMsg.GetHash") }()

	return m.GetMsgHash()
}

func (m *FedVoteProposalMsg) GetTimestamp() interfaces.Timestamp {
	return m.TS.Clone()
}

func (m *FedVoteProposalMsg) GetMsgHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "FedVoteProposalMsg.GetMsgHash") }()

	if m.MsgHash == nil {
		data, err := m.MarshalBinary()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

func (m *FedVoteProposalMsg) Type() byte {
	return constants.VOLUNTEERPROPOSAL
}

func (m *FedVoteProposalMsg) GetVolunteerMessage() interfaces.ISignableElectionMsg {
	return &m.Volunteer
}

func (m *FedVoteProposalMsg) ElectionValidate(ie interfaces.IElections) int {
	// Set the super and let the base validate
	m.FedVoteMsg.Super = m
	return m.FedVoteMsg.ElectionValidate(ie)
}

func (m *FedVoteProposalMsg) Validate(is interfaces.IState) int {
	// Set the super and let the base validate
	m.FedVoteMsg.Super = m
	return m.FedVoteMsg.Validate(is)
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *FedVoteProposalMsg) ComputeVMIndex(state interfaces.IState) {
}

// Execute the leader functions of the given message
// Leader, follower, do the same thing.
func (m *FedVoteProposalMsg) LeaderExecute(state interfaces.IState) {
	m.FollowerExecute(state)
}

func (m *FedVoteProposalMsg) FollowerExecute(is interfaces.IState) {
	s := is.(*state.State)
	if s.Elections.(*elections.Elections).Adapter == nil {
		//s.Holding[m.GetMsgHash().Fixed()] = m
		s.AddToHolding(m.GetMsgHash().Fixed(), m) // FedVoteProposalMsg.FollowerExecute

		return
	}
	is.ElectionsQueue().Enqueue(m)
}

// Acknowledgements do not go into the process list.
func (e *FedVoteProposalMsg) Process(dbheight uint32, state interfaces.IState) bool {
	panic("Election object should never have its Process() method called")
}

func (e *FedVoteProposalMsg) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *FedVoteProposalMsg) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (m *FedVoteProposalMsg) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
			llog.LogPrintf("recovery", "Error unmarshalling: %v", r)
		}
	}()

	buf := primitives.NewBuffer(data)
	if t, e := buf.PopByte(); e != nil || t != m.Type() {
		return nil, errors.New("Not a Volunteer Proposal type")
	}

	err = buf.PopBinaryMarshallable(&m.FedVoteMsg)
	if err != nil {
		return
	}

	m.Signer = new(primitives.Hash)
	err = buf.PopBinaryMarshallable(m.Signer)
	if err != nil {
		return
	}

	err = buf.PopBinaryMarshallable(&m.Volunteer)
	if err != nil {
		return
	}

	m.Signature = new(primitives.Signature)
	err = buf.PopBinaryMarshallable(m.Signature)
	if err != nil {
		return nil, err
	}

	newData = buf.DeepCopyBytes()
	return
}

func (m *FedVoteProposalMsg) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *FedVoteProposalMsg) Sign(key interfaces.Signer) error {
	signature, err := msgbase.SignSignable(m, key)
	if err != nil {
		return err
	}
	m.Signature = signature
	return nil
}

func (m *FedVoteProposalMsg) GetSignature() interfaces.IFullSignature {
	return m.Signature
}

func (m *FedVoteProposalMsg) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "FedVoteProposalMsg.MarshalBinary err:%v", *pe)
		}
	}(&err)
	var buf primitives.Buffer

	data, err := m.MarshalForSignature()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	err = buf.PushBinaryMarshallable(m.Signature)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

func (m *FedVoteProposalMsg) MarshalForSignature() (data []byte, err error) {
	var buf primitives.Buffer

	if err = buf.PushByte(m.Type()); err != nil {
		return nil, err
	}

	err = buf.PushBinaryMarshallable(&m.FedVoteMsg)
	if err != nil {
		return nil, err
	}

	err = buf.PushBinaryMarshallable(m.Signer)
	if err != nil {
		return nil, err
	}

	err = buf.PushBinaryMarshallable(&m.Volunteer)
	if err != nil {
		return nil, err
	}

	data = buf.DeepCopyBytes()
	return data, nil
}

func (m *FedVoteProposalMsg) String() string {
	return fmt.Sprintf("Fed Vote Proposal by %x, for %s", m.Signer.Bytes()[3:6], m.Volunteer.String())
}
