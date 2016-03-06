// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package block

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	. "github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// FBlockHeader defines information about a block and is used in the bitcoin
// block (MsgBlock) and headers (MsgHeaders) messages.
//
// https://github.com/FactomProject/FactomDocs/blob/master/factomDataStructureDetails.md#factoid-block
//
type FBlock struct {
	//  ChainID         interfaces.IHash     // ChainID.  But since this is a constant, we need not actually use space to store it.
	BodyMR       interfaces.IHash // Merkle root of the Factoid transactions which accompany this block.
	PrevKeyMR    interfaces.IHash // Key Merkle root of previous block.
	PrevFullHash interfaces.IHash // Sha3 of the previous Factoid Block
	ExchRate     uint64           // Factoshis per Entry Credit
	DBHeight     uint32           // Directory Block height
	// Header Expansion Size  varint
	// Transaction count
	// body size
	Transactions []interfaces.ITransaction // List of transactions in this block

	endOfPeriod [10]int // End of Minute transaction heights.  The mark the height of the first entry of
	// the NEXT period.  This entry may not exist.  The Coinbase transaction is considered
	// to be in the first period.  Factom's periods will initially be a minute long, and
	// there will be 10 of them.  This may change in the future.
}

var _ interfaces.IFBlock = (*FBlock)(nil)
var _ interfaces.Printable = (*FBlock)(nil)
var _ interfaces.BinaryMarshallableAndCopyable = (*FBlock)(nil)
var _ interfaces.DatabaseBlockWithEntries = (*FBlock)(nil)

func (c *FBlock) GetEntryHashes() []interfaces.IHash {
	entries := c.Transactions[:]
	answer := make([]interfaces.IHash, len(entries))
	for i, entry := range entries {
		answer[i] = entry.GetHash()
	}
	return answer
}

func (c *FBlock) New() interfaces.BinaryMarshallableAndCopyable {
	return new(FBlock)
}

func (c *FBlock) DatabasePrimaryIndex() interfaces.IHash {
	return c.GetKeyMR()
}

func (c *FBlock) DatabaseSecondaryIndex() interfaces.IHash {
	return c.GetHash()
}

func (c *FBlock) GetDatabaseHeight() uint32 {
	return c.DBHeight
}

// Return the timestamp of the coinbase transaction
func (b *FBlock) GetCoinbaseTimestamp() int64 {
	if len(b.Transactions) == 0 {
		return -1
	}
	return int64(b.Transactions[0].GetMilliTimestamp())
}

func (b *FBlock) EndOfPeriod(period int) {
	if period == 0 {
		return
	} else {
		period = period - 1 // Make the period zero based.
		b.endOfPeriod[period] = len(b.Transactions)
		for i := period + 1; i < len(b.endOfPeriod); i++ {
			b.endOfPeriod[i] = 0
		}
	}
}

func (b *FBlock) GetTransactions() []interfaces.ITransaction {
	return b.Transactions
}

func (b FBlock) GetNewInstance() interfaces.IFBlock {
	return new(FBlock)
}

func (b *FBlock) GetHash() interfaces.IHash {
	kmr := b.GetKeyMR()

	return kmr
}

func (b *FBlock) MarshalTrans() ([]byte, error) {
	var out bytes.Buffer
	var periodMark = 0
	var i int
	var trans interfaces.ITransaction
	for i, trans = range b.Transactions {

		for periodMark < len(b.endOfPeriod) &&
			b.endOfPeriod[periodMark] > 0 && // Ignore if markers are not set
			i == b.endOfPeriod[periodMark] {

			out.WriteByte(constants.MARKER)
			periodMark++
		}

		data, err := trans.MarshalBinary()
		if err != nil {
			return nil, err
		}
		out.Write(data)
		if err != nil {
			return nil, err
		}
	}
	for periodMark < len(b.endOfPeriod) {
		out.WriteByte(constants.MARKER)
		periodMark++
	}
	return out.Bytes(), nil
}

func (b *FBlock) MarshalHeader() ([]byte, error) {
	var out bytes.Buffer

	b.EndOfPeriod(0) // Clean up end of minute markers, if needed.

	out.Write(constants.FACTOID_CHAINID)

	if b.BodyMR == nil {
		b.BodyMR = new(primitives.Hash)
	}
	data, err := b.BodyMR.MarshalBinary()
	if err != nil {
		return nil, err
	}
	out.Write(data)

	if b.PrevKeyMR == nil {
		b.PrevKeyMR = new(primitives.Hash)
	}
	data, err = b.PrevKeyMR.MarshalBinary()
	if err != nil {
		return nil, err
	}
	out.Write(data)

	if b.PrevFullHash == nil {
		b.PrevFullHash = new(primitives.Hash)
	}
	data, err = b.PrevFullHash.MarshalBinary()
	if err != nil {
		return nil, err
	}
	out.Write(data)

	binary.Write(&out, binary.BigEndian, uint64(b.ExchRate))
	binary.Write(&out, binary.BigEndian, uint32(b.DBHeight))

	primitives.EncodeVarInt(&out, 0) // At this point in time, nothing in the Expansion Header
	// so we just write out a zero.

	binary.Write(&out, binary.BigEndian, uint32(len(b.Transactions)))

	transdata, err := b.MarshalTrans() // first get trans data
	if err != nil {
		return nil, err
	}

	binary.Write(&out, binary.BigEndian, uint32(len(transdata))) // write out its length

	return out.Bytes(), nil
}

// Write out the block
func (b *FBlock) MarshalBinary() ([]byte, error) {
	var out bytes.Buffer

	data, err := b.MarshalHeader()
	if err != nil {
		return nil, err
	}
	out.Write(data)

	transdata, err := b.MarshalTrans() // first get trans data
	if err != nil {
		return nil, err
	}
	out.Write(transdata) // write out trans data

	return out.Bytes(), nil
}

// UnmarshalBinary assumes that the Binary is all good.  We do error
// out if there isn't enough data, or the transaction is too large.
func (b *FBlock) UnmarshalBinaryData(data []byte) (newdata []byte, err error) {

	// To catch memory errors, I capture the panic and turn it into
	// a reported error.
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling transaction: %v", r)
		}
	}()

	// To capture the panic, my code needs to be in a function.  So I'm
	// creating one here, and call it at the end of this function.
	if bytes.Compare(data[:constants.ADDRESS_LENGTH], constants.FACTOID_CHAINID[:]) != 0 {
		return nil, fmt.Errorf("Block does not begin with the Factoid ChainID")
	}
	data = data[32:]

	b.BodyMR = new(primitives.Hash)
	data, err = b.BodyMR.UnmarshalBinaryData(data)
	if err != nil {
		return nil, err
	}

	b.PrevKeyMR = new(primitives.Hash)
	data, err = b.PrevKeyMR.UnmarshalBinaryData(data)
	if err != nil {
		return nil, err
	}

	b.PrevFullHash = new(primitives.Hash)
	data, err = b.PrevFullHash.UnmarshalBinaryData(data)
	if err != nil {
		return nil, err
	}

	b.ExchRate, data = binary.BigEndian.Uint64(data[0:8]), data[8:]
	b.DBHeight, data = binary.BigEndian.Uint32(data[0:4]), data[4:]

	skip, data := primitives.DecodeVarInt(data) // Skip the Expansion Header, if any, since
	data = data[skip:]                          // we don't know what to do with it.

	cnt, data := binary.BigEndian.Uint32(data[0:4]), data[4:]

	data = data[4:] // Just skip the size... We don't really need it.

	b.Transactions = make([]interfaces.ITransaction, cnt, cnt)
	for i, _ := range b.endOfPeriod {
		b.endOfPeriod[i] = 0
	}
	var periodMark = 0
	for i := uint32(0); i < cnt; i++ {

		for data[0] == constants.MARKER {
			b.endOfPeriod[periodMark] = int(i)
			data = data[1:]
			periodMark++
		}

		trans := new(Transaction)
		data, err = trans.UnmarshalBinaryData(data)
		if err != nil {
			return nil, fmt.Errorf("Failed to unmarshal a transaction in block.\n" + err.Error())
		}
		b.Transactions[i] = trans
	}
	for periodMark < len(b.endOfPeriod) {
		data = data[1:]
		periodMark++
	}

	return data, nil

}

func (b *FBlock) UnmarshalBinary(data []byte) (err error) {
	_, err = b.UnmarshalBinaryData(data)
	return err
}

// Tests if the transaction is equal in all of its structures, and
// in order of the structures.  Largely used to test and debug, but
// generally useful.
func (b1 *FBlock) IsEqual(block interfaces.IBlock) []interfaces.IBlock {

	b1.EndOfPeriod(0) // Clean up end of minute markers, if needed.

	b2, ok := block.(*FBlock)

	if !ok || // Not the right kind of interfaces.IBlock
		b1.ExchRate != b2.ExchRate ||
		b1.DBHeight != b2.DBHeight {
		r := make([]interfaces.IBlock, 0, 3)
		return append(r, b1)
	}

	r := b1.BodyMR.IsEqual(b2.BodyMR)
	if r != nil {
		return append(r, b1)
	}
	r = b1.PrevKeyMR.IsEqual(b2.PrevKeyMR)
	if r != nil {
		return append(r, b1)
	}
	r = b1.PrevFullHash.IsEqual(b2.PrevFullHash)
	if r != nil {
		return append(r, b1)
	}

	if b1.endOfPeriod != b2.endOfPeriod {
		return append(r, b1)
	}

	for i, trans := range b1.Transactions {
		r := trans.IsEqual(b2.Transactions[i])
		if r != nil {
			return append(r, b1)
		}
	}

	return nil
}

func (b *FBlock) GetChainID() interfaces.IHash {
	return primitives.NewHash(constants.FACTOID_CHAINID)
}

// Calculates the Key Merkle Root for this block and returns it.
func (b *FBlock) GetKeyMR() interfaces.IHash {

	bodyMR := b.GetBodyMR()

	data, err := b.MarshalHeader()
	if err != nil {
		panic("Failed to create KeyMR: " + err.Error())
	}
	headerHash := primitives.Sha(data)
	cat := append(headerHash.Bytes(), bodyMR.Bytes()...)
	kmr := primitives.Sha(cat)
	return kmr
}

// Calculates the Key Merkle Root for this block and returns it.
func (b *FBlock) GetFullHash() interfaces.IHash {

	ledgerMR := b.GetLedgerMR()

	data, err := b.MarshalHeader()
	if err != nil {
		panic("Failed to create FullHash: " + err.Error())
	}
	headerHash := primitives.Sha(data)
	cat := append(ledgerMR.Bytes(), headerHash.Bytes()...)
	lkmr := primitives.Sha(cat)

	return lkmr
}

// Returns the LedgerMR for this block.
func (b *FBlock) GetLedgerMR() interfaces.IHash {

	b.EndOfPeriod(0) // Clean up end of minute markers, if needed.

	hashes := make([]interfaces.IHash, 0, len(b.Transactions))
	marker := 0
	for i, trans := range b.Transactions {
		for marker < len(b.endOfPeriod) && i != 0 && i == b.endOfPeriod[marker] {
			marker++
			hashes = append(hashes, primitives.Sha(constants.ZERO))
		}
		data, err := trans.MarshalBinarySig()
		hash := primitives.Sha(data)
		if err != nil {
			panic("Failed to get LedgerMR: " + err.Error())
		}
		hashes = append(hashes, hash)
	}

	// Add any lagging markers
	for marker < len(b.endOfPeriod) {
		marker++
		hashes = append(hashes, primitives.Sha(constants.ZERO))
	}
	lmr := primitives.ComputeMerkleRoot(hashes)
	return lmr
}

func (b *FBlock) GetBodyMR() interfaces.IHash {

	b.EndOfPeriod(0) // Clean up end of minute markers, if needed.

	hashes := make([]interfaces.IHash, 0, len(b.Transactions))
	marker := 0
	for i, trans := range b.Transactions {
		for marker < len(b.endOfPeriod) && i != 0 && i == b.endOfPeriod[marker] {
			marker++
			hashes = append(hashes, primitives.Sha(constants.ZERO))
		}
		hashes = append(hashes, trans.GetHash())
	}
	// Add any lagging markers
	for marker < len(b.endOfPeriod) {
		marker++
		hashes = append(hashes, primitives.Sha(constants.ZERO))
	}

	b.BodyMR = primitives.ComputeMerkleRoot(hashes)

	return b.BodyMR
}

func (b *FBlock) GetPrevKeyMR() interfaces.IHash {
	return b.PrevKeyMR
}

func (b *FBlock) SetPrevKeyMR(hash []byte) {
	h := primitives.NewHash(hash)
	b.PrevKeyMR = h
}

func (b *FBlock) GetPrevFullHash() interfaces.IHash {
	return b.PrevFullHash
}

func (b *FBlock) SetPrevFullHash(hash []byte) {
	b.PrevFullHash.SetBytes(hash)
}

func (b *FBlock) CalculateHashes() {
	b.BodyMR = nil
	b.GetBodyMR()
}

func (b *FBlock) SetDBHeight(dbheight uint32) {
	b.DBHeight = dbheight
}

func (b *FBlock) GetDBHeight() uint32 {
	return b.DBHeight
}

func (b *FBlock) SetExchRate(rate uint64) {
	b.ExchRate = rate
}

func (b *FBlock) GetExchRate() uint64 {
	return b.ExchRate
}

func (b FBlock) ValidateTransaction(index int, trans interfaces.ITransaction) error {
	// Calculate the fee due.
	{
		err := trans.Validate(index)
		if err != nil {
			return err
		}
	}

	//Ignore coinbase transaction's signatures
	if len(b.Transactions) > 0 {
		err := trans.ValidateSignatures()
		if err != nil {
			return err
		}
	}

	fee, err := trans.CalculateFee(b.ExchRate)
	if err != nil {
		return err
	}
	tin, err := trans.TotalInputs()
	if err != nil {
		return err
	}
	tout, err := trans.TotalOutputs()
	if err != nil {
		return err
	}
	tec, err := trans.TotalECs()
	if err != nil {
		return err
	}
	sum, err := ValidateAmounts(tout, tec, fee)
	if err != nil {
		return err
	}

	if tin < sum {
		return fmt.Errorf("The inputs %s do not cover the outputs %s,\n"+
			"the Entry Credit outputs %s, and the required fee %s",
			primitives.ConvertDecimalToString(tin),
			primitives.ConvertDecimalToString(tout),
			primitives.ConvertDecimalToString(tec),
			primitives.ConvertDecimalToString(fee))
	}
	return nil
}

func (b FBlock) Validate() error {
	for i, trans := range b.Transactions {
		if err := b.ValidateTransaction(i, trans); err != nil {
			return nil
		}
		if i == 0 {
			if len(trans.GetInputs()) != 0 {
				return fmt.Errorf("Block has a coinbase transaction with inputs")
			}
		} else {
			if len(trans.GetInputs()) == 0 {
				return fmt.Errorf("Block contains transactions without inputs")
			}
		}
	}

	// Need to check balances are all good.

	// Save what we got for our hashes
	mr := b.BodyMR

	// Recalculate the hashes
	b.CalculateHashes()

	// Make sure nothing changes.  If something did, this block is bad.
	if mr != b.BodyMR {
		return fmt.Errorf("This blocks Merkle Root of the transactions does not match the transactions")
	}

	return nil
}

// Add the first transaction of a block.  This transaction makes the
// payout to the servers, so it has no inputs.   This transaction must
// be deterministic so that all servers will know and expect its output.
func (b *FBlock) AddCoinbase(trans interfaces.ITransaction) error {
	b.BodyMR = nil
	if len(b.Transactions) != 0 {
		return fmt.Errorf("The coinbase transaction must be the first transaction")
	}
	if len(trans.GetInputs()) != 0 {
		return fmt.Errorf("The coinbase transaction cannot have any inputs")
	}
	if len(trans.GetECOutputs()) != 0 {
		return fmt.Errorf("The coinbase transaction cannot buy Entry Credits")
	}
	if len(trans.GetRCDs()) != 0 {
		return fmt.Errorf("The coinbase transaction cannot have anyRCD blocks")
	}
	if len(trans.GetSignatureBlocks()) != 0 {
		return fmt.Errorf("The coinbase transaction is not signed")
	}

	// TODO Add check here for the proper payouts.

	b.Transactions = append(b.Transactions, trans)
	return nil
}

// Add the given transaction to this block.  Reports an error if this
// cannot be done, or if the transaction is invalid.
func (b *FBlock) AddTransaction(trans interfaces.ITransaction) error {
	// These tests check that the Transaction itself is valid.  If it
	// is not internally valid, it never will be valid.
	b.BodyMR = nil
	err := b.ValidateTransaction(len(b.Transactions), trans)
	if err != nil {
		return err
	}

	// Check against address balances is done at the Factom level.

	b.Transactions = append(b.Transactions, trans)
	return nil
}

func (b FBlock) String() string {
	txt, err := b.CustomMarshalText()
	if err != nil {
		return err.Error()
	}
	return string(txt)
}

// Marshal to text.  Largely a debugging thing.
func (b FBlock) CustomMarshalText() (text []byte, err error) {
	var out bytes.Buffer

	b.EndOfPeriod(0) // Clean up end of minute markers, if needed.

	out.WriteString("Transaction Block\n")
	out.WriteString("  ChainID:       ")
	out.WriteString(hex.EncodeToString(constants.FACTOID_CHAINID))
	if b.BodyMR == nil {
		b.BodyMR = new(primitives.Hash)
	}
	out.WriteString("\n  BodyMR:        ")
	out.WriteString(b.BodyMR.String())
	if b.PrevKeyMR == nil {
		b.PrevKeyMR = new(primitives.Hash)
	}
	out.WriteString("\n  PrevKeyMR:     ")
	out.WriteString(b.PrevKeyMR.String())
	if b.PrevFullHash == nil {
		b.PrevFullHash = new(primitives.Hash)
	}
	out.WriteString("\n  PrevFullHash:  ")
	out.WriteString(b.PrevFullHash.String())
	out.WriteString("\n  ExchRate:      ")
	primitives.WriteNumber64(&out, b.ExchRate)
	out.WriteString("\n  DBHeight:      ")
	primitives.WriteNumber32(&out, b.DBHeight)
	out.WriteString("\n  Period Marks:  ")
	for _, mark := range b.endOfPeriod {
		out.WriteString(fmt.Sprintf("%d ", mark))
	}
	out.WriteString("\n  #Transactions: ")
	primitives.WriteNumber32(&out, uint32(len(b.Transactions)))
	transdata, err := b.MarshalTrans()
	if err != nil {
		return out.Bytes(), err
	}
	out.WriteString("\n  Body Size:     ")
	primitives.WriteNumber32(&out, uint32(len(transdata)))
	out.WriteString("\n\n")
	markPeriod := 0

	for i, trans := range b.Transactions {

		for markPeriod < 10 && i == b.endOfPeriod[markPeriod] {
			out.WriteString(fmt.Sprintf("\n   End of Minute %d\n\n", markPeriod+1))
			markPeriod++
		}

		txt, err := trans.CustomMarshalText()
		if err != nil {
			return out.Bytes(), err
		}
		out.Write(txt)
	}
	return out.Bytes(), nil
}

func (e *FBlock) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *FBlock) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *FBlock) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

/**************************
 * Helper Functions
 **************************/

func NewFBlock(exchRate uint64, dbHeight uint32) interfaces.IFBlock {
	scb := new(FBlock)
	scb.BodyMR = new(primitives.Hash)
	scb.PrevKeyMR = new(primitives.Hash)
	scb.PrevFullHash = new(primitives.Hash)
	scb.ExchRate = exchRate
	scb.DBHeight = dbHeight
	return scb
}

func NewFBlockFromPreviousBlock(exchangeRate uint64, prev interfaces.IFBlock) interfaces.IFBlock {
	if prev != nil {
		newBlock := NewFBlock(exchangeRate, prev.GetDBHeight()+1)
		newBlock.SetPrevKeyMR(prev.GetKeyMR().Bytes())
		newBlock.SetPrevFullHash(prev.GetFullHash().Bytes())
		return newBlock
	}
	return NewFBlock(exchangeRate, 0)
}
