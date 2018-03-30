// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

// The wallet interface uses bytes.  This is because we want to
// handle fixed length values in our maps and the   If
// we try to use strings, then the lengths vary based on encoding
// and that complicates the implementation without really making
// the interface more usable by developers.
type ISCWallet interface {
	//initialize the object.  call before using other functions
	Init(string, string)
	// A New Seed is generated for the wallet.
	NewSeed(data []byte)
	// Set the seed for a wallet
	SetSeed(seed []byte)
	// Get the seed for a wallet
	GetSeed() []byte
	// Set the current deterministic root (Initialization function)
	SetRoot([]byte)
	// Returns the backing database for the wallet
	GetDB() ISCDatabaseOverlay
	// Import a key pair.  If the private key is null, this is treated as an
	// external address, useful only as a destination
	AddKeyPair(addrtype string, name []byte, public []byte, private []byte, generateRandomIfAddressPresent bool) (IAddress, error)
	// Generate a Factoid Address
	GenerateFctAddress(name []byte, m int, n int) (IAddress, error)
	// Generate an Entry Credit Address
	GenerateECAddress(name []byte) (IAddress, error)

	// Generate a Factoid Address from a private key
	GenerateFctAddressFromPrivateKey(name []byte, privateKey []byte, m int, n int) (IAddress, error)
	// Generate an Entry Credit Address from a privatekey
	GenerateECAddressFromPrivateKey(name []byte, privateKey []byte) (IAddress, error)

	// Generate a Factoid Address from a human readable private key
	GenerateFctAddressFromHumanReadablePrivateKey(name []byte, privateKey string, m int, n int) (IAddress, error)
	// Generate an Entry Credit Address from a human readable private key
	GenerateECAddressFromHumanReadablePrivateKey(name []byte, privateKey string) (IAddress, error)

	// Generate a Factoid Address from a set of 12 words from the token sale
	GenerateFctAddressFromMnemonic(name []byte, mnemonic string, m int, n int) (IAddress, error)

	// Get details for an address
	GetAddressDetailsAddr(addr []byte) (IWalletEntry, error)
	// Returns the Address hash (what we use for inputs) given the public key
	GetAddressHash(IAddress) (IAddress, error)

	/** Transaction calls **/
	// Create a transaction.  This is just the bones, to which the
	// user must add inputs, outputs, and sign before submission.
	// Must pass in the time for the transaction! UTC nanoseconds
	CreateTransaction(time uint64) ITransaction
	// Modify an input.  Used to back fill the transaction fee.
	UpdateInput(ITransaction, int, IAddress, uint64) error
	// Add an input to a transaction
	AddInput(ITransaction, IAddress, uint64) error
	// Add an output to a transaction
	AddOutput(ITransaction, IAddress, uint64) error
	// Add an Entry Credit output to a transaction.  Note that these are
	// denominated in Factoids.  So you need the exchange rate to do this
	// properly.
	AddECOutput(ITransaction, IAddress, uint64) error
	// Validate a transaction.  Just checks that the inputs and outputs are
	// there and properly constructed.
	Validate(int, ITransaction) error
	// Checks that the signatures all validate.
	ValidateSignatures(ITransaction) error
	// Sign the inputs that have public keys to which we have the private
	// keys.  In the future, we will allow transactions with particular signatures
	// to be sent to other people to complete the signing process.  This will
	// be particularly useful with multisig.
	SignInputs(ITransaction) (bool, error) // True if all inputs are signed
	// Sign a CommitEntry or a CommitChain with the eckey
	SignCommit(we IWalletEntry, data []byte) []byte
	// Get the exchange rate of Factoids per Entry Credit
	// 	GetECRate() uint64
}

type IWalletEntry interface {
	BinaryMarshallable
	Printable

	// Set the RCD for this entry.  USE WITH CAUTION!  You change
	// the hash and thus the address returned by the wallet entry!
	SetRCD(IRCD)
	// Get the RCD used to validate an input
	GetRCD() IRCD
	// Add a public and private key.  USE WITH CAUTION! You change
	// the hash and thus the address returned by the wallet entry!
	AddKey(public, private []byte)
	// Get the name for this address
	GetName() []byte
	// Get the Public Key by its index
	GetKey(i int) []byte
	// Get the Private Key by its index
	GetPrivKey(i int) []byte
	// Set the name for this address
	SetName([]byte)
	// Get the address defined by the RCD for this wallet entry.
	GetAddress() (IAddress, error)
	// Return "ec" for Entry Credit address, and "fct" for a Factoid address
	GetType() string
	SetType(string)
}
