// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package electionMsgs

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages/msgbase"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/elections"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/util/atomic"

	llog "github.com/FactomProject/factomd/log"
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
	return fmt.Sprintf("%5s%6s %10s %5s %5s %5s %5s %5s %5s",
		"", // Spacer
		"Type",
		"Node",
		"M:DBHt",
		"M:Min",
		"M:VM",
		"E:DBHt",
		"E:Min",
		"E:VM")
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

func (m *EomSigInternal) GetMsgHash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("EomSigInternal.GetMsgHash() saw an interface that was nil")
		}
	}()

	if m.MsgHash == nil {
		data, err := m.MarshalBinary()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}
func Fault(e *elections.Elections, dbheight int, minute int, timeOutId int, currentTimeoutId *atomic.AtomicInt, sigtype bool, timeoutDuration time.Duration) {
	//	e.LogPrintf("election", "Start Timeout %d", timeOutId)
	for !e.State.(*state.State).DBFinished || e.State.(*state.State).IgnoreMissing {
		time.Sleep(timeoutDuration)
	}
	time.Sleep(timeoutDuration)

	if currentTimeoutId.Load() == timeOutId {
		//		e.LogPrintf("election", "Timeout %d", timeOutId)
		/* we have NOT moved on so no timeout */
		timeout := new(TimeoutInternal)
		timeout.DBHeight = dbheight
		timeout.Minute = byte(minute)
		timeout.SigType = sigtype
		e.Input.Enqueue(timeout)
	} else {
		//		e.LogPrintf("election", "Cancel Timeout %d", timeOutId)

	}
}
func (m *EomSigInternal) ComparisonMinute() int {
	if !m.SigType {
		return -1
	}
	return int(m.Minute)
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

	//// We start sorting here on 6/28/18 at 12pm ...
	//if is.IsActive(activations.ELECTION_NO_SORT) {
	//	if int(m.DBHeight) > e.DBHeight {
	//		// Sort leaders, on block boundaries
	//		s := e.State
	//		s.LogPrintf("elections", "Election Sort FedServers EomSigInternal")
	//		changed := e.Sort(e.Federated)
	//		if changed {
	//			e.LogPrintf("election", "Sort changed e.Federated in EomSigInternal.ElectionProcess")
	//			e.LogPrintLeaders("election")
	//		}
	//		changed = e.Sort(e.Audit)
	//		if changed {
	//			e.LogPrintf("election", "Sort changed e.Audit in EomSigInternal.ElectionProcess")
	//			e.LogPrintLeaders("election")
	//		}
	//	}
	//}

	// We only do this once, as we transition into a sync event.
	// Either the height has incremented, or the minute has incremented.
	mv := int(m.DBHeight) > e.DBHeight || m.ComparisonMinute() > e.ComparisonMinute()

	if mv {
		// Set our Identity Chain (Just in case it has changed.)
		e.FedID = s.IdentityChainID

		// Reset elections as we moved forward
		if int(m.DBHeight) > e.DBHeight && e.Electing != -1 {
			e.Electing = -1
		}

		// Sort leaders, on block boundaries
		s.LogPrintf("elections", "Election Sort FedServers EomSigInternal2")
		changed := e.Sort(e.Federated)
		if changed {
			e.LogPrintf("election", "Sort changed e.Federated in EomSigInternal.ElectionProcess")
			e.LogPrintLeaders("election")
		}
		changed = e.Sort(e.Audit)
		if changed {
			e.LogPrintf("election", "Sort changed e.Audit in EomSigInternal.ElectionProcess")
			e.LogPrintLeaders("election")

		}

		e.DBHeight = int(m.DBHeight)
		e.Minute = int(m.Minute)
		e.SigType = m.SigType
		e.Msgs = append(e.Msgs[:0], m)
		e.Sync = make([]bool, len(e.Federated))
		// Set the title in the state
		s.Election0 = Title()

		e.FaultId.Store(e.FaultId.Load() + 1) // increment the timeout counter
		go Fault(e, e.DBHeight, e.Minute, e.FaultId.Load(), &e.FaultId, m.SigType, e.Timeout)

		// Drain all waiting messages as we have advanced, they can now be processed again
		// as moving forward in mins/blocks may invalidate/validate some messages
		go e.ProcessWaiting()

		t := "EOM"
		if !m.SigType {
			t = "DBSig"
		}

		e.SetElections3()

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

func (m *EomSigInternal) GetServerID() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("EomSigInternal.GetServerID() saw an interface that was nil")
		}
	}()

	return m.ServerID
}

func (m *EomSigInternal) LogFields() log.Fields {
	return log.Fields{"category": "message", "messagetype": "EomSigInternal", "dbheight": m.DBHeight, "newleader": m.ServerID.String()[4:12]}
}

func (m *EomSigInternal) GetRepeatHash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("EomSigInternal.GetRepeatHash() saw an interface that was nil")
		}
	}()

	return m.GetMsgHash()
}

// We have to return the hash of the underlying message.
func (m *EomSigInternal) GetHash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("EomSigInternal.GetHash() saw an interface that was nil")
		}
	}()

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
func (m *EomSigInternal) Process(dbheight uint32, state interfaces.IState) bool {
	panic("Ack object should never have its Process() method called")
}

func (m *EomSigInternal) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(m)
}

func (m *EomSigInternal) JSONString() (string, error) {
	return primitives.EncodeJSONString(m)
}

func (m *EomSigInternal) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
			llog.LogPrintf("recovery", "Error unmarshalling: %v", r)
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

func (m *EomSigInternal) IsSameAs(b *EomSigInternal) bool {
	return true
}

func (m *EomSigInternal) GetDBHeight() uint32 {
	return m.DBHeight
}

func (m *EomSigInternal) Label() string {
	return msgbase.GetLabel(m)
}

