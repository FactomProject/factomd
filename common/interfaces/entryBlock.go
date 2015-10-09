// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

import ()

type IEBEntry interface {
	Printable
	BinaryMarshallable

	IsValid() bool
	Hash() IHash
	GetChainID() IHash
	ExternalIDs() [][]byte
	GetContent() []byte
}
