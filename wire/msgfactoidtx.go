// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"bytes"
	"io"
	fct "github.com/FactomProject/factoid"
)

type IMsgFactoidTX interface {
    // BtcEncode encodes the receiver to w using the bitcoin protocol encoding.
    // This is part of the Message interface implementation.
    BtcEncode(w io.Writer, pver uint32) error 
    // BtcDecode decodes r using the bitcoin protocol encoding into the receiver.
    // This is part of the Message interface implementation.
    BtcDecode(r io.Reader, pver uint32) error 
    // Command returns the protocol command string for the message.  This is part
    // of the Message interface implementation.
    Command() string 
    // MaxPayloadLength returns the maximum length the payload can be for the
    // receiver.  This is part of the Message interface implementation.
    MaxPayloadLength(pver uint32) uint32 
    // NewMsgCommitEntry returns a new bitcoin Commit Entry message that conforms to
    // the Message interface.
    NewMsgFactoidTX() IMsgFactoidTX 
    // Check whether the msg can pass the message level validations
    // such as timestamp, signiture and etc
    IsValid() bool 
    // Create a sha hash from the message binary (output of BtcEncode)
    Sha() (ShaHash, error)     
}
// MsgCommitEntry implements the Message interface and represents a factom
// Commit-Entry message.  It is used by client to commit the entry before
// revealing it.
type MsgFactoidTX struct {
    IMsgFactoidTX
	transaction fct.ITransaction
}

// BtcEncode encodes the receiver to w using the bitcoin protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgFactoidTX) BtcEncode(w io.Writer, pver uint32) error {
    bytes, err := msg.transaction.MarshalBinary()
	if err != nil {
		return err
	}
	// Write the transaction length
	if err := writeElement(w, uint32(len(bytes))); err != nil {
        return err
    }
	// Write the transaction 
	if err := writeVarBytes(w, pver, bytes); err != nil {
		return err
	}

	return nil
}

// BtcDecode decodes r using the bitcoin protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgFactoidTX) BtcDecode(r io.Reader, pver uint32) error {
    var length uint32
    // Read the transaction length
    err := readElement(r , &length)
    if err != nil {
        return err
    }
    // Get the bytes for the transaction
    bytes, err := readVarBytes(r, pver, uint32(length), CmdFactoidTX)
	if err != nil {
		return err
	}
    // Unmarshal
	msg.transaction = new(fct.Transaction)
	err = msg.transaction.UnmarshalBinary(bytes)
	if err != nil {
		return err
	}

	return nil
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *MsgFactoidTX) Command() string {
	return CmdFactoidTX
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *MsgFactoidTX) MaxPayloadLength(pver uint32) uint32 {
	return MaxAppMsgPayload
}

// NewMsgCommitEntry returns a new bitcoin Commit Entry message that conforms to
// the Message interface.
func NewMsgFactoidTX() IMsgFactoidTX {
	return &MsgFactoidTX{}
}

// Check whether the msg can pass the message level validations
// such as timestamp, signiture and etc
func (msg *MsgFactoidTX) IsValid() bool {
    return msg.transaction.Validate() == fct.WELL_FORMED
}

// Create a sha hash from the message binary (output of BtcEncode)
func (msg *MsgFactoidTX) Sha() (ShaHash, error) {

	buf := bytes.NewBuffer(nil)
	msg.BtcEncode(buf, ProtocolVersion)
	var sha ShaHash
	_ = sha.SetBytes(Sha256(buf.Bytes()))

	return sha, nil
}
