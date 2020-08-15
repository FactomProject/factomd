// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/PaulSnow/factom2d/common/constants"
	"github.com/PaulSnow/factom2d/common/interfaces"
	"github.com/PaulSnow/factom2d/common/messages/msgbase"
	"github.com/PaulSnow/factom2d/common/primitives"

	llog "github.com/PaulSnow/factom2d/log"
	log "github.com/sirupsen/logrus"
)

//Structure to request missing messages in a node's process list
type MissingMsgResponse struct {
	msgbase.MessageBase

	Timestamp   interfaces.Timestamp
	AckResponse interfaces.IMsg
	MsgResponse interfaces.IMsg

	//No signature!

	//Not marshalled
	hash interfaces.IHash
}

var General interfaces.IGeneralMsg

var _ interfaces.IMsg = (*MissingMsgResponse)(nil)

func (a *MissingMsgResponse) IsSameAs(b *MissingMsgResponse) bool {
	if b == nil {
		return false
	}
	if a.Timestamp.GetTimeMilli() != b.Timestamp.GetTimeMilli() {
		return false
	}

	ah := a.MsgResponse.GetHash()
	bh := b.MsgResponse.GetHash()

	if !ah.IsSameAs(bh) {
		return false
	}

	if !a.AckResponse.GetHash().IsSameAs(b.AckResponse.GetHash()) {
		return false
	}

	return true
}

func (m *MissingMsgResponse) Process(uint32, interfaces.IState) bool {
	return true
}

func (m *MissingMsgResponse) GetRepeatHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "MissingMsgResponse.GetRepeatHash") }()

	return m.GetMsgHash()
}

func (m *MissingMsgResponse) GetHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "MissingMsgResponse.GetHash") }()

	if m.hash == nil {
		data, err := m.MarshalBinary()
		if err != nil {
			panic(fmt.Sprintf("Error in MissingMsg.GetHash(): %s", err.Error()))
		}
		m.hash = primitives.Sha(data)
	}
	return m.hash
}

func (m *MissingMsgResponse) GetMsgHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "MissingMsgResponse.GetMsgHash") }()

	if m.MsgHash == nil {
		data, err := m.MarshalBinary()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

func (m *MissingMsgResponse) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp.Clone()
}

func (m *MissingMsgResponse) Type() byte {
	return constants.MISSING_MSG_RESPONSE
}

func (m *MissingMsgResponse) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
			llog.LogPrintf("recovery", "Error unmarshalling: %v", r)
		}
	}()

	buf := primitives.NewBuffer(data)

	b, err := buf.PopByte()
	if err != nil {
		return nil, err
	}
	if b != m.Type() {
		return nil, fmt.Errorf("%s", "Invalid Message type")
	}
	m.Timestamp, err = buf.PopTimestamp()

	b, err = buf.PopByte()
	if err != nil {
		return nil, err
	}

	if b == 1 {
		m.AckResponse, err = buf.PopMsg()
		if err != nil {
			return nil, err
		}
	}
	m.MsgResponse, err = buf.PopMsg()
	if err != nil {
		return nil, err
	}
	m.Peer2Peer = true // Always a peer2peer request.

	return
}

func (m *MissingMsgResponse) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *MissingMsgResponse) MarshalBinaryOld() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "MissingMsgResponse.MarshalBinaryOld err:%v", *pe)
		}
	}(&err)
	var buf primitives.Buffer

	binary.Write(&buf, binary.BigEndian, m.Type())

	data, err := m.GetTimestamp().MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	if m.AckResponse == nil {
		buf.WriteByte(0)
	} else {
		buf.WriteByte(1)

		ackData, err := m.AckResponse.MarshalBinary()
		if err != nil {
			return nil, err
		}
		buf.Write(ackData)
	}

	msgData, err := m.MsgResponse.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(msgData)

	bb := buf.DeepCopyBytes()

	return bb, nil
}

func (m *MissingMsgResponse) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "MissingMsgResponse.MarshalBinary err:%v", *pe)
		}
	}(&err)

	var buf primitives.Buffer

	buf.PushByte(m.Type())
	buf.PushTimestamp(m.GetTimestamp())

	if m.AckResponse == nil {
		buf.PushByte(0)
	} else {
		buf.PushByte(1)
		buf.PushMsg(m.AckResponse)
	}
	buf.PushMsg(m.MsgResponse)

	bb := buf.DeepCopyBytes()

	return bb, nil
}

func (m *MissingMsgResponse) String() string {
	ack, ok := m.AckResponse.(*Ack)
	if !ok {
		return fmt.Sprint("MissingMsgResponse (no Ack) <-- ", m.MsgResponse.String())
	}

	return fmt.Sprintf("MissingMsgResponse <-- DBh/VMh/h[%15s] message %s msgHash[%x] to peer-%d %s",
		fmt.Sprintf("%d/%d/%d", ack.DBHeight, ack.VMIndex, ack.Height), m.MsgResponse.String(),
		m.GetMsgHash().Bytes()[:3], m.GetOrigin(), m.GetNetworkOrigin())
}

func (m *MissingMsgResponse) LogFields() log.Fields {
	if m == nil {
		return log.Fields{"category": "message", "messagetype": "missingmsgresponse",
			"ackhash": "nil",
			"msghash": "nil"}
	} else {
		var ahash, mshash string
		if m.AckResponse != nil {
			ahash = m.AckResponse.GetMsgHash().String()
		} else {
			ahash = "nil"
		}
		if m.MsgResponse != nil {
			mshash = m.MsgResponse.GetMsgHash().String()
		} else {
			mshash = "nil"
		}
		return log.Fields{"category": "message", "messagetype": "missingmsgresponse",
			"ackhash": ahash,
			"msghash": mshash}
	}
}

func (m *MissingMsgResponse) ChainID() []byte {
	return nil
}

func (m *MissingMsgResponse) ListHeight() int {
	return 0
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *MissingMsgResponse) Validate(state interfaces.IState) int {
	if m.AckResponse == nil {
		return -1
	}
	if m.MsgResponse == nil {
		return -1
	}
	ack, ok := m.AckResponse.(*Ack)
	if !ok {
		return -1
	}
	if ack.DBHeight < state.GetLLeaderHeight() || ack.DBHeight == state.GetLLeaderHeight() && int(ack.Minute) < int(state.GetCurrentMinute()) {
		return -1
	}
	return 1
}

func (m *MissingMsgResponse) ComputeVMIndex(state interfaces.IState) {
}

func (m *MissingMsgResponse) LeaderExecute(state interfaces.IState) {
	m.FollowerExecute(state)
}

func (m *MissingMsgResponse) FollowerExecute(state interfaces.IState) {
	state.FollowerExecuteMMR(m)

	return
}

func (e *MissingMsgResponse) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *MissingMsgResponse) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func NewMissingMsgResponse(state interfaces.IState, msgResponse interfaces.IMsg, ackResponse interfaces.IMsg) interfaces.IMsg {
	msg := new(MissingMsgResponse)

	msg.Peer2Peer = true // Always a peer2peer request.
	msg.Timestamp = state.GetTimestamp()
	msg.MsgResponse = msgResponse
	msg.AckResponse = ackResponse

	return msg
}
