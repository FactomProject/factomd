// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"

	log "github.com/FactomProject/logrus"
)

// Communicate a Directory Block State

type DataResponse struct {
	MessageBase
	Timestamp interfaces.Timestamp

	DataType   int // 0 = Entry, 1 = EntryBlock
	DataHash   interfaces.IHash
	DataObject interfaces.BinaryMarshallable //Entry or EntryBlock

	//Not signed!
}

var _ interfaces.IMsg = (*DataResponse)(nil)

func (a *DataResponse) IsSameAs(b *DataResponse) bool {
	if b == nil {
		return false
	}
	if a.Timestamp.GetTimeMilli() != b.Timestamp.GetTimeMilli() {
		return false
	}
	if a.DataType != b.DataType {
		return false
	}

	if a.DataHash == nil && b.DataHash != nil {
		return false
	}
	if a.DataHash != nil {
		if a.DataHash.IsSameAs(b.DataHash) == false {
			return false
		}
	}

	if a.DataObject == nil && b.DataObject != nil {
		return false
	}
	if a.DataObject != nil {
		hex1, err := a.DataObject.MarshalBinary()
		if err != nil {
			return false
		}
		hex2, err := b.DataObject.MarshalBinary()
		if err != nil {
			return false
		}
		if primitives.AreBytesEqual(hex1, hex2) == false {
			return false
		}
	}
	return true
}

func (m *DataResponse) GetRepeatHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *DataResponse) GetHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *DataResponse) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
		data, err := m.MarshalBinary()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

func (m *DataResponse) Type() byte {
	return constants.DATA_RESPONSE
}

func (m *DataResponse) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *DataResponse) Validate(state interfaces.IState) int {
	var dataHash interfaces.IHash
	var err error
	switch m.DataType {
	case 0: // DataType = entry
		dataObject, ok := m.DataObject.(interfaces.IEBEntry)
		if !ok {
			return -1
		}
		dataHash = dataObject.GetHash()
	case 1: // DataType = eblock
		dataObject, ok := m.DataObject.(interfaces.IEntryBlock)
		if !ok {
			return -1
		}
		dataHash, err = dataObject.KeyMR()
		if err != nil {
			return -1
		}
	default:
		// DataType currently not supported, treat as invalid
		return -1
	}

	if dataHash.IsSameAs(m.DataHash) {
		return 1
	}

	return -1
}

func (m *DataResponse) ComputeVMIndex(state interfaces.IState) {}

// Execute the leader functions of the given message
func (m *DataResponse) LeaderExecute(state interfaces.IState) {
	m.FollowerExecute(state)
}

func (m *DataResponse) FollowerExecute(state interfaces.IState) {
	state.FollowerExecuteDataResponse(m)
}

// Acknowledgements do not go into the process list.
func (e *DataResponse) Process(dbheight uint32, state interfaces.IState) bool {
	panic("Should never have its Process() method called")
}

func (e *DataResponse) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *DataResponse) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (m *DataResponse) UnmarshalBinaryData(data []byte) ([]byte, error) {
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

	t, err = buf.PopByte()
	if err != nil {
		return nil, err
	}
	m.DataType = int(t)

	m.DataHash = primitives.NewZeroHash()
	err = buf.PopBinaryMarshallable(m.DataHash)
	if err != nil {
		return nil, err
	}
	switch m.DataType {
	case 0:
		ebe := entryBlock.NewEntry()
		err = buf.PopBinaryMarshallable(ebe)
		if err != nil {
			return nil, err
		}
		m.DataObject = ebe
	case 1:
		eb := entryBlock.NewEBlock()
		err = buf.PopBinaryMarshallable(eb)
		if err != nil {
			return nil, err
		}
		m.DataObject = eb
	default:
		return nil, fmt.Errorf("DataResponse's DataType not supported for unmarshalling yet")
	}

	m.Peer2Peer = true // Always a peer2peer request.
	return data, nil
}

func (m *DataResponse) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *DataResponse) MarshalBinary() ([]byte, error) {
	buf := primitives.NewBuffer(nil)

	err := buf.PushByte(m.Type())
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(m.Timestamp)
	if err != nil {
		return nil, err
	}
	err = buf.PushByte(byte(m.DataType))
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(m.DataHash)
	if err != nil {
		return nil, err
	}

	if m.DataObject != nil {
		err = buf.PushBinaryMarshallable(m.DataObject)
		if err != nil {
			return nil, err
		}
	}

	return buf.DeepCopyBytes(), nil
}

func (m *DataResponse) String() string {
	return fmt.Sprintf("DataResponse Type: %2d Hash: %x\n",
		m.DataType,
		m.DataHash.Bytes())
}

func (m *DataResponse) LogFields() log.Fields {
	return log.Fields{"category": "message", "messagetype": "dataresponse", "datatype": m.DataType,
		"datahash": m.DataHash.String()[:6]}
}

func NewDataResponse(state interfaces.IState, dataObject interfaces.BinaryMarshallable,
	dataType int,
	dataHash interfaces.IHash) interfaces.IMsg {
	msg := new(DataResponse)

	msg.Peer2Peer = true
	msg.Timestamp = state.GetTimestamp()

	msg.DataHash = dataHash
	msg.DataType = dataType
	msg.DataObject = dataObject

	//fmt.Println("DATARESPONSE: ", msg.DataObject)

	return msg
}
