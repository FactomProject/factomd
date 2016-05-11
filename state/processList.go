package state

import (
	"fmt"

	"github.com/FactomProject/factomd/common/directoryBlock"
	//"github.com/FactomProject/factomd/common/factoid"
	"bytes"
	"log"

	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"time"
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

	State     interfaces.IState
	VMs       []*VM       // Process list for each server (up to 32)
	ServerMap [10][32]int // Map of FedServers to all Servers for each minute

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
	Matryoshka   []interfaces.IHash      // Reverse Hash
	AuditServers []interfaces.IFctServer // List of Audit Servers
	FedServers   []interfaces.IFctServer // List of Federated Servers

}

type VM struct {
	List           []interfaces.IMsg // Lists of acknowledged messages
	Height         int               // Height of messages that have been processed
	LeaderMinute   int               // Where the leader is in acknowledging messages
	MinuteComplete int               // Highest minute complete (0-9) by the follower
	MinuteHeight   int               // Height of the last minute complete
	SigComplete    bool              // Lists that are signature complete
	Undo           interfaces.IMsg   // The Leader needs one level of undo to handle DB Sigs.
	LastLeaderAck  interfaces.IMsg   // The last Acknowledgement set by this leader
	LastAck        interfaces.IMsg   // The last Acknowledgement set by this follower
	missingTime    int64             // How long we have been waiting for a missing message
}

// Returns the Virtual Server index for this hash for the given minute
func (p *ProcessList) VMIndexFor(hash []byte) int {
	v := uint64(0)
	for _, b := range hash {
		v += uint64(b)
	}
	r := int(v % uint64(len(p.FedServers)))
	return r
}

// Returns the Federated Server responsible for this hash in this minute
func (p *ProcessList) FedServerFor(minute int, hash []byte) interfaces.IFctServer {
	vs := p.VMIndexFor(hash)
	if vs < 0 {
		return nil
	}
	fedIndex := p.ServerMap[minute][vs]
	return p.FedServers[fedIndex]
}

func (p *ProcessList) LeaderFor(chainID interfaces.IHash, hash []byte) int {
	vmIndex := p.VMIndexFor(hash)
	minute := p.VMs[vmIndex].LeaderMinute
	vm := p.FedServers[p.ServerMap[minute][vmIndex]]
	if bytes.Compare(chainID.Bytes(), vm.GetChainID().Bytes()) == 0 {
		return vmIndex
	}
	return -1
}

func (p *ProcessList) GetVirtualServers(minute int, identityChainID interfaces.IHash) (found bool, index int) {
	found, fedIndex := p.GetFedServerIndexHash(identityChainID)
	if !found {
		return false, -1
	}
	// fmt.Println("Line 100 minute:",minute)
	for i, fedix := range p.ServerMap[minute] {
		if i == len(p.FedServers) {
			break
		}
		if fedix == fedIndex {
			return true, i
		}
	}

	return false, -1
}

// Returns true and the index of this server, or false and the insertion point for this server
func (p *ProcessList) GetFedServerIndexHash(identityChainID interfaces.IHash) (bool, int) {

	if p == nil {
		return false, 0
	}

	scid := identityChainID.Bytes()

	for i, fs := range p.FedServers {
		// Find and remove
		comp := bytes.Compare(scid, fs.GetChainID().Bytes())
		if comp == 0 {
			return true, i
		}
		if comp < 0 {
			return false, i
		}
	}
	return false, len(p.FedServers)
}

// Returns true and the index of this server, or false and the insertion point for this server
func (p *ProcessList) GetAuditServerIndexHash(identityChainID interfaces.IHash) (bool, int) {

	if p == nil {
		return false, 0
	}

	scid := identityChainID.Bytes()

	for i, fs := range p.AuditServers {
		// Find and remove
		if bytes.Compare(scid, fs.GetChainID().Bytes()) == 0 {
			return true, i
		}
	}
	return false, len(p.AuditServers)
}

// This function will be replaced by a calculation from the Matryoshka hashes from the servers
// but for now, we are just going to make it a function of the dbheight.
func (p *ProcessList) MakeMap() {
	n := len(p.FedServers)
	indx := int(p.DBHeight*131) % n
	for i := 0; i < 10; i++ {
		indx = (indx + 1) % n
		for j := 0; j < len(p.FedServers); j++ {
			p.ServerMap[i][j] = indx
			indx = (indx + 1) % n
		}
	}
}

// This function will be replaced by a calculation from the Matryoshka hashes from the servers
// but for now, we are just going to make it a function of the dbheight.
func (p *ProcessList) PrintMap() string {
	n := len(p.FedServers)
	prt := " min"
	for i := 0; i < n; i++ {
		prt = fmt.Sprintf("%s%3d", prt, i)
	}
	prt = prt + "\n"
	for i := 0; i < 10; i++ {
		prt = fmt.Sprintf("%s%3d  ", prt, i)
		for j := 0; j < len(p.FedServers); j++ {
			prt = fmt.Sprintf("%s%2d ", prt, p.ServerMap[i][j])
		}
		prt = prt + "\n"
	}
	return prt
}

// Take the minute that has completed.  The minute height then is 1 plus that number
// i.e. the minute height is 0, or 1, or 2, or ... or 10 (all done)
func (p *ProcessList) SetMinute(index int, minute int) {
	p.VMs[index].LeaderMinute = minute
	p.VMs[index].MinuteComplete = minute + 1
	p.VMs[index].MinuteHeight = p.VMs[index].Height
}

// Return the lowest minute number in our lists.  Note that Minute Markers END
// a minute, so After MinuteComplete=0
func (p *ProcessList) MinuteHeight() int {
	m := 10
	for i := 0; i < len(p.FedServers); i++ {
		vm := p.VMs[i]
		if vm.MinuteComplete < m {
			m = vm.MinuteComplete
		}
	}
	return m
}

// Add the given serverChain to this processlist as a Federated Server, and return
// the server index number of the added server
func (p *ProcessList) AddFedServer(identityChainID interfaces.IHash) int {
	found, i := p.GetFedServerIndexHash(identityChainID)
	if found {
		return i
	}
	p.FedServers = append(p.FedServers, nil)
	copy(p.FedServers[i+1:], p.FedServers[i:])
	p.FedServers[i] = &interfaces.Server{ChainID: identityChainID}

	p.MakeMap()

	return i
}

// Add the given serverChain to this processlist as an Audit Server, and return
// the server index number of the added server
func (p *ProcessList) AddAuditServer(identityChainID interfaces.IHash) int {
	found, i := p.GetAuditServerIndexHash(identityChainID)
	if found {
		return i
	}
	p.AuditServers = append(p.AuditServers, nil)
	copy(p.AuditServers[i+1:], p.AuditServers[i:])
	p.AuditServers[i] = &interfaces.Server{ChainID: identityChainID}

	return i
}

// Remove the given serverChain from this processlist's Federated Servers
func (p *ProcessList) RemoveFedServerHash(identityChainID interfaces.IHash) {
	found, i := p.GetFedServerIndexHash(identityChainID)
	if !found {
		return
	}
	p.FedServers = append(p.FedServers[:i], p.FedServers[i+1:]...)
}

// Remove the given serverChain from this processlist's Audit Servers
func (p *ProcessList) RemoveAuditServerHash(identityChainID interfaces.IHash) {
	found, i := p.GetAuditServerIndexHash(identityChainID)
	if !found {
		return
	}
	p.AuditServers = append(p.AuditServers[:i], p.AuditServers[i+1:]...)
}

// Given a server index, return the last Ack
func (p *ProcessList) GetLastAck(index int) interfaces.IMsg {
	return p.VMs[index].LastAck
}

// Given a server index, return the last Ack
func (p *ProcessList) SetLastAck(index int, msg interfaces.IMsg) error {
	// Check the hash of the previous msg before we over write
	p.VMs[index].LastAck = msg
	return nil
}

// Given a server index, return the last Ack
func (p *ProcessList) GetLastLeaderAck(index int) interfaces.IMsg {
	return p.VMs[index].LastLeaderAck
}

// Given a server index, return the last Ack
func (p *ProcessList) SetLastLeaderAck(index int, msg interfaces.IMsg) error {
	// Check the hash of the previous msg before we over write
	p.VMs[index].Undo = p.VMs[index].LastLeaderAck
	p.VMs[index].LastLeaderAck = msg
	return nil
}

func (p *ProcessList) UndoLeaderAck(index int) {
	p.VMs[index].Height--
	p.VMs[index].LastLeaderAck = p.VMs[index].Undo
}

func (p ProcessList) HasMessage() bool {

	for i := 0; i < len(p.FedServers); i++ {
		if len(p.VMs[i].List) > 0 {
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
	p.VMs[i].SigComplete = value
}

// Test if a process list for a server is EOM complete.  Return true if all messages
// have been recieved, and we just need the signaure.  If we need EOM messages, or we
// have all EOM messages and we have the Signature, then we return false.
func (p *ProcessList) EomComplete() bool {
	if p == nil {
		return true
	}

	for i := 0; i < len(p.FedServers); i++ {
		c := p.VMs[i]
		if c.MinuteComplete != 10 {
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
		c := p.VMs[i]
		if !c.SigComplete {
			return false
		}
	}
	return true
}

func (p *ProcessList) FinishedEOM() bool {
	if p == nil || !p.HasMessage() { // Empty or nul, return true.
		return true
	}
	n := len(p.State.GetFedServers(p.DBHeight))
	for i := 0; i < n; i++ {
		c := p.VMs[i]
		if c.Height < c.MinuteHeight {
			return false
		}
	}
	return true
}

func (p *ProcessList) FinishedSIG() bool {
	if p == nil || !p.HasMessage() { // Empty or nul, return true.
		return true
	}
	n := len(p.State.GetFedServers(p.DBHeight))
	for i := 0; i < n; i++ {
		c := p.VMs[i]
		if !c.SigComplete {
			return false
		}
		if c.Height != len(c.List) {
			return false
		}
	}
	return true
}

// Process messages and update our state.
func (p *ProcessList) Process(state *State) (progress bool) {

	if !p.good { // If we don't know this process list is good...
		last := state.DBStates.Last() // Get our last state.
		if last == nil {
			return
		}
		lht := last.DirectoryBlock.GetHeader().GetDBHeight()
		if !last.Saved || lht < p.DBHeight-1 {
			return
		}
		p.good = true
	}
	for i := 0; i < len(p.FedServers); i++ {

		plist := p.VMs[i].List
	thisVM:
		for j := p.VMs[i].Height; j < len(plist); j++ {
			if plist[j] == nil {
				now := time.Now().Unix()
				if p.VMs[i].missingTime == 0 {
					p.VMs[i].missingTime = now
				}
				if now-p.VMs[i].missingTime > 0 {
					missingMsgRequest := messages.NewMissingMsg(state, p.DBHeight, uint32(j))
					if missingMsgRequest != nil {
						state.NetworkOutMsgQueue() <- missingMsgRequest
					}
					p.VMs[i].missingTime = now
				}
				return
			}

			oldAck, ok := p.OldAcks[plist[j].GetHash().Fixed()]
			if !ok {
				// the message from OldAcks is not actually of type Ack
				plist[j] = nil
				return
			}
			thisAck, ok := oldAck.(*messages.Ack)
			if !ok {
				// corresponding acknowledgement not found
				if state.GetOut() {
					p.State.Println("!!!!!!! Missing acknowledgement in process list for", j)
				}
				plist[j] = nil
				return
			}

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
				fmt.Printf("DISCREPANCY: %d %x pl ht: %d \nDetected on: %s\n",
					i,
					p.FedServers[i].GetChainID().Bytes()[:3],
					j,
					state.GetFactomNodeName())
				fmt.Printf("LAST MESS: %x ::: LAST SERIAL: %x\n", last.MessageHash.Bytes()[:3], last.SerialHash.Bytes()[:3])
				fmt.Printf("THIS MESS: %x ::: THIS SERIAL: %x\n", thisAck.MessageHash.Bytes()[:3], thisAck.SerialHash.Bytes()[:3])
				fmt.Printf("EXPECT:    %x \n", expectedSerialHash.Bytes()[:3])
				fmt.Printf("The message that didn't work: %s\n\n", plist[j].String())
				// the SerialHash of this acknowledgment is incorrect
				// according to this node's processList
				plist[j] = nil
				return
			}
			p.SetLastAck(i, thisAck)

			if plist[j].Process(p.DBHeight, state) { // Try and Process this entry
				p.VMs[i].Height = j + 1 // Don't process it again if the process worked.
				progress = true
			} else {
				break thisVM // Don't process further in this list, go to the next.
			}
		}
	}
	return
}

func (p *ProcessList) AddToProcessList(ack *messages.Ack, m interfaces.IMsg) {

	m.SetLeaderChainID(ack.GetLeaderChainID())

	if len(p.VMs[ack.VMIndex].List) > int(ack.Height) && p.VMs[ack.VMIndex].List[ack.Height] != nil {
		fmt.Println(p.String())
		fmt.Println(p.PrintMap())
		panic(fmt.Sprintf("\t%12s %s\n\t%12s %s\n\t %12s %s",
			"OverWriting:",
			p.VMs[ack.VMIndex].List[ack.Height].String(),
			"With:",
			m.String(),
			"Detected on:",
			p.State.GetFactomNodeName(),
		))
	}

	for len(p.VMs[ack.VMIndex].List) <= int(ack.Height) {
		p.VMs[ack.VMIndex].List = append(p.VMs[ack.VMIndex].List, nil)
	}
	p.VMs[ack.VMIndex].LastAck = ack

	p.VMs[ack.VMIndex].List[ack.Height] = m
}

func (p *ProcessList) String() string {
	var buf primitives.Buffer
	if p == nil {
		buf.WriteString("-- <nil>\n")
	} else {
		buf.WriteString(fmt.Sprintf("%s #VMs %d\n", p.State.GetFactomNodeName(), len(p.FedServers)))

		for i := 0; i < len(p.FedServers); i++ {
			server := p.VMs[i]
			eom := fmt.Sprintf("Minute Complete %d Height %d ", server.MinuteComplete, server.Height)
			if p.FinishedEOM() {
				eom = eom + "Finished EOM "
			}
			sig := ""
			if server.SigComplete {
				sig = "Sig Complete "
			}
			if p.FinishedSIG() {
				eom = eom + "Finished SIG "
			}

			buf.WriteString(fmt.Sprintf("  VM %d Fed %d %s %s\n", i, p.ServerMap[server.LeaderMinute][i], eom, sig))
			for j, msg := range server.List {

				if j < server.Height {
					buf.WriteString("  P")
				} else {
					buf.WriteString("   ")
				}

				if msg != nil {
					buf.WriteString("   " + msg.String() + "\n")
				} else {
					buf.WriteString("   <nil>\n")
				}
			}
		}
		buf.WriteString("\n   Federated VMs:\n")
		for _, fed := range p.FedServers {
			buf.WriteString(fmt.Sprintf("    %x\n", fed.GetChainID().Bytes()[:3]))
		}
		buf.WriteString("\n   Audit VMs:\n")
		for _, aud := range p.AuditServers {
			buf.WriteString(fmt.Sprintf("    %x\n", aud.GetChainID().Bytes()[:3]))
		}
	}
	return buf.String()
}

/************************************************
 * Support
 ************************************************/

func NewProcessList(state interfaces.IState, previous *ProcessList, dbheight uint32) *ProcessList {
	// We default to the number of Servers previous.   That's because we always
	// allocate the FUTURE directoryblock, not the current or previous...

	pl := new(ProcessList)

	pl.State = state

	// Make a copy of the previous FedServers
	pl.FedServers = make([]interfaces.IFctServer, 0)
	pl.AuditServers = make([]interfaces.IFctServer, 0)
	if previous != nil {
		pl.FedServers = append(pl.FedServers, previous.FedServers...)
		pl.AuditServers = append(pl.AuditServers, previous.AuditServers...)
	} else {
		pl.AddFedServer(primitives.Sha([]byte("FNode0"))) // Our default for now fed server
	}

	pl.VMs = make([]*VM, 32)
	for i := 0; i < 32; i++ {
		pl.VMs[i] = new(VM)
		pl.VMs[i].List = make([]interfaces.IMsg, 0)

	}

	pl.DBHeight = dbheight

	pl.MakeMap()

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
