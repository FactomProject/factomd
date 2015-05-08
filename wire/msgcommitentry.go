// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/FactomProject/FactomCode/common"
	"github.com/agl/ed25519"
)

// MsgCommitEntry implements the Message interface and represents a factom
// Commit-Entry message.  It is used by client to commit the entry before
// revealing it.
type MsgCommitEntry struct {
	Version   int8
	MilliTime *[6]byte
	EntryHash *common.Hash
	Credits   uint8
	ECPubKey  *[32]byte
	Sig       *[64]byte
}

// BtcEncode encodes the receiver to w using the bitcoin protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgCommitEntry) BtcEncode(w io.Writer, pver uint32) error {
	// Version
	if err := writeElement(w, &msg.Version); err != nil {
		return err
	}

	//MilliTime
	if err := writeVarBytes(w, pver, msg.MilliTime[:]); err != nil {
		return err
	}

	//EntryHash
	if err := writeVarBytes(w, uint32(common.HASH_LENGTH), msg.EntryHash.Bytes); err != nil {
		return err
	}

	//Credits
	if err := writeElement(w, &msg.Credits); err != nil {
		return err
	}

	//ECPubKey
	if err := writeVarBytes(w, uint32(ed25519.PublicKeySize), msg.ECPubKey[:]); err != nil {
		return err
	}

	//Signature
	if err := writeVarBytes(w, uint32(ed25519.SignatureSize), msg.Sig[:]); err != nil {
		return err
	}

	return nil
}

// BtcDecode decodes r using the bitcoin protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgCommitEntry) BtcDecode(r io.Reader, pver uint32) error {
	// Version
	if err := readElement(r, &msg.Version); err != nil {
		return err
	}

	// MilliTime
	if bytes, err := readVarBytes(r, pver, uint32(6), CmdCommitEntry); err != nil {
		return err
	} else {
		copy(msg.MilliTime[:], bytes)
	}

	// EntryHash
	if bytes, err := readVarBytes(r, pver, uint32(common.HASH_LENGTH),
		CmdCommitEntry); err != nil {
		return err
	} else {
		copy(msg.EntryHash.Bytes, bytes[:32])
	}

	// Credits
	if err := readElement(r, &msg.Credits); err != nil {
		return err
	}

	// ECPubKey
	if bytes, err := readVarBytes(r, pver, uint32(ed25519.PublicKeySize),
		CmdCommitEntry); err != nil {
		return err
	} else {
		msg.ECPubKey = new([32]byte)
		copy(msg.ECPubKey[:], bytes)
	}

	// Signature
	if bytes, err := readVarBytes(r, pver, uint32(ed25519.SignatureSize),
		CmdCommitEntry); err != nil {
		return err
	} else {
		msg.Sig = new([64]byte)
		copy(msg.Sig[:], bytes)
	}

	return nil
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *MsgCommitEntry) Command() string {
	return CmdCommitEntry
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *MsgCommitEntry) MaxPayloadLength(pver uint32) uint32 {
	return MaxAppMsgPayload
}

// NewMsgCommitEntry returns a new bitcoin Commit Entry message that conforms to
// the Message interface.
func NewMsgCommitEntry() *MsgCommitEntry {
	m := new(MsgCommitEntry)
	m.MilliTime = new([6]byte)
	m.EntryHash = new(common.Hash)
	m.EntryHash.Bytes = make([]byte, 32)
	m.ECPubKey = new([32]byte)
	m.Sig = new([64]byte)

	return m
}

// Check whether the msg can pass the message level validations
// such as timestamp, signiture and etc
func (msg *MsgCommitEntry) IsValid() bool {
	// Verify signature (Version + MilliTime + EntryHash + Credits)
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, msg.Version)
	buf.Write(msg.MilliTime[:])
	buf.Write(msg.EntryHash.Bytes)
	binary.Write(buf, binary.BigEndian, msg.Credits)

	return ed25519.Verify(msg.ECPubKey, buf.Bytes(), msg.Sig)
}

// Create a sha hash from the message binary (output of BtcEncode)
func (msg *MsgCommitEntry) Sha() (ShaHash, error) {

	buf := bytes.NewBuffer(nil)
	msg.BtcEncode(buf, ProtocolVersion)
	var sha ShaHash
	_ = sha.SetBytes(Sha256(buf.Bytes()))

	return sha, nil
}
