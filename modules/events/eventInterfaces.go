package events

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

type ICommitChain interface {
	GetVersion() uint8
	GetMilliTime() *primitives.ByteSlice6
	GetChainIDHash() interfaces.IHash
	GetWeld() interfaces.IHash
	GetEntryHash() interfaces.IHash
	GetCredits() uint8
	GetECPubKey() *primitives.ByteSlice32
	GetSig() *primitives.ByteSlice64
}

type ICommitEntry interface {
	GetVersion() uint8
	GetMilliTime() *primitives.ByteSlice6
	GetEntryHash() interfaces.IHash
	GetCredits() uint8
	GetECPubKey() *primitives.ByteSlice32
	GetSig() *primitives.ByteSlice64
}

type IRevealEntry interface {
	GetHash() interfaces.IHash
	ExternalIDs() [][]byte
	GetContent() []byte
	GetChainIDHash() interfaces.IHash
}

type IDBState interface {
	GetDirectoryBlock() interfaces.IDirectoryBlock
	GetAdminBlock() interfaces.IAdminBlock
	GetFactoidBlock() interfaces.IFBlock
	GetEntryCreditBlock() interfaces.IEntryCreditBlock

	GetEntryBlocks() []interfaces.IEntryBlock
	GetEntries() []interfaces.IEBEntry
}
