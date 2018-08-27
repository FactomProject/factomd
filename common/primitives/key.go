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

func (e *PrivateKey) Init() {
	if e.Key == nil {
		e.Key = new([ed25519.PrivateKeySize]byte)
	}
	if e.Pub == nil {
		e.Pub = new(PublicKey)
	}
}

func RandomPrivateKey() *PrivateKey {
	return NewPrivateKeyFromHexBytes(random.RandByteSliceOfLen(ed25519.PrivateKeySize))
}

func (pk *PrivateKey) CustomMarshalText2(string) ([]byte, error) {
	return ([]byte)(hex.EncodeToString(pk.Key[:]) + pk.Pub.String()), nil
}

func (pk *PrivateKey) Public() []byte {
	return pk.Pub[:]
}

func (pk *PrivateKey) AllocateNew() {
	pk.Key = new([ed25519.PrivateKeySize]byte)
	pk.Pub = new(PublicKey)
}

// Create a new private key from a hex string
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
	pk.AllocateNew()
	copy(pk.Key[:], privKeybytes)
	err = pk.Pub.UnmarshalBinary(privKeybytes[len(privKeybytes)-ed25519.PublicKeySize:])
	if err != nil {
		return nil, err
	}
	return pk, nil
}

func NewPrivateKeyFromHexBytes(privKeybytes []byte) *PrivateKey {
	pk := new(PrivateKey)
	pk.AllocateNew()
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

// Sign signs msg with PrivateKey and return Signature
func (pk *PrivateKey) MarshalSign(msg interfaces.BinaryMarshallable) (sig interfaces.IFullSignature) {
	data, _ := msg.MarshalBinary()
	return pk.Sign(data)
}

//Generate creates new PrivateKey / PublciKey pair or returns error
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

// Returns hex-encoded string of first 32 bytes of key (private key portion)
func (pk *PrivateKey) PrivateKeyString() string {
	return hex.EncodeToString(pk.Key[:32])
}
func (pk *PrivateKey) PublicKeyString() string {
	return pk.Pub.String()
}

/******************PublicKey*******************************/

// PublicKey contains only Public part of Public/Private key pair
type PublicKey [ed25519.PublicKeySize]byte

var _ interfaces.Verifier = (*PublicKey)(nil)
var _ encoding.TextMarshaler = (*PublicKey)(nil)

func (c *PublicKey) Copy() (*PublicKey, error) {
	h := new(PublicKey)
	bytes, err := c.MarshalBinary()
	if err != nil {
		return nil, err
	}
	err = h.UnmarshalBinary(bytes)
	if err != nil {
		return nil, err
	}
	return h, nil
}

func (a PublicKey) Fixed() [ed25519.PublicKeySize]byte {
	return a
}

func (a *PublicKey) IsSameAs(b *PublicKey) bool {
	if b == nil {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (pk *PublicKey) MarshalText() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "PublicKey.MarshalText err:%v", *pe)
		}
	}(&err)
	return []byte(pk.String()), nil
}

func (pk *PublicKey) UnmarshalText(b []byte) error {
	p, err := hex.DecodeString(string(b))
	if err != nil {
		return err
	}
	copy(pk[:], p)
	return nil
}

func (pk *PublicKey) String() string {
	return hex.EncodeToString(pk[:])
}

func PubKeyFromString(instr string) (pk PublicKey) {
	p, _ := hex.DecodeString(instr)
	copy(pk[:], p)
	return
}

func (k *PublicKey) Verify(msg []byte, sig *[ed25519.SignatureSize]byte) bool {
	return ed25519.VerifyCanonical((*[32]byte)(k), msg, sig)
}

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

func (k *PublicKey) UnmarshalBinaryData(p []byte) ([]byte, error) {
	if len(p) < ed25519.PublicKeySize {
		return nil, fmt.Errorf("Invalid data passed")
	}
	copy(k[:], p)
	return p[ed25519.PublicKeySize:], nil
}

func (k *PublicKey) UnmarshalBinary(p []byte) (err error) {
	_, err = k.UnmarshalBinaryData(p)
	return
}

// Verify returns true iff sig is a valid signature of message by publicKey.
func Verify(publicKey *[ed25519.PublicKeySize]byte, message []byte, sig *[ed25519.SignatureSize]byte) bool {
	return ed25519.VerifyCanonical(publicKey, message, sig)
}

// Verify returns true iff sig is a valid signature of message by publicKey.
func VerifySlice(p []byte, message []byte, s []byte) bool {
	sig := new([ed25519.PrivateKeySize]byte)
	pub := new([ed25519.PublicKeySize]byte)
	copy(sig[:], s)
	copy(pub[:], p)
	return Verify(pub, message, sig)
}
