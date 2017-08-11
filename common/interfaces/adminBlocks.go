// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

// Administrative Block
// This is a special block which accompanies this Directory Block.
// It contains the signatures and organizational data needed to validate previous and future Directory Blocks.
// This block is included in the DB body. It appears there with a pair of the Admin AdminChainID:SHA256 of the block.
// For more details, please go to:
// https://github.com/FactomProject/FactomDocs/blob/master/factomDataStructureDetails.md#administrative-block
type IAdminBlock interface {
	Printable
	DatabaseBatchable

	IsSameAs(IAdminBlock) bool
	BackReferenceHash() (IHash, error)
	GetABEntries() []IABEntry
	GetDBHeight() uint32
	GetDBSignature() IABEntry
	GetHash() IHash
	GetHeader() IABlockHeader
	GetKeyMR() (IHash, error)
	LookupHash() (IHash, error)
	RemoveFederatedServer(IHash) error
	SetABEntries([]IABEntry)
	SetHeader(IABlockHeader)
	AddEntry(IABEntry) error

	AddABEntry(e IABEntry) error
	AddAuditServer(IHash) error
	AddDBSig(serverIdentity IHash, sig IFullSignature) error
	AddFedServer(IHash) error
	AddFederatedServerBitcoinAnchorKey(IHash, byte, byte, [20]byte) error
	AddFederatedServerSigningKey(IHash, [32]byte) error
	AddFirstABEntry(e IABEntry) error
	AddMatryoshkaHash(IHash, IHash) error
	AddServerFault(IABEntry) error
	UpdateState(IState) error
}

// Admin Block Header
type IABlockHeader interface {
	Printable
	BinaryMarshallable

	IsSameAs(IABlockHeader) bool
	GetAdminChainID() IHash
	GetDBHeight() uint32
	GetPrevBackRefHash() IHash
	SetDBHeight(uint32)
	SetPrevBackRefHash(IHash)

	GetHeaderExpansionArea() []byte
	SetHeaderExpansionArea([]byte)

	GetBodySize() uint32
	GetMessageCount() uint32
	SetBodySize(uint32)
	SetMessageCount(uint32)
}

type IABEntry interface {
	Printable
	BinaryMarshallable
	ShortInterpretable

	UpdateState(IState) error // When loading Admin Blocks,

	Type() byte
	Hash() IHash
}
