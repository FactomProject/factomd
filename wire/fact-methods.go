// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package wire

// Notes - TODO:
// RCD, sig & bitfield primitives not implemented yet

import (
	//	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	//	"os"
	"strconv"

	"github.com/FactomProject/FactomCode/factoid"
	"github.com/FactomProject/FactomCode/util"

	"github.com/davecgh/go-spew/spew"
)

var _ = util.Trace
const (
	TxVersion  = 0
	inNout_cap = 16000 // per spec

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

/*
// good check to run after deserialization
func factoid_CountCheck(tx *MsgTx) bool {
	l1 := len(tx.TxIn)
	l2 := len(tx.TxSig)

	return l1 == l2
}
*/

func readRCD(r io.Reader, pver uint32, rcd *RCDreveal) error {

	return nil
}

func writeRCD(w io.Writer, pver uint32, rcd *RCDreveal) error {
	
	return nil
}

func readSig(r io.Reader, pver uint32, sigCount int, sig *TxSig) error {

	readBitfield(r, pver, sig)
	sig.signatures = make([][]byte, sigCount)
	
	var err error
	for i := 0; i<sigCount; i++ {
		sig.signatures[i] = make([]byte, 64)		
		_, err = io.ReadFull(r, sig.signatures[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func writeSig(w io.Writer, pver uint32, sig *TxSig) error {

	writeBitfield(w, pver, sig)
	
	for i := 0; i < len(sig.signatures); i++ {
		w.Write(sig.signatures[i])
	}

	return nil
}

func readBitfield(r io.Reader, pver uint32, sig *TxSig) error {

	var err error
	sig.bitfield, err = readVarBytes(r, pver, uint32(40), "Bitfield")
	if err != nil {
		return err
	}

	return nil
}

func writeBitfield(w io.Writer, pver uint32, sig *TxSig) error {
	
	err := writeVarBytes(w, pver, sig.bitfield)
	if err != nil {
		return err
	}	

	return nil
}

// readOutPoint reads the next sequence of bytes from r as an OutPoint.
func readOutPoint(r io.Reader, pver uint32, op *OutPoint) error {
	_, err := io.ReadFull(r, op.Hash[:])
	if err != nil {
		return err
	}
	
	// op.Index = binary.BigEndian.Uint32(buf[:])

	// varint on the wire, but easily fits into uint32
	index, err := readVarInt(r, pver)
	
	// coinbase has math.MaxUint32, so that's ok
	if inNout_cap < index && math.MaxUint32 != index {
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

	b := make([]byte, RCDHashSize)
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

	/*
		if s, ok := r.(io.Seeker); ok {
		}
		nowseek, _ := s.Seek(0, os.SEEK_CUR)

		fmt.Println("nowseek= ", nowseek)

		{
			me := bufio.NewReader(r)
			peekbuf, _ := me.Peek(1000)

			fmt.Println("bufio peek= ", spew.Sdump(peekbuf))
		}
	*/

	//	s.Seek(nowseek, os.SEEK_SET)

	var buf [1]byte

	_, err := io.ReadFull(r, buf[:])
	if err != nil {
		return err
	}

	msg.Version = uint8(buf[0])


	if !factoid.FactoidTx_VersionCheck(msg.Version) {
		return errors.New("fTx version check")
	}

	var buf5 [5]byte
	_, err = io.ReadFull(r, buf5[:])
	if err != nil {
		return err
	}

	full8slice := []byte{0, 0, 0}
	full8slice = append(full8slice, buf5[:]...)

	msg.LockTime = int64(binary.BigEndian.Uint64(full8slice))

	if !factoid.FactoidTx_LocktimeCheck(msg.LockTime) {
		return errors.New("fTx decode locktime check")
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
		return messageError("MsgTx.BtcDecode maxecout", str)
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

	msg.RCDreveal = make([]*RCDreveal, rcdcount)
	for i := uint64(0); i < rcdcount; i++ {
		rcd := RCDreveal{}
		err = readRCD(r, pver, &rcd)
		if err != nil {
			return err
		}
		msg.RCDreveal[i] = &rcd
	}

	/* TODO:
	RE - ENABLE
	msg.TxSig = make([]*TxSig, incount)
	for i := uint64(0); i < incount; i++ {
		sig := TxSig{}
		err = readSig(r, pver, &sig)
		if err != nil {
			return err
		}
		msg.TxSig[i] = &sig
	}
	util.Trace()

	// ----------------------------------------------
	if !factoid_CountCheck(msg) {
		errors.New("Factoid check 1")
	}
	*/

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

	if !factoid.FactoidTx_LocktimeCheck(msg.LockTime) {
		return errors.New("fTx encode locktime check")
	}

	binary.BigEndian.PutUint64(buf8[:], uint64(msg.LockTime))
	_, err = w.Write(buf8[3:8]) // LockTime is 5 bytes
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

	rcdcount := uint64(len(msg.RCDreveal))
	err = writeVarInt(w, pver, rcdcount)
	if err != nil {
		return err
	}

	for _, rcd := range msg.RCDreveal {
		err = writeRCD(w, pver, rcd)
		if err != nil {
			return err
		}
	}

	/* TODO: RE-ENABLE
	for _, sig := range msg.TxSig {
		err = writeSig(w, pver, sig)
		if err != nil {
			return err
		}
	}
	*/

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
	err := writeVarInt(w, pver, uint64(to.Value))
	if err != nil {
		return err
	}

	//	err = writeVarBytes(w, pver, to.RCDHash[:])
	_, err = w.Write(to.RCDHash[:])

	if err != nil {
		return err
	}

	return nil
}

// writeECOut encodes to into the protocol encoding for a transaction
// output (TxOut) to w.
func writeECOut(w io.Writer, pver uint32, eco *TxEntryCreditOut) error {
	err := writeVarInt(w, pver, uint64(eco.Value))
	if err != nil {
		return err
	}

	_, err = w.Write(eco.ECpubkey[:])
	if err != nil {
		return err
	}

	return nil
}

// SerializeSize returns the number of bytes it would take to serialize the
// the transaction output.
func (t *TxOut) SerializeSize() int {
	return RCDHashSize + VarIntSerializeSize(uint64(t.Value))
}

func (t *TxEntryCreditOut) SerializeSize() int {
	return PubKeySize + VarIntSerializeSize(uint64(t.Value))
}

// we are using the OutPoint here; thus the total transaction input size is: txid (32) + index (4) + sighash (1)
func (t *TxIn) SerializeSize() int {
	return 37
}

func (rcd *RCDreveal) SerializeSize() int {
	
	return 0
}

func (sig *TxSig) SerializeSize() int {
	n := 0
	// get the size for bitfield
	n += VarIntSerializeSize(uint64(len(sig.bitfield)))
	n += len(sig.bitfield)

	for _, signature := range sig.signatures {
		n += len(signature)
	}

	return n
}

// SerializeSize returns the number of bytes it would take to serialize the
// the transaction.
func (msg *MsgTx) SerializeSize() int {

	n := 1 + // 1 byte version
		5 // 5 bytes locktime

	n += VarIntSerializeSize(uint64(len(msg.TxOut)))

	for _, txOut := range msg.TxOut {
		n += txOut.SerializeSize()
	}

	n += VarIntSerializeSize(uint64(len(msg.ECOut)))

	for _, ecOut := range msg.ECOut {
		n += ecOut.SerializeSize()
	}

	n += VarIntSerializeSize(uint64(len(msg.TxIn)))

	for _, txIn := range msg.TxIn {
		n += txIn.SerializeSize()
	}

	n += VarIntSerializeSize(uint64(len(msg.RCDreveal)))

	for _, rcd := range msg.RCDreveal {
		n += rcd.SerializeSize()
	}

	// FIXME
	// TODO: count TxSig impact here
	
	return n
}

// TxSha generates the ShaHash name for the transaction.
func (msg *MsgTx) TxSha() (ShaHash, error) {
	
	fmt.Println("TxSha spew: ", spew.Sdump(*msg))

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

	return sha, nil
}

func (msg *MsgTx) Deserialize(r io.Reader) error {
	// At the current time, there is no difference between the wire encoding
	// at protocol version 0 and the stable long-term storage format.  As
	// a result, make use of BtcDecode.
	return msg.BtcDecode(r, 0)
}

func (msg *MsgTx) Serialize(w io.Writer) error {
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

// AddECOut adds a transaction output to the message.
func (msg *MsgTx) AddECOut(eco *TxEntryCreditOut) {
	msg.ECOut = append(msg.ECOut, eco)
}

// AddRCD adds a RCD to the message.
func (msg *MsgTx) AddRCD(rcd *RCDreveal) {
	msg.RCDreveal = append(msg.RCDreveal, rcd)
}

// NewMsgTx returns a new tx message that conforms to the Message
// interface.  The return instance has a default version of TxVersion and there
// are no transaction inputs or outputs.  Also, the lock time is set to zero
// to indicate the transaction is valid immediately as opposed to some time in
// future.
func NewMsgTx() *MsgTx {
	return &MsgTx{
		Version:  TxVersion,
		LockTime: 0,
		//		LockTime:  0x123456789A, // FIXME: this is for testing only
		TxOut:     make([]*TxOut, 0, defaultTxInOutAlloc),
		ECOut:     make([]*TxEntryCreditOut, 0, defaultTxInOutAlloc),
		TxIn:      make([]*TxIn, 0, defaultTxInOutAlloc),
		RCDreveal: make([]*RCDreveal, 0, defaultTxInOutAlloc),
		//		TxSig:    make([]*TxSig, 0, defaultTxInOutAlloc), // TODO: RE-ENABLE
	}
}

// String returns the OutPoint in the human-readable form "hash:index".
func (o OutPoint) String() string {
	// Allocate enough for hash string, colon, and 10 digits.  Although
	// at the time of writing, the number of digits can be no greater than
	// the length of the decimal representation of maxTxOutPerMessage, the
	// maximum message payload may increase in the future and this
	// optimization may go unnoticed, so allocate space for 10 decimal
	// digits, which will fit any uint32.
	buf := make([]byte, 2*HashSize+1, 2*HashSize+1+10)
	copy(buf, o.Hash.String())
	buf[2*HashSize] = ':'
	buf = strconv.AppendUint(buf, uint64(o.Index), 10)

	return string(buf)
}
