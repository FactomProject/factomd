// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"fmt"

	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/util"
)

var _ = fmt.Print

//***************************************************************
// Process Loop for Consensus
//
// Returns true if some message was processed.
//***************************************************************
func (s *State) Process() (progress bool) {

	if false {
		ppl := s.ProcessLists.Get(s.LLeaderHeight)
		fmt.Println(
			s.FactomNodeName,
			"  DBHeight", s.LLeaderHeight,
			"  Finished EOM:", ppl.FinishedEOM(),
			"  EOM", s.EOM,
			"  Leader min:", s.LeaderMinute,
			"  PL Min Ht:", ppl.MinuteHeight(),
			"  PL Ht:", ppl.VMs[0].Height)
	}

	highest := s.GetHighestRecordedBlock()

	if s.LLeaderHeight == 0 {
		s.LLeaderHeight = s.GetHighestRecordedBlock() + 1
		s.LeaderPL = s.ProcessLists.Get(s.LLeaderHeight)
		s.Leader, s.LeaderVMIndex = s.LeaderPL.GetVirtualServers(0, s.IdentityChainID)
	} else if s.LLeaderHeight <= highest {

		s.LeaderMinute = 0 // Last block leaves at 10, which blows up. New block = 0

		for _, vm := range s.LeaderPL.VMs {
			ack1, ok1 := vm.LastLeaderAck.(*messages.Ack)
			ack2, ok2 := vm.LastAck.(*messages.Ack)
			if (!ok1 && ok2) || (ok1 && ok2 && ack2.Height > ack1.Height) {
				vm.LastLeaderAck = vm.LastAck
			}
		}
		s.LLeaderHeight = s.GetHighestRecordedBlock() + 1
		s.LeaderPL = s.ProcessLists.Get(s.LLeaderHeight)
		s.Leader, s.LeaderVMIndex = s.LeaderPL.GetVirtualServers(0, s.IdentityChainID)
		if s.Leader {
			s.EOM = false

			dbstate := s.DBStates.Get(s.LLeaderHeight - 1)

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
			s.leaderMsgQueue <- dbs
		}
	}

	if s.EOM && s.LeaderPL.FinishedEOM() {
		switch {
		case s.LeaderMinute <= 9:
			for _, vm := range s.LeaderPL.VMs {
				ack1, ok1 := vm.LastLeaderAck.(*messages.Ack)
				ack2, ok2 := vm.LastAck.(*messages.Ack)
				if (!ok1 && ok2) || (ok1 && ok2 && ack2.Height >= ack1.Height) {
					vm.LastLeaderAck = vm.LastAck
				}
			}
			s.EOM = false
		case s.LeaderMinute == 10:
			s.AddDBState(true, s.LeaderPL.DirectoryBlock, s.LeaderPL.AdminBlock, s.GetFactoidState().GetCurrentBlock(), s.LeaderPL.EntryCreditBlock)
		}
	}

	if !s.Leader || (!s.EOM && s.Leader) {
		var vm *VM
		if s.Leader {
			vm = s.LeaderPL.VMs[s.LeaderVMIndex]
		}
		// To process a leader message, we have to have the follower process completely
		// up to date.  Then we can validate the message.  Process is up to date if all
		// messages in the process list have been processed by the follower, ie the Height
		// is equal to the length of the process list.
		if !s.Leader || len(vm.List) >= vm.Height {
			select {
			case msg, _ := <-s.leaderMsgQueue:
				v := msg.Validate(s)
				switch v {
				case 1:
					msg.LeaderExecute(s)
					s.networkOutMsgQueue<-msg
					for s.UpdateState() {
					}
				case -1:
					s.networkInvalidMsgQueue <- msg
				}
				progress = true
			default:
			}
		}
	}
	// Followers are less strict.  Messages can be validated as they are processed, but
	// the acknowledgement from the leader is enough to put a message into the process list.
	select {
	case msg := <-s.followerMsgQueue:
		v := msg.Validate(s)
		switch v {
		case 1:
			msg.FollowerExecute(s)
			s.networkOutMsgQueue<-msg
			for s.UpdateState() {
			}
		case -1:
			s.networkInvalidMsgQueue <- msg
		}
		progress = true
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
			s.DB.ProcessEBlockBatch(eblock)
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
// Returns true if it finds a match
func (s *State) FollowerExecuteMsg(m interfaces.IMsg) (bool, error) {

	hash := m.GetHash()
	hashf := hash.Fixed()
	ack, ok := s.Acks[hashf].(*messages.Ack)
	if !ok || ack == nil {
		s.Holding[hashf] = m
		return false, nil
	} else {
		pl := s.ProcessLists.Get(ack.DBHeight)

		if m.Type() == constants.COMMIT_CHAIN_MSG || m.Type() == constants.COMMIT_ENTRY_MSG {
			s.PutCommits(hash, m)
		}

		if pl != nil {
			pl.AddToProcessList(ack, m)

			pl.OldAcks[hashf] = ack
			pl.OldMsgs[hashf] = m
			delete(s.Acks, hashf)
			delete(s.Holding, hashf)
		}
		return true, nil
	}
}

// Ack messages always match some message in the Process List.   That is
// done here, though the only msg that should call this routine is the Ack
// message.
func (s *State) FollowerExecuteAck(msg interfaces.IMsg) (bool, error) {
	ack := msg.(*messages.Ack)
	s.Acks[ack.GetHash().Fixed()] = ack
	match := s.Holding[ack.GetHash().Fixed()]
	if match != nil {
		match.FollowerExecute(s)
		return true, nil
	}

	return false, nil
}

func (s *State) FollowerExecuteDBState(msg interfaces.IMsg) error {

	dbstatemsg, ok := msg.(*messages.DBStateMsg)
	if !ok {
		return fmt.Errorf("Cannot execute the given DBStateMsg")
	}

	s.DBStates.LastTime = s.GetTimestamp()
	//	fmt.Println("DBState Message  ")
	s.AddDBState(true,
		dbstatemsg.DirectoryBlock,
		dbstatemsg.AdminBlock,
		dbstatemsg.FactoidBlock,
		dbstatemsg.EntryCreditBlock)

	return nil
}

func (s *State) FollowerExecuteAddData(msg interfaces.IMsg) error {
	dataResponseMsg, ok := msg.(*messages.DataResponse)
	if !ok {
		return fmt.Errorf("Cannot execute the given DataResponse")
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
		return fmt.Errorf("Datatype currently unsupported")
	}

	return nil
}

func (s *State) LeaderExecute(m interfaces.IMsg) error {
	h := m.GetVMHash()
	if h != nil && len(h)>0 {
		fmt.Println(s.FactomNodeName,"Leader Execute VMHash",m.GetVMIndex())
		m.SetVMIndex(s.LeaderPL.VMIndexFor(m.GetVMHash()))
	}
	fmt.Println(s.FactomNodeName,"Leader",s.Leader,"MsgVMIndex",m.GetVMIndex(),"LeaderVM",s.LeaderVMIndex)
	if !s.Leader || m.GetVMIndex() != s.LeaderVMIndex {
		if m.Follower(s) {
			m.FollowerExecute(s)
		}
		return nil
	}
	dbheight := s.LLeaderHeight
	ack, err := s.NewAck(dbheight, m)
	if err != nil {
		return err
	}

	s.followerMsgQueue <- ack
	s.followerMsgQueue <- m
	return nil
}

func (s *State) LeaderExecuteEOM(m interfaces.IMsg) error {
	if !s.Leader { // Ignore local EOM messages when a follower only.
		return nil
	}

	eom := m.(*messages.EOM)
	eom.DBHeight = s.LLeaderHeight
	eom.VMIndex = s.LeaderVMIndex
	eom.Sign(s)
	eom.SetLocal(false)
	ack, err := s.NewAck(s.LLeaderHeight, m)

	if err != nil {
		return err
	}

	s.followerMsgQueue <- m
	s.followerMsgQueue <- ack

	return nil
}

func (s *State) ProcessAddServer(dbheight uint32, addServerMsg interfaces.IMsg) bool {
	as, ok := addServerMsg.(*messages.AddServerMsg)
	if !ok {
		return true
	}

	pl := s.ProcessLists.Get(dbheight)
	if as.ServerType == 0 {
		pl.AdminBlock.AddFedServer(as.ServerChainID)
	}

	return true
}

func (s *State) ProcessCommitChain(dbheight uint32, commitChain interfaces.IMsg) bool {
	c, ok := commitChain.(*messages.CommitChainMsg)
	if !ok {
		return false
	}

	pl := s.ProcessLists.Get(dbheight)
	pl.EntryCreditBlock.GetBody().AddEntry(c.CommitChain)
	s.GetFactoidState().UpdateECTransaction(true, c.CommitChain)

	// save the Commit to match agains the Reveal later
	s.PutCommits(c.CommitChain.EntryHash, c)
	// check for a matching Reveal and, if found, execute it
	if r := s.GetReveals(c.CommitChain.EntryHash); r != nil {
		s.LeaderExecute(r)
	}

	return true
}

func (s *State) ProcessCommitEntry(dbheight uint32, commitEntry interfaces.IMsg) bool {
	c, ok := commitEntry.(*messages.CommitEntryMsg)
	if !ok {
		return false
	}

	pl := s.ProcessLists.Get(dbheight)
	pl.EntryCreditBlock.GetBody().AddEntry(c.CommitEntry)
	s.GetFactoidState().UpdateECTransaction(true, c.CommitEntry)

	// save the Commit to match agains the Reveal later
	s.PutCommits(c.CommitEntry.EntryHash, c)
	// check for a matching Reveal and, if found, execute it
	if r := s.GetReveals(c.CommitEntry.EntryHash); r != nil {
		s.LeaderExecute(r)
	}

	return true
}

// TODO: Should fault the server if we don't have the proper sequence of EOM messages.
func (s *State) ProcessEOM(dbheight uint32, msg interfaces.IMsg) bool {

	e, ok := msg.(*messages.EOM)
	if !ok {
		panic("Must pass an EOM message to ProcessEOM)")
	}

	pl := s.ProcessLists.Get(dbheight)


	// Set this list complete
	s.LeaderMinute = int(e.Minute + 1)
	pl.SetMinute(e.VMIndex, int(e.Minute))

	if pl.MinuteHeight() < s.LeaderMinute {
		return false
	}

	s.EOM = true

	if pl.VMIndexFor(constants.FACTOID_CHAINID) == e.VMIndex {
		s.FactoidState.EndOfPeriod(int(e.Minute))

		// Add EOM to the EBlocks.  We only do this once, so
		// we piggy back on the fact that we only do the FactoidState
		// EndOfPeriod once too.
		for _, eb := range pl.NewEBlocks {
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

	return true
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
	return s.Reveals[hash.Fixed()]
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
	if v, ok := s.FactoidBalancesT[adr]; !ok {
		v = s.FactoidBalancesP[adr]
		return v
	} else {
		return v
	}
}

func (s *State) PutF(rt bool, adr [32]byte, v int64) {
	if rt {
		s.FactoidBalancesT[adr] = v
	} else {
		s.FactoidBalancesP[adr] = v
	}
}

func (s *State) GetE(adr [32]byte) int64 {
	if v, ok := s.ECBalancesT[adr]; !ok {
		v = s.ECBalancesP[adr]
		return v
	} else {
		return v
	}
}

func (s *State) PutE(rt bool, adr [32]byte, v int64) {
	if rt {
		s.ECBalancesT[adr] = v
	} else {
		s.ECBalancesP[adr] = v
	}
}

// Returns the Virtual Server Index for this hash if this server is the leader;
// returns -1 if we are not the leader for this hash
func (s *State) LeaderFor(msg interfaces.IMsg, hash []byte) bool {
	if hash != nil {
		h := make([]byte, len(hash))
		copy(h, hash)
fmt.Println("LeaderFor ... ",msg.String())
		msg.SetVMHash(h) // <-- This is important
	}
	return true
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

func (s *State) GetDB() interfaces.DBOverlay {
	return s.DB
}

func (s *State) SetDB(dbase interfaces.DBOverlay) {
	s.DB = databaseOverlay.NewOverlay(dbase)
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

// Create a new Acknowledgement.  This Acknowledgement
func (s *State) NewAck(dbheight uint32, msg interfaces.IMsg) (iack interfaces.IMsg, err error) {

	vmIndex := msg.GetVMIndex()
fmt.Println(s.FactomNodeName,"NewAck",vmIndex,)
	pl := s.ProcessLists.Get(dbheight)
	if pl == nil {
		return nil, fmt.Errorf(s.FactomNodeName + ": No process list at this time")
	}
	msg.SetLeaderChainID(s.IdentityChainID)
	ack := new(messages.Ack)
	ack.DBHeight = dbheight
	ack.VMIndex = vmIndex
	ack.Timestamp = s.GetTimestamp()
	ack.MessageHash = msg.GetHash()
	ack.LeaderChainID = s.IdentityChainID

	last, ok := pl.GetLastLeaderAck(vmIndex).(*messages.Ack)
	if !ok {
		ack.Height = 0
		ack.SerialHash = ack.MessageHash
	} else {
		ack.Height = last.Height + 1
		ack.SerialHash, err = primitives.CreateHash(last.MessageHash, ack.MessageHash)
		if err != nil {
			return nil, err
		}
	}
	pl.SetLastLeaderAck(vmIndex, ack)

	ack.Sign(s)

	return ack, nil
}
