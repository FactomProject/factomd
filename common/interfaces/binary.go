// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

import (
	"encoding"
)

type BinaryMarshallable interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler

	UnmarshalBinaryData([]byte) ([]byte, error)
}

type BinaryMarshallableAndCopyable interface {
	BinaryMarshallable
	New() BinaryMarshallableAndCopyable
}
