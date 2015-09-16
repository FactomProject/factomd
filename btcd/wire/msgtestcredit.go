// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/FactomProject/FactomCode/common"
)

type MsgTestCredit struct {
	ECKey *[32]byte
	Amt   int32
}

func NewMsgTestCredit() *MsgTestCredit {
	m := new(MsgTestCredit)
	m.ECKey = new([32]byte)
	return m
}

func (m *MsgTestCredit) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.Write(m.ECKey[:])
	if err := binary.Write(buf, binary.BigEndian, m.Amt); err != nil {
		return buf.Bytes(), err
	}
	return buf.Bytes(), nil
}

func (m *MsgTestCredit) UnmarshalBinary(data []byte) error {
	buf := bytes.NewBuffer(data)
	if p := buf.Next(32); len(p) != 32 {
		return fmt.Errorf("Bad Msg length: %v", m)
	} else {
		copy(m.ECKey[:], p)
	}
	if err := binary.Read(buf, binary.BigEndian, m.Amt); err != nil {
		return err
	}
	return nil
}

// BtcEncode encodes the receiver to w using the bitcoin protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgTestCredit) BtcEncode(w io.Writer, pver uint32) error {
	bytes, err := msg.MarshalBinary()
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
func (msg *MsgTestCredit) BtcDecode(r io.Reader, pver uint32) error {
	bytes, err := readVarBytes(r, pver, uint32(common.CommitEntrySize),
		CmdEntry)
	if err != nil {
		return err
	}

	if err = msg.UnmarshalBinary(bytes); err != nil {
		return err
	}

	return nil
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *MsgTestCredit) Command() string {
	return CmdTestCredit
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *MsgTestCredit) MaxPayloadLength(pver uint32) uint32 {
	return MaxAppMsgPayload
}

// Create a sha hash from the message binary (output of BtcEncode)
func (msg *MsgTestCredit) Sha() (ShaHash, error) {

	buf := bytes.NewBuffer(nil)
	msg.BtcEncode(buf, ProtocolVersion)
	var sha ShaHash
	_ = sha.SetBytes(Sha256(buf.Bytes()))

	return sha, nil
}
