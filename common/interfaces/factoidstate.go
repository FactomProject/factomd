// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

import ()

type IFactoidState interface {

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

	// Add a transaction   Useful for catching up with the network.
	AddTransactionBlock(IFBlock) error

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

	// Process End of Block
	ProcessEndOfBlock(IState)

	// Set the End of Period.  Currently, each block in Factom is broken
	// into ten, one minute periods.
	EndOfPeriod(period int)
}

/*
type IFSbalance interface {
	IBlock
	getNumber() uint64
	setNumber(uint64)
}
*/
