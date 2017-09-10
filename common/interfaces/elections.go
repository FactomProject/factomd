package interfaces

type IElections interface {
}

type IElectionMsg interface {
	ElectionProcess(IState, IElections)
	String() string
}

type IElectionsFactory interface {
	NewAddLeaderInternal(Name string, dbheight uint32, serverID IHash) IMsg
	NewAddAuditInternal(name string, dbheight uint32, serverID IHash) IMsg
	NewRemoveLeaderInternal(name string, dbheight uint32, serverID IHash) IMsg
	NewRemoveAuditInternal(name string, dbheight uint32, serverID IHash) IMsg
	NewEomSigInternal(name string, dbheight uint32, minute uint32, height uint32, serverID IHash) IMsg
	String() string
}
