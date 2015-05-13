// Copyright (c) 2013-2015 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"bytes"
	"io"
    "github.com/FactomProject/FactomCode/common"
)



// BlockVersion is the current latest supported block version.
const BlockVersion = 2

// The MerkleRoot (32 bytes) the Previous Merkle Root (32 bytes) and the paranoid hash (32 bytes)
const MaxBlockHeaderPayload =  (HashSize * 3)

// BlockHeader defines information about a block and is used in the bitcoin
// block (MsgBlock) and headers (MsgHeaders) messages.
type BlockHeader struct {

                              // ChainID     ShaHash  unneeded by our logic in process
    // Hash of the previous block in the block chain.
    MerkleRoot  ShaHash       // BodyMR elsewhere in the Factom Code for other blocks.
                              //   This is the Merkle Root of all the transactions in the body.
    PrevBlock   ShaHash       // Key Merkle root of previous block.
    PrevHash3   Sha3Hash
                              // ExchRate    uint64   provided over a chanel... part of wire format
                              // DBHeight    uint32   provided over a chanel... part of wire format
                              // UTXOCommit  [32]byte computed when we close the block.
                              // TransCnt    uint64   Is only used over the wire
                              // BodySize    uint64   Is only used over the wire.
}

var (
    // Shared constants
    FChainID          *common.Hash
    CreditsPerChain    int32    
    
    // BTCD State Variables
    FactoshisPerCredit uint64     
)

// blockHeaderLen is a constant that represents the number of bytes for a block
// header.
const blockHeaderLen = 96

// Factom Constants for BTCD and Factom
//
func Init () {
    barray := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
                     0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0F}
    FChainID = new(common.Hash)
    FChainID.SetBytes(barray)
    
    CreditsPerChain = 10            // Entry Credits to create a chain
    
    // Shouldn't set this, but we are for now.
    FactoshisPerCredit = 666667     // .001 / .15 * 100000000 (assuming a Factoid is .15 cents, entry credit = .1 cents    

    
}



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
    
    var chainID     ShaHash
    var exchRate    uint64
    var utxoCommit  [32]byte
    var transCnt    uint64
    var bodySize    uint64
    
    err := readElements(r, &chainID, &bh.MerkleRoot, &bh.PrevBlock,  &bh.PrevHash3,
        &exchRate, &utxoCommit, &transCnt, &bodySize)
    
    if err != nil {
		return err
	}

	return nil
}

// writeBlockHeader writes a bitcoin block header to w.  See Serialize for
// encoding block headers to be stored to disk, such as in a database, as
// opposed to encoding for the wire.
func writeBlockHeader(w io.Writer, pver uint32, bh *BlockHeader) error {
	
    var utxoCommit [32]byte
    var transCnt    uint64
    var bodySize    uint64
    
    err := writeElements(w, FChainID,  &bh.MerkleRoot, &bh.PrevBlock, &bh.PrevHash3,
                         &FactoshisPerCredit, &utxoCommit, &transCnt, &bodySize) 
    

    if err != nil {
		return err
	}

	return nil
}
