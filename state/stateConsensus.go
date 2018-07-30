// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"errors"
	"fmt"
	"hash"
	"sync"
	"time"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/util"
	"github.com/FactomProject/factomd/util/atomic"

	log "github.com/sirupsen/logrus"
)

// consenLogger is the general logger for all consensus related logs. You can add additional fields,
// or create more context loggers off of this
var consenLogger = packageLogger.WithFields(log.Fields{"subpack": "consensus"})

var _ = fmt.Print
var _ = (*hash.Hash32)(nil)

//***************************************************************
// Process Loop for Consensus
//
// Returns true if some message was processed.
//***************************************************************

var once sync.Once
var debugExec_flag bool

func (s *State) DebugExec() (ret bool) {
	once.Do(func() { debugExec_flag = globals.Params.DebugLogRegEx != "" })

	//return s.FactomNodeName == "FNode0"
	return debugExec_flag
}

func (s *State) LogMessage(logName string, comment string, msg interfaces.IMsg) {
	if s.DebugExec() {
		var dbh int
		if s.LeaderPL != nil {
			dbh = int(s.LeaderPL.DBHeight)
		}
		messages.StateLogMessage(s.FactomNodeName, dbh, int(s.CurrentMinute), logName, comment, msg)
	}
}

func (s *State) LogPrintf(logName string, format string, more ...interface{}) {
	if s.DebugExec() {
		var dbh int
		if s.LeaderPL != nil {
			dbh = int(s.LeaderPL.DBHeight)
		}
		messages.StateLogPrintf(s.FactomNodeName, dbh, int(s.CurrentMinute), logName, format, more...)
	}
}
func (s *State) executeMsg(vm *VM, msg interfaces.IMsg) (ret bool) {

	if msg.GetHash() == nil {
		s.LogMessage("badMsgs", "Nil hash in executeMsg", msg)
		return false
	}

	preExecuteMsgTime := time.Now()
	_, ok := s.Replay.Valid(constants.INTERNAL_REPLAY, msg.GetRepeatHash().Fixed(), msg.GetTimestamp(), s.GetTimestamp())
	if !ok {
		consenLogger.WithFields(msg.LogFields()).Debug("executeMsg (Replay Invalid)")
		s.LogMessage("executeMsg", "replayInvalid", msg)
		return
	}
	s.SetString()
	msg.ComputeVMIndex(s)

	// never ignore DBState messages
	if s.IgnoreMissing && msg.Type() != constants.DBSTATE_MSG {
		now := s.GetTimestamp().GetTimeSeconds()
		if now-msg.GetTimestamp().GetTimeSeconds() > 60*15 {
			s.LogMessage("executeMsg", "ignoreMissing", msg)
			return
		}
	}

	valid := msg.Validate(s)
	switch valid {
	case 1:
		// The highest block for which we have received a message.  Sometimes the same as
		if msg.GetResendCnt() == 0 {
			msg.SendOut(s, msg)
		} else if msg.Resend(s) {
			msg.SendOut(s, msg)
		}

		switch msg.Type() {
		case constants.REVEAL_ENTRY_MSG, constants.COMMIT_ENTRY_MSG, constants.COMMIT_CHAIN_MSG:
			if !s.NoEntryYet(msg.GetHash(), nil) {
				delete(s.Holding, msg.GetHash().Fixed())
				s.Commits.Delete(msg.GetHash().Fixed())
				return true
			}
			s.Holding[msg.GetMsgHash().Fixed()] = msg
		}

		var vml int
		if vm == nil || vm.List == nil {
			vml = 0
		} else {
			vml = len(vm.List)
		}
		local := msg.IsLocal()
		vmi := msg.GetVMIndex()
		hkb := s.GetHighestKnownBlock()

		if s.RunLeader &&
			s.Leader &&
			!s.Saving &&
			vm != nil && int(vm.Height) == vml &&
			(!s.Syncing || !vm.Synced) &&
			(local || vmi == s.LeaderVMIndex) &&
			s.LeaderPL.DBHeight+1 >= hkb {
			if vml == 0 {
				s.SendDBSig(s.LLeaderHeight, s.LeaderVMIndex)
				TotalXReviewQueueInputs.Inc()
				s.XReview = append(s.XReview, msg)
			} else {
				s.LogMessage("executeMsg", "LeaderExecute", msg)
				msg.LeaderExecute(s)
			}
		} else {
			s.LogMessage("executeMsg", "FollowerExecute2", msg)
			msg.FollowerExecute(s)
		}

		ret = true

	case 0:
		// Sometimes messages we have already processed are in the msgQueue from holding when we execute them
		// this check makes sure we don't put them back in holding after just deleting them
		if _, valid := s.Replay.Valid(constants.INTERNAL_REPLAY, msg.GetRepeatHash().Fixed(), msg.GetTimestamp(), s.GetTimestamp()); valid {
			TotalHoldingQueueInputs.Inc()
			TotalHoldingQueueRecycles.Inc()
			s.Holding[msg.GetMsgHash().Fixed()] = msg
		} else {
			s.LogMessage("executeMsg", "drop, IReplay", msg)
		}

	default:
		if !msg.SentInvalid() {
			msg.MarkSentInvalid(true)
			s.LogMessage("executeMsg", "InvalidMsg", msg)
			s.networkInvalidMsgQueue <- msg
		}
	}

	executeMsgTime := time.Since(preExecuteMsgTime)
	TotalExecuteMsgTime.Add(float64(executeMsgTime.Nanoseconds()))

	return

}

func (s *State) Process() (progress bool) {

	if s.ResetRequest {
		s.ResetRequest = false
		s.DoReset()
		return false
	}

	// If we are not running the leader, then look to see if we have waited long enough to
	// start running the leader.  If we are, start the clock on Ignoring Missing Messages.  This
	// is so we don't conflict with past version of the network if we have to reboot the network.
	if s.CurrentMinute > 9 {
		s.Leader, s.LeaderVMIndex = s.LeaderPL.GetVirtualServers(9, s.IdentityChainID)
	} else {
		s.Leader, s.LeaderVMIndex = s.LeaderPL.GetVirtualServers(s.CurrentMinute, s.IdentityChainID)
	}

	if !s.RunLeader {
		now := s.GetTimestamp().GetTimeMilli() // Timestamps are in milliseconds, so wait 20
		if now-s.StartDelay > s.StartDelayLimit {
			if s.DBFinished == true {
				s.RunLeader = true
				if !s.IgnoreDone {
					s.StartDelay = now // Reset StartDelay for Ignore Missing
					s.IgnoreDone = true
				}
			}
		}
		s.LeaderPL = s.ProcessLists.Get(s.LLeaderHeight)

	} else if s.IgnoreMissing {
		s.LeaderPL = s.ProcessLists.Get(s.LLeaderHeight)
		now := s.GetTimestamp().GetTimeMilli() // Timestamps are in milliseconds, so wait 20
		if now-s.StartDelay > s.StartDelayLimit {
			s.IgnoreMissing = false
		}
	}

	process := []interfaces.IMsg{}

	var vm *VM
	if s.Leader && s.RunLeader {
		vm = s.LeaderPL.VMs[s.LeaderVMIndex]
		if vm.Height == 0 && s.RunLeader { // Shouldn't send DBSigs out until we have fully loaded our db
			s.SendDBSig(s.LeaderPL.DBHeight, s.LeaderVMIndex)
		}
	}

	/** Process all the DBStates  that might be pending **/

	for {
		ix := int(s.GetHighestSavedBlk()) - s.DBStatesReceivedBase + 1
		if ix < 0 || ix >= len(s.DBStatesReceived) {
			break
		}
		msg := s.DBStatesReceived[ix]
		if msg == nil {
			break
		}
		process = append(process, msg)
		s.DBStatesReceived[ix] = nil
	}

	preAckLoopTime := time.Now()
	// Process acknowledgements if we have some.
ackLoop:
	for {
		select {
		case ack := <-s.ackQueue:
			switch ack.Validate(s) {
			case -1:
				s.LogMessage("ackQueue", "Drop Invalid", ack) // Maybe put it back in the ask queue ? -- clay
				continue
			case 0:
				// toss the ack into holding and we will try again in a bit...
				TotalHoldingQueueInputs.Inc()
				TotalHoldingQueueRecycles.Inc()
				s.Holding[ack.GetMsgHash().Fixed()] = ack
				continue
			}

			if s.IgnoreMissing {
				now := s.GetTimestamp().GetTimeSeconds() //todo: Do we really need to do this every loop?
				if now-ack.GetTimestamp().GetTimeSeconds() < 60*15 {
					s.LogMessage("ackQueue", "Execute", ack)
					s.executeMsg(vm, ack)
					progress = true
				} else {
					s.LogMessage("ackQueue", "Drop Too Old", ack)
				}
				continue
			}

			s.LogMessage("ackQueue", "Execute2", ack)
			progress = s.executeMsg(vm, ack) || progress

		default:
			break ackLoop
		}
	}

	ackLoopTime := time.Since(preAckLoopTime)
	TotalAckLoopTime.Add(float64(ackLoopTime.Nanoseconds()))

	preEmptyLoopTime := time.Now()

	// Process inbound messages
emptyLoop:
	for {
		select {
		case msg := <-s.msgQueue:
			s.LogMessage("msgQueue", "Execute", msg)
			progress = s.executeMsg(vm, msg) || progress
		default:
			break emptyLoop
		}
	}
	emptyLoopTime := time.Since(preEmptyLoopTime)
	TotalEmptyLoopTime.Add(float64(emptyLoopTime.Nanoseconds()))

	preProcessXReviewTime := time.Now()
	// Reprocess any stalled messages, but not so much compared inbound messages
	// Process last first

	if s.RunLeader {
		s.ReviewHolding()
		for {
			for _, msg := range s.XReview {
				if msg == nil {
					continue
				}
				if msg.GetVMIndex() == s.LeaderVMIndex {
					process = append(process, msg)
				}
			}
			s.XReview = s.XReview[:0]
			break
		} // skip review
	}

	processXReviewTime := time.Since(preProcessXReviewTime)
	TotalProcessXReviewTime.Add(float64(processXReviewTime.Nanoseconds()))

	preProcessProcChanTime := time.Now()
	for _, msg := range process {
		newProgress := s.executeMsg(vm, msg)
		progress = newProgress || progress //
		s.LogMessage("executeMsg", fmt.Sprintf("From processq : %t", newProgress), msg)
		s.UpdateState()
	} // processLoop for{...}

	processProcChanTime := time.Since(preProcessProcChanTime)
	TotalProcessProcChanTime.Add(float64(processProcChanTime.Nanoseconds()))

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

	preReviewHoldingTime := time.Now()
	if len(s.XReview) > 0 || s.Syncing || s.Saving {
		return
	}

	now := s.GetTimestamp()
	if s.ResendHolding == nil {
		s.ResendHolding = now
	}
	if now.GetTimeMilli()-s.ResendHolding.GetTimeMilli() < 100 {
		return
	}

	s.Commits.Cleanup(s)
	s.DB.Trim()

	// Set the resend time at the END of the function. This prevents the time it takes to execute this function
	// from reducing the time we allow before another review
	defer func() {
		s.ResendHolding = now
	}()
	// Anything we are holding, we need to reprocess.
	s.XReview = make([]interfaces.IMsg, 0)

	highest := s.GetHighestKnownBlock()
	saved := s.GetHighestSavedBlk()
	for _, a := range s.Acks {
		if s.Holding[a.GetHash().Fixed()] != nil {
			a.FollowerExecute(s)
		}
	}

	// Set this flag, so it acts as a constant.  We will set s.LeaderNewMin to false
	// after processing the Holding Queue.  Ensures we only do this one per minute.
	//	processMinute := s.LeaderNewMin // Have we processed this minute
	s.LeaderNewMin++ // Either way, don't do it again until the ProcessEOM resets LeaderNewMin

	for k, v := range s.Holding {

		if v.Expire(s) {
			s.LogMessage("executeMsg", "expire from holding", v)
			s.ExpireCnt++
			TotalHoldingQueueOutputs.Inc()
			delete(s.Holding, k)
			continue
		}

		switch v.Validate(s) {
		case -1:
			s.LogMessage("executeMsg", "invalid from holding", v)
			TotalHoldingQueueOutputs.Inc()
			delete(s.Holding, k)
			continue
		case 0:
			continue
		}

		if v.GetResendCnt() == 0 {
			v.SendOut(s, v)
		} else {
			if v.Resend(s) {
				v.SendOut(s, v)
			}
		}

		if int(highest)-int(saved) > 1000 {
			TotalHoldingQueueOutputs.Inc()
			delete(s.Holding, k)
		}

		eom, ok := v.(*messages.EOM)
		if ok && ((eom.DBHeight <= saved && saved > 0) || int(eom.Minute) < s.CurrentMinute) {
			TotalHoldingQueueOutputs.Inc()
			delete(s.Holding, k)
			continue
		}

		dbsmsg, ok := v.(*messages.DBStateMsg)
		if ok && (dbsmsg.DirectoryBlock.GetHeader().GetDBHeight() < saved-1 && saved > 0) {
			TotalHoldingQueueOutputs.Inc()
			delete(s.Holding, k)
			continue
		}

		dbsigmsg, ok := v.(*messages.DirectoryBlockSignature)
		if ok && ((dbsigmsg.DBHeight <= saved && saved > 0) || (dbsigmsg.DBHeight < highest-3 && highest > 2)) {
			TotalHoldingQueueOutputs.Inc()
			delete(s.Holding, k)
			continue
		}

		_, ok = s.Replay.Valid(constants.INTERNAL_REPLAY, v.GetRepeatHash().Fixed(), v.GetTimestamp(), s.GetTimestamp())
		ok2 := s.FReplay.IsHashUnique(constants.BLOCK_REPLAY, v.GetRepeatHash().Fixed())
		if !ok || !ok2 {
			TotalHoldingQueueOutputs.Inc()
			delete(s.Holding, k)
			continue
		}

		// If it is an entryCommit/ChainCommit/RevealEntry and it has a duplicate hash to an existing entry throw it away here

		ce, ok := v.(*messages.CommitEntryMsg)
		if ok {
			x := s.NoEntryYet(ce.CommitEntry.EntryHash, ce.CommitEntry.GetTimestamp())
			if !x {
				TotalHoldingQueueOutputs.Inc()
				delete(s.Holding, k) // Drop commits with the same entry hash from holding because they are blocked by a previous entry
				continue
			}
		}

		// If it is an chainCommit and it has a duplicate hash to an existing entry throw it away here

		cc, ok := v.(*messages.CommitChainMsg)
		if ok {
			x := s.NoEntryYet(cc.CommitChain.EntryHash, cc.CommitChain.GetTimestamp())
			if !x {
				TotalHoldingQueueOutputs.Inc()
				delete(s.Holding, k) // Drop commits with the same entry hash from holding because they are blocked by a previous entry
				continue
			}
		}

		// If a Reveal Entry has a commit available, then process the Reveal Entry and send it out.
		if re, ok := v.(*messages.RevealEntryMsg); ok {
			if !s.NoEntryYet(re.GetHash(), s.GetLeaderTimestamp()) {
				delete(s.Holding, re.GetHash().Fixed())
				s.Commits.Delete(re.GetHash().Fixed())
				continue
			}
			// Only reprocess if at the top of a new minute, and if we are a leader.
			//if processMinute > 10 {
			//	continue // No need for followers to review Reveal Entry messages
			//}
			// Needs to be our VMIndex as well, or ignore.
			if re.GetVMIndex() != s.LeaderVMIndex || !s.Leader {
				continue // If we are a leader, but it isn't ours, and it isn't a new minute, ignore.
			}
		}
		//TODO: Move this earlier!
		// We don't reprocess messages if we are a leader, but it ain't ours!
		if s.LeaderVMIndex != v.GetVMIndex() {
			continue
		}

		TotalXReviewQueueInputs.Inc()
		s.XReview = append(s.XReview, v)
		TotalHoldingQueueOutputs.Inc()
	}
	reviewHoldingTime := time.Since(preReviewHoldingTime)
	TotalReviewHoldingTime.Add(float64(reviewHoldingTime.Nanoseconds()))
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
		//s.AddStatus(fmt.Sprintf("AddDBState(): Fail dbstate is nil at dbht: %d", directoryBlock.GetHeader().GetDBHeight()))
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
		//fmt.Println(fmt.Sprintf("SigType PROCESS: %10s Add DBState: s.SigType(%v)", s.FactomNodeName, s.SigType))
		s.EOM = false
		s.DBSig = false
		s.LLeaderHeight = ht
		s.ProcessLists.Get(ht + 1)
		s.CurrentMinute = 0
		s.EOMProcessed = 0
		s.DBSigProcessed = 0
		s.StartDelay = s.GetTimestamp().GetTimeMilli()
		s.RunLeader = false
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
		for s.ProcessLists.UpdateState(s.LLeaderHeight) {
		}
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
	FollowerExecutions.Inc()
	// add it to the holding queue in case AddToProcessList may remove it
	TotalHoldingQueueInputs.Inc()

	s.Holding[m.GetMsgHash().Fixed()] = m
	ack, _ := s.Acks[m.GetMsgHash().Fixed()].(*messages.Ack)

	if ack != nil {
		m.SetLeaderChainID(ack.GetLeaderChainID())
		m.SetMinute(ack.Minute)

		pl := s.ProcessLists.Get(ack.DBHeight)
		pl.AddToProcessList(ack, m)

		// Cross Boot Replay
		s.CrossReplayAddSalt(ack.DBHeight, ack.Salt)
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

	eom, ok := m.(*messages.EOM)
	if !ok {
		return
	}

	if eom.DBHeight == s.ProcessLists.Lists[0].DBHeight && int(eom.Minute) < s.CurrentMinute {
		return
	}

	FollowerEOMExecutions.Inc()
	// add it to the holding queue in case AddToProcessList may remove it
	TotalHoldingQueueInputs.Inc()
	s.Holding[m.GetMsgHash().Fixed()] = m // FollowerExecuteEOM

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

	if ack.DBHeight > s.HighestKnown {
		s.HighestKnown = ack.DBHeight
	}

	pl := s.ProcessLists.Get(ack.DBHeight)
	if pl == nil {
		s.LogMessage("executeMsg", "drop, no pl", msg)
		// Does this mean it's from the future?
		// TODO: Should we put the ack back on the inMsgQueue queue here instead of dropping it? -- clay
		return
	}
	list := pl.VMs[ack.VMIndex].List
	if len(list) > int(ack.Height) && list[ack.Height] != nil {
		// there is already a message in our slot?
		s.LogPrintf("executeMsg", "drop, len(list)(%d) > int(ack.Height)(%d) && list[ack.Height](%p) != nil", len(list), int(ack.Height), list[ack.Height])
		s.LogMessage("executeMsg", "found ", list[ack.Height])
		return
	}

	// We have an ack  and a matching message go execute the message!
	TotalAcksInputs.Inc()
	s.Acks[ack.GetHash().Fixed()] = ack
	m, _ := s.Holding[ack.GetHash().Fixed()]
	if m != nil {
		s.LogMessage("executeMsg", "FollowerExecute3 ", m)
		m.FollowerExecute(s)
	} else {
		s.LogPrintf("executeMsg", "No Msg Holding %x", ack.GetHash().Bytes()[:3])
	}
}

func (s *State) ExecuteEntriesInDBState(dbmsg *messages.DBStateMsg) {
	height := dbmsg.DirectoryBlock.GetDatabaseHeight()

	if s.EntryDBHeightComplete > height {
		return
	}
	// If no Eblocks, leave
	if len(dbmsg.EBlocks) == 0 {
		return
	}

	// All DBStates that got here are valid, so just checking the DBlock hash works
	dblock, err := s.DB.FetchDBlockByHeight(height)
	if err != nil || dblock == nil {
		consenLogger.WithFields(log.Fields{"func": "ExecuteEntriesInDBState", "height": height}).Warnf("Dblock fetched is nil")
		return // This is a weird case
	}

	if !dbmsg.DirectoryBlock.GetHash().IsSameAs(dblock.GetHash()) {
		consenLogger.WithFields(log.Fields{"func": "ExecuteEntriesInDBState", "height": height}).Errorf("Bad DBState. DBlock does not match found")
		return // Bad DBlock
	}

	s.DB.StartMultiBatch()
	for _, e := range dbmsg.Entries {
		if exists, _ := s.DB.DoesKeyExist(databaseOverlay.ENTRY, e.GetHash().Bytes()); !exists {
			s.DB.InsertEntryMultiBatch(e)
		}
	}
	err = s.DB.ExecuteMultiBatch()
	if err != nil {
		consenLogger.WithFields(log.Fields{"func": "ExecuteEntriesInDBState", "height": height}).Errorf("Was unable to execute multibatch")
		return
	}
}

func (s *State) FollowerExecuteDBState(msg interfaces.IMsg) {
	dbstatemsg, _ := msg.(*messages.DBStateMsg)

	cntFail := func() {
		if !dbstatemsg.IsInDB {
			s.DBStateIgnoreCnt++
		}
	}

	// saved := s.GetHighestSavedBlk()

	dbheight := dbstatemsg.DirectoryBlock.GetHeader().GetDBHeight()

	// ignore if too old. If its under EntryDBHeightComplete
	if dbheight > 0 && dbheight <= s.GetHighestSavedBlk() && dbheight < s.EntryDBHeightComplete {
		return
	}

	//s.AddStatus(fmt.Sprintf("FollowerExecuteDBState(): Saved %d dbht: %d", saved, dbheight))

	pdbstate := s.DBStates.Get(int(dbheight - 1))

	switch pdbstate.ValidNext(s, dbstatemsg) {
	case 0:
		//s.AddStatus(fmt.Sprintf("FollowerExecuteDBState(): DBState might be valid %d", dbheight))

		// Don't add duplicate dbstate messages.
		if s.DBStatesReceivedBase < int(s.GetHighestSavedBlk()) {
			cut := int(s.GetHighestSavedBlk()) - s.DBStatesReceivedBase
			if len(s.DBStatesReceived) > cut {
				s.DBStatesReceived = append(make([]*messages.DBStateMsg, 0), s.DBStatesReceived[cut:]...)
			}
			s.DBStatesReceivedBase += cut
		}
		ix := int(dbheight) - s.DBStatesReceivedBase
		if ix < 0 {
			// If we are missing entries at this DBState, we can apply the entries only
			s.ExecuteEntriesInDBState(dbstatemsg)
			return
		}
		for len(s.DBStatesReceived) <= ix {
			s.DBStatesReceived = append(s.DBStatesReceived, nil)
		}
		s.DBStatesReceived[ix] = dbstatemsg
		return
	case -1:
		//s.AddStatus(fmt.Sprintf("FollowerExecuteDBState(): DBState is invalid at ht %d", dbheight))
		// Do nothing because this dbstate looks to be invalid
		cntFail()
		if dbstatemsg.IsLast { // this is the last DBState in this load
			s.DBFinished = true // Just in case we toss the last one for some reason
		}
		return
	}

	if dbstatemsg.IsLast { // this is the last DBState in this load
		s.DBFinished = true // Normal case
		// Attempted hack to fix a set where one leader was ahead of the others.
		//if s.Leader {
		//	dbstatemsg.SetLocal(false) // we are going to send it out to catch everyone up
		//	dbstatemsg.SetPeer2Peer(false)
		//	dbstatemsg.SetFullBroadcast(true)
		//	dbstatemsg.SendOut(s, dbstatemsg)
		//}
	}
	/**************************
	for int(s.ProcessLists.DBHeightBase)+len(s.ProcessLists.Lists) > int(dbheight+1) {
		s.ProcessLists.Lists[len(s.ProcessLists.Lists)-1].Clear()
		s.ProcessLists.Lists = s.ProcessLists.Lists[:len(s.ProcessLists.Lists)-1]
	}
	***************************/
	if dbheight > 1 && dbheight >= s.ProcessLists.DBHeightBase {
		dbs := s.DBStates.Get(int(dbheight))
		if pdbstate.SaveStruct != nil {
			//s.AddStatus(fmt.Sprintf("FollowerExecuteDBState(): Reset to previous state before applying at ht %d", dbheight))
			pdbstate.SaveStruct.TrimBack(s, dbs)
		}
	}

	dbstate := s.AddDBState(false,
		dbstatemsg.DirectoryBlock,
		dbstatemsg.AdminBlock,
		dbstatemsg.FactoidBlock,
		dbstatemsg.EntryCreditBlock,
		dbstatemsg.EBlocks,
		dbstatemsg.Entries)
	if dbstate == nil {
		//s.AddStatus(fmt.Sprintf("FollowerExecuteDBState(): dbstate fail at ht %d", dbheight))
		cntFail()
		return
	}

	// Check all the transaction IDs (do not include signatures).  Only check, don't set flags.
	for i, fct := range dbstatemsg.FactoidBlock.GetTransactions() {
		// Check the prior blocks for a replay.
		_, valid := s.FReplay.Valid(
			constants.BLOCK_REPLAY,
			fct.GetSigHash().Fixed(),
			fct.GetTimestamp(),
			dbstatemsg.DirectoryBlock.GetHeader().GetTimestamp())
		// If not the coinbase TX, and we are past 100,000, and the TX is not valid,then we don't accept this block.
		if i > 0 && // Don't test the coinbase TX
			((dbheight > 0 && dbheight < 2000) || dbheight > 100000) && // Test the first 2000 blks, so we can unit test, then after
			!valid { // 100K for the running system.  If a TX isn't valid, ignore.
			return //Totally ignore the block if it has a double spend.
		}
	}

	for _, ebs := range dbstatemsg.EBlocks {
		blktime := dbstatemsg.DirectoryBlock.GetTimestamp()
		for _, e := range ebs.GetEntryHashes() {
			if e.IsMinuteMarker() {
				continue
			}
			s.FReplay.IsTSValidAndUpdateState(
				constants.BLOCK_REPLAY,
				e.Fixed(),
				blktime,
				blktime)
			s.Replay.IsTSValidAndUpdateState(
				constants.INTERNAL_REPLAY,
				e.Fixed(),
				blktime,
				blktime)

		}
	}

	// Only set the flag if we know the whole block is valid.  We know it is because we checked them all in the loop
	// above
	for _, fct := range dbstatemsg.FactoidBlock.GetTransactions() {
		s.FReplay.IsTSValidAndUpdateState(
			constants.BLOCK_REPLAY,
			fct.GetSigHash().Fixed(),
			fct.GetTimestamp(),
			dbstatemsg.DirectoryBlock.GetHeader().GetTimestamp())
	}

	if dbstatemsg.IsInDB == false {
		//s.AddStatus(fmt.Sprintf("FollowerExecuteDBState(): dbstate added from network at ht %d", dbheight))
		dbstate.ReadyToSave = true
		dbstate.Locked = false
		dbstate.Signed = true
		s.DBStateAppliedCnt++
		s.DBStates.UpdateState()
	} else {
		//s.AddStatus(fmt.Sprintf("FollowerExecuteDBState(): dbstate added from local db at ht %d", dbheight))
		dbstate.Saved = true
		dbstate.IsNew = false
		dbstate.Locked = false
	}

	//fmt.Println(fmt.Sprintf("SigType PROCESS: %10s Clear SigType follower execute DBState:  !s.SigType(%v)", s.FactomNodeName, s.SigType))
	s.EOM = false
	s.EOMDone = false
	s.EOMSys = false
	s.DBSig = false
	s.DBSigDone = false
	s.DBSigSys = false
	s.Saving = true
	s.Syncing = false

	// Hurry up our next ask.  When we get to where we have the data we asked for, then go ahead and ask for the next set.
	if s.DBStates.LastEnd < int(dbheight) {
		s.DBStates.Catchup(true)
	}
	if s.DBStates.LastBegin < int(dbheight)+1 {
		s.DBStates.LastBegin = int(dbheight)
	}
	s.DBStates.TimeToAsk = nil

	if dbstatemsg.IsLocal() {
		if s.StateSaverStruct.FastBoot {
			dbstate.SaveStruct = SaveFactomdState(s, dbstate)

			err := s.StateSaverStruct.SaveDBStateList(s.DBStates, s.Network)
			if err != nil {
				panic(err)
			}
		}
	}
}

func (s *State) FollowerExecuteMMR(m interfaces.IMsg) {

	if s.inMsgQueue.Length() > constants.INMSGQUEUE_HIGH {
		s.LogMessage("executeMsg", "Drop INMSGQUEUE_HIGH", m)
		return
	}
	// Just ignore missing messages for a period after going off line or starting up.

	if s.IgnoreMissing {
		s.LogMessage("executeMsg", "Drop IgnoreMissing", m)
		return
	}
	// Drop the missing message response if it's already in the process list
	_, valid := s.Replay.Valid(constants.INTERNAL_REPLAY, m.GetRepeatHash().Fixed(), m.GetTimestamp(), s.GetTimestamp())
	if !valid {
		s.LogMessage("executeMsg", "drop, INTERNAL_REPLAY", m)
		return
	}

	mmr, _ := m.(*messages.MissingMsgResponse)

	ack, ok := mmr.AckResponse.(*messages.Ack)

	// If we don't need this message, we don't have to do everything else.
	if !ok {
		s.LogMessage("executeMsg", "Drop no ack", m)
		return
	}

	// If we don't need this message, we don't have to do everything else.
	if ack.Validate(s) == -1 {
		s.LogMessage("executeMsg", "Drop ack invalid", m)
		return
	}
	ack.Response = true
	msg := mmr.MsgResponse

	if msg == nil {
		s.LogMessage("executeMsg", "Drop nil message", m)
		return
	}

	pl := s.ProcessLists.Get(ack.DBHeight)

	if pl == nil {
		s.LogMessage("executeMsg", "Drop No Processlist", m)
		return
	}
	_, okm := s.Replay.Valid(constants.INTERNAL_REPLAY, msg.GetRepeatHash().Fixed(), msg.GetTimestamp(), s.GetTimestamp())

	TotalAcksInputs.Inc()

	if okm {
		s.LogMessage("executeMsg", "FollowerExecute3", msg)
		msg.FollowerExecute(s)

		s.LogMessage("executeMsg", "FollowerExecute4", ack)
		ack.FollowerExecute(s)

		s.MissingResponseAppliedCnt++

	}
}

func (s *State) FollowerExecuteDataResponse(m interfaces.IMsg) {
	msg, ok := m.(*messages.DataResponse)
	if !ok {
		return
	}

	switch msg.DataType {
	case 1: // Data is an entryBlock
		eblock, ok := msg.DataObject.(interfaces.IEntryBlock)
		if !ok {
			return
		}

		ebKeyMR, _ := eblock.KeyMR()
		if ebKeyMR == nil {
			return
		}

		for i, missing := range s.MissingEntryBlocks {
			eb := missing.EBHash
			if !eb.IsSameAs(ebKeyMR) {
				continue
			}

			db, err := s.DB.FetchDBlockByHeight(eblock.GetHeader().GetDBHeight())
			if err != nil || db == nil {
				return
			}

			var missing []MissingEntryBlock
			missing = append(missing, s.MissingEntryBlocks[:i]...)
			missing = append(missing, s.MissingEntryBlocks[i+1:]...)
			s.MissingEntryBlocks = missing

			s.DB.ProcessEBlockBatch(eblock, true)

			break
		}

	case 0: // Data is an entry
		entry, ok := msg.DataObject.(interfaces.IEBEntry)
		if !ok {
			return
		}
		if len(s.WriteEntry) < cap(s.WriteEntry) {
			s.WriteEntry <- entry
		}
	}
}

func (s *State) FollowerExecuteMissingMsg(msg interfaces.IMsg) {
	// Don't respond to missing messages if we are behind.
	if s.inMsgQueue.Length() > constants.INMSGQUEUE_LOW {
		return
	}

	m := msg.(*messages.MissingMsg)

	pl := s.ProcessLists.Get(m.DBHeight)

	if pl == nil {
		s.MissingRequestIgnoreCnt++
		return
	}
	FollowerMissingMsgExecutions.Inc()
	sent := false
	if len(pl.System.List) > int(m.SystemHeight) && pl.System.List[m.SystemHeight] != nil {
		msgResponse := messages.NewMissingMsgResponse(s, pl.System.List[m.SystemHeight], nil)
		msgResponse.SetOrigin(m.GetOrigin())
		msgResponse.SetNetworkOrigin(m.GetNetworkOrigin())
		s.NetworkOutMsgQueue().Enqueue(msgResponse)
		s.MissingRequestReplyCnt++
		sent = true
	}

	for _, h := range m.ProcessListHeight {
		missingmsg, ackMsg, err := s.LoadSpecificMsgAndAck(m.DBHeight, m.VMIndex, h)

		if missingmsg != nil && ackMsg != nil && err == nil {
			// If I don't have this message, ignore.
			msgResponse := messages.NewMissingMsgResponse(s, missingmsg, ackMsg)
			msgResponse.SetOrigin(m.GetOrigin())
			msgResponse.SetNetworkOrigin(m.GetNetworkOrigin())
			s.NetworkOutMsgQueue().Enqueue(msgResponse)
			s.MissingRequestReplyCnt++
			sent = true
		}
	}

	if !sent {
		s.MissingRequestIgnoreCnt++
	}
	return
}

func (s *State) FollowerExecuteCommitChain(m interfaces.IMsg) {
	FollowerExecutions.Inc()
	s.FollowerExecuteMsg(m)
	cc := m.(*messages.CommitChainMsg)
	re := s.Holding[cc.CommitChain.EntryHash.Fixed()]
	if re != nil {
		re.FollowerExecute(s)
		re.SendOut(s, re)
	}
	m.SendOut(s, m)
}

func (s *State) FollowerExecuteCommitEntry(m interfaces.IMsg) {
	ce := m.(*messages.CommitEntryMsg)
	FollowerExecutions.Inc()
	s.FollowerExecuteMsg(m)
	re := s.Holding[ce.CommitEntry.EntryHash.Fixed()]
	if re != nil {
		re.FollowerExecute(s)
		re.SendOut(s, re)
	}
	m.SendOut(s, m)
}

func (s *State) FollowerExecuteRevealEntry(m interfaces.IMsg) {
	FollowerExecutions.Inc()
	TotalHoldingQueueInputs.Inc()

	s.Holding[m.GetMsgHash().Fixed()] = m // hold in  FollowerExecuteRevealEntry

	ack, _ := s.Acks[m.GetMsgHash().Fixed()].(*messages.Ack)

	if ack == nil {
		// todo: prevent this log from double logging
		s.LogMessage("executeMsg", "hold, no ack yet", m)
		return
	}

	//ack.SendOut(s, ack)
	m.SetLeaderChainID(ack.GetLeaderChainID())
	m.SetMinute(ack.Minute)
	m.SendOut(s, m)

	pl := s.ProcessLists.Get(ack.DBHeight)
	if pl == nil {
		s.LogMessage("executeMsg", "hold, no process list yet", m)
		return
	}

	// Add the message and ack to the process list.
	pl.AddToProcessList(ack, m)

	// Check to make sure AddToProcessList removed it from holding (added it to the list)
	if s.Holding[m.GetMsgHash().Fixed()] != nil {
		s.LogMessage("executeMsg", "hold, no process list yet2", m)
		return
	}

	msg := m.(*messages.RevealEntryMsg)
	TotalCommitsOutputs.Inc()

	// This is so the api can determine if a chainhead is about to be updated. It fixes a race condition
	// on the api. MUST BE BEFORE THE REPLAY FILTER ADD
	pl.PendingChainHeads.Put(msg.Entry.GetChainID().Fixed(), msg)
	// Okay the Reveal has been recorded.  Record this as an entry that cannot be duplicated.
	s.Replay.IsTSValidAndUpdateState(constants.REVEAL_REPLAY, msg.Entry.GetHash().Fixed(), msg.Timestamp, s.GetLeaderTimestamp())
}

func (s *State) LeaderExecute(m interfaces.IMsg) {
	LeaderExecutions.Inc()
	_, ok := s.Replay.Valid(constants.INTERNAL_REPLAY, m.GetRepeatHash().Fixed(), m.GetTimestamp(), s.GetTimestamp())
	if !ok {
		TotalHoldingQueueOutputs.Inc()
		delete(s.Holding, m.GetMsgHash().Fixed())
		if s.DebugExec() {
			s.LogMessage("executeMsg", "Drop replay", m)
		}
		return
	}

	ack := s.NewAck(m, nil).(*messages.Ack)
	m.SetLeaderChainID(ack.GetLeaderChainID())
	m.SetMinute(ack.Minute)

	s.ProcessLists.Get(ack.DBHeight).AddToProcessList(ack, m)
}

func (s *State) setCurrentMinute(m int) {
	if m != s.CurrentMinute && m != s.CurrentMinute+1 && !(m == 0 && s.CurrentMinute == 10) {
		s.LogPrintf("dbsig-eom", " Jump s.CurrentMinute = %d, from %d %s", m, s.CurrentMinute, atomic.WhereAmIString(1))

	} else {
		if m != s.CurrentMinute {
			s.LogPrintf("dbsig-eom", "s.CurrentMinute = %d from %s", m, atomic.WhereAmIString(1))
		}
	}
	s.CurrentMinute = m
}

func (s *State) LeaderExecuteEOM(m interfaces.IMsg) {
	LeaderEOMExecutions.Inc()
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

	// Put the System Height and Serial Hash into the EOM
	eom.SysHeight = uint32(pl.System.Height)

	if s.Syncing && vm.Synced {
		return
	} else if !s.Syncing {
		s.Syncing = true
		//fmt.Println(fmt.Sprintf("SigType PROCESS: %10s LeaderExecuteEOM: !s.SigType(%v)", s.FactomNodeName, s.SigType))
		s.EOM = true
		s.EOMsyncing = true
		s.EOMProcessed = 0
		for _, vm := range pl.VMs {
			vm.Synced = false
		}
		s.EOMLimit = len(pl.FedServers)
		s.EOMMinute = int(s.CurrentMinute)
	}

	if vm.EomMinuteIssued >= s.CurrentMinute+1 {
		//os.Stderr.WriteString(fmt.Sprintf("Bump detected %s minute %2d\n", s.FactomNodeName, s.CurrentMinute))
		return
	}

	//_, vmindex := pl.GetVirtualServers(s.EOMMinute, s.IdentityChainID)

	if eom.DBHeight != s.LLeaderHeight || eom.VMIndex != s.LeaderVMIndex || eom.Minute != byte(s.CurrentMinute) {
		s.LogPrintf("executeMsg", "EOM has wrong data expected DBH/VM/M %d/%d/%d", s.LLeaderHeight, s.LeaderVMIndex, s.CurrentMinute)
	}
	eom.DBHeight = s.LLeaderHeight
	eom.VMIndex = s.LeaderVMIndex
	// eom.Minute is zerobased, while LeaderMinute is 1 based.  So
	// a simple assignment works.
	eom.Minute = byte(s.CurrentMinute)
	vm.EomMinuteIssued = s.CurrentMinute + 1
	eom.Sign(s)
	eom.MsgHash = nil
	ack := s.NewAck(m, nil).(*messages.Ack)

	TotalAcksInputs.Inc()
	s.Acks[eom.GetMsgHash().Fixed()] = ack
	m.SetLocal(false)
	ack.SendOut(s, ack)
	m.SendOut(s, m)
	s.FollowerExecuteEOM(m)
	s.UpdateState()
	delete(s.Acks, ack.GetHash().Fixed())
	delete(s.Holding, m.GetMsgHash().Fixed())
}

func (s *State) LeaderExecuteDBSig(m interfaces.IMsg) {
	LeaderExecutions.Inc()
	dbs := m.(*messages.DirectoryBlockSignature)
	pl := s.ProcessLists.Get(dbs.DBHeight)

	if dbs.DBHeight != s.LLeaderHeight {
		s.LogMessage("executeMsg", "followerExec", m)
		m.FollowerExecute(s)
		return
	}
	if len(pl.VMs[dbs.VMIndex].List) > 0 {
		s.LogMessage("executeMsg", "drop, slot 0 taken by", pl.VMs[dbs.VMIndex].List[0])
		return
	}

	// Put the System Height and Serial Hash into the EOM
	dbs.SysHeight = uint32(pl.System.Height)

	_, ok := s.Replay.Valid(constants.INTERNAL_REPLAY, m.GetRepeatHash().Fixed(), m.GetTimestamp(), s.GetTimestamp())
	if !ok {
		TotalHoldingQueueOutputs.Inc()
		HoldingQueueDBSigOutputs.Inc()
		delete(s.Holding, m.GetMsgHash().Fixed())
		s.LogMessage("executeMsg", "drop INTERNAL_REPLAY", m)
		return
	}

	ack := s.NewAck(m, s.Balancehash).(*messages.Ack)

	m.SetLeaderChainID(ack.GetLeaderChainID())
	m.SetMinute(ack.Minute)

	s.ProcessLists.Get(ack.DBHeight).AddToProcessList(ack, m)
}

func (s *State) LeaderExecuteCommitChain(m interfaces.IMsg) {
	cc := m.(*messages.CommitChainMsg)
	// Check if this commit has more entry credits than any previous that we have.
	if !s.IsHighestCommit(cc.GetHash(), m) {
		// This commit is not higher than any previous, so we can discard it and prevent a double spend
		return
	}

	s.LeaderExecute(m)

	if re := s.Holding[cc.GetHash().Fixed()]; re != nil {
		re.SendOut(s, re) // If I was waiting on the commit, go ahead and send out the reveal
	}
}

func (s *State) LeaderExecuteCommitEntry(m interfaces.IMsg) {
	ce := m.(*messages.CommitEntryMsg)

	// Check if this commit has more entry credits than any previous that we have.
	if !s.IsHighestCommit(ce.GetHash(), m) {
		// This commit is not higher than any previous, so we can discard it and prevent a double spend
		return
	}

	s.LeaderExecute(m)

	if re := s.Holding[ce.GetHash().Fixed()]; re != nil {
		re.SendOut(s, re) // If I was waiting on the commit, go ahead and send out the reveal
	}
}

func (s *State) LeaderExecuteRevealEntry(m interfaces.IMsg) {
	LeaderExecutions.Inc()
	re := m.(*messages.RevealEntryMsg)
	eh := re.Entry.GetHash()

	ack := s.NewAck(m, nil).(*messages.Ack)

	// Debugging thing.
	m.SetLeaderChainID(ack.GetLeaderChainID())
	m.SetMinute(ack.Minute)

	// Put the acknowledgement in the Acks so we can tell if AddToProcessList() adds it.
	s.Acks[m.GetMsgHash().Fixed()] = ack
	TotalAcksInputs.Inc()
	s.ProcessLists.Get(ack.DBHeight).AddToProcessList(ack, m)

	// If it was not added, then handle as a follower, and leave.
	if s.Acks[m.GetMsgHash().Fixed()] != nil {
		m.FollowerExecute(s)
		return
	}

	// Okay the Reveal has been recorded.  Record this as an entry that cannot be duplicated.
	s.Replay.IsTSValidAndUpdateState(constants.REVEAL_REPLAY, eh.Fixed(), m.GetTimestamp(), s.GetLeaderTimestamp())
	TotalCommitsOutputs.Inc()
}

func (s *State) ProcessAddServer(dbheight uint32, addServerMsg interfaces.IMsg) bool {
	as, ok := addServerMsg.(*messages.AddServerMsg)
	if ok && !ProcessIdentityToAdminBlock(s, as.ServerChainID, as.ServerType) {
		//s.AddStatus(fmt.Sprintf("Failed to add %x as server type %d", as.ServerChainID.Bytes()[2:5], as.ServerType))
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
		return true
	}

	if s.GetAuthorityServerType(rs.ServerChainID) != rs.ServerType {
		return true
	}

	if len(s.LeaderPL.FedServers) < 2 && rs.ServerType == 0 {
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
		return true
	}

	switch ask.AdminBlockChange {
	case constants.TYPE_ADD_BTC_ANCHOR_KEY:
		var btcKey [20]byte
		copy(btcKey[:], ask.Key.Bytes()[:20])
		s.LeaderPL.AdminBlock.AddFederatedServerBitcoinAnchorKey(ask.IdentityChainID, ask.KeyPriority, ask.KeyType, btcKey)
	case constants.TYPE_ADD_FED_SERVER_KEY:
		pub := ask.Key.Fixed()
		s.LeaderPL.AdminBlock.AddFederatedServerSigningKey(ask.IdentityChainID, pub)
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
		// save the Commit to match against the Reveal later
		h := c.GetHash()
		s.PutCommit(h, c)
		entry := s.Holding[h.Fixed()]
		if entry != nil {
			entry.FollowerExecute(s)
			entry.SendOut(s, entry)
			TotalXReviewQueueInputs.Inc()
			s.XReview = append(s.XReview, entry)
			TotalHoldingQueueOutputs.Inc()
		}

		return true
	}
	//s.AddStatus("Cannot process Commit Chain")

	return false
}

func (s *State) ProcessCommitEntry(dbheight uint32, commitEntry interfaces.IMsg) bool {
	c, _ := commitEntry.(*messages.CommitEntryMsg)

	pl := s.ProcessLists.Get(dbheight)
	pl.EntryCreditBlock.GetBody().AddEntry(c.CommitEntry)
	if e := s.GetFactoidState().UpdateECTransaction(true, c.CommitEntry); e == nil {
		// save the Commit to match against the Reveal later
		h := c.GetHash()
		s.PutCommit(h, c)
		entry := s.Holding[h.Fixed()]
		if entry != nil && entry.Validate(s) == 1 {
			entry.FollowerExecute(s)
			entry.SendOut(s, entry)
			TotalXReviewQueueInputs.Inc()
			s.XReview = append(s.XReview, entry)
			TotalHoldingQueueOutputs.Inc()
		}
		return true
	}
	//s.AddStatus("Cannot Process Commit Entry")

	return false
}

func (s *State) ProcessRevealEntry(dbheight uint32, m interfaces.IMsg) bool {
	pl := s.ProcessLists.Get(dbheight)
	if pl == nil {
		return false
	}
	TotalProcessListProcesses.Inc()
	msg := m.(*messages.RevealEntryMsg)
	TotalCommitsOutputs.Inc()
	s.Commits.Delete(msg.Entry.GetHash().Fixed()) // 	delete(s.Commits, msg.Entry.GetHash().Fixed())

	// This is so the api can determine if a chainhead is about to be updated. It fixes a race condition
	// on the api. MUST BE BEFORE THE REPLAY FILTER ADD
	pl.PendingChainHeads.Put(msg.Entry.GetChainID().Fixed(), msg)
	// Okay the Reveal has been recorded.  Record this as an entry that cannot be duplicated.
	s.Replay.IsTSValidAndUpdateState(constants.REVEAL_REPLAY, msg.Entry.GetHash().Fixed(), msg.Timestamp, s.GetTimestamp())
	myhash := msg.Entry.GetHash()

	chainID := msg.Entry.GetChainID()

	TotalCommitsOutputs.Inc()
	s.Commits.Delete(msg.Entry.GetHash().Fixed()) // delete(s.Commits, msg.Entry.GetHash().Fixed())

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
			//s.AddStatus("Failed to add to process Reveal Entry because no Entry Block found")
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

	s.IncEntries()
	return true
}

func (s *State) CreateDBSig(dbheight uint32, vmIndex int) (interfaces.IMsg, interfaces.IMsg) {
	dbstate := s.DBStates.Get(int(dbheight - 1))
	if dbstate == nil && dbheight > 0 {
		s.LogPrintf("executeMsg", "Can not create DBSig because %d because there is no dbstate", dbheight)
		return nil, nil
	}
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
	ack := s.NewAck(dbs, s.Balancehash).(*messages.Ack)
	return dbs, ack
}

// dbheight is the height of the process list, and vmIndex is the vm
// that is missing the DBSig.  If the DBSig isn't our responsibility, then
// this call will do nothing.  Assumes the state for the leader is set properly
func (s *State) SendDBSig(dbheight uint32, vmIndex int) {
	dbslog := consenLogger.WithFields(log.Fields{"func": "SendDBSig"})

	ht := s.GetHighestSavedBlk()
	if dbheight <= ht { // if it's in the past, just return.
		return
	}
	if s.EOM { // If we are counting up EOMs don't generate a DBSig .. why ? -- clay
		return
	}
	pl := s.ProcessLists.Get(dbheight)
	vm := pl.VMs[vmIndex]
	if vm.Height > 0 {
		return // If we already have the DBSIG (it's always in slot 0) then just return
	}
	leader, lvm := pl.GetVirtualServers(vm.LeaderMinute, s.IdentityChainID)
	if !leader || lvm != vmIndex {
		return // If I'm not a leader or this is not my VM then return
	}

	if !vm.Signed {

		if !pl.DBSigAlreadySent {

			dbs, _ := s.CreateDBSig(dbheight, vmIndex)
			if dbs == nil {
				return
			}

			dbslog.WithFields(dbs.LogFields()).WithFields(log.Fields{"lheight": s.GetLeaderHeight(), "node-Name": s.GetFactomNodeName()}).Infof("Generate DBSig")
			dbs.LeaderExecute(s)
			vm.Signed = true
			pl.DBSigAlreadySent = true
		}
		// used to ask here for the message we already made and sent...
	}
}

// TODO: Should fault the server if we don't have the proper sequence of EOM messages.
func (s *State) ProcessEOM(dbheight uint32, msg interfaces.IMsg) bool {
	TotalProcessEOMs.Inc()
	e := msg.(*messages.EOM)
	// plog := consenLogger.WithFields(log.Fields{"func": "ProcessEOM", "msgheight": e.DBHeight, "lheight": s.GetLeaderHeight(), "min", e.Minute})
	pl := s.ProcessLists.Get(dbheight)
	vmIndex := msg.GetVMIndex()
	vm := pl.VMs[vmIndex]

	s.LogPrintf("dbsig-eom", "ProcessEOM@%d/%d/%d minute %d, Syncing %v , EOM %v, EOMDone %v, EOMProcessed %v, EOMLimit %v DBSigDone %v",
		dbheight, msg.GetVMIndex(), len(vm.List), s.CurrentMinute, s.Syncing, s.EOM, s.EOMDone, s.EOMProcessed, s.EOMLimit, s.DBSigDone)

	// debug
	if s.DebugExec() {
		if s.Syncing && s.EOM && !s.EOMDone && s.DBSigDone {
			ids := s.GetUnSyncedServers(dbheight)
			if len(ids) > 0 {
				s.LogPrintf("dbsig-eom", "Waiting for EOMs from %s", ids)
			}
		}
	}

	// Check a bunch of reason not to handle an EOM

	if s.Syncing && s.DBSig { // this means we are syncing DBSigs
		//fmt.Println(fmt.Sprintf("SigType PROCESS: %10s vm %2d Will Not Process: return on s.Syncing(%v) && !s.SigType(%v)", s.FactomNodeName, e.VMIndex, s.Syncing, s.SigType))
		s.LogPrintf("dbsig-eom", "ProcessEOM skip wait for DBSigs to be done")
		return false
	}

	if s.EOM && e.DBHeight != dbheight { // EOM for the wrong dbheight
		//fmt.Println(fmt.Sprintf("SigType PROCESS: %10s vm %2d Invalid SigType s.SigType(%v) && e.DBHeight(%v) != dbheight(%v)", s.FactomNodeName, e.VMIndex, s.SigType, e.DBHeight, dbheight))
		s.LogPrintf("dbsig-eom", "ProcessEOM Found EOM for a different height e.DBHeight(%d) != dbheight(%d) ", e.DBHeight, dbheight)
		// Really we are just going to process this?
	}

	if s.EOM && int(e.Minute) > s.EOMMinute {
		//fmt.Println(fmt.Sprintf("SigType PROCESS: %10s vm %2d Will Not Process: return on s.SigType(%v) && int(e.Minute(%v)) > s.EOMMinute(%v)", s.FactomNodeName, e.VMIndex, s.SigType, e.Minute, s.EOMMinute))
		s.LogPrintf("dbsig-eom", "ProcessEOM skip EOM for a future minute e.Minute(%d) > s.EOMMinute(%d)", e.Minute, s.EOMMinute)
		return false
	}

	if s.CurrentMinute == 0 && !s.DBSigDone {
		s.LogPrintf("dbsig-eom", "ProcessEOM wait for DBSIg in minute 0")
		return false
	}

	if uint32(pl.System.Height) >= e.SysHeight {
		s.EOMSys = true
	}

	s.LogMessage("dbsig-eom", "ProcessEOM ", msg)

	// If I have done everything for all EOMs for all VMs, then and only then do I
	// let processing continue.
	if s.EOMDone && s.EOMSys {
		s.LogPrintf("dbsig-eom", "ProcessEOM finalize EOM processing")

		dbstate := s.GetDBState(dbheight - 1)
		if dbstate == nil {
			//fmt.Println(fmt.Sprintf("SigType PROCESS: %10s vm %2d DBState == nil: return on s.SigType(%v) && int(e.Minute(%v)) > s.EOMMinute(%v)", s.FactomNodeName, e.VMIndex, s.SigType, e.Minute, s.EOMMinute))
			s.LogPrintf("dbsig-eom", "ProcessEOM wait prev dbstate == nil")
			return false
		}
		if !dbstate.Saved {
			//fmt.Println(fmt.Sprintf("SigType PROCESS: %10s vm %2d DBState not saved: return on s.SigType(%v) && int(e.Minute(%v)) > s.EOMMinute(%v)", s.FactomNodeName, e.VMIndex, s.SigType, e.Minute, s.EOMMinute))
			s.LogPrintf("dbsig-eom", "ProcessEOM wait prev !dbstate.Saved")
			return false
		}

		//fmt.Println(fmt.Sprintf("EOM PROCESS: %10s vm %2d Done! s.EOMDone(%v) && s.EOMSys(%v)", s.FactomNodeName, e.VMIndex, s.EOMDone, s.EOMSys))
		s.EOMProcessed--
		if s.EOMProcessed <= 0 { // why less than or equal?
			s.EOM = false
			s.EOMDone = false
			s.Syncing = false
			s.EOMProcessed = 0
			s.SendHeartBeat() // Only do this once
			s.LogPrintf("dbsig-eom", "ProcessEOM complete for %d", e.Minute)
		}
		return true
	}

	// What I do once  for all VMs at the beginning of processing a particular EOM
	if !s.EOM {
		s.LogPrintf("dbsig-eom", "ProcessEOM start EOM processing for %d", e.Minute)

		//fmt.Println(fmt.Sprintf("SigType PROCESS: %10s vm %2d Start SigType Processing: !s.SigType(%v) SigType: %s", s.FactomNodeName, e.VMIndex, s.SigType, e.String()))
		s.EOMSys = false
		s.Syncing = true
		s.EOM = true
		s.EOMLimit = len(s.LeaderPL.FedServers)
		if s.CurrentMinute != int(e.Minute) {
			s.LogPrintf("dbsig-eom", "Follower jump to minute %d from %d", s.CurrentMinute, int(e.Minute))
		}
		s.EOMMinute = int(e.Minute)
		s.EOMsyncing = true
		s.EOMProcessed = 0

		for _, vm := range pl.VMs {
			vm.Synced = false
		}
		//fmt.Println(fmt.Sprintf("SigType PROCESS: %10s vm  %2d First SigType processed: return on s.SigType(%v) && int(e.Minute(%v)) > s.EOMMinute(%v)", s.FactomNodeName, e.VMIndex, s.SigType, e.Minute, s.EOMMinute))
		return false
	}

	// What I do for each EOM
	if !vm.Synced {
		s.LogPrintf("dbsig-eom", "ProcessEOM Handle VM(%v) minute %d", msg.GetVMIndex(), e.Minute)

		InMsg := s.EFactory.NewEomSigInternal(
			s.FactomNodeName,
			e.DBHeight,
			uint32(e.Minute),
			msg.GetVMIndex(),
			uint32(vm.Height),
			e.ChainID,
		)
		s.electionsQueue.Enqueue(InMsg)

		//fmt.Println(fmt.Sprintf("SigType PROCESS: %10s vm %2d Process Once: !e.Processed(%v) SigType: %s", s.FactomNodeName, e.VMIndex, e.Processed, e.String()))
		vm.LeaderMinute++
		s.EOMProcessed++
		//fmt.Println(fmt.Sprintf("EOM PROCESS: %10s vm %2d EOMProcessed++ (%2d)", s.FactomNodeName, e.VMIndex, s.EOMProcessed))
		vm.Synced = true
		markNoFault(pl, msg.GetVMIndex())
		if s.LeaderPL.SysHighest < int(e.SysHeight) {
			s.LeaderPL.SysHighest = int(e.SysHeight)
		}
		//fmt.Println(fmt.Sprintf("SigType PROCESS: %10s vm %2d Process this SigType: return on s.SigType(%v) && int(e.Minute(%v)) > s.EOMMinute(%v)", s.FactomNodeName, e.VMIndex, s.SigType, e.Minute, s.EOMMinute))
		return false
	}

	allfaults := s.LeaderPL.System.Height >= s.LeaderPL.SysHighest

	// After all EOM markers are processed, Claim we are done.  Now we can unwind

	if allfaults && s.EOMProcessed == s.EOMLimit && !s.EOMDone {
		s.LogPrintf("dbsig-eom", "ProcessEOM stop EOM processing minute %d", s.CurrentMinute)

		//fmt.Println(fmt.Sprintf("SigType PROCESS: SigType Complete: %10s vm %2d allfaults(%v) && s.EOMProcessed(%v) == s.EOMLimit(%v) && !s.EOMDone(%v)",
		//	s.FactomNodeName,
		//	e.VMIndex, allfaults, s.EOMProcessed, s.EOMLimit, s.EOMDone))

		s.EOMDone = true
		s.LeaderNewMin = 0
		for _, eb := range pl.NewEBlocks {
			eb.AddEndOfMinuteMarker(byte(e.Minute + 1))
		}

		s.FactoidState.EndOfPeriod(int(e.Minute))

		ecblk := pl.EntryCreditBlock
		ecbody := ecblk.GetBody()
		mn := entryCreditBlock.NewMinuteNumber(e.Minute + 1)
		ecbody.AddEntry(mn)

		if !s.Leader {
			if s.CurrentMinute != int(e.Minute) {
				s.LogPrintf("dbsig-eom", "Follower jump to minute %d from %d", s.CurrentMinute, int(e.Minute))
			}
			s.setCurrentMinute(int(e.Minute))
		}

		s.setCurrentMinute(s.CurrentMinute + 1)
		s.CurrentMinuteStartTime = time.Now().UnixNano()
		// If an election took place, our lists will be unsorted. Fix that
		pl.SortAuditServers()
		pl.SortFedServers()

		switch {
		case s.CurrentMinute < 10:
			if s.CurrentMinute == 1 {
				dbstate := s.GetDBState(dbheight - 1)
				// Panic had arose when leaders would reboot and the follower was on a future minute
				if dbstate == nil {
					// We recognize that this will leave us "Done" without finishing the process.  But
					// a Follower can heal themselves by asking for a block, and overwriting this block.
					return false
				}
				if !dbstate.Saved {
					dbstate.ReadyToSave = true
				}
			}
			s.LeaderPL = s.ProcessLists.Get(s.LLeaderHeight)
			s.Leader, s.LeaderVMIndex = s.LeaderPL.GetVirtualServers(s.CurrentMinute, s.IdentityChainID)

		case s.CurrentMinute == 10:
			s.LogPrintf("dbsig-eom", "Start new block")
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

			s.setCurrentMinute(0)
			s.LLeaderHeight++

			s.GetAckChange()
			s.CheckForIDChange()

			s.LeaderPL = s.ProcessLists.Get(s.LLeaderHeight)
			s.Leader, s.LeaderVMIndex = s.LeaderPL.GetVirtualServers(0, s.IdentityChainID)

			s.DBSigProcessed = 0
			s.TempBalanceHash = s.FactoidState.GetBalanceHash(true)

			// Note about dbsigs.... If we processed the previous minute, then we generate the DBSig for the next block.
			// But if we didn't process the previous block, like we start from scratch, or we had to reset the entire
			// network, then no dbsig exists.  This code doesn't execute, and so we have no dbsig.  In that case, on
			// the next EOM, we see the block hasn't been signed, and we sign the block (That is the call to SendDBSig()
			// above).
			pldbs := s.ProcessLists.Get(s.LLeaderHeight)
			if s.Leader && !pldbs.DBSigAlreadySent {
				// dbstate is already set.
				dbs := new(messages.DirectoryBlockSignature)
				db := dbstate.DirectoryBlock
				dbs.DirectoryBlockHeader = db.GetHeader()
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
				pldbs.DBSigAlreadySent = true

				dbslog := consenLogger.WithFields(log.Fields{"func": "SendDBSig", "lheight": s.GetLeaderHeight(), "node-name": s.GetFactomNodeName()}).WithFields(dbs.LogFields())
				dbslog.Infof("Generate DBSig")

				s.LogMessage("executeMsg", "LeaderExec2", dbs)
				dbs.LeaderExecute(s)
			}
			s.Saving = true
		}

		s.Commits.RemoveExpired(s)

		for k := range s.Acks {
			v := s.Acks[k].(*messages.Ack)
			if v.DBHeight < s.LLeaderHeight {
				TotalAcksOutputs.Inc()
				delete(s.Acks, k)
			}
		}
		//fmt.Println(fmt.Sprintf("SigType PROCESS: %10s vm %2d Saving: return on s.SigType(%v) && int(e.Minute(%v)) > s.EOMMinute(%v)", s.FactomNodeName, e.VMIndex, s.SigType, e.Minute, s.EOMMinute))
	} else {
		//fmt.Println(fmt.Sprintf("SigType PROCESS: %10s vm %2d Do nothing: return on s.SigType(%v) && int(e.Minute(%v)) > s.EOMMinute(%v)", s.FactomNodeName, e.VMIndex, s.SigType, e.Minute, s.EOMMinute))
	}

	return false
}

// Return a string with the short IDs for all unsynced VMs
func (s *State) GetUnSyncedServers(dbheight uint32) string {
	var ids string
	p := s.ProcessLists.Get(dbheight)
	for index, l := range s.GetFedServers(dbheight) {
		vmIndex := p.ServerMap[s.CurrentMinute][index]
		vm := p.VMs[vmIndex]
		if !vm.Synced {
			ids = ids + "," + l.GetChainID().String()[6:12]
		}
	}
	if len(ids) > 0 {
		ids = ids[1:] // drop the leading comma
	}
	return ids
}

func (s *State) CheckForIDChange() {
	var reloadIdentity bool = false
	if s.AckChange > 0 {
		if s.LLeaderHeight >= s.AckChange {
			reloadIdentity = true
		}
	}
	if reloadIdentity {
		config := util.ReadConfig(s.ConfigFilePath)
		var err error
		s.IdentityChainID, err = primitives.NewShaHashFromStr(config.App.IdentityChainID)
		if err != nil {
			panic(err)
		}
		s.LocalServerPrivKey = config.App.LocalServerPrivKey
		s.initServerKeys()
	}
}

// When we process the directory Signature, and we are the leader for said signature, it
// is then that we push it out to the rest of the network.  Otherwise, if we are not the
// leader for the signature, it marks the sig complete for that list
func (s *State) ProcessDBSig(dbheight uint32, msg interfaces.IMsg) bool {
	//fmt.Println(fmt.Sprintf("ProcessDBSig: %10s %s ", s.FactomNodeName, msg.String()))

	dbs := msg.(*messages.DirectoryBlockSignature)
	//plog makes logging anything in ProcessDBSig() easier
	//		The instantiation as a function makes it almost no overhead if you do not use it
	plog := func(format string, args ...interface{}) {
		consenLogger.WithFields(log.Fields{"func": "ProcessDBSig", "msgheight": dbs.DBHeight, "lheight": s.GetLeaderHeight(), "msg": msg.String()}).Errorf(format, args...)
	}
	// debug
	if s.DebugExec() {
		var ids string
		if s.Syncing && s.DBSig && !s.DBSigDone {
			p := s.ProcessLists.Get(dbheight - 1)
			for i, l := range p.FedServers {
				vm := p.VMs[i]
				if !vm.Synced {
					ids = ids + "," + l.GetChainID().String()[6:12]
				}
			}
			if len(ids) > 0 {
				s.LogPrintf("dbsig-eom", "Waiting for DBSIGs from %s", ids[1:])
			}
		}
	}
	// Don't process if syncing an EOM
	if s.Syncing && !s.DBSig {
		//fmt.Println(fmt.Sprintf("ProcessDBSig(): %10s Will Not Process: dbht: %d return on s.Syncing(%v) && !s.DBSig(%v)", s.FactomNodeName,
		//	dbs.DBHeight, s.Syncing, s.DBSig))
		s.LogPrintf("dbsig-eom", "ProcessDBSig skip wait for EOMs to be done")
		return false
	}

	pl := s.ProcessLists.Get(dbheight)
	vm := s.ProcessLists.Get(dbheight).VMs[msg.GetVMIndex()]

	// debug
	s.LogPrintf("dbsig-eom", "ProcessDBSig@%d/%d/%d minute %d, Syncing %v , DBSID %v, DBSigDone %v, DBSigProcessed %v, DBSigLimit %v DBSigDone %v",
		dbheight, msg.GetVMIndex(), len(vm.List), s.CurrentMinute, s.Syncing, s.DBSig, s.DBSigDone, s.DBSigProcessed, s.DBSigLimit, s.DBSigDone)

	// debug
	if s.DebugExec() {
		if s.Syncing && s.DBSig && !s.DBSigDone {
			ids := s.GetUnSyncedServers(dbheight)
			if len(ids) > 0 {
				s.LogPrintf("dbsig-eom", "Waiting for DBSigs from %s", ids)
			}
		}
	}
	if uint32(pl.System.Height) >= dbs.SysHeight {
		s.DBSigSys = true
	}

	s.LogMessage("dbsig-eom", "ProcessDBSig ", msg)
	// If we are done with DBSigs, and this message is processed, then we are done.  Let everything go!
	if s.DBSigSys && s.DBSig && s.DBSigDone {
		s.LogPrintf("dbsig-eom", "ProcessDBSig finalize DBSig processing")
		//fmt.Println(fmt.Sprintf("ProcessDBSig(): %10s Finished with DBSig: s.DBSigSys(%v) && s.DBSig(%v) && s.DBSigDone(%v)", s.FactomNodeName, s.DBSigSys, s.DBSig, s.DBSigDone))
		s.DBSigProcessed--
		if s.DBSigProcessed <= 0 {
			s.EOMDone = false
			s.EOMSys = false
			s.EOM = false
			s.DBSig = false
			s.Syncing = false
			s.LogPrintf("dbsig-eom", "ProcessDBSig complete for %d", dbs.Minute)
		}
		vm.Signed = true
		//s.LeaderPL.AdminBlock
		return true
	}

	// Put the stuff that only executes once at the start of DBSignatures here
	if !s.DBSig {
		s.LogPrintf("dbsig-eom", "ProcessDBSig start DBSig processing for %d", dbs.Minute)

		//fmt.Printf("ProcessDBSig(): %s Start DBSig %s\n", s.FactomNodeName, dbs.String())
		s.DBSigLimit = len(pl.FedServers)
		s.DBSigProcessed = 0
		s.DBSig = true
		s.Syncing = true
		s.DBSigDone = false // p
		//		s.LogPrintf("dbsig-eom", "DBSIGDone written %v @ %s", s.DBSigDone, atomic.WhereAmIString(0))
		for _, vm := range pl.VMs {
			vm.Synced = false
		}
		pl.ResetDiffSigTally()
	}

	// Put the stuff that executes per DBSignature here
	if !vm.Synced {
		s.LogPrintf("dbsig-eom", "ProcessDBSig Handle VM(%v) minute %d", msg.GetVMIndex(), dbs.Minute)

		if s.LLeaderHeight > 0 && s.GetHighestCompletedBlk()+1 < s.LLeaderHeight {

			pl := s.ProcessLists.Get(dbs.DBHeight - 1)
			if !pl.Complete() {
				dbstate := s.DBStates.Get(int(dbs.DBHeight - 1))
				if dbstate == nil || (!dbstate.Locked && !dbstate.Saved) {
					db, _ := s.DB.FetchDBlockByHeight(dbs.DBHeight - 1)
					if db == nil {
						//fmt.Printf("ProcessDBSig(): %10s Previous Process List isn't complete. %s\n", s.FactomNodeName, dbs.String())
						return false
					}
				}
			}
		}

		//fmt.Println(fmt.Sprintf("ProcessDBSig(): %10s Process the %d DBSig: %v", s.FactomNodeName, s.DBSigProcessed, dbs.String()))
		if dbs.VMIndex == 0 {
			//fmt.Println(fmt.Sprintf("ProcessDBSig(): %10s Set Leader Timestamp to: %v %d", s.FactomNodeName, dbs.GetTimestamp().String(), dbs.GetTimestamp().GetTimeMilli()))
			s.SetLeaderTimestamp(dbs.GetTimestamp())
		}

		dblk, err := s.DB.FetchDBlockByHeight(dbheight - 1)
		if err != nil || dblk == nil {
			dbstate := s.GetDBState(dbheight - 1)
			if dbstate == nil || !(!dbstate.IsNew || dbstate.Locked || dbstate.Saved) {
				//fmt.Println(fmt.Sprintf("ProcessingDBSig(): %10s The prior dbsig %d is nil", s.FactomNodeName, dbheight-1))
				return false
			}
			dblk = dbstate.DirectoryBlock
		}

		if dbs.DirectoryBlockHeader.GetBodyMR().Fixed() != dblk.GetHeader().GetBodyMR().Fixed() {
			pl.IncrementDiffSigTally()
			plog("Failed. DBlocks do not match Expected-Body-Mr: %x, Got: %x",
				dblk.GetHeader().GetBodyMR().Fixed(), dbs.DirectoryBlockHeader.GetBodyMR().Fixed())
			return false
		}

		// Adds DB Sig to be added to Admin block if passes sig checks
		data, err := dbs.DirectoryBlockHeader.MarshalBinary()
		if err != nil {
			return false
		}
		if !dbs.DBSignature.Verify(data) {
			return false
		}

		valid, err := s.FastVerifyAuthoritySignature(data, dbs.DBSignature, dbs.DBHeight)
		if err != nil || valid != 1 {
			s.LogPrintf("executeMsg", "Failed. Invalid Auth Sig: Pubkey: %x", dbs.Signature.GetKey())
			return false
		}

		dbs.Matches = true
		s.AddDBSig(dbheight, dbs.ServerIdentityChainID, dbs.DBSignature)

		s.DBSigProcessed++
		//fmt.Println(fmt.Sprintf("Process DBSig %10s vm %2v DBSigProcessed++ (%2d)", s.FactomNodeName, dbs.VMIndex, s.DBSigProcessed))
		vm.Synced = true

		InMsg := s.EFactory.NewDBSigSigInternal(
			s.FactomNodeName,
			dbs.DBHeight,
			uint32(0),
			msg.GetVMIndex(),
			uint32(vm.Height),
			dbs.LeaderChainID,
		)
		s.electionsQueue.Enqueue(InMsg)
	}

	allfaults := s.LeaderPL.System.Height >= s.LeaderPL.SysHighest

	// Put the stuff that executes once for set of DBSignatures (after I have them all) here
	if allfaults && !s.DBSigDone && s.DBSigProcessed >= s.DBSigLimit {
		s.LogPrintf("dbsig-eom", "ProcessDBSig stop DBSig processing minute %d", s.CurrentMinute)
		//fmt.Println(fmt.Sprintf("All DBSigs are processed: allfaults(%v), && !s.DBSigDone(%v) && s.DBSigProcessed(%v)>= s.DBSigLimit(%v)",
		//	allfaults, s.DBSigDone, s.DBSigProcessed, s.DBSigLimit))
		fails := 0
		for i := range pl.FedServers {
			vm := pl.VMs[i]
			if len(vm.List) > 0 {
				tdbsig, ok := vm.List[0].(*messages.DirectoryBlockSignature)
				if !ok || !tdbsig.Matches {
					s.DBSigProcessed--
					return false
				}
			}
		}
		if fails > 0 {
			//fmt.Println("DBSig Fails Detected")
			return false
		}

		// TODO: check signatures here.  Count what match and what don't.  Then if a majority
		// disagree with us, null our entry out.  Otherwise toss our DBState and ask for one from
		// our neighbors.
		if !s.KeepMismatch && !pl.CheckDiffSigTally() {
			return false
		}

		s.ReviewHolding()
		s.Saving = false
		s.DBSigDone = true // p
		//		s.LogPrintf("dbsig-eom", "DBSIGDone written %v @ %s", s.DBSigDone, atomic.WhereAmIString(0))
	}
	return false
	/*
		err := s.LeaderPL.AdminBlock.AddDBSig(dbs.ServerIdentityChainID, dbs.DBSignature)
		if err != nil {
			fmt.Printf("Error in adding DB sig to admin block, %s\n", err.Error())
		}
	*/
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
			hb.DBHeight = s.LLeaderHeight
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
		if ok && s.Replay.IsTSValidAndUpdateState(constants.INTERNAL_REPLAY, cc.GetSigHash().Fixed(), cc.GetTimestamp(), now) {
			if s.NoEntryYet(cc.EntryHash, cc.GetTimestamp()) {
				cmsg := new(messages.CommitChainMsg)
				cmsg.CommitChain = cc
				s.PutCommit(cc.EntryHash, cmsg)
			}
			continue
		}
		ce, ok := entry.(*entryCreditBlock.CommitEntry)
		if ok && s.Replay.IsTSValidAndUpdateState(constants.INTERNAL_REPLAY, ce.GetSigHash().Fixed(), ce.GetTimestamp(), now) {
			if s.NoEntryYet(ce.EntryHash, ce.GetTimestamp()) {
				emsg := new(messages.CommitEntryMsg)
				emsg.CommitEntry = ce
				s.PutCommit(ce.EntryHash, emsg)
			}
			continue
		}
	}
}

func (s *State) GetNewEBlocks(dbheight uint32, hash interfaces.IHash) interfaces.IEntryBlock {
	if dbheight <= s.GetHighestSavedBlk()+2 {
		pl := s.ProcessLists.Get(dbheight)
		if pl == nil {
			return nil
		}
		return pl.GetNewEBlocks(hash)
	}
	return nil
}

func (s *State) IsNewOrPendingEBlocks(dbheight uint32, hash interfaces.IHash) bool {
	if dbheight <= s.GetHighestSavedBlk()+2 {
		pl := s.ProcessLists.Get(dbheight)
		if pl == nil {
			return false
		}
		eblk := pl.GetNewEBlocks(hash)
		if eblk != nil {
			return true
		}

		return pl.IsPendingChainHead(hash)
	}
	return false
}

func (s *State) PutNewEBlocks(dbheight uint32, hash interfaces.IHash, eb interfaces.IEntryBlock) {
	pl := s.ProcessLists.Get(dbheight)
	pl.AddNewEBlocks(hash, eb)
	// We no longer need them in this map, as they are in the other
	pl.PendingChainHeads.Delete(hash.Fixed())
}

func (s *State) PutNewEntries(dbheight uint32, hash interfaces.IHash, e interfaces.IEntry) {
	pl := s.ProcessLists.Get(dbheight)
	pl.AddNewEntry(hash, e)
}

// Returns the oldest, not processed, Commit received
func (s *State) NextCommit(hash interfaces.IHash) interfaces.IMsg {
	c := s.Commits.Get(hash.Fixed()) //  s.Commits[hash.Fixed()]
	return c
}

// IsHighestCommit will determine if the commit given has more entry credits than the current
// commit in the commit hashmap. If there is no prior commit, this will also return true.
func (s *State) IsHighestCommit(hash interfaces.IHash, msg interfaces.IMsg) bool {
	e, ok1 := s.Commits.Get(hash.Fixed()).(*messages.CommitEntryMsg)
	m, ok1b := msg.(*messages.CommitEntryMsg)
	ec, ok2 := s.Commits.Get(hash.Fixed()).(*messages.CommitChainMsg)
	mc, ok2b := msg.(*messages.CommitChainMsg)

	// Keep the most entry credits. If the current (e,ec) is >=, then the message is not
	// the highest.
	switch {
	case ok1 && ok1b && e.CommitEntry.Credits >= m.CommitEntry.Credits:
	case ok2 && ok2b && ec.CommitChain.Credits >= mc.CommitChain.Credits:
	default:
		// Returns true when new commit is greater than old, or if old does not exist
		return true
	}
	return false
}

func (s *State) PutCommit(hash interfaces.IHash, msg interfaces.IMsg) {
	if s.IsHighestCommit(hash, msg) {
		s.Commits.Put(hash.Fixed(), msg)
	}
}

func (s *State) GetHighestAck() uint32 {
	return s.HighestAck
}

func (s *State) SetHighestAck(dbht uint32) {
	if dbht > s.HighestAck {
		s.HighestAck = dbht
	}
}

// This is the highest block signed off and recorded in the Database.
func (s *State) GetHighestSavedBlk() uint32 {
	v := s.DBStates.GetHighestSavedBlk()
	HighestSaved.Set(float64(v))
	return v
}

// This is the highest block signed off, but not necessarily validated.
func (s *State) GetHighestCompletedBlk() uint32 {
	v := s.DBStates.GetHighestCompletedBlk()
	HighestCompleted.Set(float64(v))
	return v
}

// This is lowest block currently under construction under the "leader".
func (s *State) GetLeaderHeight() uint32 {
	return s.LLeaderHeight
}

// The highest block for which we have received a message.  Sometimes the same as
// BuildingBlock(), but can be different depending or the order messages are received.
func (s *State) GetHighestKnownBlock() uint32 {
	if s.ProcessLists == nil {
		return 0
	}
	HighestKnown.Set(float64(s.HighestKnown))
	return s.HighestKnown
}

func (s *State) GetF(rt bool, adr [32]byte) (v int64) {
	ok := false
	if rt {
		pl := s.ProcessLists.Get(s.LLeaderHeight)
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
		pl := s.ProcessLists.Get(s.LLeaderHeight)
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
	ok := false
	if rt {
		pl := s.ProcessLists.Get(s.LLeaderHeight)
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
		pl := s.ProcessLists.Get(s.LLeaderHeight)
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
func (s *State) NewAck(msg interfaces.IMsg, balanceHash interfaces.IHash) interfaces.IMsg {

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
	ack.BalanceHash = balanceHash
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
