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
	"os"
)

var _ = fmt.Print

//***************************************************************
// Process Loop for Consensus
//
// Returns true if some message was processed.
//***************************************************************
func (s *State) NewMinute() {
	s.Review = make([]interfaces.IMsg, 0, len(s.Holding))
	// Anything we are holding, we need to reprocess.
	for k := range s.Holding {
		if v:= s.Holding[k]; v != nil {
			s.Review = append(s.Review,v)
			s.Holding[k] = nil
		}
	}
	// Clear the holding map
	s.Holding = make(map[[32]byte]interfaces.IMsg)
	s.Reveals = make(map[[32]byte]interfaces.IMsg)
	s.EOM=0
}

func (s *State) Process() (progress bool) {

	//s.DebugPrt("Process")

	highest := s.GetHighestRecordedBlock()

	if s.EOM <= 9 {
		s.LeaderPL = s.ProcessLists.Get(s.LLeaderHeight)
		s.Leader, s.LeaderVMIndex = s.LeaderPL.GetVirtualServers(s.LeaderMinute, s.IdentityChainID)
		minFin := s.LeaderPL.MinuteFinished()
		if s.EOM < minFin {
			s.LeaderMinute = minFin
		}
	}

	dbstate := s.DBStates.Get(s.LLeaderHeight)
	if s.LLeaderHeight <= highest && (dbstate == nil || dbstate.Saved){
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
			s.leaderMsgQueue <- dbs
		}
		s.LeaderMinute = 0
		s.NewMinute()
	}

	if s.EOM > 0 && s.LeaderPL.Unseal(s.EOM) {
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

func (s *State) TryToProcess(msg interfaces.IMsg) {
	// First make sure the message is valid.

//fmt.Println("xxxxxxxxxx", s.FactomNodeName, msg.String())

	ExeFollow := func() {
		if msg.Follower(s) {
			if !s.Leader && msg.IsLocal() {
				if _, ok := msg.(*messages.EOM); ok {
					return
				}
			}
			err := msg.FollowerExecute(s)
			if err == nil && !msg.IsRepeat() {
				s.networkOutMsgQueue <- msg
			} else {
				s.StallMsg(msg)
			}
		}
	}


	msgLeader := msg.Leader(s)
	if ack, ok := msg.(*messages.Ack);
	      s.LeaderPL.GoodTo(msg.GetVMIndex()) &&
	      (!ok || int(ack.Height) == s.LeaderPL.VMs[ack.VMIndex].Height) {
		v := msg.Validate(s)
		if v == 1 {
			// If we are a leader, we are way more strict than simple followers.
			if msgLeader && s.Leader && s.EOM == 0  &&
			(s.LeaderVMIndex == msg.GetVMIndex() || msg.IsLocal()) {
				err := msg.LeaderExecute(s)
				if err == nil {
					// If all went well, then send it to the world.
					if !msg.IsRepeat() {
						s.networkOutMsgQueue <- msg
					}
				} else {
					// If bad, stall as long as it isn't our own EOM
					if _, ok := msg.(*messages.EOM); !ok {
						s.StallMsg(msg)
					}
				}
			} else {
				ExeFollow()
			}
		} else if v == 0 {
			// Could be good, might not be.  Stall it.
			s.StallMsg(msg)
			// If the transaction isn't valid (or we can't tell) we just drop it.
		} else {
			s.networkInvalidMsgQueue <- msg
		}
	}else{
		s.StallMsg(msg)
	}
}

func (s *State) ProcessQueues() (progress bool) {

	var msg interfaces.IMsg

	for msg == nil && s.Review != nil && len(s.Review) > 0 {
		msg = s.Review[0]
		s.Review = s.Review[1:]
		progress = true
		if msg != nil {
			msg.SetRepeat(true)
		}
	}

	if msg == nil {
		select {
		case msg = <-s.leaderMsgQueue:
			_, ok := s.InternalReplay.Valid(msg.GetHash().Fixed(), int64(msg.GetTimestamp()), int64(s.GetTimestamp()))
			if !ok {
				msg = nil
			}
			progress = true
		default:
		}
	}
	// If all my messages are empy, see if I can process a stalled message
	if msg == nil {
		select {
		case msg = <-s.stallQueue:
			_, ok := s.InternalReplay.Valid(msg.GetHash().Fixed(), int64(msg.GetTimestamp()), int64(s.GetTimestamp()))
			if !ok {
				msg = nil
			} else {
				msg.SetStalled(true) // Allow them to rebroadcast if they work.
			}

		case msg = <-s.followerMsgQueue:
			_, ok := s.InternalReplay.Valid(msg.GetHash().Fixed(), int64(msg.GetTimestamp()), int64(s.GetTimestamp()))
			if !ok {
				msg = nil
			} else {
			}
			progress = true
		default:
		}
	}

	if msg != nil {
		if s.LeaderPL != nil {
			if !s.NetStateOff {
				s.TryToProcess(msg)
			}else{
				fmt.Println(s.FactomNodeName, "Msg: ", msg.String())
			}
		} else {
			if !s.NetStateOff {
				s.TryToProcess(msg)
			}else{
				fmt.Println(s.FactomNodeName, "Msg: ", msg.String())
			}
		}
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
		s.EOM=0
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
// Returns true if it finds a match
func (s *State) FollowerExecuteMsg(m interfaces.IMsg) (bool, error) {
	hash := m.GetHash()
	hashf := hash.Fixed()
	ack, ok := s.Acks[hashf].(*messages.Ack)
	if !ok || ack == nil {
		s.Holding[hashf] = m
		return false, nil
	} else {
		delete(s.Acks, hashf)    // No matter what, we don't want to
		delete(s.Holding, hashf) // rematch.. If we stall, we will see them again.

		pl := s.ProcessLists.Get(ack.DBHeight)

		if pl != nil {
			if !pl.AddToProcessList(ack, m) {
				return false, fmt.Errorf("Could not add message")
			}

			pl.OldAcks[hashf] = ack
			pl.OldMsgs[hashf] = m
		} else {
			if ack.DBHeight >= s.ProcessLists.DBHeightBase {
				return false, fmt.Errorf("Could not add message")
			}
		}

		return true, nil
	}
}

// Ack messages always match some message in the Process List.   That is
// done here, though the only msg that should call this routine is the Ack
// message.
func (s *State) FollowerExecuteAck(msg interfaces.IMsg) (bool, error) {
	ack := msg.(*messages.Ack)
	match := s.Holding[ack.GetHash().Fixed()]
	if match != nil {
		s.Acks[ack.GetHash().Fixed()] = ack
		if err := match.FollowerExecute(s); err != nil {
			s.Acks[ack.GetHash().Fixed()] = nil
			return false, err
		}
		return true, nil
	}
	s.Acks[ack.GetHash().Fixed()] = ack
	return false, nil
}

func (s *State) FollowerExecuteDBState(msg interfaces.IMsg) error {

	dbstatemsg, _ := msg.(*messages.DBStateMsg)

	s.DBStates.LastTime = s.GetTimestamp()
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

	if s.EOM > 0 {
		return fmt.Errorf("Cannot Lead right now")
	}

	dbheight := s.LLeaderHeight
	ack, err := s.NewAck(dbheight, m)
	if err != nil {
		return err
	}

	if err := ack.FollowerExecute(s); err == nil {
		m.FollowerExecute(s);
		s.networkOutMsgQueue <- ack
	}else{
		return err
	}

	return nil
}

// Leader Execute for Reveal Entry
func (s *State) LeaderExecuteRE(m interfaces.IMsg) error {

	if s.EOM > 0 {
		return fmt.Errorf("Cannot Lead right now")
	}

	dbheight := s.LLeaderHeight
	ack, err := s.NewAck(dbheight, m)
	if err != nil {
		return err
	}

	if err := ack.FollowerExecute(s); err != nil {
		return err
	}
	m.FollowerExecute(s);
	s.networkOutMsgQueue <- ack

	return nil
}


func (s *State) LeaderExecuteEOM(m interfaces.IMsg) error {

	eom := m.(*messages.EOM)

	if s.EOM > 0 {
		return fmt.Errorf("Stalling")
	}

	s.EOM = int(s.LeaderMinute + 1)
	if s.LeaderPL.VMIndexFor(constants.FACTOID_CHAINID) == s.LeaderVMIndex {
		eom.FactoidVM = true
	}
	eom.DBHeight = s.LLeaderHeight
	eom.VMIndex = s.LeaderVMIndex
	eom.Minute = byte(s.LeaderMinute)
	eom.Sign(s)
	eom.SetLocal(false)
	ack, err := s.NewAck(s.LLeaderHeight, m)
	if err != nil {
		return err
	}

	if err := ack.FollowerExecute(s); err == nil {
		m.FollowerExecute(s);
		s.networkOutMsgQueue <- ack
	}else{
		return err
	}

	return nil
}

func (s *State) ProcessAddServer(dbheight uint32, addServerMsg interfaces.IMsg) bool {
	as, ok := addServerMsg.(*messages.AddServerMsg)
	if !ok {
		return true
	}

	pl := s.ProcessLists.Get(dbheight)
	if leader, _ := pl.GetFedServerIndexHash(as.ServerChainID); leader {
		return true
	}

	if as.ServerType == 0 {
		pl.AdminBlock.AddFedServer(as.ServerChainID)
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

	pl.SetMinute(e.VMIndex, int(e.Minute))

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
		if pl.VMIndexFor(eb.GetChainID().Bytes())==e.VMIndex {
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
		msg.SetVMHash(h)
		msg.SetVMIndex(s.LeaderPL.VMIndexFor(h))
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
	pl := s.ProcessLists.Get(dbheight)

	if pl == nil {
		if s.DebugConsensus {
			fmt.Printf("%-30s %10s %s\n", "aaa Ack PL==nil", s.FactomNodeName, msg.String())
		}
		err = fmt.Errorf(s.FactomNodeName + ": No process list at this time")
		return
	}
	msg.SetLeaderChainID(s.IdentityChainID)
	ack := new(messages.Ack)
	ack.DBHeight = dbheight
	ack.VMIndex = vmIndex
	ack.Minute = byte(s.LeaderMinute)
	ack.Timestamp = s.GetTimestamp()
	ack.MessageHash = msg.GetHash()
	ack.LeaderChainID = s.IdentityChainID

	last := pl.GetAckAt(vmIndex,pl.VMs[vmIndex].Height-1)
	if last == nil {
		ack.Height = 0
		ack.SerialHash = ack.MessageHash
	} else {
		ack.Height = last.Height + 1
		ack.SerialHash, err = primitives.CreateHash(last.MessageHash, ack.MessageHash)
		if err != nil {
			if s.DebugConsensus {
				fmt.Printf("%-30s %10s %s\n", "aaa Ack Serial Hash Failed", s.FactomNodeName, msg.String())
			}
			return nil, err
		}
	}

	ack.Sign(s)

	return ack, nil
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
