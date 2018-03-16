// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

type IEBEntry interface {
	DatabaseBatchable
	Printable

	GetHash() IHash
	ExternalIDs() [][]byte
	GetContent() []byte
	GetChainIDHash() IHash
	IsSameAs(IEBEntry) bool
}

type IEntry interface {
	IEBEntry
	KSize() int
}

type IPendingEntry struct {
	EntryHash IHash  `json:"entryhash"`
	ChainID   IHash  `json:"chainid"`
	Status    string `json:"status"`
}
