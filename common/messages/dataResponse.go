// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"

	"github.com/FactomProject/factomd/common/messages/msgbase"
	llog "github.com/FactomProject/factomd/log"
	log "github.com/sirupsen/logrus"
)

// Communicate a Directory Block State

type DataResponse struct {
	msgbase.MessageBase
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

func (m *DataResponse) GetRepeatHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "DataResponse.GetRepeatHash") }()

	return m.GetMsgHash()
}

func (m *DataResponse) GetHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "DataResponse.GetHash") }()

	return m.GetMsgHash()
}

func (m *DataResponse) GetMsgHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "DataResponse.GetMsgHash") }()

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
	return m.Timestamp.Clone()
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

func (m *DataResponse) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
			llog.LogPrintf("recovery", "Error unmarshalling: %v", r)
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

	m.DataType = int(newData[0])
	newData = newData[1:]

	m.DataHash = primitives.NewHash(constants.ZERO_HASH)
	newData, err = m.DataHash.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}
	switch m.DataType {
	case 0:
		entryAttempt, err := attemptEntryUnmarshal(newData)
		if err != nil {
			return nil, err
		} else {
			m.DataObject = entryAttempt
		}
	case 1:
		eblockAttempt, err := attemptEBlockUnmarshal(newData)
		if err != nil {
			return nil, err
		} else {
			m.DataObject = eblockAttempt
		}
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

func attemptEntryUnmarshal(data []byte) (entry interfaces.IEBEntry, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Bytes do not represent an entry %v", r)
			llog.LogPrintf("recovery", "Bytes do not represent an entry %v", r)
		}
	}()

	entry, err = entryBlock.UnmarshalEntry(data)
	if err != nil {
		return nil, err
	}

	return entry, nil
}

func attemptEBlockUnmarshal(data []byte) (eblock interfaces.IEntryBlock, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Bytes do not represent an eblock: %v\n", r)
			llog.LogPrintf("recovery", "Bytes do not represent an eblock: %v", r)
		}
	}()

	eblock, err = entryBlock.UnmarshalEBlock(data)
	if err != nil {
		return nil, err
	}

	return eblock, nil
}

func (m *DataResponse) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "DataResponse.MarshalBinary err:%v", *pe)
		}
	}(&err)
	var buf primitives.Buffer
	buf.Write([]byte{m.Type()})
	if d, err := m.Timestamp.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}

	binary.Write(&buf, binary.BigEndian, uint8(m.DataType))

	if d, err := m.DataHash.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}

	if m.DataObject != nil {
		d, err := m.DataObject.MarshalBinary()
		if err != nil {
			return nil, err
		}
		buf.Write(d)
	}

	return buf.DeepCopyBytes(), nil
}

func (m *DataResponse) String() string {
	return fmt.Sprintf("DataResponse Type: %2d Hash: %x", m.DataType, m.DataHash.Bytes())
}

func (m *DataResponse) LogFields() log.Fields {
	return log.Fields{"category": "message", "messagetype": "dataresponse", "datatype": m.DataType,
		"datahash": m.DataHash.String()}
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
