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

// defaultInvListAlloc is the default size used for the backing array for an
// inventory list.  The array will dynamically grow as needed, but this
// figure is intended to provide enough space for the max number of inventory
// vectors in a *typical* inventory message without needing to grow the backing
// array multiple times.  Technically, the list can grow to MaxInvPerMsg, but
// rather than using that large figure, this figure more accurately reflects the
// typical case.
const defaultInvListAlloc = 1000

// MsgInv implements the MsgInv interface and represents a bitcoin inv message.
// It is used to advertise a peer's known data such as blocks and transactions
// through inventory vectors.  It may be sent unsolicited to inform other peers
// of the data or in response to a getblocks message (MsgGetBlocks).  Each
// message is limited to a maximum number of inventory vectors, which is
// currently 50,000.
//
// Use the AddInvVect function to build up the list of inventory vectors when
// sending an inv message to another peer.
type MsgInv struct {
	InvList []*InvVect
}

// AddInvVect adds an inventory vector to the message.
func (msg *MsgInv) AddInvVect(iv *InvVect) error {
	if len(msg.InvList)+1 > MaxInvPerMsg {
		str := fmt.Sprintf("too many invvect in message [max %v]",
			MaxInvPerMsg)
		return messageError("MsgInv.AddInvVect", str)
	}

	msg.InvList = append(msg.InvList, iv)
	return nil
}

// BtcDecode decodes r using the bitcoin protocol encoding into the receiver.
// This is part of the MsgInv interface implementation.
func (msg *MsgInv) BtcDecode(r io.Reader, pver uint32) error {
	count, err := readVarInt(r, pver)
	if err != nil {
		return err
	}

	// Limit to max inventory vectors per message.
	if count > MaxInvPerMsg {
		str := fmt.Sprintf("too many invvect in message [%v]", count)
		return messageError("MsgInv.BtcDecode", str)
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
// This is part of the MsgInv interface implementation.
func (msg *MsgInv) BtcEncode(w io.Writer, pver uint32) error {
	// Limit to max inventory vectors per message.
	count := len(msg.InvList)
	if count > MaxInvPerMsg {
		str := fmt.Sprintf("too many invvect in message [%v]", count)
		return messageError("MsgInv.BtcEncode", str)
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
// of the MsgInv interface implementation.
func (msg *MsgInv) Command() string {
	return CmdInv
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the MsgInv interface implementation.
func (msg *MsgInv) MaxPayloadLength(pver uint32) uint32 {
	// Num inventory vectors (varInt) + max allowed inventory vectors.
	return uint32(MaxVarIntPayload + (MaxInvPerMsg * maxInvVectPayload))
}

// NewMsgInv returns a new bitcoin inv message that conforms to the MsgInv
// interface.  See MsgInv for details.
func NewMsgInv() *MsgInv {
	return &MsgInv{
		InvList: make([]*InvVect, 0, defaultInvListAlloc),
	}
}

// NewMsgInvSizeHint returns a new bitcoin inv message that conforms to the
// MsgInv interface.  See MsgInv for details.  This function differs from
// NewMsgInv in that it allows a default allocation size for the backing array
// which houses the inventory vector list.  This allows callers who know in
// advance how large the inventory list will grow to avoid the overhead of
// growing the internal backing array several times when appending large amounts
// of inventory vectors with AddInvVect.  Note that the specified hint is just
// that - a hint that is used for the default allocation size.  Adding more
// (or less) inventory vectors will still work properly.  The size hint is
// limited to MaxInvPerMsg.
func NewMsgInvSizeHint(sizeHint uint) *MsgInv {
	// Limit the specified hint to the maximum allow per message.
	if sizeHint > MaxInvPerMsg {
		sizeHint = MaxInvPerMsg
	}

	return &MsgInv{
		InvList: make([]*InvVect, 0, sizeHint),
	}
}

var _ interfaces.IMsg = (*MsgInv)(nil)

func (m *MsgInv) Process(uint32, interfaces.IState) {}

func (m *MsgInv) GetHash() interfaces.IHash {
	return nil
}

func (m *MsgInv) GetTimestamp() interfaces.Timestamp {
	return 0
}

func (m *MsgInv) Type() int {
	return -1
}

func (m *MsgInv) Int() int {
	return -1
}

func (m *MsgInv) Bytes() []byte {
	return nil
}

func (m *MsgInv) UnmarshalBinaryData(data []byte) (newdata []byte, err error) {
	return nil, nil
}

func (m *MsgInv) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *MsgInv) MarshalBinary() (data []byte, err error) {
	return nil, nil
}

func (m *MsgInv) MarshalForSignature() (data []byte, err error) {
	return nil, nil
}

func (m *MsgInv) String() string {
	return ""
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- MsgInv is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- MsgInv is valid
func (m *MsgInv) Validate(dbheight uint32, state interfaces.IState) int {
	return 0
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *MsgInv) Leader(state interfaces.IState) bool {
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
func (m *MsgInv) LeaderExecute(state interfaces.IState) error {
	return nil
}

// Returns true if this is a message for this server to execute as a follower
func (m *MsgInv) Follower(interfaces.IState) bool {
	return true
}

func (m *MsgInv) FollowerExecute(interfaces.IState) error {
	return nil
}

func (e *MsgInv) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *MsgInv) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *MsgInv) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}
