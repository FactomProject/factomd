// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

import (
	"encoding"
)

// BinaryMarshallable represents an object which is binary marshallable and unmarshallable
type BinaryMarshallable interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler

	UnmarshalBinaryData([]byte) ([]byte, error)
}

// BinaryMarshallableAndCopyable represents an object which is binary marshallable/unmarshallable and copyable
type BinaryMarshallableAndCopyable interface {
	BinaryMarshallable                  // All of the above
	New() BinaryMarshallableAndCopyable // Plust the copy piece
}
