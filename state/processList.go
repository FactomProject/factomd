package state

import (
	"fmt"
	"sync"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
)

var _ = fmt.Print

type ProcessLists struct {
	State        *State // Pointer to the state object
	DBHeightBase uint32 // Height of the first Process List in this structure.
	ListsMutex   *sync.Mutex
	Lists        []*ProcessList // Pointer to the ProcessList structure for each DBHeight under construction
}

func (lists *ProcessLists) UpdateState() {
	lists.ListsMutex.Lock()
	defer lists.ListsMutex.Unlock()

	// First let's start at the lowest Process List not yet complete.
	dbstate := lists.State.DBStates.Last()
	if dbstate == nil {
		return
	}
	heightBuilding := dbstate.DirectoryBlock.GetHeader().GetDBHeight() + 1
	pl := lists.Get(heightBuilding)
	// Create DState blocks for all completed Process Lists
	pl.Process(lists.State)
	if pl.Complete() {
		lists.State.DBStates.NewDBState(true, pl.DirectoryBlock, pl.AdminBlock, pl.FactoidBlock, pl.EntryCreditBlock)
	}
}

func (lists *ProcessLists) Get(dbheight uint32) *ProcessList {

	i := int(dbheight) - int(lists.DBHeightBase)
	if i < 0 {
		return nil
	}
	for len(lists.Lists) <= i {
		lists.Lists = append(lists.Lists, nil)
	}
	pl := lists.Lists[i]
	if pl == nil {
		pl = NewProcessList(lists.State.GetTotalServers(), dbheight)
		lists.Lists[i] = pl
	}
	return pl
}

type ProcessList struct {
	DBHeight uint32 // The directory block height for these lists
	Servers  []ListServer

	Acks *map[[32]byte]interfaces.IMsg // acknowlegments by hash
	Msgs *map[[32]byte]interfaces.IMsg // messages by hash

	// Maps
	// ====
	// For Follower

	NewEBlocksSem *sync.Mutex
	NewEBlocks    map[[32]byte]interfaces.IEntryBlock // Entry Blocks added within 10 minutes (follower and leader)

	CommitsSem *sync.Mutex
	Commits    map[[32]byte]interfaces.IMsg // Used by the leader, validate

	// Lists
	// =====
	//
	// The index into the Matryoshka Index handed off to the network by this server.
	MatryoshkaIndex int
	AuditServers    []interfaces.IServer   // List of Audit Servers
	ServerOrder     [][]interfaces.IServer // 10 lists for Server Order for each minute
	FedServers      []interfaces.IServer   // List of Federated Servers
	// Index of this server in the FedServers list, if this is a Federated Server
	ServerIndex int

	// State information about the directory block while it is under construction.  We may
	// have to start building the next block while still building the previous block.
	FactoidBlock     interfaces.IFBlock
	AdminBlock       interfaces.IAdminBlock
	EntryCreditBlock interfaces.IEntryCreditBlock
	DirectoryBlock   interfaces.IDirectoryBlock
}

type ListServer struct {
	List        []interfaces.IMsg // Lists of acknowledged messages
	Height      int               // Height of messages that have been processed
	EomComplete bool              // Lists that are end of minute complete
	SigComplete bool              // Lists that are signature complete
}

func (p *ProcessList) GetLen(list int) int {
	if list >= len(p.Servers) {
		return -1
	}
	l := len(p.Servers[list].List)
	return l
}

func (p *ProcessList) SetSigComplete(value bool) {
	p.Servers[p.ServerIndex].SigComplete = value
}

func (p *ProcessList) SetEomComplete(value bool) {
	p.Servers[p.ServerIndex].EomComplete = value
}

func (p *ProcessList) GetNewEBlocks(key [32]byte) interfaces.IEntryBlock {
	p.NewEBlocksSem.Lock()
	defer p.NewEBlocksSem.Unlock()

	eb := p.NewEBlocks[key]
	return eb
}

func (p *ProcessList) GetCommits(key [32]byte) interfaces.IMsg {
	p.CommitsSem.Lock()
	defer p.CommitsSem.Unlock()

	c := p.Commits[key]
	return c
}

func (p *ProcessList) PutNewEBlocks(dbheight uint32, key interfaces.IHash, value interfaces.IEntryBlock) {
	p.NewEBlocksSem.Lock()
	defer p.NewEBlocksSem.Unlock()

	p.NewEBlocks[key.Fixed()] = value

}

// Test if the process list is complete.  Return true if all messages
// have been recieved, and we have all the signaures for the directory blocks.
func (p *ProcessList) Complete() bool {
	if p == nil {
		return true
	}
	for _, c := range p.Servers {
		if !c.SigComplete {
			return false
		}
	}
	return true
}

// When we begin building on a Process List, we start it.  That marks everything
// as needing to be complete.  When we get all the messages we need, then Complete() will
// return true, because each process list will be signed off.
func (p *ProcessList) SetComplete(v bool) {
	if p == nil {
		return
	}
	for i, _ := range p.Servers {
		p.Servers[i].SigComplete = v
	}
}

// Return true if the process list is complete
func (p *ProcessList) Process(state interfaces.IState) {
	for i := 0; i < len(p.Servers); i++ {
		plist := p.Servers[i].List
		for j := p.Servers[i].Height; j < len(plist); j++ {
			if plist[j] == nil {
				break
			}
			p.Servers[i].Height = j + 1         // Don't process it again.
			plist[j].Process(p.DBHeight, state) // Process this entry

			eom, ok := plist[j].(*messages.EOM)
			if ok && eom.Minute == 9 {
				p.Servers[i].EomComplete = true
			}
			_, ok = plist[j].(*messages.DirectoryBlockSignature)
			if ok {
				p.Servers[i].SigComplete = true
			}

		}
	}
}

func (p *ProcessList) AddToProcessList(ack *messages.Ack, m interfaces.IMsg) {
	processlist := p.Servers[ack.ServerIndex].List
	for len(processlist) <= int(ack.Height) {
		processlist = append(processlist, nil)
	}
	processlist[ack.Height] = m
	p.Servers[ack.ServerIndex].List = processlist
}

func (p *ProcessList) PutCommits(key interfaces.IHash, value interfaces.IMsg) {
	p.CommitsSem.Lock()
	{
		cmsg, ok := value.(interfaces.ICounted)
		if ok {
			v := p.Commits[key.Fixed()]
			if v != nil {
				_, ok := v.(interfaces.ICounted)
				if ok {
					cmsg.SetCount(v.(interfaces.ICounted).GetCount() + 1)
				} else {
					fmt.Println(v)
					panic("Should never happen")
				}
			}
		}

		p.Commits[key.Fixed()] = value
	}
	p.CommitsSem.Unlock()
}

/************************************************
 * Support
 ************************************************/

func NewProcessLists(state interfaces.IState) *ProcessLists {

	pls := new(ProcessLists)

	s, ok := state.(*State)
	if !ok {
		panic("Failed to initalize Process Lists because the wrong state object was used")
	}
	pls.State = s
	pls.DBHeightBase = 0
	pls.ListsMutex = new(sync.Mutex)
	pls.Lists = make([]*ProcessList, 0)

	return pls
}

func NewProcessList(totalServers int, dbheight uint32) *ProcessList {
	// We default to the number of Servers previous.   That's because we always
	// allocate the FUTURE directoryblock, not the current or previous...

	fmt.Println("total Servers ", totalServers)

	pl := new(ProcessList)

	pl.Servers = make([]ListServer, totalServers)

	pl.DBHeight = dbheight
	pl.Acks = new(map[[32]byte]interfaces.IMsg)
	pl.Msgs = new(map[[32]byte]interfaces.IMsg)

	pl.NewEBlocksSem = new(sync.Mutex)
	pl.NewEBlocks = make(map[[32]byte]interfaces.IEntryBlock)

	pl.CommitsSem = new(sync.Mutex)
	pl.Commits = make(map[[32]byte]interfaces.IMsg)

	// If a federated server, this is the server index, which is our index in the FedServers list

	pl.AuditServers = make([]interfaces.IServer, 0)
	pl.FedServers = make([]interfaces.IServer, 0)
	pl.ServerOrder = make([][]interfaces.IServer, 0)

	return pl
}
