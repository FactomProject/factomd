// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package primitives

import (
	//"encoding/hex"
	"github.com/FactomProject/ed25519"
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

func (sig *Signature) Key() []byte {
	return (*sig.Pub.Key)[:]
}

/*
func (sig *Signature) DetachSig() *DetachedSignature {
	return (*DetachedSignature)(sig.Sig)
}

func (ds *DetachedSignature) String() string {
	return hex.EncodeToString(ds[:])
}*/

func UnmarshalBinarySignature(data []byte) (sig Signature) {
	sig.Pub.Key = new([ed25519.PublicKeySize]byte)
	sig.Sig = new([ed25519.SignatureSize]byte)
	copy(sig.Pub.Key[:], data[:ed25519.PublicKeySize])
	data = data[ed25519.PublicKeySize:]
	copy(sig.Sig[:], data[:ed25519.SignatureSize])
	return
}

// Verify returns true iff sig is a valid signature of msg by PublicKey.
func (sig *Signature) Verify(msg []byte) bool {
	return ed25519.VerifyCanonical(sig.Pub.Key, msg, sig.Sig)
}
