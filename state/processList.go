package state

import (
	"bytes"
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"log"
)

var _ = fmt.Print
var _ = log.Print

type ProcessList struct {
	DBHeight uint32 // The directory block height for these lists
	
	// List of messsages that came in before the previous block was built
	// We can not completely validate these messages until the previous block
	// is built.
	MsgsQueue	[] interfaces.IMsg			
	
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

// Add the given serverChain to this processlist, and return the server index number of the
// added server
func (p *ProcessList) AddFedServer(server *interfaces.Server) int {
	found, i := p.GetFedServerIndex(server.ChainID)
	if server.ChainID.Bytes()[0] == 0 && server.ChainID.Bytes()[1] == 0 {
		panic("Grrr")
	}
	if found {
		return i
	}
	p.FedServers = append(p.FedServers, nil)
	copy(p.FedServers[i+1:], p.FedServers[i:])
	p.FedServers[i] = server
	return i
}

// Add the given serverChain to this processlist, and return the server index number of the
// added server
func (p *ProcessList) RemoveFedServer(server *interfaces.Server) {
	found, i := p.GetFedServerIndex(server.ChainID)
	if !found {
		return
	}
	p.FedServers = append(p.FedServers[:i], p.FedServers[i+1:]...)
}

// Returns true and the index of this server, or false and the insertion point for this server
func (p *ProcessList) GetFedServerIndex(serverChainID interfaces.IHash) (bool, int) {
	scid := serverChainID.Bytes()
	if p == nil || p.FedServers == nil {
		return false, 0
	}
	for i, ifs := range p.FedServers {
		fs := ifs.(*interfaces.Server)
		// Find and remove
		switch bytes.Compare(scid, fs.ChainID.Bytes()) {
		case 0: // Found the ID!
			return true, i
		case -1: // Past the ID, can't be in list
			return false, i
		}
	}
	return false, len(p.FedServers)
}

func (p *ProcessList) GetLen(list int) int {
	if list >= len(p.Servers) {
		return -1
	}
	l := len(p.Servers[list].List)
	return l
}

func (p ProcessList) HasMessage() bool {
	if len(*p.Acks) > 0 { return true }
	if len(*p.Msgs) > 0 { return true }
	
	for _,ls := range p.Servers {
		if len(ls.List) > 0 {
			return true
		}
	}
	
	return false
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
				//p.State.Println("!!!!!!! Missing entry in process list at", j)
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
	if p == nil || p.Servers[ack.ServerIndex].List == nil {
		panic("This should not happen")
	}

	for len(p.Servers[ack.ServerIndex].List) <= int(ack.Height) {
		p.Servers[ack.ServerIndex].List = append(p.Servers[ack.ServerIndex].List, nil)
	}
	p.Servers[ack.ServerIndex].List[ack.Height] = m
}

/************************************************
 * Support
 ************************************************/

func NewProcessList(state interfaces.IState, totalServers int, dbheight uint32) *ProcessList {
	// We default to the number of Servers previous.   That's because we always
	// allocate the FUTURE directoryblock, not the current or previous...

	pl := new(ProcessList)

	pl.State = state
	pl.Servers = make([]*ListServer, totalServers)
	for i := 0; i < totalServers; i++ {
		pl.Servers[i] = new(ListServer)
		pl.Servers[i].List = make([]interfaces.IMsg, 0)

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
