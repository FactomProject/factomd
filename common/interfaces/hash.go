// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

// IHash is an interface to an object interpretable as a hash
type IHash interface {
	BinaryMarshallableAndCopyable
	Printable

	Copy() IHash                  // Returns a copy of this IHash
	Fixed() [32]byte              // Returns the fixed array for use in maps
	PFixed() *[32]byte            // Return a pointer to a Fixed array
	Bytes() []byte                // Return the byte slice for this Hash
	SetBytes([]byte) error        // Set the bytes
	IsSameAs(IHash) bool          // Compare two Hashes
	IsMinuteMarker() bool         // Returns true if the hash can be interpreted as a minute marker (all zeros except last byte, which represents the minute)
	UnmarshalText(b []byte) error // Unmarshal the input data into this IHash
	IsZero() bool                 // Is this hash the zero hash?
	ToMinute() byte               // Returns the last byte of the hash
	IsHashNil() bool              // Returns true if this hash is nil
	//MarshalText() ([]byte, error)
}
