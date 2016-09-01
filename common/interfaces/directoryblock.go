// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

type IDirectoryBlock interface {
	Printable
	DatabaseBlockWithEntries

	GetHeader() IDirectoryBlockHeader
	SetHeader(IDirectoryBlockHeader)
	GetDBEntries() []IDBEntry
	SetDBEntries([]IDBEntry) error
	AddEntry(chainID IHash, keyMR IHash) error
	BuildKeyMerkleRoot() (IHash, error)
	BuildBodyMR() (IHash, error)
	GetKeyMR() IHash
	GetHash() IHash
	GetFullHash() IHash

	GetTimestamp() Timestamp
	HeaderHash() (IHash, error)
	BodyKeyMR() IHash
	GetEntryHashesForBranch() []IHash

	SetEntryHash(hash, chainID IHash, index int)
	SetABlockHash(aBlock IAdminBlock) error
	SetECBlockHash(ecBlock IEntryCreditBlock) error
	SetFBlockHash(fBlock IFBlock) error
	CheckNetworkID(network int) error
}

type IDirectoryBlockHeader interface {
	Printable
	BinaryMarshallable

	GetVersion() byte
	SetVersion(byte)
	GetPrevFullHash() IHash
	SetPrevFullHash(IHash)
	GetBodyMR() IHash
	SetBodyMR(IHash)
	GetPrevKeyMR() IHash
	SetPrevKeyMR(IHash)
	GetDBHeight() uint32
	SetDBHeight(uint32)
	GetBlockCount() uint32
	SetBlockCount(uint32)
	GetNetworkID() uint32
	SetNetworkID(uint32)
	GetTimestamp() Timestamp
	SetTimestamp(Timestamp)
}

type IDBEntry interface {
	Printable
	BinaryMarshallable
	GetChainID() IHash
	SetChainID(IHash)
	GetKeyMR() IHash
	SetKeyMR(IHash)
}
