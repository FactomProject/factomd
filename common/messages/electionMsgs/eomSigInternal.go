// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package electionMsgs

import (
	"errors"
	"fmt"

	"time"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages/msgbase"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/elections"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/util/atomic"
	log "github.com/sirupsen/logrus"
)

//General acknowledge message
type EomSigInternal struct {
	msgbase.MessageBase
	NName       string
	SigType     bool             // True of EOM, False if DBSig
	ServerID    interfaces.IHash // Hash of message acknowledged
	DBHeight    uint32           // Directory Block Height that owns this ack
	Height      uint32           // Height of this ack in this process list
	MessageHash interfaces.IHash
}

var _ interfaces.IMsg = (*EomSigInternal)(nil)
var _ interfaces.IElectionMsg = (*EomSigInternal)(nil)

func Title() string {
	return fmt.Sprintf("%5s%6s %10s %8s %8s %8s %8s",
		"", // Spacer
		"Type",
		"Node",
		"M:DBHt",
		"M:Min",
		"E:DBHt",
		"E:Min")
}

func (m *EomSigInternal) MarshalBinary() (data []byte, err error) {
	var buf primitives.Buffer

	if err = buf.PushByte(constants.INTERNALEOMSIG); err != nil {
		return nil, err
	}
	if e := buf.PushIHash(m.ServerID); e != nil {
		return nil, e
	}
	if e := buf.PushInt(int(m.DBHeight)); e != nil {
		return nil, e
	}
	if e := buf.PushByte(m.Minute); e != nil {
		return nil, e
	}
	if e := buf.PushByte(m.Minute); e != nil {
		return nil, e
	}
	data = buf.Bytes()
	return data, nil
}

func (m *EomSigInternal) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
		data, err := m.MarshalBinary()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}
func Fault(e *elections.Elections, dbheight int, minute int, round int, timeOutId int, currentTimeoutId *atomic.AtomicInt, sigtype bool) {
	//	e.LogPrintf("election", "Start Timeout %d", timeOutId)
	time.Sleep(e.Timeout)

	if currentTimeoutId.Load() == timeOutId {
		//		e.LogPrintf("election", "Timeout %d", timeOutId)
		/* we have NOT moved on so no timeout */
		timeout := new(TimeoutInternal)
		timeout.DBHeight = dbheight
		timeout.Minute = byte(minute)
		timeout.Round = round
		timeout.SigType = sigtype
		e.Input.Enqueue(timeout)
	} else {
		//		e.LogPrintf("election", "Cancel Timeout %d", timeOutId)

	}
}

func (m *EomSigInternal) ElectionProcess(is interfaces.IState, elect interfaces.IElections) {
	e := elect.(*elections.Elections) // Could check, but a nil pointer error is just as good.
	s := is.(*state.State)            // Same here.

	if m.ServerID == nil {
		return // Someone could send us a msg with a nil chainid
	}
	idx := e.LeaderIndex(m.ServerID)
	if idx == -1 {
		return // EOM but not from a server, just ignore it.
	}

	// We only do this once, as we transition into a sync event.
	// Either the height has incremented, or the minute has incremented.
	mv := int(m.DBHeight) > e.DBHeight || int(m.Minute) > e.Minute

	if mv {
		// Set our Identity Chain (Just in case it has changed.)
		e.FedID = s.IdentityChainID

		// Reset elections as we moved forward
		if int(m.DBHeight) > e.DBHeight && e.Electing != -1 {
			e.Electing = -1
		}
		e.DBHeight = int(m.DBHeight)
		e.Minute = int(m.Minute)
		e.Msgs = append(e.Msgs[:0], m)
		e.Sync = make([]bool, len(e.Federated))
		// Set the title in the state
		s.Election0 = Title()

		// Start our timer to timeout this sync
		round := 0

		e.FaultId.Store(e.FaultId.Load() + 1) // increment the timeout counter
		go Fault(e, e.DBHeight, e.Minute, round, e.FaultId.Load(), &e.FaultId, m.SigType)

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
	} else {
		e.Msgs = append(e.Msgs, m)
	}

	s.Election2 = e.FeedBackStr("m", true, idx)
	if len(e.Sync) <= idx {
		panic(errors.New("e.sync too short"))
	}

	e.Sync[idx] = true // Mark the leader at idx as synced.
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

// We have to return the hash of the underlying message.
func (m *EomSigInternal) GetHash() interfaces.IHash {
	return m.MessageHash
}

func (m *EomSigInternal) GetTimestamp() interfaces.Timestamp {
	return primitives.NewTimestampNow()
}

func (m *EomSigInternal) Type() byte {
	return constants.INTERNALEOMSIG
}

func (m *EomSigInternal) Validate(state interfaces.IState) int {
	return 1
}

func (m *EomSigInternal) ElectionValidate(ie interfaces.IElections) int {
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
