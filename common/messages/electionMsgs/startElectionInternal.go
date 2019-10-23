// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package electionMsgs

import (
	"fmt"
	"reflect"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages/msgbase"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/elections"
	"github.com/FactomProject/factomd/state"

	llog "github.com/FactomProject/factomd/log"
	log "github.com/sirupsen/logrus"
)

//General acknowledge message
type StartElectionInternal struct {
	msgbase.MessageBase

	VMHeight       int
	DBHeight       uint32
	PreviousDBHash interfaces.IHash
	SigType        bool
	IsLeader       bool
}

var _ interfaces.IMsg = (*StartElectionInternal)(nil)
var _ interfaces.IElectionMsg = (*StartElectionInternal)(nil)

func (m *StartElectionInternal) ElectionProcess(s interfaces.IState, elect interfaces.IElections) {
	e := elect.(*elections.Elections)

	// If the electing is set to -1, that election has ended before we got to start it.
	// Still trigger the Fault loop, it will self terminate if we've moved forward
	if e.Electing == -1 {
		go Fault(e, e.DBHeight, e.Minute, e.FaultId.Load(), &e.FaultId, m.SigType, e.RoundTimeout)
		return
	}
	e.Adapter = NewElectionAdapter(e, m.PreviousDBHash)
	s.LogPrintf("election", "Create Election Adapter")
	// An election that finishes may make us a leader. We need to know that for the next election that
	// takes place. So use the election's list of fed servers to determine if we are a leader
	for _, id := range e.Federated {
		if id.GetChainID().IsSameAs(s.GetIdentityChainID()) {
			e.Adapter.SetObserver(false)
			break
		}
		e.Adapter.SetObserver(true)
	}

	go Fault(e, e.DBHeight, e.Minute, e.FaultId.Load(), &e.FaultId, m.SigType, e.RoundTimeout)
}

// Execute the leader functions of the given message
// Leader, follower, do the same thing.
func (m *StartElectionInternal) LeaderExecute(state interfaces.IState) {
	m.FollowerExecute(state)
}

func (m *StartElectionInternal) FollowerExecute(is interfaces.IState) {
	s := is.(*state.State)
	m.IsLeader = is.IsLeader()
	// TODO: State related things about starting an election
	pl := s.ProcessLists.Get(m.DBHeight)
	if pl == nil {
		//s.Holding[m.GetHash().Fixed()] = m
		s.AddToHolding(m.GetMsgHash().Fixed(), m) // StartElectionInternal.FollowerExecute
		return
	}
	vm := pl.VMs[m.VMIndex]
	if vm == nil {
		return
	}

	// Process all the messages that we can
	for s.LeaderPL.Process(s) {
	}

	m.VMHeight = vm.Height

	// Send to elections
	is.ElectionsQueue().Enqueue(m)
}

func (m *StartElectionInternal) ElectionValidate(ie interfaces.IElections) int {
	if int(m.DBHeight) < ie.(*elections.Elections).DBHeight {
		return -1
	}
	return 1
}

func (m *StartElectionInternal) Validate(state interfaces.IState) int {
	return 1
}

func (m *StartElectionInternal) MarshalBinary() (data []byte, err error) {
	//var buf primitives.Buffer

	return nil, fmt.Errorf("Not implmented for StartElectionInternal")
}

func (m *StartElectionInternal) GetMsgHash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("StartElectionInternal.GetMsgHash() saw an interface that was nil")
		}
	}()

	// Internal messages don't have marshal code. Give them some hash to be happy
	if m.MsgHash == nil {
		m.MsgHash = primitives.RandomHash()
	}
	return m.MsgHash
}

func (m *StartElectionInternal) LogFields() log.Fields {
	return log.Fields{"category": "message", "messagetype": "StartElectionInternal", "dbheight": m.DBHeight}
}

func (m *StartElectionInternal) GetRepeatHash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("StartElectionInternal.GetRepeatHash() saw an interface that was nil")
		}
	}()

	return m.GetMsgHash()
}

// We have to return the hash of the underlying message.
func (m *StartElectionInternal) GetHash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("StartElectionInternal.GetHash() saw an interface that was nil")
		}
	}()

	return m.GetMsgHash()
}

func (m *StartElectionInternal) GetTimestamp() interfaces.Timestamp {
	return primitives.NewTimestampNow()
}

func (m *StartElectionInternal) Type() byte {
	return constants.INTERNALSTARTELECTION
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *StartElectionInternal) ComputeVMIndex(state interfaces.IState) {
}

// Acknowledgements do not go into the process list.
func (e *StartElectionInternal) Process(dbheight uint32, state interfaces.IState) bool {
	panic("Ack object should never have its Process() method called")
}

func (e *StartElectionInternal) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *StartElectionInternal) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (m *StartElectionInternal) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
			llog.LogPrintf("recovery", "Error unmarshalling: %v", r)
		}
	}()
	return nil, fmt.Errorf("Not implmented for StartElectionInternal")
}

func (m *StartElectionInternal) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *StartElectionInternal) String() string {
	return fmt.Sprintf("%20s dbheight %d min %d vm %d", "Start Election Internal", m.DBHeight, int(m.Minute), m.VMIndex)
}

func (a *StartElectionInternal) IsSameAs(b *StartElectionInternal) bool {
	return true
}
