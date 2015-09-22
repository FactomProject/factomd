// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

import ()

type IFactoidState interface {
	// Set the database for the Coin State.  This is where
	// we manage the balances for transactions.  We also look
	// for previous blocks here.
	SetDB(IFDatabase)
	GetDB() IFDatabase

	// Load the address state of Factoids
	LoadState() error

	// Get the wallet used to help manage the Factoid State in
	// some applications.
	GetWallet() ISCWallet
	SetWallet(ISCWallet)

	// The Exchange Rate for Entry Credits in Factoshis per
	// Entry Credits
	GetFactoshisPerEC() uint64
	SetFactoshisPerEC(uint64)

	// Get the current transaction block
	GetCurrentBlock() IFBlock

	// Update balance updates the balance for a Factoid address in
	// the database.  Note that we take an int64 to allow debits
	// as well as credits
	UpdateBalance(address IAddress, amount int64) error

	// Update balance updates the balance for an Entry Credit address
	// in the database.  Note that we take an int64 to allow debits
	// as well as credits
	UpdateECBalance(address IAddress, amount int64) error

	// Use Entry Credits, which lowers their balance
	UseECs(address IAddress, amount uint64) error

	// Return the Factoid balance for an address
	GetBalance(address IAddress) uint64

	// Return the Entry Credit balance for an address
	GetECBalance(address IAddress) uint64

	// Add a transaction   Useful for catching up with the network.
	AddTransactionBlock(IFBlock) error

	// Return the Factoid block with this hash.  If unknown, returns
	// a null.
	GetTransactionBlock(IHash) IFBlock
	// Put a Factoid block with this hash into the database.
	PutTransactionBlock(IHash, IFBlock)

	// Time is something that can vary across multiple systems, and
	// must be controlled in order to build reliable, repeatable
	// tests.  Therefore, no node should directly querry system
	// time.
	GetTimeMilli() uint64 // Count of milliseconds from Jan 1,1970
	GetTime() uint64      // Count of seconds from Jan 1, 1970

	// Validate transaction
	// Return zero len string if the balance of an address covers each input
	Validate(int, ITransaction) error

	// Check the transaction timestamp for to ensure it can be included
	// in the current   Transactions that are too old, or dated to
	// far in the future cannot be included in the current block
	ValidateTransactionAge(trans ITransaction) error

	// Update Transaction just updates the balance sheet with the
	// addition of a transaction.
	UpdateTransaction(ITransaction) error

	// Add a Transaction to the current   The transaction is
	// validated against the address balances, which must cover The
	// inputs.  Returns true if the transaction is added.
	AddTransaction(int, ITransaction) error

	// Process End of Minute.
	ProcessEndOfMinute()

	// Process End of
	ProcessEndOfBlock() // to be replaced by ProcessEndOfBlock2
	ProcessEndOfBlock2(uint32)

	// Get the current Directory Block Height
	GetDBHeight() uint32

	// Set the End of Period.  Currently, each block in Factom is broken
	// into ten, one minute periods.
	EndOfPeriod(period int)
}

type IFSbalance interface {
	IBlock
	getNumber() uint64
	setNumber(uint64)
}
