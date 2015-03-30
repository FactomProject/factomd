// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package wire

import (
	//	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	//	"strconv"

	"github.com/FactomProject/FactomCode/factoid"
	"github.com/FactomProject/FactomCode/util"
)

const (
	inNout_cap = 16000 // per spec
)

// good check to run after deserialization
func factoid_CountCheck(tx *MsgTx) bool {
	l1 := len(tx.TxIn)
	l2 := len(tx.TxSig)

	return l1 == l2
}

func readRCD(r io.Reader, pver uint32, rcd *RCD) error {
	return nil
}

// readOutPoint reads the next sequence of bytes from r as an OutPoint.
func readOutPoint(r io.Reader, pver uint32, op *OutPoint) error {
	_, err := io.ReadFull(r, op.Hash[:])
	if err != nil {
		return err
	}

	var buf [4]byte
	_, err = io.ReadFull(r, buf[:])
	if err != nil {
		return err
	}

	// op.Index = binary.LittleEndian.Uint32(buf[:])

	// varint on the wire, but easily fits into uint32
	index, err := readVarInt(r, pver)

	if inNout_cap < index {
		return fmt.Errorf("OutPoint trouble, index too large: %d", index)
	}

	op.Index = uint32(index)

	//	return nil
	return err
}

// writeOutPoint encodes op to the protocol encoding for an OutPoint to w.
func writeOutPoint(w io.Writer, pver uint32, op *OutPoint) error {
	_, err := w.Write(op.Hash[:])
	if err != nil {
		return err
	}

	err = writeVarInt(w, pver, uint64(op.Index))
	if err != nil {
		return err
	}

	return nil
}

// readTxIn reads the next sequence of bytes from r as a transaction input
func readTxIn(r io.Reader, pver uint32, ti *TxIn) error {
	var op OutPoint

	err := readOutPoint(r, pver, &op)
	if err != nil {
		return err
	}

	ti.PreviousOutPoint = op

	var buf [1]byte
	_, err = io.ReadFull(r, buf[:])
	if err != nil {
		return err
	}

	ti.sighash = uint8(buf[0])

	return nil
}

// readTxOut reads the next sequence of bytes from r as a transaction output
// (TxOut).
func readTxOut(r io.Reader, pver uint32, to *TxOut) error {
	value, err := readVarInt(r, pver)
	if err != nil {
		return err
	}

	to.Value = int64(value)

	b := make([]byte, 32)
	_, err = io.ReadFull(r, b)

	copy(to.RCDHash[:], b)

	return nil
}

func readECOut(r io.Reader, pver uint32, eco *TxEntryCreditOut) error {
	value, err := readVarInt(r, pver)
	if err != nil {
		return err
	}

	eco.Value = int64(value)

	b := make([]byte, 32)
	_, err = io.ReadFull(r, b)

	copy(eco.ECpubkey[:], b)

	return nil
}

// BtcDecode decodes r using the protocol encoding into the receiver.
// This is part of the Message interface implementation.
// See Deserialize for decoding transactions stored to disk, such as in a
// database, as opposed to decoding transactions from the wire.
func (msg *MsgTx) BtcDecode(r io.Reader, pver uint32) error {
	util.Trace()
	var buf [1]byte
	_, err := io.ReadFull(r, buf[:])
	if err != nil {
		return err
	}

	msg.Version = uint8(buf[0])

	if !factoid.FactoidTx_VersionCheck(msg.Version) {
		return errors.New("fTx version check")
	}

	msg.LockTime = int64(binary.BigEndian.Uint64(buf[:])) // FIXME: must do 5 bytes here

	if !factoid.FactoidTx_LocktimeCheck(msg.LockTime) {
		return errors.New("fTx locktime check")
	}

	outcount, err := readVarInt(r, pver)
	if err != nil {
		return err
	}

	// Prevent more input transactions than could possibly fit into a
	// message.  It would be possible to cause memory exhaustion and panics
	// without a sane upper bound on this count.
	if outcount > uint64(inNout_cap) {
		str := fmt.Sprintf("too many input transactions to fit into "+
			"max message size [count %d, max %d]", outcount,
			inNout_cap)
		return messageError("MsgTx.BtcDecode maxtxout", str)
	}

	msg.TxOut = make([]*TxOut, outcount)
	for i := uint64(0); i < outcount; i++ {
		to := TxOut{}
		err = readTxOut(r, pver, &to)
		if err != nil {
			return err
		}
		msg.TxOut[i] = &to
	}

	eccount, err := readVarInt(r, pver)
	if err != nil {
		return err
	}

	// Prevent more input transactions than could possibly fit into a
	// message.  It would be possible to cause memory exhaustion and panics
	// without a sane upper bound on this count.
	if eccount > uint64(inNout_cap) {
		str := fmt.Sprintf("too many input transactions to fit into "+
			"max message size [count %d, max %d]", eccount,
			inNout_cap)
		return messageError("MsgTx.BtcDecode maxtxout", str)
	}

	// Prevent more input transactions than could possibly fit into a
	// message.  It would be possible to cause memory exhaustion and panics
	// without a sane upper bound on this count.
	if eccount > uint64(inNout_cap) {
		str := fmt.Sprintf("too many input transactions to fit into "+
			"max message size [count %d, max %d]", eccount,
			inNout_cap)
		return messageError("MsgTx.BtcDecode maxtxout", str)
	}

	msg.ECOut = make([]*TxEntryCreditOut, eccount)
	for i := uint64(0); i < eccount; i++ {
		eco := TxEntryCreditOut{}
		err = readECOut(r, pver, &eco)
		if err != nil {
			return err
		}
		msg.ECOut[i] = &eco
	}

	incount, err := readVarInt(r, pver)

	msg.TxIn = make([]*TxIn, incount)
	for i := uint64(0); i < incount; i++ {
		ti := TxIn{}
		err = readTxIn(r, pver, &ti)
		if err != nil {
			return err
		}
		msg.TxIn[i] = &ti
	}

	_, err = io.ReadFull(r, buf[:])
	if err != nil {
		return err
	}

	rcdcount, err := readVarInt(r, pver)

	if rcdcount > uint64(inNout_cap) {
		str := fmt.Sprintf("too many RCDs to fit into "+
			"max message size [count %d, max %d]", rcdcount,
			inNout_cap)
		return messageError("MsgTx.BtcDecode max rcd", str)
	}

	msg.RCD = make([]*RCD, rcdcount)
	for i := uint64(0); i < rcdcount; i++ {
		rcd := RCD{}
		err = readRCD(r, pver, &rcd)
		if err != nil {
			return err
		}
		msg.RCD[i] = &rcd
	}

	// ----------------------------------------------
	if !factoid_CountCheck(msg) {
		errors.New("Factoid check 1")
	}

	return nil
}

// FactoidEncode encodes the receiver to w using the protocol encoding.
// This is part of the Message interface implementation.
// See Serialize for encoding transactions to be stored to disk, such as in a
// database, as opposed to encoding transactions for the wire.
func (msg *MsgTx) BtcEncode(w io.Writer, pver uint32) error {

	util.Trace(" WIP !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")

	var buf [1]byte

	buf[0] = msg.Version

	_, err := w.Write(buf[:])
	if err != nil {
		return err
	}

	var buf8 [8]byte
	binary.LittleEndian.PutUint64(buf8[:], uint64(msg.LockTime)) // FIXME: must do 5 bytes here
	_, err = w.Write(buf8[:])
	if err != nil {
		return err
	}

	txoutcount := uint64(len(msg.TxOut))
	err = writeVarInt(w, pver, txoutcount)
	if err != nil {
		return err
	}

	for _, to := range msg.TxOut {
		err = writeTxOut(w, pver, to)
		if err != nil {
			return err
		}
	}

	ecoutcount := uint64(len(msg.ECOut))
	err = writeVarInt(w, pver, ecoutcount)
	if err != nil {
		return err
	}

	/*
		count = uint64(len(msg.TxOut))
		err = writeVarInt(w, pver, count)
		if err != nil {
			return err
		}

		for _, to := range msg.TxOut {
			err = writeTxOut(w, pver, to)
			if err != nil {
				return err
			}
		}

		binary.BigEndian.PutUint64(buf[:], uint64(msg.LockTime)) // FIXME: must do 5 bytes here
		_, err = w.Write(buf[:])
		if err != nil {
			return err
		}

	*/
	return nil
}

// writeTxIn encodes ti to the protocol encoding for a transaction
// input (TxIn) to w.
func writeTxIn(w io.Writer, pver uint32, ti *TxIn) error {
	util.Trace(" NOT IMPLEMENTED !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")

	/*
		err := writeOutPoint(w, pver, &ti.PreviousOutPoint)
		if err != nil {
			return err
		}

		err = writeVarBytes(w, pver, ti.SignatureScript)
		if err != nil {
			return err
		}

		var buf [4]byte
		binary.BigEndian.PutUint32(buf[:], ti.Sequence)
		_, err = w.Write(buf[:])
		if err != nil {
			return err
		}

	*/
	return nil
}

// writeTxOut encodes to into the protocol encoding for a transaction
// output (TxOut) to w.
func writeTxOut(w io.Writer, pver uint32, to *TxOut) error {
	util.Trace(" NOT IMPLEMENTED !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	/*
		var buf [8]byte
		binary.BigEndian.PutUint64(buf[:], uint64(to.Value))
		_, err := w.Write(buf[:])
		if err != nil {
			return err
		}

		err = writeVarBytes(w, pver, to.PkScript)
		if err != nil {
			return err
		}
	*/
	return nil
}
