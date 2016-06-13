// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"bytes"
	//"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// Communicate a Directory Block State

type SendRawMsg struct {
	MessageBase
	Timestamp interfaces.Timestamp // Message Timestamp
	Message   string
}

var _ interfaces.IMsg = (*SendRawMsg)(nil)

func (m *SendRawMsg) GetHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *SendRawMsg) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
		data, err := m.MarshalBinary()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

func (m *SendRawMsg) Type() byte {
	//return m.MsgType
	return 0
}

func (m *SendRawMsg) Int() int {
	return -1
}

func (m *SendRawMsg) Bytes() []byte {
	return nil
}

func (m *SendRawMsg) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp
}

func (m *SendRawMsg) Validate(state interfaces.IState) int {
	return 1
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *SendRawMsg) ComputeVMIndex(state interfaces.IState) {

}

// Execute the leader functions of the given message
func (m *SendRawMsg) LeaderExecute(state interfaces.IState) {
	state.LeaderExecute(m)
}

func (m *SendRawMsg) FollowerExecute(state interfaces.IState) {
	state.FollowerExecuteMsg(m)
}

// Acknowledgements do not go into the process list.
func (e *SendRawMsg) Process(dbheight uint32, state interfaces.IState) bool {
	/*switch e.MsgType {
	case 0x14: // AddServer
		return state.ProcessAddServer(dbheight, e)
	case 0x0c:
		return state.ProcessRevealEntry(dbheight, e)
	}*/

	/*
			ProcessAddServer(dbheight uint32, addServerMsg IMsg) bool
		ProcessCommitChain(dbheight uint32, commitChain IMsg) bool
		ProcessCommitEntry(dbheight uint32, commitChain IMsg) bool
		ProcessDBSig(dbheight uint32, commitChain IMsg) bool
		ProcessEOM(dbheight uint32, eom IMsg) bool
		ProcessRevealEntry(dbheight uint32, m IMsg) bool
	*/
	//panic("Should not be processed")
	//return state.ProcessAddServer(dbheight, e)
	return false
}

func (e *SendRawMsg) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *SendRawMsg) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *SendRawMsg) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

/*
func (m *SendRawMsg) Sign(key interfaces.Signer) error {
	return nil
}

func (m *SendRawMsg) GetSignature() interfaces.IFullSignature {
	return nil
}

func (m *SendRawMsg) VerifySignature() (bool, error) {
	return VerifyMessage(m)
}*/

// Not Implemented
func (m *SendRawMsg) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		return
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling Send Raw Message: %v", r)
		}
	}()
	return nil, nil
}

func (m *SendRawMsg) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

/*func (m *SendRawMsg) MarshalForSignature() ([]byte, error) {
	var buf primitives.Buffer

	binary.Write(&buf, binary.BigEndian, m.Type())

	t := m.GetTimestamp()
	data, err := t.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	data, err = m.ServerChainID.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	binary.Write(&buf, binary.BigEndian, uint8(m.ServerType))

	return buf.DeepCopyBytes(), nil
}*/

func (m *SendRawMsg) MarshalBinary() ([]byte, error) {
	var buf primitives.Buffer

	//binary.Write(&buf, binary.BigEndian, m.Type())
	/*
		t := m.GetTimestamp()
		data, err := t.MarshalBinary()
		if err != nil {
			return nil, err
		}
		buf.Write(data)
	*/
	data, err := hex.DecodeString(m.Message)
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	return buf.DeepCopyBytes(), nil
}

func (m *SendRawMsg) String() string {
	return fmt.Sprintf("SendRawMessage: Message: %x Time: %x Msg Hash %x ",
		m.Message[:6],
		m.Timestamp,
		m.GetMsgHash().Bytes()[:3])

}

func (m *SendRawMsg) IsSameAs(b *SendRawMsg) bool {
	if b == nil {
		return false
	}
	if uint64(m.Timestamp) != uint64(b.Timestamp) {
		return false
	}
	if strings.Compare(m.Message, b.Message) != 0 {
		return false
	}
	return true
}

func NewSendRawMsg(state interfaces.IState, message string) interfaces.IMsg {
	msg := new(SendRawMsg)
	msg.Message = message
	//t := uint8(msgType)
	//msg.MsgType = t
	msg.Timestamp = state.GetTimestamp()

	return msg

}
