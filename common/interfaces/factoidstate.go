// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

type IFactoidState interface {

	// Get the current transaction block
	GetCurrentBlock() IFBlock

	// Get the current balance for a transaction
	GetFactoidBalance(address [32]byte) int64
	GetECBalance(address [32]byte) int64

	// Add a transaction   Useful for catching up with the network.
	AddTransactionBlock(IFBlock) error
	AddECBlock(IEntryCreditBlock) error

	// Get a hash of all the balances at this height
	GetBalanceHash(bool) IHash

	// Validate transaction
	// Return zero len string if the balance of an address covers each input
	Validate(int, ITransaction) (err error, holdAddr [32]byte)

	// Check the transaction timestamp for to ensure it can be included
	// in the current   Transactions that are too old, or dated to
	// far in the future cannot be included in the current block
	ValidateTransactionAge(trans ITransaction) error

	// Update Transaction just updates the balance sheet with the
	// addition of a transaction.  bool must be true if this is a realtime update,
	// and false if processing a block.  This provides real time balances, without
	// double counting transactions when we process blocks.
	UpdateTransaction(bool, ITransaction) error
	UpdateECTransaction(bool, IECBlockEntry) error

	// Add a Transaction to the current   The transaction is
	// validated against the address balances, which must cover The
	// inputs.  Returns true if the transaction is added.
	AddTransaction(int, ITransaction) error

	// Process End of Block
	ProcessEndOfBlock(IState)

	// Set the End of Period.  Currently, each block in Factom is broken
	// into ten, one minute periods.
	EndOfPeriod(period int)

	//GetFactiodAccounts(interface{}) (uint32, []string)

	//GetMultipleFactoidBalances([][32]byte) (uint32, [][]string)
	GetMultipleFactoidBalances([32]byte) (uint32, uint32, int64, int64, string)

	GetMultipleECBalances([32]byte) (uint32, uint32, int64, int64, string)
}
