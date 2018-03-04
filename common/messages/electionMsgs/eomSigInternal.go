// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package electionMsgs

import (
	"fmt"

	"time"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages/msgbase"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/elections"
	"github.com/FactomProject/factomd/state"
	log "github.com/sirupsen/logrus"
)

//General acknowledge message
type EomSigInternal struct {
	msgbase.MessageBase
	NName       string
	SigType     bool             // True of EOM, False if DBSig
	ServerID    interfaces.IHash // Hash of message acknowledged
	DBHeight    uint32           // Directory Block Height that owns this ack
	Minute      uint32           // Minute of the issue
	Height      uint32           // Height of this ack in this process list
	MessageHash interfaces.IHash
}

var _ interfaces.IMsg = (*EomSigInternal)(nil)

func Title() string {
	return fmt.Sprintf("%6s %10s %8s %8s %8s %8s",
		"Type",
		"Node",
		"M:DBHt",
		"M:Min",
		"E:DBHt",
		"E:Min")
}

func Fault(e *elections.Elections, dbheight int, minute int, vmIndex int, round int) {

	time.Sleep(e.Timeout)
	timeout := new(TimeoutInternal)
	timeout.Minute = byte(minute)
	timeout.DBHeight = dbheight
	timeout.VMIndex = vmIndex
	timeout.Round = round
	e.Input.Enqueue(timeout)

}

func (m *EomSigInternal) ElectionProcess(is interfaces.IState, elect interfaces.IElections) {
	e := elect.(*elections.Elections) // Could check, but a nil pointer error is just as good.
	s := is.(*state.State)            // Same here.

	// We only do this once, as we transition into a sync event.
	// Either the height has incremented, or the minute has incremented.
	mv := int(m.DBHeight) > e.DBHeight || int(m.Minute) > e.Minute
	if mv {
		// Set our Identity Chain (Just in case it has changed.)
		e.FedID = s.IdentityChainID

		e.DBHeight = int(m.DBHeight)
		e.Minute = int(m.Minute)
		e.Msgs = append(e.Msgs, m)
		e.Sync = make([]bool, len(e.Federated))
		// Set the title in the state
		s.Election0 = Title()

		// Start our timer to timeout this sync
		go Fault(e, int(m.DBHeight), int(m.Minute), m.VMIndex, 0)
		t := "EOM"
		if !m.SigType {
			t = "DBSig"
		}
		s.Election1 = fmt.Sprintf("%6s %10s %8d %8d %8d %8d",
			t,
			s.FactomNodeName,
			m.DBHeight,
			m.Minute,
			e.DBHeight,
			e.Minute)
	}
	idx := e.LeaderIndex(m.ServerID)
	s.Election2 = e.FeedBackStr("m", true, idx)
	if idx >= 0 {
		for len(e.Sync) <= idx {
			e.Sync = append(e.Sync, false)
		}
		e.Sync[idx] = true // Mark the leader at idx as synced.
	} else {
		return // Not a server, just ignore the while thing.
	}
	for _, b := range e.Sync {
		if !b {
			return // If any leader is not yet synced, then return.
		}
	}
	e.NewFeedback()
	s.Election2 = e.FeedBackStr("", true, 0)
	e.Round = e.Round[:0] // Get rid of any previous round counting.
}

func (m *EomSigInternal) GetServerID() interfaces.IHash {
	return m.ServerID
}

func (m *EomSigInternal) LogFields() log.Fields {
	return log.Fields{"category": "message", "messagetype": "EomSigInternal", "dbheight": m.DBHeight, "newleader": m.ServerID.String()[4:12]}
}

func (m *EomSigInternal) GetRepeatHash() interfaces.IHash {
	return m.GetMsgHash()
}

// We have to return the haswh of the underlying message.
func (m *EomSigInternal) GetHash() interfaces.IHash {
	return m.MessageHash
}

func (m *EomSigInternal) GetTimestamp() interfaces.Timestamp {
	return primitives.NewTimestampNow()
}

func (m *EomSigInternal) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
	}
	return m.MsgHash
}

func (m *EomSigInternal) Type() byte {
	return constants.INTERNALSIG
}

func (m *EomSigInternal) Validate(state interfaces.IState) int {
	return 1
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *EomSigInternal) ComputeVMIndex(state interfaces.IState) {
}

// Execute the leader functions of the given message
// Leader, follower, do the same thing.
func (m *EomSigInternal) LeaderExecute(state interfaces.IState) {
	m.FollowerExecute(state)
}

func (m *EomSigInternal) FollowerExecute(state interfaces.IState) {

}

// Acknowledgements do not go into the process list.
func (e *EomSigInternal) Process(dbheight uint32, state interfaces.IState) bool {
	panic("Ack object should never have its Process() method called")
}

func (e *EomSigInternal) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *EomSigInternal) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (m *EomSigInternal) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	return
}

func (m *EomSigInternal) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *EomSigInternal) MarshalBinary() (data []byte, err error) {
	return
}

func (m *EomSigInternal) String() string {
	if m.ServerID == nil {
		m.ServerID = primitives.NewZeroHash()
	}
	return fmt.Sprintf("%6s %10s %20s %x dbheight %5d minute %2d",
		"",
		m.NName,
		"EOM",
		m.ServerID.Bytes(),
		m.DBHeight,
		m.Minute)
}

func (a *EomSigInternal) IsSameAs(b *EomSigInternal) bool {
	return true
}

func (a *EomSigInternal) GetDBHeight() uint32 {
	return a.DBHeight
}
