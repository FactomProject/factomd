// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package electionMsgs

import (
	"fmt"

	"github.com/PaulSnow/factom2d/common/constants"
	"github.com/PaulSnow/factom2d/common/interfaces"
	"github.com/PaulSnow/factom2d/common/messages/msgbase"
	"github.com/PaulSnow/factom2d/common/primitives"
	"github.com/PaulSnow/factom2d/elections"

	llog "github.com/PaulSnow/factom2d/log"
	log "github.com/sirupsen/logrus"
)

//General acknowledge message
type RemoveAuditInternal struct {
	msgbase.MessageBase
	NName    string
	ServerID interfaces.IHash // Hash of message acknowledged
	DBHeight uint32           // Directory Block Height that owns this ack
	Height   uint32           // Height of this ack in this process list
}

var _ interfaces.IMsg = (*RemoveAuditInternal)(nil)
var _ interfaces.IElectionMsg = (*RemoveAuditInternal)(nil)

func (m *RemoveAuditInternal) MarshalBinary() (data []byte, err error) {
	var buf primitives.Buffer

	if err = buf.PushByte(constants.INTERNALREMOVEAUDIT); err != nil {
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

func (m *RemoveAuditInternal) GetMsgHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "RemoveAuditInternal.GetMsgHash") }()

	if m.MsgHash == nil {
		data, err := m.MarshalBinary()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

func (m *RemoveAuditInternal) ElectionProcess(state interfaces.IState, elect interfaces.IElections) {
	e, ok := elect.(*elections.Elections)
	if !ok {
		panic("Invalid elections object")
	}
	idx := -1
	for i, s := range e.Audit {
		if s.GetChainID().IsSameAs(m.GetServerID()) {
			idx = i
			break
		}
	}
	if idx != -1 {
		e.Audit = append(e.Audit[:idx], e.Audit[idx+1:]...)
	}
}

func (m *RemoveAuditInternal) GetServerID() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "RemoveAuditInternal.GetServerID") }()

	return m.ServerID
}

func (m *RemoveAuditInternal) LogFields() log.Fields {
	return log.Fields{"category": "message", "messagetype": "RemoveAuditInternal", "dbheight": m.DBHeight, "newleader": m.ServerID.String()[4:12]}
}

func (m *RemoveAuditInternal) GetRepeatHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "RemoveAuditInternal.GetRepeatHash") }()

	return m.GetMsgHash()
}

// We have to return the hash of the underlying message.
func (m *RemoveAuditInternal) GetHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "RemoveAuditInternal.GetHash") }()

	return m.GetMsgHash()
}

func (m *RemoveAuditInternal) GetTimestamp() interfaces.Timestamp {
	return primitives.NewTimestampNow()
}

func (m *RemoveAuditInternal) Type() byte {
	return constants.INTERNALREMOVEAUDIT
}

func (m *RemoveAuditInternal) Validate(state interfaces.IState) int {
	return 1
}

func (m *RemoveAuditInternal) ElectionValidate(ie interfaces.IElections) int {
	return 1
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *RemoveAuditInternal) ComputeVMIndex(state interfaces.IState) {
}

// Execute the leader functions of the given message
// Leader, follower, do the same thing.
func (m *RemoveAuditInternal) LeaderExecute(state interfaces.IState) {
	m.FollowerExecute(state)
}

func (m *RemoveAuditInternal) FollowerExecute(state interfaces.IState) {

}

// Acknowledgements do not go into the process list.
func (e *RemoveAuditInternal) Process(dbheight uint32, state interfaces.IState) bool {
	panic("Ack object should never have its Process() method called")
}

func (e *RemoveAuditInternal) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *RemoveAuditInternal) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (m *RemoveAuditInternal) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
			llog.LogPrintf("recovery", "Error unmarshalling: %v", r)
		}
	}()
	return
}

func (m *RemoveAuditInternal) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *RemoveAuditInternal) String() string {
	if m.LeaderChainID == nil {
		m.LeaderChainID = primitives.NewZeroHash()
	}
	return fmt.Sprintf("%20s %x %10s dbheight %d", "Remove Audit Internal", m.ServerID.Bytes(), m.NName, m.DBHeight)
}

func (a *RemoveAuditInternal) IsSameAs(b *RemoveAuditInternal) bool {
	return true
}
