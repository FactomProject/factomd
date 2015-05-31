// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package block

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
    sc "github.com/FactomProject/simplecoin"
)

type ISCBlock interface {
	sc.IBlock
	MarshalTrans() ([]byte, error)
    AddCoinbase(sc.ITransaction) (bool, error)
	AddTransaction(sc.ITransaction) (bool, error)
	CalculateHashes()
	SetDBHeight(uint32)
	GetDBHeight() uint32
	SetExchRate(uint64)
	GetExchRate() uint64
}

// FBlockHeader defines information about a block and is used in the bitcoin
// block (MsgBlock) and headers (MsgHeaders) messages.
//
// https://github.com/FactomProject/FactomDocs/blob/master/factomDataStructureDetails.md#factoid-block
//
type SCBlock struct {
	ISCBlock
	//  ChainID         IHash           // ChainID.  But since this is a constant, we need not actually use space to store it.
	MerkleRoot sc.IHash // Merkle root of the Factoid transactions which accompany this block.
	PrevBlock  sc.IHash // Key Merkle root of previous block.
	PrevHash3  sc.IHash // Sha3 of the previous Factoid Block
	ExchRate   uint64   // Factoshis per Entry Credit
	DBHeight   uint32   // Directory Block height
	UTXOCommit sc.IHash // This field will hold a Merkle root of an array containing all unspent transactions.
	// Transaction count
	// body size
	transactions []sc.ITransaction // List of transactions in this block

}

var _ ISCBlock = (*SCBlock)(nil)

func (b *SCBlock) MarshalTrans() ([]byte, error) {
	var out bytes.Buffer
	for _, trans := range b.transactions {
		data, err := trans.MarshalBinary()
		if err != nil {
			return nil, err
		}
		out.Write(data)
		if err != nil {
			return nil, err
		}
	}
	return out.Bytes(), nil
}

// Write out the block
func (b *SCBlock) MarshalBinary() ([]byte, error) {
	var out bytes.Buffer
	b.CalculateHashes()
	out.Write(sc.FACTOID_CHAINID)
	b.MerkleRoot = new(sc.Hash)
	data, err := b.MerkleRoot.MarshalBinary()
	if err != nil {
		return nil, err
	}
	out.Write(data)
	b.PrevBlock = new(sc.Hash)
	data, err = b.PrevBlock.MarshalBinary()
	if err != nil {
		return nil, err
	}
	out.Write(data)
	b.PrevHash3 = new(sc.Hash)
	data, err = b.PrevHash3.MarshalBinary()
	if err != nil {
		return nil, err
	}
	out.Write(data)
	binary.Write(&out, binary.BigEndian, uint64(b.ExchRate))
	binary.Write(&out, binary.BigEndian, uint32(b.DBHeight))
	b.UTXOCommit = new(sc.Hash)
	data, err = b.UTXOCommit.MarshalBinary()
	if err != nil {
		return nil, err
	}
	out.Write(data)
	binary.Write(&out, binary.BigEndian, uint64(len(b.transactions)))

	transdata, err := b.MarshalTrans()                           // first get trans data
	binary.Write(&out, binary.BigEndian, uint64(len(transdata))) // write out its length
	out.Write(transdata)                                         // write out trans data

	return out.Bytes(), nil
}

// UnmarshalBinary assumes that the Binary is all good.  We do error
// out if there isn't enough data, or the transaction is too large.
func (b *SCBlock) UnmarshalBinaryData(data []byte) ([]byte, error) {

	if bytes.Compare(data[:sc.ADDRESS_LENGTH], sc.FACTOID_CHAINID[:]) != 0 {
		return nil, fmt.Errorf("Block does not begin with the Factoid ChainID")
	}
	data = data[32:]
	b.MerkleRoot = new(sc.Hash)
	data, err := b.MerkleRoot.UnmarshalBinaryData(data)
	if err != nil {
		return nil, err
	}

	b.PrevBlock = new(sc.Hash)
	data, err = b.PrevBlock.UnmarshalBinaryData(data)
	if err != nil {
		return nil, err
	}
	b.PrevHash3 = new(sc.Hash)
	data, err = b.PrevHash3.UnmarshalBinaryData(data)
	if err != nil {
		return nil, err
	}

	b.ExchRate, data = binary.BigEndian.Uint64(data[0:8]), data[8:]
	b.DBHeight, data = binary.BigEndian.Uint32(data[0:4]), data[4:]

	b.UTXOCommit = new(sc.Hash)
	data, err = b.UTXOCommit.UnmarshalBinaryData(data)
	if err != nil {
		return nil, err
	}

	cnt, data := binary.BigEndian.Uint64(data[0:8]), data[8:]
	data = data[8:] // Just skip the size... We don't really need it.

	b.transactions = make([]sc.ITransaction, 0, cnt)
	for i := uint64(0); i < cnt; i++ {
		b.transactions = append(b.transactions, new(sc.Transaction))
	}

	return data, nil
}

func (b *SCBlock) UnmarshalBinary(data []byte) (err error) {
	data, err = b.UnmarshalBinaryData(data)
	return err
}

// Tests if the transaction is equal in all of its structures, and
// in order of the structures.  Largely used to test and debug, but
// generally useful.
func (b1 SCBlock) IsEqual(block sc.IBlock) bool {

	b2, ok := block.(*SCBlock)

	if !ok || // Not the right kind of IBlock
		!b1.MerkleRoot.IsEqual(b2.MerkleRoot) ||
		!b1.PrevBlock.IsEqual(b2.PrevBlock) ||
		!b1.PrevHash3.IsEqual(b2.PrevHash3) ||
		b1.ExchRate != b2.ExchRate ||
		b1.DBHeight != b2.DBHeight ||
		!b1.UTXOCommit.IsEqual(b2.UTXOCommit) {
		return false
	}

	for i, trans := range b1.transactions {
		if !trans.IsEqual(b2.transactions[i]) {
			return false
		}
	}

	return true
}

func (b *SCBlock) CalculateHashes() {
}

func (b *SCBlock) SetDBHeight(dbheight uint32) {
	b.DBHeight = dbheight
}
func (b SCBlock) GetDBHeight() uint32 {
	return b.DBHeight
}
func (b *SCBlock) SetExchRate(rate uint64) {
	b.ExchRate = rate
}
func (b SCBlock) GetExchRate() uint64 {
	return b.ExchRate
}

func (b SCBlock) Validate() (bool, error) {
	for _, trans := range b.transactions {
		valid := trans.Validate()
		if !valid {
			return false, fmt.Errorf("Block contains invalid transactions")
		}
	}

	// Need to check balances are all good.

	// Save what we got for our hashes
	mr := b.MerkleRoot
	pb := b.PrevBlock
	ph := b.PrevHash3

	// Recalculate the hashes
	b.CalculateHashes()

	// Make sure nothing changes.  If something did, this block is bad.
	return mr == b.MerkleRoot && pb == b.PrevBlock && ph == b.PrevHash3, nil
}

// Add the first transaction of a block.  This transaction makes the 
// payout to the servers, so it has no inputs.   This transaction must
// be deterministic so that all servers will know and expect its output.
func (b *SCBlock) AddCoinbase(trans sc.ITransaction) (bool, error) {
    if len(b.transactions)              != 0 ||
       len(trans.GetInputs())           != 0 || 
       len(trans.GetOutECs())           != 0 ||
       len(trans.GetRCDs())             != 0 ||
       len(trans.GetSignatureBlocks())  != 0 {
        return false, fmt.Errorf("Cannot have inputs or EC outputs in the coinbase.")
    }

    // TODO Add check here for the proper payouts.
    
    b.transactions = append(b.transactions, trans)
    return true, nil
}
    

// Add a transaction to the Facoid block. If there is an error,
// then the transaction can be discarded.  If it returns true,
// then the transaction was added, if false it was not.
func (b *SCBlock) AddTransaction(trans sc.ITransaction) (bool, error) {
	// These tests check that the Transaction itself is valid.  If it
	// is not internally valid, it never will be valid.
	ok := trans.Validate()
	if !ok {
		return false, fmt.Errorf("Invalid Transaction")
	}

	// These checks may pass in the future

	// Check against address balances

	b.transactions = append(b.transactions, trans)
	return true, nil
}

// Marshal to text.  Largely a debugging thing.
func (b SCBlock) MarshalText() (text []byte, err error) {
	var out bytes.Buffer

	out.WriteString("Transaction Block\n")
	out.WriteString("  ChainID: ")
	out.WriteString(hex.EncodeToString(sc.FACTOID_CHAINID))
    if b.MerkleRoot == nil { b.MerkleRoot = new (sc.Hash) }
    out.WriteString("\n  MerkleRoot: ")
    out.WriteString(b.MerkleRoot.String())
    if b.PrevBlock == nil { b.PrevBlock = new (sc.Hash) }
    out.WriteString("\n  PrevBlock: ")
	out.WriteString(b.PrevBlock.String())
    if b.PrevHash3 == nil { b.PrevHash3 = new (sc.Hash) }
    out.WriteString("\n  PrevHash3: ")
	out.WriteString(b.PrevHash3.String())
	out.WriteString("\n  ExchRate: ")
	sc.WriteNumber64(&out, b.ExchRate)
	out.WriteString("\n  DBHeight: ")
	sc.WriteNumber32(&out, b.DBHeight)
    if b.UTXOCommit == nil { b.UTXOCommit = new (sc.Hash) }
    out.WriteString("\n  UTXOCommit: ")
	out.WriteString(b.UTXOCommit.String())
	out.WriteString("\n  Number Transactions: ")
	sc.WriteNumber64(&out, uint64(len(b.transactions)))
	transdata, err := b.MarshalTrans()
	if err != nil {
		return out.Bytes(), err
	}
	out.WriteString("\n  Body Size: ")
	sc.WriteNumber64(&out, uint64(len(transdata)))
	out.WriteString("\n")
	for _, trans := range b.transactions {
		txt, err := trans.MarshalText()
		if err != nil {
			return out.Bytes(), err
		}
		out.Write(txt)
	}

	return out.Bytes(), nil
}

/**************************
 * Helper Functions
 **************************/

func NewSCBlock(ExchRate uint64, DBHeight uint32) ISCBlock {
	scb := new(SCBlock)
	scb.ExchRate = ExchRate
	scb.DBHeight = DBHeight
	return scb
}
