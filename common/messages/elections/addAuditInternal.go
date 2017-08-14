// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package elections

import (
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	log "github.com/FactomProject/logrus"
	"github.com/FactomProject/factomd/common/messages/msgbase"
)

//General acknowledge message
type AddAuditInternal struct {
	msgbase.MessageBase
	NName       string
	ServerID    interfaces.IHash // Hash of message acknowledged
	DBHeight    uint32           // Directory Block Height that owns this ack
	Height      uint32           // Height of this ack in this process list
	MessageHash interfaces.IHash
}

var _ interfaces.IMsg = (*AddAuditInternal)(nil)

func (m *AddAuditInternal) GetServerID() interfaces.IHash {
	return m.ServerID
}

func (m *AddAuditInternal) LogFields() log.Fields {
	return log.Fields{"category": "message", "messagetype": "AddAuditInternal", "dbheight": m.DBHeight, "newleader": m.ServerID.String()[4:12]}
}

func (m *AddAuditInternal) GetRepeatHash() interfaces.IHash {
	return m.GetMsgHash()
}

// We have to return the haswh of the underlying message.
func (m *AddAuditInternal) GetHash() interfaces.IHash {
	return m.MessageHash
}

func (m *AddAuditInternal) GetTimestamp() interfaces.Timestamp {
	return primitives.NewTimestampNow()
}

func (m *AddAuditInternal) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
	}
	return m.MsgHash
}

func (m *AddAuditInternal) Type() byte {
	return constants.INTERNALADDAUDIT
}

func (m *AddAuditInternal) Validate(state interfaces.IState) int {
	return 1
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *AddAuditInternal) ComputeVMIndex(state interfaces.IState) {
}

// Execute the leader functions of the given message
// Leader, follower, do the same thing.
func (m *AddAuditInternal) LeaderExecute(state interfaces.IState) {
	m.FollowerExecute(state)
}

func (m *AddAuditInternal) FollowerExecute(state interfaces.IState) {

}

// Acknowledgements do not go into the process list.
func (e *AddAuditInternal) Process(dbheight uint32, state interfaces.IState) bool {
	panic("Ack object should never have its Process() method called")
}

func (e *AddAuditInternal) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *AddAuditInternal) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (m *AddAuditInternal) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	return
}

func (m *AddAuditInternal) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *AddAuditInternal) MarshalBinary() (data []byte, err error) {
	return
}

func (m *AddAuditInternal) String() string {
	if m.LeaderChainID == nil {
		m.LeaderChainID = primitives.NewZeroHash()
	}
	return fmt.Sprintf("%20s %x %10s dbheight %d", "Add Audit Internal", m.ServerID.Bytes(), m.NName, m.DBHeight)
}

func (a *AddAuditInternal) IsSameAs(b *AddAuditInternal) bool {
	return true
}
