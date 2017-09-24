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
	log "github.com/FactomProject/logrus"
	"github.com/FactomProject/factomd/state"
)

//General acknowledge message
type TimeoutInternal struct {
	msgbase.MessageBase
	NName       string
	DBHeight    int
	Minute      int
	Round       int
	MessageHash interfaces.IHash
}

var _ interfaces.IMsg = (*TimeoutInternal)(nil)

/*
type Elections struct {
	ServerID  interfaces.IHash
	Name      string
	Sync      []bool
	Federated []interfaces.IServer
	Audit     []interfaces.IServer
	FPriority []interfaces.IHash
	APriority []interfaces.IHash
	DBHeight  int
	Minute    int
	Input     interfaces.IQueue
	Output    interfaces.IQueue
	Round     []int
	Electing  int

	LeaderElecting  int // This is the federated Server we are electing, if we are a leader
	LeaderVolunteer int // This is the volunteer that we expect
}
*/

func (m *TimeoutInternal) ElectionProcess(is interfaces.IState, elect interfaces.IElections) {
	e, ok := elect.(*elections.Elections)
	if !ok {
		panic("Invalid elections object")
	}

	// We have advanced, so do nothing.
	if e.DBHeight > m.DBHeight || e.Minute > m.Minute {

		fmt.Printf("eee %10s %20s e.DBHeight %d m.DBHeight %d e.Minute %d m.Minute %d m.Round %d\n",
			e.State.GetFactomNodeName(),
			"Close Timeout",
			e.DBHeight,
			m.DBHeight,
			e.Minute,
			m.Minute,
			m.Round)

		e.Round = e.Round[:0]
		e.Electing = -1
		return
	}

	cnt := 0
	e.Electing = -1
	for i, b := range e.Sync {
		if !b {
			cnt++
			if e.Electing < 0 {
				e.Electing = i
			}
		}
	}
	// Hey, if all is well, then continue.
	if e.Electing < 0 {
		return
	}

	for len(e.Round) <= e.Electing {
		e.Round = append(e.Round, 0)
	}

	// New timeout, new round of elections.
	e.Round[e.Electing]++

	// If we don't have all our sync messages, we will have to come back around and see if all is well.
	go Fault(e, int(m.DBHeight), int(m.Minute), e.Round[e.Electing])

	fmt.Printf("eee %10s %20s Server Index: %d Round: %d %10s\n",
		e.State.GetFactomNodeName(),
		"Timeout",
		e.Electing,
		e.Round[e.Electing],
		e.Name)

	// Can we see a majority of the federated servers?
	if cnt <= len(e.Federated)/2 {
		// Reset the timeout and give up if we can't see a majority.
		return
	}
	fmt.Printf("eee %10s %10s on %10s\n", e.State.GetFactomNodeName(), "Fault", e.Name)

	// Get the priority order list of audit servers in the priority order
	e.APriority = Order(e.Audit, e.DBHeight, e.Minute, e.Electing, e.Round[e.Electing])

	idx := e.LeaderIndex(e.ServerID)
	// We are a leader
	if idx >= 0 {

	}

	idx = e.AuditIndex(e.ServerID)
	if idx >= 0 {
		serverMap := state.MakeMap(len(e.Federated), e.DBHeight)
		fmt.Printf("eee %10s %20s\n", e.Name, "I'm an Audit Server")
		auditIdx := MaxIdx(e.APriority)
		if idx == auditIdx {
			V := new(VolunteerAudit)
			V.TS = primitives.NewTimestampNow()
			V.NName = e.Name
			V.ServerIdx = uint32(e.Electing)
			V.ServerID = e.ServerID
			V.Weight = e.APriority[idx]
			V.DBHeight = uint32(e.DBHeight)
			V.Minute = byte(e.Minute)
			V.Round = e.Round[e.Electing]
			fmt.Printf("eee %10s %20s %s\n", e.Name, "I'm an Audit Server and I Volunteer", V.String())
			V.SendOut(is, V)
		}
	}
}

func (m *TimeoutInternal) GetServerID() interfaces.IHash {
	return nil
}

func (m *TimeoutInternal) LogFields() log.Fields {
	return log.Fields{"category": "message", "messagetype": "TimeoutInternal", "dbheight": m.DBHeight}
}

func (m *TimeoutInternal) GetRepeatHash() interfaces.IHash {
	return m.GetMsgHash()
}

// We have to return the haswh of the underlying message.
func (m *TimeoutInternal) GetHash() interfaces.IHash {
	return m.MessageHash
}

func (m *TimeoutInternal) GetTimestamp() interfaces.Timestamp {
	return primitives.NewTimestampNow()
}

func (m *TimeoutInternal) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
	}
	return m.MsgHash
}

func (m *TimeoutInternal) Type() byte {
	return constants.INTERNALTIMEOUT
}

func (m *TimeoutInternal) Validate(state interfaces.IState) int {
	return 1
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *TimeoutInternal) ComputeVMIndex(state interfaces.IState) {
}

// Execute the leader functions of the given message
// Leader, follower, do the same thing.
func (m *TimeoutInternal) LeaderExecute(state interfaces.IState) {
	m.FollowerExecute(state)
}

func (m *TimeoutInternal) FollowerExecute(state interfaces.IState) {

}

// Acknowledgements do not go into the process list.
func (e *TimeoutInternal) Process(dbheight uint32, state interfaces.IState) bool {
	panic("Ack object should never have its Process() method called")
}

func (e *TimeoutInternal) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *TimeoutInternal) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (m *TimeoutInternal) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	return
}

func (m *TimeoutInternal) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *TimeoutInternal) MarshalBinary() (data []byte, err error) {
	return
}

func (m *TimeoutInternal) String() string {
	if m.LeaderChainID == nil {
		m.LeaderChainID = primitives.NewZeroHash()
	}
	return fmt.Sprintf(" %20s %10s dbheight %d", "Add Audit Internal", m.NName, m.DBHeight)
}

func (a *TimeoutInternal) IsSameAs(b *TimeoutInternal) bool {
	return true
}
