// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"fmt"

	"hash"

	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/util"
)

var _ = fmt.Print
var _ = (*hash.Hash32)(nil)

//***************************************************************
// Process Loop for Consensus
//
// Returns true if some message was processed.
//***************************************************************
func (s *State) NewMinute() {
	// Anything we are holding, we need to reprocess.
	for k := range s.Holding {
		v := s.Holding[k]
		v.ComputeVMIndex(s)
		s.XReview = append(s.XReview, v)
		delete(s.Holding, k)
	}
}

func (s *State) Process() (progress bool) {

	//fmt.Printf("dddd %20s %10s --- %10s %10v %10s %10v\n", "Process() Start?", s.FactomNodeName, "RunLeader", s.RunLeader, "Leader", s.Leader)
	// Check if we the leader isn't running, and if so, can we start it?
	if !s.RunLeader {
		//fmt.Printf("dddd %20s %10s --- \n", "Process() Start", s.FactomNodeName)
		now := s.GetTimestamp() // Timestamps are in milliseconds, so wait 20
		if now-s.StartDelay > 5*1000 {
			s.RunLeader = true
		}
		s.LeaderPL = s.ProcessLists.Get(s.LLeaderHeight)
		s.Leader, s.LeaderVMIndex = s.LeaderPL.GetVirtualServers(0, s.IdentityChainID)
	}

	dbstate := s.DBStates.Get(int(s.LLeaderHeight - 1))

	//lock := true
	//if dbstate != nil {
	//	lock = dbstate.Locked
	//}
	//fmt.Printf("dddd %20s %10s --- %10s %10v %10s %10v %10s %10v\n", "Process() EOB?", s.FactomNodeName, "LLeaderHt", s.LLeaderHeight, "Saving", s.Saving, "Locked", lock)
	if s.Saving && ((s.LLeaderHeight == 0 && dbstate != nil) || (dbstate != nil && dbstate.Locked)) {

		s.NewMinute()
		s.LeaderPL = s.ProcessLists.Get(s.LLeaderHeight)
		s.Leader, s.LeaderVMIndex = s.LeaderPL.GetVirtualServers(0, s.IdentityChainID)
		//fmt.Printf("dddd %20s %10s --- %10s %10v %10s %10v %10s %10v\n", "NEW BLOCK", s.FactomNodeName, "DBHeight", s.LLeaderHeight, "Leader", s.Leader, "VM", s.LeaderVMIndex)

		s.Saving = false
	}

	if s.RunLeader && s.EOM && s.EOMProcessed >= len(s.LeaderPL.FedServers) {

		//fmt.Printf("dddd %20s %10s --- %10s %10v %10s %10v %10s %10v %10s %10v\n", "NEW MINUTE", s.FactomNodeName, "EOM", s.EOM,
		//	"EomCnt:", s.EOMProcessed, "FedServ#", len(s.ProcessLists.Get(s.LLeaderHeight).FedServers), "Saving", s.Saving)
		// Out of the EOM processing, open all the VMs again.
		for _, vm := range s.LeaderPL.VMs {
			vm.EOM = false
		}

		s.CurrentMinute++
		if s.CurrentMinute > 9 {
			s.CurrentMinute = 0
		}
		switch {

		case s.CurrentMinute > 0:
			//fmt.Printf("dddd %20s %10s --- %10s %10v\n", "Process() lmin  > 0", s.FactomNodeName, "LeaderMin", min)
			s.LeaderPL = s.ProcessLists.Get(s.LLeaderHeight)
			s.Leader, s.LeaderVMIndex = s.LeaderPL.GetVirtualServers(s.CurrentMinute-1, s.IdentityChainID)
			s.NewMinute()
		case s.CurrentMinute == 0:
			//fmt.Printf("dddd %20s %10s --- %10s %10v %10s %10v\n", "Process() lmin == 0", s.FactomNodeName, "LeaderMin", min, "DBHT", s.LeaderPL.DirectoryBlock.GetHeader().GetDBHeight())
			dbstate := s.AddDBState(true, s.LeaderPL.DirectoryBlock, s.LeaderPL.AdminBlock, s.GetFactoidState().GetCurrentBlock(), s.LeaderPL.EntryCreditBlock)
			if s.LLeaderHeight > 0 {
				prev := s.DBStates.Get(int(s.LLeaderHeight))
				if s.DBStates.FixupLinks(prev, dbstate) {
					fmt.Println("dddd Fixed", s.FactomNodeName)
				}
			}
			if s.DBStates.ProcessBlocks(dbstate) {
				fmt.Println("dddd Processed", s.FactomNodeName)
			}

			s.LastHeight = s.LLeaderHeight
			s.LLeaderHeight++
			s.LeaderPL = s.ProcessLists.Get(s.LLeaderHeight)
			s.Leader, s.LeaderVMIndex = s.LeaderPL.GetVirtualServers(0, s.IdentityChainID)
			s.Saving = true
			s.DBSigProcessed = 0

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
					//	fmt.Println("dddd ERROR:", s.FactomNodeName, err.Error())
					panic(err)
				}
				//fmt.Println("dddd DBSig:", s.FactomNodeName, dbs.String())
				dbs.LeaderExecute(s)
			}

			if s.DBStates.SaveDBStateToDB(dbstate) {
				fmt.Println("dddd Saved", s.FactomNodeName)
			}

		}
		s.EOM = false
	}

	return s.ProcessQueues()
}

func (s *State) ProcessQueues() (progress bool) {

	// Executing a message means looking if it is valid, checking if we are a leader.
	executeMsg := func(msg interfaces.IMsg) (ret bool) {
		_, ok := s.Replay.Valid(constants.INTERNAL_REPLAY, msg.GetMsgHash().Fixed(), msg.GetTimestamp(), s.GetTimestamp())
		if !ok {
			return
		}

		msg.ComputeVMIndex(s)

		switch msg.Validate(s) {
		case 1:
			//fmt.Printf("dddd %20s %10s --- %10s %10v %10s %10v %10s %10v %10s %10v\n", "ProcessQ()>", s.FactomNodeName,
			//	"EOM", s.EOM, "Saving", s.Saving, "RunLeader", s.RunLeader, "leader", s.Leader)
			if s.Leader &&
				!s.Saving &&
				!s.LeaderPL.VMs[s.LeaderVMIndex].EOM &&
				(msg.IsLocal() || msg.GetVMIndex() == s.LeaderVMIndex) {

				msg.LeaderExecute(s)
				for i := 0; i < 10 && s.UpdateState(); i++ {
				}

			} else {
				//fmt.Printf("dddd %20s %10s --- %10s %s \n", "xLeader()>", s.FactomNodeName, "Msg", msg.String())
				msg.FollowerExecute(s)
			}
			ret = true
		case 0:
			//fmt.Println("dddd msg holding", s.FactomNodeName, msg.String())
			s.Holding[msg.GetMsgHash().Fixed()] = msg
		default:
			if s.DebugConsensus {
				//fmt.Println("dddd Deleted=== Msg:", s.FactomNodeName, msg.String())
			}
			s.Holding[msg.GetMsgHash().Fixed()] = msg
			s.networkInvalidMsgQueue <- msg
		}

		return
	}

	// Reprocess any stalled Acknowledgements
	for len(s.XReview) > 0 {
		msg := s.XReview[0]
		executeMsg(msg)
		s.XReview = s.XReview[1:]
	}

	select {
	case ack := <-s.ackQueue:
		a := ack.(*messages.Ack)
		if a.DBHeight >= s.LLeaderHeight && ack.Validate(s) == 1 {
			ack.FollowerExecute(s)
		}
		progress = true
	case msg := <-s.msgQueue:
		if executeMsg(msg) {
			s.networkOutMsgQueue <- msg
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
	entryCreditBlock interfaces.IEntryCreditBlock) *DBState {

	dbState := s.DBStates.NewDBState(isNew, directoryBlock, adminBlock, factoidBlock, entryCreditBlock)
	s.DBStates.Put(dbState)
	ht := dbState.DirectoryBlock.GetHeader().GetDBHeight()
	if ht > s.LLeaderHeight {
		s.LLeaderHeight = ht
		s.ProcessLists.Get(ht + 1)
		s.CurrentMinute = 0
	}
	if ht == 0 {
		s.LLeaderHeight = 1
	}

	return dbState
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

	s.Holding[m.GetMsgHash().Fixed()] = m
	ack, _ := s.Acks[m.GetMsgHash().Fixed()].(*messages.Ack)
	if ack != nil {
		m.SetLeaderChainID(ack.GetLeaderChainID())
		m.SetMinute(ack.Minute)

		pl := s.ProcessLists.Get(ack.DBHeight)
		pl.AddToProcessList(ack, m)
	}
}

// Messages that will go into the Process List must match an Acknowledgement.
// The code for this is the same for all such messages, so we put it here.
//
// Returns true if it finds a match, puts the message in holding, or invalidates the message
func (s *State) FollowerExecuteEOM(m interfaces.IMsg) {

	if m.IsLocal() {
		return // This is an internal EOM message.  We are not a leader so ignore.
	}

	eom, _ := m.(*messages.EOM)

	s.Holding[m.GetMsgHash().Fixed()] = m

	ack, _ := s.Acks[m.GetMsgHash().Fixed()].(*messages.Ack)
	if ack != nil {

		// For debugging, note who the leader is for this message, and the minute.
		m.SetLeaderChainID(ack.GetLeaderChainID())
		m.SetMinute(eom.Minute + 1)

		pl := s.ProcessLists.Get(ack.DBHeight)
		pl.AddToProcessList(ack, m)
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
		m.FollowerExecute(s)
	}
}

func (s *State) FollowerExecuteDBState(msg interfaces.IMsg) {

	dbstatemsg, _ := msg.(*messages.DBStateMsg)

	s.DBStates.LastTime = s.GetTimestamp()
	dbstate := s.AddDBState(false, // Not a new block; got it from the network
		dbstatemsg.DirectoryBlock,
		dbstatemsg.AdminBlock,
		dbstatemsg.FactoidBlock,
		dbstatemsg.EntryCreditBlock)
	dbstate.ReadyToSave = true
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

func (s *State) FollowerExecuteSFault(m interfaces.IMsg) {
	sf, _ := m.(*messages.ServerFault)
	pl := s.ProcessLists.Get(sf.DBHeight)
	if pl != nil {
		pl.FaultCnt[sf.ServerID.Fixed()]++
		cnt := pl.FaultCnt[sf.ServerID.Fixed()]
		if s.Leader && cnt > len(pl.FedServers)/2 {

		}
	}
}

func (s *State) FollowerExecuteMMR(m interfaces.IMsg) {
	mmr, _ := m.(*messages.MissingMsgResponse)
	ackResp := mmr.AckResponse.(*messages.Ack)
	//s.Holding[mmr.MsgResponse.GetHash().Fixed()] = mmr.MsgResponse
	//s.Acks[ackResp.GetHash().Fixed()] = ackResp

	pl := s.ProcessLists.Get(ackResp.DBHeight)
	pl.AddToProcessList(ackResp, mmr.MsgResponse)
}

func (s *State) LeaderExecute(m interfaces.IMsg) {

	_, ok := s.Replay.Valid(constants.INTERNAL_REPLAY, m.GetMsgHash().Fixed(), m.GetTimestamp(), s.GetTimestamp())
	if !ok {
		delete(s.Holding, m.GetMsgHash().Fixed())
		return
	}

	if s.Saving {
		m.FollowerExecute(s)
		return
	}

	ack := s.NewAck(m).(*messages.Ack)
	m.SetLeaderChainID(ack.GetLeaderChainID())
	m.SetMinute(ack.Minute)
	s.ProcessLists.Get(ack.DBHeight).AddToProcessList(ack, m)

}

func (s *State) LeaderExecuteEOM(m interfaces.IMsg) {

	if !m.IsLocal() || s.EOM || s.Saving {
		s.FollowerExecuteEOM(m)
		return
	}

	// The zero based minute for the message is equal to
	// the one based "LastMinute".  This way we know we are
	// generating minutes in order.

	eom := m.(*messages.EOM)
	vm := s.ProcessLists.Get(s.LLeaderHeight).VMs[s.LeaderVMIndex]
	if vm.EOM {
		return
	}

	if s.LeaderPL.VMIndexFor(constants.FACTOID_CHAINID) == s.LeaderVMIndex {
		eom.FactoidVM = true
	}
	eom.DBHeight = s.LLeaderHeight
	eom.VMIndex = s.LeaderVMIndex
	// eom.Minute is zerobased, while LeaderMinute is 1 based.  So
	// a simple assignment works.
	eom.Minute = byte(s.CurrentMinute)
	eom.Sign(s)
	ack := s.NewAck(m)
	s.Acks[eom.GetMsgHash().Fixed()] = ack
	m.SetLocal(false)

	s.FollowerExecuteEOM(m)

}

func (s *State) LeaderExecuteRevealEntry(m interfaces.IMsg) {
	re := m.(*messages.RevealEntryMsg)
	commit := s.NextCommit(re.Entry.GetHash())
	if commit == nil {
		m.FollowerExecute(s)
	}
	s.LeaderExecute(m)
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
	} else if as.ServerType == 1 {
		s.LeaderPL.AdminBlock.AddAuditServer(as.ServerChainID)
	}

	return true
}

func (s *State) ProcessChangeServerKey(dbheight uint32, changeServerKeyMsg interfaces.IMsg) bool {
	ask, ok := changeServerKeyMsg.(*messages.ChangeServerKeyMsg)
	if !ok {
		return true
	}

	// TODO: Signiture && Checking

	//fmt.Printf("DEBUG: Processed: %x", ask.AdminBlockChange)
	switch ask.AdminBlockChange {
	case constants.TYPE_ADD_BTC_ANCHOR_KEY:
		var btcKey [20]byte
		copy(btcKey[:], ask.Key.Bytes()[:20])
		fmt.Println("Add BTC to admin block")
		s.LeaderPL.AdminBlock.AddFederatedServerBitcoinAnchorKey(ask.IdentityChainID, ask.KeyPriority, ask.KeyType, &btcKey)
	case constants.TYPE_ADD_FED_SERVER_KEY:
		pub := ask.Key.Fixed()
		fmt.Println("Add Block Key to admin block : " + s.IdentityChainID.String())
		s.LeaderPL.AdminBlock.AddFederatedServerSigningKey(ask.IdentityChainID, &pub)
	case constants.TYPE_ADD_MATRYOSHKA:
		fmt.Println("Add MHash to admin block")
		s.LeaderPL.AdminBlock.AddMatryoshkaHash(ask.IdentityChainID, ask.Key)
	}
	return true
}

func (s *State) ProcessCommitChain(dbheight uint32, commitChain interfaces.IMsg) bool {
	c, _ := commitChain.(*messages.CommitChainMsg)

	pl := s.ProcessLists.Get(dbheight)
	pl.EntryCreditBlock.GetBody().AddEntry(c.CommitChain)
	s.GetFactoidState().UpdateECTransaction(true, c.CommitChain)

	// save the Commit to match agains the Reveal later
	s.PutCommit(c.CommitChain.EntryHash, c)

	return true
}

func (s *State) ProcessCommitEntry(dbheight uint32, commitEntry interfaces.IMsg) bool {
	c, _ := commitEntry.(*messages.CommitEntryMsg)

	pl := s.ProcessLists.Get(dbheight)
	pl.EntryCreditBlock.GetBody().AddEntry(c.CommitEntry)
	s.GetFactoidState().UpdateECTransaction(true, c.CommitEntry)

	// save the Commit to match agains the Reveal later
	s.PutCommit(c.CommitEntry.EntryHash, c)

	return true
}

func (s *State) ProcessRevealEntry(dbheight uint32, m interfaces.IMsg) bool {

	msg := m.(*messages.RevealEntryMsg)
	myhash := msg.Entry.GetHash()

	if m.Validate(s) != 1 {
		commit := s.NextCommit(myhash)
		if commit == nil {
			return false
		}
		return s.ProcessRevealEntry(dbheight, m)
	}

	chainID := msg.Entry.GetChainID()

	s.NextCommit(myhash)

	eb, _ := s.DB.FetchEBlockHead(chainID)

	if !msg.IsEntry && eb == nil {
		eb, _ := s.DB.FetchEBlockHead(chainID)
		if eb == nil {
			// Create a new Entry Block for a new Entry Block Chain
			eb = entryBlock.NewEBlock()
			// Set the Chain ID
			eb.GetHeader().SetChainID(chainID)
			// Set the Directory Block Height for this Entry Block
			eb.GetHeader().SetDBHeight(dbheight)
			// Add our new entry
			eb.AddEBEntry(msg.Entry)
			// Put it in our list of new Entry Blocks for this Directory Block
			s.PutNewEBlocks(dbheight, chainID, eb)
			s.PutNewEntries(dbheight, myhash, msg.Entry)

			s.IncEntryChains()
			s.IncEntries()
			return true
		}
	}

	ebc := s.GetNewEBlocks(dbheight, chainID)
	if ebc == nil {
		prev := s.GetNewEBlocks(dbheight-1, chainID)
		prevdb := eb
		if prev == nil && prevdb == nil {
			return false
		}

		if prev != nil {
			mr1, _ := prev.KeyMR()
			mr2, _ := prevdb.KeyMR()
			if mr2 == nil || !mr1.IsSameAs(mr2) {
				fmt.Printf("dddd BAD EB Block Head %s DBHeigth: %3d Min %2d PL: %x DB: %x\n",
					s.FactomNodeName, s.LLeaderHeight, int(s.LeaderPL.VMs[0].LeaderMinute), mr1.Bytes(), mr2.Bytes())
			}
		} else {
			prev = prevdb
		}

		eb = entryBlock.NewEBlock()
		eb.GetHeader().SetEBSequence(prev.GetHeader().GetEBSequence() + 1)
		eb.GetHeader().SetPrevFullHash(prev.GetHash())
		// Set the Chain ID
		eb.GetHeader().SetChainID(chainID)
		// Set the Directory Block Height for this Entry Block
		eb.GetHeader().SetDBHeight(dbheight)
		// Set the PrevKeyMR
		key, _ := prev.KeyMR()
		eb.GetHeader().SetPrevKeyMR(key)
	}
	// Add our new entry
	eb.AddEBEntry(msg.Entry)
	// Put it in our list of new Entry Blocks for this Directory Block
	s.PutNewEBlocks(dbheight, chainID, eb)
	s.PutNewEntries(dbheight, myhash, msg.Entry)

	s.IncEntries()
	return true
}

// TODO: Should fault the server if we don't have the proper sequence of EOM messages.
func (s *State) ProcessEOM(dbheight uint32, msg interfaces.IMsg) bool {

	e := msg.(*messages.EOM)

	if s.EOMDone && e.Processed {
		return true
	}

	pl := s.ProcessLists.Get(dbheight)

	vm := s.ProcessLists.Get(dbheight).VMs[msg.GetVMIndex()]
	vm.LeaderMinute++
	vm.EOM = true

	if !s.EOM {
		s.EOM = true
		s.EOMProcessed = 0

	}

	s.EOMProcessed++

	// After all EOM markers are processed, but before anything else is done
	// we do any cleanup required.
	if s.EOMProcessed == len(s.LeaderPL.FedServers) {

		s.EOMDone = true

		s.FactoidState.EndOfPeriod(int(e.Minute))

		// Add EOM to the EBlocks.  We only do this once, so
		// we piggy back on the fact that we only do the FactoidState
		// EndOfPeriod once too.

		for _, eb := range pl.NewEBlocks {
			eb.AddEndOfMinuteMarker(byte(e.Minute + 1))
		}

		ecblk := pl.EntryCreditBlock
		ecbody := ecblk.GetBody()
		mn := entryCreditBlock.NewMinuteNumber(e.Minute + 1)
		ecbody.AddEntry(mn)
	}

	e.Processed = true
	return false
}

// When we process the directory Signature, and we are the leader for said signature, it
// is then that we push it out to the rest of the network.  Otherwise, if we are not the
// leader for the signature, it marks the sig complete for that list
func (s *State) ProcessDBSig(dbheight uint32, msg interfaces.IMsg) bool {

	dbs := msg.(*messages.DirectoryBlockSignature)

	//fmt.Printf("dddd %20s %10s --- %10s %10v \n", "ProcessDBSig()", s.FactomNodeName, "DBHeight", dbheight)

	resp := dbs.Validate(s)
	if resp != 1 {
		//fmt.Printf("dddd %20s %10s --- %10s %10v \n", "ProcessDBSig()-", s.FactomNodeName, "DBHeight", dbheight)
		return false
	}

	if dbs.VMIndex == 0 {
		s.SetLeaderTimestamp(dbs.GetTimestamp())
	}

	if !dbs.Once {
		s.DBSigProcessed++
		dbs.Once = true
	}

	if s.DBSigProcessed >= len(s.LeaderPL.FedServers) {
		// TODO: check signatures here.  Count what match and what don't.  Then if a majority
		// disagree with us, null our entry out.  Otherwise toss our DBState and ask for one from
		// our neighbors.
		dbstate := s.DBStates.Get(int(dbheight - 1))
		if dbstate.Saved {
			return true
		} else {
			dbstate.ReadyToSave = true
		}
	}
	return false
}

func (s *State) ConsiderSaved(dbheight uint32) {
	for _, dbs := range s.DBStates.DBStates {
		if dbs.DirectoryBlock.GetDatabaseHeight() == dbheight {
			dbs.Saved = true
		}
	}
}

func (s *State) GetNewEBlocks(dbheight uint32, hash interfaces.IHash) interfaces.IEntryBlock {
	pl := s.ProcessLists.Get(dbheight)
	if pl == nil {
		return nil
	}
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

// Returns the oldest, not processed, Commit received
func (s *State) NextCommit(hash interfaces.IHash) interfaces.IMsg {
	cs := s.Commits[hash.Fixed()]
	if cs == nil || len(cs) == 0 {
		return nil
	}
	r := cs[0]
	s.Commits[hash.Fixed()] = cs[1:]
	return r
}

func (s *State) PutCommit(hash interfaces.IHash, msg interfaces.IMsg) {
	cs := s.Commits[hash.Fixed()]
	if cs == nil {
		cs = make([]interfaces.IMsg, 0)
	}
	s.Commits[hash.Fixed()] = append(cs, msg)
}

// This is the highest block signed off and recorded in the Database.
func (s *State) GetHighestRecordedBlock() uint32 {
	return s.DBStates.GetHighestRecordedBlock()
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
	plh := s.ProcessLists.DBHeightBase + uint32(len(s.ProcessLists.Lists)-1)
	dbsh := s.DBStates.Base + uint32(len(s.DBStates.DBStates))
	if dbsh > plh {
		return dbsh
	}
	return plh
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

	vmIndex := msg.GetVMIndex()

	msg.SetLeaderChainID(s.IdentityChainID)
	ack := new(messages.Ack)
	ack.DBHeight = s.LLeaderHeight
	ack.VMIndex = vmIndex
	ack.Minute = byte(s.ProcessLists.Get(s.LLeaderHeight).VMs[vmIndex].LeaderMinute)
	ack.Timestamp = s.GetTimestamp()
	ack.MessageHash = msg.GetMsgHash()
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
