// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

import ()

type IHash interface {
	BinaryMarshallableAndCopyable
	Printable

	Copy() IHash
	Fixed() [32]byte       // Returns the fixed array for use in maps
	Bytes() []byte         // Return the byte slice for this Hash
	SetBytes([]byte) error // Set the bytes
	IsSameAs(IHash) bool   // Compare two Hashes
	IsMinuteMarker() bool
	UnmarshalText(b []byte) error
	IsZero() bool
	//MarshalText() ([]byte, error)
}
