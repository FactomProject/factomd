package state

import (
	"fmt"

	"bytes"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"log"

	"time"

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
	Sealing      bool                    // We are in the process of sealing this process list
}

type VM struct {
	List           []interfaces.IMsg // Lists of acknowledged messages
	ListAck        []*messages.Ack   // Acknowledgements
	Height         int               // Height of messages that have been processed
	LeaderMinute   int               // Where the leader is in acknowledging messages
	Seal           int               // Sealed with an EOM minute, and released (0) when all EOM are found.
	SealTime       int64             // The time since we started waiting
	SealHeight     uint32            // Entries belowe the seal can still be recorded.
	MinuteComplete int               // Highest minute complete recorded (0-9) by the follower
	MinuteFinished int               // Highest minute processed (0-9) by the follower
	MinuteHeight   int               // Height of the last minute complete
	missingTime    int64             // How long we have been waiting for a missing message
}

// Attempts to unseal. Takes a minute (1-10) Returns false if it cannot.
// Returns false if no seal is found.
func (p *ProcessList) Unsealable(minute int) bool {
searchVMs:
	for i := 0; i < len(p.FedServers); i++ {
		vm := p.VMs[i]
		if len(vm.List) != vm.Height {
			return false
		}
		for _, v := range vm.List {
			if v == nil {
				return false
			}
			if eom, ok := v.(*messages.EOM); ok {
				if int(eom.Minute+1) == minute {
					continue searchVMs
				}
			}
		}
		return false
	}
	return true
}

func (p *ProcessList) Unseal(minute int) bool {
	if !p.Unsealable(minute) {
		return false
	}
	for i := 0; i < len(p.FedServers); i++ {
		p.VMs[i].Seal = 0
		p.VMs[i].SealHeight = 0
		p.Sealing = false
	}
	return true
}

// Returns the Virtual Server index for this hash for the given minute
func (p *ProcessList) VMIndexFor(hash []byte) int {
	return 0
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
	if !p.State.Green() && p.State.GetIdentityChainID().IsSameAs(identityChainID) {
		return false, -1
	}

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
	prt := fmt.Sprintf("===PrintMapStart=== %d\n", p.DBHeight)
	prt = prt + " min"
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
	prt = prt + fmt.Sprintf("===PrintMapEnd=== %d\n", p.DBHeight)
	return prt
}

// Return the lowest minute number in our lists.  Note that Minute Markers END
// a minute, so After MinuteComplete=0
func (p *ProcessList) MinuteComplete() int {
	m := 10
	for i := 0; i < len(p.FedServers); i++ {
		mm := 0
		for _, msg := range p.VMs[i].List {
			if eom, ok := msg.(*messages.EOM); ok {
				mm = int(eom.Minute + 1)
				p.VMs[i].MinuteComplete = mm
			}
		}
		if m > mm {
			m = mm
		}
	}
	return m
}

// Return the lowest minute number in our lists.  Note that Minute Markers END
// a minute, so After MinuteComplete=0
func (p *ProcessList) MinuteFinished() int {
	m := 10
	for i := 0; i < len(p.FedServers); i++ {
		mm := 0
		for j, msg := range p.VMs[i].List {
			if j == p.VMs[i].Height {
				break
			}
			if eom, ok := msg.(*messages.EOM); ok {
				mm = int(eom.Minute + 1)
				p.VMs[i].MinuteFinished = mm
			}
		}
		if m > mm {
			m = mm
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

func (p *ProcessList) FinishedEOM() bool {
	if p == nil || !p.HasMessage() { // Empty or nul, return true.
		return true
	}
	n := len(p.FedServers)
	for i := 0; i < n; i++ {
		c := p.VMs[i]
		if c.Height <= c.MinuteHeight {
			return false
		}
	}
	return true
}

// Process messages and update our state.
func (p *ProcessList) Process(state *State) (progress bool) {

	now := time.Now().Unix()
	ask := func(vmIndex int, vm *VM, thetime int64, j int) int64 {
		if thetime == 0 {
			thetime = now
		}
		if now-thetime > 1 {
			missingMsgRequest := messages.NewMissingMsg(state, p.DBHeight, uint32(j))
			if missingMsgRequest != nil {
				state.NetworkOutMsgQueue() <- missingMsgRequest
			}
			thetime = now
		}
		if p.State.Leader && now-thetime > 2 {
			id := p.FedServers[p.ServerMap[vm.MinuteComplete][vmIndex]].GetChainID()
			sf := messages.NewServerFault(state.GetTimestamp(), id, vmIndex, p.DBHeight, uint32(j))
			if sf != nil {
				state.NetworkOutMsgQueue() <- sf
			}
		}

		return thetime
	}

	if !p.good { // If we don't know this process list is good...
		if p.DBHeight == 0 {
			p.good = true
		}else {
			prev := state.DBStates.Get(p.DBHeight - 1)

			if prev == nil {
				return
			}
			if !prev.Saved {
				return
			}
			p.good = true
		}
		fmt.Println("dddd xxxx Not Good", p.State.FactomNodeName)
	}
	for i := 0; i < len(p.FedServers); i++ {
		// Just in case, set p.diffSigTally to 0 when initiating pass-through
		if i == 0 {
			p.diffSigTally = 0
		}
		vm := p.VMs[i]

		plist := vm.List
		alist := vm.ListAck

		if vm.Height == len(plist) && p.Sealing && vm.Seal == 0 {
			vm.SealTime = ask(i, vm, vm.SealTime+2, vm.Height)
		}

		for j := vm.Height; j < len(plist); j++ {
			if plist[j] == nil {
				vm.missingTime = ask(i, vm, vm.missingTime, j)
				break
			}

			// When processing DirectoryBlockSignatures, we check to see if the signed block
			// matches our own saved block. If the majority of VMs' signatures do not match
			// our saved block, we discard that block from our database.
			//if plist[j].Type() == constants.DIRECTORY_BLOCK_SIGNATURE_MSG {
			//dbs := plist[j].(*messages.DirectoryBlockSignature)
			//myDBlock := state.GetDirectoryBlockByHeight(dbs.DBHeight - 1)
			//myDBlock.GetHeader().SetTimestamp(p.GetLeaderTimestamp())
			//if !dbs.DirectoryBlockKeyMR.IsSameAs(state.ProcessLists.Lists[0].DirectoryBlock.GetKeyMR()) {
			//	p.diffSigTally++
			//	if p.diffSigTally > 0 && p.diffSigTally > (len(p.FedServers)/2) {
			//		state.DB.Delete([]byte{byte(databaseOverlay.DIRECTORYBLOCK)}, state.ProcessLists.Lists[0].DirectoryBlock.GetKeyMR().Bytes())
			//	}
			//}
			//if i >= len(p.FedServers) {
			//	p.diffSigTally = 0
			//}
			//}

			thisAck := alist[j]
			if thisAck == nil { // IF I don't have an Ack to match this entry
				plist[j] = nil // throw the entry away, and continue to the
				break          // next list.  SHOULD NEVER HAPPEN.
			}

			var expectedSerialHash interfaces.IHash
			var err error
			if vm.Height == 0 {
				expectedSerialHash = thisAck.SerialHash
			} else {
				last := p.GetAckAt(i, vm.Height-1)
				expectedSerialHash, err = primitives.CreateHash(last.MessageHash, thisAck.MessageHash)
				if err != nil {
					// cannot create a expectedSerialHash to compare to
					plist[j] = nil
					break
				}

				// compare the SerialHash of this acknowledgement with the
				// expected serialHash (generated above)
				if !expectedSerialHash.IsSameAs(thisAck.SerialHash) {
					p.State.DebugPrt("Process List")
					fmt.Printf("Error detected on %s\nSerial Hash failure: Fed Server %d  Leader ID %x List Ht: %d \nDetected on: %s\n",
						state.GetFactomNodeName(),
						i,
						p.FedServers[i].GetChainID().Bytes()[:3],
						j,
						plist[j].String())
					fmt.Printf("Last Ack: %6x  Last Serial: %6x\n", last.GetHash().Bytes()[:3], last.SerialHash.Bytes()[:3])
					fmt.Printf("This Ack: %6x  This Serial: %6x\n", thisAck.GetHash().Bytes()[:3], thisAck.SerialHash.Bytes()[:3])
					fmt.Printf("Expected: %6x\n", expectedSerialHash.Bytes()[:3])
					fmt.Printf("The message that didn't work: %s\n\n", plist[j].String())
					fmt.Println(p.PrintMap())
					// the SerialHash of this acknowledgment is incorrect
					// according to this node's processList
					plist[j] = nil
					break
				}
			}

			if plist[j].Process(p.DBHeight, state) { // Try and Process this entry
				vm.Height = j + 1 // Don't process it again if the process worked.
				progress = true
			} else {
				break // Don't process further in this list, go to the next.
			}
		}
	}
	return
}

func (p *ProcessList) GetLeaderTimestamp() uint32 {
	for _, msg := range p.VMs[0].List {
		if msg.Type() == constants.DIRECTORY_BLOCK_SIGNATURE_MSG {
			ts := msg.GetTimestamp()
			return uint32(ts.GetTime().Unix())
		}
	}
	return 0
}

// Check to assure that the given VM has been completely processed
func (p *ProcessList) GoodTo(vmIndex int) bool {
	if vmIndex < 0 {
		vmIndex = p.State.LeaderVMIndex
	}
	vm := p.VMs[vmIndex]
	if len(vm.List) > vm.Height {
		return false
	}
	for _, v := range vm.List {
		if v == nil {
			return false
		}
	}
	return true
}

func (p *ProcessList) AddToProcessList(ack *messages.Ack, m interfaces.IMsg) {

	stall := func(hint string) {
		p.State.StallAck(ack)
		p.State.Holding[m.GetHash().Fixed()] = m
		delete(p.State.Acks, ack.GetHash().Fixed())
		if p.State.DebugConsensus {
			fmt.Println("dddd", hint, p.State.FactomNodeName, "Stall", m.String())
			fmt.Println("dddd", hint, p.State.FactomNodeName, "Stall", ack.String())
		}
	}

	outOfOrder := func(hint string) {
		p.State.OutOfOrderAck(ack)
		p.State.Holding[m.GetHash().Fixed()] = m
		delete(p.State.Acks, ack.GetHash().Fixed())
		if p.State.DebugConsensus {
			fmt.Println("dddd", hint, p.State.FactomNodeName, "OutOfOrder", m.String())
			fmt.Println("dddd", hint, p.State.FactomNodeName, "OutOfOrder", ack.String())
		}
	}

	toss := func(hint string) {
		delete(p.State.Holding, ack.GetHash().Fixed())
		delete(p.State.Acks, ack.GetHash().Fixed())
		if p.State.DebugConsensus {
			fmt.Println("dddd", hint, p.State.FactomNodeName, "Toss", m.String())
			fmt.Println("dddd", hint, p.State.FactomNodeName, "Toss", ack.String())
		}
	}

	now := int64(p.State.GetTimestamp() / 1000)

	_, isnew := p.State.InternalReplay.Valid(m.GetHash().Fixed(), int64(m.GetTimestamp()/1000), now)
	if !isnew {
		toss("seen before")
		return
	}

	vm := p.VMs[ack.VMIndex]

	if ack.DBHeight != p.DBHeight {
		panic(fmt.Sprintf("Ack is wrong height.  Expected: %d Ack: %s", p.DBHeight, ack.String()))
		return
	}

	// If this vm is sealed, then we can't add more messages.
	if vm.Seal > 0 && ack.Height >= vm.SealHeight {
		stall("b")
		return
	}

	if len(vm.List) > vm.Height {
		outOfOrder("c")
		return
	}

	if int(ack.Height) > vm.Height {
		outOfOrder("d")
		return
	}

	if len(vm.List) > int(ack.Height) && vm.List[ack.Height] != nil {

		if vm.List[ack.Height].GetMsgHash().IsSameAs(m.GetMsgHash()) {
			fmt.Printf("dddd %-30s %10s %s\n", "xxxxxxxxx PL Duplicate", p.State.GetFactomNodeName(), m.String())
			toss("2")
			return
		}

		if vm.List[ack.Height] != nil {
			fmt.Println(p.String())
			fmt.Println(p.PrintMap())
			fmt.Printf("dddd\t%12s %s %s\n", "OverWriting:", vm.List[ack.Height].String(), "with")
			fmt.Printf("dddd\t%12s %s\n", "with:", m.String())
			fmt.Printf("dddd\t%12s %s\n", "Detected on:", p.State.GetFactomNodeName())
			fmt.Printf("dddd\t%12s %s\n", "old ack", vm.ListAck[ack.Height].String())
			fmt.Printf("dddd\t%12s %s\n", "new ack", ack.String())
			fmt.Printf("dddd\t%12s %s\n", "VM Index", ack.VMIndex)
			toss("3")
			return
		}
	}

	// From this point on, we consider the transaction recorded.  If we detect it has already been
	// recorded, then we still treat it as if we recorded it.

	eom, ok := m.(*messages.EOM)
	if ok {
		if vm.Seal > 0 {
			if p.State.DebugConsensus {
				fmt.Println("dddd EOM after Seal", p.State.FactomNodeName, m.String())
				fmt.Println("dddd EOM after Seal", p.State.FactomNodeName, ack.String())
			}
			outOfOrder("eom")
		}
		if p.State.Leader && eom.IsLocal() {
			p.State.EOM = int(eom.Minute + 1)
		}
		p.Sealing = true
		vm.Seal = int(eom.Minute + 1)
		vm.SealHeight = ack.Height
		vm.MinuteComplete = int(eom.Minute + 1)
		vm.MinuteHeight = vm.Height
	}

	msgOk := p.State.InternalReplay.IsTSValid_(m.GetHash().Fixed(), int64(m.GetTimestamp()/1000), now)

	if !msgOk { // If we already have this message or acknowledgement recorded,
		if p.State.DebugConsensus {
			fmt.Println("dddd Msg Repeat", p.State.FactomNodeName, m.String())
			fmt.Println("dddd Msg Repeat", p.State.FactomNodeName, ack.String())
		}
		toss("4")
		return // we don't have to do anything.  Just say we got it handled.
	}

	p.State.NetworkOutMsgQueue() <- ack
	p.State.NetworkOutMsgQueue() <- m
	delete(p.State.Acks, ack.GetHash().Fixed())
	delete(p.State.Holding, m.GetHash().Fixed())

	m.SetLeaderChainID(ack.GetLeaderChainID())
	m.SetMinute(ack.Minute)

	// Both the ack and the message hash to the same GetHash()
	m.SetLocal(false)
	ack.SetLocal(false)
	ack.SetPeer2Peer(false)
	m.SetPeer2Peer(false)

	length := len(p.VMs[ack.VMIndex].List)
	for length <= int(ack.Height) {
		p.VMs[ack.VMIndex].List = append(p.VMs[ack.VMIndex].List, nil)
		p.VMs[ack.VMIndex].ListAck = append(p.VMs[ack.VMIndex].ListAck, nil)
		length = len(p.VMs[ack.VMIndex].List)
	}


	p.VMs[ack.VMIndex].List[ack.Height] = m
	p.VMs[ack.VMIndex].ListAck[ack.Height] = ack
	p.OldMsgs[m.GetHash().Fixed()] = m
	p.OldAcks[m.GetHash().Fixed()] = ack

}

func (p *ProcessList) String() string {
	var buf primitives.Buffer
	if p == nil {
		buf.WriteString("-- <nil>\n")
	} else {
		buf.WriteString(fmt.Sprintf("===ProcessListStart=== %s %d\n", p.State.GetFactomNodeName(), p.DBHeight))
		buf.WriteString(fmt.Sprintf("%s #VMs %d\n", p.State.GetFactomNodeName(), len(p.FedServers)))

		for i := 0; i < len(p.FedServers); i++ {
			vm := p.VMs[i]
			eom := fmt.Sprintf("MinuteComplete() %2d MinuteFinished() %2d vm.Height %3d Len %3d ",
				p.MinuteComplete(),
				p.MinuteFinished(),
				vm.Height,
				len(vm.List))
			min := vm.LeaderMinute
			if min > 9 {
				min = 9
			}
			buf.WriteString(fmt.Sprintf("  VM %d VM Minute %d Fed Server: %d %s\n", i, vm.LeaderMinute, p.ServerMap[min][i], eom))
			for j, msg := range vm.List {
				buf.WriteString(fmt.Sprintf("   %3d", j))
				if j < vm.Height {
					buf.WriteString(" P")
				} else {
					buf.WriteString("  ")
				}

				if msg != nil {
					buf.WriteString("   " + msg.String() + "\n")
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
