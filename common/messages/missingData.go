// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"fmt"
	"os"
	"reflect"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"

	"github.com/FactomProject/factomd/common/messages/msgbase"

	llog "github.com/FactomProject/factomd/log"
	log "github.com/sirupsen/logrus"
)

//Structure to request missing messages in a node's process list
type MissingData struct {
	msgbase.MessageBase
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

func (m *MissingData) GetRepeatHash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("MissingData.GetRepeatHash() saw an interface that was nil")
		}
	}()

	return m.GetMsgHash()
}

func (m *MissingData) GetHash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("MissingData.GetHash() saw an interface that was nil")
		}
	}()

	return m.GetMsgHash()
}

func (m *MissingData) GetMsgHash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("MissingData.GetMsgHash() saw an interface that was nil")
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

func (m *MissingData) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp.Clone()
}

func (m *MissingData) Type() byte {
	return constants.MISSING_DATA
}

func (m *MissingData) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
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

	m.RequestHash = primitives.NewHash(constants.ZERO_HASH)
	newData, err = m.RequestHash.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	m.Peer2Peer = true // Always a peer2peer request.

	return
}

func (m *MissingData) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *MissingData) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "MissingData.MarshalBinary err:%v", *pe)
		}
	}(&err)
	var buf primitives.Buffer
	buf.Write([]byte{m.Type()})
	if d, err := m.Timestamp.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}

	if d, err := m.RequestHash.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}

	return buf.DeepCopyBytes(), nil
}

func (m *MissingData) String() string {
	return fmt.Sprintf("MissingData: [%x]", m.RequestHash.Bytes()[:5])
}

func (m *MissingData) LogFields() log.Fields {
	return log.Fields{"category": "message", "messagetype": "missingdata",
		"hash": m.GetHash().String(), "requesthash": m.RequestHash.String()}
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
	panic("deprecated") // go routine in NetworkProcessorNet now handles this
}

func (m *MissingData) SendResponse(state interfaces.IState) {
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
