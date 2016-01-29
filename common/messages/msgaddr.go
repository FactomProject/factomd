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

// MaxAddrPerMsg is the maximum number of addresses that can be in a single
// bitcoin addr message (MsgAddr).
const MaxAddrPerMsg = 1000

// MsgAddr implements the MsgAddr interface and represents a bitcoin
// addr message.  It is used to provide a list of known active peers on the
// network.  An active peer is considered one that has transmitted a message
// within the last 3 hours.  Nodes which have not transmitted in that time
// frame should be forgotten.  Each message is limited to a maximum number of
// addresses, which is currently 1000.  As a result, multiple messages must
// be used to relay the full list.
//
// Use the AddAddress function to build up the list of known addresses when
// sending an addr message to another peer.
type MsgAddr struct {
	AddrList []*NetAddress
}

// AddAddress adds a known active peer to the message.
func (msg *MsgAddr) AddAddress(na *NetAddress) error {
	if len(msg.AddrList)+1 > MaxAddrPerMsg {
		str := fmt.Sprintf("too many addresses in message [max %v]",
			MaxAddrPerMsg)
		return messageError("MsgAddr.AddAddress", str)
	}

	msg.AddrList = append(msg.AddrList, na)
	return nil
}

// AddAddresses adds multiple known active peers to the message.
func (msg *MsgAddr) AddAddresses(netAddrs ...*NetAddress) error {
	for _, na := range netAddrs {
		err := msg.AddAddress(na)
		if err != nil {
			return err
		}
	}
	return nil
}

// ClearAddresses removes all addresses from the message.
func (msg *MsgAddr) ClearAddresses() {
	msg.AddrList = []*NetAddress{}
}

// BtcDecode decodes r using the bitcoin protocol encoding into the receiver.
// This is part of the MsgAddr interface implementation.
func (msg *MsgAddr) BtcDecode(r io.Reader, pver uint32) error {
	count, err := readVarInt(r, pver)
	if err != nil {
		return err
	}

	// Limit to max addresses per message.
	if count > MaxAddrPerMsg {
		str := fmt.Sprintf("too many addresses for message "+
			"[count %v, max %v]", count, MaxAddrPerMsg)
		return messageError("MsgAddr.BtcDecode", str)
	}

	msg.AddrList = make([]*NetAddress, 0, count)
	for i := uint64(0); i < count; i++ {
		na := NetAddress{}
		err := readNetAddress(r, pver, &na, true)
		if err != nil {
			return err
		}
		msg.AddAddress(&na)
	}
	return nil
}

// BtcEncode encodes the receiver to w using the bitcoin protocol encoding.
// This is part of the MsgAddr interface implementation.
func (msg *MsgAddr) BtcEncode(w io.Writer, pver uint32) error {
	// Protocol versions before MultipleAddressVersion only allowed 1 address
	// per message.
	count := len(msg.AddrList)
	if pver < MultipleAddressVersion && count > 1 {
		str := fmt.Sprintf("too many addresses for message of "+
			"protocol version %v [count %v, max 1]", pver, count)
		return messageError("MsgAddr.BtcEncode", str)

	}
	if count > MaxAddrPerMsg {
		str := fmt.Sprintf("too many addresses for message "+
			"[count %v, max %v]", count, MaxAddrPerMsg)
		return messageError("MsgAddr.BtcEncode", str)
	}

	err := writeVarInt(w, pver, uint64(count))
	if err != nil {
		return err
	}

	for _, na := range msg.AddrList {
		err = writeNetAddress(w, pver, na, true)
		if err != nil {
			return err
		}
	}

	return nil
}

// Command returns the protocol command string for the message.  This is part
// of the MsgAddr interface implementation.
func (msg *MsgAddr) Command() string {
	return CmdAddr
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the MsgAddr interface implementation.
func (msg *MsgAddr) MaxPayloadLength(pver uint32) uint32 {
	if pver < MultipleAddressVersion {
		// Num addresses (varInt) + a single net addresses.
		return MaxVarIntPayload + maxNetAddressPayload(pver)
	}

	// Num addresses (varInt) + max allowed addresses.
	return MaxVarIntPayload + (MaxAddrPerMsg * maxNetAddressPayload(pver))
}

// NewMsgAddr returns a new bitcoin addr message that conforms to the
// MsgAddr interface.  See MsgAddr for details.
func NewMsgAddr() *MsgAddr {
	return &MsgAddr{
		AddrList: make([]*NetAddress, 0, MaxAddrPerMsg),
	}
}

var _ interfaces.IMsg = (*MsgAddr)(nil)

func (m *MsgAddr) Process(uint32, interfaces.IState) {}

func (m *MsgAddr) GetHash() interfaces.IHash {
	return nil
}

func (m *MsgAddr) GetTimestamp() interfaces.Timestamp {
	return 0
}

func (m *MsgAddr) Type() int {
	return -1
}

func (m *MsgAddr) Int() int {
	return -1
}

func (m *MsgAddr) Bytes() []byte {
	return nil
}

func (m *MsgAddr) UnmarshalBinaryData(data []byte) (newdata []byte, err error) {
	return nil, nil
}

func (m *MsgAddr) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *MsgAddr) MarshalBinary() (data []byte, err error) {
	return nil, nil
}

func (m *MsgAddr) MarshalForSignature() (data []byte, err error) {
	return nil, nil
}

func (m *MsgAddr) String() string {
	return ""
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- MsgAddr is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- MsgAddr is valid
func (m *MsgAddr) Validate(dbheight uint32, state interfaces.IState) int {
	return 0
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *MsgAddr) Leader(state interfaces.IState) bool {
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
func (m *MsgAddr) LeaderExecute(state interfaces.IState) error {
	return nil
}

// Returns true if this is a message for this server to execute as a follower
func (m *MsgAddr) Follower(interfaces.IState) bool {
	return true
}

func (m *MsgAddr) FollowerExecute(interfaces.IState) error {
	return nil
}

func (e *MsgAddr) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *MsgAddr) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *MsgAddr) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}
