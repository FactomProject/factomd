// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package primitives

import (
	"crypto/rand"
	"encoding"
	"encoding/hex"
	"errors"
	"fmt"
	"os"

	"github.com/FactomProject/ed25519"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives/random"
)

// PrivateKey contains Public/Private key pair
type PrivateKey struct {
	Key *[ed25519.PrivateKeySize]byte
	Pub *PublicKey
}

var _ interfaces.Signer = (*PrivateKey)(nil)

// Init initializes the internal Key/Pub pair to new []byte if previously nil
func (pk *PrivateKey) Init() {
	if pk.Key == nil {
		pk.Key = new([ed25519.PrivateKeySize]byte)
	}
	if pk.Pub == nil {
		pk.Pub = new(PublicKey)
	}
}

// RandomPrivateKey returns a new random private key
func RandomPrivateKey() *PrivateKey {
	return NewPrivateKeyFromHexBytes(random.RandByteSliceOfLen(ed25519.PrivateKeySize))
}

// CustomMarshalText2 is a badly named function which encodes the private key to a string and concatenates
// with the public key to return single string with both private and public keys - suggest renaming:
// MarshalPrivateAndPublicKeys
func (pk *PrivateKey) CustomMarshalText2(string) ([]byte, error) {
	return ([]byte)(hex.EncodeToString(pk.Key[:]) + pk.Pub.String()), nil
}

// Public returns the public key of PrivateKey struct
func (pk *PrivateKey) Public() []byte {
	return pk.Pub[:]
}

// NewPrivateKeyFromHex creates a new private key from a hex string
func NewPrivateKeyFromHex(s string) (*PrivateKey, error) {
	privKeybytes, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	if privKeybytes == nil {
		return nil, errors.New("Invalid private key input string!")
	}
	pk := new(PrivateKey)
	if len(privKeybytes) == ed25519.PrivateKeySize-ed25519.PublicKeySize {
		_, privKeybytes, err = GenerateKeyFromPrivateKey(privKeybytes)
		if err != nil {
			return nil, err
		}
	}
	if len(privKeybytes) != ed25519.PrivateKeySize {
		return nil, errors.New("Invalid private key input string!")
	}
	pk.Init()
	copy(pk.Key[:], privKeybytes)
	err = pk.Pub.UnmarshalBinary(privKeybytes[len(privKeybytes)-ed25519.PublicKeySize:])
	if err != nil {
		return nil, err
	}
	return pk, nil
}

// NewPrivateKeyFromHexBytes creates a new private key from hex []byte
func NewPrivateKeyFromHexBytes(privKeybytes []byte) *PrivateKey {
	pk := new(PrivateKey)
	pk.Init()
	copy(pk.Key[:], privKeybytes)
	pk.Pub.UnmarshalBinary(ed25519.GetPublicKey(pk.Key)[:])
	return pk
}

// Sign signs msg with PrivateKey and return Signature
func (pk *PrivateKey) Sign(msg []byte) (sig interfaces.IFullSignature) {
	sig = new(Signature)
	sig.SetPub(pk.Pub[:])
	s := ed25519.Sign(pk.Key, msg)
	sig.SetSignature(s[:])
	return
}

// MarshalSign marshals and signs msg with PrivateKey and return Signature
func (pk *PrivateKey) MarshalSign(msg interfaces.BinaryMarshallable) (sig interfaces.IFullSignature) {
	data, _ := msg.MarshalBinary()
	return pk.Sign(data)
}

//GenerateKey creates new PrivateKey / PublicKey pair or returns error
func (pk *PrivateKey) GenerateKey() error {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	pk.Key = priv
	pk.Pub = new(PublicKey)
	err = pk.Pub.UnmarshalBinary(pub[:])
	return err
}

// PrivateKeyString returns hex-encoded string of first 32 bytes of key (private key portion)
func (pk *PrivateKey) PrivateKeyString() string {
	return hex.EncodeToString(pk.Key[:32])
}

// PublicKeyString returns string of the public key
func (pk *PrivateKey) PublicKeyString() string {
	return pk.Pub.String()
}

/******************PublicKey*******************************/

// PublicKey contains only Public part of Public/Private key pair
type PublicKey [ed25519.PublicKeySize]byte

var _ interfaces.Verifier = (*PublicKey)(nil)
var _ encoding.TextMarshaler = (*PublicKey)(nil)

// Copy creates a new copy of the public key
func (k *PublicKey) Copy() (*PublicKey, error) {
	h := new(PublicKey)
	bytes, err := k.MarshalBinary()
	if err != nil {
		return nil, err
	}
	err = h.UnmarshalBinary(bytes)
	if err != nil {
		return nil, err
	}
	return h, nil
}

// Fixed returns itself
func (k PublicKey) Fixed() [ed25519.PublicKeySize]byte {
	return k
}

// IsSameAs returns true iff input 'b' is the same as 'k'
func (k *PublicKey) IsSameAs(b *PublicKey) bool {
	if b == nil {
		return false
	}
	for i := range k {
		if k[i] != b[i] {
			return false
		}
	}
	return true
}

// MarshalText marshals the public key into a []byte
func (k *PublicKey) MarshalText() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "PublicKey.MarshalText err:%v", *pe)
		}
	}(&err)
	return []byte(k.String()), nil
}

// UnmarshalText decodes the input []byte public key into 'k'
func (k *PublicKey) UnmarshalText(b []byte) error {
	p, err := hex.DecodeString(string(b))
	if err != nil {
		return err
	}
	copy(k[:], p)
	return nil
}

// String encodes the public key into a string
func (k *PublicKey) String() string {
	return hex.EncodeToString(k[:])
}

// PubKeyFromString decodes an input string into a public key
func PubKeyFromString(instr string) (pk PublicKey) {
	p, _ := hex.DecodeString(instr)
	copy(pk[:], p)
	return
}

// Verify returns true iff sig is a valid signature of message by publicKey k
func (k *PublicKey) Verify(msg []byte, sig *[ed25519.SignatureSize]byte) bool {
	return ed25519.VerifyCanonical((*[32]byte)(k), msg, sig)
}

// MarshalBinary marshals and returns a new []byte containing the PublicKey k
func (k *PublicKey) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "PublicKey.MarshalBinary err:%v", *pe)
		}
	}(&err)
	var buf Buffer
	buf.Write(k[:])
	return buf.DeepCopyBytes(), nil
}

// UnmarshalBinaryData unmarshals the first section of input p of the size of a public key, and returns
// the residual data p along with an error if needed
func (k *PublicKey) UnmarshalBinaryData(p []byte) ([]byte, error) {
	if len(p) < ed25519.PublicKeySize {
		return nil, fmt.Errorf("Invalid data passed")
	}
	copy(k[:], p)
	return p[ed25519.PublicKeySize:], nil
}

// UnmarshalBinary unmarshals the input p into the public key k, returning an error if needed
func (k *PublicKey) UnmarshalBinary(p []byte) (err error) {
	_, err = k.UnmarshalBinaryData(p)
	return
}

// Verify returns true iff sig is a valid signature of message by publicKey.
func Verify(publicKey *[ed25519.PublicKeySize]byte, message []byte, sig *[ed25519.SignatureSize]byte) bool {
	return ed25519.VerifyCanonical(publicKey, message, sig)
}

// VerifySlice returns true iff sig is a valid signature of message by publicKey.
func VerifySlice(p []byte, message []byte, s []byte) bool {
	sig := new([ed25519.PrivateKeySize]byte)
	pub := new([ed25519.PublicKeySize]byte)
	copy(sig[:], s)
	copy(pub[:], p)
	return Verify(pub, message, sig)
}
