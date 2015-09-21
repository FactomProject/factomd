// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid

import (
	// "fmt"
	"encoding"
)

type IBlock interface {
	encoding.BinaryMarshaler   // Easy to support this, just drop the slice.
	encoding.BinaryUnmarshaler // And once in Binary, it must come back.
	//encoding.TextMarshaler     // Using this mostly for debugging
	CustomMarshalText() ([]byte, error)

	// We need the progress through the slice, so we really can't use the stock spec
	// for the UnmarshalBinary() method from encode.  We define our own method that
	// makes the code easier to read and way more efficent.
	UnmarshalBinaryData(data []byte) ([]byte, error)
	String() string // Makes debugging, logging, and error reporting easier

	IsEqual(IBlock) []IBlock // Check if this block is the same as itself.
	//   Returns nil, or the path to the first difference.

	GetDBHash() IHash       // Identifies the class of the object
	GetHash() IHash         // Returns the hash of the object
	GetNewInstance() IBlock // Get a new instance of this object
}

type IFBlock interface {
	fct.IBlock
	fct.Printable
	// Get the ChainID. This is a constant for all Factoids.
	GetChainID() fct.IHash
	// Validation functions
	Validate() error
	ValidateTransaction(int, fct.ITransaction) error
	// Marshal just the header for the block. This is to include the header
	// in the LedgerKeyMR
	MarshalHeader() ([]byte, error)
	// Marshal just the transactions.  This is because we need the length
	MarshalTrans() ([]byte, error)
	// Add a coinbase transaction.  This transaction has no inputs
	AddCoinbase(fct.ITransaction) error
	// Add a proper transaction.  Transactions are validated before
	// being added to the block.
	AddTransaction(fct.ITransaction) error
	// Calculate all the MR and serial hashes for this block.  Done just
	// prior to being persisted.
	CalculateHashes()
	// Hash accessors
	// Get Key MR() hashes the header with the GetBodyMR() of the transactions
	GetKeyMR() fct.IHash
	// Get the MR for the list of transactions
	GetBodyMR() fct.IHash
	// Get the KeyMR of the previous block.
	GetPrevKeyMR() fct.IHash
	SetPrevKeyMR([]byte)
	GetLedgerMR() fct.IHash
	GetLedgerKeyMR() fct.IHash
	GetPrevLedgerKeyMR() fct.IHash
	SetPrevLedgerKeyMR([]byte)
	// Accessors for the Directory Block Height
	SetDBHeight(uint32)
	GetDBHeight() uint32
	// Accessors for the Exchange rate
	SetExchRate(uint64)
	GetExchRate() uint64
	// Accessors for the transactions
	GetTransactions() []fct.ITransaction

	// Mark an end of Minute.  If there are multiple calls with the same minute value
	// the later one simply overwrites the previous one.  Since this is an informational
	// data point, we do not enforce much, other than order (the end of period one can't
	// come before period 2.  We just adjust the periods accordingly.
	EndOfPeriod(min int)

	// Returns the milliTimestamp of the coinbase transaction.  This is used to validate
	// the timestamps of transactions included in the block. Transactions prior to the
	// TRANSACTION_PRIOR_LIMIT or after the TRANSACTION_POST_LIMIT are considered invalid
	// for this block. -1 is returned if no coinbase transaction is found.
	GetCoinbaseTimestamp() int64
}
