// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package block

import sc "github.com/FactomProject/simplecoin"
import (
  	"bytes"
 	"encoding/binary"
// 	"encoding/hex"
 	"fmt" 
)

type ISCBlock interface {
    sc.IBlock
    AddTransaction(sc.ITransaction)
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
    MerkleRoot      sc.IHash        // Merkle root of the Factoid transactions which accompany this block.
    PrevBlock       sc.IHash        // Key Merkle root of previous block.
    PrevHash3       sc.IHash        // Sha3 of the previous Factoid Block
    ExchRate        uint64          // Factoshis per Entry Credit
    DBHeight        uint32          // Directory Block height
    UTXOCommit      sc.IHash        // This field will hold a Merkle root of an array containing all unspent transactions.    
    transactions    []sc.ITransaction  // List of transactions in this block
}

var _ ISCBlock = (*SCBlock)(nil)

// This is what Gets Signed.  Yet signature blocks are part of the transaction.
// We don't include them here, and tack them on later.
func (b *SCBlock) MarshalBinary() ([]byte, error) {
    var out bytes.Buffer
    b.CalculateHashes()
    out.Write(sc.FACTOID_CHAINID)
    data,err := b.MerkleRoot.MarshalBinary(); if err != nil {return nil,err}; out.Write(data)
    data,err  = b.MerkleRoot.MarshalBinary(); if err != nil {return nil,err}; out.Write(data)
    data,err  = b.MerkleRoot.MarshalBinary(); if err != nil {return nil,err}; out.Write(data)
    binary.Write(&out, binary.BigEndian, uint64(b.ExchRate)) 
    binary.Write(&out, binary.BigEndian, uint32(b.DBHeight))
    data,err  = b.MerkleRoot.MarshalBinary(); if err != nil {return nil,err}; out.Write(data)
    binary.Write(&out, binary.BigEndian, uint64(len(b.transactions))) 
    
    var out2 bytes.Buffer
    for _,trans := range b.transactions {
        data,err := trans.MarshalBinary(); if err != nil {return nil,err}; out.Write(data)
    }
    
    length := uint64(out.Len()+8+out2.Len())
    binary.Write(&out, binary.BigEndian, length)
    
    out.Write(out2.Bytes())
    
    return out.Bytes(),nil
}

// UnmarshalBinary assumes that the Binary is all good.  We do error
// out if there isn't enough data, or the transaction is too large.
func (b *SCBlock) UnmarshalBinaryData(data []byte) ( []byte,  error) {
    
    if bytes.Compare(data[:sc.ADDRESS_LENGTH],sc.FACTOID_CHAINID[:]) != 0 {
        return nil, fmt.Errorf("Block does not begin with the Factoid ChainID")
    }
    data = data[32:]
    b.MerkleRoot = new(sc.Hash)
    data, err := b.MerkleRoot.UnmarshalBinaryData(data)
    if err != nil { return nil, err }
    
    b.PrevBlock = new(sc.Hash)
    data, err = b.PrevBlock.UnmarshalBinaryData(data)
    if err != nil { return nil, err }
    b.PrevHash3 = new(sc.Hash)
    data, err = b.PrevHash3.UnmarshalBinaryData(data)
    if err != nil { return nil, err }
    
    b.ExchRate, data = binary.BigEndian.Uint64(data[0:8]), data[8:]
    b.DBHeight, data = binary.BigEndian.Uint32(data[0:4]), data[4:]
    
    b.UTXOCommit = new(sc.Hash)
    data, err = b.UTXOCommit.UnmarshalBinaryData(data)
    if err != nil { return nil, err }

    cnt, data := binary.BigEndian.Uint64(data[0:8]), data[8:]
    data = data[8:] // Just skip the size... We don't really need it.
    
    b.transactions = make([]sc.ITransaction,0,cnt)
    for i:=uint64(0);i<cnt;i++ {
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
        !b1.UTXOCommit.IsEqual(b2.UTXOCommit){
            return false
    }
    
    for i,trans := range b1.transactions {
        if !trans.IsEqual(b2.transactions[i]) {
            return false
        }
    }
        
    return true
}

func (b SCBlock) CalculateHashes() {
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
    var allvalid bool = true
    for _,trans := range b.transactions {
        valid, err := trans.Validate(b.ExchRate)
        if err != nil { return false, err }
        allvalid = allvalid && valid
    }
    
    // Save what we got for our hashes
    mr := b.MerkleRoot
    pb := b.PrevBlock
    ph := b.PrevHash3
    
    // Recalculate the hashes
    b.CalculateHashes()
    
    // Make sure nothing changes.  If something did, this block is bad.
    return  mr == b.MerkleRoot && pb == b.PrevBlock  && ph == b.PrevHash3, nil
}


// Helper function for building transactions.  Add an input to
// the transaction.  I'm guessing 5 inputs is about all anyone
// will need, so I'll default to 5.  Of course, go will grow
// past that if needed.
func (b *SCBlock) AddTransaction( input IAddress,amount uint64) {
    if t.inputs == nil {
        t.inputs = make([]IInAddress, 0, 5)
    }
    out := NewInAddress(input, amount)
    t.inputs = append(t.inputs, out)
}

/*
// Helper function for building transactions.  Add an output to
// the transaction.  I'm guessing 5 outputs is about all anyone
// will need, so I'll default to 5.  Of course, go will grow
// past that if needed.
func (t *Transaction) AddOutput(output IAddress,amount uint64) {
    if t.outputs == nil {
        t.outputs = make([]IOutAddress, 0, 5)
    }
    out := NewOutAddress(output, amount)
    t.outputs = append(t.outputs, out)
    
}

// Add a EntryCredit output.  Validating this is going to require
// access to the exchange rate.  This is literally how many entry
// credits are being added to the specified Entry Credit address.
func (t *Transaction) AddECOutput( ecoutput IAddress, amount uint64) {
    if t.outECs == nil {
        t.outECs = make([]IOutECAddress, 0, 5)
    }
    out := NewOutECAddress(ecoutput, amount)
    t.outECs = append(t.outECs, out)
    
}

// Marshal to text.  Largely a debugging thing.
func (t Transaction) MarshalText() (text []byte, err error) {
    var out bytes.Buffer
    
    out.WriteString("locktime")
    WriteNumber64(&out, uint64(t.lockTime))
    out.WriteString("in  ")
    WriteNumber16(&out, uint16(len(t.inputs)))
    out.WriteString("\nout ")
    WriteNumber16(&out, uint16(len(t.outputs)))
    out.WriteString("\nec  ")
    WriteNumber16(&out, uint16(len(t.outECs)))
    out.WriteString("\n")
    
    for _, address := range t.inputs {
        text, _ := address.MarshalText()
        out.Write(text)
    }
    for _, address := range t.outputs {
        text, _ := address.MarshalText()
        out.Write(text)
    }
    for _, ecaddress := range t.outECs {
        text, _ := ecaddress.MarshalText()
        out.Write(text)
    }
    for _, rcd := range t.rcds {
        text, err = rcd.MarshalText()
        if err != nil {
            return nil, err
        }
        out.Write(text)
    }
    for i := 0; i < len(t.inputs); i++ {
        if len(t.sigBlocks) < i {
            t.sigBlocks = append(t.sigBlocks, new(SignatureBlock))
        }
        text, err := t.sigBlocks[i].MarshalText()
        if err != nil {
            return nil, err
        }
        out.Write(text)
    }
    
    return out.Bytes(), nil
}

// Helper Function.  This simply adds an Authorization to a
// transaction.  DOES NO VALIDATION.  Not the job of construction.
// That's why we have a validation call.
func (t *Transaction) AddAuthorization(auth IRCD) {
    if t.rcds == nil {
        t.rcds = make([]IRCD, 0, 5)
    }
    t.rcds = append(t.rcds, auth)
}

*/