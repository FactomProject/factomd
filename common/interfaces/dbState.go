package interfaces

type IDBState interface {
	GetDirectoryBlock() IDirectoryBlock
	GetAdminBlock() IAdminBlock
	GetFactoidBlock() IFBlock
	GetEntryCreditBlock() IEntryCreditBlock

	GetEntryBlocks() []IEntryBlock
	GetEntries() []IEBEntry
}
