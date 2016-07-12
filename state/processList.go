package state

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/database/databaseOverlay"
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

	State     *State
	VMs       []*VM       // Process list for each server (up to 32)
	ServerMap [10][64]int // Map of FedServers to all Servers for each minute

	diffSigTally int /*     Tally of how many VMs have provided different
		                    Directory Block Signatures than what we have
	                        (discard DBlock if > 1/2 have sig differences) */
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
	FaultCnt     map[[32]byte]int        // Count of faults against the Federated Servers

}

type VM struct {
	List           []interfaces.IMsg // Lists of acknowledged messages
	ListAck        []*messages.Ack   // Acknowledgements
	Height         int               // Height of messages that have been processed
	LeaderMinute   int               // Where the leader is in acknowledging messages
	MinuteComplete int               // Highest minute complete recorded (0-9) by the follower
	Synced         bool              // Is this VM synced yet?
	missingTime    int64             // How long we have been waiting for a missing message
	missingEOM     int64             // Ask for EOM
	heartBeat      int64             // Just ping ever so often if we have heard nothing.
}

func (p *ProcessList) Complete() bool {
	for i := 0; i < len(p.FedServers); i++ {
		vm := p.VMs[i]
		if vm.LeaderMinute < 10 {
			return false
		}
		if vm.Height < len(vm.List) {
			return false
		}
	}
	return true
}

// Returns the Virtual Server index for this hash for the given minute
func (p *ProcessList) VMIndexFor(hash []byte) int {

	if p.State.OneLeader {
		return 0
	}

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

func (p *ProcessList) GetVirtualServers(minute int, identityChainID interfaces.IHash) (found bool, index int) {
	found, fedIndex := p.GetFedServerIndexHash(identityChainID)
	if !found {
		return false, -1
	}

	for i := 0; i < len(p.FedServers); i++ {
		fedix := p.ServerMap[minute][i]
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
	prt := fmt.Sprintf("===PrintMapStart=== %d\n", p.DBHeight)
	prt = prt + fmt.Sprintf("dddd %s minute map:  s.LeaderVMIndex %d pl.dbht %d  s.dbht %d s.EOM %v\ndddd     ",
		p.State.FactomNodeName, p.State.LeaderVMIndex, p.DBHeight, p.State.LLeaderHeight, p.State.EOM)
	for i := 0; i < n; i++ {
		prt = fmt.Sprintf("%s%3d", prt, i)
	}
	prt = prt + "\ndddd "
	for i := 0; i < 10; i++ {
		prt = fmt.Sprintf("%s%3d  ", prt, i)
		for j := 0; j < len(p.FedServers); j++ {
			prt = fmt.Sprintf("%s%2d ", prt, p.ServerMap[i][j])
		}
		prt = prt + "\ndddd "
	}
	prt = prt + fmt.Sprintf("\n===PrintMapEnd=== %d\n", p.DBHeight)
	return prt
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
func (p *ProcessList) GetAck(vmIndex int) *messages.Ack {
	return p.GetAckAt(vmIndex, p.VMs[vmIndex].Height)
}

// Given a server index, return the last Ack
func (p *ProcessList) GetAckAt(vmIndex int, height int) *messages.Ack {
	vm := p.VMs[vmIndex]
	if height < 0 || height >= len(vm.ListAck) {
		return nil
	}
	return vm.ListAck[height]
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

func (p *ProcessList) GetLeaderTimestamp() interfaces.Timestamp {
	for _, msg := range p.VMs[0].List {
		if msg.Type() == constants.DIRECTORY_BLOCK_SIGNATURE_MSG {
			return msg.GetTimestamp()
		}
	}
	return new(primitives.Timestamp)
}

func (p *ProcessList) ResetDiffSigTally() {
	p.diffSigTally = 0
}

func (p *ProcessList) IncrementDiffSigTally() {
	p.diffSigTally++
}

func (p *ProcessList) CheckDiffSigTally() {
	// If the majority of VMs' signatures do not match our
	// saved block, we discard that block from our database.
	if p.diffSigTally > 0 && p.diffSigTally > (len(p.FedServers)/2) {
		p.State.DB.Delete([]byte(databaseOverlay.DIRECTORYBLOCK), p.State.ProcessLists.Lists[0].DirectoryBlock.GetKeyMR().Bytes())
	}
}

func ask(p *ProcessList, vmIndex int, waitSeconds int64, vm *VM, thetime int64, height int) int64 {
	now := time.Now().Unix()

	if thetime == 0 {
		thetime = now
	}
	if now-thetime >= waitSeconds {
		missingMsgRequest := messages.NewMissingMsg(p.State, vmIndex, p.DBHeight, uint32(height))
		if missingMsgRequest != nil {
			p.State.NetworkOutMsgQueue() <- missingMsgRequest
		}
		thetime = now
	}
	if p.State.Leader && now-thetime >= waitSeconds+2 {
		id := p.FedServers[p.ServerMap[0][vmIndex]].GetChainID()
		sf := messages.NewServerFault(p.State.GetTimestamp(), id, vmIndex, p.DBHeight, uint32(height))
		if sf != nil {
			p.State.NetworkOutMsgQueue() <- sf
		}
	}

	return thetime
}

// Process messages and update our state.
func (p *ProcessList) Process(state *State) (progress bool) {

	for i := 0; i < len(p.FedServers); i++ {
		vm := p.VMs[i]

		if vm.Height == len(vm.List) && p.State.Syncing && !vm.Synced {
			vm.missingTime = ask(p, i, 1, vm, vm.missingTime, vm.Height)
		}

		vm.heartBeat = ask(p, i, 10, vm, vm.heartBeat, len(vm.List))

	VMListLoop:
		for j := vm.Height; j < len(vm.List); j++ {
			if vm.List[j] == nil {
				vm.missingTime = ask(p, i, 1, vm, vm.missingTime, j)
				break VMListLoop
			}

			thisAck := vm.ListAck[j]

			var expectedSerialHash interfaces.IHash
			var err error

			if vm.Height == 0 {
				expectedSerialHash = thisAck.SerialHash
			} else {
				last := vm.ListAck[vm.Height-1]
				expectedSerialHash, err = primitives.CreateHash(last.MessageHash, thisAck.MessageHash)
				if err != nil {
					vm.List[j] = nil
					vm.ListAck[j] = nil
					// Ask for the correct ack if this one is no good.
					vm.missingTime = ask(p, i, 1, vm, vm.missingTime, j)
					break VMListLoop
				}

				// compare the SerialHash of this acknowledgement with the
				// expected serialHash (generated above)
				if !expectedSerialHash.IsSameAs(thisAck.SerialHash) {
					fmt.Printf("dddd %20s %10s --- %10s %10x %10s %10x \n", "Conflict", p.State.FactomNodeName, "expected", expectedSerialHash.Bytes()[:3], "This", thisAck.Bytes()[:3])
					fmt.Printf("dddd Error detected on %s\nSerial Hash failure: Fed Server %d  Leader ID %x List Ht: %d \nDetected on: %s\n",
						state.GetFactomNodeName(),
						i,
						p.FedServers[i].GetChainID().Bytes()[:3],
						j,
						vm.List[j].String())
					fmt.Printf("dddd Last Ack: %6x  Last Serial: %6x\n", last.GetHash().Bytes()[:3], last.SerialHash.Bytes()[:3])
					fmt.Printf("dddd his Ack: %6x  This Serial: %6x\n", thisAck.GetHash().Bytes()[:3], thisAck.SerialHash.Bytes()[:3])
					fmt.Printf("dddd Expected: %6x\n", expectedSerialHash.Bytes()[:3])
					fmt.Printf("dddd The message that didn't work: %s\n\n", vm.List[j].String())
					// the SerialHash of this acknowledgment is incorrect
					// according to this node's processList
					vm.List[j] = nil
					vm.missingTime = ask(p, i, 1, vm, vm.missingTime, j)
					break VMListLoop
				}
			}

			if vm.List[j].Process(p.DBHeight, state) { // Try and Process this entry
				vm.heartBeat = 0
				vm.Height = j + 1 // Don't process it again if the process worked.
				progress = true
			} else {
				break VMListLoop // Don't process further in this list, go to the next.
			}
		}
	}
	return
}

func (p *ProcessList) AddToProcessList(ack *messages.Ack, m interfaces.IMsg) {

	toss := func(hint string) {
		delete(p.State.Holding, ack.GetHash().Fixed())
		delete(p.State.Acks, ack.GetHash().Fixed())
	}

	if p == nil {
		return
	}

	now := p.State.GetTimestamp()

	_, isNew2 := p.State.Replay.Valid(constants.INTERNAL_REPLAY, m.GetMsgHash().Fixed(), m.GetTimestamp(), now)
	if !isNew2 {
		toss("seen before, or too old")
		return
	}

	vm := p.VMs[ack.VMIndex]

	if ack.DBHeight != p.DBHeight {
		panic(fmt.Sprintf("Ack is wrong height.  Expected: %d Ack: %s", p.DBHeight, ack.String()))
		return
	}

	if len(vm.List) > int(ack.Height) && vm.List[ack.Height] != nil {

		if vm.List[ack.Height].GetMsgHash().IsSameAs(m.GetMsgHash()) {
			fmt.Printf("dddd %-30s %10s %s\n", "xxxxxxxxx PL Duplicate   ", p.State.GetFactomNodeName(), m.String())
			fmt.Printf("dddd %-30s %10s %s\n", "xxxxxxxxx PL Duplicate   ", p.State.GetFactomNodeName(), ack.String())
			fmt.Printf("dddd %-30s %10s %s\n", "xxxxxxxxx PL Duplicate vm", p.State.GetFactomNodeName(), vm.List[ack.Height].String())
			fmt.Printf("dddd %-30s %10s %s\n", "xxxxxxxxx PL Duplicate vm", p.State.GetFactomNodeName(), vm.ListAck[ack.Height].String())
			toss("2")
			return
		}

		fmt.Printf("dddd\t%12s %s %s\n", "OverWriting:", vm.List[ack.Height].String(), "with")
		fmt.Printf("dddd\t%12s %s\n", "with:", m.String())
		fmt.Printf("dddd\t%12s %s\n", "Detected on:", p.State.GetFactomNodeName())
		fmt.Printf("dddd\t%12s %s\n", "old ack", vm.ListAck[ack.Height].String())
		fmt.Printf("dddd\t%12s %s\n", "new ack", ack.String())
		fmt.Printf("dddd\t%12s %s\n", "VM Index", ack.VMIndex)
		toss("3")
		return
	}

	// From this point on, we consider the transaction recorded.  If we detect it has already been
	// recorded, then we still treat it as if we recorded it.

	vm.heartBeat = 0 // We have heard from this VM

	// We have already tested and found m to be a new message.  We now record its hashes so later, we
	// can detect that it has been recorded.  We don't care about the results of IsTSValid_ at this point.
	p.State.Replay.IsTSValid_(constants.INTERNAL_REPLAY, m.GetMsgHash().Fixed(), m.GetTimestamp(), now)

	delete(p.State.Acks, ack.GetHash().Fixed())
	delete(p.State.Holding, m.GetMsgHash().Fixed())

	// Both the ack and the message hash to the same GetHash()
	m.SetLocal(false)
	ack.SetLocal(false)
	ack.SetPeer2Peer(false)
	m.SetPeer2Peer(false)

	p.State.NetworkOutMsgQueue() <- ack
	p.State.NetworkOutMsgQueue() <- m

	for len(vm.List) <= int(ack.Height) {
		vm.List = append(vm.List, nil)
		vm.ListAck = append(vm.ListAck, nil)
	}

	p.VMs[ack.VMIndex].List[ack.Height] = m
	p.VMs[ack.VMIndex].ListAck[ack.Height] = ack
	p.OldMsgs[m.GetHash().Fixed()] = m
	p.OldAcks[m.GetMsgHash().Fixed()] = ack

}

func (p *ProcessList) String() string {
	var buf primitives.Buffer
	if p == nil {
		buf.WriteString("-- <nil>\n")
	} else {
		buf.WriteString(fmt.Sprintf("===ProcessListStart=== %s %d\n", p.State.GetFactomNodeName(), p.DBHeight))
		buf.WriteString(fmt.Sprintf("%s #VMs %d Complete %v\n", p.State.GetFactomNodeName(), len(p.FedServers), p.Complete()))

		for i := 0; i < len(p.FedServers); i++ {
			vm := p.VMs[i]
			buf.WriteString(fmt.Sprintf("  VM %d  vMin %d vHeight %v len(List)%d Syncing %v Synced %v EOMProcessed %d DBSigProcessed %d\n",
				i, vm.LeaderMinute, vm.Height, len(vm.List), p.State.Syncing, vm.Synced, p.State.EOMProcessed, p.State.DBSigProcessed))
			for j, msg := range vm.List {
				buf.WriteString(fmt.Sprintf("   %3d", j))
				if j < vm.Height {
					buf.WriteString(" P")
				} else {
					buf.WriteString("  ")
				}

				if msg != nil {
					leader := fmt.Sprintf("[%x] ", vm.ListAck[j].LeaderChainID.Bytes()[:3])
					buf.WriteString("   " + leader + msg.String() + "\n")
				} else {
					buf.WriteString("   <nil>\n")
				}
			}
		}
		buf.WriteString(fmt.Sprintf("===FederatedServersStart=== %d\n", len(p.FedServers)))
		for _, fed := range p.FedServers {
			buf.WriteString(fmt.Sprintf("    %x\n", fed.GetChainID().Bytes()[:3]))
		}
		buf.WriteString(fmt.Sprintf("===FederatedServersEnd=== %d\n", len(p.FedServers)))
		buf.WriteString(fmt.Sprintf("===AuditServersStart=== %d\n", len(p.AuditServers)))
		for _, aud := range p.AuditServers {
			buf.WriteString(fmt.Sprintf("    %x\n", aud.GetChainID().Bytes()[:3]))
		}
		buf.WriteString(fmt.Sprintf("===AuditServersEnd=== %d\n", len(p.AuditServers)))
	}
	buf.WriteString(fmt.Sprintf("===ProcessListEnd=== %s %d\n", p.State.GetFactomNodeName(), p.DBHeight))
	return buf.String()
}

/************************************************
 * Support
 ************************************************/

func NewProcessList(state interfaces.IState, previous *ProcessList, dbheight uint32) *ProcessList {
	// We default to the number of Servers previous.   That's because we always
	// allocate the FUTURE directoryblock, not the current or previous...

	pl := new(ProcessList)

	pl.State = state.(*State)

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
		pl.VMs[i].Synced = true
	}

	pl.DBHeight = dbheight

	pl.MakeMap()

	pl.OldMsgs = make(map[[32]byte]interfaces.IMsg)
	pl.OldAcks = make(map[[32]byte]interfaces.IMsg)

	pl.NewEBlocks = make(map[[32]byte]interfaces.IEntryBlock)
	pl.NewEntries = make(map[[32]byte]interfaces.IEntry)
	pl.Commits = make(map[[32]byte]interfaces.IMsg)

	// If a federated server, this is the server index, which is our index in the FedServers list

	var err error

	if previous != nil {
		pl.DirectoryBlock = directoryBlock.NewDirectoryBlock(previous.DirectoryBlock)
		pl.AdminBlock = adminBlock.NewAdminBlock(previous.AdminBlock)
		pl.EntryCreditBlock, err = entryCreditBlock.NextECBlock(previous.EntryCreditBlock)
	} else {
		pl.DirectoryBlock = directoryBlock.NewDirectoryBlock(nil)
		pl.AdminBlock = adminBlock.NewAdminBlock(nil)
		pl.EntryCreditBlock, err = entryCreditBlock.NextECBlock(nil)
	}

	pl.ResetDiffSigTally()

	if err != nil {
		panic(err.Error())
	}

	return pl
}
