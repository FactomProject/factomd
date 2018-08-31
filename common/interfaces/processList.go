package interfaces

type IProcessList interface {
	//Clear()
	GetKeysNewEntries() (keys [][32]byte)
	GetNewEntry(key [32]byte) IEntry
	LenNewEntries() int
	Complete() bool
	VMIndexFor(hash []byte) int
	SortFedServers()
	SortAuditServers()
	SortDBSigs()
	FedServerFor(minute int, hash []byte) IServer
	GetVirtualServers(minute int, identityChainID IHash) (found bool, index int)
	GetFedServerIndexHash(identityChainID IHash) (bool, int)
	GetAuditServerIndexHash(identityChainID IHash) (bool, int)
	MakeMap()
	PrintMap() string
	AddFedServer(identityChainID IHash) int
	AddAuditServer(identityChainID IHash) int
	RemoveFedServerHash(identityChainID IHash)
	RemoveAuditServerHash(identityChainID IHash)
	String() string
	GetDBHeight() uint32
}

type IRequest interface {
	//Key() (thekey [32]byte)
}
