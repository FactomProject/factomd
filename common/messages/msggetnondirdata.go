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

// MsgGetNonDirData implements the MsgGetNonDirData interface and represents a factom
// get non dir block data message.  It is used to request data such as blocks and transactions
// from another peer.  It should be used in response to the inv (MsgDirInv) message
// to request the actual data referenced by each inventory vector the receiving
// peer doesn't already have.  Each message is limited to a maximum number of
// inventory vectors, which is currently 50,000.  As a result, multiple messages
// must be used to request larger amounts of data.
//
// Use the AddInvVect function to build up the list of inventory vectors when
// sending a getdata message to another peer.
type MsgGetNonDirData struct {
	InvList []*InvVect
}

// AddInvVect adds an inventory vector to the message.
func (msg *MsgGetNonDirData) AddInvVect(iv *InvVect) error {
	if len(msg.InvList)+1 > MaxInvPerMsg {
		str := fmt.Sprintf("too many invvect in message [max %v]",
			MaxInvPerMsg)
		return messageError("MsgGetNonDirData.AddInvVect", str)
	}

	msg.InvList = append(msg.InvList, iv)
	return nil
}

// BtcDecode decodes r using the bitcoin protocol encoding into the receiver.
// This is part of the MsgGetNonDirData interface implementation.
func (msg *MsgGetNonDirData) BtcDecode(r io.Reader, pver uint32) error {
	count, err := readVarInt(r, pver)
	if err != nil {
		return err
	}

	// Limit to max inventory vectors per message.
	if count > MaxInvPerMsg {
		str := fmt.Sprintf("too many invvect in message [%v]", count)
		return messageError("MsgGetNonDirData.BtcDecode", str)
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
// This is part of the MsgGetNonDirData interface implementation.
func (msg *MsgGetNonDirData) BtcEncode(w io.Writer, pver uint32) error {
	// Limit to max inventory vectors per message.
	count := len(msg.InvList)
	if count > MaxInvPerMsg {
		str := fmt.Sprintf("too many invvect in message [%v]", count)
		return messageError("MsgGetNonDirData.BtcEncode", str)
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
// of the MsgGetNonDirData interface implementation.
func (msg *MsgGetNonDirData) Command() string {
	return CmdGetNonDirData
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the MsgGetNonDirData interface implementation.
func (msg *MsgGetNonDirData) MaxPayloadLength(pver uint32) uint32 {
	// Num inventory vectors (varInt) + max allowed inventory vectors.
	return uint32(MaxVarIntPayload + (MaxInvPerMsg * maxInvVectPayload))
}

// NewMsgGetNonDirData returns a new factom get non dir data message that conforms to the
// MsgGetNonDirData interface.  See MsgGetNonDirData for details.
func NewMsgGetNonDirData() *MsgGetNonDirData {
	return &MsgGetNonDirData{
		InvList: make([]*InvVect, 0, defaultInvListAlloc),
	}
}

// NewMsgGetNonDirDataSizeHint returns a new bitcoin getdata message that conforms to
// the MsgGetNonDirData interface.  See MsgGetNonDirData for details.  This function differs
// from NewMsgGetDirData in that it allows a default allocation size for the
// backing array which houses the inventory vector list.  This allows callers
// who know in advance how large the inventory list will grow to avoid the
// overhead of growing the internal backing array several times when appending
// large amounts of inventory vectors with AddInvVect.  Note that the specified
// hint is just that - a hint that is used for the default allocation size.
// Adding more (or less) inventory vectors will still work properly.  The size
// hint is limited to MaxInvPerMsg.
func NewMsgGetNonDirDataSizeHint(sizeHint uint) *MsgGetNonDirData {
	// Limit the specified hint to the maximum allow per message.
	if sizeHint > MaxInvPerMsg {
		sizeHint = MaxInvPerMsg
	}

	return &MsgGetNonDirData{
		InvList: make([]*InvVect, 0, sizeHint),
	}
}

var _ interfaces.IMsg = (*MsgGetNonDirData)(nil)

func (m *MsgGetNonDirData) Process(uint32, interfaces.IState) {}

func (m *MsgGetNonDirData) GetHash() interfaces.IHash {
	return nil
}

func (m *MsgGetNonDirData) GetTimestamp() interfaces.Timestamp {
	return 0
}

func (m *MsgGetNonDirData) Type() int {
	return -1
}

func (m *MsgGetNonDirData) Int() int {
	return -1
}

func (m *MsgGetNonDirData) Bytes() []byte {
	return nil
}

func (m *MsgGetNonDirData) UnmarshalBinaryData(data []byte) (newdata []byte, err error) {
	return nil, nil
}

func (m *MsgGetNonDirData) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *MsgGetNonDirData) MarshalBinary() (data []byte, err error) {
	return nil, nil
}

func (m *MsgGetNonDirData) MarshalForSignature() (data []byte, err error) {
	return nil, nil
}

func (m *MsgGetNonDirData) String() string {
	return ""
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- MsgGetNonDirData is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- MsgGetNonDirData is valid
func (m *MsgGetNonDirData) Validate(dbheight uint32, state interfaces.IState) int {
	return 0
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *MsgGetNonDirData) Leader(state interfaces.IState) bool {
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
func (m *MsgGetNonDirData) LeaderExecute(state interfaces.IState) error {
	return nil
}

// Returns true if this is a message for this server to execute as a follower
func (m *MsgGetNonDirData) Follower(interfaces.IState) bool {
	return true
}

func (m *MsgGetNonDirData) FollowerExecute(interfaces.IState) error {
	return nil
}

func (e *MsgGetNonDirData) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *MsgGetNonDirData) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *MsgGetNonDirData) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}
