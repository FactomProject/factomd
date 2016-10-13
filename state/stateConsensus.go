// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"fmt"
	"hash"

	"time"

	"errors"

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

func (s *State) executeMsg(vm *VM, msg interfaces.IMsg) (ret bool) {
	_, ok := s.Replay.Valid(constants.INTERNAL_REPLAY, msg.GetRepeatHash().Fixed(), msg.GetTimestamp(), s.GetTimestamp())
	if !ok {
		return
	}
	s.SetString()
	msg.ComputeVMIndex(s)

	switch msg.Validate(s) {
	case 1:
		if s.RunLeader &&
			s.Leader &&
			!s.Saving &&
			vm != nil && int(vm.Height) == len(vm.List) &&
			(!s.Syncing || !vm.Synced) &&
			(msg.IsLocal() || msg.GetVMIndex() == s.LeaderVMIndex) &&
			s.LeaderPL.DBHeight+1 >= s.GetHighestKnownBlock() {
			if len(vm.List) == 0 {
				s.SendDBSig(s.LLeaderHeight, s.LeaderVMIndex)
				s.XReview = append(s.XReview, msg)
			} else {
				msg.LeaderExecute(s)
			}
		} else {
			msg.FollowerExecute(s)
		}
		ret = true
	case 0:
		s.Holding[msg.GetMsgHash().Fixed()] = msg
	default:
		s.Holding[msg.GetMsgHash().Fixed()] = msg
		if !msg.SentInvlaid() {
			msg.MarkSentInvalid(true)
			s.networkInvalidMsgQueue <- msg
		}
	}

	return

}

func (s *State) Process() (progress bool) {

	if !s.RunLeader {
		now := s.GetTimestamp().GetTimeMilli() // Timestamps are in milliseconds, so wait 20
		if now-s.StartDelay > s.StartDelayLimit {
			s.RunLeader = true
		}
		s.LeaderPL = s.ProcessLists.Get(s.LLeaderHeight)
		s.Leader, s.LeaderVMIndex = s.LeaderPL.GetVirtualServers(s.CurrentMinute, s.IdentityChainID)
	}

	var vm *VM
	if s.Leader {
		vm = s.LeaderPL.VMs[s.LeaderVMIndex]
	}

	s.ReviewHolding()

	more := false
	// Process acknowledgements if we have some.
ackLoop:
	for i := 0; i < 55; i++ {
		select {
		case ack := <-s.ackQueue:
			a := ack.(*messages.Ack)
			if a.DBHeight >= s.LLeaderHeight && ack.Validate(s) == 1 {
				ack.FollowerExecute(s)
			}
			progress = true
		default:
			more = true
			break ackLoop
		}
	}

	// Process inbound messages
emptyLoop:
	for i := 0; i < 55; i++ {
		select {
		case msg := <-s.msgQueue:
			if s.executeMsg(vm, msg) && !msg.IsPeer2Peer() {
				msg.SendOut(s, msg)
			}
		default:
			more = true
			break emptyLoop
		}
	}

	// Reprocess any stalled messages, but not so much compared inbound messages
	// Process last first
	for i := 0; i < 30 && len(s.XReview) > 0; i++ {
		l := len(s.XReview) - 1
		msg := s.XReview[l]
		progress = s.executeMsg(vm, msg)
		s.XReview = s.XReview[:l]
	}
	if !more {
		time.Sleep(10 * time.Millisecond)
	}

	return
}

//***************************************************************
// Checkpoint DBKeyMR
//***************************************************************
func CheckDBKeyMR(s *State, ht uint32, hash string) error {
	if s.Network != "MAIN" && s.Network != "main" {
		return nil
	}
	if val, ok := constants.CheckPoints[ht]; ok {
		if val != hash {
			return fmt.Errorf("%20s CheckPoints at %d DB height failed\n", s.FactomNodeName, ht)
		}
	}
	return nil
}

//***************************************************************
// Consensus Methods
//***************************************************************

// Places the entries in the holding map back into the XReview list for
// review if this is a leader, and those messages are that leader's
// responsibility
func (s *State) ReviewHolding() {
	if len(s.XReview) > 0 {
		return
	}
	now := s.GetTimestamp()
	if s.resendHolding == nil {
		s.resendHolding = now
	}
	if now.GetTimeMilli()-s.resendHolding.GetTimeMilli() < 100 {
		return
	}
	s.resendHolding = now
	// Anything we are holding, we need to reprocess.
	s.XReview = make([]interfaces.IMsg, 0)

	for k := range s.Holding {
		v := s.Holding[k]

		_, ok := s.Replay.Valid(constants.INTERNAL_REPLAY, v.GetRepeatHash().Fixed(), v.GetTimestamp(), s.GetTimestamp())
		if !ok {
			delete(s.Holding, k)
			continue
		}

		if v.Expire(s) {
			s.ExpireCnt++
			delete(s.Holding, k)
			continue
		}

		if v.Resend(s) {
			if v.Validate(s) == 1 {
				s.ResendCnt++
				v.SendOut(s, v)
			}
		}
		// Pointless to review a Reveal Entry;  it will be pulled into play when its commit
		// comes around.
		if re, ok := v.(*messages.RevealEntryMsg); ok {
			if s.Commits[re.Entry.GetHash().Fixed()] != nil {
				s.XReview = append(s.XReview, v)
				delete(s.Holding, k)
			}
		} else {
			s.XReview = append(s.XReview, v)
			delete(s.Holding, k)
		}
	}
}

// Adds blocks that are either pulled locally from a database, or acquired from peers.
func (s *State) AddDBState(isNew bool,
	directoryBlock interfaces.IDirectoryBlock,
	adminBlock interfaces.IAdminBlock,
	factoidBlock interfaces.IFBlock,
	entryCreditBlock interfaces.IEntryCreditBlock,
	eBlocks []interfaces.IEntryBlock,
	entries []interfaces.IEBEntry) *DBState {

	dbState := s.DBStates.NewDBState(isNew, directoryBlock, adminBlock, factoidBlock, entryCreditBlock, eBlocks, entries)

	if dbState == nil {
		return nil
	}

	ht := dbState.DirectoryBlock.GetHeader().GetDBHeight()
	DBKeyMR := dbState.DirectoryBlock.GetKeyMR().String()

	err := CheckDBKeyMR(s, ht, DBKeyMR)
	if err != nil {
		panic(fmt.Errorf("Found block at height %d that didn't match a checkpoint. Got %s, expected %s", ht, DBKeyMR, constants.CheckPoints[ht])) //TODO make failing when given bad blocks fail more elegantly
	}

	if ht > s.LLeaderHeight {
		s.Syncing = false
		s.EOM = false
		s.DBSig = false
		s.LLeaderHeight = ht
		s.ProcessLists.Get(ht + 1)
		s.CurrentMinute = 0
		s.EOMProcessed = 0
		s.DBSigProcessed = 0
		s.StartDelay = s.GetTimestamp().GetTimeMilli()
		s.RunLeader = false
		s.Newblk = true
		s.LeaderPL = s.ProcessLists.Get(s.LLeaderHeight)

		{
			// Okay, we have just loaded a new DBState.  The temp balances are no longer valid, if they exist.  Nuke them.
			s.LeaderPL.FactoidBalancesTMutex.Lock()
			defer s.LeaderPL.FactoidBalancesTMutex.Unlock()

			s.LeaderPL.ECBalancesTMutex.Lock()
			defer s.LeaderPL.ECBalancesTMutex.Unlock()

			s.LeaderPL.FactoidBalancesT = map[[32]byte]int64{}
			s.LeaderPL.ECBalancesT = map[[32]byte]int64{}
		}

		s.Leader, s.LeaderVMIndex = s.LeaderPL.GetVirtualServers(s.CurrentMinute, s.IdentityChainID)
		s.ProcessLists.UpdateState(s.LLeaderHeight)
	}
	if ht == 0 && s.LLeaderHeight < 1 {
		s.LLeaderHeight = 1
	}

	return dbState
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

	s.Holding[m.GetMsgHash().Fixed()] = m

	ack, _ := s.Acks[m.GetMsgHash().Fixed()].(*messages.Ack)
	if ack != nil {
		pl := s.ProcessLists.Get(ack.DBHeight)
		pl.AddToProcessList(ack, m)
	}
}

// Ack messages always match some message in the Process List.   That is
// done here, though the only msg that should call this routine is the Ack
// message.
func (s *State) FollowerExecuteAck(msg interfaces.IMsg) {
	ack := msg.(*messages.Ack)

	pl := s.ProcessLists.Get(ack.DBHeight)
	if pl == nil {
		return
	}
	list := pl.VMs[ack.VMIndex].List
	if len(list) > int(ack.Height) && list[ack.Height] != nil {
		return
	}

	s.Acks[ack.GetHash().Fixed()] = ack
	m, _ := s.Holding[ack.GetHash().Fixed()]
	if m != nil {
		m.FollowerExecute(s)
	}
}

func (s *State) FollowerExecuteDBState(msg interfaces.IMsg) {
	dbstatemsg, _ := msg.(*messages.DBStateMsg)

	dbheight := dbstatemsg.DirectoryBlock.GetHeader().GetDBHeight()

	if s.GetHighestCompletedBlock() > dbheight {
		return
	}
	pdbstate := s.DBStates.Get(int(dbheight - 1))

	switch pdbstate.ValidNext(s, dbstatemsg.DirectoryBlock) {
	case 0:
		k := fmt.Sprint("dbstate", dbheight-1)
		key := primitives.NewHash([]byte(k))
		s.Holding[key.Fixed()] = msg
		return
	case -1:
		// Kill the previous DBState, because it could be bad.
		if dbheight > s.DBStates.Base && len(s.DBStates.DBStates) > int(dbheight-s.DBStates.Base) {
			s.DBStates.DBStates[dbheight-s.DBStates.Base] = nil
			s.DBStateFailsCnt++
			s.networkInvalidMsgQueue <- msg
		}
		return
	}

	s.DBStates.LastTime = s.GetTimestamp()
	dbstate := s.AddDBState(false, // Not a new block; got it from the network
		dbstatemsg.DirectoryBlock,
		dbstatemsg.AdminBlock,
		dbstatemsg.FactoidBlock,
		dbstatemsg.EntryCreditBlock,
		dbstatemsg.EBlocks,
		dbstatemsg.Entries)
	if dbstate == nil {
		s.DBStateFailsCnt++
	} else {
		dbstate.ReadyToSave = true
	}
}

func (s *State) FollowerExecuteNegotiation(m interfaces.IMsg) {
	fmt.Println("JUSTIN : No more negotiation messages")
}

func (s *State) FollowerExecuteMMR(m interfaces.IMsg) {
	mmr, _ := m.(*messages.MissingMsgResponse)
	ack := mmr.AckResponse.(*messages.Ack)

	// If we don't need this message, we don't have to do everything else.
	if ack.Validate(s) == -1 {
		return
	}

	ack.Response = true
	msg := mmr.MsgResponse
	pl := s.ProcessLists.Get(ack.DBHeight)
	_, okr := s.Replay.Valid(constants.INTERNAL_REPLAY, ack.GetRepeatHash().Fixed(), ack.GetTimestamp(), s.GetTimestamp())
	_, okm := s.Replay.Valid(constants.INTERNAL_REPLAY, msg.GetRepeatHash().Fixed(), msg.GetTimestamp(), s.GetTimestamp())

	if pl == nil {
		return
	}

	s.Acks[ack.GetHash().Fixed()] = ack

	if okr {
		ack.FollowerExecute(s)
	}
	if okm {
		msg.FollowerExecute(s)
	}
	if !okr && !okm {
		pl.AddToProcessList(ack, msg)
	}

	if s.Acks[ack.GetHash().Fixed()] == nil {
		s.MissingResponseAppliedCnt++
	}

}

func (s *State) FollowerExecuteDataResponse(m interfaces.IMsg) {
	msg, ok := m.(*messages.DataResponse)
	if !ok {
		return
	}

	now := s.GetTimestamp()

	//fmt.Println("JUSTIN", s.FactomNodeName, "FOLLEX DR:", msg.DataType, msg.DataHash.String())

	switch msg.DataType {
	case 1: // Data is an entryBlock
		eblock, ok := msg.DataObject.(interfaces.IEntryBlock)
		if !ok {
			//fmt.Println("JUSTIN", s.FactomNodeName, "EBLOCK NOT OK", msg.DataHash.String())
			return
		}

		ebKeyMR, _ := eblock.KeyMR()
		if ebKeyMR == nil {
			//fmt.Println("JUSTIN", s.FactomNodeName, "EBKMR NIL", msg.DataHash.String(), ebKeyMR.String())
			return
		}

		for i, missing := range s.MissingEntryBlocks {
			eb := missing.ebhash
			if !eb.IsSameAs(ebKeyMR) {
				continue
			}

			db, err := s.DB.FetchDBlockByHeight(eblock.GetHeader().GetDBHeight())
			if err != nil || db == nil {
				return
			}

			//fmt.Println("JUSTIN", s.FactomNodeName, "FOUND EB", msg.DataHash.String())
			s.MissingEntryBlocks = append(s.MissingEntryBlocks[:i], s.MissingEntryBlocks[i+1:]...)
			s.DB.ProcessEBlockBatch(eblock, true)

			//s.DB.ProcessEBlockBatch(eblock, true)

			for _, entryhash := range eblock.GetEntryHashes() {
				if entryhash.IsMinuteMarker() {
					continue
				}
				e, _ := s.DB.FetchEntry(entryhash)
				if e == nil {
					/*if s.EntryBlockDBHeightComplete >= eblock.GetDatabaseHeight() {
						s.EntryBlockDBHeightComplete = eblock.GetDatabaseHeight() - 1
						if s.EntryBlockDBHeightComplete < 0 {
							s.EntryBlockDBHeightComplete = 0
						}
					}*/
					var v struct {
						ebhash    interfaces.IHash
						entryhash interfaces.IHash
						dbheight  uint32
					}

					v.dbheight = eblock.GetHeader().GetDBHeight()
					v.entryhash = entryhash
					v.ebhash = eb
					//fmt.Println("JUSTIN", s.FactomNodeName, "FROM EB APP ", entryhash.String())

					s.MissingEntries = append(s.MissingEntries, v)

					// Save the entry hash, and remove from commits IF this hash is valid in this current timeframe.
					if s.Replay.IsTSValid_(constants.REVEAL_REPLAY, entryhash.Fixed(), db.GetTimestamp(), now) {
						delete(s.Commits, entryhash.Fixed())
					}

				}

			}

			mindb := s.GetDBHeightComplete() + 1
			for _, missingleft := range s.MissingEntryBlocks {
				if missingleft.dbheight <= mindb {
					mindb = missingleft.dbheight
				}
			}
			s.EntryBlockDBHeightComplete = mindb - 1
			//fmt.Println("JUSTIN", s.FactomNodeName, "NOW EBDHBC IS", s.EntryBlockDBHeightComplete)
			break
		}

	case 0: // Data is an entry
		entry, ok := msg.DataObject.(interfaces.IEBEntry)
		if !ok {
			//fmt.Println("JUSTIN", s.FactomNodeName, "NOT OK ENTRY", msg.DataHash.String())
			return
		}

		for i, missing := range s.MissingEntries {
			e := missing.entryhash
			//fmt.Println("JUSTIN", s.FactomNodeName, "FOUND ENT", msg.DataHash.String())

			if e.IsSameAs(entry.GetHash()) {
				//fmt.Println("JUSTIN", s.FactomNodeName, "FOUND ENT AND MATCH", msg.DataHash.String())
				s.DB.InsertEntry(entry)
				s.MissingEntries = append(s.MissingEntries[:i], s.MissingEntries[i+1:]...)
				break
			}
		}
	}
}

func (s *State) FollowerExecuteMissingMsg(msg interfaces.IMsg) {
	m := msg.(*messages.MissingMsg)

	for _, h := range m.ProcessListHeight {
		missingmsg, ackMsg, err := s.LoadSpecificMsgAndAck(m.DBHeight, m.VMIndex, h)

		if missingmsg != nil && ackMsg != nil && err == nil {
			// If I don't have this message, ignore.
			msgResponse := messages.NewMissingMsgResponse(s, missingmsg, ackMsg)
			msgResponse.SetOrigin(m.GetOrigin())
			msgResponse.SetNetworkOrigin(m.GetNetworkOrigin())
			s.NetworkOutMsgQueue() <- msgResponse
			s.MissingRequestReplyCnt++
		} else {
			s.MissingRequestIgnoreCnt++
		}
	}
	return
}

func (s *State) FollowerExecuteCommitChain(m interfaces.IMsg) {
	s.FollowerExecuteMsg(m)
	cc := m.(*messages.CommitChainMsg)
	re := s.Holding[cc.CommitChain.EntryHash.Fixed()]
	if re != nil {
		s.XReview = append(s.XReview, re)
		re.SendOut(s, re)
	}
}

func (s *State) FollowerExecuteCommitEntry(m interfaces.IMsg) {
	s.FollowerExecuteMsg(m)
	ce := m.(*messages.CommitEntryMsg)
	re := s.Holding[ce.CommitEntry.EntryHash.Fixed()]
	if re != nil {
		s.XReview = append(s.XReview, re)
		re.SendOut(s, re)
	}
}

func (s *State) FollowerExecuteRevealEntry(m interfaces.IMsg) {
	s.Holding[m.GetMsgHash().Fixed()] = m
	ack, _ := s.Acks[m.GetMsgHash().Fixed()].(*messages.Ack)

	if ack != nil {

		m.SetLeaderChainID(ack.GetLeaderChainID())
		m.SetMinute(ack.Minute)

		pl := s.ProcessLists.Get(ack.DBHeight)
		pl.AddToProcessList(ack, m)

		// If we added the ack, then it will be cleared from the ack map.
		if s.Acks[m.GetMsgHash().Fixed()] == nil {
			msg := m.(*messages.RevealEntryMsg)
			delete(s.Commits, msg.Entry.GetHash().Fixed())
			// Okay the Reveal has been recorded.  Record this as an entry that cannot be duplicated.
			s.Replay.IsTSValid_(constants.REVEAL_REPLAY, msg.Entry.GetHash().Fixed(), msg.Timestamp, s.GetTimestamp())
		}

	}

}

func (s *State) LeaderExecute(m interfaces.IMsg) {

	_, ok := s.Replay.Valid(constants.INTERNAL_REPLAY, m.GetRepeatHash().Fixed(), m.GetTimestamp(), s.GetTimestamp())
	if !ok {
		delete(s.Holding, m.GetRepeatHash().Fixed())
		delete(s.Holding, m.GetMsgHash().Fixed())
		return
	}

	ack := s.NewAck(m).(*messages.Ack)
	m.SetLeaderChainID(ack.GetLeaderChainID())
	m.SetMinute(ack.Minute)
	s.ProcessLists.Get(ack.DBHeight).AddToProcessList(ack, m)

}

func (s *State) LeaderExecuteEOM(m interfaces.IMsg) {

	if !m.IsLocal() {
		s.FollowerExecuteEOM(m)
		return
	}

	// The zero based minute for the message is equal to
	// the one based "LastMinute".  This way we know we are
	// generating minutes in order.

	eom := m.(*messages.EOM)
	pl := s.ProcessLists.Get(s.LLeaderHeight)
	vm := pl.VMs[s.LeaderVMIndex]
	if s.Syncing && vm.Synced {
		return
	} else if !s.Syncing {
		s.Syncing = true
		s.EOM = true
		s.EOMsyncing = true
		s.EOMProcessed = 0
		for _, vm := range pl.VMs {
			vm.Synced = false
		}
		s.EOMLimit = len(pl.FedServers)
		s.EOMMinute = int(s.CurrentMinute)
	}

	//_, vmindex := pl.GetVirtualServers(s.EOMMinute, s.IdentityChainID)

	eom.DBHeight = s.LLeaderHeight
	eom.VMIndex = s.LeaderVMIndex
	// eom.Minute is zerobased, while LeaderMinute is 1 based.  So
	// a simple assignment works.
	eom.Minute = byte(s.CurrentMinute)
	eom.Sign(s)
	eom.MsgHash = nil
	ack := s.NewAck(m)
	s.Acks[eom.GetMsgHash().Fixed()] = ack
	m.SetLocal(false)
	s.FollowerExecuteEOM(m)
	s.UpdateState()
}

func (s *State) LeaderExecuteCommitChain(m interfaces.IMsg) {
	s.LeaderExecute(m)
	cc := m.(*messages.CommitChainMsg)
	re := s.Holding[cc.CommitChain.EntryHash.Fixed()]
	if re != nil {
		s.XReview = append(s.XReview, re)
		re.SendOut(s, re)
	}
}

func (s *State) LeaderExecuteCommitEntry(m interfaces.IMsg) {
	s.LeaderExecute(m)
	ce := m.(*messages.CommitEntryMsg)
	re := s.Holding[ce.CommitEntry.EntryHash.Fixed()]
	if re != nil {
		s.XReview = append(s.XReview, re)
		re.SendOut(s, re)
	}
}

func (s *State) LeaderExecuteRevealEntry(m interfaces.IMsg) {
	re := m.(*messages.RevealEntryMsg)
	eh := re.Entry.GetHash()

	commit, rtn := re.ValidateRTN(s)

	switch rtn {
	case 0:
		m.FollowerExecute(s)
	case -1:
		return
	}
	now := s.GetTimestamp()
	// If we have already recorded a Reveal Entry with this hash in this period, just ignore.
	if _, v := s.Replay.Valid(constants.REVEAL_REPLAY, eh.Fixed(), s.GetLeaderTimestamp(), now); !v {
		return
	}

	ack := s.NewAck(m).(*messages.Ack)

	m.SetLeaderChainID(ack.GetLeaderChainID())
	m.SetMinute(ack.Minute)

	// Put the acknowledgement in the Acks so we can tell if AddToProcessList() adds it.
	s.Acks[m.GetMsgHash().Fixed()] = ack
	s.ProcessLists.Get(ack.DBHeight).AddToProcessList(ack, m)
	// If it was added, then get rid of the matching Commit.
	if s.Acks[m.GetMsgHash().Fixed()] != nil {
		m.FollowerExecute(s)
		s.PutCommit(eh, commit)
	} else {
		// Okay the Reveal has been recorded.  Record this as an entry that cannot be duplicated.
		s.Replay.IsTSValid_(constants.REVEAL_REPLAY, eh.Fixed(), m.GetTimestamp(), now)
		delete(s.Commits, eh.Fixed())
	}
}

func (s *State) ProcessAddServer(dbheight uint32, addServerMsg interfaces.IMsg) bool {
	as, ok := addServerMsg.(*messages.AddServerMsg)
	if ok && !ProcessIdentityToAdminBlock(s, as.ServerChainID, as.ServerType) {
		return false
	}
	return true
}

func (s *State) ProcessRemoveServer(dbheight uint32, removeServerMsg interfaces.IMsg) bool {
	rs, ok := removeServerMsg.(*messages.RemoveServerMsg)
	if !ok {
		return true
	}

	if !s.VerifyIsAuthority(rs.ServerChainID) {
		fmt.Printf("dddd %s %s\n", s.FactomNodeName, "RemoveServer message did not add to admin block. Not an Authority")
		return true
	}

	if s.GetAuthorityServerType(rs.ServerChainID) != rs.ServerType {
		fmt.Printf("dddd %s %s\n", s.FactomNodeName, "RemoveServer message did not add to admin block. Servertype of message did not match authority's")
		return true
	}

	if len(s.LeaderPL.FedServers) < 2 && rs.ServerType == 0 {
		fmt.Printf("dddd %s %s\n", s.FactomNodeName, "RemoveServer message did not add to admin block. Only 1 federated server exists.")
		return true
	}
	s.LeaderPL.AdminBlock.RemoveFederatedServer(rs.ServerChainID)

	return true
}

func (s *State) ProcessChangeServerKey(dbheight uint32, changeServerKeyMsg interfaces.IMsg) bool {
	ask, ok := changeServerKeyMsg.(*messages.ChangeServerKeyMsg)
	if !ok {
		return true
	}

	if !s.VerifyIsAuthority(ask.IdentityChainID) {
		fmt.Printf("dddd %s %s\n", s.FactomNodeName, "ChangeServerKey message did not add to admin block.")
		return true
	}

	switch ask.AdminBlockChange {
	case constants.TYPE_ADD_BTC_ANCHOR_KEY:
		var btcKey [20]byte
		copy(btcKey[:], ask.Key.Bytes()[:20])
		s.LeaderPL.AdminBlock.AddFederatedServerBitcoinAnchorKey(ask.IdentityChainID, ask.KeyPriority, ask.KeyType, &btcKey)
	case constants.TYPE_ADD_FED_SERVER_KEY:
		pub := ask.Key.Fixed()
		s.LeaderPL.AdminBlock.AddFederatedServerSigningKey(ask.IdentityChainID, &pub)
	case constants.TYPE_ADD_MATRYOSHKA:
		s.LeaderPL.AdminBlock.AddMatryoshkaHash(ask.IdentityChainID, ask.Key)
	}
	return true
}

func (s *State) ProcessCommitChain(dbheight uint32, commitChain interfaces.IMsg) bool {
	c, _ := commitChain.(*messages.CommitChainMsg)

	pl := s.ProcessLists.Get(dbheight)
	pl.EntryCreditBlock.GetBody().AddEntry(c.CommitChain)
	if e := s.GetFactoidState().UpdateECTransaction(true, c.CommitChain); e == nil {
		// save the Commit to match agains the Reveal later
		h := c.CommitChain.EntryHash
		s.PutCommit(h, c)
		if s.Holding[h.Fixed()] != nil {
			entry := s.Holding[h.Fixed()]
			entry.SendOut(s, entry)
			s.XReview = append(s.XReview, entry)
			delete(s.Holding, h.Fixed())
		}
		return true
	} else {
		fmt.Println(e)
	}

	return false
}

func (s *State) ProcessCommitEntry(dbheight uint32, commitEntry interfaces.IMsg) bool {
	c, _ := commitEntry.(*messages.CommitEntryMsg)

	pl := s.ProcessLists.Get(dbheight)
	pl.EntryCreditBlock.GetBody().AddEntry(c.CommitEntry)
	if e := s.GetFactoidState().UpdateECTransaction(true, c.CommitEntry); e == nil {
		// save the Commit to match agains the Reveal later
		h := c.CommitEntry.EntryHash
		s.PutCommit(h, c)
		if s.Holding[h.Fixed()] != nil {
			entry := s.Holding[h.Fixed()]
			entry.SendOut(s, entry)
			s.XReview = append(s.XReview, entry)
			delete(s.Holding, h.Fixed())
		}
		return true
	} else {
		fmt.Println(e)
	}
	return false
}

func (s *State) ProcessRevealEntry(dbheight uint32, m interfaces.IMsg) bool {

	msg := m.(*messages.RevealEntryMsg)
	myhash := msg.Entry.GetHash()

	chainID := msg.Entry.GetChainID()

	delete(s.Commits, msg.Entry.GetHash().Fixed())

	eb := s.GetNewEBlocks(dbheight, chainID)
	eb_db := s.GetNewEBlocks(dbheight-1, chainID)
	if eb_db == nil {
		eb_db, _ = s.DB.FetchEBlockHead(chainID)
	}
	// Handle the case that this is a Entry Chain create
	// Must be built with CommitChain (i.e. !msg.IsEntry).  Also
	// cannot have an existing chaing (eb and eb_db == nil)
	if !msg.IsEntry && eb == nil && eb_db == nil {
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

	// Create an entry (even if they used commitChain).  Means there must
	// be a chain somewhere.  If not, we return false.
	if eb == nil {
		if eb_db == nil {
			return false
		}
		eb = entryBlock.NewEBlock()
		eb.GetHeader().SetEBSequence(eb_db.GetHeader().GetEBSequence() + 1)
		eb.GetHeader().SetPrevFullHash(eb_db.GetHash())
		// Set the Chain ID
		eb.GetHeader().SetChainID(chainID)
		// Set the Directory Block Height for this Entry Block
		eb.GetHeader().SetDBHeight(dbheight)
		// Set the PrevKeyMR
		key, _ := eb_db.KeyMR()
		eb.GetHeader().SetPrevKeyMR(key)
	}
	// Add our new entry
	eb.AddEBEntry(msg.Entry)
	// Put it in our list of new Entry Blocks for this Directory Block
	s.PutNewEBlocks(dbheight, chainID, eb)
	s.PutNewEntries(dbheight, myhash, msg.Entry)

	// Monitor key changes for fed/audit servers
	LoadIdentityByEntry(msg.Entry, s, dbheight, false)

	s.IncEntries()
	return true
}

// dbheight is the height of the process list, and vmIndex is the vm
// that is missing the DBSig.  If the DBSig isn't our responsiblity, then
// this call will do nothing.  Assumes the state for the leader is set properly
func (s *State) SendDBSig(dbheight uint32, vmIndex int) {
	ht := s.GetHighestCompletedBlock()
	if dbheight <= ht || s.EOM {
		return
	}
	pl := s.ProcessLists.Get(dbheight)
	vm := pl.VMs[vmIndex]
	if vm.LeaderMinute > 9 {
		return
	}
	leader, lvm := pl.GetVirtualServers(vm.LeaderMinute, s.IdentityChainID)
	if leader && !vm.Signed {
		dbstate := s.DBStates.Get(int(dbheight - 1))
		if dbstate == nil && dbheight > 0 {
			s.SendDBSig(dbheight-1, vmIndex)
			return
		}
		if lvm == vmIndex {
			dbs := new(messages.DirectoryBlockSignature)
			dbs.DirectoryBlockHeader = dbstate.DirectoryBlock.GetHeader()
			//dbs.DirectoryBlockKeyMR = dbstate.DirectoryBlock.GetKeyMR()
			dbs.ServerIdentityChainID = s.GetIdentityChainID()
			dbs.DBHeight = dbheight
			dbs.Timestamp = s.GetTimestamp()
			dbs.SetVMHash(nil)
			dbs.SetVMIndex(vmIndex)
			dbs.SetLocal(true)
			dbs.Sign(s)
			err := dbs.Sign(s)
			if err != nil {
				panic(err)
			}
			dbs.LeaderExecute(s)
			vm.Signed = true
		}
	}
}

// TODO: Should fault the server if we don't have the proper sequence of EOM messages.
func (s *State) ProcessEOM(dbheight uint32, msg interfaces.IMsg) bool {

	e := msg.(*messages.EOM)

	if s.Syncing && !s.EOM {
		return false
	}

	if s.EOM && int(e.Minute) > s.EOMMinute {
		return false
	}

	pl := s.ProcessLists.Get(dbheight)
	vm := s.ProcessLists.Get(dbheight).VMs[msg.GetVMIndex()]

	// If I have done everything for all EOMs for all VMs, then and only then do I
	// let processing continue.
	if s.EOMDone {
		s.EOMProcessed--
		if s.EOMProcessed == 0 {
			s.EOM = false
			s.EOMDone = false
			s.ReviewHolding()
			s.Syncing = false
		}
		s.SendHeartBeat()

		return true
	}

	// What I do once  for all VMs at the beginning of processing a particular EOM
	if !s.EOM {
		s.Syncing = true
		s.EOM = true
		s.EOMLimit = len(s.LeaderPL.FedServers)
		s.EOMMinute = int(e.Minute)
		s.EOMsyncing = true
		s.EOMProcessed = 0
		s.Newblk = false

		for _, vm := range pl.VMs {
			vm.Synced = false
		}
		return false
	}

	// What I do for each EOM
	if !e.Processed {
		vm.LeaderMinute++
		s.EOMProcessed++
		e.Processed = true
		vm.Synced = true

		return false
	}

	allfaults := s.LeaderPL.System.Height == s.LeaderPL.SysHighest

	// After all EOM markers are processed, Claim we are done.  Now we can unwind
	if allfaults && s.EOMProcessed == s.EOMLimit && !s.EOMDone {

		s.EOMDone = true
		for _, eb := range pl.NewEBlocks {
			eb.AddEndOfMinuteMarker(byte(e.Minute + 1))
		}

		s.FactoidState.EndOfPeriod(int(e.Minute))

		ecblk := pl.EntryCreditBlock
		ecbody := ecblk.GetBody()
		mn := entryCreditBlock.NewMinuteNumber(e.Minute + 1)
		ecbody.AddEntry(mn)

		if !s.Leader {
			s.CurrentMinute = int(e.Minute)
		}

		s.CurrentMinute++

		switch {
		case s.CurrentMinute < 10:
			s.LeaderPL = s.ProcessLists.Get(s.LLeaderHeight)
			s.Leader, s.LeaderVMIndex = s.LeaderPL.GetVirtualServers(s.CurrentMinute, s.IdentityChainID)
		case s.CurrentMinute == 10:
			eBlocks := []interfaces.IEntryBlock{}
			entries := []interfaces.IEBEntry{}
			for _, v := range pl.NewEBlocks {
				eBlocks = append(eBlocks, v)
			}
			for _, v := range pl.NewEntries {
				entries = append(entries, v)
			}

			dbstate := s.AddDBState(true, s.LeaderPL.DirectoryBlock, s.LeaderPL.AdminBlock, s.GetFactoidState().GetCurrentBlock(), s.LeaderPL.EntryCreditBlock, eBlocks, entries)
			if dbstate == nil {
				dbstate = s.DBStates.Get(int(s.LeaderPL.DirectoryBlock.GetHeader().GetDBHeight()))
			}
			dbht := int(dbstate.DirectoryBlock.GetHeader().GetDBHeight())
			if dbht > 0 {
				prev := s.DBStates.Get(dbht - 1)
				s.DBStates.FixupLinks(prev, dbstate)
			}
			s.DBStates.ProcessBlocks(dbstate)

			s.CurrentMinute = 0
			s.LLeaderHeight++
			s.LeaderPL = s.ProcessLists.Get(s.LLeaderHeight)
			s.Leader, s.LeaderVMIndex = s.LeaderPL.GetVirtualServers(0, s.IdentityChainID)

			s.DBSigProcessed = 0

			// Note about dbsigs.... If we processed the previous minute, then we generate the DBSig for the next block.
			// But if we didn't process the preivious block, like we start from scratch, or we had to reset the entire
			// network, then no dbsig exists.  This code doesn't execute, and so we have no dbsig.  In that case, on
			// the next EOM, we see the block hasn't been signed, and we sign the block (Thats the call to SendDBSig()
			// above).
			if s.Leader {
				dbstate := s.DBStates.Get(int(s.LLeaderHeight - 1))
				dbs := new(messages.DirectoryBlockSignature)
				dbs.DirectoryBlockHeader = dbstate.DirectoryBlock.GetHeader()
				//dbs.DirectoryBlockKeyMR = dbstate.DirectoryBlock.GetKeyMR()
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

			s.Saving = true
		}

		for k := range s.Commits {
			vs := s.Commits[k]
			if len(vs) == 0 {
				delete(s.Commits, k)
				continue
			}
			v, ok := vs[0].(interfaces.IMsg)
			if ok {
				_, ok := s.Replay.Valid(constants.TIME_TEST, v.GetRepeatHash().Fixed(), v.GetTimestamp(), s.GetTimestamp())
				if !ok {
					copy(vs, vs[1:])
					vs[len(vs)-1] = nil
					s.Commits[k] = vs[:len(vs)-1]
				}
			}
		}

		for k := range s.Acks {
			v := s.Acks[k].(*messages.Ack)
			if v.DBHeight < s.LLeaderHeight {
				delete(s.Acks, k)
			}
		}
	}

	return false
}

// When we process the directory Signature, and we are the leader for said signature, it
// is then that we push it out to the rest of the network.  Otherwise, if we are not the
// leader for the signature, it marks the sig complete for that list
func (s *State) ProcessDBSig(dbheight uint32, msg interfaces.IMsg) bool {

	dbs := msg.(*messages.DirectoryBlockSignature)
	// Don't process if syncing an EOM
	if s.Syncing && !s.DBSig {
		return false
	}

	pl := s.ProcessLists.Get(dbheight)
	vm := s.ProcessLists.Get(dbheight).VMs[msg.GetVMIndex()]

	// If we are done with DBSigs, and this message is processed, then we are done.  Let everything go!
	if s.DBSig && s.DBSigDone {
		s.DBSigProcessed--
		if s.DBSigProcessed <= 0 {
			s.DBSig = false
			s.Syncing = false
		}
		vm.Signed = true
		//s.LeaderPL.AdminBlock
		return true
	}

	// Put the stuff that only executes once at the start of DBSignatures here
	if !s.DBSig {
		s.DBSigLimit = len(pl.FedServers)
		s.DBSigProcessed = 0
		s.DBSig = true
		s.Syncing = true
		s.DBSigDone = false
		for _, vm := range pl.VMs {
			vm.Synced = false
		}
		pl.ResetDiffSigTally()
	}

	// Put the stuff that executes per DBSignature here
	if !dbs.Processed {
		if dbs.VMIndex == 0 {
			s.SetLeaderTimestamp(dbs.GetTimestamp())
		}
		if !dbs.DirectoryBlockHeader.GetBodyMR().IsSameAs(s.GetDBState(dbheight - 1).DirectoryBlock.GetHeader().GetBodyMR()) {
			fmt.Println(s.FactomNodeName, "JUST COMPARED", dbs.DirectoryBlockHeader.GetBodyMR().String()[:10], " : ", s.GetDBState(dbheight - 1).DirectoryBlock.GetHeader().GetBodyMR().String()[:10])
			pl.IncrementDiffSigTally()
		}

		// Adds DB Sig to be added to Admin block if passes sig checks
		allChecks := false
		data, err := dbs.DirectoryBlockHeader.MarshalBinary()
		if err != nil {
			fmt.Println("Debug: DBSig Signature Error, Marshal binary errored")
		} else {
			if !dbs.DBSignature.Verify(data) {
				fmt.Println("Debug: DBSig Signature Error, Verify errored")
			} else {
				if valid, err := s.VerifyAuthoritySignature(data, dbs.DBSignature.GetSignature(), dbs.DBHeight); err == nil && valid == 1 {
					allChecks = true
				}
			}
		}

		if allChecks {
			s.AddDBSig(dbheight, dbs.ServerIdentityChainID, dbs.DBSignature)
		}

		dbs.Processed = true
		s.DBSigProcessed++
		vm.Synced = true
	}

	// Put the stuff that executes once for set of DBSignatures (after I have them all) here
	if s.DBSigProcessed >= s.DBSigLimit {
		dbstate := s.DBStates.Get(int(dbheight - 1))

		// TODO: check signatures here.  Count what match and what don't.  Then if a majority
		// disagree with us, null our entry out.  Otherwise toss our DBState and ask for one from
		// our neighbors.
		if s.KeepMismatch || pl.CheckDiffSigTally() {
			if !dbstate.Saved {
				dbstate.ReadyToSave = true
				s.DBStates.SaveDBStateToDB(dbstate)
				//s.LeaderPL.AddDBSig(dbs.ServerIdentityChainID, dbs.DBSignature)
			}
		} else {
			s.DBSigFails++
			s.DBStates.DBStates = s.DBStates.DBStates[:len(s.DBStates.DBStates)-1]

			msg := messages.NewDBStateMissing(s, uint32(dbheight-1), uint32(dbheight-1))

			if msg != nil {
				s.RunLeader = false
				s.StartDelay = s.GetTimestamp().GetTimeMilli()
				s.NetworkOutMsgQueue() <- msg
			}
		}
		s.ReviewHolding()
		s.Saving = false
		s.DBSigDone = true
	}
	return false
	/*
		err := s.LeaderPL.AdminBlock.AddDBSig(dbs.ServerIdentityChainID, dbs.DBSignature)
		if err != nil {
			fmt.Printf("Error in adding DB sig to admin block, %s\n", err.Error())
		}
	*/
}

func (s *State) ProcessFullServerFault(dbheight uint32, msg interfaces.IMsg) (haveReplaced bool) {
	// If we are here, this means that the FullFault message is complete
	// and we can execute it as such (replacing the faulted Leader with
	// the nominated Audit server)

	//fmt.Println("FULL FAULT:", s.FactomNodeName, s.GetTimestamp().GetTimeSeconds())

	fullFault, _ := msg.(*messages.FullServerFault)
	relevantPL := s.ProcessLists.Get(fullFault.DBHeight)

	auditServerList := s.GetAuditServers(fullFault.DBHeight)
	var theAuditReplacement interfaces.IFctServer

	for _, auditServer := range auditServerList {
		if auditServer.GetChainID().IsSameAs(fullFault.AuditServerID) {
			theAuditReplacement = auditServer
		}
	}
	if theAuditReplacement == nil {
		// If we don't have any Audit Servers in our Authority set
		// that match the nominated Audit Server in the FullFault,
		// we can't really do anything useful with it
		return
	}

	for listIdx, fedServ := range relevantPL.FedServers {
		if fedServ.GetChainID().IsSameAs(fullFault.ServerID) {
			//fmt.Println("FULL FAULT X:", s.FactomNodeName, fullFault.ServerID.String()[:10], fullFault.AuditServerID.String()[:10], s.GetTimestamp().GetTimeSeconds())
			relevantPL.FedServers[listIdx] = theAuditReplacement
			relevantPL.FedServers[listIdx].SetOnline(true)
			relevantPL.AddAuditServer(fedServ.GetChainID())
			s.RemoveAuditServer(fullFault.DBHeight, theAuditReplacement.GetChainID())
			// After executing the FullFault successfully, we want to reset
			// to the default state (No One At Fault)
			relevantPL.Unfault()
			haveReplaced = true
			break
		}
	}

	s.Leader, s.LeaderVMIndex = s.LeaderPL.GetVirtualServers(s.CurrentMinute, s.IdentityChainID)
	return haveReplaced
}

func (s *State) GetMsg(vmIndex int, dbheight int, height int) (interfaces.IMsg, error) {

	pl := s.ProcessLists.Get(uint32(dbheight))
	if pl == nil {
		return nil, errors.New("No Process List")
	}
	vms := pl.VMs
	if len(vms) <= vmIndex {
		return nil, errors.New("Bad VM Index")
	}
	vm := vms[vmIndex]
	if vm.Height > height {
		return vm.List[height], nil
	}
	return nil, nil
}

func (s *State) SendHeartBeat() {
	dbstate := s.DBStates.Get(int(s.LLeaderHeight - 1))
	if dbstate == nil {
		return
	}
	for _, auditServer := range s.GetAuditServers(s.LLeaderHeight) {
		if auditServer.GetChainID().IsSameAs(s.IdentityChainID) {
			hb := new(messages.Heartbeat)
			hb.Timestamp = primitives.NewTimestampNow()
			hb.SecretNumber = s.GetSalt(hb.Timestamp)
			hb.DBlockHash = dbstate.DBHash
			hb.IdentityChainID = s.IdentityChainID
			hb.Sign(s.GetServerPrivateKey())
			hb.SendOut(s, hb)
		}
	}
}

func (s *State) UpdateECs(ec interfaces.IEntryCreditBlock) {
	now := s.GetTimestamp()
	for _, entry := range ec.GetEntries() {
		cc, ok := entry.(*entryCreditBlock.CommitChain)
		if ok && s.Replay.IsTSValid_(constants.INTERNAL_REPLAY, cc.GetHash().Fixed(), cc.GetTimestamp(), now) {
			if s.NoEntryYet(cc.EntryHash, cc.GetTimestamp()) {
				cmsg := new(messages.CommitChainMsg)
				cmsg.CommitChain = cc
				s.PutCommit(cc.EntryHash, cmsg)
			}
			continue
		}
		ce, ok := entry.(*entryCreditBlock.CommitEntry)
		if ok && s.Replay.IsTSValid_(constants.INTERNAL_REPLAY, ce.GetHash().Fixed(), ce.GetTimestamp(), now) {
			if s.NoEntryYet(ce.EntryHash, ce.GetTimestamp()) {
				emsg := new(messages.CommitEntryMsg)
				emsg.CommitEntry = ce
				s.PutCommit(ce.EntryHash, emsg)
			}
			continue
		}
	}
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
	pl.AddNewEBlocks(hash, eb)
}

func (s *State) PutNewEntries(dbheight uint32, hash interfaces.IHash, e interfaces.IEntry) {
	pl := s.ProcessLists.Get(dbheight)
	pl.AddNewEntry(hash, e)
}

// Returns the oldest, not processed, Commit received
func (s *State) NextCommit(hash interfaces.IHash) interfaces.IMsg {
	cs := s.Commits[hash.Fixed()]
	if cs == nil {
		return nil
	}

	if len(cs) == 0 {
		delete(s.Commits, hash.Fixed())
		return nil
	}

	r := cs[0]

	copy(cs[:], cs[1:])
	cs[len(cs)-1] = nil
	s.Commits[hash.Fixed()] = cs[:len(cs)-1]

	return r
}

func (s *State) PutCommit(hash interfaces.IHash, msg interfaces.IMsg) {
	cs := s.Commits[hash.Fixed()]
	if cs == nil {
		cs = make([]interfaces.IMsg, 0)
	}
	s.Commits[hash.Fixed()] = append(cs, msg)
}

// This is the highest block signed off, but not necessarily validted.
func (s *State) GetHighestCompletedBlock() uint32 {
	return s.DBStates.GetHighestCompletedBlock()
}

// This is the highest block signed off and recorded in the Database.
func (s *State) GetHighestSavedBlock() uint32 {
	return s.DBStates.GetHighestSavedBlock()
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

func (s *State) GetF(rt bool, adr [32]byte) (v int64) {
	var ok bool
	if rt {
		pl := s.ProcessLists.Get(s.GetHighestCompletedBlock() + 1)
		if pl != nil {
			pl.FactoidBalancesTMutex.Lock()
			defer pl.FactoidBalancesTMutex.Unlock()
			v, ok = pl.FactoidBalancesT[adr]
		}
	}
	if !ok {
		s.FactoidBalancesPMutex.Lock()
		defer s.FactoidBalancesPMutex.Unlock()
		v = s.FactoidBalancesP[adr]
	}

	return v

}

// If rt == true, update the Temp balances.  Otherwise update the Permenent balances.
func (s *State) PutF(rt bool, adr [32]byte, v int64) {
	if rt {
		pl := s.ProcessLists.Get(s.GetHighestCompletedBlock() + 1)
		if pl != nil {
			pl.FactoidBalancesTMutex.Lock()
			defer pl.FactoidBalancesTMutex.Unlock()

			pl.FactoidBalancesT[adr] = v
		}
	} else {
		s.FactoidBalancesPMutex.Lock()
		defer s.FactoidBalancesPMutex.Unlock()
		s.FactoidBalancesP[adr] = v
	}
}

func (s *State) GetE(rt bool, adr [32]byte) (v int64) {
	var ok bool
	if rt {
		pl := s.ProcessLists.Get(s.GetHighestCompletedBlock() + 1)
		if pl != nil {
			pl.ECBalancesTMutex.Lock()
			defer pl.ECBalancesTMutex.Unlock()
			v, ok = pl.ECBalancesT[adr]
		}
	}
	if !ok {
		s.ECBalancesPMutex.Lock()
		defer s.ECBalancesPMutex.Unlock()
		v = s.ECBalancesP[adr]
	}
	return v

}

// If rt == true, update the Temp balances.  Otherwise update the Permenent balances.
func (s *State) PutE(rt bool, adr [32]byte, v int64) {
	if rt {
		pl := s.ProcessLists.Get(s.GetHighestCompletedBlock() + 1)
		if pl != nil {
			pl.ECBalancesTMutex.Lock()
			defer pl.ECBalancesTMutex.Unlock()
			pl.ECBalancesT[adr] = v
		}
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

func (s *State) GetNetworkName() string {
	return (s.Cfg.(*util.FactomdConfig)).App.Network

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
func (s *State) NewAck(msg interfaces.IMsg) interfaces.IMsg {

	vmIndex := msg.GetVMIndex()

	msg.SetLeaderChainID(s.IdentityChainID)
	ack := new(messages.Ack)
	ack.DBHeight = s.LLeaderHeight
	ack.VMIndex = vmIndex
	ack.Minute = byte(s.ProcessLists.Get(s.LLeaderHeight).VMs[vmIndex].LeaderMinute)
	ack.Timestamp = s.GetTimestamp()
	ack.SaltNumber = s.GetSalt(ack.Timestamp)
	copy(ack.Salt[:8], s.Salt.Bytes()[:8])
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
