package state

import (
	"fmt"
	"log"

	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
)

var _ = fmt.Print
var _ = log.Print

type ProcessLists struct {
	State        *State         // Pointer to the state object
	DBHeightBase uint32         // Height of the first Process List in this structure.
	Lists        []*ProcessList // Pointer to the ProcessList structure for each DBHeight under construction
}

// Returns the height of the Process List under construction.  There
// can be another list under construction (because of missing messages.
func (lists *ProcessLists) GetDBHeight() uint32 {
	// First let's start at the lowest Process List not yet complete.
	length := len(lists.Lists)
	if length == 0 { return 0 }
	last := lists.Lists[length-1]
	if last == nil {
		return 0
	}
	return last.DBHeight
}


func (lists *ProcessLists) UpdateState() {
	
	heightBuilding := lists.GetDBHeight()

	if heightBuilding == 0 { 
		return
	}

	dbstate := lists.State.DBStates.Last()
	
	pl := lists.Get(heightBuilding)

	diff := heightBuilding - lists.DBHeightBase
	if diff > 0 {
		lists.DBHeightBase += diff
		lists.Lists = lists.Lists[diff:]
	}

	//*******************************************************************//
	// Do initialization of blocks for the next Process List level here
	//*******************************************************************
	if pl.DirectoryBlock == nil {
		pl.DirectoryBlock = directoryBlock.NewDirectoryBlock(heightBuilding, nil)
		pl.FactoidBlock = lists.State.GetFactoidState().GetCurrentBlock()
		pl.AdminBlock = lists.State.NewAdminBlock(heightBuilding)
		var err error
		pl.EntryCreditBlock, err = entryCreditBlock.NextECBlock(dbstate.EntryCreditBlock)
		if err != nil {
			panic(err.Error())
		}

	}
	// Create DState blocks for all completed Process Lists
	pl.Process(lists.State)

	lastHeight := dbstate.DirectoryBlock.GetHeader().GetDBHeight()
	// Only when we are sig complete that we can move on.
	if pl.Complete() &&  lastHeight+1 == heightBuilding {
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
		pl = NewProcessList(lists.State, lists.State.GetTotalServers(), dbheight)
		lists.Lists[i] = pl
	}
	return pl
}

type ProcessList struct {
	DBHeight uint32 // The directory block height for these lists
	State    interfaces.IState
	Servers  []*ListServer

	Acks *map[[32]byte]interfaces.IMsg // acknowlegments by hash
	Msgs *map[[32]byte]interfaces.IMsg // messages by hash

	// Maps
	// ====
	// For Follower

	NewEBlocks map[[32]byte]interfaces.IEntryBlock // Entry Blocks added within 10 minutes (follower and leader)
	Commits    map[[32]byte]interfaces.IMsg        // Used by the leader, validate

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

func (p *ProcessList) GetCommits(key interfaces.IHash) interfaces.IMsg {
	c := p.Commits[key.Fixed()]
	return c
}

func (p *ProcessList) PutCommits(key interfaces.IHash, value interfaces.IMsg) {

	{
		cmsg, ok := value.(interfaces.ICounted)
		if ok {
			v := p.Commits[key.Fixed()]
			if v != nil {
				_, ok := v.(interfaces.ICounted)
				if ok {
					cmsg.SetCount(v.(interfaces.ICounted).GetCount() + 1)
				} else {
					p.State.Println(v)
					panic("Should never happen")
				}
			}
		}

		p.Commits[key.Fixed()] = value
	}
}

func (p *ProcessList) GetNewEBlocks(key interfaces.IHash) interfaces.IEntryBlock {

	eb := p.NewEBlocks[key.Fixed()]
	return eb
}

func (p *ProcessList) PutNewEBlocks(dbheight uint32, key interfaces.IHash, value interfaces.IEntryBlock) {

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

// Process messages and update our state.
func (p *ProcessList) Process(state *State) {
	
	for i := 0; i < len(p.Servers); i++ {
		plist := p.Servers[i].List
		//fmt.Println("Process List: DBHEight, height in list, len(plist)", p.DBHeight, p.Servers[i].Height, len(plist))
		for j := p.Servers[i].Height; !p.Servers[i].SigComplete && j < len(plist); j++ {
			if plist[j] == nil {
				p.State.Println("!!!!!!! Missing entry in process list at", j)
				return
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
	if p == nil || p.Servers[ack.ServerIndex].List == nil { panic("This should not happen")}
	
	for len(p.Servers[ack.ServerIndex].List) <= int(ack.Height) {
		p.Servers[ack.ServerIndex].List = append(p.Servers[ack.ServerIndex].List, nil)
	}
	p.Servers[ack.ServerIndex].List[ack.Height] = m
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
	pls.Lists = make([]*ProcessList, 0)

	return pls
}

func NewProcessList(state interfaces.IState, totalServers int, dbheight uint32) *ProcessList {
	// We default to the number of Servers previous.   That's because we always
	// allocate the FUTURE directoryblock, not the current or previous...
	
	pl := new(ProcessList)

	pl.State   = state
	pl.Servers = make([]*ListServer, totalServers)
	for i:=0; i< totalServers; i++ {
		pl.Servers[i]=new(ListServer)
		pl.Servers[i].List = make([]interfaces.IMsg,0)
	}
	pl.DBHeight = dbheight
	pl.Acks = new(map[[32]byte]interfaces.IMsg)
	pl.Msgs = new(map[[32]byte]interfaces.IMsg)

	pl.NewEBlocks = make(map[[32]byte]interfaces.IEntryBlock)
	pl.Commits = make(map[[32]byte]interfaces.IMsg)

	// If a federated server, this is the server index, which is our index in the FedServers list

	pl.AuditServers = make([]interfaces.IServer, 0)
	pl.FedServers = make([]interfaces.IServer, 0)
	pl.ServerOrder = make([][]interfaces.IServer, 0)

	return pl
}
