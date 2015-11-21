// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

import ()

type IEBEntry interface {
	BinaryMarshallable
	GetHash() IHash
	ExternalIDs() [][]byte
	GetContent() []byte
	GetChainIDHash() IHash
}

type IEntry interface {
	BinaryMarshallable
	GetChainID() IHash
	GetKeyMR() (IHash, error)
}
