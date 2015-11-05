// Copyright 2015 Factom Foundation
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
	GetHeader() IABlockHeader
	SetHeader(IABlockHeader)
	GetABEntries() []IABEntry
	SetABEntries([]IABEntry)
	GetDBHeight() uint32
	GetKeyMR() (IHash, error)
	LedgerKeyMR() (IHash, error)
	PartialHash() (IHash, error)
	BuildFullBHash() (err error)
	BuildPartialHash() (err error)
	AddABEntry(e IABEntry) (err error)
	AddEndOfMinuteMarker(eomType byte) (err error)
	GetDBSignature() IABEntry
}

// Admin Block Header
type IABlockHeader interface {
	Printable
	BinaryMarshallable

	GetAdminChainID() IHash
	GetPrevLedgerKeyMR() IHash
	SetPrevLedgerKeyMR(IHash)
	GetDBHeight() uint32
	SetDBHeight(uint32)

	GetHeaderExpansionArea() []byte
	SetHeaderExpansionArea([]byte)

	GetMessageCount() uint32
	SetMessageCount(uint32)
	GetBodySize() uint32
	SetBodySize(uint32)
}

type IABEntry interface {
	Printable
	BinaryMarshallable
	ShortInterpretable

	MarshalledSize() uint64
	Type() byte
	Hash() IHash
}
