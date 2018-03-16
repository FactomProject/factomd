//  Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

type IEntryCreditBlock interface {
	Printable
	DatabaseBatchable

	GetHeader() IECBlockHeader
	GetBody() IECBlockBody
	GetHash() IHash
	HeaderHash() (IHash, error)
	GetFullHash() (IHash, error)
	GetEntryHashes() []IHash
	GetEntrySigHashes() []IHash
	GetEntries() []IECBlockEntry
	GetEntryByHash(hash IHash) IECBlockEntry

	UpdateState(IState) error
	IsSameAs(IEntryCreditBlock) bool
	BuildHeader() error
}

type IECBlockHeader interface {
	BinaryMarshallable

	String() string
	GetBodyHash() IHash
	SetBodyHash(IHash)
	GetPrevHeaderHash() IHash
	SetPrevHeaderHash(IHash)
	GetPrevFullHash() IHash
	SetPrevFullHash(IHash)
	GetDBHeight() uint32
	SetDBHeight(uint32)
	GetECChainID() IHash
	SetHeaderExpansionArea([]byte)
	GetHeaderExpansionArea() []byte
	GetObjectCount() uint64
	SetObjectCount(uint64)
	GetBodySize() uint64
	SetBodySize(uint64)
	IsSameAs(IECBlockHeader) bool
}

type IECBlockBody interface {
	String() string
	GetEntries() []IECBlockEntry
	SetEntries([]IECBlockEntry)
	AddEntry(IECBlockEntry)
	IsSameAs(IECBlockBody) bool
}

type IECBlockEntry interface {
	Printable
	ShortInterpretable

	ECID() byte
	MarshalBinary() ([]byte, error)
	UnmarshalBinary(data []byte) error
	UnmarshalBinaryData(data []byte) ([]byte, error)
	Hash() IHash
	GetHash() IHash
	GetEntryHash() IHash
	GetSigHash() IHash
	GetTimestamp() Timestamp
	IsSameAs(IECBlockEntry) bool
}
