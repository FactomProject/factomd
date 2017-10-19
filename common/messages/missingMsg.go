// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"

	log "github.com/FactomProject/logrus"
)

//Structure to request missing messages in a node's process list
type MissingMsg struct {
	MessageBase

	Timestamp         interfaces.Timestamp
	Asking            interfaces.IHash
	DBHeight          uint32
	SystemHeight      uint32 // Might as well check for a missing Server Fault
	ProcessListHeight []uint32

	//No signature!

	//Not marshalled
	hash interfaces.IHash
}

var _ interfaces.IMsg = (*MissingMsg)(nil)

func (a *MissingMsg) IsSameAs(b *MissingMsg) bool {
	if b == nil {
		return false
	}
	if a.Timestamp.GetTimeMilli() != b.Timestamp.GetTimeMilli() {
		return false
	}

	if a.DBHeight != b.DBHeight {
		return false
	}

	if a.VMIndex != b.VMIndex {
		return false
	}

	if len(a.ProcessListHeight) != len(b.ProcessListHeight) {
		return false
	}
	for i, v := range a.ProcessListHeight {
		if v != b.ProcessListHeight[i] {
			return false
		}
	}

	return true
}

func (m *MissingMsg) Process(uint32, interfaces.IState) bool {
	panic("MissingMsg should not have its Process() method called")
}

func (m *MissingMsg) GetRepeatHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *MissingMsg) GetHash() interfaces.IHash {
	if m.hash == nil {
		data, err := m.MarshalBinary()
		if err != nil {
			panic(fmt.Sprintf("Error in MissingMsg.GetHash(): %s", err.Error()))
		}
		m.hash = primitives.Sha(data)
	}
	return m.hash
}

func (m *MissingMsg) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
		data, err := m.MarshalBinary()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

func (m *MissingMsg) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp
}

func (m *MissingMsg) Type() byte {
	return constants.MISSING_MSG
}

func (m *MissingMsg) UnmarshalBinaryData(data []byte) ([]byte, error) {
	buf := primitives.NewBuffer(data)

	t, err := buf.PopByte()
	if err != nil {
		return nil, err
	}
	if t != m.Type() {
		return nil, fmt.Errorf("%s", "Invalid Message type")
	}

	m.Timestamp = new(primitives.Timestamp)
	err = buf.PopBinaryMarshallable(m.Timestamp)
	if err != nil {
		return nil, err
	}

	m.Asking = new(primitives.Hash)
	err = buf.PopBinaryMarshallable(m.Asking)
	if err != nil {
		return nil, err
	}

	t, err = buf.PopByte()
	if err != nil {
		return nil, err
	}
	m.VMIndex = int(t)

	m.DBHeight, err = buf.PopUInt32()
	if err != nil {
		return nil, err
	}
	m.SystemHeight, err = buf.PopUInt32()
	if err != nil {
		return nil, err
	}

	// Get all the missing messages...
	lenl, err := buf.PopUInt32()
	if err != nil {
		return nil, err
	}
	for i := 0; i < int(lenl); i++ {
		height, err := buf.PopUInt32()
		if err != nil {
			return nil, err
		}
		m.ProcessListHeight = append(m.ProcessListHeight, height)
	}

	m.Peer2Peer = true // Always a peer2peer request.

	return buf.DeepCopyBytes(), nil
}

func (m *MissingMsg) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *MissingMsg) MarshalBinary() ([]byte, error) {
	buf := primitives.NewBuffer(nil)

	err := buf.PushByte(m.Type())
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(m.GetTimestamp())
	if err != nil {
		return nil, err
	}

	if m.Asking == nil {
		m.Asking = primitives.NewZeroHash()
	}
	err = buf.PushBinaryMarshallable(m.Asking)
	if err != nil {
		return nil, err
	}
	err = buf.PushByte(byte(m.VMIndex))
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(m.DBHeight)
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(m.SystemHeight)
	if err != nil {
		return nil, err
	}

	err = buf.PushUInt32(uint32(len(m.ProcessListHeight)))
	if err != nil {
		return nil, err
	}
	for _, h := range m.ProcessListHeight {
		err = buf.PushUInt32(h)
		if err != nil {
			return nil, err
		}
	}

	return buf.DeepCopyBytes(), nil
}

func (m *MissingMsg) String() string {
	str := ""
	for _, n := range m.ProcessListHeight {
		str = fmt.Sprintf("%s%d,", str, n)
	}
	return fmt.Sprintf("MissingMsg --> Asking %x DBHeight:%3d vm=%3d Hts::[%s] Sys: %d msgHash[%x]",
		m.Asking.Bytes()[:8],
		m.DBHeight,
		m.VMIndex,
		str,
		m.SystemHeight,
		m.GetMsgHash().Bytes()[:3])
}

func (m *MissingMsg) LogFields() log.Fields {
	return log.Fields{"category": "message", "messagetype": "missingmsg",
		"vm":        m.VMIndex,
		"dbheight":  m.DBHeight,
		"asking":    m.Asking.String()[:8],
		"sysheight": m.SystemHeight,
		"hash":      m.GetMsgHash().String()[:6]}
}

func (m *MissingMsg) ChainID() []byte {
	return nil
}

func (m *MissingMsg) ListHeight() int {
	return 0
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *MissingMsg) Validate(state interfaces.IState) int {
	if m.Asking == nil {
		return -1
	}
	if m.Asking.IsZero() {
		return -1
	}
	return 1
}

func (m *MissingMsg) ComputeVMIndex(state interfaces.IState) {
}

func (m *MissingMsg) LeaderExecute(state interfaces.IState) {
	m.FollowerExecute(state)
}

func (m *MissingMsg) FollowerExecute(state interfaces.IState) {
	state.FollowerExecuteMissingMsg(m)
}

func (e *MissingMsg) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *MissingMsg) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

// AddHeight: Add a Missing Message Height to the request
func (e *MissingMsg) AddHeight(h uint32) {
	e.ProcessListHeight = append(e.ProcessListHeight, h)
}

// NewMissingMsg: Build a missing Message request, and add the first Height
func NewMissingMsg(state interfaces.IState, vm int, dbHeight uint32, processlistHeight uint32) *MissingMsg {
	msg := new(MissingMsg)

	msg.Asking = state.GetIdentityChainID()
	msg.Peer2Peer = true // Always a peer2peer request // .
	msg.VMIndex = vm
	msg.Timestamp = state.GetTimestamp()
	msg.DBHeight = dbHeight
	msg.ProcessListHeight = append(msg.ProcessListHeight, processlistHeight)
	msg.SystemHeight = uint32(state.GetSystemHeight(dbHeight))
	return msg
}
