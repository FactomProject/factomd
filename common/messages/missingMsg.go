// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

//Structure to request missing messages in a node's process list
type MissingMsg struct {
	MessageBase

	Timestamp         interfaces.Timestamp
	DBHeight          uint32
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

func (m *MissingMsg) Int() int {
	return -1
}

func (m *MissingMsg) Bytes() []byte {
	return nil
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

	m.VMIndex, newData = int(newData[0]), newData[1:]
	m.DBHeight, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]

	// Get all the missing messages...
	lenl, newData := binary.BigEndian.Uint32(newData[0:4]), newData[4:]
	for i := 0; i < int(lenl); i++ {
		var height uint32
		height, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]
		m.ProcessListHeight = append(m.ProcessListHeight, height)
	}

	m.Peer2Peer = true // Always a peer2peer request.

	return data, nil
}

func (m *MissingMsg) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *MissingMsg) MarshalBinary() ([]byte, error) {
	var buf primitives.Buffer

	binary.Write(&buf, binary.BigEndian, m.Type())

	t := m.GetTimestamp()
	data, err := t.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	buf.WriteByte(uint8(m.VMIndex))
	binary.Write(&buf, binary.BigEndian, m.DBHeight)

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
		str = fmt.Sprintf("%s%d,", str, n)
	}
	return fmt.Sprintf("MissingMsg --> DBHeight:%3d vm=%3d Hts::%s msgHash[%x]", m.DBHeight, m.VMIndex, str, m.GetMsgHash().Bytes()[:3])
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

func (e *MissingMsg) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

// AddHeight: Add a Missing Message Height to the request
func (e *MissingMsg) AddHeight(h uint32) {
	e.ProcessListHeight = append(e.ProcessListHeight, h)
}

// NewMissingMsg: Build a missing Message request, and add the first Height
func NewMissingMsg(state interfaces.IState, vm int, dbHeight uint32, processlistHeight uint32) *MissingMsg {

	msg := new(MissingMsg)

	msg.Peer2Peer = true // Always a peer2peer request // .
	msg.VMIndex = vm
	msg.Timestamp = state.GetTimestamp()
	msg.DBHeight = dbHeight
	msg.ProcessListHeight = append(msg.ProcessListHeight, processlistHeight)

	return msg
}
