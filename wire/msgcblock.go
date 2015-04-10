// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"io"

	"github.com/FactomProject/FactomCode/common"
)

// MsgCBlock implements the Message interface and represents a factom
// Reveal-Entry message.  It is used by client to reveal the entry.
type MsgCBlock struct {
	CBlk *common.CBlock
}

// BtcEncode encodes the receiver to w using the bitcoin protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgCBlock) BtcEncode(w io.Writer, pver uint32) error {

	bytes, err := msg.CBlk.MarshalBinary()
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
func (msg *MsgCBlock) BtcDecode(r io.Reader, pver uint32) error {
	//Entry
	bytes, err := readVarBytes(r, pver, uint32(10000), CmdRevealEntry)
	if err != nil {
		return err
	}

	msg.CBlk = new(common.CBlock)
	err = msg.CBlk.UnmarshalBinary(bytes)
	if err != nil {
		return err
	}

	return nil
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *MsgCBlock) Command() string {
	return CmdCBlock
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *MsgCBlock) MaxPayloadLength(pver uint32) uint32 {
	return MaxAppMsgPayload
}

// NewMsgCBlock returns a new bitcoin inv message that conforms to the Message
// interface.  See MsgInv for details.
func NewMsgCBlock() *MsgCBlock {
	return &MsgCBlock{}
}
