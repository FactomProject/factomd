// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

import ()

type IBlock interface {
	BinaryMarshallable
	Printable

	CustomMarshalText() ([]byte, error)

	// We need the progress through the slice, so we really can't use the stock spec
	// for the UnmarshalBinary() method from encode.  We define our own method that
	// makes the code easier to read and way more efficent.
	//String() string // Makes debugging, logging, and error reporting easier

	IsEqual(IBlock) []IBlock // Check if this block is the same as itself.
	//   Returns nil, or the path to the first difference.

	GetHash() IHash // Returns the hash of the object
}

type IFBlock interface {
	IBlock

	//DatabaseBlockWithEntries
	GetDatabaseHeight() uint32
	DatabasePrimaryIndex() IHash   //block.KeyMR()
	DatabaseSecondaryIndex() IHash //block.GetHash()
	New() BinaryMarshallableAndCopyable
	GetEntryHashes() []IHash
	GetEntrySigHashes() []IHash
	GetTransactionByHash(hash IHash) ITransaction

	// Get the ChainID. This is a constant for all Factoids.
	GetChainID() IHash
	// Validation functions
	Validate() error
	ValidateTransaction(int, ITransaction) error
	// Marshal just the header for the block. This is to include the header
	// in the FullHash
	MarshalHeader() ([]byte, error)
	// Marshal just the transactions.  This is because we need the length
	MarshalTrans() ([]byte, error)
	// Add a coinbase transaction.  This transaction has no inputs
	AddCoinbase(ITransaction) error
	// Add a proper transaction.  Transactions are validated before
	// being added to the block.
	AddTransaction(ITransaction) error
	// Calculate all the MR and serial hashes for this block.  Done just
	// prior to being persisted.
	CalculateHashes()
	// Hash accessors
	// Get Key MR() hashes the header with the GetBodyMR() of the transactions
	GetKeyMR() IHash
	// Get the MR for the list of transactions
	GetBodyMR() IHash
	// Get the KeyMR of the previous block.
	GetPrevKeyMR() IHash
	SetPrevKeyMR(IHash)
	GetLedgerKeyMR() IHash
	GetPrevLedgerKeyMR() IHash
	SetPrevLedgerKeyMR(IHash)
	// Accessors for the Directory Block Height
	SetDBHeight(uint32)
	GetDBHeight() uint32
	// Accessors for the Exchange rate
	SetExchRate(uint64)
	GetExchRate() uint64
	// Accessors for the transactions
	GetTransactions() []ITransaction

	// Mark an end of Minute.  If there are multiple calls with the same minute value
	// the later one simply overwrites the previous one.  Since this is an informational
	// data point, we do not enforce much, other than order (the end of period one can't
	// come before period 2.  We just adjust the periods accordingly.
	EndOfPeriod(min int)
	GetEndOfPeriod() [10]int
	// Returns the milliTimestamp of the coinbase transaction.  This is used to validate
	// the timestamps of transactions included in the block. Transactions prior to the
	// TRANSACTION_PRIOR_LIMIT or after the TRANSACTION_POST_LIMIT are considered invalid
	// for this block. -1 is returned if no coinbase transaction is found.
	GetCoinbaseTimestamp() Timestamp

	GetNewInstance() IFBlock // Get a new instance of this object
}
