// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package primitives

import (
	//"encoding/hex"
	"fmt"
	"encoding/hex"
	"github.com/FactomProject/ed25519"
	"github.com/FactomProject/factomd/common/interfaces"
)

/*
type DetachedSignature [ed25519.SignatureSize]byte
type DetachedPublicKey [ed25519.PublicKeySize]byte
*/
//Signature has signed data and its corresponsing PublicKey
type Signature struct {
	Pub PublicKey
	Sig *[ed25519.SignatureSize]byte
}

var _ interfaces.BinaryMarshallable = (*Signature)(nil)
var _ interfaces.IFullSignature = (*Signature)(nil)

func (sig *Signature) CustomMarshalText() ([]byte, error) {
	return ([]byte)(sig.Pub.String()+hex.EncodeToString(sig.Sig[:])),nil
}

func (sig *Signature) SetPub(publicKey[]byte) {
	sig.Pub.Key = new([ed25519.PublicKeySize]byte)
	copy(sig.Pub.Key[:],publicKey)
}

func (sig *Signature) GetKey() []byte {
	return (*sig.Pub.Key)[:]
}

func (sig *Signature) SetSignature(signature []byte) error {
	if len(signature) != ed25519.SignatureSize {
		return fmt.Errorf("Signature wrong size")
	}
	sig.Sig = new([ed25519.SignatureSize]byte)
	copy(sig.Sig[:],signature)
	return nil
}

func (sig *Signature) GetSignature() *[ed25519.SignatureSize]byte {
	return sig.Sig
}


func (s *Signature) MarshalBinary() ([]byte, error) {
	if s.Pub.Key == nil || s.Sig == nil {
		return nil, fmt.Errorf("Signature not complete")
	}
	return append(s.Pub.Key[:], s.Sig[:]...), nil
}

func (sig *Signature) UnmarshalBinaryData(data []byte) ([]byte, error) {
	sig.Pub.Key = new([ed25519.PublicKeySize]byte)
	sig.Sig = new([ed25519.SignatureSize]byte)
	copy(sig.Pub.Key[:], data[:ed25519.PublicKeySize])
	data = data[ed25519.PublicKeySize:]
	copy(sig.Sig[:], data[:ed25519.SignatureSize])
	data = data[ed25519.SignatureSize:]
	return data, nil
}

func (s *Signature) UnmarshalBinary(data []byte) error {
	_, err := s.UnmarshalBinaryData(data)
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
	return ed25519.VerifyCanonical(sig.Pub.Key, msg, sig.Sig)
}
