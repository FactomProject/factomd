package state

import (
	"bytes"
	"fmt"
	"github.com/FactomProject/factomd/common/directoryBlock"
	//"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"log"
)

var _ = fmt.Print
var _ = log.Print

type ProcessList struct {
	DBHeight uint32 // The directory block height for these lists
	good     bool   // Means we have the previous blocks, so we can process!

	// List of messsages that came in before the previous block was built
	// We can not completely validate these messages until the previous block
	// is built.
	MsgQueue []interfaces.IMsg

	State   interfaces.IState
	Servers []*ListServer

	// Maps
	// ====
	OldMsgs map[[32]byte]interfaces.IMsg // messages processed in this list
	OldAcks map[[32]byte]interfaces.IMsg // messages processed in this list

	NewEBlocks map[[32]byte]interfaces.IEntryBlock // Entry Blocks added within 10 minutes (follower and leader)
	Commits    map[[32]byte]interfaces.IMsg        // Used by the leader, validate

	// Lists
	// =====
	//
	// The index into the Matryoshka Index handed off to the network by this server.
	MatryoshkaIndex int
	AuditServers    []interfaces.IFctServer   // List of Audit Servers
	ServerOrder     [][]interfaces.IFctServer // 10 lists for Server Order for each minute
	FedServers      []interfaces.IFctServer   // List of Federated Servers

	// State information about the directory block while it is under construction.  We may
	// have to start building the next block while still building the previous block.
	AdminBlock       interfaces.IAdminBlock
	EntryCreditBlock interfaces.IEntryCreditBlock
	DirectoryBlock   interfaces.IDirectoryBlock
}

type ListServer struct {
	List        []interfaces.IMsg // Lists of acknowledged messages
	Height      int               // Height of messages that have been processed
	EomComplete bool              // Lists that are end of minute complete
	SigComplete bool              // Lists that are signature complete
	LastAck     interfaces.IMsg   // The last Acknowledgement set by this server
}

// Given a server index, return the last Ack
func (p *ProcessList) GetLastAck(index int) interfaces.IMsg {
	return p.Servers[index].LastAck
}

// Given a server index, return the last Ack
func (p *ProcessList) SetLastAck(index int, msg interfaces.IMsg) error {
	// Check the hash of the previous msg before we over write
	p.Servers[index].LastAck = msg
	return nil
}

// Add the given serverChain to this processlist, and return the server index number of the
// added server
func (p *ProcessList) AddFedServer(server interfaces.IFctServer) int {
	found, i := p.GetFedServerIndex(server.GetChainID())
	if server.GetChainID().Bytes()[0] == 0 && server.GetChainID().Bytes()[1] == 0 {
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
func (p *ProcessList) RemoveFedServer(server interfaces.IFctServer) {
	found, i := p.GetFedServerIndex(server.GetChainID())
	if !found {
		return
	}
	p.FedServers = append(p.FedServers[:i], p.FedServers[i+1:]...)
}

// Returns true and the index of this server, or false and the insertion point for this server
func (p *ProcessList) GetFedServerIndex(serverChainID interfaces.IHash) (bool, int) {
	scid := serverChainID.Bytes()
	if p == nil || p.FedServers == nil {
		if p == nil {
			return false, 0
		}
		if bytes.Compare(scid, p.State.GetCoreChainID().Bytes()) == 0 {
			return true, 0
		} else {
			return false, 0
		}
	}
	if len(p.FedServers) == 0 {
		server := new(interfaces.Server)
		server.ChainID = p.State.GetCoreChainID()
		p.FedServers = append(p.FedServers, server)
	}
	for i, fs := range p.FedServers {
		// Find and remove
		switch bytes.Compare(scid, fs.GetChainID().Bytes()) {
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

	for _, ls := range p.Servers {
		if len(ls.List) > 0 {
			return true
		}
	}

	return false
}

func (p *ProcessList) GetNewEBlocks(key interfaces.IHash) interfaces.IEntryBlock {

	eb := p.NewEBlocks[key.Fixed()]
	return eb
}

func (p *ProcessList) PutNewEBlocks(dbheight uint32, key interfaces.IHash, value interfaces.IEntryBlock) {

	p.NewEBlocks[key.Fixed()] = value

}

// TODO:  Need to map the server identity to the process list for which it
// is responsible.  Right now, works with only one server!
func (p *ProcessList) SetSigComplete(value bool) {
	found, i := p.GetFedServerIndex(p.State.GetIdentityChainID())
	if !found {
		return
	}
	p.Servers[i].SigComplete = value
}

// Set the EomComplete for the ith list
func (p *ProcessList) SetEomComplete(i int, value bool) {
	found, i := p.GetFedServerIndex(p.State.GetIdentityChainID())
	if !found {
		return
	}
	p.Servers[i].EomComplete = value
}

// Test if a process list for a server is EOM complete.  Return true if all messages
// have been recieved, and we just need the signaure.  If we need EOM messages, or we
// have all EOM messages and we have the Signature, then we return false.
func (p *ProcessList) EomComplete() bool {
	if p == nil {
		return true
	}
	for _, c := range p.Servers {
		if !c.EomComplete {
			return false
		}
	}
	return true
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
func (p *ProcessList) SetComplete(index int, v bool) {
	if p == nil {
		return
	}
	p.Servers[index].SigComplete = v
}

// Process messages and update our state.
func (p *ProcessList) Process(state *State) {

	if !p.good { // If we don't know this process list is good...
		last := state.DBStates.Last() // Get our last state.
		if last == nil {
			return
		}
		lht := last.DirectoryBlock.GetHeader().GetDBHeight()
		if last.Saved && lht == p.DBHeight-1 {
			p.good = true
		} else {
			//fmt.Println("ht/lht: ", p.DBHeight, " ", lht, " ", last.Saved)
			return
		}
	}

	for i := 0; i < len(p.Servers); i++ {
		plist := p.Servers[i].List

		// state.Println("Process List: DBHeight, height in list, len(plist)", p.DBHeight, "/", p.Servers[i].Height, "/", len(plist))

		for j := p.Servers[i].Height; j < len(plist); j++ {
			if plist[j] == nil {
				p.State.Println("!!!!!!! Missing entry in process list at", j)
				return
			}

			if oldAck, ok := p.OldAcks[plist[j].GetHash().Fixed()]; ok {
				if thisAck, ok := oldAck.(*messages.Ack); ok {
					if thisAck.GetOrigin() != i {
						// if the acknowledgement didn't originate on this server
						// it may be invalid
						var expectedSerialHash interfaces.IHash
						var err error
						last, ok := p.GetLastAck(i).(*messages.Ack)
						if !ok {
							expectedSerialHash = thisAck.MessageHash
						} else {
							expectedSerialHash, err = primitives.CreateHash(last.MessageHash, thisAck.MessageHash)
							if err != nil {
								// cannot create a expectedSerialHash to compare to
								plist[j] = nil
								return
							}
						}
						// compare the SerialHash of this acknowledgement with the
						// expected serialHash (generated above)
						if !expectedSerialHash.IsSameAs(thisAck.SerialHash) {
							// the SerialHash of this acknowledgment is incorrect
							// according to this node's processList
							plist[j] = nil
							return
						}
					}
					p.SetLastAck(i, thisAck)
				} else {
					// the message from OldAcks is not actually of type Ack
					plist[j] = nil
					return
				}
			} else {
				// corresponding acknowledgement not found
				p.State.Println("!!!!!!! Missing acknowledgement in process list for", j)
				plist[j] = nil
				return
			}

			if plist[j].Process(p.DBHeight, state) { // Try and Process this entry
				p.Servers[i].Height = j + 1 // Don't process it again if the process worked.
			}

			// TODO:  If we carefully manage our state as we process messages, we
			// would not need to check the messages here!  Checking for EOM and DBS
			// as follows is a bit of a kludge.
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
	//p.State.Println("AddToProcessList +++++++++++++++++++++++++++++++++++++++++++++++",
	//				m.String(),
	//				" ",
	//			 len(p.Servers[ack.ServerIndex].List))
	if p == nil || p.Servers[ack.ServerIndex].List == nil {
		panic("This should not happen")
	}

	for len(p.Servers[ack.ServerIndex].List) <= int(ack.Height) {
		p.Servers[ack.ServerIndex].List = append(p.Servers[ack.ServerIndex].List, nil)
	}
	p.Servers[ack.ServerIndex].List[ack.Height] = m
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

func (p *ProcessList) String() string {
	prt := ""
	if p == nil {
		prt = "-- <nil>\n"
	} else {
		prt = p.State.GetFactomNodeName() + "\n"
		for _, id := range p.FedServers {
			prt = fmt.Sprintf("%s   %x\n", prt, id.GetChainID().Bytes())
		}
		for i, server := range p.Servers {
			prt = prt + fmt.Sprintf("  Server %d \n", i)
			for _, msg := range server.List {
				if msg != nil {
					prt = prt + "   " + msg.String() + "\n"
				} else {
					prt = prt + "   <nil>\n"
				}
			}
			prt = prt + "\n"
		}
	}
	return prt
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

	pl.OldMsgs = make(map[[32]byte]interfaces.IMsg)
	pl.OldAcks = make(map[[32]byte]interfaces.IMsg)

	pl.NewEBlocks = make(map[[32]byte]interfaces.IEntryBlock)
	pl.Commits = make(map[[32]byte]interfaces.IMsg)

	// If a federated server, this is the server index, which is our index in the FedServers list

	pl.AuditServers = make([]interfaces.IFctServer, 0)
	pl.FedServers = make([]interfaces.IFctServer, 0)
	pl.ServerOrder = make([][]interfaces.IFctServer, 0)

	s := state.(*State)
	var err error

	pl.DirectoryBlock = directoryBlock.NewDirectoryBlock(dbheight, nil)
	pl.AdminBlock = s.NewAdminBlock(dbheight)
	pl.EntryCreditBlock, err = entryCreditBlock.NextECBlock(nil)

	if err != nil {
		panic(err.Error())
	}

	return pl
}
