// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"fmt"

	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/util"
	"hash"
	"os"
)

var _ = fmt.Print
var _ = (*hash.Hash32)(nil)

//***************************************************************
// Process Loop for Consensus
//
// Returns true if some message was processed.
//***************************************************************
func (s *State) NewMinute() {
	s.LeaderPL.Unseal(s.EOM)
	// Anything we are holding, we need to reprocess.
	for k := range s.Holding {
		v := s.Holding[k]
		v.ComputeVMIndex(s)
		if s.Leader {
			s.XReview = append(s.XReview, v)
			delete(s.Holding, k)
		}
	}

	s.EOM = 0

	for k := range s.Acks {
		s.StallMsg(s.Acks[k])
	}

	s.LeaderPL = s.ProcessLists.Get(s.LLeaderHeight)
	s.Leader, s.LeaderVMIndex = s.LeaderPL.GetVirtualServers(s.LeaderMinute, s.IdentityChainID)
	/**
	fmt.Println(s.FactomNodeName, ">>>")

	for k := range s.Holding {
		m := s.Holding[k]
		fmt.Println(s.FactomNodeName, ">>>", m.String())
	}
	**/
}

func (s *State) Process() (progress bool) {

	//s.DebugPrt("Process")

	highest := s.GetHighestRecordedBlock()

	dbstate := s.DBStates.Get(s.LLeaderHeight)
	if s.LLeaderHeight <= highest && (dbstate == nil || dbstate.Saved) {
		s.LLeaderHeight = highest + 1
		s.LeaderPL = s.ProcessLists.Get(s.LLeaderHeight)
		s.Leader, s.LeaderVMIndex = s.LeaderPL.GetVirtualServers(0, s.IdentityChainID)

		if s.Leader && dbstate != nil {
			dbs := new(messages.DirectoryBlockSignature)
			dbs.DirectoryBlockKeyMR = dbstate.DirectoryBlock.GetKeyMR()
			dbs.ServerIdentityChainID = s.GetIdentityChainID()
			dbs.DBHeight = s.LLeaderHeight
			dbs.Timestamp = s.GetTimestamp()
			dbs.SetVMHash(nil)
			dbs.SetVMIndex(s.LeaderVMIndex)
			dbs.SetLocal(true)
			dbs.Sign(s)
			err := dbs.Sign(s)
			if err != nil {
				panic(err)
			}
			dbs.LeaderExecute(s)
		}
		s.LeaderMinute = 0
		s.NewMinute()
	}

	if s.EOM > 0 && s.LeaderPL.Unsealable(s.EOM) {
		s.LeaderMinute++

		switch {
		case s.LeaderMinute <= 9:
			s.LeaderPL = s.ProcessLists.Get(s.LLeaderHeight)
			s.Leader, s.LeaderVMIndex = s.LeaderPL.GetVirtualServers(s.LeaderMinute, s.IdentityChainID)
			s.NewMinute()
		case s.LeaderMinute == 10:
			s.AddDBState(true, s.LeaderPL.DirectoryBlock, s.LeaderPL.AdminBlock, s.GetFactoidState().GetCurrentBlock(), s.LeaderPL.EntryCreditBlock)
			s.LeaderPL = s.ProcessLists.Get(s.LLeaderHeight + 1)
			s.Leader, s.LeaderVMIndex = s.LeaderPL.GetVirtualServers(0, s.IdentityChainID)
		}

	}

	return s.ProcessQueues()
}

func (s *State) ProcessQueues() (progress bool) {

	// Reprocess any stalled Acknowledgements
	for s.Leader && s.EOM == 0 && len(s.XReview) > 0 {
		msg := s.XReview[0]
		if _, ok := s.InternalReplay.Valid(msg.GetHash().Fixed(), int64(msg.GetTimestamp()), int64(s.GetTimestamp())); !ok {
			msg = nil
		} else if s.Leader {
			msg.LeaderExecute(s)
		} else {
			msg.FollowerExecute(s)
		}
		s.XReview = s.XReview[1:]
		s.UpdateState()
	}

	select {
	case ack := <-s.ackQueue:
		_, ok := s.InternalReplay.Valid(ack.GetHash().Fixed(), int64(ack.GetTimestamp()), int64(s.GetTimestamp()))
		if ok && ack.Validate(s) == 1 {
			ack.FollowerExecute(s)
		}
		progress = true
	case ack := <-s.stallQueue:
		_, ok := s.InternalReplay.Valid(ack.GetHash().Fixed(), int64(ack.GetTimestamp()), int64(s.GetTimestamp()))
		if ok && ack.Validate(s) == 1 {
			ack.FollowerExecute(s)
			progress = true
		}
		progress = true
	case msg := <-s.msgQueue:
		_, ok := s.InternalReplay.Valid(msg.GetHash().Fixed(), int64(msg.GetTimestamp()), int64(s.GetTimestamp()))
		if ok {
			msg.ComputeVMIndex(s)
			switch msg.Validate(s) {
			case 1:
				if s.EOM == 0 {
					if s.Leader {
						msg.LeaderExecute(s)
					} else {
						s.networkOutMsgQueue <-msg
						msg.FollowerExecute(s)
					}
				} else {
					s.networkOutMsgQueue <-msg
					msg.FollowerExecute(s)
				}

			case 0: // Put at the end of the line, and hopefully we will resolve it.
				s.msgQueue <- msg
			default:
				delete(s.Acks, msg.GetHash().Fixed())
				s.networkInvalidMsgQueue <- msg
			}
		}
	default:
	}
	return
}

//***************************************************************
// Consensus Methods
//***************************************************************

// Adds blocks that are either pulled locally from a database, or acquired from peers.
func (s *State) AddDBState(isNew bool,
	directoryBlock interfaces.IDirectoryBlock,
	adminBlock interfaces.IAdminBlock,
	factoidBlock interfaces.IFBlock,
	entryCreditBlock interfaces.IEntryCreditBlock) {

	// TODO:  Need to validate before we add, or at least validate once we have a contiguous set of blocks.

	// 	fmt.Printf("AddDBState %s: DirectoryBlock %d %x %x %x %x\n",
	// 			   s.FactomNodeName,
	// 			   directoryBlock.GetHeader().GetDBHeight(),
	// 			   directoryBlock.GetKeyMR().Bytes()[:5],
	// 			   adminBlock.GetHash().Bytes()[:5],
	// 			   factoidBlock.GetHash().Bytes()[:5],
	// 			   entryCreditBlock.GetHash().Bytes()[:5])

	dbState := s.DBStates.NewDBState(isNew, directoryBlock, adminBlock, factoidBlock, entryCreditBlock)
	s.DBStates.Put(dbState)
	ht := dbState.DirectoryBlock.GetHeader().GetDBHeight()
	if ht > s.LLeaderHeight {
		s.LLeaderHeight = ht
		s.ProcessLists.Get(ht + 1)
		s.EOM = 0
	}
	//	dbh := directoryBlock.GetHeader().GetDBHeight()
	//	if s.LLeaderHeight < dbh {
	//		s.LLeaderHeight = dbh + 1
	//	}
}

func (s *State) addEBlock(eblock interfaces.IEntryBlock) {
	hash, err := eblock.KeyMR()

	if err == nil {
		if s.HasDataRequest(hash) {

			s.DB.ProcessEBlockBatch(eblock, true)
			delete(s.DataRequests, hash.Fixed())

			if s.GetAllEntries(hash) {
				if s.GetEBDBHeightComplete() < eblock.GetDatabaseHeight() {
					s.SetEBDBHeightComplete(eblock.GetDatabaseHeight())
				}
			}
		}
	}
}

// Messages that will go into the Process List must match an Acknowledgement.
// The code for this is the same for all such messages, so we put it here.
//
// Returns true if it finds a match, puts the message in holding, or invalidates the message
func (s *State) FollowerExecuteMsg(m interfaces.IMsg) {

	if eom, ok := m.(*messages.EOM); ok && m.IsLocal() {
		return // This is an internal EOM message.  We are not a leader so ignore.
	} else if ok && !s.Leader {
		s.EOM = int(eom.Minute + 1)
	}

	hash := m.GetHash()
	hashf := hash.Fixed()
	s.Holding[hashf] = m
	ack, _ := s.Acks[hashf].(*messages.Ack)
	if ack != nil {
		s.LeaderPL.AddToProcessList(ack, m)
	}
}

// Ack messages always match some message in the Process List.   That is
// done here, though the only msg that should call this routine is the Ack
// message.
func (s *State) FollowerExecuteAck(msg interfaces.IMsg) {
	ack := msg.(*messages.Ack)
	s.Acks[ack.GetHash().Fixed()] = ack
	m, _ := s.Holding[ack.GetHash().Fixed()]
	if m != nil {
		s.LeaderPL.AddToProcessList(ack, m)
	}
}

func (s *State) FollowerExecuteDBState(msg interfaces.IMsg) {

	dbstatemsg, _ := msg.(*messages.DBStateMsg)

	s.DBStates.LastTime = s.GetTimestamp()
	s.AddDBState(true,
		dbstatemsg.DirectoryBlock,
		dbstatemsg.AdminBlock,
		dbstatemsg.FactoidBlock,
		dbstatemsg.EntryCreditBlock)

}

func (s *State) FollowerExecuteAddData(msg interfaces.IMsg) {
	dataResponseMsg, ok := msg.(*messages.DataResponse)
	if !ok {
		return
	}

	switch dataResponseMsg.DataType {
	case 0: // DataType = entry
		entry := dataResponseMsg.DataObject.(interfaces.IEBEntry)

		if entry.GetHash().IsSameAs(dataResponseMsg.DataHash) {

			s.DB.InsertEntry(entry)
			delete(s.DataRequests, entry.GetHash().Fixed())
		}
	case 1: // DataType = eblock
		eblock := dataResponseMsg.DataObject.(interfaces.IEntryBlock)
		dataHash, _ := eblock.KeyMR()
		if dataHash.IsSameAs(dataResponseMsg.DataHash) {
			s.addEBlock(eblock)
		}
	default:
		s.networkInvalidMsgQueue <- msg
	}

}

func (s *State) LeaderExecute(m interfaces.IMsg) {

	if !s.Leader || s.EOM > 0 || m.GetVMIndex() != s.LeaderVMIndex {
		m.FollowerExecute(s)
		return
	}

	vm := s.LeaderPL.VMs[s.LeaderVMIndex]
	if len(vm.List) > vm.Height {
		s.msgQueue <- m
	}

	ack := s.NewAck(m)

	s.LeaderPL.AddToProcessList(ack.(*messages.Ack), m)

}

func (s *State) LeaderExecuteEOM(m interfaces.IMsg) {

	vm := s.LeaderPL.VMs[s.LeaderVMIndex]
	if len(vm.List) > vm.Height {
		s.msgQueue <- m
	}

	if !m.IsLocal() {
		m.FollowerExecute(s)
		return
	}

	eom := m.(*messages.EOM)

	s.EOM = int(s.LeaderMinute + 1)
	if s.LeaderPL.VMIndexFor(constants.FACTOID_CHAINID) == s.LeaderVMIndex {
		eom.FactoidVM = true
	}
	eom.DBHeight = s.LLeaderHeight
	eom.VMIndex = s.LeaderVMIndex
	eom.Minute = byte(s.LeaderMinute)
	eom.Sign(s)
	eom.SetLocal(false)
	ack := s.NewAck(m)
	s.LeaderPL.AddToProcessList(ack.(*messages.Ack), eom)

}

func (s *State) ProcessAddServer(dbheight uint32, addServerMsg interfaces.IMsg) bool {
	as, ok := addServerMsg.(*messages.AddServerMsg)
	if !ok {
		return true
	}

	if leader, _ := s.LeaderPL.GetFedServerIndexHash(as.ServerChainID); leader {
		return true
	}

	if as.ServerType == 0 {
		s.LeaderPL.AdminBlock.AddFedServer(as.ServerChainID)
	}

	return true
}

func (s *State) ProcessCommitChain(dbheight uint32, commitChain interfaces.IMsg) bool {
	c, _ := commitChain.(*messages.CommitChainMsg)

	pl := s.ProcessLists.Get(dbheight)
	pl.EntryCreditBlock.GetBody().AddEntry(c.CommitChain)
	s.GetFactoidState().UpdateECTransaction(true, c.CommitChain)

	// save the Commit to match agains the Reveal later
	s.PutCommits(c.CommitChain.EntryHash, c)

	return true
}

func (s *State) ProcessCommitEntry(dbheight uint32, commitEntry interfaces.IMsg) bool {
	c, _ := commitEntry.(*messages.CommitEntryMsg)

	pl := s.ProcessLists.Get(dbheight)
	pl.EntryCreditBlock.GetBody().AddEntry(c.CommitEntry)
	s.GetFactoidState().UpdateECTransaction(true, c.CommitEntry)

	// save the Commit to match agains the Reveal later
	s.PutCommits(c.CommitEntry.EntryHash, c)

	return true
}

// TODO: Should fault the server if we don't have the proper sequence of EOM messages.
func (s *State) ProcessEOM(dbheight uint32, msg interfaces.IMsg) bool {

	e, ok := msg.(*messages.EOM)
	if !ok {
		panic("Must pass an EOM message to ProcessEOM)")
	}

	if s.EOM == 0 && !s.Leader {
		s.EOM = int(e.Minute + 1)
	}

	pl := s.ProcessLists.Get(dbheight)

	if pl.MinuteComplete() < s.LeaderMinute {
		return false
	}

	if e.FactoidVM {
		s.FactoidState.EndOfPeriod(int(e.Minute))

		// Add EOM to the EBlocks.  We only do this once, so
		// we piggy back on the fact that we only do the FactoidState
		// EndOfPeriod once too.
	}

	for _, eb := range pl.NewEBlocks {
		if pl.VMIndexFor(eb.GetChainID().Bytes()) == e.VMIndex {
			eb.AddEndOfMinuteMarker(e.Bytes()[0])
		}
	}

	if pl.VMIndexFor(constants.ADMIN_CHAINID) == e.VMIndex {
		pl.AdminBlock.AddEndOfMinuteMarker(e.Minute)
	}

	if pl.VMIndexFor(constants.EC_CHAINID) == e.VMIndex {
		ecblk := pl.EntryCreditBlock
		ecbody := ecblk.GetBody()
		mn := entryCreditBlock.NewMinuteNumber2(e.Minute)
		ecbody.AddEntry(mn)
	}

	vm := pl.VMs[e.VMIndex]

	vm.MinuteFinished = int(e.Minute) + 1

	return true
}

func (s *State) ProcessRevealEntry(dbheight uint32, m interfaces.IMsg) bool {
	msg := m.(*messages.RevealEntryMsg)
	myhash := msg.GetHash()
	commit := s.GetCommits(myhash)
	if commit == nil {
		return false
	}

	if _, isNewChain := commit.(*messages.CommitChainMsg); isNewChain {
		chainID := msg.Entry.GetChainID()
		eb, err := s.DB.FetchEBlockHead(chainID)
		if err != nil || eb != nil {
			panic(fmt.Sprintf("%s\n%s", "Chain already exists", msg.String()))
		}

		// Create a new Entry Block for a new Entry Block Chain
		eb = entryBlock.NewEBlock()
		// Set the Chain ID
		eb.GetHeader().SetChainID(msg.Entry.GetChainID())
		// Set the Directory Block Height for this Entry Block
		eb.GetHeader().SetDBHeight(dbheight)
		// Add our new entry
		eb.AddEBEntry(msg.Entry)
		// Put it in our list of new Entry Blocks for this Directory Block
		s.PutNewEBlocks(dbheight, msg.Entry.GetChainID(), eb)
		s.PutNewEntries(dbheight, msg.Entry.GetHash(), msg.Entry)

		if v := s.GetReveals(myhash); v != nil {
			s.PutReveals(myhash, nil)
		}

		s.PutCommits(myhash, nil)
		s.IncEntryChains()
		s.IncEntries()
		return true
	} else if _, isNewEntry := commit.(*messages.CommitEntryMsg); isNewEntry {
		chainID := msg.Entry.GetChainID()
		eb := s.GetNewEBlocks(dbheight, chainID)
		if eb == nil {
			prev, err := s.DB.FetchEBlockHead(chainID)
			if prev == nil || err != nil {
				return false
			}
			eb = entryBlock.NewEBlock()
			// Set the Chain ID
			eb.GetHeader().SetChainID(msg.Entry.GetChainID())
			// Set the Directory Block Height for this Entry Block
			eb.GetHeader().SetDBHeight(dbheight)
			// Set the PrevKeyMR
			key, _ := prev.KeyMR()
			eb.GetHeader().SetPrevKeyMR(key)
		}
		// Add our new entry
		eb.AddEBEntry(msg.Entry)
		// Put it in our list of new Entry Blocks for this Directory Block
		s.PutNewEBlocks(dbheight, msg.Entry.GetChainID(), eb)
		s.PutNewEntries(dbheight, msg.Entry.GetHash(), msg.Entry)

		if v := s.GetReveals(myhash); v != nil {
			s.PutReveals(myhash, nil)
		}

		s.PutCommits(myhash, nil)
		s.IncEntries()
		return true
	}
	return false
}

// When we process the directory Signature, and we are the leader for said signature, it
// is then that we push it out to the rest of the network.  Otherwise, if we are not the
// leader for the signature, it marks the sig complete for that list
func (s *State) ProcessDBSig(dbheight uint32, msg interfaces.IMsg) bool {

	dbs := msg.(*messages.DirectoryBlockSignature)

	resp := dbs.Validate(s)
	if resp != 1 {
		return false
	}

	return true
}

func (s *State) GetNewEBlocks(dbheight uint32, hash interfaces.IHash) interfaces.IEntryBlock {
	pl := s.ProcessLists.Get(dbheight)
	return pl.GetNewEBlocks(hash)
}

func (s *State) PutNewEBlocks(dbheight uint32, hash interfaces.IHash, eb interfaces.IEntryBlock) {
	pl := s.ProcessLists.Get(dbheight)
	pl.PutNewEBlocks(dbheight, hash, eb)
}

func (s *State) PutNewEntries(dbheight uint32, hash interfaces.IHash, e interfaces.IEntry) {
	pl := s.ProcessLists.Get(dbheight)
	pl.PutNewEntries(dbheight, hash, e)
}

func (s *State) GetCommits(hash interfaces.IHash) interfaces.IMsg {
	return s.Commits[hash.Fixed()]
}

func (s *State) GetReveals(hash interfaces.IHash) interfaces.IMsg {
	v := s.Reveals[hash.Fixed()]
	return v
}

func (s *State) PutCommits(hash interfaces.IHash, msg interfaces.IMsg) {
	cmsg, ok := msg.(interfaces.ICounted)
	if ok {
		v := s.Commits[hash.Fixed()]
		if v != nil {
			_, ok := v.(interfaces.ICounted)
			if ok {
				cmsg.SetCount(v.(interfaces.ICounted).GetCount() + 1)
			} else {
				panic("Should never happen")
			}
		}
	}
	s.Commits[hash.Fixed()] = msg
}

func (s *State) PutReveals(hash interfaces.IHash, msg interfaces.IMsg) {
	cmsg, ok := msg.(interfaces.ICounted)
	if ok {
		v := s.Reveals[hash.Fixed()]
		if v != nil {
			_, ok := v.(interfaces.ICounted)
			if ok {
				cmsg.SetCount(v.(interfaces.ICounted).GetCount() + 1)
			} else {
				panic("Should never happen")
			}
		}
	}
	s.Reveals[hash.Fixed()] = msg
}

// This is the highest block signed off and recorded in the Database.
func (s *State) GetHighestRecordedBlock() uint32 {
	return s.DBStates.GetHighestRecordedBlock()
}

// If Green, this server is, to the best of its knowledge, caught up with the
// network.  TODO there should be a timeout that requires seeing a message within
// some period of time, but not there yet.
//
// We hare caught up with the network IF:
// The highest recorded block is equal to or just below the highest known block
func (s *State) Green() bool {
	if s.GreenCnt > 1000 {
		return true
	}

	rec := s.DBStates.GetHighestRecordedBlock()
	high := s.GetHighestKnownBlock()
	s.GreenFlg = rec >= high-1
	if s.GreenFlg {
		s.GreenCnt++
	} else {
		s.GreenCnt = 0
	}
	return s.GreenFlg
}

// This is lowest block currently under construction under the "leader".
func (s *State) GetLeaderHeight() uint32 {
	return s.LLeaderHeight
}

// The highest block for which we have received a message.  Sometimes the same as
// BuildingBlock(), but can be different depending or the order messages are recieved.
func (s *State) GetHighestKnownBlock() uint32 {
	if s.ProcessLists == nil {
		return 0
	}
	return s.ProcessLists.DBHeightBase + uint32(len(s.ProcessLists.Lists)-1)
}

func (s *State) GetF(adr [32]byte) int64 {
	s.FactoidBalancesTMutex.Lock()
	defer s.FactoidBalancesTMutex.Unlock()

	if v, ok := s.FactoidBalancesT[adr]; !ok {
		s.FactoidBalancesPMutex.Lock()
		defer s.FactoidBalancesPMutex.Unlock()
		v = s.FactoidBalancesP[adr]
		return v
	} else {
		return v
	}
}

func (s *State) PutF(rt bool, adr [32]byte, v int64) {
	if rt {
		s.FactoidBalancesTMutex.Lock()
		defer s.FactoidBalancesTMutex.Unlock()
		s.FactoidBalancesT[adr] = v
	} else {
		s.FactoidBalancesPMutex.Lock()
		defer s.FactoidBalancesPMutex.Unlock()
		s.FactoidBalancesP[adr] = v
	}
}

func (s *State) GetE(adr [32]byte) int64 {
	s.ECBalancesTMutex.Lock()
	defer s.ECBalancesTMutex.Unlock()

	if v, ok := s.ECBalancesT[adr]; !ok {
		s.ECBalancesPMutex.Lock()
		defer s.ECBalancesPMutex.Unlock()
		v = s.ECBalancesP[adr]
		return v
	} else {
		return v
	}
}

func (s *State) PutE(rt bool, adr [32]byte, v int64) {
	if rt {
		s.ECBalancesTMutex.Lock()
		defer s.ECBalancesTMutex.Unlock()
		s.ECBalancesT[adr] = v
	} else {
		s.ECBalancesPMutex.Lock()
		defer s.ECBalancesPMutex.Unlock()
		s.ECBalancesP[adr] = v
	}
}

// Returns the Virtual Server Index for this hash if this server is the leader;
// returns -1 if we are not the leader for this hash
func (s *State) ComputeVMIndex(hash []byte) int {
	return s.LeaderPL.VMIndexFor(hash)
}

func (s *State) NewAdminBlock(dbheight uint32) interfaces.IAdminBlock {
	ab := new(adminBlock.AdminBlock)
	ab.Header = s.NewAdminBlockHeader(dbheight)
	return ab
}

func (s *State) NewAdminBlockHeader(dbheight uint32) interfaces.IABlockHeader {
	header := new(adminBlock.ABlockHeader)
	header.DBHeight = dbheight
	header.PrevFullHash = primitives.NewHash(constants.ZERO_HASH)
	header.HeaderExpansionSize = 0
	header.HeaderExpansionArea = make([]byte, 0)
	header.MessageCount = 0
	header.BodySize = 0
	return header
}

func (s *State) GetNetworkName() string {
	return (s.Cfg.(util.FactomdConfig)).App.Network

}

func (s *State) GetDBHeightComplete() uint32 {
	db := s.GetDirectoryBlock()
	if db == nil {
		return 0
	}
	return db.GetHeader().GetDBHeight()
}

func (s *State) GetDirectoryBlock() interfaces.IDirectoryBlock {
	if s.DBStates.Last() == nil {
		return nil
	}
	return s.DBStates.Last().DirectoryBlock
}

func (s *State) GetNewHash() interfaces.IHash {
	return new(primitives.Hash)
}

// Create a new Acknowledgement.  Must be called by a leader.  This
// call assumes all the pieces are in place to create a new acknowledgement
func (s *State) NewAck(msg interfaces.IMsg) (iack interfaces.IMsg) {
	s.DBStates.UpdateState()

	vmIndex := msg.GetVMIndex()

	msg.SetLeaderChainID(s.IdentityChainID)
	ack := new(messages.Ack)
	ack.DBHeight = s.LLeaderHeight
	ack.VMIndex = vmIndex
	ack.Minute = byte(s.LeaderMinute)
	ack.Timestamp = s.GetTimestamp()
	ack.MessageHash = msg.GetHash()
	ack.LeaderChainID = s.IdentityChainID

	listlen := len(s.LeaderPL.VMs[vmIndex].List)
	if listlen == 0 {
		ack.Height = 0
		ack.SerialHash = ack.MessageHash
	} else {
		last := s.LeaderPL.GetAckAt(vmIndex, listlen-1)
		ack.Height = last.Height + 1
		ack.SerialHash, _ = primitives.CreateHash(last.MessageHash, ack.MessageHash)
	}

	ack.Sign(s)

	return ack
}

// ****************************************************************
//                          Support
// ****************************************************************

func (s *State) DebugPrt(what string) {

	ppl := s.ProcessLists.Get(s.LLeaderHeight)

	fmt.Printf("tttt %8s %8s:  %v %v  %v %v   %v %v  %v %v  %v %v  %v %v  %v %v  %v %v \n",
		what,
		s.FactomNodeName,
		"LeaderVMIndex", s.LeaderVMIndex,
		"Is Leader", s.Leader,
		"LLeaderHeight", s.LLeaderHeight,
		"Highest rec blk", s.GetHighestRecordedBlock(),
		"Leader Minute", s.LeaderMinute,
		"EOM", s.EOM,
		"PL Min Complete", ppl.MinuteComplete(),
		"PL Min Finish", ppl.MinuteFinished())
	fmt.Printf("tttt\t\t%12s %2v %2v %2v %2v %2v %2v %2v %2v %2v %2v\n",
		"VM Ht:",
		ppl.VMs[0].Height,
		ppl.VMs[1].Height,
		ppl.VMs[2].Height,
		ppl.VMs[3].Height,
		ppl.VMs[4].Height,
		ppl.VMs[5].Height,
		ppl.VMs[6].Height,
		ppl.VMs[7].Height,
		ppl.VMs[8].Height,
		ppl.VMs[9].Height,
	)
	fmt.Printf("tttt\t\t%12s %2v %2v %2v %2v %2v %2v %2v %2v %2v %2v\n",
		"List Len:",
		len(ppl.VMs[0].List),
		len(ppl.VMs[1].List),
		len(ppl.VMs[2].List),
		len(ppl.VMs[3].List),
		len(ppl.VMs[4].List),
		len(ppl.VMs[5].List),
		len(ppl.VMs[6].List),
		len(ppl.VMs[7].List),
		len(ppl.VMs[8].List),
		len(ppl.VMs[9].List),
	)
	fmt.Printf("tttt\t\t%12s %2v %2v %2v %2v %2v %2v %2v %2v %2v %2v\n",
		"Complete:",
		ppl.VMs[0].MinuteComplete,
		ppl.VMs[1].MinuteComplete,
		ppl.VMs[2].MinuteComplete,
		ppl.VMs[3].MinuteComplete,
		ppl.VMs[4].MinuteComplete,
		ppl.VMs[5].MinuteComplete,
		ppl.VMs[6].MinuteComplete,
		ppl.VMs[7].MinuteComplete,
		ppl.VMs[8].MinuteComplete,
		ppl.VMs[9].MinuteComplete,
	)
	fmt.Printf("tttt\t\t%12s %2v %2v %2v %2v %2v %2v %2v %2v %2v %2v\n",
		"Finished:",
		ppl.VMs[0].MinuteFinished,
		ppl.VMs[1].MinuteFinished,
		ppl.VMs[2].MinuteFinished,
		ppl.VMs[3].MinuteFinished,
		ppl.VMs[4].MinuteFinished,
		ppl.VMs[5].MinuteFinished,
		ppl.VMs[6].MinuteFinished,
		ppl.VMs[7].MinuteFinished,
		ppl.VMs[8].MinuteFinished,
		ppl.VMs[9].MinuteFinished,
	)
	fmt.Printf("tttt\t\t%12s %2v %2v %2v %2v %2v %2v %2v %2v %2v %2v\n",
		"Min Ht:",
		ppl.VMs[0].MinuteHeight,
		ppl.VMs[1].MinuteHeight,
		ppl.VMs[2].MinuteHeight,
		ppl.VMs[3].MinuteHeight,
		ppl.VMs[4].MinuteHeight,
		ppl.VMs[5].MinuteHeight,
		ppl.VMs[6].MinuteHeight,
		ppl.VMs[7].MinuteHeight,
		ppl.VMs[8].MinuteHeight,
		ppl.VMs[9].MinuteHeight,
	)
	os.Stdout.Sync()
}
