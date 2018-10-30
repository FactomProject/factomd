// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/FactomProject/factomd/common/constants"
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
	BodyMR          interfaces.IHash `json:"bodymr"`          // Merkle root of the Factoid transactions which accompany this block.
	PrevKeyMR       interfaces.IHash `json:"prevkeymr"`       // Key Merkle root of previous block.
	PrevLedgerKeyMR interfaces.IHash `json:"prevledgerkeymr"` // Sha3 of the previous Factoid Block
	ExchRate        uint64           `json:"exchrate"`        // Factoshis per Entry Credit
	DBHeight        uint32           `json:"dbheight"`        // Directory Block height
	// Header Expansion Size  varint
	// Transaction count
	// body size
	Transactions []interfaces.ITransaction `json:"transactions"` // List of transactions in this block

	endOfPeriod [10]int // End of Minute transaction heights.  The mark the height of the first entry of
	// the NEXT period.  This entry may not exist.  The Coinbase transaction is considered
	// to be in the first period.  Factom's periods will initially be a minute long, and
	// there will be 10 of them.  This may change in the future.
}

var _ interfaces.IFBlock = (*FBlock)(nil)
var _ interfaces.Printable = (*FBlock)(nil)
var _ interfaces.BinaryMarshallableAndCopyable = (*FBlock)(nil)
var _ interfaces.DatabaseBlockWithEntries = (*FBlock)(nil)

func (a *FBlock) Init() {
	if a.BodyMR == nil {
		a.BodyMR = primitives.NewZeroHash()
	}
	if a.PrevKeyMR == nil {
		a.PrevKeyMR = primitives.NewZeroHash()
	}
	if a.PrevLedgerKeyMR == nil {
		a.PrevLedgerKeyMR = primitives.NewZeroHash()
	}
}

func (a *FBlock) IsSameAs(b interfaces.IFBlock) bool {
	return true
}

func (c *FBlock) GetEntryHashes() []interfaces.IHash {
	entries := c.Transactions[:]
	answer := make([]interfaces.IHash, len(entries))
	for i, entry := range entries {
		answer[i] = entry.GetHash()
	}
	return answer
}

func (c *FBlock) GetTransactionByHash(hash interfaces.IHash) interfaces.ITransaction {
	if hash == nil {
		return nil
	}

	txs := c.GetTransactions()
	for _, tx := range txs {
		if hash.IsSameAs(tx.GetHash()) {
			return tx
		}
		if hash.IsSameAs(tx.GetSigHash()) {
			return tx
		}
	}
	return nil
}

func (c *FBlock) GetEntrySigHashes() []interfaces.IHash {
	entries := c.Transactions[:]
	answer := make([]interfaces.IHash, len(entries))
	for i, entry := range entries {
		answer[i] = entry.GetSigHash()
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
	return c.GetLedgerKeyMR()
}

func (c *FBlock) GetDatabaseHeight() uint32 {
	return c.DBHeight
}

// Return the timestamp of the coinbase transaction
func (b *FBlock) GetCoinbaseTimestamp() interfaces.Timestamp {
	if len(b.Transactions) == 0 {
		return nil
	}
	return b.Transactions[0].GetTimestamp()
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

func (b *FBlock) GetEndOfPeriod() [10]int {
	return b.endOfPeriod
}

func (b *FBlock) MarshalTrans() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "FBlock.MarshalTrans err:%v", *pe)
		}
	}(&err)
	var out primitives.Buffer
	var periodMark = 0
	var i int
	var trans interfaces.ITransaction

	// 	for _, v := range b.GetEndOfPeriod() {
	// 		if v == 0 {
	// 			return nil, fmt.Errorf("Factoid Block is incomplete.  Missing EOM markers detected: %v",b.endOfPeriod)
	// 		}
	// 	}

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
	return out.DeepCopyBytes(), nil
}

func (b *FBlock) MarshalHeader() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "FBlock.MarshalHeader err:%v", *pe)
		}
	}(&err)
	var out primitives.Buffer

	out.Write(constants.FACTOID_CHAINID)

	b.BodyMR = b.GetBodyMR()
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

	if b.PrevLedgerKeyMR == nil {
		b.PrevLedgerKeyMR = new(primitives.Hash)
	}
	data, err = b.PrevLedgerKeyMR.MarshalBinary()
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

	h := out.DeepCopyBytes()
	return h, nil
}

// Write out the block
func (b *FBlock) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "FBlock.MarshalBinary err:%v", *pe)
		}
	}(&err)
	b.Init()
	var out primitives.Buffer

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

	return out.DeepCopyBytes(), nil
}

func UnmarshalFBlock(data []byte) (interfaces.IFBlock, error) {
	block := new(FBlock)

	err := block.UnmarshalBinary(data)
	if err != nil {
		return nil, err
	}

	return block, nil
}

// UnmarshalBinary assumes that the Binary is all good.  We do error
// out if there isn't enough data, or the transaction is too large.
func (b *FBlock) UnmarshalBinaryData(data []byte) ([]byte, error) {
	b.Init()
	buf := primitives.NewBuffer(data)
	h := primitives.NewZeroHash()
	err := buf.PopBinaryMarshallable(h)
	if err != nil {
		return nil, err
	}
	if h.String() != "000000000000000000000000000000000000000000000000000000000000000f" {
		return nil, fmt.Errorf("Block does not begin with the Factoid ChainID")
	}

	err = buf.PopBinaryMarshallable(b.BodyMR)
	if err != nil {
		return nil, err
	}
	err = buf.PopBinaryMarshallable(b.PrevKeyMR)
	if err != nil {
		return nil, err
	}
	err = buf.PopBinaryMarshallable(b.PrevLedgerKeyMR)
	if err != nil {
		return nil, err
	}

	b.ExchRate, err = buf.PopUInt64()
	if err != nil {
		return nil, err
	}
	b.DBHeight, err = buf.PopUInt32()
	if err != nil {
		return nil, err
		return nil, err
	}

	// Skip the Expansion Header, if any, since
	// we don't know what to do with it.
	skip, err := buf.PopVarInt()
	if err != nil {
		return nil, err
	}
	_, err = buf.PopLen(int(skip))
	if err != nil {
		return nil, err
	}

	txCount, err := buf.PopUInt32()
	if err != nil {
		return nil, err
	}
	if txCount > 1000 {
		// TODO: replace this message with a proper error
		return nil, fmt.Errorf("Error: fblock.UnmarshalBinary: transaction count %d too high (uint underflow?)", txCount)
	}

	// Just skip the size... We don't really need it.
	_, err = buf.PopUInt32()
	if err != nil {
		return nil, err
	}

	b.Transactions = make([]interfaces.ITransaction, int(txCount), int(txCount))
	for i, _ := range b.endOfPeriod {
		b.endOfPeriod[i] = 0
	}
	var periodMark = 0

	for i := uint32(0); i < txCount; i++ {
		by, err := buf.PeekByte()
		if err != nil {
			return nil, err
		}
		for by == constants.MARKER {
			_, err = buf.PopByte()
			if err != nil {
				return nil, err
			}
			b.endOfPeriod[periodMark] = int(i)
			periodMark++

			by, err = buf.PeekByte()
			if err != nil {
				return nil, err
			}
		}

		trans := new(Transaction)
		err = buf.PopBinaryMarshallable(trans)
		if err != nil {
			return nil, err
		}
		if err != nil {
			return nil, fmt.Errorf("Failed to unmarshal a transaction in block.\n" + err.Error())
		}
		b.Transactions[i] = trans
	}
	for periodMark < len(b.endOfPeriod) {
		_, err = buf.PopByte()
		if err != nil {
			return nil, err
		}
		b.endOfPeriod[periodMark] = int(txCount)
		periodMark++
	}

	return buf.DeepCopyBytes(), nil
}

func (b *FBlock) UnmarshalBinary(data []byte) (err error) {
	_, err = b.UnmarshalBinaryData(data)
	return err
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

func (b *FBlock) GetHash() interfaces.IHash {
	return b.GetLedgerKeyMR()
}

func (b *FBlock) GetLedgerKeyMR() interfaces.IHash {
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
	hashes := make([]interfaces.IHash, 0, len(b.Transactions))
	marker := 0
	for i, trans := range b.Transactions {
		for marker < len(b.endOfPeriod) && i != 0 && i == b.endOfPeriod[marker] {
			marker++
			hashes = append(hashes, primitives.Sha(constants.ZERO))
		}
		hashes = append(hashes, trans.GetSigHash())
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

func (b *FBlock) SetPrevKeyMR(hash interfaces.IHash) {
	b.PrevKeyMR = hash
}

func (b *FBlock) GetPrevLedgerKeyMR() interfaces.IHash {
	return b.PrevLedgerKeyMR
}

func (b *FBlock) SetPrevLedgerKeyMR(hash interfaces.IHash) {
	b.PrevLedgerKeyMR = hash
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
	var out primitives.Buffer

	out.WriteString("Transaction Block\n")
	out.WriteString("  ChainID:       ")
	out.WriteString(hex.EncodeToString(constants.FACTOID_CHAINID))
	keyMR := b.GetKeyMR()
	out.WriteString("\n  KeyMR (NM):    ")
	out.WriteString(keyMR.String())

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
	if b.PrevLedgerKeyMR == nil {
		b.PrevLedgerKeyMR = new(primitives.Hash)
	}
	out.WriteString("\n  PrevLedgerKeyMR:  ")
	out.WriteString(b.PrevLedgerKeyMR.String())
	out.WriteString("\n  ExchRate:      ")
	out.WriteString(fmt.Sprintf("%d.%08d fct, %d factoshis", b.ExchRate/100000000, b.ExchRate%100000000, b.ExchRate))
	out.WriteString(fmt.Sprintf("\n  DBHeight:      %v", b.DBHeight))
	out.WriteString("\n  Period Marks:  ")
	for _, mark := range b.endOfPeriod {
		out.WriteString(fmt.Sprintf("%d ", mark))
	}
	out.WriteString("\n  #Transactions: ")
	primitives.WriteNumber32(&out, uint32(len(b.Transactions)))
	transdata, err := b.MarshalTrans()
	if err != nil {
		return nil, err
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
			return nil, err
		}
		out.Write(txt)
	}
	return out.DeepCopyBytes(), nil
}

func (e *FBlock) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *FBlock) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

type ExpandedFBlock FBlock

func (e FBlock) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ExpandedFBlock
		ChainID     string `json:"chainid"`
		KeyMR       string `json:"keymr"`
		LedgerKeyMR string `json:"ledgerkeymr"`
	}{
		ExpandedFBlock: ExpandedFBlock(e),
		ChainID:        "000000000000000000000000000000000000000000000000000000000000000f",
		KeyMR:          e.GetKeyMR().String(),
		LedgerKeyMR:    e.GetLedgerKeyMR().String(),
	})
}

/**************************
 * Helper Functions
 **************************/

func NewFBlock(prev interfaces.IFBlock) interfaces.IFBlock {
	scb := new(FBlock)
	scb.BodyMR = new(primitives.Hash)
	if prev != nil {
		scb.PrevKeyMR = prev.GetKeyMR()
		scb.PrevLedgerKeyMR = prev.GetLedgerKeyMR()
		scb.ExchRate = prev.GetExchRate()
		scb.DBHeight = prev.GetDBHeight() + 1
	} else {
		scb.PrevKeyMR = primitives.NewZeroHash()
		scb.PrevLedgerKeyMR = primitives.NewZeroHash()
		scb.ExchRate = 1
		scb.DBHeight = 0
	}
	return scb
}

func CheckBlockPairIntegrity(block interfaces.IFBlock, prev interfaces.IFBlock) error {
	if block == nil {
		return fmt.Errorf("No block specified")
	}

	if prev == nil {
		if block.GetPrevKeyMR().IsZero() == false {
			return fmt.Errorf("Invalid PrevKeyMR")
		}
		if block.GetPrevLedgerKeyMR().IsZero() == false {
			return fmt.Errorf("Invalid PrevLedgerKeyMR")
		}
		if block.GetDBHeight() != 0 {
			return fmt.Errorf("Invalid DBHeight")
		}
	} else {
		if block.GetPrevKeyMR().IsSameAs(prev.GetKeyMR()) == false {
			return fmt.Errorf("Invalid PrevKeyMR")
		}
		if block.GetPrevLedgerKeyMR().IsSameAs(prev.GetLedgerKeyMR()) == false {
			return fmt.Errorf("Invalid PrevLedgerKeyMR")
		}
		if block.GetDBHeight() != (prev.GetDBHeight() + 1) {
			return fmt.Errorf("Invalid DBHeight")
		}
	}

	return nil
}
