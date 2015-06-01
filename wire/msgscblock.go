// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire


import (
	"io"
    "github.com/FactomProject/simplecoin/block"
)

// Simplecoin block
type MsgSCBlock struct {
	SC block.ISCBlock
}

// BtcEncode encodes the receiver to w using the bitcoin protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgSCBlock) BtcEncode(w io.Writer, pver uint32) error {

	bytes, err := msg.SC.MarshalBinary()
	if err != nil {
		return err
	}

	err = writeVarBytes(w, pver, bytes)
	if err != nil {
		return err
	}

	return nil
}

// BtcDecode decodes r using the bitcoin protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgSCBlock) BtcDecode(r io.Reader, pver uint32) error {

	bytes, err := readVarBytes(r, pver, uint32(10000), CmdSCBlock)
	if err != nil {
		return err
	}

	msg.SC = new(block.SCBlock)
	err = msg.SC.UnmarshalBinary(bytes)
	if err != nil {
		return err
	}

	return nil
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *MsgSCBlock) Command() string {
	return CmdSCBlock
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *MsgSCBlock) MaxPayloadLength(pver uint32) uint32 {
	return MaxAppMsgPayload
}

// NewMsgABlock returns a new bitcoin inv message that conforms to the Message
// interface.  See MsgInv for details.
func NewMsgSCBlock() *MsgSCBlock {
    return &MsgSCBlock{}
}
