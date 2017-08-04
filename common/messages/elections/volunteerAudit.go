// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package elections

import (
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	log "github.com/FactomProject/logrus"
)

//General acknowledge message
type VolunteerAudit struct {
	messages.MessageBase
	NName       string           // Server name
	ServerIdx   int              // Index of Server replacing
	ServerID    interfaces.IHash // Hash of message acknowledged
	DBHeight    uint32           // Directory Block Height that owns this ack
	Minute      byte
	Height      uint32 // Height of this ack in this process list
	MessageHash interfaces.IHash
}

var _ interfaces.IMsg = (*VolunteerAudit)(nil)

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
	return m.MessageHash
}

func (m *VolunteerAudit) GetTimestamp() interfaces.Timestamp {
	return primitives.NewTimestampNow()
}

func (m *VolunteerAudit) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
	}
	return m.MsgHash
}

func (m *VolunteerAudit) Type() byte {
	return constants.INTERNALADDLEADER
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
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	buf := primitives.NewBuffer(data)
	if m.NName, err = buf.PopString(); err != nil {
		return nil, err
	}
	if m.ServerIdx, err = buf.PopInt(); err != nil {
		return nil, err
	}
	ServerID, e := buf.PopBytes()
	if e != nil {
		return nil, err
	}
	m.ServerID = primitives.NewHash(ServerID)
	if m.DBHeight, err = buf.PopUInt32(); err != nil {
		return nil, err
	}
	if m.Minute, err = buf.PopByte(); err != nil {
		return nil, err
	}
	if m.Height, err = buf.PopUInt32(); err != nil {
		return nil, err
	}
	return buf.PopBytes()
}

func (m *VolunteerAudit) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *VolunteerAudit) MarshalBinary() (data []byte, err error) {

	var buf primitives.Buffer

	if e := buf.PushString(m.NName); e != nil {
		return nil, e
	}
	if e := buf.PushInt(m.ServerIdx); e != nil {
		return nil, e
	}
	if e := buf.PushBytes(m.ServerID.Bytes()); e != nil {
		return nil, e
	}
	if e := buf.PushUInt32(m.DBHeight); e != nil {
		return nil, e
	}
	if e := buf.PushByte(m.Minute); e != nil {
		return nil, e
	}
	if e := buf.PushUInt32(m.Height); e != nil {
		return nil, e
	}
	return buf.PopBytes()
}

func (m *VolunteerAudit) String() string {
	if m.LeaderChainID == nil {
		m.LeaderChainID = primitives.NewZeroHash()
	}
	return fmt.Sprintf("%20s %x %10s dbheight %d", "Add Audit Internal", m.ServerID.Bytes(), m.NName, m.DBHeight)
}

func (a *VolunteerAudit) IsSameAs(b *VolunteerAudit) bool {
	return true
}
