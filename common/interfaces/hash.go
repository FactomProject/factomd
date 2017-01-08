// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

import ()

type IHash interface {
	IBlock // Implements IBlock

	Copy() IHash
	Fixed() [32]byte                                   // Returns the fixed array for use in maps
	Bytes() []byte                                     // Return the byte slice for this Hash
	SetBytes([]byte) error                             // Set the bytes
	IsSameAs(IHash) bool                               // Compare two Hashes
	CreateHash(a ...BinaryMarshallable) (IHash, error) // Create a serial Hash from arguments
	HexToHash(hexStr string) (IHash, error)            // Convert a Hex string to a Hash
	IsMinuteMarker() bool
	UnmarshalText(b []byte) error
	IsZero() bool
}
