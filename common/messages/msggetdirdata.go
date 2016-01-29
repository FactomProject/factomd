// Copyright (c) 2013-2015 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package messages

import (
	"bytes"
	"fmt"
	"io"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// MsgGetDirData implements the MsgGetDirData interface and represents a factom
// getdirdata message.  It is used to request data such as blocks and transactions
// from another peer.  It should be used in response to the inv (MsgDirInv) message
// to request the actual data referenced by each inventory vector the receiving
// peer doesn't already have.  Each message is limited to a maximum number of
// inventory vectors, which is currently 50,000.  As a result, multiple messages
// must be used to request larger amounts of data.
//
// Use the AddInvVect function to build up the list of inventory vectors when
// sending a getdata message to another peer.
type MsgGetDirData struct {
	InvList []*InvVect
}

// AddInvVect adds an inventory vector to the message.
func (msg *MsgGetDirData) AddInvVect(iv *InvVect) error {
	if len(msg.InvList)+1 > MaxInvPerMsg {
		str := fmt.Sprintf("too many invvect in message [max %v]",
			MaxInvPerMsg)
		return messageError("MsgGetDirData.AddInvVect", str)
	}

	msg.InvList = append(msg.InvList, iv)
	return nil
}

// BtcDecode decodes r using the bitcoin protocol encoding into the receiver.
// This is part of the MsgGetDirData interface implementation.
func (msg *MsgGetDirData) BtcDecode(r io.Reader, pver uint32) error {
	count, err := readVarInt(r, pver)
	if err != nil {
		return err
	}

	// Limit to max inventory vectors per message.
	if count > MaxInvPerMsg {
		str := fmt.Sprintf("too many invvect in message [%v]", count)
		return messageError("MsgGetDirData.BtcDecode", str)
	}

	msg.InvList = make([]*InvVect, 0, count)
	for i := uint64(0); i < count; i++ {
		iv := InvVect{}
		err := readInvVect(r, pver, &iv)
		if err != nil {
			return err
		}
		msg.AddInvVect(&iv)
	}

	return nil
}

// BtcEncode encodes the receiver to w using the bitcoin protocol encoding.
// This is part of the MsgGetDirData interface implementation.
func (msg *MsgGetDirData) BtcEncode(w io.Writer, pver uint32) error {
	// Limit to max inventory vectors per message.
	count := len(msg.InvList)
	if count > MaxInvPerMsg {
		str := fmt.Sprintf("too many invvect in message [%v]", count)
		return messageError("MsgGetDirData.BtcEncode", str)
	}

	err := writeVarInt(w, pver, uint64(count))
	if err != nil {
		return err
	}

	for _, iv := range msg.InvList {
		err := writeInvVect(w, pver, iv)
		if err != nil {
			return err
		}
	}

	return nil
}

// Command returns the protocol command string for the message.  This is part
// of the MsgGetDirData interface implementation.
func (msg *MsgGetDirData) Command() string {
	return CmdGetDirData
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the MsgGetDirData interface implementation.
func (msg *MsgGetDirData) MaxPayloadLength(pver uint32) uint32 {
	// Num inventory vectors (varInt) + max allowed inventory vectors.
	return uint32(MaxVarIntPayload + (MaxInvPerMsg * maxInvVectPayload))
}

// NewMsgGetDirData returns a new bitcoin getdata message that conforms to the
// MsgGetDirData interface.  See MsgGetDirData for details.
func NewMsgGetDirData() *MsgGetDirData {
	return &MsgGetDirData{
		InvList: make([]*InvVect, 0, defaultInvListAlloc),
	}
}

// NewMsgGetDirDataSizeHint returns a new bitcoin getdata message that conforms to
// the MsgGetDirData interface.  See MsgGetDirData for details.  This function differs
// from NewMsgGetDirData in that it allows a default allocation size for the
// backing array which houses the inventory vector list.  This allows callers
// who know in advance how large the inventory list will grow to avoid the
// overhead of growing the internal backing array several times when appending
// large amounts of inventory vectors with AddInvVect.  Note that the specified
// hint is just that - a hint that is used for the default allocation size.
// Adding more (or less) inventory vectors will still work properly.  The size
// hint is limited to MaxInvPerMsg.
func NewMsgGetDirDataSizeHint(sizeHint uint) *MsgGetDirData {
	// Limit the specified hint to the maximum allow per message.
	if sizeHint > MaxInvPerMsg {
		sizeHint = MaxInvPerMsg
	}

	return &MsgGetDirData{
		InvList: make([]*InvVect, 0, sizeHint),
	}
}

var _ interfaces.IMsg = (*MsgGetDirData)(nil)

func (m *MsgGetDirData) Process(uint32, interfaces.IState) {}

func (m *MsgGetDirData) GetHash() interfaces.IHash {
	return nil
}

func (m *MsgGetDirData) GetTimestamp() interfaces.Timestamp {
	return 0
}

func (m *MsgGetDirData) Type() int {
	return -1
}

func (m *MsgGetDirData) Int() int {
	return -1
}

func (m *MsgGetDirData) Bytes() []byte {
	return nil
}

func (m *MsgGetDirData) UnmarshalBinaryData(data []byte) (newdata []byte, err error) {
	return nil, nil
}

func (m *MsgGetDirData) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *MsgGetDirData) MarshalBinary() (data []byte, err error) {
	return nil, nil
}

func (m *MsgGetDirData) MarshalForSignature() (data []byte, err error) {
	return nil, nil
}

func (m *MsgGetDirData) String() string {
	return ""
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- MsgGetDirData is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- MsgGetDirData is valid
func (m *MsgGetDirData) Validate(dbheight uint32, state interfaces.IState) int {
	return 0
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *MsgGetDirData) Leader(state interfaces.IState) bool {
	switch state.GetNetworkNumber() {
	case 0: // Main Network
		panic("Not implemented yet")
	case 1: // Test Network
		panic("Not implemented yet")
	case 2: // Local Network
		panic("Not implemented yet")
	default:
		panic("Not implemented yet")
	}

}

// Execute the leader functions of the given message
func (m *MsgGetDirData) LeaderExecute(state interfaces.IState) error {
	return nil
}

// Returns true if this is a message for this server to execute as a follower
func (m *MsgGetDirData) Follower(interfaces.IState) bool {
	return true
}

func (m *MsgGetDirData) FollowerExecute(interfaces.IState) error {
	return nil
}

func (e *MsgGetDirData) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *MsgGetDirData) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *MsgGetDirData) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}
