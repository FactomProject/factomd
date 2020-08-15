// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"fmt"

	"github.com/PaulSnow/factom2d/common/constants"
	"github.com/PaulSnow/factom2d/common/interfaces"
	"github.com/PaulSnow/factom2d/common/messages/msgbase"
	"github.com/PaulSnow/factom2d/common/primitives"

	llog "github.com/PaulSnow/factom2d/log"
	log "github.com/sirupsen/logrus"
)

//A placeholder structure for messages
type RequestBlock struct {
	msgbase.MessageBase
	Timestamp interfaces.Timestamp

	//TODO: figure whether this should be signed or not?

	//Not marshalled
	hash interfaces.IHash
}

var _ interfaces.IMsg = (*RequestBlock)(nil)

func (a *RequestBlock) IsSameAs(b *RequestBlock) bool {
	if b == nil {
		return false
	}
	if a.Timestamp.GetTimeMilli() != b.Timestamp.GetTimeMilli() {
		return false
	}

	//TODO: expand

	return true
}

func (m *RequestBlock) Process(uint32, interfaces.IState) bool { return true }

func (m *RequestBlock) GetRepeatHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "RequestBlock.GetRepeatHash") }()

	return m.GetMsgHash()
}

func (m *RequestBlock) GetHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "RequestBlock.GetHash") }()

	if m.hash == nil {
		data, err := m.MarshalForSignature()
		if err != nil {
			panic(fmt.Sprintf("Error in RequestBlock.GetHash(): %s", err.Error()))
		}
		m.hash = primitives.Sha(data)
	}
	return m.hash
}

func (m *RequestBlock) GetMsgHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "RequestBlock.GetMsgHash") }()

	if m.MsgHash == nil {
		data, err := m.MarshalBinary()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

func (m *RequestBlock) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp
}

func (m *RequestBlock) Type() byte {
	return constants.REQUEST_BLOCK_MSG
}

func (m *RequestBlock) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling RequestBlock: %v", r)
			llog.LogPrintf("recovery", "Error unmarshalling RequestBlock: %v", r)
		}
	}()
	newData = data
	if newData[0] != m.Type() {
		return nil, fmt.Errorf("Invalid Message type")
	}
	newData = newData[1:]

	m.Timestamp = new(primitives.Timestamp)
	newData, err = m.Timestamp.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	//TODO: expand

	return newData, nil
}

func (m *RequestBlock) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *RequestBlock) MarshalForSignature() (data []byte, err error) {
	var buf primitives.Buffer
	buf.Write([]byte{m.Type()})
	if d, err := m.Timestamp.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}

	//TODO: expand

	return buf.DeepCopyBytes(), nil
}

func (m *RequestBlock) MarshalBinary() (data []byte, err error) {
	//TODO: sign or delete
	return m.MarshalForSignature()
}

func (m *RequestBlock) String() string {
	return "Request Block"
}

func (m *RequestBlock) LogFields() log.Fields {
	return log.Fields{}
}

func (m *RequestBlock) DBHeight() int {
	return 0
}

func (m *RequestBlock) ChainID() []byte {
	return nil
}

func (m *RequestBlock) ListHeight() int {
	return 0
}

func (m *RequestBlock) SerialHash() []byte {
	return nil
}

func (m *RequestBlock) Signature() []byte {
	return nil
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *RequestBlock) Validate(state interfaces.IState) int {
	return 0
}

func (m *RequestBlock) ComputeVMIndex(state interfaces.IState) {
}

func (m *RequestBlock) LeaderExecute(state interfaces.IState) {
}

func (m *RequestBlock) FollowerExecute(interfaces.IState) {
}

func (e *RequestBlock) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *RequestBlock) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}
