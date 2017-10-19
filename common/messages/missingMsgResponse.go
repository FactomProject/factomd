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
type MissingMsgResponse struct {
	MessageBase

	Timestamp   interfaces.Timestamp
	AckResponse interfaces.IMsg
	MsgResponse interfaces.IMsg

	//No signature!

	//Not marshalled
	hash interfaces.IHash
}

var _ interfaces.IMsg = (*MissingMsgResponse)(nil)

func (a *MissingMsgResponse) IsSameAs(b *MissingMsgResponse) bool {
	if b == nil {
		return false
	}
	if a.Timestamp.GetTimeMilli() != b.Timestamp.GetTimeMilli() {
		return false
	}

	if !a.MsgResponse.GetHash().IsSameAs(b.MsgResponse.GetHash()) {
		fmt.Println("MissingMsgResponse IsNotSameAs because MsgResp GetHash mismatch")
		return false
	}

	if !a.AckResponse.GetHash().IsSameAs(b.AckResponse.GetHash()) {
		fmt.Println("MissingMsgResponse IsNotSameAs because Ack GetHash mismatch")
		return false
	}

	return true
}

func (m *MissingMsgResponse) Process(uint32, interfaces.IState) bool {
	return true
}

func (m *MissingMsgResponse) GetRepeatHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *MissingMsgResponse) GetHash() interfaces.IHash {
	if m.hash == nil {
		data, err := m.MarshalBinary()
		if err != nil {
			panic(fmt.Sprintf("Error in MissingMsg.GetHash(): %s", err.Error()))
		}
		m.hash = primitives.Sha(data)
	}
	return m.hash
}

func (m *MissingMsgResponse) GetMsgHash() interfaces.IHash {
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
	return m.Timestamp
}

func (m *MissingMsgResponse) Type() byte {
	return constants.MISSING_MSG_RESPONSE
}

func (m *MissingMsgResponse) UnmarshalBinaryData(data []byte) ([]byte, error) {
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

	t, err = buf.PopByte()
	if err != nil {
		return nil, err
	}

	if t == 1 {
		m.AckResponse = new(Ack)
		err = buf.PopBinaryMarshallable(m.AckResponse)
		if err != nil {
			return nil, err
		}
	}

	rest, mr, err := UnmarshalMessageData(buf.DeepCopyBytes())
	if err != nil {
		return nil, err
	}
	m.MsgResponse = mr

	m.Peer2Peer = true // Always a peer2peer request.

	return rest, nil
}

func (m *MissingMsgResponse) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *MissingMsgResponse) MarshalBinary() ([]byte, error) {
	buf := primitives.NewBuffer(nil)

	err := buf.PushByte(m.Type())
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(m.GetTimestamp())
	if err != nil {
		return nil, err
	}

	if m.AckResponse == nil {
		err = buf.PushByte(0)
		if err != nil {
			return nil, err
		}
	} else {
		err = buf.PushByte(1)
		if err != nil {
			return nil, err
		}
		err = buf.PushBinaryMarshallable(m.AckResponse)
		if err != nil {
			return nil, err
		}
	}

	err = buf.PushBinaryMarshallable(m.MsgResponse)
	if err != nil {
		return nil, err
	}

	var mmm MissingMsgResponse

	bb := buf.DeepCopyBytes()

	//TODO: delete this once we have unit tests
	if unmarshalErr := mmm.UnmarshalBinary(bb); unmarshalErr != nil {
		fmt.Println("MissingMsgResponse failed to marshal/unmarshal: ", unmarshalErr)
		return nil, unmarshalErr
	}

	return bb, nil
}

func (m *MissingMsgResponse) String() string {
	ack, ok := m.AckResponse.(*Ack)
	if !ok {
		return fmt.Sprint("MissingMsgResponse (no Ack) <-- ", m.MsgResponse.String())
	}
	return fmt.Sprintf("MissingMsgResponse <-- DBHeight:%3d vm=%3d PL Height:%3d msgHash[%x]", ack.DBHeight, ack.VMIndex, ack.Height, m.GetMsgHash().Bytes()[:3])
}

func (m *MissingMsgResponse) LogFields() log.Fields {
	return log.Fields{"category": "message", "messagetype": "missingmsgresponse",
		"ackhash": m.Ack.GetMsgHash().String()[:10],
		"msghash": m.MsgResponse.GetMsgHash().String()[:10]}
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
