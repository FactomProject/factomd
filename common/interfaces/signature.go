// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

import ()

/**************************************
 * ISign
 *
 * Interface for RCB Signatures
 *
 * Data structure to support signatures, including multisig.
 **************************************/

type ISignature interface {
	BinaryMarshallable
	SetSignature(sig []byte) error // Set or update the signature
	GetSignature() *[64]byte
	CustomMarshalText() ([]byte, error)
}

type IFullSignature interface {
	ISignature
	SetPub(publicKey []byte)
	// Get the private key
	GetKey() []byte
	// Validate data against this signature
	Verify(data []byte) bool
}

/**************************************
 * ISign
 *
 * Interface for RCB Signatures
 *
 * The signature block holds the signatures that validate one of the RCBs.
 * Each signature has an index, so if the RCD is a multisig, you can know
 * how to apply the signatures to the addresses in the RCD.
 **************************************/
type ISignatureBlock interface {
	IBlock
	GetSignatures() []ISignature
	AddSignature(sig ISignature)
	GetSignature(int) ISignature
}
