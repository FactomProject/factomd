// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package primitives

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/FactomProject/ed25519"

	"github.com/FactomProject/btcutil/base58"
	"github.com/FactomProject/go-bip32"
	"github.com/FactomProject/go-bip39"
)

// Mnemonic strings are a series of words put together which can be used to generate a new seed for the random number
// generator. This allows a private key to be deterministically created from that seed. The mnemonic strings may then
// be used when a private key is lost, to restore the proper seed to the generator and recover the original private key.

// Public / Private Keys in this file are either:
// 1) []byte arrays representing the key
// 2) strings representing the private key in characters:
//    a) Human readable is the Wallet Import Format (WIF) which uses the base 58 characters for the string
//    b) Hexidecimal format (0-F)

// MnemonicStringToPrivateKey transforms the input mnemonic word string into a new seed such that a new private key may
// be generated ([]byte) in a 'deterministic' manner
func MnemonicStringToPrivateKey(mnemonic string) ([]byte, error) {
	mnemonic = strings.ToLower(strings.TrimSpace(mnemonic))
	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, "")
	if err != nil {
		return nil, err
	}

	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		return nil, err
	}

	child, err := masterKey.NewChildKey(bip32.FirstHardenedChild + 7)
	if err != nil {
		return nil, err
	}

	return child.Key, nil
}

// MnemonicStringToPrivateKeyString transforms the input mnemonic word string into a new seed such that a new private key
// may be generated (hexidecimal 0-F) in a 'deterministic' manner
func MnemonicStringToPrivateKeyString(mnemonic string) (string, error) {
	key, err := MnemonicStringToPrivateKey(mnemonic)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", key), nil
}

/******************************************************************************/
/********************Human-readable private keys*******************************/
/******************************************************************************/

// HumanReadableFactoidPrivateKeyToPrivateKey returns the []byte array conversion of the input
// human readable Factoid key string (WIF base 58 char)
func HumanReadableFactoidPrivateKeyToPrivateKey(human string) ([]byte, error) {
	human = strings.TrimSpace(human)
	base, v1, v2, err := base58.CheckDecodeWithTwoVersionBytes(human)
	if err != nil {
		return nil, err
	}

	if v1 != 0x64 || v2 != 0x78 {
		return nil, fmt.Errorf("Invalid prefix")
	}

	return base, nil
}

// HumanReadableFactoidPrivateKeyToPrivateKeyString returns the hexidecimal string conversion (0-F) of the input
// human readable Factoid key string (WIF base 58 char)
func HumanReadableFactoidPrivateKeyToPrivateKeyString(human string) (string, error) {
	key, err := HumanReadableFactoidPrivateKeyToPrivateKey(human)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", key), nil
}

// HumanReadableECPrivateKeyToPrivateKey returns the []byte array conversion of the input
// human readable Entry Credit string (WIF base 58 char)
func HumanReadableECPrivateKeyToPrivateKey(human string) ([]byte, error) {
	human = strings.TrimSpace(human)
	base, v1, v2, err := base58.CheckDecodeWithTwoVersionBytes(human)
	if err != nil {
		return nil, err
	}

	if v1 != 0x5d || v2 != 0xb6 {
		return nil, fmt.Errorf("Invalid prefix")
	}

	return base, nil
}

// HumanReadableECPrivateKeyToPrivateKeyString returns the hexidecimal string conversion (0-F) of the input
// human readable Entry Credit string (WIF base 58 char)
func HumanReadableECPrivateKeyToPrivateKeyString(human string) (string, error) {
	key, err := HumanReadableECPrivateKeyToPrivateKey(human)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", key), nil
}

// PrivateKeyStringToHumanReadableFactoidPrivateKey returns the human readable Factoid string (WIF base 58 char) conversion
// from the hexidecimal string (0-F) input
func PrivateKeyStringToHumanReadableFactoidPrivateKey(priv string) (string, error) {
	return PrivateKeyStringToHumanReadablePrivateKey(priv, 0x64, 0x78)
}

// PrivateKeyStringToHumanReadableECPrivateKey returns the human readable Entry Credit string (WIF base 58 char) conversion
// from the hexidecimal string (0-F) input
func PrivateKeyStringToHumanReadableECPrivateKey(priv string) (string, error) {
	return PrivateKeyStringToHumanReadablePrivateKey(priv, 0x5d, 0xb6)
}

// PrivateKeyStringToHumanReadablePrivateKey returns the human readable string (WIF base 58 char) conversion
// from the hexidecimal string (0-F) input
func PrivateKeyStringToHumanReadablePrivateKey(priv string, b1, b2 byte) (string, error) {
	priv = strings.TrimSpace(priv)

	h, err := hex.DecodeString(priv)
	if err != nil {
		return "", err
	}

	return base58.CheckEncodeWithVersionBytes(h, b1, b2), nil
}

/******************************************************************************/
/***************************To public key**************************************/
/******************************************************************************/

// PrivateKeyStringToPublicKeyString generates a public key hexidecimal string (0-F) from an
// input private key hexidecimal string (0-F)
func PrivateKeyStringToPublicKeyString(private string) (string, error) {
	pub, err := PrivateKeyStringToPublicKey(private)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", pub), nil
}

// PrivateKeyStringToPublicKey generates a public key []byte array from an
// input private key hexidecimal string (0-F)
func PrivateKeyStringToPublicKey(private string) ([]byte, error) {
	privHex, err := hex.DecodeString(private)
	if err != nil {
		return nil, err
	}
	return PrivateKeyToPublicKey(privHex)
}

// PrivateKeyToPublicKey returns the 32 byte public key generated from the input private key. Private key may be the key alone
// (size 32) or a keypair (size 64) where the 32 MSBytes are assumed to be the private key.
func PrivateKeyToPublicKey(private []byte) ([]byte, error) {
	pub, _, err := GenerateKeyFromPrivateKey(private)
	if err != nil {
		return nil, err
	}
	return pub, nil
}

// GenerateKeyFromPrivateKey generates a public key from the input private key. The input private key may be the private
// key alone (size 32) or a keypair (size 64) where the 32 MSBytes are assumed to be the private key.
func GenerateKeyFromPrivateKey(privateKey []byte) (public []byte, private []byte, err error) {
	if len(privateKey) == 64 {
		privateKey = privateKey[:32]
	}
	if len(privateKey) != 32 {
		return nil, nil, fmt.Errorf("Wrong privateKey size")
	}
	keypair := new([64]byte)

	copy(keypair[:32], privateKey[:])
	// the crypto library puts the pubkey in the lower 32 bytes and returns the same 32 bytes.
	pub := ed25519.GetPublicKey(keypair)

	return pub[:], keypair[:], err
}
