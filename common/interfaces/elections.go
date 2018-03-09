package interfaces

type IElections interface {
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
}

type IElectionMsg interface {
	ElectionProcess(IState, IElections)
	String() string
}

type IElectionsFactory interface {
	// Messages
	NewAddLeaderInternal(Name string, dbheight uint32, serverID IHash) IMsg
	NewAddAuditInternal(name string, dbheight uint32, serverID IHash) IMsg
	NewRemoveLeaderInternal(name string, dbheight uint32, serverID IHash) IMsg
	NewRemoveAuditInternal(name string, dbheight uint32, serverID IHash) IMsg
	NewEomSigInternal(name string, dbheight uint32, minute uint32, vmIndex int, height uint32, serverID IHash) IMsg
	NewDBSigSigInternal(name string, dbheight uint32, minute uint32, vmIndex int, height uint32, serverID IHash) IMsg

	//
	NewElectionAdapter(el IElections) IElectionAdapter
}
