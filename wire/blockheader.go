// Copyright (c) 2013-2015 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"bytes"
	"io"
)

// BlockVersion is the current latest supported block version.
const BlockVersion = 2

// Version 4 bytes + Timestamp 4 bytes + Bits 4 bytes + Nonce 4 bytes +
// PrevBlock and MerkleRoot hashes.
const MaxBlockHeaderPayload = 16 + (HashSize * 2)

// BlockHeader defines information about a block and is used in the bitcoin
// block (MsgBlock) and headers (MsgHeaders) messages.
type BlockHeader struct {

	// Hash of the previous block in the block chain.
	PrevBlock ShaHash

	// Merkle tree reference to hash of all transactions for the block.
	MerkleRoot ShaHash

	BodyMR    ShaHash
	PrevKeyMR ShaHash

	PrevHash3 Sha3Hash
}

// blockHeaderLen is a constant that represents the number of bytes for a block
// header.
const blockHeaderLen = 80

// BlockSha computes the block identifier hash for the given block header.
func (h *BlockHeader) BlockSha() (ShaHash, error) {
	// Encode the header and run double sha256 everything prior to the
	// number of transactions.  Ignore the error returns since there is no
	// way the encode could fail except being out of memory which would
	// cause a run-time panic.  Also, SetBytes can't fail here due to the
	// fact DoubleSha256 always returns a []byte of the right size
	// regardless of input.
	var buf bytes.Buffer
	var sha ShaHash
	_ = writeBlockHeader(&buf, 0, h)
	_ = sha.SetBytes(DoubleSha256(buf.Bytes()[0:blockHeaderLen]))

	// Even though this function can't currently fail, it still returns
	// a potential error to help future proof the API should a failure
	// become possible.
	return sha, nil
}

// Deserialize decodes a block header from r into the receiver using a format
// that is suitable for long-term storage such as a database while respecting
// the Version field.
func (h *BlockHeader) Deserialize(r io.Reader) error {
	// At the current time, there is no difference between the wire encoding
	// at protocol version 0 and the stable long-term storage format.  As
	// a result, make use of readBlockHeader.
	return readBlockHeader(r, 0, h)
}

// Serialize encodes a block header from r into the receiver using a format
// that is suitable for long-term storage such as a database while respecting
// the Version field.
func (h *BlockHeader) Serialize(w io.Writer) error {
	// At the current time, there is no difference between the wire encoding
	// at protocol version 0 and the stable long-term storage format.  As
	// a result, make use of writeBlockHeader.
	return writeBlockHeader(w, 0, h)
}

// NewBlockHeader returns a new BlockHeader using the provided previous block
// hash, merkle root hash, difficulty bits, and nonce used to generate the
// block with defaults for the remaining fields.
func NewBlockHeader(prevHash *ShaHash, merkleRootHash *ShaHash, prevHash3 *Sha3Hash) *BlockHeader {

	return &BlockHeader{
		PrevBlock:  *prevHash,
        PrevHash3:  *prevHash3,
        MerkleRoot: *merkleRootHash,

	}
}

// readBlockHeader reads a bitcoin block header from r.  See Deserialize for
// decoding block headers stored to disk, such as in a database, as opposed to
// decoding from the wire.
func readBlockHeader(r io.Reader, pver uint32, bh *BlockHeader) error {
	err := readElements(r, &bh.PrevBlock, &bh.MerkleRoot, &bh.PrevHash3)
	
    if err != nil {
		return err
	}

	return nil
}

// writeBlockHeader writes a bitcoin block header to w.  See Serialize for
// encoding block headers to be stored to disk, such as in a database, as
// opposed to encoding for the wire.
func writeBlockHeader(w io.Writer, pver uint32, bh *BlockHeader) error {
	err := writeElements(w, &bh.PrevBlock, &bh.MerkleRoot, &bh.PrevHash3)

    if err != nil {
		return err
	}

	return nil
}
