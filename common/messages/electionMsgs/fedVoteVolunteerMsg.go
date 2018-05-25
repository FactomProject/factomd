// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package electionMsgs

import (
	"bytes"
	"fmt"
	"os"
	//"github.com/FactomProject/factomd/state"

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

// FedVoteVolunteerMsg
// We vote on the Audit Server to replace a Federated Server that fails
// We vote to move to the next round, if the audit server fails.
// Could make these two messages, but for now we will do it in one.
type FedVoteVolunteerMsg struct {
	FedVoteMsg
	// Volunteer fields
	Name       string           // Server name
	FedIdx     uint32           // Server faulting
	FedID      interfaces.IHash // Server faulting
	ServerIdx  uint32           // Index of Server replacing
	ServerID   interfaces.IHash // Volunteer Server ChainID
	ServerName string           // Volunteer Name
	Weight     interfaces.IHash
	Missing    interfaces.IMsg // The Missing DBSig or EOM
	Ack        interfaces.IMsg // The acknowledgement for the missing message
	Round      int

	messageHash interfaces.IHash

	Signature interfaces.IFullSignature

	// Non-marshalled fields
	// The authority set to be used in this election
	AuditServers []interfaces.IServer // List of Audit Servers
	FedServers   []interfaces.IServer // List of Federated Servers
}

var _ interfaces.IMsg = (*FedVoteVolunteerMsg)(nil)
var _ interfaces.IElectionMsg = (*FedVoteVolunteerMsg)(nil)

func (m *FedVoteVolunteerMsg) ElectionProcess(is interfaces.IState, elect interfaces.IElections) {
	e := elect.(*elections.Elections)
	// This message picked up the authority set for the affected processlist in state before arriving here

	//e.Audit = m.AuditServers
	//e.Federated = m.FedServers

	idx := e.LeaderIndex(is.GetIdentityChainID())
	aidx := e.AuditIndex(is.GetIdentityChainID())

	// Really this seems like a bad plan. We scramble the audit server order for each volunteer instead of once at the
	// start of an election.
	// TODO -- revisit this It can nominate the same audit server in different rounds.

	auditIdx := e.AuditPriority()
	if aidx >= 0 && auditIdx == aidx { // If I am an audit server and I am the volunteer ...
		e.FeedBackStr(fmt.Sprintf("V%d", m.ServerIdx), false, aidx)
	} else if idx >= 0 { // else if I am a leader ...
		e.FeedBackStr(fmt.Sprintf("V%d", m.ServerIdx), true, idx)
	} else if aidx >= 0 { // else if I am an audit server but not the volunteer
		e.FeedBackStr(fmt.Sprintf("*%d", m.ServerIdx), false, aidx)
	}
	e.Msg = m.Missing
	e.Ack = m.Ack
	e.VName = m.ServerName

	/******  Election Adapter Control   ******/
	/**	Controlling the inner election state**/
	// Response from non-fed is nil.
	resp := e.Adapter.Execute(m)
	if resp == nil {
		return
	}

	resp.SendOut(is, resp)
	/*_____ End Election Adapter Control  _____*/

}

// Execute the leader functions of the given message
// Leader, follower, do the same thing.
func (m *FedVoteVolunteerMsg) LeaderExecute(state interfaces.IState) {
	m.FollowerExecute(state)
}

func (m *FedVoteVolunteerMsg) FollowerExecute(is interfaces.IState) {
	s := is.(*state.State)
	e := s.Elections.(*elections.Elections)
	if e.Adapter == nil {
		s.Holding[m.GetMsgHash().Fixed()] = m
		return
	}

	elections.CheckAuthSetsMatch("FedVoteVolunteerMsg.FollowerExecute", e, s)

	// Add the authority set this election involves from the process list
	// may this should live in the election adapter? It's life mirrors the election ...
	pl := s.ProcessLists.Get(uint32(m.DBHeight))
	if pl == nil {
		return
	}
	//s_fservers := pl.FedServers
	//s_aservers := pl.AuditServers

	//m.FedServers = nil
	//m.AuditServers = nil
	//
	//for _, s := range s_fservers {
	//	m.FedServers = append(m.FedServers, s) // Append the federated servers
	//}
	//for _, s := range s_aservers {
	//	m.AuditServers = append(m.AuditServers, s) // Append the audit servers
	//}
	// these will be sorted to priority order in the election.
	is.ElectionsQueue().Enqueue(m)
}

var _ interfaces.IMsg = (*FedVoteVolunteerMsg)(nil)

func (a *FedVoteVolunteerMsg) IsSameAs(msg interfaces.IMsg) bool {
	b, ok := msg.(*FedVoteVolunteerMsg)
	if !ok {
		return false
	}
	if !a.FedVoteMsg.IsSameAs(&b.FedVoteMsg) {
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
	if a.FedIdx != b.FedIdx {
		return false
	}
	if a.FedID.Fixed() != b.FedID.Fixed() {
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

func (m *FedVoteVolunteerMsg) GetServerID() interfaces.IHash {
	return m.ServerID
}

func (m *FedVoteVolunteerMsg) LogFields() log.Fields {
	return log.Fields{"category": "message", "messagetype": "FedVoteVolunteerMsg", "dbheight": m.DBHeight, "newleader": m.ServerID.String()[4:12]}
}

func (m *FedVoteVolunteerMsg) GetRepeatHash() interfaces.IHash {
	return m.GetMsgHash()
}

// We have to return the hash of the underlying message.

func (m *FedVoteVolunteerMsg) GetHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *FedVoteVolunteerMsg) GetTimestamp() interfaces.Timestamp {
	return m.TS
}

func (m *FedVoteVolunteerMsg) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
		data, err := m.MarshalBinary()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

func (m *FedVoteVolunteerMsg) Type() byte {
	return constants.VOLUNTEERAUDIT
}

func (m *FedVoteVolunteerMsg) GetVolunteerMessage() interfaces.ISignableElectionMsg {
	return m
}

func (m *FedVoteVolunteerMsg) ElectionValidate(ie interfaces.IElections) int {
	// Set the super and let the base validate
	m.FedVoteMsg.Super = m
	return m.FedVoteMsg.ElectionValidate(ie)
}

func (m *FedVoteVolunteerMsg) Validate(is interfaces.IState) int {
	// Set the super and let the base validate
	m.FedVoteMsg.Super = m
	valid := m.FedVoteMsg.Validate(is)
	if valid <= 0 {
		return valid
	}

	// If valid is 1
	if is.(*state.State).Elections.(*elections.Elections).Electing < 0 {
		return 0
	}
	return valid
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *FedVoteVolunteerMsg) ComputeVMIndex(state interfaces.IState) {
}

// Acknowledgements do not go into the process list.
func (e *FedVoteVolunteerMsg) Process(dbheight uint32, state interfaces.IState) bool {
	panic("Ack object should never have its Process() method called")
}

func (e *FedVoteVolunteerMsg) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *FedVoteVolunteerMsg) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (m *FedVoteVolunteerMsg) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()

	buf := primitives.NewBuffer(data)
	if t, e := buf.PopByte(); e != nil || t != m.Type() {
		return newData, errors.New("Not a Volunteer Audit type")
	}
	if m.TS, err = buf.PopTimestamp(); err != nil {
		return newData, err
	}
	if m.Name, err = buf.PopString(); err != nil {
		return newData, err
	}
	if m.SigType, err = buf.PopBool(); err != nil {
		return newData, err
	}
	if m.ServerIdx, err = buf.PopUInt32(); err != nil {
		return newData, err
	}
	if m.ServerID, err = buf.PopIHash(); err != nil {
		return newData, err
	}
	if m.ServerName, err = buf.PopString(); err != nil {
		return newData, err
	}
	if m.FedIdx, err = buf.PopUInt32(); err != nil {
		return newData, err
	}
	if m.FedID, err = buf.PopIHash(); err != nil {
		return newData, err
	}
	if m.DBHeight, err = buf.PopUInt32(); err != nil {
		return newData, err
	}
	if m.VMIndex, err = buf.PopInt(); err != nil {
		return newData, err
	}
	if m.Minute, err = buf.PopByte(); err != nil {
		return newData, err
	}
	if m.Weight, err = buf.PopIHash(); err != nil {
		return newData, err
	}
	if m.Ack, err = buf.PopMsg(); err != nil {
		return newData, err
	}
	if m.Missing, err = buf.PopMsg(); err != nil {
		return newData, err
	}
	if m.Round, err = buf.PopInt(); err != nil {
		return newData, err
	}

	m.Signature = new(primitives.Signature)
	err = buf.PopBinaryMarshallable(m.Signature)
	if err != nil {
		return nil, err
	}
	newData = buf.Bytes()
	return
}

func (m *FedVoteVolunteerMsg) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *FedVoteVolunteerMsg) Sign(key interfaces.Signer) error {
	signature, err := msgbase.SignSignable(m, key)
	if err != nil {
		return err
	}
	m.Signature = signature
	return nil
}

func (m *FedVoteVolunteerMsg) GetSignature() interfaces.IFullSignature {
	return m.Signature
}

func (m *FedVoteVolunteerMsg) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "FedVoteVolunteerMsg.MarshalBinary err:%v", *pe)
		}
	}(&err)
	var buf primitives.Buffer

	data, err := m.MarshalForSignature()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	if m.Signature != nil {
		err = buf.PushBinaryMarshallable(m.Signature)
		if err != nil {
			return nil, err
		}
	}

	return buf.DeepCopyBytes(), nil
}

func (m *FedVoteVolunteerMsg) MarshalForSignature() (data []byte, err error) {
	var buf primitives.Buffer

	if err = buf.PushByte(m.Type()); err != nil {
		return nil, err
	}
	if e := buf.PushTimestamp(m.TS); e != nil {
		return nil, e
	}
	if e := buf.PushString(m.Name); e != nil {
		return nil, e
	}
	if e := buf.PushBool(m.SigType); e != nil {
		return nil, e
	}
	if e := buf.PushUInt32(m.ServerIdx); e != nil {
		return nil, e
	}
	if e := buf.PushIHash(m.ServerID); e != nil {
		return nil, e
	}
	if e := buf.PushString(m.ServerName); e != nil {
		return nil, e
	}
	if e := buf.PushUInt32(m.FedIdx); e != nil {
		return nil, e
	}
	if e := buf.PushIHash(m.FedID); e != nil {
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

	if m.Weight == nil {
		m.Weight = primitives.NewZeroHash()
	}
	if e := buf.PushIHash(m.Weight); e != nil {
		return nil, e
	}
	if e := buf.PushMsg(m.Ack); e != nil {
		return nil, e
	}
	if e := buf.PushMsg(m.Missing); e != nil {
		return nil, e
	}
	if e := buf.PushInt(m.Round); e != nil {
		return nil, e
	}
	data = buf.DeepCopyBytes()
	return data, nil
}

func (m *FedVoteVolunteerMsg) String() string {
	if m.LeaderChainID == nil {
		m.LeaderChainID = primitives.NewZeroHash()
	}
	return fmt.Sprintf("%19s %20s %20s ID: %x weight %x serverIdx: %d vmIdx: %d round %d dbheight: %d minute: %d ",
		m.Name,
		"Volunteer Audit",
		m.TS.String(),
		m.ServerID.Bytes()[2:5],
		m.Weight.Bytes()[2:5],
		m.ServerIdx,
		m.VMIndex,
		m.Round,
		m.DBHeight,
		m.Minute)
}
