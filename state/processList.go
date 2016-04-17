package state

import (
	"fmt"

	"github.com/FactomProject/factomd/common/directoryBlock"
	//"github.com/FactomProject/factomd/common/factoid"
	"log"
    "bytes"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
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

	State         interfaces.IState
	NumberServers int           // How many servers we are tracking
	Servers       []*ListServer // Process list for each server (up to 32)

	// Maps
	// ====
	OldMsgs map[[32]byte]interfaces.IMsg // messages processed in this list
	OldAcks map[[32]byte]interfaces.IMsg // messages processed in this list

	NewEBlocks map[[32]byte]interfaces.IEntryBlock // Entry Blocks added within 10 minutes (follower and leader)
	NewEntries map[[32]byte]interfaces.IEntry      // Entries added within 10 minutes (follower and leader)
	Commits    map[[32]byte]interfaces.IMsg        // Used by the leader, validate

	// State information about the directory block while it is under construction.  We may
	// have to start building the next block while still building the previous block.
	AdminBlock       interfaces.IAdminBlock
	EntryCreditBlock interfaces.IEntryCreditBlock
	DirectoryBlock   interfaces.IDirectoryBlock
    
    // Number of Servers acknowledged by Factom
	Matryoshka   []interfaces.IHash        // Reverse Hash
	AuditServers []interfaces.IFctServer   // List of Audit Servers
	ServerOrder  [][]interfaces.IFctServer // 10 lists for Server Order for each minute
	FedServers   []interfaces.IFctServer   // List of Federated Servers

}

type ListServer struct {
	List          []interfaces.IMsg // Lists of acknowledged messages
	Height        int               // Height of messages that have been processed
	EomComplete   bool              // Lists that are end of minute complete
	SigComplete   bool              // Lists that are signature complete
	Undo          interfaces.IMsg   // The Leader needs one level of undo to handle DB Sigs.
    LastLeaderAck interfaces.IMsg   // The last Acknowledgement set by this leader
	LastAck       interfaces.IMsg   // The last Acknowledgement set by this follower
}

// Returns true and the index of this server, or false and the insertion point for this server
func (p *ProcessList) GetFedServerIndexHash(identityChainID interfaces.IHash) (bool, int) {
	scid := identityChainID.Bytes()

	for i, fs := range p.FedServers {
		// Find and remove
		if bytes.Compare(scid, fs.GetChainID().Bytes()) == 0 {
			return true, i
		}
	}
	return false, len(p.FedServers)
}

// Add the given serverChain to this processlist, and return the server index number of the
// added server
func (s *ProcessList) AddFedServer(identityChainID interfaces.IHash) int {
	found, i := s.GetFedServerIndexHash(identityChainID)
	if found {
		return i
	}
	s.FedServers = append(s.FedServers, nil)
	copy(s.FedServers[i+1:], s.FedServers[i:])
	s.FedServers[i] = &interfaces.Server{ChainID: identityChainID}
	return i
}

// Add the given serverChain to this processlist, and return the server index number of the
// added server
func (p *ProcessList) RemoveFedServerHash(identityChainID interfaces.IHash) {
	found, i := p.GetFedServerIndexHash(identityChainID)
	if !found {
		return
	}
	p.FedServers = append(p.FedServers[:i], p.FedServers[i+1:]...)
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

// Given a server index, return the last Ack
func (p *ProcessList) GetLastLeaderAck(index int) interfaces.IMsg {
	return p.Servers[index].LastLeaderAck
}

// Given a server index, return the last Ack
func (p *ProcessList) SetLastLeaderAck(index int, msg interfaces.IMsg) error {
	// Check the hash of the previous msg before we over write
    p.Servers[index].Undo = p.Servers[index].LastLeaderAck
	p.Servers[index].LastLeaderAck = msg
	return nil
}

func (p *ProcessList) UndoLeaderAck(index int) {
    p.Servers[index].LastLeaderAck = p.Servers[index].Undo
}

func (p *ProcessList) GetLen(list int) int {
	if list >= p.NumberServers {
		return -1
	}
	l := len(p.Servers[list].List)
	return l
}

func (p ProcessList) HasMessage() bool {

	for i := 0; i < p.NumberServers; i++ {
		if len(p.Servers[i].List) > 0 {
			return true
		}
	}

	return false
}

func (p *ProcessList) GetNewEBlocks(key interfaces.IHash) interfaces.IEntryBlock {
	return p.NewEBlocks[key.Fixed()]
}

func (p *ProcessList) PutNewEBlocks(dbheight uint32, key interfaces.IHash, value interfaces.IEntryBlock) {
	p.NewEBlocks[key.Fixed()] = value
}

func (p *ProcessList) PutNewEntries(dbheight uint32, key interfaces.IHash, value interfaces.IEntry) {
	p.NewEntries[key.Fixed()] = value
}

// TODO:  Need to map the server identity to the process list for which it
// is responsible.  Right now, works with only one server!
func (p *ProcessList) SetSigComplete(i int, value bool) {
	p.Servers[i].SigComplete = value
}

// Set the EomComplete for the ith list
func (p *ProcessList) SetEomComplete(i int, value bool) {
	p.Servers[i].EomComplete = value
}

// Test if a process list for a server is EOM complete.  Return true if all messages
// have been recieved, and we just need the signaure.  If we need EOM messages, or we
// have all EOM messages and we have the Signature, then we return false.
func (p *ProcessList) EomComplete() bool {
	if p == nil {
		return true
	}
	n := len(p.State.GetFedServers(p.DBHeight))
	for i := 0; i < n; i++ {
		c := p.Servers[i]
		if !c.EomComplete {
			return false
		}
	}
	return true
}

// Test if the process list is complete.  Return true if all messages
// have been recieved, and we have all the signaures for the directory blocks.
func (p *ProcessList) SigComplete() bool {
	if p == nil {
		return true
	}
	n := len(p.State.GetFedServers(p.DBHeight))
	for i := 0; i < n; i++ {
		c := p.Servers[i]
		if !c.SigComplete {
			return false
		}
	}
	return true
}

// Process messages and update our state.
func (p *ProcessList) Process(state *State) {

	if !p.good { // If we don't know this process list is good...
		last := state.DBStates.Last() // Get our last state.
		if last == nil {
			return
		}
		lht := last.DirectoryBlock.GetHeader().GetDBHeight()
		if last.Saved && lht >= p.DBHeight-1 {
			p.good = true
			p.NumberServers = len(p.FedServers)
		} else {
			//fmt.Println("ht/lht: ", p.DBHeight, " ", lht, " ", last.Saved)
			return
		}
	}

	for i := 0; i < p.NumberServers; i++ {
		plist := p.Servers[i].List

		for j := p.Servers[i].Height; j < len(plist); j++ {
			if plist[j] == nil {
				if !state.IsThrottled {
					missingMsgRequest := messages.NewMissingMsg(state, p.DBHeight, uint32(j))
					if missingMsgRequest != nil {
						state.NetworkOutMsgQueue() <- missingMsgRequest
					}
					p.State.Println("!!!!!!! Missing entry in process list at", j)
					state.IsThrottled = true
				}
				return
			}

			if oldAck, ok := p.OldAcks[plist[j].GetHash().Fixed()]; ok {
				if thisAck, ok := oldAck.(*messages.Ack); ok {
					var expectedSerialHash interfaces.IHash
					var err error
					last, ok := p.GetLastAck(i).(*messages.Ack)
					if !ok || last.IsSameAs(thisAck) {
						expectedSerialHash = thisAck.SerialHash
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
						fmt.Println("DISCREPANCY: ", i, j, "on", state.GetFactomNodeName())
						fmt.Printf("LAST MESS: %+v ::: LAST SERIAL: %+v\n", last.MessageHash, last.SerialHash)
						fmt.Printf("THIS MESS: %+v ::: THIS SERIAL: %+v\n", thisAck.MessageHash, thisAck.SerialHash)
						fmt.Println("EXPECT: ", expectedSerialHash)
						// the SerialHash of this acknowledgment is incorrect
						// according to this node's processList
						plist[j] = nil
						return
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
			} else {
                break
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

func (p *ProcessList) String() string {
	prt := ""
	if p == nil {
		prt = "-- <nil>\n"
	} else {
		prt = p.State.GetFactomNodeName() + "\n"

		for i, server := range p.Servers {
			if i >= p.NumberServers {
				break
			}
            eom := ""
            sig := ""
            if server.EomComplete {
                eom = "EOM Complete"
            }
            if server.SigComplete {
                sig = "Sig Complete"
            } 
           
			prt = prt + fmt.Sprintf("  Server %d %s %s\n", i,eom,sig)
			for j, msg := range server.List {
                
                if j < server.Height {
                    prt = prt + "  p"
                }else{
                    prt = prt + "   "
                }
            
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

func NewProcessList(state interfaces.IState, previous *ProcessList, dbheight uint32) *ProcessList {
	// We default to the number of Servers previous.   That's because we always
	// allocate the FUTURE directoryblock, not the current or previous...

	pl := new(ProcessList)

	pl.State = state
	pl.Servers = make([]*ListServer, 32)
	for i := 0; i < 32; i++ {
		pl.Servers[i] = new(ListServer)
		pl.Servers[i].List = make([]interfaces.IMsg, 0)

	}
    
    // Make a copy of the previous FedServers
    pl.FedServers = append(pl.FedServers, previous.FedServers ...)
   	pl.AuditServers = append(pl.AuditServers, previous.AuditServers ...)
    pl.ServerOrder = append(pl.ServerOrder, previous.ServerOrder ...)
 

    
	pl.DBHeight = dbheight

	pl.OldMsgs = make(map[[32]byte]interfaces.IMsg)
	pl.OldAcks = make(map[[32]byte]interfaces.IMsg)

	pl.NewEBlocks = make(map[[32]byte]interfaces.IEntryBlock)
	pl.NewEntries = make(map[[32]byte]interfaces.IEntry)
	pl.Commits = make(map[[32]byte]interfaces.IMsg)

	// If a federated server, this is the server index, which is our index in the FedServers list

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
