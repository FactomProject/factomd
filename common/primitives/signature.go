// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package primitives

import (
	//"encoding/hex"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/FactomProject/ed25519"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives/random"
)

/*
type DetachedSignature [ed25519.SignatureSize]byte
type DetachedPublicKey [ed25519.PublicKeySize]byte
*/
//Signature has signed data and its corresponding PublicKey
type Signature struct {
	Pub *PublicKey    `json:"pub"`
	Sig *ByteSliceSig `json:"sig"`
}

var _ interfaces.BinaryMarshallable = (*Signature)(nil)
var _ interfaces.IFullSignature = (*Signature)(nil)

// Init initializes the Signature member pointers with new objects
func (sig *Signature) Init() {
	if sig.Pub == nil {
		sig.Pub = new(PublicKey)
	}
	if sig.Sig == nil {
		sig.Sig = new(ByteSliceSig)
	}
}

// GetPubBytes returns the internal public key, initializing the key if nil
func (sig *Signature) GetPubBytes() []byte {
	sig.Init()
	return sig.Pub[:]
}

// GetSigBytes returns the internal signature, initializing the signature if nil
func (sig *Signature) GetSigBytes() []byte {
	sig.Init()
	return sig.Sig[:]
}

// RandomSignatureSet creates a random private key, random data, and signs the data
func RandomSignatureSet() ([]byte, interfaces.Signer, interfaces.IFullSignature) {
	priv := RandomPrivateKey()
	data := random.RandNonEmptyByteSlice()
	sig := priv.Sign(data)

	return data, priv, sig
}

// IsSameAs returns true iff all aspects of the signatures match
func (a *Signature) IsSameAs(b interfaces.IFullSignature) bool {
	if b == nil {
		return false
	}
	s := b.(*Signature)

	if a.Sig == nil && s.Sig != nil {
		return false
	}
	if s.Sig == nil && a.Sig != nil {
		return false
	}
	for i := range a.Sig {
		if a.Sig[i] != s.Sig[i] {
			return false
		}
	}

	if a.Pub.IsSameAs(s.Pub) == false {
		return false
	}

	return true
}

// CustomMarshalText returns the public key + hecidecimal encoded signature, initializes signature if nil
func (sig *Signature) CustomMarshalText() ([]byte, error) {
	sig.Init()
	return ([]byte)(sig.Pub.String() + hex.EncodeToString(sig.Sig[:])), nil
}

// Bytes returns the signature bytes, if signature is nil, returns nil instead of initializing it
func (sig *Signature) Bytes() []byte {
	if sig.Sig == nil {
		return nil
	}
	return sig.Sig[:]
}

// SetPub sets the public key to the input value
func (sig *Signature) SetPub(publicKey []byte) {
	sig.Pub = new(PublicKey)
	sig.Pub.UnmarshalBinary(publicKey)
}

// GetKey returns the public key, if key is nil, initializes it
func (sig *Signature) GetKey() []byte {
	sig.Init()
	return sig.Pub[:]
}

// SetSignature sets the signature to the input value
func (sig *Signature) SetSignature(signature []byte) error {
	if len(signature) != ed25519.SignatureSize {
		return fmt.Errorf("Signature wrong size")
	}
	sig.Sig = new(ByteSliceSig)
	copy(sig.Sig[:], signature)
	return nil
}

// GetSignature returns the signature
func (sig *Signature) GetSignature() *[ed25519.SignatureSize]byte {
	sig.Init()
	return (*[ed25519.SignatureSize]byte)(sig.Sig)
}

// MarshalBinary marshals the public key and signature
func (sig *Signature) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "Signature.MarshalBinary err:%v", *pe)
		}
	}(&err)
	if sig.Sig == nil {
		return nil, fmt.Errorf("Signature not complete")
	}
	sig.Init()
	return append(sig.Pub[:], sig.Sig[:]...), nil
}

// UnmarshalBinaryData unmarshals the input data (key + signature) into the Signature object
func (sig *Signature) UnmarshalBinaryData(data []byte) ([]byte, error) {
	if data == nil || len(data) < ed25519.SignatureSize+ed25519.PublicKeySize {
		return nil, fmt.Errorf("Not enough data to unmarshal")
	}
	sig.Sig = new(ByteSliceSig)
	var err error
	sig.Pub = new(PublicKey)
	data, err = sig.Pub.UnmarshalBinaryData(data)
	if err != nil {
		return nil, err
	}
	copy(sig.Sig[:], data[:ed25519.SignatureSize])
	data = data[ed25519.SignatureSize:]
	return data, nil
}

// UnmarshalBinary unmarshals the input data (key + signature) into the Signature object
func (sig *Signature) UnmarshalBinary(data []byte) error {
	_, err := sig.UnmarshalBinaryData(data)
	return err
}

/*
func (sig *Signature) DetachSig() *DetachedSignature {
	return (*DetachedSignature)(sig.Sig)
}

func (ds *DetachedSignature) String() string {
	return hex.EncodeToString(ds[:])
}*/

// Verify returns true iff sig is a valid signature of msg by PublicKey.
func (sig *Signature) Verify(msg []byte) bool {
	sig.Init()
	return ed25519.VerifyCanonical((*[32]byte)(sig.Pub), msg, (*[ed25519.SignatureSize]byte)(sig.Sig))
}

// SignSignable signs input data with input private key
func SignSignable(priv []byte, data interfaces.ISignable) ([]byte, error) {
	d, err := data.MarshalBinarySig()
	if err != nil {
		return nil, err
	}
	return Sign(priv, d), nil
}

// Sign signs the input data with input private key
func Sign(priv, data []byte) []byte {
	priv2 := [64]byte{}
	if len(priv) == 64 {
		copy(priv2[:], priv[:])
	} else if len(priv) == 32 {
		copy(priv2[:], priv[:])
		pub := ed25519.GetPublicKey(&priv2)
		copy(priv2[:], append(priv, pub[:]...)[:])
	} else {
		return nil
	}

	return ed25519.Sign(&priv2, data)[:constants.SIGNATURE_LENGTH]
}

// VerifySignature returns an error iff input signature is NOT a valid signature of data by the public key.
func VerifySignature(data, publicKey, signature []byte) error {
	pub := [32]byte{}
	sig := [64]byte{}

	if len(publicKey) == 32 {
		copy(pub[:], publicKey[:])
	} else {
		return fmt.Errorf("Invalid public key length")
	}

	if len(signature) == 64 {
		copy(sig[:], signature[:])
	} else {
		return fmt.Errorf("Invalid signature length")
	}

	valid := ed25519.VerifyCanonical(&pub, data, &sig)
	if valid == false {
		return fmt.Errorf("Invalid signature")
	}
	return nil
}
