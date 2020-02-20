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

	llog "github.com/FactomProject/factomd/log"
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

	return nil, fmt.Errorf("Not implmented for AuthorityListInternal")
}

var msgCount int

func (m *AuthorityListInternal) GetMsgHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "AuthorityListInternal.GetMsgHash") }()

	// because this is an internal only message it has no serialization so no real hash. make a fake hash so it
	// doesn't trigger pokemon tracking and can be logged.
	if m.MsgHash == nil {
		msgCount++
		m.MsgHash = new(primitives.Hash)
		m.MsgHash.PFixed()[0] = byte((msgCount >> 0) & 0xFF)
		m.MsgHash.PFixed()[1] = byte((msgCount >> 8) & 0xFF)
		m.MsgHash.PFixed()[2] = byte((msgCount >> 16) & 0xFF)
		m.MsgHash.PFixed()[3] = byte((msgCount >> 24) & 0xFF)
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

func (m *AuthorityListInternal) GetRepeatHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "AuthorityListInternal.GetRepeatHash") }()

	return m.GetMsgHash()
}

// We have to return the hash of the underlying message.
func (m *AuthorityListInternal) GetHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "AuthorityListInternal.GetHash") }()

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
			llog.LogPrintf("recovery", "Error unmarshalling: %v", r)
		}
	}()
	return nil, fmt.Errorf("Not implmented for AuthorityListInternal")
}

func (m *AuthorityListInternal) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

/*
type AuthorityListInternal struct {
	msgbase.MessageBase

	Federated []interfaces.IServer
	Audit     []interfaces.IServer
	DBHeight  uint32 // Directory Block Height that owns this ack
}
*/
func (m *AuthorityListInternal) String() string {
	var f_str, a_str string
	for _, f := range m.Federated {
		f_str = f_str + fmt.Sprintf("%x, ", f.GetChainID().Bytes()[3:6])
	}
	for _, a := range m.Audit {
		a_str = a_str + fmt.Sprintf("%x, ", a.GetChainID().Bytes()[3:6])
	}

	return fmt.Sprintf("AuthorityListInternal DBH %d fed [%s] aud[%s]", m.DBHeight, f_str, a_str)
}

func (a *AuthorityListInternal) IsSameAs(b *AuthorityListInternal) bool {
	return true
}
