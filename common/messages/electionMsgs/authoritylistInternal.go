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
	log "github.com/sirupsen/logrus"
)

//General acknowledge message
type AuthorityListInternal struct {
	msgbase.MessageBase

	Federated []interfaces.IServer
	Audit     []interfaces.IServer
	DBHeight  uint32 // Directory Block Height that owns this ack
}

var _ interfaces.IMsg = (*AuthorityListInternal)(nil)
var _ interfaces.IElectionMsg = (*AuthorityListInternal)(nil)

func (m *AuthorityListInternal) MarshalBinary() (data []byte, err error) {
	//var buf primitives.Buffer

	return nil, nil
}

func (m *AuthorityListInternal) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
		data, err := m.MarshalBinary()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

func (m *AuthorityListInternal) ElectionProcess(s interfaces.IState, elect interfaces.IElections) {
	e := elect.(*elections.Elections)
	e.Federated = m.Federated
	e.Audit = m.Audit
	e.SetElections3()
}

func (m *AuthorityListInternal) LogFields() log.Fields {
	return log.Fields{"category": "message", "messagetype": "AuthorityListInternal", "dbheight": m.DBHeight}
}

func (m *AuthorityListInternal) GetRepeatHash() interfaces.IHash {
	return m.GetMsgHash()
}

// We have to return the hash of the underlying message.
func (m *AuthorityListInternal) GetHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *AuthorityListInternal) GetTimestamp() interfaces.Timestamp {
	return primitives.NewTimestampNow()
}

func (m *AuthorityListInternal) Type() byte {
	return constants.INTERNALAUTHLIST
}

func (m *AuthorityListInternal) ElectionValidate(ie interfaces.IElections) int {
	if int(m.DBHeight) < ie.(*elections.Elections).DBHeight {
		return -1
	}
	return 1
}

func (m *AuthorityListInternal) Validate(state interfaces.IState) int {
	return 1
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *AuthorityListInternal) ComputeVMIndex(state interfaces.IState) {
}

// Execute the leader functions of the given message
// Leader, follower, do the same thing.
func (m *AuthorityListInternal) LeaderExecute(state interfaces.IState) {
	m.FollowerExecute(state)
}

func (m *AuthorityListInternal) FollowerExecute(state interfaces.IState) {
	state.ElectionsQueue().Enqueue(m)
}

// Acknowledgements do not go into the process list.
func (e *AuthorityListInternal) Process(dbheight uint32, state interfaces.IState) bool {
	panic("Ack object should never have its Process() method called")
}

func (e *AuthorityListInternal) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *AuthorityListInternal) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (m *AuthorityListInternal) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	return
}

func (m *AuthorityListInternal) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *AuthorityListInternal) String() string {
	return "Not implemented, AuthorityListInternal"
}

func (a *AuthorityListInternal) IsSameAs(b *AuthorityListInternal) bool {
	return true
}
