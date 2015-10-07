// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// Defines the state for factoid.  By using the proper
// interfaces, the functionality of factoid can be imported
// into any framework.
package state

import (
	"bytes"
	"fmt"
	. "github.com/FactomProject/factomd/common/constants"
	. "github.com/FactomProject/factomd/common/factoid"
	. "github.com/FactomProject/factomd/common/factoid/block"
	. "github.com/FactomProject/factomd/common/factoid/block/coinbase"
	. "github.com/FactomProject/factomd/common/interfaces"
	. "github.com/FactomProject/factomd/common/primitives"
	cp "github.com/FactomProject/factomd/controlpanel"
	"time"
)

var _ = time.Sleep
var FACTOID_CHAINID_HASH = NewHash(FACTOID_CHAINID)
var FactoidStateGlobal IFactoidState

type FactoidState struct {
	database        IFDatabase
	factoshisPerEC  uint64
	currentBlock    IFBlock
	dbheight        uint32
	wallet          ISCWallet
	numTransactions int
}

var _ IFactoidState = (*FactoidState)(nil)

func (fs *FactoidState) EndOfPeriod(period int) {
	fs.GetCurrentBlock().EndOfPeriod(period)
}

func (fs *FactoidState) GetWallet() ISCWallet {
	return fs.wallet
}

func (fs *FactoidState) SetWallet(w ISCWallet) {
	fs.wallet = w
}

func (fs *FactoidState) GetCurrentBlock() IFBlock {
	return fs.currentBlock
}

func (fs *FactoidState) GetDBHeight() uint32 {
	return fs.dbheight
}

// When we are playing catchup, adding the transaction block is a pretty
// useful feature.
func (fs *FactoidState) AddTransactionBlock(blk IFBlock) error {

	if err := blk.Validate(); err != nil {
		return err
	}

	transactions := blk.GetTransactions()
	for _, trans := range transactions {
		err := fs.UpdateTransaction(trans)
		if err != nil {
			return err
		}
	}
	fs.currentBlock = blk
	fs.SetFactoshisPerEC(blk.GetExchRate())

	cp.CP.AddUpdate(
		"FAddBlk", // tag
		"status",  // Category
		fmt.Sprintf("Added Factoid Block %d", blk.GetDBHeight()), // Title
		"", // message
		60) // sixty seconds should be enough

	return nil
}

// Checks the transaction timestamp for validity in being included in the current
// No node has any responsiblity to forward on transactions that do not fall within
// the timeframe around a block defined by TRANSACTION_PRIOR_LIMIT and TRANSACTION_POST_LIMIT
func (fs *FactoidState) ValidateTransactionAge(trans ITransaction) error {
	tsblk := fs.GetCurrentBlock().GetCoinbaseTimestamp()
	if tsblk < 0 {
		return fmt.Errorf("Block has no coinbase transaction at this time")
	}

	tstrans := int64(trans.GetMilliTimestamp())

	if tsblk-tstrans > TRANSACTION_PRIOR_LIMIT {
		return fmt.Errorf("Transaction is too old to be included in the current block")
	}

	if tstrans-tsblk > TRANSACTION_POST_LIMIT {
		return fmt.Errorf("Transaction is dated too far in the future to be included in the current block")
	}
	return nil
}

// Only add valid transactions to the current
func (fs *FactoidState) AddTransaction(index int, trans ITransaction) error {
	if err := fs.Validate(index, trans); err != nil {
		return err
	}
	if err := fs.ValidateTransactionAge(trans); err != nil {
		return err
	}
	if err := fs.UpdateTransaction(trans); err != nil {
		return err
	}
	if err := fs.currentBlock.AddTransaction(trans); err != nil {
		return err
	}

	return nil
}

// Assumes validation has already been done.
func (fs *FactoidState) UpdateTransaction(trans ITransaction) error {
	for _, input := range trans.GetInputs() {
		err := fs.UpdateBalance(input.GetAddress(), -int64(input.GetAmount()))
		if err != nil {
			return err
		}
	}
	for _, output := range trans.GetOutputs() {
		err := fs.UpdateBalance(output.GetAddress(), int64(output.GetAmount()))
		if err != nil {
			return err
		}
	}
	for _, ecoutput := range trans.GetECOutputs() {
		err := fs.UpdateECBalance(ecoutput.GetAddress(), int64(ecoutput.GetAmount()))
		if err != nil {
			return err
		}
	}

	fs.numTransactions++
	cp.CP.AddUpdate(
		"transprocessed", // tag
		"status",         // Category
		fmt.Sprintf("Factoid Transactions Processed: %d", fs.numTransactions), // Title
		"", // Message
		0)  // When to expire the message; 0 is never

	return nil
}

func (fs *FactoidState) ProcessEndOfMinute() {
}

// End of Block means packing the current block away, and setting
// up the next
func (fs *FactoidState) ProcessEndOfBlock() {
	var hash, hash2 IHash

	if fs.GetCurrentBlock() == nil {
		panic("Invalid state on initialization")
	}

	hash = fs.currentBlock.GetHash()
	hash2 = fs.currentBlock.GetLedgerKeyMR()

	fs.PutTransactionBlock(hash, fs.currentBlock)
	fs.PutTransactionBlock(FACTOID_CHAINID_HASH, fs.currentBlock)

	fs.dbheight += 1
	fs.currentBlock = NewFBlock(fs.GetFactoshisPerEC(), fs.dbheight)

	t := GetCoinbase(fs.GetTimeMilli())
	err := fs.currentBlock.AddCoinbase(t)
	if err != nil {
		panic(err.Error())
	}
	fs.UpdateTransaction(t)

	if hash != nil {
		fs.currentBlock.SetPrevKeyMR(hash.Bytes())
		fs.currentBlock.SetPrevLedgerKeyMR(hash2.Bytes())
	}

	cp.CP.AddUpdate(
		"blockheight", // tag
		"status",      // Category
		fmt.Sprintf("Directory Block Height: %d", fs.GetDBHeight()), // Title
		"", // Msg
		0)
}

// End of Block means packing the current block away, and setting
// up the next
// this function is to replace the existing function: ProcessEndOfBlock
func (fs *FactoidState) ProcessEndOfBlock2(nextBlkHeight uint32) {
	var hash, hash2 IHash

	if fs.currentBlock != nil { // If no blocks, the current block is nil
		hash = fs.currentBlock.GetHash()
		hash2 = fs.currentBlock.GetLedgerKeyMR()
	}

	fs.currentBlock = NewFBlock(fs.GetFactoshisPerEC(), nextBlkHeight)

	t := GetCoinbase(fs.GetTimeMilli())
	err := fs.currentBlock.AddCoinbase(t)
	if err != nil {
		panic(err.Error())
	}
	fs.UpdateTransaction(t)

	if hash != nil {
		fs.currentBlock.SetPrevKeyMR(hash.Bytes())
		fs.currentBlock.SetPrevLedgerKeyMR(hash2.Bytes())
	}

	cp.CP.AddUpdate(
		"blockheight", // tag
		"status",      // Category
		fmt.Sprintf("Directory Block Height: %d", nextBlkHeight), // Title
		"", // Msg
		0)

}

func (fs *FactoidState) LoadState() error {
	var hashes []IHash
	cblk := fs.GetTransactionBlock(FACTOID_CHAINID_HASH)
	// If there is no head for the Factoids in the database, we have an
	// uninitialized database.  We need to add the Genesis  TODO
	if cblk == nil {
		cp.CP.AddUpdate(
			"Creating Factoid Genesis Block", // tag
			"info", // Category
			"Creating the Factoid Genesis Block", // Title
			"", // Msg
			60) // Expire
		//gb := GetGenesisFBlock(fs.GetTimeMilli(), 1000000,10,200000000000)
		gb := GetGenesisFBlock()
		fs.PutTransactionBlock(gb.GetHash(), gb)
		fs.PutTransactionBlock(FACTOID_CHAINID_HASH, gb)
		err := fs.AddTransactionBlock(gb)
		if err != nil {
			Prtln("Failed to build initial state.\n", err)
			return err
		}
		fs.ProcessEndOfBlock()
		return nil
	}
	blk := cblk
	// First run back from the head back to the genesis block, collecting hashes.
	for {
		if blk == nil {
			return fmt.Errorf("Block not found or not formated properly")

		}
		h := blk.GetHash()
		for _, hash := range hashes {
			if bytes.Compare(hash.Bytes(), h.Bytes()) == 0 {
				return fmt.Errorf("Corrupted database; same hash found twice")
			}
		}
		hashes = append(hashes, h)
		if bytes.Compare(blk.GetPrevKeyMR().Bytes(), ZERO_HASH) == 0 {
			break
		}
		tblk := fs.GetTransactionBlock(blk.GetPrevKeyMR())
		if tblk == nil {
			return fmt.Errorf("Failed to find the block at height: %d", blk.GetDBHeight()-1)
		}
		if !bytes.Equal(tblk.GetHash().Bytes(), blk.GetPrevKeyMR().Bytes()) {
			return fmt.Errorf("Hash Failure!  Database must be rebuilt")
		}

		blk = tblk
		time.Sleep(time.Second / 100)
		cp.CP.AddUpdate(
			"loadState",
			"status", // Category
			"Loading State",
			fmt.Sprintf("Scanning backwards. Block: %d", blk.GetDBHeight()),
			0)
	}

	// Now run forward, and build our accounting
	for i := len(hashes) - 1; i >= 0; i-- {
		blk = fs.GetTransactionBlock(hashes[i])
		if blk == nil {

			return fmt.Errorf("Should never happen.  Block not found in the Database\n"+
				"No block found for: %s", hashes[i].String())

		}

		err := fs.AddTransactionBlock(blk) // updates accounting for this block
		if err != nil {
			Prtln("Failed to rebuild state.\n", err)
			return err
		}
		time.Sleep(time.Second / 100)
		cp.CP.AddUpdate(
			"loadState",
			"status", // Category
			"Loading State",
			fmt.Sprintf("Loading and Processing. Block: %d", blk.GetDBHeight()),
			0)
	}

	fs.dbheight = blk.GetDBHeight()
	fs.ProcessEndOfBlock()
	return nil
}

// Returns an error message about what is wrong with the transaction if it is
// invalid, otherwise you are good to go.
func (fs *FactoidState) Validate(index int, trans ITransaction) error {
	err := fs.currentBlock.ValidateTransaction(index, trans)
	if err != nil {
		return err
	}

	var sums = make(map[[32]byte]uint64, 10)  // Look at the sum of an address's inputs
	for _, input := range trans.GetInputs() { //    to a transaction.
		bal, err := ValidateAmounts(sums[input.GetAddress().Fixed()], input.GetAmount())
		if err != nil {
			return err
		}
		if bal > fs.GetBalance(input.GetAddress()) {
			return fmt.Errorf("Not enough funds in input addresses for the transaction")
		}
		sums[input.GetAddress().Fixed()] = bal
	}
	return nil
}

func (fs *FactoidState) GetFactoshisPerEC() uint64 {
	return fs.factoshisPerEC
}

func (fs *FactoidState) SetFactoshisPerEC(factoshisPerEC uint64) {
	fs.factoshisPerEC = factoshisPerEC
}

func (fs *FactoidState) PutTransactionBlock(hash IHash, trans IFBlock) {
	fs.database.Put(DB_FACTOID_BLOCKS, hash, trans)
}

func (fs *FactoidState) GetTransactionBlock(hash IHash) IFBlock {
	transblk := fs.database.Get(DB_FACTOID_BLOCKS, hash)
	if transblk == nil {
		return nil
	}
	return transblk.(IFBlock)
}

func (fs *FactoidState) GetTimeMilli() uint64 {
	return uint64(time.Now().UnixNano()) / 1000000 // 10^-9 >> 10^-3
}

func (fs *FactoidState) GetTime() uint64 {
	return uint64(time.Now().Unix())
}

func (fs *FactoidState) SetDB(database IFDatabase) {
	fs.database = database
}

func (fs *FactoidState) GetDB() IFDatabase {
	return fs.database
}

// Any address that is not defined has a zero balance.
func (fs *FactoidState) GetBalance(address IAddress) uint64 {
	balance := uint64(0)
	b := fs.database.GetRaw([]byte(DB_F_BALANCES), address.Bytes())
	if b != nil {
		balance = b.(*FSbalance).number
	}
	return balance
}

// Any address that is not defined has a zero balance.
func (fs *FactoidState) GetECBalance(address IAddress) uint64 {
	balance := uint64(0)
	b := fs.database.GetRaw([]byte(DB_EC_BALANCES), address.Bytes())
	if b != nil {
		balance = b.(*FSbalance).number
	}
	return balance
}

// Update balance throws an error if your update will drive the balance negative.
func (fs *FactoidState) UpdateBalance(address IAddress, amount int64) error {
	nbalance := int64(fs.GetBalance(address)) + amount
	if nbalance < 0 {
		return fmt.Errorf("The update to this address would drive the balance negative.")
	}
	balance := uint64(nbalance)
	fs.database.PutRaw([]byte(DB_F_BALANCES), address.Bytes(), &FSbalance{number: balance})

	return nil
}

// Update ec balance throws an error if your update will drive the balance negative.
func (fs *FactoidState) UpdateECBalance(address IAddress, amount int64) error {
	nbalance := int64(fs.GetECBalance(address)) + amount
	if nbalance < 0 {
		return fmt.Errorf("The update to this Entry Credit address would drive the balance negative.")
	}
	balance := uint64(nbalance)
	fs.database.PutRaw([]byte(DB_EC_BALANCES), address.Bytes(), &FSbalance{number: balance})

	return nil
}

// Add to Entry Credit Balance.  Note Entry Credit balances are maintained
// as entry credits, not Factoids.  But adding is done in Factoids, using
// done in Entry Credits. Using lowers the Entry Credit Balance.
func (fs *FactoidState) AddToECBalance(address IAddress, amount uint64) error {
	ecs := amount / fs.GetFactoshisPerEC()
	balance := fs.GetECBalance(address) + ecs
	fs.database.PutRaw([]byte(DB_EC_BALANCES), address.Bytes(), &FSbalance{number: balance})
	return nil
}

// Use Entry Credits.  Note Entry Credit balances are maintained
// as entry credits, not Factoids.  But adding is done in Factoids, using
// done in Entry Credits.  Using lowers the Entry Credit Balance.
func (fs *FactoidState) UseECs(address IAddress, amount uint64) error {
	balance := fs.GetECBalance(address) - amount
	if balance < 0 {
		return fmt.Errorf("Overdraft of Entry Credits attempted.")
	}
	fs.database.PutRaw([]byte(DB_EC_BALANCES), address.Bytes(), &FSbalance{number: balance})
	return nil
}
