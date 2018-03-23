// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package electionMsgs

import (
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages/msgbase"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/elections"
	"github.com/FactomProject/factomd/state"
	log "github.com/sirupsen/logrus"
)

//General acknowledge message
type StartElectionInternal struct {
	msgbase.MessageBase

	VMHeight       int
	DBHeight       uint32
	PreviousDBHash interfaces.IHash
	IsLeader       bool
}

var _ interfaces.IMsg = (*StartElectionInternal)(nil)
var _ interfaces.IElectionMsg = (*StartElectionInternal)(nil)

func (m *StartElectionInternal) ElectionProcess(s interfaces.IState, elect interfaces.IElections) {
	e := elect.(*elections.Elections)

	e.Adapter = NewElectionAdapter(e, m.PreviousDBHash)
	e.Adapter.SetObserver(!m.IsLeader)
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
		s.Holding[m.GetHash().Fixed()] = m
	}
	vm := pl.VMs[m.VMIndex]
	if vm == nil {
		return
	}

	end := len(vm.List)
	if end > vm.Height {
		for _, msg := range vm.List[vm.Height:] {
			if msg != nil {
				hash := msg.GetRepeatHash()
				s.Replay.Clear(constants.INTERNAL_REPLAY, hash.Fixed())
				s.Holding[msg.GetMsgHash().Fixed()] = msg
			}
		}
	}

	m.VMHeight = vm.Height
	vm.List = vm.List[:vm.Height]
	vm.ListAck = vm.ListAck[:vm.Height]
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

func (m *StartElectionInternal) GetMsgHash() interfaces.IHash {
	// Internal messages don't have marshal code. Give them some hash to be happy
	if m.MsgHash == nil {
		m.MsgHash = primitives.RandomHash()
	}
	return m.MsgHash
}

func (m *StartElectionInternal) LogFields() log.Fields {
	return log.Fields{"category": "message", "messagetype": "StartElectionInternal", "dbheight": m.DBHeight}
}

func (m *StartElectionInternal) GetRepeatHash() interfaces.IHash {
	return m.GetMsgHash()
}

// We have to return the hash of the underlying message.
func (m *StartElectionInternal) GetHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *StartElectionInternal) GetTimestamp() interfaces.Timestamp {
	return primitives.NewTimestampNow()
}

func (m *StartElectionInternal) Type() byte {
	return constants.INTERNALAUTHLIST
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
		}
	}()
	return nil, fmt.Errorf("Not implmented for StartElectionInternal")
}

func (m *StartElectionInternal) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *StartElectionInternal) String() string {
	return "Not implemented, StartElectionInternal"
}

func (a *StartElectionInternal) IsSameAs(b *StartElectionInternal) bool {
	return true
}
