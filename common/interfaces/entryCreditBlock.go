// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

import ()

type IEntryCreditBlock interface {
	Printable
	DatabaseBatchable
	GetHeader() IECBlockHeader
	GetBody() IECBlockBody
	HeaderHash() (IHash, error)
	Hash() (IHash, error)
}

type IECBlockHeader interface {
	GetBodyHash() IHash
	SetBodyHash(IHash)
	GetPrevHeaderHash() IHash
	SetPrevHeaderHash(IHash)
	GetPrevLedgerKeyMR() IHash
	SetPrevLedgerKeyMR(IHash)
	GetDBHeight() uint32
	SetDBHeight(uint32)
	GetECChainID() IHash
	SetHeaderExpansionArea([]byte)
	GetHeaderExpansionArea() []byte
	GetObjectCount() uint64
	SetObjectCount(uint64)
	GetBodySize() uint64
	SetBodySize(uint64)
}

type IECBlockBody interface {
	GetEntries() []IECBlockEntry
	SetEntries([]IECBlockEntry)
}

type IECBlockEntry interface {
	Printable
	ShortInterpretable
	
	ECID() byte
	MarshalBinary() ([]byte, error)
	UnmarshalBinary(data []byte) error
	Hash() IHash
}