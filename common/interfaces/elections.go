package interfaces

type IElections interface {
	ElectionProcess(IState, IElections)
	NewAddLeaderInternal(Name string, dbheight uint32, serverID IHash) IMsg
	NewAddAuditInternal(name string, dbheight uint32, serverID IHash) IMsg
	NewRemoveLeaderInternal(name string, dbheight uint32, serverID IHash) IMsg
	NewRemoveAuditInternal(name string, dbheight uint32, serverID IHash) IMsg
}
