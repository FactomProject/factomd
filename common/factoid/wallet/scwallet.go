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
	"github.com/FactomProject/factomd/common/constants"
	. "github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/database/bytestore"
	"github.com/FactomProject/factomd/database/mapdb"
)

var factoshisPerEC uint64 = 100000

type SCWallet struct {
	db            mapdb.MapDB
	isInitialized bool //defaults to 0 and false
	RootSeed      []byte
	NextSeed      []byte
}

var _ interfaces.ISCWallet = (*SCWallet)(nil)

/*************************************
 *       Stubs
 *************************************/

func (SCWallet) GetHash() interfaces.IHash {
	return nil
}

/***************************************
 *       Methods
 ***************************************/

func (w *SCWallet) SetRoot(root []byte) {
	w.RootSeed = root
}

func (w *SCWallet) GetDB() interfaces.IDatabase {
	return &w.db
}

func (w *SCWallet) SignInputs(trans interfaces.ITransaction) (bool, error) {

	data, err := trans.MarshalBinarySig() // Get the part of the transaction we sign
	if err != nil {
		return false, err
	}

	var numSigs int = 0

	inputs := trans.GetInputs()
	rcds := trans.GetRCDs()
	for i, rcd := range rcds {
		rcd1, ok := rcd.(*RCD_1)
		if ok {
			pub := rcd1.GetPublicKey()
			wex, err := w.db.Get([]byte(constants.W_ADDRESS_PUB_KEY), pub, new(WalletEntry))
			if err != nil {
				return false, err
			}
			we := wex.(*WalletEntry)
			if we != nil {
				var pri [constants.SIGNATURE_LENGTH]byte
				copy(pri[:], we.private[0])
				bsig := ed25519.Sign(&pri, data)
				sig := new(FactoidSignature)
				sig.SetSignature(bsig[:])
				sigblk := new(SignatureBlock)
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
func (w *SCWallet) SignCommit(we interfaces.IWalletEntry, data []byte) []byte {
	pub := new([constants.ADDRESS_LENGTH]byte)
	copy(pub[:], we.GetKey(0))
	pri := new([constants.PRIVATE_LENGTH]byte)
	copy(pri[:], we.GetPrivKey(0))
	sig := ed25519.Sign(pri, data)
	r := append(data, pub[:]...)
	r = append(r, sig[:]...)

	return r
}

func (w *SCWallet) GetECRate() uint64 {
	return factoshisPerEC
}

func (w *SCWallet) GetAddressDetailsAddr(name []byte) (interfaces.IWalletEntry, error) {
	we, err := w.db.Get([]byte("wallet.address.addr"), name, new(WalletEntry))
	return we.(interfaces.IWalletEntry), err
}

func (w *SCWallet) generateAddressFromPrivateKey(addrtype string, name []byte, privateKey []byte, m int, n int) (interfaces.IAddress, error) {
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

func (w *SCWallet) generateAddress(addrtype string, name []byte, m int, n int) (interfaces.IAddress, error) {
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

func (w *SCWallet) AddKeyPair(addrtype string, name []byte, pub []byte, pri []byte, generateRandomIfAddressPresent bool) (address interfaces.IAddress, err error) {

	we := new(WalletEntry)

	nm, err := w.db.Get([]byte(constants.W_NAME), name, new(WalletEntry))
	if err != nil {
		return nil, err
	}
	if nm != nil {
		str := fmt.Sprintf("The name '%s' already exists. Duplicate names are not supported", string(name))
		return nil, fmt.Errorf(str)
	}

	// Make sure we have not generated this pair before;  Keep
	// generating until we have a unique pair.
	for {
		p, err := w.db.Get([]byte(constants.W_ADDRESS_PUB_KEY), pub, new(WalletEntry))
		if p == nil {
			break
		}
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
	we.SetRCD(NewRCD_1(pub))
	if addrtype == "fct" {
		we.SetType("fct")
	} else {
		we.SetType("ec")
	}
	//
	address, _ = we.GetAddress()
	err = w.db.Put([]byte(constants.W_RCD_ADDRESS_HASH), address.Bytes(), we)
	if err != nil {
		return nil, err
	}
	err = w.db.Put([]byte(constants.W_ADDRESS_PUB_KEY), pub, we)
	if err != nil {
		return nil, err
	}
	err = w.db.Put([]byte(constants.W_NAME), name, we)
	if err != nil {
		return nil, err
	}

	return
}

func (w *SCWallet) GenerateECAddress(name []byte) (hash interfaces.IAddress, err error) {
	return w.generateAddress("ec", name, 1, 1)
}
func (w *SCWallet) GenerateFctAddress(name []byte, m int, n int) (hash interfaces.IAddress, err error) {
	return w.generateAddress("fct", name, m, n)
}

func (w *SCWallet) GenerateECAddressFromPrivateKey(name []byte, privateKey []byte) (hash interfaces.IAddress, err error) {
	return w.generateAddressFromPrivateKey("ec", name, privateKey, 1, 1)
}
func (w *SCWallet) GenerateFctAddressFromPrivateKey(name []byte, privateKey []byte, m int, n int) (hash interfaces.IAddress, err error) {
	return w.generateAddressFromPrivateKey("fct", name, privateKey, m, n)
}

func (w *SCWallet) GenerateECAddressFromHumanReadablePrivateKey(name []byte, privateKey string) (interfaces.IAddress, error) {
	priv, err := HumanReadableECPrivateKeyToPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	return w.GenerateECAddressFromPrivateKey(name, priv)
}

func (w *SCWallet) GenerateFctAddressFromHumanReadablePrivateKey(name []byte, privateKey string, m int, n int) (interfaces.IAddress, error) {
	priv, err := HumanReadableFactoidPrivateKeyToPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	return w.GenerateFctAddressFromPrivateKey(name, priv, m, n)
}

func (w *SCWallet) GenerateFctAddressFromMnemonic(name []byte, mnemonic string, m int, n int) (interfaces.IAddress, error) {
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
	b := new(bytestore.ByteStore)
	b.SetBytes(w.RootSeed)
	err := w.db.Put([]byte(constants.W_SEEDS), constants.CURRENT_SEED[:], b)
	if err != nil {
		panic(err)
	}
	err = w.db.Put([]byte(constants.W_SEEDS), w.RootSeed[:32], b)
	if err != nil {
		panic(err)
	}
	err = w.db.Put([]byte(constants.W_SEED_HEADS), w.RootSeed[:32], b)
	if err != nil {
		panic(err)
	}
}

func (w *SCWallet) SetSeed(seed []byte) {
	w.NextSeed = seed
	w.RootSeed = seed
	b := new(bytestore.ByteStore)
	b.SetBytes(w.RootSeed)
	err := w.db.Put([]byte(constants.W_SEEDS), constants.CURRENT_SEED[:], b)
	if err != nil {
		panic(err)
	}
	err = w.db.Put([]byte(constants.W_SEEDS), w.RootSeed[:32], b)
	if err != nil {
		panic(err)
	}
	err = w.db.Put([]byte(constants.W_SEED_HEADS), w.RootSeed[:32], b)
	if err != nil {
		panic(err)
	}
}

func (w *SCWallet) GetSeed() []byte {
	iroot, err := w.db.Get([]byte(constants.W_SEEDS), constants.CURRENT_SEED[:], new(bytestore.ByteStore))
	if err != nil {
		panic(err)
	}
	if iroot == nil {
		randomstuff := make([]byte, 1024)
		rand.Read(randomstuff)
		w.NewSeed(randomstuff)
	}
	hasher := sha512.New()
	hasher.Write([]byte(w.NextSeed))
	seedhash := hasher.Sum(nil)
	w.NextSeed = seedhash

	b := new(bytestore.ByteStore)
	b.SetBytes(w.NextSeed)
	err = w.db.Put([]byte(constants.W_SEED_HEADS), w.RootSeed[:32], b)
	if err != nil {
		panic(err)
	}
	return w.NextSeed
}

func (w *SCWallet) Init(a ...interface{}) {
	if w.isInitialized != false {
		return
	}
	w.isInitialized = true
	w.db.Init(nil)
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

func (w *SCWallet) CreateTransaction(time uint64) interfaces.ITransaction {
	t := new(Transaction)
	t.SetMilliTimestamp(time)
	return t
}

func (w *SCWallet) getWalletEntry(bucket []byte, address interfaces.IAddress) (interfaces.IWalletEntry, interfaces.IAddress, error) {

	v, err := w.db.Get([]byte(constants.W_RCD_ADDRESS_HASH), address.Bytes(), new(WalletEntry))
	if err != nil {
		return nil, nil, err
	}
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
func (w *SCWallet) GetAddressHash(address interfaces.IAddress) (interfaces.IAddress, error) {
	_, adr, err := w.getWalletEntry([]byte(constants.W_RCD_ADDRESS_HASH), address)
	if err != nil {
		return nil, err
	}
	return CreateAddress(adr), nil
}

func (w *SCWallet) AddInput(trans interfaces.ITransaction, address interfaces.IAddress, amount uint64) error {
	// Check if this is an address we know.
	we, adr, err := w.getWalletEntry([]byte(constants.W_RCD_ADDRESS_HASH), address)
	// If it isn't, we assume the user knows what they are doing.
	if we == nil || err != nil {
		rcd := NewRCD_1(address.Bytes())
		trans.AddRCD(rcd)
		adr, err := rcd.GetAddress()
		if err != nil {
			return err
		}
		trans.AddInput(CreateAddress(adr), amount)
	} else {
		trans.AddRCD(we.GetRCD())
		trans.AddInput(CreateAddress(adr), amount)
	}

	return nil
}

func (w *SCWallet) UpdateInput(trans interfaces.ITransaction, index int, address interfaces.IAddress, amount uint64) error {

	we, adr, err := w.getWalletEntry([]byte(constants.W_RCD_ADDRESS_HASH), address)
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

func (w *SCWallet) AddOutput(trans interfaces.ITransaction, address interfaces.IAddress, amount uint64) error {

	_, adr, err := w.getWalletEntry([]byte(constants.W_RCD_ADDRESS_HASH), address)
	if err != nil {
		adr = address
	}

	trans.AddOutput(CreateAddress(adr), amount)

	return nil
}

func (w *SCWallet) AddECOutput(trans interfaces.ITransaction, address interfaces.IAddress, amount uint64) error {

	_, adr, err := w.getWalletEntry([]byte(constants.W_RCD_ADDRESS_HASH), address)
	if err != nil {
		adr = address
	}

	trans.AddECOutput(CreateAddress(adr), amount)
	return nil
}

func (w *SCWallet) Validate(index int, trans interfaces.ITransaction) error {
	err := trans.Validate(index)
	return err
}

func (w *SCWallet) ValidateSignatures(trans interfaces.ITransaction) error {
	if trans == nil {
		return fmt.Errorf("Missing Transaction")
	}
	return trans.ValidateSignatures()
}
