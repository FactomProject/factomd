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

// MsgGetData implements the MsgGetData interface and represents a bitcoin
// getdata message.  It is used to request data such as blocks and transactions
// from another peer.  It should be used in response to the inv (MsgInv) message
// to request the actual data referenced by each inventory vector the receiving
// peer doesn't already have.  Each message is limited to a maximum number of
// inventory vectors, which is currently 50,000.  As a result, multiple messages
// must be used to request larger amounts of data.
//
// Use the AddInvVect function to build up the list of inventory vectors when
// sending a getdata message to another peer.
type MsgGetData struct {
	InvList []*InvVect
}

// AddInvVect adds an inventory vector to the message.
func (msg *MsgGetData) AddInvVect(iv *InvVect) error {
	if len(msg.InvList)+1 > MaxInvPerMsg {
		str := fmt.Sprintf("too many invvect in message [max %v]",
			MaxInvPerMsg)
		return messageError("MsgGetData.AddInvVect", str)
	}

	msg.InvList = append(msg.InvList, iv)
	return nil
}

// BtcDecode decodes r using the bitcoin protocol encoding into the receiver.
// This is part of the MsgGetData interface implementation.
func (msg *MsgGetData) BtcDecode(r io.Reader, pver uint32) error {
	count, err := readVarInt(r, pver)
	if err != nil {
		return err
	}

	// Limit to max inventory vectors per message.
	if count > MaxInvPerMsg {
		str := fmt.Sprintf("too many invvect in message [%v]", count)
		return messageError("MsgGetData.BtcDecode", str)
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
// This is part of the MsgGetData interface implementation.
func (msg *MsgGetData) BtcEncode(w io.Writer, pver uint32) error {
	// Limit to max inventory vectors per message.
	count := len(msg.InvList)
	if count > MaxInvPerMsg {
		str := fmt.Sprintf("too many invvect in message [%v]", count)
		return messageError("MsgGetData.BtcEncode", str)
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
// of the MsgGetData interface implementation.
func (msg *MsgGetData) Command() string {
	return CmdGetData
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the MsgGetData interface implementation.
func (msg *MsgGetData) MaxPayloadLength(pver uint32) uint32 {
	// Num inventory vectors (varInt) + max allowed inventory vectors.
	return uint32(MaxVarIntPayload + (MaxInvPerMsg * maxInvVectPayload))
}

// NewMsgGetData returns a new bitcoin getdata message that conforms to the
// MsgGetData interface.  See MsgGetData for details.
func NewMsgGetData() *MsgGetData {
	return &MsgGetData{
		InvList: make([]*InvVect, 0, defaultInvListAlloc),
	}
}

// NewMsgGetDataSizeHint returns a new bitcoin getdata message that conforms to
// the MsgGetData interface.  See MsgGetData for details.  This function differs
// from NewMsgGetData in that it allows a default allocation size for the
// backing array which houses the inventory vector list.  This allows callers
// who know in advance how large the inventory list will grow to avoid the
// overhead of growing the internal backing array several times when appending
// large amounts of inventory vectors with AddInvVect.  Note that the specified
// hint is just that - a hint that is used for the default allocation size.
// Adding more (or less) inventory vectors will still work properly.  The size
// hint is limited to MaxInvPerMsg.
func NewMsgGetDataSizeHint(sizeHint uint) *MsgGetData {
	// Limit the specified hint to the maximum allow per message.
	if sizeHint > MaxInvPerMsg {
		sizeHint = MaxInvPerMsg
	}

	return &MsgGetData{
		InvList: make([]*InvVect, 0, sizeHint),
	}
}

var _ interfaces.IMsg = (*MsgGetData)(nil)

func (m *MsgGetData) Process(uint32, interfaces.IState) {}

func (m *MsgGetData) GetHash() interfaces.IHash {
	return nil
}

func (m *MsgGetData) GetTimestamp() interfaces.Timestamp {
	return 0
}

func (m *MsgGetData) Type() int {
	return -1
}

func (m *MsgGetData) Int() int {
	return -1
}

func (m *MsgGetData) Bytes() []byte {
	return nil
}

func (m *MsgGetData) UnmarshalBinaryData(data []byte) (newdata []byte, err error) {
	return nil, nil
}

func (m *MsgGetData) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *MsgGetData) MarshalBinary() (data []byte, err error) {
	return nil, nil
}

func (m *MsgGetData) MarshalForSignature() (data []byte, err error) {
	return nil, nil
}

func (m *MsgGetData) String() string {
	return ""
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- MsgGetData is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- MsgGetData is valid
func (m *MsgGetData) Validate(dbheight uint32, state interfaces.IState) int {
	return 0
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *MsgGetData) Leader(state interfaces.IState) bool {
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
func (m *MsgGetData) LeaderExecute(state interfaces.IState) error {
	return nil
}

// Returns true if this is a message for this server to execute as a follower
func (m *MsgGetData) Follower(interfaces.IState) bool {
	return true
}

func (m *MsgGetData) FollowerExecute(interfaces.IState) error {
	return nil
}

func (e *MsgGetData) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *MsgGetData) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *MsgGetData) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}
