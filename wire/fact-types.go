// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package wire

const RCDHashSize = 32
const PubKeySize = 32

type RCDHash [RCDHashSize]byte
type ECPubKey [PubKeySize]byte
type PubKey [PubKeySize]byte

// Use the AddTxIn and AddTxOut functions to build up the list of transaction
// inputs and outputs.
type MsgTx struct {
	Version  uint8
	LockTime int64 // is really 5 bytes, TODO: fix in Encode/Decode

	//	FactoidOut []*TxFactoidOut
	TxOut []*TxOut
	ECOut []*TxEntryCreditOut
	TxIn  []*TxIn
	RCD   []*RCD
	TxSig []*TxSig
}

// type TxFactoidOut struct {
type TxOut struct {
	Value   int64
	RCDHash RCDHash
}

type TxEntryCreditOut struct {
	Value    int64
	ECpubkey ECPubKey
}

type TxIn struct {
	PreviousOutPoint OutPoint
	sighash          uint8 // sighash type
}

// OutPoint defines a bitcoin data type that is used to track previous
// transaction outputs.
type OutPoint struct {
	Hash  ShaHash
	Index uint32
}

// NewOutPoint returns a new bitcoin transaction outpoint point with the
// provided hash and index.
func NewOutPoint(hash *ShaHash, index uint32) *OutPoint {
	return &OutPoint{
		Hash:  *hash,
		Index: index,
	}
}

/*
// OutPoint defines a bitcoin data type that is used to track previous
// transaction outputs.
type OutPoint struct {
	txid [32]byte
	idx  uint64
}
*/

type RCD struct {
	Version     uint8
	Type        uint8
	PubKey      []PubKey
	MinRequired uint64 // minimum number of keys required for validity
}

type TxSig struct {
	bitfield   uint64 // TODO: extend to beyond 64bits
	signatures [64][]byte
}
