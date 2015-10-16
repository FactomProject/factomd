// Copyright (c) 2013-2015 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"bytes"
	"fmt"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	. "github.com/FactomProject/factomd/common/primitives"
	"io"
)

var (
	// Shared constants
	FChainID        interfaces.IHash
	CreditsPerChain int32

	// BTCD State Variables
	FactoshisPerCredit uint64
)

// Factom Constants for BTCD and Factom
//
func Init() {
	barray := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0F}
	FChainID = new(Hash)
	FChainID.SetBytes(barray)

	CreditsPerChain = 10 // Entry Credits to create a chain

	// Shouldn't set this, but we are for now.
	FactoshisPerCredit = 666667 // .001 / .15 * 100000000 (assuming a Factoid is .15 cents, entry credit = .1 cents

}

// BlockHeader defines information about a block and is used in the bitcoin
// block (MsgBlock) and headers (MsgHeaders) messages.
type BlockHeader struct {
	ChainID    interfaces.IHash // ChainID.  But since this is a constant, we need not actually use space to store it.
	MerkleRoot interfaces.IHash // Merkle root of the Factoid transactions which accompany this block.
	PrevBlock  interfaces.IHash // Key Merkle root of previous block.
	PrevHash3  Sha3Hash         // Sha3 of the previous Factoid Block
	ExchRate   uint64           // Factoshis per Entry Credit
	DBHeight   uint32           // Directory Block height
	UTXOCommit interfaces.IHash // This field will hold a Merkle root of an array containing all unspent transactions.

	// transaction count & body size are "read-only" (future) fields since serialization logic is handling both
	TransCnt uint64 // Count of transactions in this block
	BodySize uint64 // Bytes in the body of this block.
}

// blockHeaderLen is a constant that represents the number of bytes for a block
// header.
const blockHeaderLen = 28 + 5*constants.HASH_LENGTH

// BlockSha computes the block identifier hash for the given block header.
func (h *BlockHeader) BlockSha() (interfaces.IHash, error) {
	// Encode the header and run double sha256 everything prior to the
	// number of transactions.  Ignore the error returns since there is no
	// way the encode could fail except being out of memory which would
	// cause a run-time panic.  Also, SetBytes can't fail here due to the
	// fact DoubleSha256 always returns a []byte of the right size
	// regardless of input.
	var buf bytes.Buffer
	var sha interfaces.IHash
	_ = writeBlockHeader(&buf, 0, h)
	fmt.Println("Len: ", len(buf.Bytes()), " ", blockHeaderLen)
	_ = sha.SetBytes(DoubleSha256(buf.Bytes()[0:blockHeaderLen]))

	// Even though this function can't currently fail, it still returns
	// a potential error to help future proof the API should a failure
	// become possible.
	return sha, nil
}

// Deserialize decodes a block header from r into the receiver using a format
// that is suitable for long-term storage such as a database.
func (h *BlockHeader) Deserialize(r io.Reader) error {
	// At the current time, there is no difference between the wire encoding
	// at protocol version 0 and the stable long-term storage format.  As
	// a result, make use of readBlockHeader.
	return readBlockHeader(r, 0, h)
}

// Serialize encodes a block header from r into the receiver using a format
// that is suitable for long-term storage such as a database.
func (h *BlockHeader) Serialize(w io.Writer) error {
	// At the current time, there is no difference between the wire encoding
	// at protocol version 0 and the stable long-term storage format.  As
	// a result, make use of writeBlockHeader.
	return writeBlockHeader(w, 0, h)
}

// NewBlockHeader returns a new BlockHeader using the provided previous block
// hash, merkle root hash, difficulty bits, and nonce used to generate the
// block with defaults for the remaining fields.
func NewBlockHeader(prevHash interfaces.IHash, merkleRootHash interfaces.IHash) *BlockHeader {

	return &BlockHeader{
		PrevBlock:  prevHash,
		MerkleRoot: merkleRootHash,
	}
}

// readBlockHeader reads a bitcoin block header from r.  See Deserialize for
// decoding block headers stored to disk, such as in a database, as opposed to
// decoding from the wire.
func readBlockHeader(r io.Reader, pver uint32, bh *BlockHeader) error {

	err := readElements(r, &bh.ChainID, &bh.MerkleRoot, &bh.PrevBlock, &bh.PrevHash3, &bh.ExchRate,
		&bh.DBHeight, &bh.UTXOCommit, &bh.TransCnt, &bh.BodySize)

	if err != nil {
		return err
	}

	return nil
}

// writeBlockHeader writes a bitcoin block header to w.  See Serialize for
// encoding block headers to be stored to disk, such as in a database, as
// opposed to encoding for the wire.
func writeBlockHeader(w io.Writer, pver uint32, bh *BlockHeader) error {
	err := bh.ChainID.SetBytes(FChainID.Bytes())
	if err != nil {
		return err
	}
	err = writeElements(w, &bh.ChainID, &bh.MerkleRoot, &bh.PrevBlock, &bh.PrevHash3, bh.ExchRate,
		bh.DBHeight, &bh.UTXOCommit, &bh.TransCnt, bh.BodySize)
	if err != nil {
		return err
	}

	return nil
}
