package interfaces

type IProcessList interface {
	Clear()
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
	//GetAck(vmIndex int) IMsg
	//GetAckAt(vmIndex int, height int) IMsg
	HasMessage() bool
	AddOldMsgs(m IMsg)
	DeleteOldMsgs(key IHash)
	GetOldMsgs(key IHash) IMsg
	AddNewEBlocks(key IHash, value IEntryBlock)
	GetNewEBlocks(key IHash) IEntryBlock
	DeleteEBlocks(key IHash)
	AddNewEntry(key IHash, value IEntry)
	DeleteNewEntry(key IHash)
	GetLeaderTimestamp() Timestamp
	ResetDiffSigTally()
	IncrementDiffSigTally()
	CheckDiffSigTally() bool
	Ask(vmIndex int, height int, waitSeconds int64, tag int) int
	TrimVMList(height uint32, vmIndex int)
	//Process(state IState) (progress bool)
	//AddToProcessList(ack IMsg, m IMsg)
	ContainsDBSig(serverID IHash) bool
	AddDBSig(serverID IHash, sig IFullSignature)
	String() string
	IsPendingChainHead(chainID IHash) bool
	GetDBHeight() uint32
}

type IRequest interface {
	//Key() (thekey [32]byte)
}
