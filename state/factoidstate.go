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
	fct "github.com/FactomProject/factoid"
	"github.com/FactomProject/factoid/block"
	db "github.com/FactomProject/factoid/database"
	"github.com/FactomProject/factoid/wallet"
	"time"
)

var _ = time.Sleep

type IFactoidState interface {
	// Set the database for the Coin State.  This is where
	// we manage the balances for transactions.  We also look
	// for previous blocks here.
	SetDB(db.IFDatabase)
	GetDB() db.IFDatabase

	// Load the address state of Factoids
	LoadState() error

	// Get the wallet used to help manage the Factoid State in
	// some applications.
	GetWallet() wallet.ISCWallet
	SetWallet(wallet.ISCWallet)

	// The Exchange Rate for Entry Credits in Factoshis per
	// Entry Credits
	GetFactoshisPerEC() uint64
	SetFactoshisPerEC(uint64)

	// Get the current transaction block
	GetCurrentBlock() block.IFBlock

	// Update balance updates the balance for a Factoid address in
	// the database.  Note that we take an int64 to allow debits
	// as well as credits
	UpdateBalance(address fct.IAddress, amount int64) error

	// Update balance updates the balance for an Entry Credit address
	// in the database.  Note that we take an int64 to allow debits
	// as well as credits
	UpdateECBalance(address fct.IAddress, amount int64) error

	// Use Entry Credits, which lowers their balance
	UseECs(address fct.IAddress, amount uint64) error

	// Return the Factoid balance for an address
	GetBalance(address fct.IAddress) uint64

	// Return the Entry Credit balance for an address
	GetECBalance(address fct.IAddress) uint64

	// Add a transaction block.  Useful for catching up with the network.
	AddTransactionBlock(block.IFBlock) error

	// Return the Factoid block with this hash.  If unknown, returns
	// a null.
	GetTransactionBlock(fct.IHash) block.IFBlock
	// Put a Factoid block with this hash into the database.
	PutTransactionBlock(fct.IHash, block.IFBlock)

	// Time is something that can vary across multiple systems, and
	// must be controlled in order to build reliable, repeatable
	// tests.  Therefore, no node should directly querry system
	// time.
	GetTimeNano() uint64 // Count of nanoseconds from Jan 1,1970
	GetTime() uint64     // Count of seconds from Jan 1, 1970

	// Validate transaction
	// Return true if the balance of an address covers each input
	Validate(fct.ITransaction) bool

	// Update Transaction just updates the balance sheet with the
	// addition of a transaction.
	UpdateTransaction(fct.ITransaction) bool

	// Add a Transaction to the current block.  The transaction is
	// validated against the address balances, which must cover The
	// inputs.  Returns true if the transaction is added.
	AddTransaction(fct.ITransaction) bool

	// Process End of Minute.
	ProcessEndOfMinute()

	// Process End of Block.
	ProcessEndOfBlock() // to be replaced by ProcessEndOfBlock2
	ProcessEndOfBlock2(uint32)

	// Get the current Directory Block Height
	GetDBHeight() uint32
}

type FactoidState struct {
	IFactoidState
	database       db.IFDatabase
	factoshisPerEC uint64
	currentBlock   block.IFBlock
	dbheight       uint32
	wallet         wallet.ISCWallet
}

var _ IFactoidState = (*FactoidState)(nil)

func (fs *FactoidState) GetWallet() wallet.ISCWallet {
	return fs.wallet
}

func (fs *FactoidState) SetWallet(w wallet.ISCWallet) {
	fs.wallet = w
}

func (fs *FactoidState) GetCurrentBlock() block.IFBlock {
	return fs.currentBlock
}

func (fs *FactoidState) GetDBHeight() uint32 {
	return fs.dbheight
}

// When we are playing catchup, adding the transaction block is a pretty
// useful feature.
func (fs *FactoidState) AddTransactionBlock(blk block.IFBlock) error {
	transactions := blk.GetTransactions()
	for _, trans := range transactions {
		ok := fs.UpdateTransaction(trans)
		if !ok {
			return fmt.Errorf("Failed to add transaction")
		}
	}
	return nil
}

func (fs *FactoidState) AddTransaction(trans fct.ITransaction) bool {
	if !fs.Validate(trans) {
		return false
	}
	if fs.UpdateTransaction(trans) {
		fs.currentBlock.AddTransaction(trans)
		return true
	}
	return false
}

// Assumes validation has already been done.
func (fs *FactoidState) UpdateTransaction(trans fct.ITransaction) bool {
	for _, input := range trans.GetInputs() {
		fmt.Println("Input", input)
		fs.UpdateBalance(input.GetAddress(), -int64(input.GetAmount()))
	}
	for _, output := range trans.GetOutputs() {
		fmt.Println("Output", output)
		fs.UpdateBalance(output.GetAddress(), int64(output.GetAmount()))
	}
	for _, ecoutput := range trans.GetECOutputs() {
		fs.UpdateECBalance(ecoutput.GetAddress(), int64(ecoutput.GetAmount()))
	}
	return true
}

func (fs *FactoidState) ProcessEndOfMinute() {
}

// End of Block means packing the current block away, and setting
// up the next block.
func (fs *FactoidState) ProcessEndOfBlock() {
	var hash fct.IHash

	if fs.currentBlock != nil { // If no blocks, the current block is nil
		hash = fs.currentBlock.GetHash()
		fs.PutTransactionBlock(hash, fs.currentBlock)
		fs.PutTransactionBlock(fct.FACTOID_CHAINID_HASH, fs.currentBlock)
	}

	fs.dbheight += 1
	fs.currentBlock = block.NewFBlock(fs.GetFactoshisPerEC(), fs.dbheight)
	flg, err := fs.currentBlock.AddCoinbase(new(fct.Transaction))
	if !flg || err != nil {
		panic("Failed to add coinbase transaction")
	}
	if hash != nil {
		fs.currentBlock.SetPrevBlock(hash.Bytes())
	}

}

// End of Block means packing the current block away, and setting
// up the next block.
// this function is to replace the existing function: ProcessEndOfBlock
func (fs *FactoidState) ProcessEndOfBlock2(nextBlkHeight uint32) {
	var hash fct.IHash

	if fs.currentBlock != nil { // If no blocks, the current block is nil
		hash = fs.currentBlock.GetHash()
	}

	fs.currentBlock = block.NewFBlock(fs.GetFactoshisPerEC(), nextBlkHeight)
	flg, err := fs.currentBlock.AddCoinbase(new(fct.Transaction))
	if !flg || err != nil {
		panic("Failed to add coinbase transaction")
	}
	if hash != nil {
		fs.currentBlock.SetPrevBlock(hash.Bytes())
	}

}

func (fs *FactoidState) LoadState() error {
	var hashes []fct.IHash
	blk := fs.GetTransactionBlock(fct.FACTOID_CHAINID_HASH)
	// If there is no head for the Factoids in the database, we have an
	// uninitialized database.  We need to add the Genesis Block. TODO
	if blk == nil {
		fct.Prtln("No Genesis Block for Factoids detected.  Adding Genesis Block")
		gb := block.GetGenesisBlock(fs.GetTimeNano(), 1000000, 10, 200000000000)
		fs.PutTransactionBlock(gb.GetHash(), gb)
		fs.PutTransactionBlock(fct.FACTOID_CHAINID_HASH, gb)
		err := fs.AddTransactionBlock(gb)
		if err != nil {
			fct.Prtln("Failed to build initial state.\n", err)
			return err
		}
		fs.dbheight = 0
		fs.currentBlock = gb
		return nil
	}
	// First run back from the head back to the genesis block, collecting hashes.
	for {
		if blk == nil {
			return fmt.Errorf("Block not found or not formated properly")
		}
		hashes = append(hashes, blk.GetHash())
		if bytes.Compare(blk.GetPrevBlock().Bytes(), fct.ZERO_HASH) == 0 {
			break
		}
		tblk := fs.GetTransactionBlock(blk.GetPrevBlock())
		if tblk.GetHash().IsEqual(blk.GetPrevBlock()) != nil {
			return fmt.Errorf("Hash Failure!  Database must be rebuilt")
		}
		blk = tblk
	}

	// Now run forward, and build our accounting
	for i := len(hashes) - 1; i >= 0; i-- {
		blk = fs.GetTransactionBlock(hashes[i])
		if blk == nil {
			return fmt.Errorf("Should never happen.  Block not found in LoadState")
		}
		fct.Prt(blk.GetDBHeight(), " ")
		err := fs.AddTransactionBlock(blk) // updates accounting for this block
		if err != nil {
			fct.Prtln("Failed to rebuild state.\n", err)
			return err
		}
	}
	fs.dbheight = blk.GetDBHeight() + 1
	fs.currentBlock = block.NewFBlock(fs.GetFactoshisPerEC(), fs.dbheight)
	fs.currentBlock.SetPrevBlock(blk.GetHash().Bytes())
	return nil
}

// TODO: TBD - maybe it's better to return false, since panic gets handled elsewhere in the web server and does not crash the program
func (fs *FactoidState) Validate(trans fct.ITransaction) bool {
	if nil == fs {
		//		panic("\n\n!!! ERROR: fs is nil !!!")
		fmt.Println("\n\n!!! ERROR: fs is nil !!!")
		return false
	}

	if nil == fs.currentBlock {
		//		panic("\n\n!!! ERROR: fs.currentBlock is nil !!!")
		fmt.Println("\n\n!!! ERROR: fs.currentBlock is nil !!!")
		return false
	}

	if !fs.currentBlock.ValidateTransaction(trans) {
		return false
	}

	for _, input := range trans.GetInputs() {
		bal := fs.GetBalance(input.GetAddress())
		if input.GetAmount() > bal {
			return false
		}
	}
	return true
}

func (fs *FactoidState) GetFactoshisPerEC() uint64 {
	return fs.factoshisPerEC
}

func (fs *FactoidState) SetFactoshisPerEC(factoshisPerEC uint64) {
	fs.factoshisPerEC = factoshisPerEC
}

func (fs *FactoidState) PutTransactionBlock(hash fct.IHash, trans block.IFBlock) {
	fs.database.Put(fct.DB_FACTOID_BLOCKS, hash, trans)
}

func (fs *FactoidState) GetTransactionBlock(hash fct.IHash) block.IFBlock {
	transblk := fs.database.Get(fct.DB_FACTOID_BLOCKS, hash)
	if transblk == nil {
		return nil
	}
	return transblk.(block.IFBlock)
}

func (fs *FactoidState) GetTimeNano() uint64 {
	return uint64(time.Now().UnixNano())
}

func (fs *FactoidState) GetTime() uint64 {
	return uint64(time.Now().Unix())
}

func (fs *FactoidState) SetDB(database db.IFDatabase) {
	fs.database = database
}

func (fs *FactoidState) GetDB() db.IFDatabase {
	return fs.database
}

// Any address that is not defined has a zero balance.
func (fs *FactoidState) GetBalance(address fct.IAddress) uint64 {
	balance := uint64(0)
	b := fs.database.GetRaw([]byte(fct.DB_F_BALANCES), address.Bytes())
	if b != nil {
		balance = b.(*FSbalance).number
	}
	return balance
}

// Update balance throws an error if your update will drive the balance negative.
func (fs *FactoidState) UpdateBalance(address fct.IAddress, amount int64) error {
	nbalance := int64(fs.GetBalance(address)) + amount
	if nbalance < 0 {
		return fmt.Errorf("New balance cannot be negative")
	}
	balance := uint64(nbalance)
	fs.database.PutRaw([]byte(fct.DB_F_BALANCES), address.Bytes(), &FSbalance{number: balance})
	return nil
}

// Update ec balance throws an error if your update will drive the balance negative.
func (fs *FactoidState) UpdateECBalance(address fct.IAddress, amount int64) error {
	nbalance := int64(fs.GetBalance(address)) + amount
	if nbalance < 0 {
		return fmt.Errorf("New balance cannot be negative")
	}
	balance := uint64(nbalance)
	fs.database.PutRaw([]byte(fct.DB_EC_BALANCES), address.Bytes(), &FSbalance{number: balance})
	return nil
}

// Add to Entry Credit Balance.  Note Entry Credit balances are maintained
// as entry credits, not Factoids.  But adding is done in Factoids, using
// done in Entry Credits. Using lowers the Entry Credit Balance.
func (fs *FactoidState) AddToECBalance(address fct.IAddress, amount uint64) error {
	ecs := amount / fs.GetFactoshisPerEC()
	balance := fs.GetBalance(address) + ecs
	fs.database.PutRaw([]byte(fct.DB_EC_BALANCES), address.Bytes(), &FSbalance{number: balance})
	return nil
}

// Use Entry Credits.  Note Entry Credit balances are maintained
// as entry credits, not Factoids.  But adding is done in Factoids, using
// done in Entry Credits.  Using lowers the Entry Credit Balance.
func (fs *FactoidState) UseECs(address fct.IAddress, amount uint64) error {
	balance := fs.GetBalance(address) - amount
	if balance < 0 {
		return fmt.Errorf("Overdraft of Entry Credits attempted.")
	}
	fs.database.PutRaw([]byte(fct.DB_EC_BALANCES), address.Bytes(), &FSbalance{number: balance})
	return nil
}

// Any address that is not defined has a zero balance.
func (fs *FactoidState) GetECBalance(address fct.IAddress) uint64 {
	balance := uint64(0)
	b := fs.database.GetRaw([]byte(fct.DB_EC_BALANCES), address.Bytes())
	if b != nil {
		balance = b.(*FSbalance).number
	}
	return balance
}
