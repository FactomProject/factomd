// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

//Structure to request missing messages in a node's process list
type MissingData struct {
	MessageBase
	Timestamp interfaces.Timestamp

	RequestHash interfaces.IHash

	//No signature!
}

var _ interfaces.IMsg = (*MissingData)(nil)

func (a *MissingData) IsSameAs(b *MissingData) bool {
	if b == nil {
		return false
	}
	if a.Timestamp.GetTimeMilli() != b.Timestamp.GetTimeMilli() {
		return false
	}

	if a.RequestHash == nil && b.RequestHash != nil {
		return false
	}
	if a.RequestHash != nil {
		if a.RequestHash.IsSameAs(b.RequestHash) == false {
			return false
		}
	}

	return true
}

func (m *MissingData) Process(uint32, interfaces.IState) bool {
	return true
}

func (m *MissingData) GetRepeatHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *MissingData) GetHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *MissingData) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
		data, err := m.MarshalBinary()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

func (m *MissingData) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp
}

func (m *MissingData) Type() byte {
	return constants.MISSING_DATA
}

func (m *MissingData) UnmarshalBinaryData(data []byte) ([]byte, error) {
	buf := primitives.NewBuffer(data)

	t, err := buf.PopByte()
	if err != nil {
		return nil, err
	}
	if t != m.Type() {
		return nil, fmt.Errorf("Invalid Message type")
	}

	m.Timestamp = new(primitives.Timestamp)
	err = buf.PopBinaryMarshallable(m.Timestamp)
	if err != nil {
		return nil, err
	}

	m.RequestHash = primitives.NewZeroHash()
	err = buf.PopBinaryMarshallable(m.RequestHash)
	if err != nil {
		return nil, err
	}

	m.Peer2Peer = true // Always a peer2peer request.

	return buf.DeepCopyBytes(), nil
}

func (m *MissingData) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *MissingData) MarshalBinary() ([]byte, error) {
	buf := primitives.NewBuffer(nil)

	err := buf.PushByte(m.Type())
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(m.Timestamp)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(m.RequestHash)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

func (m *MissingData) String() string {
	return fmt.Sprintf("MissingData: [%x]", m.RequestHash.Bytes()[:5])
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *MissingData) Validate(state interfaces.IState) int {
	return 1
}

func (m *MissingData) ComputeVMIndex(state interfaces.IState) {
}

func (m *MissingData) LeaderExecute(state interfaces.IState) {
	m.FollowerExecute(state)
}

func (m *MissingData) FollowerExecute(state interfaces.IState) {
	var dataObject interfaces.BinaryMarshallable
	//var dataHash interfaces.IHash
	rawObject, dataType, err := state.LoadDataByHash(m.RequestHash)

	if rawObject != nil && err == nil { // If I don't have this message, ignore.
		switch dataType {
		case 0: // DataType = entry
			dataObject = rawObject.(interfaces.IEBEntry)
			//dataHash = dataObject.(interfaces.IEBEntry).GetHash()
		case 1: // DataType = eblock
			dataObject = rawObject.(interfaces.IEntryBlock)
			//dataHash, _ = dataObject.(interfaces.IEntryBlock).Hash()
		default:
			return
		}

		msg := NewDataResponse(state, dataObject, dataType, m.RequestHash)

		msg.SetOrigin(m.GetOrigin())
		msg.SetNetworkOrigin(m.GetNetworkOrigin())
		msg.SendOut(state, msg)
	}
	return
}

func (e *MissingData) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *MissingData) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func NewMissingData(state interfaces.IState, requestHash interfaces.IHash) interfaces.IMsg {
	msg := new(MissingData)

	msg.Peer2Peer = true // Always a peer2peer request.
	msg.Timestamp = state.GetTimestamp()
	msg.RequestHash = requestHash

	return msg
}
