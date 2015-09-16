// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// This is a minimum wallet to be used to test the coin
// There isn't much in the way of interest in security
// here, but rather provides a mechanism to create keys
// and sign transactions, etc.

package wallet

import (
	"crypto/rand"
	"crypto/sha512"
	"errors"
	"fmt"
	"github.com/FactomProject/ed25519"
	fct "github.com/FactomProject/factoid"
	"github.com/FactomProject/factoid/database"
)

// The wallet interface uses bytes.  This is because we want to
// handle fixed length values in our maps and the database.  If
// we try to use strings, then the lengths vary based on encoding
// and that complicates the implementation without really making
// the interface more usable by developers.
type ISCWallet interface {

	//initialize the object.  call before using other functions
	Init(a ...interface{})
	// A New Seed is generated for the wallet.
	NewSeed(data []byte)
	// Set the seed for a wallet
	SetSeed(seed []byte)
	// Get the seed for a wallet
	GetSeed() []byte
	// Set the current deterministic root (Initialization function)
	SetRoot([]byte)
	// Returns the backing database for the wallet
	GetDB() database.IFDatabase
	// Import a key pair.  If the private key is null, this is treated as an
	// external address, useful only as a destination
	AddKeyPair(addrtype string, name []byte, public []byte, private []byte, generateRandomIfAddressPresent bool) (fct.IAddress, error)
	// Generate a Factoid Address
	GenerateFctAddress(name []byte, m int, n int) (fct.IAddress, error)
	// Generate an Entry Credit Address
	GenerateECAddress(name []byte) (fct.IAddress, error)

	// Generate a Factoid Address from a private key
	GenerateFctAddressFromPrivateKey(name []byte, privateKey []byte, m int, n int) (fct.IAddress, error)
	// Generate an Entry Credit Address from a privatekey
	GenerateECAddressFromPrivateKey(name []byte, privateKey []byte) (fct.IAddress, error)

	// Generate a Factoid Address from a human readable private key
	GenerateFctAddressFromHumanReadablePrivateKey(name []byte, privateKey string, m int, n int) (fct.IAddress, error)
	// Generate an Entry Credit Address from a human readable private key
	GenerateECAddressFromHumanReadablePrivateKey(name []byte, privateKey string) (fct.IAddress, error)

	// Generate a Factoid Address from a set of 12 words from the token sale
	GenerateFctAddressFromMnemonic(name []byte, mnemonic string, m int, n int) (fct.IAddress, error)

	// Get details for an address
	GetAddressDetailsAddr(addr []byte) IWalletEntry
	// Returns the Address hash (what we use for inputs) given the public key
	GetAddressHash(fct.IAddress) (fct.IAddress, error)

	/** Transaction calls **/
	// Create a transaction.  This is just the bones, to which the
	// user must add inputs, outputs, and sign before submission.
	// Must pass in the time for the transaction! UTC nanoseconds
	CreateTransaction(time uint64) fct.ITransaction
	// Modify an input.  Used to back fill the transaction fee.
	UpdateInput(fct.ITransaction, int, fct.IAddress, uint64) error
	// Add an input to a transaction
	AddInput(fct.ITransaction, fct.IAddress, uint64) error
	// Add an output to a transaction
	AddOutput(fct.ITransaction, fct.IAddress, uint64) error
	// Add an Entry Credit output to a transaction.  Note that these are
	// denominated in Factoids.  So you need the exchange rate to do this
	// properly.
	AddECOutput(fct.ITransaction, fct.IAddress, uint64) error
	// Validate a transaction.  Just checks that the inputs and outputs are
	// there and properly constructed.
	Validate(int, fct.ITransaction) error
	// Checks that the signatures all validate.
	ValidateSignatures(fct.ITransaction) error
	// Sign the inputs that have public keys to which we have the private
	// keys.  In the future, we will allow transactions with partical signatures
	// to be sent to other people to complete the signing process.  This will
	// be particularly useful with multisig.
	SignInputs(fct.ITransaction) (bool, error) // True if all inputs are signed
	// Sign a CommitEntry or a CommitChain with the eckey
	SignCommit(we IWalletEntry, data []byte) []byte
	// Get the exchange rate of Factoids per Entry Credit
	// 	GetECRate() uint64
}

var factoshisPerEC uint64 = 100000

type SCWallet struct {
	db            database.MapDB
	isInitialized bool //defaults to 0 and false
	RootSeed      []byte
	NextSeed      []byte
}

var _ ISCWallet = (*SCWallet)(nil)

/*************************************
 *       Stubs
 *************************************/

func (SCWallet) GetHash() fct.IHash {
	return nil
}

/***************************************
 *       Methods
 ***************************************/

func (w *SCWallet) SetRoot(root []byte) {
	w.RootSeed = root
}

func (w *SCWallet) GetDB() database.IFDatabase {
	return &w.db
}

func (SCWallet) GetDBHash() fct.IHash {
	return fct.Sha([]byte("SCWallet"))
}

func (w *SCWallet) SignInputs(trans fct.ITransaction) (bool, error) {

	data, err := trans.MarshalBinarySig() // Get the part of the transaction we sign
	if err != nil {
		return false, err
	}

	var numSigs int = 0

	inputs := trans.GetInputs()
	rcds := trans.GetRCDs()
	for i, rcd := range rcds {
		rcd1, ok := rcd.(*fct.RCD_1)
		if ok {
			pub := rcd1.GetPublicKey()
			we := w.db.GetRaw([]byte(fct.W_ADDRESS_PUB_KEY), pub).(*WalletEntry)
			if we != nil {
				var pri [fct.SIGNATURE_LENGTH]byte
				copy(pri[:], we.private[0])
				bsig := ed25519.Sign(&pri, data)
				sig := new(fct.Signature)
				sig.SetSignature(bsig[:])
				sigblk := new(fct.SignatureBlock)
				sigblk.AddSignature(sig)
				trans.SetSignatureBlock(i, sigblk)
				numSigs += 1
			}
		}
	}

	return numSigs == len(inputs), nil
}

// SignCommit will sign the []byte with the Entry Credit Key and return the
// slice with the signature and pubkey appended.
func (w *SCWallet) SignCommit(we IWalletEntry, data []byte) []byte {
	pub := new([fct.ADDRESS_LENGTH]byte)
	copy(pub[:], we.GetKey(0))
	pri := new([fct.PRIVATE_LENGTH]byte)
	copy(pri[:], we.GetPrivKey(0))
	sig := ed25519.Sign(pri, data)
	r := append(data, pub[:]...)
	r = append(r, sig[:]...)

	return r
}

func (w *SCWallet) GetECRate() uint64 {
	return factoshisPerEC
}

func (w *SCWallet) GetAddressDetailsAddr(name []byte) IWalletEntry {
	return w.db.GetRaw([]byte("wallet.address.addr"), name).(IWalletEntry)
}

func (w *SCWallet) generateAddressFromPrivateKey(addrtype string, name []byte, privateKey []byte, m int, n int) (fct.IAddress, error) {
	if addrtype == "fct" && (m != 1 || n != 1) {
		return nil, fmt.Errorf("Multisig addresses are not supported at this time")
	}

	// Get a new public/private key pair
	pub, pri, err := w.generateKeyFromPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	return w.AddKeyPair(addrtype, name, pub, pri, false)
}

func (w *SCWallet) generateAddress(addrtype string, name []byte, m int, n int) (fct.IAddress, error) {
	if addrtype == "fct" && (m != 1 || n != 1) {
		return nil, fmt.Errorf("Multisig addresses are not supported at this time")
	}

	// Get a new public/private key pair
	pub, pri, err := w.generateKey()
	if err != nil {
		return nil, err
	}

	return w.AddKeyPair(addrtype, name, pub, pri, true)
}

func (w *SCWallet) AddKeyPair(addrtype string, name []byte, pub []byte, pri []byte, generateRandomIfAddressPresent bool) (address fct.IAddress, err error) {

	we := new(WalletEntry)

	nm := w.db.GetRaw([]byte(fct.W_NAME), name)
	if nm != nil {
		str := fmt.Sprintf("The name '%s' already exists. Duplicate names are not supported", string(name))
		return nil, fmt.Errorf(str)
	}

	// Make sure we have not generated this pair before;  Keep
	// generating until we have a unique pair.
	for w.db.GetRaw([]byte(fct.W_ADDRESS_PUB_KEY), pub) != nil {
		if generateRandomIfAddressPresent {
			pub, pri, err = w.generateKey()
			if err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("Address already exists in the wallet")
		}
	}

	we.AddKey(pub, pri)
	we.SetName(name)
	we.SetRCD(fct.NewRCD_1(pub))
	if addrtype == "fct" {
		we.SetType("fct")
	} else {
		we.SetType("ec")
	}
	//
	address, _ = we.GetAddress()
	w.db.PutRaw([]byte(fct.W_RCD_ADDRESS_HASH), address.Bytes(), we)
	w.db.PutRaw([]byte(fct.W_ADDRESS_PUB_KEY), pub, we)
	w.db.PutRaw([]byte(fct.W_NAME), name, we)

	return
}

func (w *SCWallet) GenerateECAddress(name []byte) (hash fct.IAddress, err error) {
	return w.generateAddress("ec", name, 1, 1)
}
func (w *SCWallet) GenerateFctAddress(name []byte, m int, n int) (hash fct.IAddress, err error) {
	return w.generateAddress("fct", name, m, n)
}

func (w *SCWallet) GenerateECAddressFromPrivateKey(name []byte, privateKey []byte) (hash fct.IAddress, err error) {
	return w.generateAddressFromPrivateKey("ec", name, privateKey, 1, 1)
}
func (w *SCWallet) GenerateFctAddressFromPrivateKey(name []byte, privateKey []byte, m int, n int) (hash fct.IAddress, err error) {
	return w.generateAddressFromPrivateKey("fct", name, privateKey, m, n)
}

func (w *SCWallet) GenerateECAddressFromHumanReadablePrivateKey(name []byte, privateKey string) (fct.IAddress, error) {
	priv, err := HumanReadableECPrivateKeyToPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	return w.GenerateECAddressFromPrivateKey(name, priv)
}

func (w *SCWallet) GenerateFctAddressFromHumanReadablePrivateKey(name []byte, privateKey string, m int, n int) (fct.IAddress, error) {
	priv, err := HumanReadableFactoidPrivateKeyToPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	return w.GenerateFctAddressFromPrivateKey(name, priv, m, n)
}

func (w *SCWallet) GenerateFctAddressFromMnemonic(name []byte, mnemonic string, m int, n int) (fct.IAddress, error) {
	priv, err := MnemonicStringToPrivateKey(mnemonic)
	if err != nil {
		return nil, err
	}
	return w.GenerateFctAddressFromPrivateKey(name, priv, m, n)
}

func (w *SCWallet) NewSeed(data []byte) {
	if len(data) == 0 {
		return
	} // No data, no change
	hasher := sha512.New()
	hasher.Write(data)
	seedhash := hasher.Sum(nil)
	w.NextSeed = seedhash
	w.RootSeed = seedhash
	b := new(database.ByteStore)
	b.SetBytes(w.RootSeed)
	w.db.PutRaw([]byte(fct.W_SEEDS), fct.CURRENT_SEED[:], b)
	w.db.PutRaw([]byte(fct.W_SEEDS), w.RootSeed[:32], b)
	w.db.PutRaw([]byte(fct.W_SEED_HEADS), w.RootSeed[:32], b)
}

func (w *SCWallet) SetSeed(seed []byte) {
	w.NextSeed = seed
	w.RootSeed = seed
	b := new(database.ByteStore)
	b.SetBytes(w.RootSeed)
	w.db.PutRaw([]byte(fct.W_SEEDS), fct.CURRENT_SEED[:], b)
	w.db.PutRaw([]byte(fct.W_SEEDS), w.RootSeed[:32], b)
	w.db.PutRaw([]byte(fct.W_SEED_HEADS), w.RootSeed[:32], b)
}

func (w *SCWallet) GetSeed() []byte {
	iroot := w.db.GetRaw([]byte(fct.W_SEEDS), fct.CURRENT_SEED[:])
	if iroot == nil {
		randomstuff := make([]byte, 1024)
		rand.Read(randomstuff)
		w.NewSeed(randomstuff)
	}
	hasher := sha512.New()
	hasher.Write([]byte(w.NextSeed))
	seedhash := hasher.Sum(nil)
	w.NextSeed = seedhash

	b := new(database.ByteStore)
	b.SetBytes(w.NextSeed)
	w.db.PutRaw([]byte(fct.W_SEED_HEADS), w.RootSeed[:32], b)

	return w.NextSeed
}

func (w *SCWallet) Init(a ...interface{}) {
	if w.isInitialized != false {
		return
	}
	w.isInitialized = true
	w.db.Init()
}

// This function pulls the next private key from the deterministic
// private key generator, gets the public key associated with it
// then prepares the generator for the next time a private key is needed.
// To prepare the next state, it sha512s the previous sha512 output.
// It returns a 32 byte public key, a 64 byte private key, and an error condition.
// The private key is the SUPERCOP style with the private key in the first 32 bytes
// and the public key is the last 32 bytes.
// The public key essentially returns twice because of this.
func (w *SCWallet) generateKey() (public []byte, private []byte, err error) {

	keypair := new([64]byte)
	// the secret part of the keypair is the top 32 bytes of the sha512 hash
	copy(keypair[:32], w.GetSeed()[:32])
	// the crypto library puts the pubkey in the lower 32 bytes and returns the same 32 bytes.
	pub := ed25519.GetPublicKey(keypair)

	return pub[:], keypair[:], err
}

func GenerateKeyFromPrivateKey(privateKey []byte) (public []byte, private []byte, err error) {
	if len(privateKey) == 64 {
		privateKey = privateKey[:32]
	}
	if len(privateKey) != 32 {
		return nil, nil, errors.New("Wrong privateKey size")
	}
	keypair := new([64]byte)

	copy(keypair[:32], privateKey[:])
	// the crypto library puts the pubkey in the lower 32 bytes and returns the same 32 bytes.
	pub := ed25519.GetPublicKey(keypair)

	return pub[:], keypair[:], err
}

func (w *SCWallet) generateKeyFromPrivateKey(privateKey []byte) (public []byte, private []byte, err error) {
	return GenerateKeyFromPrivateKey(privateKey)
}

func (w *SCWallet) CreateTransaction(time uint64) fct.ITransaction {
	t := new(fct.Transaction)
	t.SetMilliTimestamp(time)
	return t
}

func (w *SCWallet) getWalletEntry(bucket []byte, address fct.IAddress) (IWalletEntry, fct.IAddress, error) {
	
	v := w.db.GetRaw([]byte(fct.W_RCD_ADDRESS_HASH), address.Bytes())
	if v == nil {
		return nil, nil, fmt.Errorf("Unknown address")
	}

	we := v.(*WalletEntry)

	adr, err := we.GetAddress()
	if err != nil {
		return nil, nil, err
	}

	return we, adr, nil
}

// Returns the Address hash (what we use for inputs) given the public key
func (w *SCWallet) GetAddressHash(address fct.IAddress) (fct.IAddress, error) {
	_, adr, err := w.getWalletEntry([]byte(fct.W_RCD_ADDRESS_HASH), address)
	if err != nil {
		return nil, err
	}
	return fct.CreateAddress(adr), nil
}

func (w *SCWallet) AddInput(trans fct.ITransaction, address fct.IAddress, amount uint64) error {
	// Check if this is an address we know.
	we, adr, err := w.getWalletEntry([]byte(fct.W_RCD_ADDRESS_HASH), address)
	// If it isn't, we assume the user knows what they are doing.
	if we == nil || err != nil {
		rcd := fct.NewRCD_1(address.Bytes()) 
		trans.AddRCD(rcd)
		adr, err := rcd.GetAddress()
		if err != nil {
			return err
		}
		trans.AddInput(fct.CreateAddress(adr), amount)
	}else{		
		trans.AddRCD(we.GetRCD())
		trans.AddInput(fct.CreateAddress(adr), amount)
	}
	
	return nil
}

func (w *SCWallet) UpdateInput(trans fct.ITransaction, index int, address fct.IAddress, amount uint64) error {

	we, adr, err := w.getWalletEntry([]byte(fct.W_RCD_ADDRESS_HASH), address)
	if err != nil {
		return err
	}

	in, err := trans.GetInput(index)
	if err != nil {
		return err
	}

	trans.GetRCDs()[index] = we.GetRCD() // The RCD must match the (possibly) new input

	in.SetAddress(adr)
	in.SetAmount(amount)

	return nil
}

func (w *SCWallet) AddOutput(trans fct.ITransaction, address fct.IAddress, amount uint64) error {

	_, adr, err := w.getWalletEntry([]byte(fct.W_RCD_ADDRESS_HASH), address)
	if err != nil {
		adr = address
	}

	trans.AddOutput(fct.CreateAddress(adr), amount)

	return nil
}

func (w *SCWallet) AddECOutput(trans fct.ITransaction, address fct.IAddress, amount uint64) error {

	_, adr, err := w.getWalletEntry([]byte(fct.W_RCD_ADDRESS_HASH), address)
	if err != nil {
		adr = address
	}

	trans.AddECOutput(fct.CreateAddress(adr), amount)
	return nil
}

func (w *SCWallet) Validate(index int, trans fct.ITransaction) error {
	err := trans.Validate(index)
	return err
}

func (w *SCWallet) ValidateSignatures(trans fct.ITransaction) error {
	if trans == nil {
		return fmt.Errorf("Missing Transaction")
	}
	return trans.ValidateSignatures()
}
