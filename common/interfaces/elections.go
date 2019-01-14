package interfaces

type IElections interface {
	GetFedID() IHash
	GetElecting() int
	GetVMIndex() int
	GetRound() []int
	GetFederatedServers() []IServer
	GetAuditServers() []IServer
	GetAdapter() IElectionAdapter
	String() string
}

type IElectionAdapter interface {
	Execute(IMsg) IMsg
	GetDBHeight() int
	GetMinute() int
	GetElecting() int
	GetVMIndex() int

	MessageLists() string
	Status() string
	VolunteerControlsStatus() string

	// An observer does not participate in election voting
	IsObserver() bool
	SetObserver(o bool)

	// Processed indicates the election swap happened
	IsElectionProcessed() bool
	SetElectionProcessed(swapped bool)
	IsStateProcessed() bool
	SetStateProcessed(swapped bool)

	GetAudits() []IHash
}

type IElectionMsg interface {
	IMsg
	ElectionProcess(IState, IElections)
	ElectionValidate(IElections) int
}

type IFedVoteMsg interface {
	ComparisonMinute() int
}

type ISignableElectionMsg interface {
	IElectionMsg
	Signable
	GetVolunteerMessage() ISignableElectionMsg
}

type IElectionsFactory interface {
	// Messages
	NewAddLeaderInternal(Name string, dbheight uint32, serverID IHash) IMsg
	NewAddAuditInternal(name string, dbheight uint32, serverID IHash) IMsg
	NewRemoveLeaderInternal(name string, dbheight uint32, serverID IHash) IMsg
	NewRemoveAuditInternal(name string, dbheight uint32, serverID IHash) IMsg
	NewEomSigInternal(name string, dbheight uint32, minute uint32, vmIndex int, height uint32, serverID IHash) IMsg
	NewDBSigSigInternal(name string, dbheight uint32, minute uint32, vmIndex int, height uint32, serverID IHash) IMsg
	NewAuthorityListInternal(feds []IServer, auds []IServer, height uint32) IMsg

	//
	//	NewElectionAdapter(el IElections) IElectionAdapter
}
