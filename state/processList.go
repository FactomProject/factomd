// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"bytes"
	"fmt"
	"sync"

	"encoding/binary"

	"os"

	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	//"github.com/FactomProject/factomd/database/databaseOverlay"

	log "github.com/sirupsen/logrus"
)

var _ = fmt.Print
var _ = log.Print

var plLogger = packageLogger.WithFields(log.Fields{"subpack": "process-list"})

type Request struct {
	vmIndex    int    // VM Index
	vmheight   uint32 // Height in the Process List where we are missing a message
	wait       int64  // How long to wait before we actually request
	sent       int64  // Last time sent (zero means none have been sent)
	requestCnt int
}

var _ interfaces.IRequest = (*Request)(nil)

func (r *Request) key() (thekey [32]byte) {
	binary.BigEndian.PutUint32(thekey[0:4], uint32(r.vmIndex))
	binary.BigEndian.PutUint64(thekey[5:13], uint64(r.wait))
	binary.BigEndian.PutUint64(thekey[14:22], uint64(r.vmheight))
	return thekey
}

/*
func (r *Request) key() (thekey [20]byte) {
	binary.BigEndian.PutUint32(thekey[0:4], uint32(r.vmIndex))
	binary.BigEndian.PutUint64(thekey[4:12], uint64(r.wait))
	binary.BigEndian.PutUint64(thekey[12:20], uint64(r.sent))
	return
}
*/

type ProcessList struct {
	DBHeight uint32 // The directory block height for these lists

	// Temporary balances from updating transactions in real time.
	FactoidBalancesT      map[[32]byte]int64
	FactoidBalancesTMutex sync.Mutex
	ECBalancesT           map[[32]byte]int64
	ECBalancesTMutex      sync.Mutex

	State        *State
	VMs          []*VM       // Process list for each server (up to 32)
	ServerMap    [10][64]int // Map of FedServers to all Servers for each minute
	System       VM          // System Faults and other system wide messages
	SysHighest   int
	diffSigTally int /* Tally of how many VMs have provided different
		                    					             Directory Block Signatures than what we have
	                                            (discard DBlock if > 1/2 have sig differences) */

	// messages processed in this list
	OldMsgs     map[[32]byte]interfaces.IMsg
	oldmsgslock *sync.Mutex

	// Chains that are executed, but not processed. There is a small window of a pending chain that the ack
	// will pass and the chainhead will fail. This covers that window. This is only used by WSAPI,
	// do not use it anywhere internally.
	PendingChainHeads *SafeMsgMap

	OldAcks     map[[32]byte]interfaces.IMsg
	oldackslock *sync.Mutex

	// Entry Blocks added within 10 minutes (follower and leader)
	NewEBlocks     map[[32]byte]interfaces.IEntryBlock
	neweblockslock *sync.Mutex

	NewEntriesMutex sync.RWMutex
	NewEntries      map[[32]byte]interfaces.IEntry

	// State information about the directory block while it is under construction.  We may
	// have to start building the next block while still building the previous block.
	AdminBlock       interfaces.IAdminBlock
	EntryCreditBlock interfaces.IEntryCreditBlock
	DirectoryBlock   interfaces.IDirectoryBlock

	// Number of Servers acknowledged by Factom
	Matryoshka   []interfaces.IHash   // Reverse Hash
	AuditServers []interfaces.IServer // List of Audit Servers
	FedServers   []interfaces.IServer // List of Federated Servers

	// AmINegotiator is just used for displaying an "N" next to a node
	// that is the assigned negotiator for a particular processList
	// height
	AmINegotiator bool

	// DB Sigs
	DBSignatures     []DBSig
	DBSigAlreadySent bool

	Requests map[[32]byte]*Request
	//Requests map[[20]byte]*Request
	NextHeightToProcess [64]int
}

var _ interfaces.IProcessList = (*ProcessList)(nil)

func (p *ProcessList) GetAmINegotiator() bool {
	return p.AmINegotiator
}

func (p *ProcessList) SetAmINegotiator(b bool) {
	p.AmINegotiator = b
}

// Data needed to add to admin block
type DBSig struct {
	ChainID   interfaces.IHash
	Signature interfaces.IFullSignature
	VMIndex   int
}

type VM struct {
	List            []interfaces.IMsg // Lists of acknowledged messages
	ListAck         []*messages.Ack   // Acknowledgements
	Height          int               // Height of messages that have been processed
	EomMinuteIssued int               // Last Minute Issued on this VM (from the leader, when we are the leader)
	LeaderMinute    int               // Where the leader is in acknowledging messages
	Synced          bool              // Is this VM synced yet?
	//faultingEOM           int64             // Faulting for EOM because it is too late
	heartBeat   int64 // Just ping ever so often if we have heard nothing.
	Signed      bool  // We have signed the previous block.
	WhenFaulted int64 // WhenFaulted is a timestamp of when this VM was faulted
	// vm.WhenFaulted serves as a bool flag (if > 0, the vm is currently considered faulted)
	FaultFlag int // FaultFlag tracks what the VM was faulted for (0 = EOM missing, 1 = negotiation issue)
}

func (p *ProcessList) Clear() {
	return
	//p.State.AddStatus(fmt.Sprintf("PROCESSLIST.Clear dbht %d", p.DBHeight))
	p.FactoidBalancesTMutex.Lock()
	defer p.FactoidBalancesTMutex.Unlock()
	p.FactoidBalancesT = nil

	p.ECBalancesTMutex.Lock()
	defer p.ECBalancesTMutex.Unlock()
	p.ECBalancesT = nil

	p.oldmsgslock.Lock()
	defer p.oldmsgslock.Unlock()
	p.OldMsgs = nil

	p.oldackslock.Lock()
	defer p.oldackslock.Unlock()
	p.OldAcks = nil

	p.neweblockslock.Lock()
	defer p.neweblockslock.Unlock()
	p.NewEBlocks = nil

	p.NewEntriesMutex.Lock()
	defer p.NewEntriesMutex.Unlock()
	p.NewEntries = nil

	p.AdminBlock = nil
	p.EntryCreditBlock = nil
	p.DirectoryBlock = nil

	p.Matryoshka = nil
	p.AuditServers = nil
	p.FedServers = nil

	p.DBSignatures = nil

	p.Requests = nil
}

func (p *ProcessList) GetKeysNewEntries() (keys [][32]byte) {
	keys = make([][32]byte, p.LenNewEntries())

	if p == nil {
		return
	}
	p.NewEntriesMutex.RLock()
	defer p.NewEntriesMutex.RUnlock()
	i := 0
	for k := range p.NewEntries {
		keys[i] = k
		i++
	}
	return
}

func (p *ProcessList) GetNewEntry(key [32]byte) interfaces.IEntry {
	p.NewEntriesMutex.RLock()
	defer p.NewEntriesMutex.RUnlock()
	return p.NewEntries[key]
}

func (p *ProcessList) LenNewEntries() int {
	if p == nil {
		return 0
	}
	p.NewEntriesMutex.RLock()
	defer p.NewEntriesMutex.RUnlock()
	return len(p.NewEntries)
}

func (p *ProcessList) Complete() bool {
	if p == nil {
		return false
	}
	if p.DBHeight <= p.State.GetHighestSavedBlk() {
		return true
	}
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

func SortServers(servers []interfaces.IServer) []interfaces.IServer {
	for i := 0; i < len(servers)-1; i++ {
		done := true
		for j := 0; j < len(servers)-1-i; j++ {
			fs1 := servers[j].GetChainID().Bytes()
			fs2 := servers[j+1].GetChainID().Bytes()
			if bytes.Compare(fs1, fs2) > 0 {
				tmp := servers[j]
				servers[j] = servers[j+1]
				servers[j+1] = tmp
				done = false
			}
		}
		if done {
			return servers
		}
	}
	return servers
}

func (p *ProcessList) SortFedServers() {
	p.FedServers = SortServers(p.FedServers)
}

func (p *ProcessList) SortAuditServers() {
	p.AuditServers = SortServers(p.AuditServers)
}

func (p *ProcessList) SortDBSigs() {
	// Sort by VMIndex
	for i := 0; i < len(p.DBSignatures)-1; i++ {
		done := true
		for j := 0; j < len(p.DBSignatures)-1-i; j++ {
			if p.DBSignatures[j].VMIndex > p.DBSignatures[j+1].VMIndex {
				tmp := p.DBSignatures[j]
				p.DBSignatures[j] = p.DBSignatures[j+1]
				p.DBSignatures[j+1] = tmp
				done = false
			}
		}
		if done {
			return
		}
	}
	/* Sort by ChainID
	for i := 0; i < len(p.DBSignatures)-1; i++ {
		done := true
		for j := 0; j < len(p.DBSignatures)-1-i; j++ {
			fs1 := p.DBSignatures[j].ChainID.Bytes()
			fs2 := p.DBSignatures[j+1].ChainID.Bytes()
			if bytes.Compare(fs1, fs2) > 0 {
				tmp := p.DBSignatures[j]
				p.DBSignatures[j] = p.DBSignatures[j+1]
				p.DBSignatures[j+1] = tmp
				done = false
			}
		}
		if done {
			return
		}
	}*/
}

// Returns the Federated Server responsible for this hash in this minute
func (p *ProcessList) FedServerFor(minute int, hash []byte) interfaces.IServer {
	vs := p.VMIndexFor(hash)
	if vs < 0 {
		return nil
	}
	fedIndex := p.ServerMap[minute][vs]
	return p.FedServers[fedIndex]
}

func FedServerVM(serverMap [10][64]int, numberOfFedServers int, minute int, fedIndex int) int {
	for i := 0; i < numberOfFedServers; i++ {
		if serverMap[minute][i] == fedIndex {
			return i
		}
	}
	return -1
}

func (p *ProcessList) GetVirtualServers(minute int, identityChainID interfaces.IHash) (found bool, index int) {
	found, fedIndex := p.GetFedServerIndexHash(identityChainID)
	if !found {
		return false, -1
	}

	p.MakeMap()

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
// serverMap[minute][vmIndex] => Index of the Federated Server responsible for that minute
func MakeMap(numberFedServers int, dbheight uint32) (serverMap [10][64]int) {
	if numberFedServers > 0 {
		indx := int(dbheight*131) % numberFedServers
		for i := 0; i < 10; i++ {
			indx = (indx + 1) % numberFedServers
			for j := 0; j < numberFedServers; j++ {
				serverMap[i][j] = indx
				indx = (indx + 1) % numberFedServers
			}
		}
	}
	return
}

func (p *ProcessList) MakeMap() {
	p.ServerMap = MakeMap(len(p.FedServers), p.DBHeight)
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
	p.SortFedServers()
	found, i := p.GetFedServerIndexHash(identityChainID)
	if found {
		//p.State.AddStatus(fmt.Sprintf("ProcessList.AddFedServer Server already there %x at height %d", identityChainID.Bytes()[2:6], p.DBHeight))
		return i
	}
	if i < 0 {
		return i
	}
	// If an audit server, it gets promoted
	auditFound, _ := p.GetAuditServerIndexHash(identityChainID)
	if auditFound {
		//p.State.AddStatus(fmt.Sprintf("ProcessList.AddFedServer Server %x was an audit server at height %d", identityChainID.Bytes()[2:6], p.DBHeight))
		p.RemoveAuditServerHash(identityChainID)
	}

	// Inform Elections of a new leader
	InMsg := p.State.EFactory.NewAddLeaderInternal(
		p.State.FactomNodeName,
		p.DBHeight,
		identityChainID,
	)
	p.State.electionsQueue.Enqueue(InMsg)

	p.FedServers = append(p.FedServers, nil)
	copy(p.FedServers[i+1:], p.FedServers[i:])
	p.FedServers[i] = &Server{ChainID: identityChainID, Online: true}
	//p.State.AddStatus(fmt.Sprintf("ProcessList.AddFedServer Server added at index %d %x at height %d", i, identityChainID.Bytes()[2:6], p.DBHeight))

	p.MakeMap()
	//p.State.AddStatus(fmt.Sprintf("PROCESSLIST.AddFedServer: Adding Server %x", identityChainID.Bytes()[3:8]))
	return i
}

// Add the given serverChain to this processlist as an Audit Server, and return
// the server index number of the added server
func (p *ProcessList) AddAuditServer(identityChainID interfaces.IHash) int {
	found, i := p.GetAuditServerIndexHash(identityChainID)
	if found {
		//p.State.AddStatus(fmt.Sprintf("ProcessList.AddAuditServer Server already there %x at height %d", identityChainID.Bytes()[2:6], p.DBHeight))
		return i
	}
	// If a fed server, demote
	fedFound, _ := p.GetFedServerIndexHash(identityChainID)
	if fedFound {
		//p.State.AddStatus(fmt.Sprintf("ProcessList.AddAuditServer Server %x was a fed server at height %d", identityChainID.Bytes()[2:6], p.DBHeight))
		p.RemoveFedServerHash(identityChainID)
	}

	InMsg := p.State.EFactory.NewAddAuditInternal(
		p.State.FactomNodeName,
		p.DBHeight,
		identityChainID,
	)
	p.State.electionsQueue.Enqueue(InMsg)

	p.AuditServers = append(p.AuditServers, nil)
	copy(p.AuditServers[i+1:], p.AuditServers[i:])
	p.AuditServers[i] = &Server{ChainID: identityChainID, Online: true}
	//p.State.AddStatus(fmt.Sprintf("PROCESSLIST.AddAuditServer Server added at index %d %x at height %d", i, identityChainID.Bytes()[2:6], p.DBHeight))

	return i
}

// Remove the given serverChain from this processlist's Federated Servers
func (p *ProcessList) RemoveFedServerHash(identityChainID interfaces.IHash) {
	found, i := p.GetFedServerIndexHash(identityChainID)
	if !found {
		p.RemoveAuditServerHash(identityChainID) // SOF-201
		return
	}

	InMsg := p.State.EFactory.NewRemoveLeaderInternal(
		p.State.FactomNodeName,
		p.DBHeight,
		identityChainID,
	)
	p.State.electionsQueue.Enqueue(InMsg)

	p.FedServers = append(p.FedServers[:i], p.FedServers[i+1:]...)
	p.MakeMap()
	//p.State.AddStatus(fmt.Sprintf("PROCESSLIST.RemoveFedServer: Removing Server %x", identityChainID.Bytes()[3:8]))
}

// Remove the given serverChain from this processlist's Audit Servers
func (p *ProcessList) RemoveAuditServerHash(identityChainID interfaces.IHash) {
	found, i := p.GetAuditServerIndexHash(identityChainID)
	if !found {
		return
	}

	InMsg := p.State.EFactory.NewRemoveAuditInternal(
		p.State.FactomNodeName,
		p.DBHeight,
		identityChainID,
	)
	p.State.electionsQueue.Enqueue(InMsg)

	p.AuditServers = append(p.AuditServers[:i], p.AuditServers[i+1:]...)
	//p.State.AddStatus(fmt.Sprintf("PROCESSLIST.RemoveAuditServer: Removing Audit Server %x", identityChainID.Bytes()[3:8]))
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

func (p *ProcessList) AddOldMsgs(m interfaces.IMsg) {
	p.oldmsgslock.Lock()
	defer p.oldmsgslock.Unlock()
	p.OldMsgs[m.GetHash().Fixed()] = m
}

func (p *ProcessList) DeleteOldMsgs(key interfaces.IHash) {
	p.oldmsgslock.Lock()
	defer p.oldmsgslock.Unlock()
	delete(p.OldMsgs, key.Fixed())
}

func (p *ProcessList) GetOldMsgs(key interfaces.IHash) interfaces.IMsg {
	if p == nil {
		return nil
	}
	if p.oldmsgslock == nil {
		return nil
	}
	p.oldmsgslock.Lock()
	defer p.oldmsgslock.Unlock()
	return p.OldMsgs[key.Fixed()]
}

func (p *ProcessList) AddNewEBlocks(key interfaces.IHash, value interfaces.IEntryBlock) {
	p.neweblockslock.Lock()
	defer p.neweblockslock.Unlock()
	p.NewEBlocks[key.Fixed()] = value
}

func (p *ProcessList) GetNewEBlocks(key interfaces.IHash) interfaces.IEntryBlock {
	p.neweblockslock.Lock()
	defer p.neweblockslock.Unlock()
	return p.NewEBlocks[key.Fixed()]
}

func (p *ProcessList) DeleteEBlocks(key interfaces.IHash) {
	p.neweblockslock.Lock()
	defer p.neweblockslock.Unlock()
	delete(p.NewEBlocks, key.Fixed())
}

func (p *ProcessList) AddNewEntry(key interfaces.IHash, value interfaces.IEntry) {
	p.NewEntriesMutex.Lock()
	defer p.NewEntriesMutex.Unlock()
	p.NewEntries[key.Fixed()] = value
}

func (p *ProcessList) DeleteNewEntry(key interfaces.IHash) {
	p.NewEntriesMutex.Lock()
	defer p.NewEntriesMutex.Unlock()
	delete(p.NewEntries, key.Fixed())
}

func (p *ProcessList) CurrentFault() *messages.FullServerFault {
	if len(p.System.List) < 1 || len(p.System.List) <= p.System.Height {
		return nil
	}
	return p.System.List[p.System.Height].(*messages.FullServerFault)
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

func (p *ProcessList) CheckDiffSigTally() bool {
	// If the majority of VMs' signatures do not match our
	// saved block, we discard that block from our database.
	if p.diffSigTally > 0 && p.diffSigTally > (len(p.FedServers)/2) {
		fmt.Println("**** dbstate diffSigTally", p.diffSigTally, "len/2", len(p.FedServers)/2)

		// p.State.DB.Delete([]byte(databaseOverlay.DIRECTORYBLOCK), p.State.ProcessLists.Lists[0].DirectoryBlock.GetKeyMR().Bytes())
		return false
	}

	return true
}

func (p *ProcessList) GetRequest(now int64, vmIndex int, height int, waitSeconds int64) *Request {
	r := new(Request)
	r.wait = waitSeconds
	r.vmIndex = vmIndex
	r.vmheight = uint32(height)

	if p.Requests[r.key()] == nil {
		r.sent = now + 2000
		p.Requests[r.key()] = r
	} else {
		r = p.Requests[r.key()]
	}

	return r

}

// Return the number of times we have tripped an ask for this request.
func (p *ProcessList) Ask(vmIndex int, height int, waitSeconds int64, tag int) int {
	now := p.State.GetTimestamp().GetTimeMilli()

	r := p.GetRequest(now, vmIndex, len(p.VMs[0].List), waitSeconds)

	if r == nil {
		return 0
	}

	if now-r.sent >= waitSeconds*1000+500 && p.State.inMsgQueue.Length() < constants.INMSGQUEUE_MED {
		missingMsgRequest := messages.NewMissingMsg(p.State, r.vmIndex, p.DBHeight, r.vmheight)

		// The System (handling full faults) is a special VM.  Let's guess it first.
		vm := &p.System
		if vmIndex >= 0 {
			// Ah, not the System VM, so let's look up the one we are really talking about.
			vm = p.VMs[vmIndex]
		}

		for k := range p.Requests {
			r2 := p.Requests[k]
			if r2.vmIndex == vmIndex && int(r2.vmheight) < vm.Height {
				delete(p.Requests, k)
			}
		}

		missingMsgRequest.AddHeight(uint32(height))
		// Okay, we are going to send one, so ask for all nil messages for this vm
		for i := 0; i < len(vm.List); i++ {
			if vm.List[i] == nil {
				missingMsgRequest.AddHeight(uint32(i))
			}
		}
		// Might as well as for the next message too.  Won't hurt.
		missingMsgRequest.AddHeight(uint32(len(vm.List)))

		if vmIndex < 0 {
			missingMsgRequest.SystemHeight = uint32(p.System.Height)
		}

		missingMsgRequest.SendOut(p.State, missingMsgRequest)
		p.State.MissingRequestAskCnt++

		r.sent = now
		r.requestCnt++
	}

	return r.requestCnt
}

func (p *ProcessList) TrimVMList(height uint32, vmIndex int) {
	if !(uint32(len(p.VMs[vmIndex].List)) > height) {
		p.VMs[vmIndex].List = p.VMs[vmIndex].List[:height]
	}
}

// Process messages and update our state.
func (p *ProcessList) Process(state *State) (progress bool) {
	dbht := state.GetHighestSavedBlk()
	if dbht >= p.DBHeight {
		//p.State.AddStatus(fmt.Sprintf("ProcessList.Process: VM Height is %d and Saved height is %d", dbht, state.GetHighestSavedBlk()))
		return true
	}

	state.PLProcessHeight = p.DBHeight

	if len(p.System.List) >= p.System.Height {
	systemloop:
		for i, f := range p.System.List[p.System.Height:] {
			if f == nil {
				p.Ask(-1, i, 10, 100)
				break systemloop
			}

			fault, ok := f.(*messages.FullServerFault)

			if ok {
				vm := p.VMs[fault.VMIndex]
				if vm.Height < int(fault.Height) {
					//p.State.AddStatus(fmt.Sprint("VM HEIGHT IS", vm.Height, "FH IS", fault.Height))
					break systemloop
				}
				if !fault.Process(p.DBHeight, p.State) {
					fault.SetAlreadyProcessed()
					break systemloop
				}
				p.System.Height++
				progress = true
			}
		}
	}

	for i := 0; i < len(p.FedServers); i++ {
		vm := p.VMs[i]

		if !p.State.Syncing {
			markNoFault(p, i)
		} else {
			if !vm.Synced {
				if vm.WhenFaulted == 0 {
					markFault(p, i, 0)
				}
			} else {
				if vm.FaultFlag == 0 {
					markNoFault(p, i)
				}
			}
		}

		FaultCheck(p)

		if vm.Height == len(vm.List) && p.State.Syncing && !vm.Synced {
			// means that we are missing an EOM
			p.Ask(i, vm.Height, 0, 1)
		}

		// If we haven't heard anything from a VM, ask for a message at the last-known height
		if vm.Height == len(vm.List) {
			p.Ask(i, vm.Height, 20, 2)
		}

	VMListLoop:
		for j := vm.Height; j < len(vm.List); j++ {
			if vm.List[j] == nil {
				//p.State.AddStatus(fmt.Sprintf("ProcessList.go Process: Found nil list at vm %d vm height %d ", i, j))
				p.Ask(i, j, 0, 3)
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
					//p.State.AddStatus(fmt.Sprintf("ProcessList.go Process: Error computing serial hash at dbht: %d vm %d  vm-height %d ", p.DBHeight, i, j))
					p.Ask(i, j, 3, 4)
					break VMListLoop
				}

				// compare the SerialHash of this acknowledgement with the
				// expected serialHash (generated above)
				if !expectedSerialHash.IsSameAs(thisAck.SerialHash) {
					//p.State.AddStatus(fmt.Sprintf("processList.Process(): SerialHash fail: dbht: %d vm %d msg %s", p.DBHeight, i, vm.List[j]))

					//fmt.Printf("dddd %20s %10s --- %10s %10x %10s %10x \n", "Conflict", p.State.FactomNodeName, "expected", expectedSerialHash.Bytes()[:3], "This", thisAck.Bytes()[:3])
					//fmt.Printf("dddd Error detected on %s\nSerial Hash failure: Fed Server %d  Leader ID %x List Ht: %d \nDetected on: %s\n",
					//	state.GetFactomNodeName(),
					//	i,
					//	p.FedServers[i].GetChainID().Bytes()[:3],
					//	j,
					//	vm.List[j].String())
					//fmt.Printf("dddd Last Ack: %6x  Last Serial: %6x\n", last.GetHash().Bytes()[:3], last.SerialHash.Bytes()[:3])
					//fmt.Printf("dddd This Ack: %6x  This Serial: %6x\n", thisAck.GetHash().Bytes()[:3], thisAck.SerialHash.Bytes()[:3])
					//fmt.Printf("dddd Expected: %6x\n", expectedSerialHash.Bytes()[:3])
					//fmt.Printf("dddd The message that didn't work: %s\n\n", vm.List[j].String())
					// the SerialHash of this acknowledgment is incorrect
					// according to this node's processList

					//fault(p, i, 0, vm, 0, j, 2)
					//p.State.AddStatus(fmt.Sprintf("ProcessList.go Process: SerialHash fails to match at dbht %d vm %d vm-height %d ", p.DBHeight, i, j))
					p.State.Reset()
					return
				}
			}

			// So here is the deal.  After we have processed a block, we have to allow the DirectoryBlockSignatures a chance to save
			// to disk.  Then we can insist on having the entry blocks.
			diff := p.DBHeight - state.EntryDBHeightComplete

			// Keep in mind, the process list is processing at a height one greater than the database. 1 is caught up.  2 is one behind.
			// Until the first couple signatures are processed, we will be 2 behind.
			if !p.State.WaitForEntries || (vm.LeaderMinute < 2 && diff <= 3) || diff <= 2 {
				// If we can't process this entry (i.e. returns false) then we can't process any more.
				p.NextHeightToProcess[i] = j + 1
				msg := vm.List[j]

				now := p.State.GetTimestamp()

				if _, valid := p.State.Replay.Valid(constants.INTERNAL_REPLAY, msg.GetRepeatHash().Fixed(), msg.GetTimestamp(), now); !valid {
					vm.List[j] = nil // If we have seen this message, we don't process it again.  Ever.
					break VMListLoop
				}

				if msg.Process(p.DBHeight, state) { // Try and Process this entry
					vm.heartBeat = 0
					vm.Height = j + 1 // Don't process it again if the process worked.

					progress = true

					// We have already tested and found m to be a new message.  We now record its hashes so later, we
					// can detect that it has been recorded.  We don't care about the results of IsTSValid_ at this point.
					p.State.Replay.IsTSValid_(constants.INTERNAL_REPLAY, msg.GetRepeatHash().Fixed(), msg.GetTimestamp(), now)
					p.State.Replay.IsTSValid_(constants.INTERNAL_REPLAY, msg.GetMsgHash().Fixed(), msg.GetTimestamp(), now)

					delete(p.State.Acks, msg.GetMsgHash().Fixed())
					delete(p.State.Holding, msg.GetMsgHash().Fixed())

				} else {
					//p.State.AddStatus(fmt.Sprintf("processList.Process(): Could not process entry dbht: %d VM: %d  msg: [[%s]]", p.DBHeight, i, msg.String()))
					break VMListLoop // Don't process further in this list, go to the next.
				}
			} else {
				// If we don't have the Entry Blocks (or we haven't processed the signatures) we can't do more.
				// p.State.AddStatus(fmt.Sprintf("Can't do more: dbht: %d vm: %d vm-height: %d Entry Height: %d", p.DBHeight, i, j, state.EntryDBHeightComplete))
				break VMListLoop
			}
		}
	}
	return
}

func (p *ProcessList) AddToSystemList(m interfaces.IMsg) bool {
	// Make sure we have a list, and punt if we don't.
	if p == nil {
		p.State.Holding[m.GetMsgHash().Fixed()] = m
		return false
	}

	fullFault, ok := m.(*messages.FullServerFault)
	if !ok {
		//p.State.AddStatus(fmt.Sprintf("FULL FAULT AddToSystemList Fail (not a FullFault) %s", m))
		return false // Should never happen;  Don't pass junk to be added to the System List
	}

	// If we have already processed past this fault, just ignore.
	if p.System.Height > int(fullFault.SystemHeight) {
		//p.State.AddStatus(fmt.Sprintf("FULL FAULT AddToSystemList Fail (System.Height > fullFault.SystemHeight) (%d > %d) : %s",
		//	p.System.Height,
		//	int(fullFault.SystemHeight),
		//	fullFault.String()))
		return false
	}

	// If the fault is in the future, hold it.
	if p.System.Height < int(fullFault.SystemHeight) {
		//p.State.AddStatus(fmt.Sprintf("FULL FAULT AddToSystemList Holding (System.Height(%d) < fullFault.SystemHeight(%d)) : %s",
		//	p.System.Height,
		//	int(fullFault.SystemHeight),
		//	fullFault.String()))
		p.State.Holding[m.GetMsgHash().Fixed()] = m
		return false
	}

	// If we are here, fullFault.SystemHeight == p.System.Height
	if len(p.System.List) <= p.System.Height {
		// Nothing in our list a this slot yet, so insert this FullFault message
		p.System.List = append(p.System.List, fullFault)
		//p.State.AddStatus(fmt.Sprintf("FULL FAULT AddToSystemList Success (append) : %s",
		//	fullFault.String()))
		return true
	}

	// Something is in our SystemList at this height;
	// We will prioritize the FullFault with the highest VMIndex
	existingSystemFault, _ := p.System.List[p.System.Height].(*messages.FullServerFault)
	if existingSystemFault.GetHash().IsSameAs(fullFault.GetHash()) {
		if p.VMs[existingSystemFault.VMIndex].WhenFaulted > 0 {
			//p.State.AddStatus(fmt.Sprintf("FULL FAULT AddToSystemList Fail (already have) : %s",
			//	fullFault.String()))
			return false
		}
	}

	if existingSystemFault.HasEnoughSigs(p.State) && p.State.pledgedByAudit(existingSystemFault) {
		//p.State.AddStatus(fmt.Sprintf("FULL FAULT AddToSystemList Fail (existingFault is complete) : %s",
		//	existingSystemFault.String()))
		return false
	}

	if fullFault.Priority(p.State) < existingSystemFault.Priority(p.State) {
		//p.State.AddStatus(fmt.Sprintf("FULL FAULT AddToSystemList Fail (priority %d < %d) :: Exist: %s /// New: %s",
		//	fullFault.Priority(p.State),
		//	existingSystemFault.Priority(p.State),
		//	existingSystemFault.String(),
		//	fullFault.String()))
		return false
	}

	if existingSystemFault.SigTally(p.State) > fullFault.SigTally(p.State) {
		if fullFault.GetCoreHash().IsSameAs(existingSystemFault.GetCoreHash()) {
			//p.State.AddStatus(fmt.Sprintf("FULL FAULT AddToSystemList Fail (less sigs than existingFault's) (%d > %d) : %s",
			//	existingSystemFault.SigTally(p.State),
			//	fullFault.SigTally(p.State),
			//	fullFault.String()))
			return false
		}
	}
	//p.State.regularFullFaultExecution(fullFault, p)

	p.System.List[p.System.Height] = fullFault

	//p.State.AddStatus(fmt.Sprintf("FULL FAULT AddToSystemList Success (create) : %s sigs:%d",
	//	fullFault.String(), fullFault.SigTally(p.State)))

	return true

}

func (p *ProcessList) AddToProcessList(ack *messages.Ack, m interfaces.IMsg) {
	if p == nil {
		return
	}

	if ack == nil || ack.GetMsgHash() == nil {
		return
	}

	TotalProcessListInputs.Inc()

	if ack.DBHeight > p.State.HighestAck && ack.Minute > 0 {
		p.State.HighestAck = ack.DBHeight
	}

	TotalAcksInputs.Inc()

	// If this is us, make sure we ignore (if old or in the ignore period) or die because two instances are running.
	//
	if !ack.Response && ack.LeaderChainID.IsSameAs(p.State.IdentityChainID) {
		now := p.State.GetTimestamp()
		if now.GetTimeSeconds()-ack.Timestamp.GetTimeSeconds() > 120 {
			// Us and too old?  Just ignore.
			return
		}

		num := p.State.GetSalt(ack.Timestamp)
		if num != ack.SaltNumber {
			os.Stderr.WriteString(fmt.Sprintf("This  AckHash    %x\n", ack.GetHash().Bytes()))
			os.Stderr.WriteString(fmt.Sprintf("This  ChainID    %x\n", p.State.IdentityChainID.Bytes()))
			os.Stderr.WriteString(fmt.Sprintf("This  Salt       %x\n", p.State.Salt.Bytes()[:8]))
			os.Stderr.WriteString(fmt.Sprintf("This  SaltNumber %x\n for this ack", num))
			os.Stderr.WriteString(fmt.Sprintf("Ack   ChainID    %x\n", ack.LeaderChainID.Bytes()))
			os.Stderr.WriteString(fmt.Sprintf("Ack   Salt       %x\n", ack.Salt))
			os.Stderr.WriteString(fmt.Sprintf("Ack   SaltNumber %x\n for this ack", ack.SaltNumber))
			panic("There are two leaders configured with the same Identity in this network!  This is a configuration problem!")
		}
	}

	toss := func(hint string) {
		TotalHoldingQueueOutputs.Inc()
		TotalAcksOutputs.Inc()
		delete(p.State.Holding, ack.GetHash().Fixed())
		delete(p.State.Acks, ack.GetHash().Fixed())
	}

	now := p.State.GetTimestamp()

	vm := p.VMs[ack.VMIndex]

	if len(vm.List) > int(ack.Height) && vm.List[ack.Height] != nil {
		_, isNew2 := p.State.Replay.Valid(constants.INTERNAL_REPLAY, m.GetRepeatHash().Fixed(), m.GetTimestamp(), now)
		if !isNew2 {
			toss("seen before, or too old")
			return
		}
	}

	if ack.DBHeight != p.DBHeight {
		// panic(fmt.Sprintf("Ack is wrong height.  Expected: %d Ack: ", p.DBHeight))
		return
	}

	if len(vm.List) > int(ack.Height) && vm.List[ack.Height] != nil {
		if vm.List[ack.Height].GetMsgHash().IsSameAs(m.GetMsgHash()) {
			toss("2")
			return
		}

		vm.List[ack.Height] = nil

		return
	}

	// From this point on, we consider the transaction recorded.  If we detect it has already been
	// recorded, then we still treat it as if we recorded it.

	vm.heartBeat = 0 // We have heard from this VM

	TotalHoldingQueueOutputs.Inc()
	TotalAcksOutputs.Inc()
	delete(p.State.Acks, m.GetMsgHash().Fixed())
	delete(p.State.Holding, m.GetMsgHash().Fixed())

	// Both the ack and the message hash to the same GetHash()
	m.SetLocal(false)
	ack.SetLocal(false)
	ack.SetPeer2Peer(false)
	m.SetPeer2Peer(false)

	ack.SendOut(p.State, ack)
	m.SendOut(p.State, m)

	for len(vm.List) <= int(ack.Height) {
		vm.List = append(vm.List, nil)
		vm.ListAck = append(vm.ListAck, nil)
	}

	delete(p.State.Acks, m.GetMsgHash().Fixed())
	p.VMs[ack.VMIndex].List[ack.Height] = m
	p.VMs[ack.VMIndex].ListAck[ack.Height] = ack
	p.AddOldMsgs(m)
	p.OldAcks[m.GetMsgHash().Fixed()] = ack

	plLogger.WithFields(log.Fields{"func": "AddToProcessList", "node-name": p.State.GetFactomNodeName(), "plheight": ack.Height, "dbheight": p.DBHeight}).WithFields(m.LogFields()).Info("Add To Process List")
}

func (p *ProcessList) ContainsDBSig(serverID interfaces.IHash) bool {
	for _, dbsig := range p.DBSignatures {
		if dbsig.ChainID.IsSameAs(serverID) {
			return true
		}
	}
	return false
}

func (p *ProcessList) AddDBSig(serverID interfaces.IHash, sig interfaces.IFullSignature) {
	found, _ := p.GetFedServerIndexHash(serverID)
	if !found || p.ContainsDBSig(serverID) {
		return // Duplicate, or not a federated server
	}
	dbsig := new(DBSig)
	dbsig.ChainID = serverID
	dbsig.Signature = sig
	found, dbsig.VMIndex = p.GetVirtualServers(0, serverID) //set the vmindex of the dbsig to the vm this server should sign
	if !found {                                             // Should never happen.
		return
	}
	p.DBSignatures = append(p.DBSignatures, *dbsig)
	p.SortDBSigs()
}

func (p *ProcessList) String() string {
	var buf primitives.Buffer
	if p == nil {
		buf.WriteString("-- <nil>\n")
	} else {
		buf.WriteString("===ProcessListStart===\n")

		pdbs := p.State.DBStates.Get(int(p.DBHeight - 1))
		saved := ""
		if pdbs == nil {
			saved = "nil"
		} else if pdbs.Signed {
			saved = "signed"
		} else if pdbs.Saved {
			saved = "saved"
		} else {
			saved = "constructing"
		}

		buf.WriteString(fmt.Sprintf("%s #VMs %d Complete %v DBHeight %d DBSig %v EOM %v p-dbstate = %s Entries Complete %d\n",
			p.State.GetFactomNodeName(),
			len(p.FedServers),
			p.Complete(),
			p.DBHeight,
			p.State.DBSig,
			p.State.EOM,
			saved,
			p.State.EntryDBHeightComplete))

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
					leader := fmt.Sprintf("[%x] ", vm.ListAck[j].LeaderChainID.Bytes()[:4])
					buf.WriteString("   " + leader + msg.String() + "\n")
				} else {
					buf.WriteString("   <nil>\n")
				}
			}
		}
		buf.WriteString(fmt.Sprintf("===FederatedServersStart=== %d\n", len(p.FedServers)))
		for _, fed := range p.FedServers {
			fedOnline := ""
			if !fed.IsOnline() {
				fedOnline = " F"
			}
			buf.WriteString(fmt.Sprintf("    %x%s\n", fed.GetChainID().Bytes()[:10], fedOnline))
		}
		buf.WriteString(fmt.Sprintf("===FederatedServersEnd=== %d\n", len(p.FedServers)))
		buf.WriteString(fmt.Sprintf("===AuditServersStart=== %d\n", len(p.AuditServers)))
		for _, aud := range p.AuditServers {
			audOnline := " offline"
			if aud.IsOnline() {
				audOnline = " online"
			}
			buf.WriteString(fmt.Sprintf("    %x%v\n", aud.GetChainID().Bytes()[:10], audOnline))
		}
		buf.WriteString(fmt.Sprintf("===AuditServersEnd=== %d\n", len(p.AuditServers)))
		buf.WriteString(fmt.Sprintf("===ProcessListEnd=== %s %d\n", p.State.GetFactomNodeName(), p.DBHeight))
	}
	return buf.String()
}

func (p *ProcessList) Reset() bool {
	return true
	previous := p.State.ProcessLists.Get(p.DBHeight - 1)

	if previous == nil {
		return false
	}

	//p.State.AddStatus(fmt.Sprintf("PROCESSLIST.Reset(): at dbht %d", p.DBHeight))

	// Make a copy of the previous FedServers
	p.System.List = p.System.List[:0]
	p.System.Height = 0
	p.Requests = make(map[[32]byte]*Request)
	//pl.Requests = make(map[[20]byte]*Request)

	p.FactoidBalancesT = map[[32]byte]int64{}
	p.ECBalancesT = map[[32]byte]int64{}

	p.FedServers = append(p.FedServers[:0], previous.FedServers...)
	p.AuditServers = append(p.AuditServers[:0], previous.AuditServers...)
	for _, auditServer := range p.AuditServers {
		auditServer.SetOnline(false)
		if p.State.GetIdentityChainID().IsSameAs(auditServer.GetChainID()) {
			// Always consider yourself "online"
			auditServer.SetOnline(true)
		}
	}
	for _, fedServer := range p.FedServers {
		fedServer.SetOnline(true)
	}
	p.SortFedServers()
	p.SortAuditServers()

	// empty my maps --
	p.OldMsgs = make(map[[32]byte]interfaces.IMsg)
	p.OldAcks = make(map[[32]byte]interfaces.IMsg)

	p.NewEBlocks = make(map[[32]byte]interfaces.IEntryBlock)
	p.NewEntries = make(map[[32]byte]interfaces.IEntry)

	p.SetAmINegotiator(false)

	p.DBSignatures = make([]DBSig, 0)

	// If a federated server, this is the server index, which is our index in the FedServers list

	var err error

	if previous != nil {
		p.DirectoryBlock = directoryBlock.NewDirectoryBlock(previous.DirectoryBlock)
		p.AdminBlock = adminBlock.NewAdminBlock(previous.AdminBlock)
		p.EntryCreditBlock, err = entryCreditBlock.NextECBlock(previous.EntryCreditBlock)
	} else {
		p.DirectoryBlock = directoryBlock.NewDirectoryBlock(nil)
		p.AdminBlock = adminBlock.NewAdminBlock(nil)
		p.EntryCreditBlock, err = entryCreditBlock.NextECBlock(nil)
	}
	if err != nil {
		panic(err.Error())
	}

	p.ResetDiffSigTally()

	for i := range p.FedServers {
		vm := p.VMs[i]

		vm.Height = 0 // Knock all the VMs back
		vm.LeaderMinute = 0
		vm.heartBeat = 0
		vm.Signed = false
		vm.Synced = false
		vm.WhenFaulted = 0

		p.VMs[i].List = p.VMs[i].List[:0]       // Knock all the lists back.
		p.VMs[i].ListAck = p.VMs[i].ListAck[:0] // Knock all the lists back.
		//p.State.SendDBSig(p.DBHeight, i)
	}

	/*fs := p.State.FactoidState.(*FactoidState)
	if previous.NextTimestamp != nil {
		fs.Reset(previous)
	}*/

	s := p.State
	s.LLeaderHeight--
	s.Saving = true
	s.Syncing = false
	s.EOM = false
	s.EOMDone = false
	s.DBSig = false
	s.DBSigDone = false
	s.CurrentMinute = 0
	s.EOMProcessed = 0
	s.DBSigProcessed = 0
	s.StartDelay = s.GetTimestamp().GetTimeMilli()
	s.RunLeader = false

	s.LLeaderHeight = s.GetHighestSavedBlk() + 1
	s.LeaderPL = s.ProcessLists.Get(s.LLeaderHeight)

	s.Leader, s.LeaderVMIndex = s.LeaderPL.GetVirtualServers(s.CurrentMinute, s.IdentityChainID)

	return true
}

/************************************************
 * Support
 ************************************************/

func NewProcessList(state interfaces.IState, previous *ProcessList, dbheight uint32) *ProcessList {
	// We default to the number of Servers previous.   That's because we always
	// allocate the FUTURE directoryblock, not the current or previous...

	state.AddStatus(fmt.Sprintf("PROCESSLISTS.NewProcessList at height %d", dbheight))
	pl := new(ProcessList)

	pl.State = state.(*State)

	// Make a copy of the previous FedServers
	pl.FedServers = make([]interfaces.IServer, 0)
	pl.AuditServers = make([]interfaces.IServer, 0)
	pl.Requests = make(map[[32]byte]*Request)
	//pl.Requests = make(map[[20]byte]*Request)

	pl.FactoidBalancesT = map[[32]byte]int64{}
	pl.ECBalancesT = map[[32]byte]int64{}

	if previous != nil {
		pl.FedServers = append(pl.FedServers, previous.FedServers...)
		pl.AuditServers = append(pl.AuditServers, previous.AuditServers...)
		for _, auditServer := range pl.AuditServers {
			auditServer.SetOnline(false)
			if state.GetIdentityChainID().IsSameAs(auditServer.GetChainID()) {
				// Always consider yourself "online"
				auditServer.SetOnline(true)
			}
		}
		for _, fedServer := range pl.FedServers {
			fedServer.SetOnline(true)
		}
		pl.SortFedServers()
	} else {
		pl.AddFedServer(state.GetNetworkBootStrapIdentity()) // Our default fed server, dependent on network type
		// pl.AddFedServer(primitives.Sha([]byte("FNode0"))) // Our default for now fed server on LOCAL network
	}

	// We just make lots of VMs as they have nearly no impact if not used.
	pl.VMs = make([]*VM, 65)
	for i := 0; i < 65; i++ {
		pl.VMs[i] = new(VM)
		pl.VMs[i].List = make([]interfaces.IMsg, 0)
		pl.VMs[i].Synced = true
		pl.VMs[i].WhenFaulted = 0
	}

	pl.DBHeight = dbheight

	pl.MakeMap()

	pl.PendingChainHeads = NewSafeMsgMap()
	pl.OldMsgs = make(map[[32]byte]interfaces.IMsg)
	pl.oldmsgslock = new(sync.Mutex)
	pl.OldAcks = make(map[[32]byte]interfaces.IMsg)
	pl.oldackslock = new(sync.Mutex)

	pl.NewEBlocks = make(map[[32]byte]interfaces.IEntryBlock)
	pl.neweblockslock = new(sync.Mutex)
	pl.NewEntries = make(map[[32]byte]interfaces.IEntry)

	pl.DBSignatures = make([]DBSig, 0)

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

// IsPendingChainHead returns if a chainhead is about to be updated (In PL)
func (p *ProcessList) IsPendingChainHead(chainid interfaces.IHash) bool {
	if p.PendingChainHeads.Get(chainid.Fixed()) != nil {
		return true
	}
	return false
}
