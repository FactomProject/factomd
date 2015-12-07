// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"bytes"
	"io"

	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/interfaces"
)

// MsgEOM implements the Message interface and represents a factom
// End of Minute (EOM) message.
type MsgEOM struct {
	EOM *messages.EOM
}

// NewMsgEOM returns a new EOM message that conforms to the
// Message interface.
func NewMsgEOM() *MsgEOM {
	m := new(MsgEOM)
	m.EOM = new(messages.EOM)
	return m
}

// BtcEncode encodes the receiver to w using the bitcoin protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgEOM) BtcEncode(w io.Writer, pver uint32) error {
	bytes, err := msg.EOM.MarshalBinary()
	if err != nil {
		return err
	}

	if err := writeVarBytes(w, pver, bytes); err != nil {
		return err
	}

	return nil
}

// BtcDecode decodes r using the bitcoin protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgEOM) BtcDecode(r io.Reader, pver uint32) error {
	bytes, err := readVarBytes(r, pver, uint32(8+1+4+32+32+64), CmdEOM)
	if err != nil {
		return err
	}

	msg.EOM = new(messages.EOM)
	if err := msg.EOM.UnmarshalBinary(bytes); err != nil {
		return err
	}

	return nil
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *MsgEOM) Command() string {
	return CmdEOM
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *MsgEOM) MaxPayloadLength(pver uint32) uint32 {
	return MaxAppMsgPayload
}

// Check whether the msg can pass the message level validations
// such as timestamp, signiture and etc
func (msg *MsgEOM) IsValid() bool {
	//return msg.EOM.IsValid()
	return true
}

// Create a sha hash from the message binary (output of BtcEncode)
func (msg *MsgEOM) Sha() (interfaces.IHash, error) {

	buf := bytes.NewBuffer(nil)
	msg.BtcEncode(buf, ProtocolVersion)
	var sha interfaces.IHash
	_ = sha.SetBytes(Sha256(buf.Bytes()))

	return sha, nil
}
