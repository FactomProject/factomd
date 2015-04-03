// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package wire

// Notes - TODO:
// Value is fixed uint64 right now, must be VarInt
// LockTime is int64 right now, must be 5 bytes
// RCD, sig & bitfield primitives not implemented yet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	//	"strconv"

	"github.com/FactomProject/FactomCode/factoid"
	"github.com/FactomProject/FactomCode/util"

	"github.com/davecgh/go-spew/spew"
)

const (
	TxVersion  = 0
	inNout_cap = 16000 // per spec

	//	MaxPrevOutIndex = 0xffffffff // there are some checks that expect math.MaxUint32 here... hm: IsCoinBase()
	MaxPrevOutIndex = math.MaxUint32 // there are some checks that expect math.MaxUint32 here... hm: IsCoinBase()

	// minTxPayload is the minimum payload size for a transaction.  Note
	// that any realistically usable transaction must have at least one
	// input or output, but that is a rule enforced at a higher layer, so
	// it is intentionally not included here.
	// Version 4 bytes + Varint number of transaction inputs 1 byte + Varint
	// number of transaction outputs 1 byte + LockTime 4 bytes + min input
	// payload + min output payload.
	minTxPayload = 10 // TODO: revisit for Factoid, blind copy from Bitcoin

	defaultTxInOutAlloc = 4
)

// 10KiB is the limit for entries, per Brian we apply the same here
func (msg *MsgTx) MaxPayloadLength(pver uint32) uint32 {
	return 1024 * 10
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *MsgTx) Command() string {
	return CmdTx
}

// good check to run after deserialization
func factoid_CountCheck(tx *MsgTx) bool {
	l1 := len(tx.TxIn)
	l2 := len(tx.TxSig)

	return l1 == l2
}

func readRCD(r io.Reader, pver uint32, rcd *RCD) error {
	util.Trace("NOT IMPLEMENTED !!!")

	return nil
}

func writeRCD(w io.Writer, pver uint32, rcd *RCD) error {
	util.Trace("NOT IMPLEMENTED !!!")

	return nil
}

func readSig(r io.Reader, pver uint32, sig *TxSig) error {
	util.Trace("NOT IMPLEMENTED !!!")

	readBitfield(r, pver)

	return nil
}

func writeSig(w io.Writer, pver uint32, sig *TxSig) error {
	util.Trace("NOT IMPLEMENTED !!!")

	writeBitfield(w, pver, sig)

	return nil
}

func readBitfield(r io.Reader, pver uint32) error {
	util.Trace("NOT IMPLEMENTED !!!")

	return nil
}

func writeBitfield(w io.Writer, pver uint32, sig *TxSig) error {
	util.Trace("NOT IMPLEMENTED !!!")

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

	to.Value = int64(value) // FIXME: make it VarInt

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

	eco.Value = int64(value) // FIXME: make it VarInt

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

	/*
		txid, _ := msg.TxSha()
		fmt.Println("BtcDecode, txid: ", txid, spew.Sdump(msg))
	*/
	fmt.Println("BtcDecode spew: ", spew.Sdump(msg))

	var buf [1]byte

	_, err := io.ReadFull(r, buf[:])
	if err != nil {
		return err
	}

	msg.Version = uint8(buf[0])

	fmt.Printf("buf= %v (%d)\n", buf, msg.Version)

	if !factoid.FactoidTx_VersionCheck(msg.Version) {
		return errors.New("fTx version check")
	}

	var buf8 [8]byte
	_, err = io.ReadFull(r, buf8[:])
	if err != nil {
		return err
	}

	fmt.Printf("buf8= %v\n", buf8)

	msg.LockTime = int64(binary.BigEndian.Uint64(buf8[:])) // FIXME: must do 5 bytes here

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

	msg.TxSig = make([]*TxSig, incount)
	for i := uint64(0); i < incount; i++ {
		sig := TxSig{}
		err = readSig(r, pver, &sig)
		if err != nil {
			return err
		}
		msg.TxSig[i] = &sig
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

	for _, eco := range msg.ECOut {
		err = writeECOut(w, pver, eco)
		if err != nil {
			return err
		}
	}

	incount := uint64(len(msg.TxIn))
	err = writeVarInt(w, pver, incount)
	if err != nil {
		return err
	}

	for _, ti := range msg.TxIn {
		err = writeTxIn(w, pver, ti)
		if err != nil {
			return err
		}
	}

	rcdcount := uint64(len(msg.RCD))
	err = writeVarInt(w, pver, rcdcount)
	if err != nil {
		return err
	}

	for _, rcd := range msg.RCD {
		err = writeRCD(w, pver, rcd)
		if err != nil {
			return err
		}
	}

	for _, sig := range msg.TxSig {
		err = writeSig(w, pver, sig)
		if err != nil {
			return err
		}
	}

	return nil
}

// writeTxIn encodes ti to the protocol encoding for a transaction
// input (TxIn) to w.
func writeTxIn(w io.Writer, pver uint32, ti *TxIn) error {
	err := writeOutPoint(w, pver, &ti.PreviousOutPoint)
	if err != nil {
		return err
	}

	var buf [1]byte

	buf[0] = ti.sighash

	_, err = w.Write(buf[:])
	if err != nil {
		return err
	}

	return nil
}

// writeTxOut encodes to into the protocol encoding for a transaction
// output (TxOut) to w.
func writeTxOut(w io.Writer, pver uint32, to *TxOut) error {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], uint64(to.Value)) // FIXME: make it VarInt
	_, err := w.Write(buf[:])
	if err != nil {
		return err
	}

	err = writeVarBytes(w, pver, to.RCDHash[:])
	if err != nil {
		return err
	}

	return nil
}

// writeECOut encodes to into the protocol encoding for a transaction
// output (TxOut) to w.
func writeECOut(w io.Writer, pver uint32, eco *TxEntryCreditOut) error {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], uint64(eco.Value)) // FIXME: make it VarInt
	_, err := w.Write(buf[:])
	if err != nil {
		return err
	}

	err = writeVarBytes(w, pver, eco.ECpubkey[:])
	if err != nil {
		return err
	}

	return nil
}

// SerializeSize returns the number of bytes it would take to serialize the
// the transaction output.
func (t *TxOut) SerializeSize() int {
	// FIXME: switch to this VarInt calculation
	//	return 32 + VarIntSerializeSize(uint64(t.Value))  // this is good for VarInt.. future

	return 40 // TODO: replace with the VarInt calc, this is for fixed size value (8 bytes)
}

func (t *TxEntryCreditOut) SerializeSize() int {
	// FIXME: make it VarInt
	return 40
}

// the transaction input: txid (32) + vaue (8) + sighash (1)
// FIXME: make it VarInt
func (t *TxIn) SerializeSize() int {
	return 41
}

// SerializeSize returns the number of bytes it would take to serialize the
// the transaction.
func (msg *MsgTx) SerializeSize() int {
	n := 1 + // 1 byte version
		8 // FIXME: 5 bytes locktime

	for _, txOut := range msg.TxOut {
		n += txOut.SerializeSize()
	}

	for _, ecOut := range msg.ECOut {
		n += ecOut.SerializeSize()
	}

	for _, txIn := range msg.TxIn {
		n += txIn.SerializeSize()
	}

	return n
}

// TxSha generates the ShaHash name for the transaction.
func (msg *MsgTx) TxSha() (ShaHash, error) {
	util.Trace()

	// Encode the transaction and calculate double sha256 on the result.
	// Ignore the error returns since the only way the encode could fail
	// is being out of memory or due to nil pointers, both of which would
	// cause a run-time panic.  Also, SetBytes can't fail here due to the
	// fact DoubleSha256 always returns a []byte of the right size
	// regardless of input.
	buf := bytes.NewBuffer(make([]byte, 0, msg.SerializeSize()))
	_ = msg.Serialize(buf)
	var sha ShaHash
	_ = sha.SetBytes(DoubleSha256(buf.Bytes()))

	// Even though this function can't currently fail, it still returns
	// a potential error to help future proof the API should a failure
	// become possible.

	util.Trace(sha.String())

	return sha, nil
}

func (msg *MsgTx) Deserialize(r io.Reader) error {
	util.Trace()
	// At the current time, there is no difference between the wire encoding
	// at protocol version 0 and the stable long-term storage format.  As
	// a result, make use of BtcDecode.
	return msg.BtcDecode(r, 0)
}

func (msg *MsgTx) Serialize(w io.Writer) error {
	util.Trace()
	// At the current time, there is no difference between the wire encoding
	// at protocol version 0 and the stable long-term storage format.  As
	// a result, make use of BtcEncode.

	return msg.BtcEncode(w, 0)
}

// NewTxIn returns a new bitcoin transaction input with the provided
// previous outpoint point and signature script with a default sequence of
// MaxTxInSequenceNum.
func NewTxIn(prevOut *OutPoint, signatureScript []byte) *TxIn {
	return &TxIn{
		PreviousOutPoint: *prevOut,
		//    SignatureScript:  signatureScript,
		//    Sequence:         MaxTxInSequenceNum,
	}
}

// NewTxOut returns a new bitcoin transaction output with the provided
// transaction value and public key.
func NewTxOut(value int64, pk []byte) *TxOut {
	var rcdhash RCDHash

	copy(rcdhash[:], pk)

	return &TxOut{
		Value:   value,
		RCDHash: rcdhash,
	}
}

// AddTxIn adds a transaction input to the message.
func (msg *MsgTx) AddTxIn(ti *TxIn) {
	msg.TxIn = append(msg.TxIn, ti)
}

// AddTxOut adds a transaction output to the message.
func (msg *MsgTx) AddTxOut(to *TxOut) {
	msg.TxOut = append(msg.TxOut, to)
}

// AddRCD adds a RCD to the message.
func (msg *MsgTx) AddRCD(rcd *RCD) {
	msg.RCD = append(msg.RCD, rcd)
}

// NewMsgTx returns a new tx message that conforms to the Message
// interface.  The return instance has a default version of TxVersion and there
// are no transaction inputs or outputs.  Also, the lock time is set to zero
// to indicate the transaction is valid immediately as opposed to some time in
// future.
func NewMsgTx() *MsgTx {
	return &MsgTx{
		Version:  TxVersion,
		LockTime: 0, // FIXME: ensure 5-byte locktime
		TxOut:    make([]*TxOut, 0, defaultTxInOutAlloc),
		ECOut:    make([]*TxEntryCreditOut, 0, defaultTxInOutAlloc),
		TxIn:     make([]*TxIn, 0, defaultTxInOutAlloc),
		RCD:      make([]*RCD, 0, defaultTxInOutAlloc),
		TxSig:    make([]*TxSig, 0, defaultTxInOutAlloc),
	}
}
