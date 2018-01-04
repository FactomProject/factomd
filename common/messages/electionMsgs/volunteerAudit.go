// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package electionMsgs

import (
	"bytes"
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages/msgbase"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/elections"
	"github.com/FactomProject/goleveldb/leveldb/errors"
	log "github.com/sirupsen/logrus"
)

var _ = fmt.Print

//General acknowledge message
type VolunteerAudit struct {
	msgbase.MessageBase
	TS          interfaces.Timestamp // Message Timestamp
	EOM         bool                 // True if an EOM, false if a DBSig
	Name        string               // Server name
	ServerIdx   uint32               // Index of Server replacing
	ServerID    interfaces.IHash     // Volunteer Server ChainID
	Weight      interfaces.IHash     // Computed Weight at this DBHeight, Minute, Round
	DBHeight    uint32               // Directory Block Height that owns this ack
	Minute      byte                 // Minute (-1 for dbsig)
	Round       int                  // Voting Round
	Missing     interfaces.IMsg      // The Missing DBSig or EOM
	Ack         interfaces.IMsg      // The acknowledgement for the missing message
	messageHash interfaces.IHash
}

func (m *VolunteerAudit) ElectionProcess(is interfaces.IState, elect interfaces.IElections) {
	e := elect.(*elections.Elections)
	idx := e.LeaderIndex(is.GetIdentityChainID())
	if idx > 0 {

	}
}

var _ interfaces.IMsg = (*VolunteerAudit)(nil)

func (a *VolunteerAudit) IsSameAs(msg interfaces.IMsg) bool {
	b, ok := msg.(*VolunteerAudit)
	if !ok {
		return false
	}
	if a.TS.GetTimeMilli() != b.TS.GetTimeMilli() {
		return false
	}
	if a.Name != b.Name {
		return false
	}
	if a.EOM != b.EOM {
		return false
	}
	if a.ServerIdx != b.ServerIdx {
		return false
	}
	if a.ServerID.Fixed() != b.ServerID.Fixed() {
		return false
	}
	if a.Weight.Fixed() != b.Weight.Fixed() {
		return false
	}
	if a.DBHeight != b.DBHeight {
		return false
	}
	if a.VMIndex != b.VMIndex {
		return false
	}
	if a.Round != b.Round {
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

func (m *VolunteerAudit) GetServerID() interfaces.IHash {
	return m.ServerID
}

func (m *VolunteerAudit) LogFields() log.Fields {
	return log.Fields{"category": "message", "messagetype": "VolunteerAudit", "dbheight": m.DBHeight, "newleader": m.ServerID.String()[4:12]}
}

func (m *VolunteerAudit) GetRepeatHash() interfaces.IHash {
	return m.GetMsgHash()
}

// We have to return the haswh of the underlying message.
func (m *VolunteerAudit) GetHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *VolunteerAudit) GetTimestamp() interfaces.Timestamp {
	return m.TS
}

func (m *VolunteerAudit) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
		data, err := m.MarshalBinary()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

func (m *VolunteerAudit) Type() byte {
	return constants.VOLUNTEERAUDIT
}

func (m *VolunteerAudit) Validate(state interfaces.IState) int {
	return 1
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *VolunteerAudit) ComputeVMIndex(state interfaces.IState) {
}

// Execute the leader functions of the given message
// Leader, follower, do the same thing.
func (m *VolunteerAudit) LeaderExecute(state interfaces.IState) {
	m.FollowerExecute(state)
}

func (m *VolunteerAudit) FollowerExecute(state interfaces.IState) {
	state.ElectionsQueue().Enqueue(m)
}

// Acknowledgements do not go into the process list.
func (e *VolunteerAudit) Process(dbheight uint32, state interfaces.IState) bool {
	panic("Ack object should never have its Process() method called")
}

func (e *VolunteerAudit) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *VolunteerAudit) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (m *VolunteerAudit) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		return
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()

	buf := primitives.NewBuffer(data)
	if t, e := buf.PopByte(); e != nil || t != constants.VOLUNTEERAUDIT {
		return nil, errors.New("Not a Volunteer Audit type")
	}
	if m.TS, err = buf.PopTimestamp(); err != nil {
		return nil, err
	}
	if m.Name, err = buf.PopString(); err != nil {
		return nil, err
	}
	if m.EOM, err = buf.PopBool(); err != nil {
		return nil, err
	}
	if m.ServerIdx, err = buf.PopUInt32(); err != nil {
		return nil, err
	}
	if m.ServerID, err = buf.PopIHash(); err != nil {
		return nil, err
	}
	if m.Weight, err = buf.PopIHash(); err != nil {
		return nil, err
	}
	if m.DBHeight, err = buf.PopUInt32(); err != nil {
		return nil, err
	}
	if m.VMIndex, err = buf.PopInt(); err != nil {
		return nil, err
	}
	if m.Round, err = buf.PopInt(); err != nil {
		return nil, err
	}
	if m.Minute, err = buf.PopByte(); err != nil {
		return nil, err
	}
	//	if m.Ack, err = buf.PopMsg(); err != nil {
	//		return nil, err
	//	}
	//	if m.Missing, err = buf.PopMsg(); err != nil {
	//		return nil, err
	//	}
	return buf.PopBytes()
}

func (m *VolunteerAudit) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *VolunteerAudit) MarshalBinary() (data []byte, err error) {
	var buf primitives.Buffer

	if e := buf.PushByte(constants.VOLUNTEERAUDIT); e != nil {
		return nil, e
	}
	if e := buf.PushTimestamp(m.TS); e != nil {
		return nil, e
	}
	if e := buf.PushString(m.Name); e != nil {
		return nil, e
	}
	if e := buf.PushBool(m.EOM); e != nil {
		return nil, e
	}
	if e := buf.PushUInt32(m.ServerIdx); e != nil {
		return nil, e
	}
	if e := buf.PushIHash(m.ServerID); e != nil {
		return nil, e
	}
	if e := buf.PushIHash(m.Weight); e != nil {
		return nil, e
	}
	if e := buf.PushUInt32(m.DBHeight); e != nil {
		return nil, e
	}
	if e := buf.PushInt(m.VMIndex); e != nil {
		return nil, e
	}
	if e := buf.PushInt(m.Round); e != nil {
		return nil, e
	}
	if e := buf.PushByte(m.Minute); e != nil {
		return nil, e
	}
	//if e := buf.PushMsg(m.Ack); e != nil {
	//	return nil, e
	//}
	//if e := buf.PushMsg(m.Missing); e != nil {
	//	return nil, e
	//}
	return buf.DeepCopyBytes(), nil
}

func (m *VolunteerAudit) String() string {
	if m.LeaderChainID == nil {
		m.LeaderChainID = primitives.NewZeroHash()
	}
	return fmt.Sprintf("%19s %20s %20s ID: %x WT: %x serverIdx: %d vmIdx: %d round: %d dbheight: %d minute: %d ",
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
