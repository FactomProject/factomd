// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/util/atomic"

	//"github.com/FactomProject/factomd/database/databaseOverlay"

	log "github.com/sirupsen/logrus"
)

var _ = fmt.Print
var _ = log.Print

var plLogger = packageLogger.WithFields(log.Fields{"subpack": "process-list"})

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
	oldmsgslock sync.Mutex

	// Chains that are executed, but not processed. There is a small window of a pending chain that the ack
	// will pass and the chainhead will fail. This covers that window. This is only used by WSAPI,
	// do not use it anywhere internally.
	PendingChainHeads *SafeMsgMap

	OldAcks     map[[32]byte]interfaces.IMsg
	oldackslock sync.Mutex

	// Entry Blocks added within 10 minutes (follower and leader)
	NewEBlocks     map[[32]byte]interfaces.IEntryBlock
	neweblockslock sync.Mutex

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

	// The Fedlist and Audlist at the START of the block. Server faults
	// can change the list, and we can calculate the deltas at the end
	StartingAuditServers []interfaces.IServer // List of Audit Servers
	StartingFedServers   []interfaces.IServer // List of Federated Servers

	// DB Sigs
	DBSignatures     []DBSig
	DBSigAlreadySent bool

	NextHeightToProcess [64]int

	// Cut overhead
	stringCnt int
}

var _ interfaces.IProcessList = (*ProcessList)(nil)

// Data needed to add to admin block
type DBSig struct {
	ChainID   interfaces.IHash
	Signature interfaces.IFullSignature
	VMIndex   int
}

type VM struct {
	List            []interfaces.IMsg    // Lists of acknowledged messages
	ListAck         []*messages.Ack      // Acknowledgements
	Height          int                  // Height of first unprocessed message (count of messages processed)
	EomMinuteIssued int                  // Last Minute Issued on this VM (from the leader, when we are the leader)
	LeaderMinute    int                  // Where the leader is in acknowledging messages
	Synced          bool                 // Is this VM synced yet?
	heartBeat       int64                // Just ping ever so often if we have heard nothing.
	Signed          bool                 // We have signed the previous block.
	WhenFaulted     int64                // WhenFaulted is a timestamp of when this VM was faulted
	FaultFlag       int                  // FaultFlag tracks what the VM was faulted for (0 = EOM missing, 1 = negotiation issue)
	ProcessTime     interfaces.Timestamp // Last time we made progress on this VM
	VmIndex         int                  // the index of this MV
	HighestAsk      int                  // highest ask sent to MMR for this VM
	HighestNil      int                  // Debug highest nil reported
	p               *ProcessList         // processList this VM part of
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

	if p == nil || p.State.OneLeader {
		return 0
	}

	v := uint64(0)
	for _, b := range hash {
		v += uint64(b)
	}
	r := int(v % uint64(len(p.FedServers)))
	return r
}

func (p *ProcessList) GetVMStatsForFedServer(index int) (vmIndex int, listHeight int, listLength int, nextNil int) {
	vmIndex = FedServerVM(p.ServerMap, len(p.FedServers), p.State.GetCurrentMinute(), index)
	if vmIndex < 0 {
		return vmIndex, -1, -1, -1
	}

	listHeight = p.VMs[vmIndex].Height
	listLength = len(p.VMs[vmIndex].List)
	nextNil = p.VMs[vmIndex].HighestNil

	return vmIndex, listHeight, listLength, nextNil
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

// duplicate function in election but cannot import because of a dependency loop
func Sort(serv []interfaces.IServer) bool {
	changed := false
	for i := 0; i < len(serv)-1; i++ {
		allgood := true
		for j := 0; j < len(serv)-1-i; j++ {
			if bytes.Compare(serv[j].GetChainID().Bytes(), serv[j+1].GetChainID().Bytes()) > 0 {
				s := serv[j]
				serv[j] = serv[j+1]
				serv[j+1] = s
				allgood = false
				changed = true
			}
		}
		if allgood {
			return changed
		}
	}
	return changed
}

func (p *ProcessList) LogPrintLeaders(log string) {
	s := p.State
	s.LogPrintf(log, "%6s | %6s", "Fed", "Aud")
	limit := len(p.FedServers)
	if limit < len(p.AuditServers) {
		limit = len(p.AuditServers)
	}
	for i := 0; i < limit; i++ {
		f := ""
		a := ""
		if i < len(p.FedServers) {
			f = fmt.Sprintf("%x", p.FedServers[i].GetChainID().Bytes()[3:6])
		}
		if i < len(p.AuditServers) {
			a = fmt.Sprintf("%x", p.AuditServers[i].GetChainID().Bytes()[3:6])
		}
		s.LogPrintf(log, "%s | %s", f, a)
	}
}
func (p *ProcessList) SortFedServers() {
	s := p.State
	if p.FedServers != nil {
		changed := Sort(p.FedServers)
		if changed {
			s.LogPrintf("election", "Sort changed p.Federated in ProcessList.SortFedServers")
			p.LogPrintLeaders("process")
		}
	}
}

func (p *ProcessList) SortAuditServers() {
	s := p.State
	if p.AuditServers != nil {
		changed := Sort(p.AuditServers)
		if changed {
			s.LogPrintf("election", "Sort changed p.Audit in ProcessList.SortAuditServers")
			p.LogPrintLeaders("process")
		}
	}
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
	//fmt.Fprintf(os.Stderr, "GetVirtualServers(%d,%x)", minute, identityChainID.Bytes()[3:6])
	found, fedIndex := p.GetFedServerIndexHash(identityChainID)
	if !found {
		return false, -1
	}

	p.MakeMap()

	if minute > 9 {
		minute = 9 // in case we get called between blocks.
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
		// Check first byte first.
		chainID := fs.GetChainID().Bytes()
		if scid[20] == chainID[20] {
			comp := bytes.Compare(scid, chainID)
			if comp == 0 {
				return true, i
			}
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
	//p.State.LogPrintf("executeMsg", "MakeMap(%d)", p.DBHeight)
	p.ServerMap = MakeMap(len(p.FedServers), p.DBHeight)
	//p.State.LogPrintf("executeMsg", "%s", p.PrintMap())

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

// Will set the starting fed/aud list for delta comparison at the end of the block
func (p *ProcessList) SetStartingAuthoritySet() {
	copyServer := func(os interfaces.IServer) interfaces.IServer {
		s := new(Server)
		s.ChainID = os.GetChainID()
		s.Name = os.GetName()
		s.Online = os.IsOnline()
		s.Replace = os.LeaderToReplace()
		return s
	}

	p.StartingFedServers = []interfaces.IServer{}
	p.StartingAuditServers = []interfaces.IServer{}

	for _, f := range p.FedServers {
		p.StartingFedServers = append(p.StartingFedServers, copyServer(f))
	}

	for _, a := range p.AuditServers {
		p.StartingAuditServers = append(p.StartingAuditServers, copyServer(a))
	}

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

	// Debugging a panic
	if p.State == nil {
		fmt.Println("-- Debug p.State", p.State)
	}

	if p.State.EFactory == nil {
		fmt.Println("-- Debug p.State.EFactory", p.State.EFactory, p.State.FactomNodeName)
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
func (p *ProcessList) GetAckAt(vmIndex int, height int) *messages.Ack {
	vm := p.VMs[vmIndex]
	if height < 0 || height >= len(vm.ListAck) {
		return nil
	}
	return vm.ListAck[height]
}

func (p *ProcessList) AddOldMsgs(m interfaces.IMsg) {
	p.oldmsgslock.Lock()
	defer p.oldmsgslock.Unlock()
	p.OldMsgs[m.GetHash().Fixed()] = m
}

func (p *ProcessList) GetOldAck(key interfaces.IHash) interfaces.IMsg {
	if p == nil {
		return nil
	}
	p.oldackslock.Lock()
	defer p.oldackslock.Unlock()
	a, ok := p.OldAcks[key.Fixed()]
	if !ok {
		return nil
	}
	return a
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

func (p *ProcessList) AddNewEntry(key interfaces.IHash, value interfaces.IEntry) {
	p.NewEntriesMutex.Lock()
	defer p.NewEntriesMutex.Unlock()
	p.NewEntries[key.Fixed()] = value
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

func (p *ProcessList) TrimVMList(h uint32, vmIndex int) {
	height := int(h)
	if len(p.VMs[vmIndex].List) < height {
		p.State.LogPrintf("processList", "TrimVMList() %d/%d/%d", p.DBHeight, vmIndex, height)
		p.VMs[vmIndex].List = p.VMs[vmIndex].List[:height]
		if p.State.DebugExec() {
			if p.VMs[vmIndex].HighestNil > height {
				p.VMs[vmIndex].HighestNil = height // Drag report limit back
			}
		}
		// make sure we will ask again for nil's above this height
		if p.VMs[vmIndex].HighestAsk > height {
			p.VMs[vmIndex].HighestAsk = height // Drag Ask limit back
		}
	} else {
		p.State.LogPrintf("process", "Attempt to trim higher than list list=%d h=%d", len(p.VMs[vmIndex].List), height)

	}
}
func (p *ProcessList) GetDBHeight() uint32 {
	return p.DBHeight
}

type foo struct {
	Syncing, DBSig, EOM, DBSigDone, EOMDone, EOMmax, EOMmin, DBSigMax, DBSigMin bool
}

var decodeMap map[foo]string = map[foo]string{
	//grep "Unexpected state" FNode0*process.txt | awk ' {print substr($0,index($0,"0x"));}' | sort -u
	//0x043 {true true false false false false true false false}
	//0x04b {true true false true false false true false false}
	//0x0cb {true true false true false false true true false}
	//0x10d {true false true true false false false false true}
	//0x11d {true false true true true false false false true}
	//0x12d {true false true true false true false false true}
	//0x13d {true false true true true true false false true}
	//0x140 {false false false false false false true false true}
	//0x148 {false false false true false false true false true}
	//0x14d {true false true true false false true false true}

	foo{true, true, false, false, false, false, true, false, false}:  "Syncing DBSig",             //0x043
	foo{true, true, false, true, false, false, true, false, false}:   "Syncing DBSig Done",        //0x04b
	foo{true, true, false, true, false, false, true, true, false}:    "Syncing DBSig Stop",        //0x0cb
	foo{true, false, true, true, false, false, false, false, true}:   "Syncing EOM",               //0x10d
	foo{true, false, true, true, true, false, false, false, true}:    "Syncing EOM Done",          //0x11d
	foo{true, false, true, true, false, true, false, false, true}:    "Syncing EOM Stop",          //0x12d
	foo{true, false, true, true, true, true, false, false, true}:     "Syncing EOM Done",          //0x13d
	foo{false, false, false, false, false, false, true, false, true}: "Normal (Begining of time)", //0x140
	foo{false, false, false, true, false, false, true, false, true}:  "Normal",                    //0x148
	foo{true, false, true, true, false, false, true, false, true}:    "Syncing EOM Start",         //0x14d

	// old code used to also hit these states some of which are problematic as they allow both DBSIG and EOM concurrently
	//foo{true, true, false, true, false, false, true, true, false}:     "Stop Syncing DBSig",              //0x0cb
	//foo{true, false, false, false, false, false, false, false, false}: "Sync Only??",                     //0x100 ***
	//foo{true, false, true, true, false, false, false, false, true}:   "Syncing EOM ... ",                 //0x10d
	//foo{true, true, false, false, false, false, true, false, true}:    "Start Syncing DBSig",             //0x143
	//foo{true, false, true, false, false, false, true, false, true}:    "Syncing EOM Start (DBSIG !Done)", //0x145 ***
	//foo{true, false, true, true, true, false, true, false, true}:      "Syncing EOM ... ",                //0x15d
}

func (p *ProcessList) decodeState(Syncing bool, DBSig bool, EOM bool, DBSigDone bool, EOMDone bool, FedServers int, EOMProcessed int, DBSigProcessed int) string {

	p.stringCnt++
	if p.stringCnt%1000 == 0 {
		return ""
	}

	if EOMProcessed > FedServers || EOMProcessed < 0 {
		p.State.LogPrintf("process", "Unexpected EOMProcessed %v of %v", EOMProcessed, FedServers)
	}
	if DBSigProcessed > FedServers || DBSigProcessed < 0 {
		p.State.LogPrintf("process", "Unexpected DBSigProcessed %v of %v", DBSigProcessed, FedServers)
	}

	var x foo = foo{Syncing, DBSig, EOM, DBSigDone, EOMDone,
		EOMProcessed == FedServers, EOMProcessed == 0, DBSigProcessed == FedServers, DBSigProcessed == 0}

	xx := 0
	var z []bool = []bool{Syncing, DBSig, EOM, DBSigDone, EOMDone, EOMProcessed == FedServers, EOMProcessed == 0, DBSigProcessed == FedServers, DBSigProcessed == 0}
	for i, b := range z {
		if b {
			xx = xx | (1 << uint(i))
		}
	}

	s, ok := decodeMap[x]
	if !ok {

		p.State.LogPrintf("process", "Unexpected 0x%03x %v", xx, x)
		s = "Unknown"
	}
	// divide ProcessListProcessCnt by a big number to make it not change the status string very often
	return fmt.Sprintf("SyncingStatus: %d-:-%d 0x%03x %25s EOM/DBSIG %02d/%02d of %02d -- %d",
		p.State.LeaderPL.DBHeight, p.State.CurrentMinute, xx, s, EOMProcessed, DBSigProcessed, FedServers, p.State.ProcessListProcessCnt/5000)

}

var extraDebug bool = false

// Process messages and update our state.
func (p *ProcessList) Process(s *State) (progress bool) {
	dbht := s.GetHighestSavedBlk()
	if dbht >= p.DBHeight {
		//s.AddStatus(fmt.Sprintf("ProcessList.Process: VM Height is %d and Saved height is %d", dbht, s.GetHighestSavedBlk()))
		return false
	}

	s.PLProcessHeight = p.DBHeight

	now := s.GetTimestamp()

	for i := 0; i < len(p.FedServers); i++ {
		vm := p.VMs[i]

		if vm.Height == len(vm.List) && s.Syncing && !vm.Synced {
			// means that we are missing an EOM
			vm.ReportMissing(vm.Height, 0)
		}

		// If we haven't heard anything from a VM in 2 seconds, ask for a message at the last-known height
		if vm.Height == len(vm.List) && now.GetTimeMilli()-vm.ProcessTime.GetTimeMilli() > 2000 {
			vm.ReportMissing(vm.Height, 2000) // Ask for one past the end of the list
		}

	VMListLoop:
		for j := vm.Height; j < len(vm.List); j++ {

			s.ProcessListProcessCnt++

			x := p.decodeState(s.Syncing, s.DBSig, s.EOM, s.DBSigDone, s.EOMDone,
				len(s.LeaderPL.FedServers), s.EOMProcessed, s.DBSigProcessed)

			// Compute a syncing s string and report if it has changed
			if x != "" && s.SyncingState[s.SyncingStateCurrent] != x {
				s.LogPrintf("processStatus", x)
				s.SyncingStateCurrent = (s.SyncingStateCurrent + 1) % len(s.SyncingState)
				s.SyncingState[s.SyncingStateCurrent] = x
			}
			if vm.List[j] == nil {
				//p.State.AddStatus(fmt.Sprintf("ProcessList.go Process: Found nil list at vm %d vm height %d ", i, j))
				cnt := 0
				for k := j; k < len(vm.List); k++ {
					if vm.List[k] == nil {
						cnt++
						vm.ReportMissing(k, 0)
					}
				}
				if s.DebugExec() {
					if vm.HighestNil < j {
						s.LogPrintf("process", "%d nils  at  %v/%v/%v", cnt, p.DBHeight, i, j)
						vm.HighestNil = j
					}
				}

				//				s.LogPrintf("process","nil  at  %v/%v/%v", p.DBHeight, i, j)
				break VMListLoop
			}

			if extraDebug {
				s.LogMessage("process", fmt.Sprintf("Consider %v/%v/%v", p.DBHeight, i, j), vm.List[j])
			}

			thisAck := vm.ListAck[j]
			thisMsg := vm.List[j]

			//todo: Need to re-validate the signatures of the message and ACK at this point to make sure they are current federated servers

			var expectedSerialHash interfaces.IHash
			var err error

			if vm.Height == 0 {
				expectedSerialHash = thisAck.SerialHash
			} else {
				last := vm.ListAck[vm.Height-1]
				expectedSerialHash, err = primitives.CreateHash(last.MessageHash, thisAck.MessageHash)
				if err != nil {
					s.LogMessage("process", fmt.Sprintf("nil out message %v/%v/%v, hash INTERNAL_REPLAY", p.DBHeight, i, j), vm.List[j]) //todo: revisit message
					vm.List[j] = nil
					if vm.HighestNil > j {
						vm.HighestNil = j // Drag report limit back
					}
					if vm.HighestAsk > j {
						vm.HighestAsk = j // Drag Ask limit back
					}
					//s.AddStatus(fmt.Sprintf("ProcessList.go Process: Error computing serial hash at dbht: %d vm %d  vm-height %d ", p.DBHeight, i, j))
					vm.ReportMissing(j, 0)
					//todo: report this... it's probably bad
					break VMListLoop
				}

				// compare the SerialHash of this acknowledgement with the
				// expected serialHash (generated above)
				if !expectedSerialHash.IsSameAs(thisAck.SerialHash) {
					s.LogMessage("process", "SerialHash Mismatch", thisMsg)
					s.LogMessage("process", "This ACK", thisAck)
					s.LogMessage("process", "Prev ACK", last)

					s.LogPrintf("process", "expected %x", expectedSerialHash.Bytes())
					s.LogPrintf("process", "thisAck  %x", thisAck.SerialHash.Bytes())
					s.Reset() // This currently does nothing.. see comments in reset
					//todo: report this... it's probably bad
					return
				}
			}

			// So here is the deal.  After we have processed a block, we have to allow the DirectoryBlockSignatures a chance to save
			// to disk.  Then we can insist on having the entry blocks.
			diff := p.DBHeight - s.EntryDBHeightComplete

			// Keep in mind, the process list is processing at a height one greater than the database. 1 is caught up.  2 is one behind.
			// Until the first couple signatures are processed, we will be 2 behind.
			//TODO: Why is this in the execution per message per VM when it's global to the processlist -- clay
			if s.WaitForEntries {
				s.LogPrintf("processList", "s.WaitForEntries %d-:-%d [%d] > %d + 2", p.DBHeight, vm.LeaderMinute, s.EntryDBHeightComplete)
				break VMListLoop // Don't process further in this list, go to the next.
			}

			// Try an process this message
			msg := thisMsg
			// If the block is not yet being written to disk (22 minutes old...)
			// dif >2 means the second pass sync is not complete so don't process yet.
			// this prevent you from becoming a leader when you don't have complete identities
			if (vm.LeaderMinute < 2 && diff <= 3) || diff <= 2 {
				// If we can't process this entry (i.e. returns false) then we can't process any more.
				p.NextHeightToProcess[i] = j + 1 // unused...

				now := p.State.GetTimestamp()

				msgRepeatHashFixed := msg.GetRepeatHash().Fixed()
				msgHashFixed := msg.GetMsgHash().Fixed()

				if _, valid := s.Replay.Valid(constants.INTERNAL_REPLAY, msgRepeatHashFixed, msg.GetTimestamp(), now); !valid {
					s.LogMessage("process", fmt.Sprintf("drop %v/%v/%v, hash INTERNAL_REPLAY", p.DBHeight, i, j), thisMsg)
					vm.List[j] = nil // If we have seen this message, we don't process it again.  Ever.
					if vm.HighestNil > j {
						vm.HighestNil = j // Drag report limit back
					}
					if vm.HighestAsk > j {
						vm.HighestAsk = j // Drag Ask limit back
					}
					//todo: report this... it's probably bad
					vm.ReportMissing(j, 0)
					break VMListLoop
				}
				vm.ProcessTime = now

				if msg.Process(p.DBHeight, s) { // Try and Process this entry

					p.State.LogMessage("processList", "done", msg)
					vm.heartBeat = 0
					vm.Height = j + 1 // Don't process it again if the process worked.
					s.LogMessage("process", fmt.Sprintf("done %v/%v/%v", p.DBHeight, i, j), msg)
					//s.LogPrintf("process", "thisAck  %x", thisAck.SerialHash.Bytes())

					progress = true

					// We have already tested and found m to be a new message.  We now record its hashes so later, we
					// can detect that it has been recorded.  We don't care about the results of IsTSValidAndUpdateState at this point.
					// block network replay too since we have already seen this message there is not need to see it again
					s.Replay.IsTSValidAndUpdateState(constants.INTERNAL_REPLAY|constants.NETWORK_REPLAY, msgRepeatHashFixed, msg.GetTimestamp(), now)
					s.Replay.IsTSValidAndUpdateState(constants.INTERNAL_REPLAY, msgHashFixed, msg.GetTimestamp(), now)

					delete(s.Acks, msgHashFixed)
					//delete(s.Holding, msgHashFixed)

					// REVIEW: does this leave msg in dependent holding?
					s.DeleteFromHolding(msgHashFixed, msg, "msg.Process done")
				} else {
					s.LogMessage("process", fmt.Sprintf("retry %v/%v/%v", p.DBHeight, i, j), msg)
					//s.AddStatus(fmt.Sprintf("processList.Process(): Could not process entry dbht: %d VM: %d  msg: [[%s]]", p.DBHeight, i, msg.String()))
					break VMListLoop // Don't process further in this list, go to the next.
				}
			} else {
				s.LogMessage("process", "Waiting on saving", msg)
				s.LogPrintf("EntrySync", "Waiting on saving EntryDBHeightComplete = %d", s.EntryDBHeightComplete)

				// If we don't have the Entry Blocks (or we haven't processed the signatures) we can't do more.
				// p.State.AddStatus(fmt.Sprintf("Can't do more: dbht: %d vm: %d vm-height: %d Entry Height: %d", p.DBHeight, i, j, s.EntryDBHeightComplete))
				if extraDebug {
					p.State.LogPrintf("process", "Waiting on saving blocks to progress complete %d processing %d-:-%d", s.EntryDBHeightComplete, p.DBHeight, vm.LeaderMinute)
				}
				break VMListLoop
			}
		}
	}
	return
}

func (p *ProcessList) AddToProcessList(s *State, ack *messages.Ack, m interfaces.IMsg) {
	//s.LogMessage("processList", "Message:", m) // also logged with the ack
	s.LogMessage("processList", "Ack:", ack)
	if p == nil {
		s.LogPrintf("processList", "Drop no process list to add to")
		return
	}

	if ack == nil {
		s.LogPrintf("processList", "drop Ack==nil")
		return
	}
	if ack.GetMsgHash() == nil {
		s.LogPrintf("processList", "Drop ack.GetMsgHash() == nil")
		return
	}

	TotalProcessListInputs.Inc()
	messageHash := ack.GetHash() // This is the has of the message bring acknowledged not the hash of the ack message
	msgHash := m.GetMsgHash()
	if !messageHash.IsSameAs(msgHash) {
		panic("Hash mismatch")
	}

	TotalProcessListInputs.Inc()

	// Make sure we don't put in an old ack (outside our repeat range)
	blktime := s.GetLeaderTimestamp().GetTime().UnixNano()
	tlim := int64(Range * 60 * 1000 * 1000 * 1000)

	if blktime != 0 && m.Type() != constants.DIRECTORY_BLOCK_SIGNATURE_MSG {
		acktime := ack.GetTimestamp().GetTime().UnixNano()
		msgtime := m.GetTimestamp().GetTime().UnixNano()
		Delta := blktime - acktime

		if Delta > tlim || -Delta > tlim {
			p.State.LogPrintf("processList", "Drop message pair, because the ack is out of range")
			return
		}

		// Make sure we don't put in an old msg (outside our repeat range)
		Delta = blktime - msgtime
		if Delta > tlim || -Delta > tlim {
			p.State.LogPrintf("processList", "Drop message pair, because the msg is out of range")
			return
		}
	}

	if ack.DBHeight > s.HighestAck && ack.Minute > 0 {
		s.HighestAck = ack.DBHeight

	}

	TotalAcksInputs.Inc()

	// If this is us, make sure we ignore (if old or in the ignore period) or die because two instances are running.
	//
	if !ack.Response && ack.LeaderChainID.IsSameAs(s.IdentityChainID) {
		now := s.GetTimestamp().GetTimeSeconds()
		ackSeconds := ack.Timestamp.GetTimeSeconds()
		if now-ackSeconds > 120 {
			s.LogPrintf("processList", "Drop old msg")
			// Us and too old?  Just ignore.
			return
		}

		num := s.GetSalt(ack.Timestamp)
		if num != ack.SaltNumber {
			os.Stderr.WriteString(fmt.Sprintf("This  AckHash    %x\n", messageHash.Bytes()))
			os.Stderr.WriteString(fmt.Sprintf("This  ChainID    %x\n", s.IdentityChainID.Bytes()))
			os.Stderr.WriteString(fmt.Sprintf("This  Salt       %x\n", s.Salt.Bytes()[:8]))
			os.Stderr.WriteString(fmt.Sprintf("This  SaltNumber %x\n for this ack", num))
			os.Stderr.WriteString(fmt.Sprintf("Ack   ChainID    %x\n", ack.LeaderChainID.Bytes()))
			os.Stderr.WriteString(fmt.Sprintf("Ack   Salt       %x\n", ack.Salt))
			os.Stderr.WriteString(fmt.Sprintf("Ack   SaltNumber %x\n for this ack", ack.SaltNumber))
			panic("There are two leaders configured with the same Identity in this network!  This is a configuration problem!")
		}
	}

	toss := func(hint string) {
		s.LogPrintf("processList", "Drop "+hint)
		TotalHoldingQueueOutputs.Inc()
		TotalAcksOutputs.Inc()
		//delete(s.Holding, msgHash.Fixed())

		s.DeleteFromHolding(m.GetMsgHash().Fixed(), m, "Toss"+hint)
		delete(s.Acks, msgHash.Fixed())
	}

	now := s.GetTimestamp()

	vm := p.VMs[ack.VMIndex]

	if len(vm.List) > int(ack.Height) && vm.List[ack.Height] != nil {
		_, isNew2 := s.Replay.Valid(constants.INTERNAL_REPLAY, m.GetRepeatHash().Fixed(), m.GetTimestamp(), now)
		if !isNew2 {
			toss("seen before, or too old")
			return
		}
	}

	if ack.DBHeight != p.DBHeight {
		// panic(fmt.Sprintf("Ack is wrong height.  Expected: %d Ack: ", p.DBHeight))
		s.LogPrintf("processList", "Drop Ack is wrong height.  Expected: %d Ack: ", p.DBHeight)
		return
	}

	if len(vm.List) > int(ack.Height) && vm.List[ack.Height] != nil {
		if vm.List[ack.Height].GetMsgHash().IsSameAs(msgHash) {
			s.LogPrintf("processList", "Drop duplicate")
			toss("Drop duplicate")
			return
		}

		s.LogMessage("processList", "drop from pl", vm.List[ack.Height])
		s.LogMessage("processList", "drop from pl", vm.ListAck[ack.Height])
		// the code below will blindly overwrite the old message/ack
	}

	// From this point on, we consider the transaction recorded.  If we detect it has already been
	// recorded, then we still treat it as if we recorded it.

	vm.heartBeat = 0 // We have heard from this VM

	// Both the ack and the message hash to the same GetHash()
	if ack.GetHash().Fixed() != m.GetMsgHash().Fixed() {
		s.LogPrintf("executeMsg", "m/ack mismatch m-%x a-%x", m.GetMsgHash().Fixed(), ack.GetHash().Fixed())
	}

	if ack.GetHash().Fixed() != m.GetMsgHash().Fixed() {
		s.LogPrintf("executeMsg", "m/ack mismatch m-%x a-%x", m.GetMsgHash().Fixed(), ack.GetHash().Fixed())
	}

	for len(vm.List) <= int(ack.Height) {
		vm.List = append(vm.List, nil)
		vm.ListAck = append(vm.ListAck, nil)
	}

	s.LogPrintf("executeMsg", "remove from holding M-%v|R-%v", m.GetMsgHash().String()[:6], m.GetRepeatHash().String()[:6])
	TotalHoldingQueueOutputs.Inc()
	TotalAcksOutputs.Inc()
	s.DeleteFromHolding(msgHash.Fixed(), m, "Process()")
	delete(s.Acks, msgHash.Fixed())
	p.VMs[ack.VMIndex].List[ack.Height] = m
	p.VMs[ack.VMIndex].ListAck[ack.Height] = ack
	p.AddOldMsgs(m)
	p.OldAcks[msgHash.Fixed()] = ack

	if s.adds != nil {
		s.adds <- plRef{int(p.DBHeight), ack.VMIndex, int(ack.Height)}
	}

	s.LogMessage("processList", fmt.Sprintf("Added at %d/%d/%d by %s", ack.DBHeight, ack.VMIndex, ack.Height, atomic.WhereAmIString(1)), m)

	// If we add the message to the process list, ensure we actually process that
	// message, so the next msg will be able to added without going into holding.
	if ack.IsLocal() {
		for p.Process(s) {
		}
	}

	// Both the ack and the message hash to the same GetHash()
	ack.SetLocal(false)
	ack.SetPeer2Peer(false)
	m.SetPeer2Peer(false)
	m.SetLocal(false)

	m.SendOut(s, m)
	ack.SendOut(s, ack)
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
					leader := fmt.Sprintf("[%x] ", vm.ListAck[j].LeaderChainID.Bytes()[3:6])
					msgStr := msg.String()
					index := strings.Index(msgStr, "\n")
					if index > 0 {
						msgStr = msgStr[0:index]
					}
					index = len(msgStr)
					if index > 128 {
						msgStr = msgStr[0:125] + "..."
					}
					buf.WriteString("   " + leader + msgStr + "\n")
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

// Intended to let a demoted leader come back before the next DB state but interfered with boot under load so disable for now
// that means demoted leaders are not sane till the next DBState (up to 10 minutes). Maybe revisit after the missing message storms are fixed.
func (p *ProcessList) Reset() bool {
	p.State.LogPrintf("processList", "ProcessList.Reset %d minute %s @ %s", p.DBHeight, p.State.CurrentMinute, atomic.WhereAmIString(1))

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
	//pl.Requests = make(map[[20]byte]*Request)

	pl.FactoidBalancesTMutex.Lock()
	pl.FactoidBalancesT = map[[32]byte]int64{}
	pl.FactoidBalancesTMutex.Unlock()

	pl.ECBalancesTMutex.Lock()
	pl.ECBalancesT = map[[32]byte]int64{}
	pl.ECBalancesTMutex.Unlock()

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

	now := state.GetTimestamp()
	// We just make lots of VMs as they have nearly no impact if not used.
	pl.VMs = make([]*VM, 65)
	for i := 0; i < 65; i++ {
		pl.VMs[i] = new(VM)
		pl.VMs[i].List = make([]interfaces.IMsg, 0)
		pl.VMs[i].Synced = true
		pl.VMs[i].WhenFaulted = 0
		pl.VMs[i].ProcessTime = now
		pl.VMs[i].VmIndex = i
		pl.VMs[i].p = pl
	}

	pl.DBHeight = dbheight

	pl.MakeMap()

	pl.PendingChainHeads = NewSafeMsgMap("PendingChainHeads", pl.State)
	pl.OldMsgs = make(map[[32]byte]interfaces.IMsg)
	pl.OldAcks = make(map[[32]byte]interfaces.IMsg)

	pl.NewEBlocks = make(map[[32]byte]interfaces.IEntryBlock)
	pl.NewEntries = make(map[[32]byte]interfaces.IEntry)

	pl.DBSignatures = make([]DBSig, 0)

	// If a federated server, this is the server index, which is our index in the FedServers list

	var err error

	if previous != nil {
		pl.DirectoryBlock = directoryBlock.NewDirectoryBlock(previous.DirectoryBlock)
		pl.AdminBlock = adminBlock.NewAdminBlock(previous.AdminBlock)
		pl.EntryCreditBlock, err = entryCreditBlock.NextECBlock(previous.EntryCreditBlock)
	} else {
		if pl.DBHeight > 0 {
			pl.DirectoryBlock, _ = state.GetDB().FetchDBlockByHeight(pl.DBHeight)
			pl.AdminBlock, _ = state.GetDB().FetchABlockByHeight(pl.DBHeight)
			pl.EntryCreditBlock, _ = state.GetDB().FetchECBlockByHeight(pl.DBHeight)
		} else {
			pl.DirectoryBlock = directoryBlock.NewDirectoryBlock(nil)
			pl.AdminBlock = adminBlock.NewAdminBlock(nil)
			pl.EntryCreditBlock, err = entryCreditBlock.NextECBlock(nil)
		}
	}

	pl.ResetDiffSigTally()

	if pl.DirectoryBlock != nil {
		pl.DirectoryBlock.GetHeader().SetTimestamp(now) // Well this is awkwardly after it's created but ....
	}
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
