// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"encoding/binary"
	"fmt"
	"os"
	"reflect"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"

	"github.com/FactomProject/factomd/common/messages/msgbase"
	log "github.com/sirupsen/logrus"
)

//Structure to request missing messages in a node's process list
type MissingMsg struct {
	msgbase.MessageBase

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

func (m *MissingMsg) GetRepeatHash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("MissingMsg.GetRepeatHash() saw an interface that was nil")
		}
	}()

	return m.GetMsgHash()
}

func (m *MissingMsg) GetHash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("MissingMsg.GetHash() saw an interface that was nil")
		}
	}()

	if m.hash == nil {
		data, err := m.MarshalBinary()
		if err != nil {
			panic(fmt.Sprintf("Error in MissingMsg.GetHash(): %s", err.Error()))
		}
		m.hash = primitives.Sha(data)
	}
	return m.hash
}

func (m *MissingMsg) GetMsgHash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("MissingMsg.GetMsgHash() saw an interface that was nil")
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

func (m *MissingMsg) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp
}

func (m *MissingMsg) Type() byte {
	return constants.MISSING_MSG
}

func (m *MissingMsg) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	newData = data
	if newData[0] != m.Type() {
		return nil, fmt.Errorf("%s", "Invalid Message type")
	}
	newData = newData[1:]

	m.Timestamp = new(primitives.Timestamp)
	newData, err = m.Timestamp.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	m.Asking = new(primitives.Hash)
	newData, err = m.Asking.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	m.VMIndex, newData = int(newData[0]), newData[1:]
	m.DBHeight, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]
	m.SystemHeight, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]

	// Get all the missing messages...
	lenl, newData := binary.BigEndian.Uint32(newData[0:4]), newData[4:]
	for i := 0; i < int(lenl); i++ {
		var height uint32
		height, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]
		m.ProcessListHeight = append(m.ProcessListHeight, height)
	}

	m.Peer2Peer = true // Always a peer2peer request.

	return
}

func (m *MissingMsg) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *MissingMsg) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "MissingMsg.MarshalBinary err:%v", *pe)
		}
	}(&err)
	var buf primitives.Buffer

	binary.Write(&buf, binary.BigEndian, m.Type())

	t := m.GetTimestamp()
	data, err := t.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	if m.Asking == nil {
		m.Asking = primitives.NewHash(constants.ZERO_HASH)
	}
	data, err = m.Asking.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	buf.WriteByte(uint8(m.VMIndex))
	binary.Write(&buf, binary.BigEndian, m.DBHeight)
	binary.Write(&buf, binary.BigEndian, m.SystemHeight)

	binary.Write(&buf, binary.BigEndian, uint32(len(m.ProcessListHeight)))
	for _, h := range m.ProcessListHeight {
		binary.Write(&buf, binary.BigEndian, h)
	}

	bb := buf.DeepCopyBytes()

	return bb, nil
}

func (m *MissingMsg) String() string {
	str := ""
	for _, n := range m.ProcessListHeight {
		str += fmt.Sprintf("%d/%d/%d, ", m.DBHeight, m.VMIndex, n)
	}
	return fmt.Sprintf("MissingMsg --> %x asking for DBh/VMh/h[%s] Sys: %d msgHash[%x] from peer-%d %s",
		m.Asking.Bytes()[3:6],
		str,
		m.SystemHeight,
		m.GetMsgHash().Bytes()[:3], m.GetOrigin(), m.GetNetworkOrigin())
}

func (m *MissingMsg) LogFields() log.Fields {
	return log.Fields{"category": "message", "messagetype": "missingmsg",
		"vm":        m.VMIndex,
		"dbheight":  m.DBHeight,
		"asking":    m.Asking.String(),
		"sysheight": m.SystemHeight,
		"hash":      m.GetMsgHash().String()}
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
	// can't answer about the future
	if m.DBHeight > state.GetLLeaderHeight() {
		return -1
	}
	// can't answer about the past before our earliest pl
	// use int so at height near 0 we can go negative
	if int(m.DBHeight) < int(state.GetLLeaderHeight())-2 {
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
	// search to see if the height is already there.
	for _, ht := range e.ProcessListHeight {
		if ht == h {
			return // if it's already there just return
		}
	}
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
