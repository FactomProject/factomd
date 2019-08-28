// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"errors"
	"fmt"
	"hash"
	"os"
	"reflect"
	"sort"
	"time"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
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

func (s *State) CheckFileName(name string) bool {
	return messages.CheckFileName(name)
}

func (s *State) DebugExec() (ret bool) {
	return globals.Params.DebugLogRegEx != ""
}

func (s *State) LogMessage(logName string, comment string, msg interfaces.IMsg) {
	if s.DebugExec() {
		if s == nil {
			messages.StateLogMessage("unknown", 0, 0, logName, comment, msg)
		} else {
			messages.StateLogMessage(s.FactomNodeName, int(s.LLeaderHeight), int(s.CurrentMinute), logName, comment, msg)
		}
	}
}

func (s *State) LogPrintf(logName string, format string, more ...interface{}) {
	if s.DebugExec() {
		if s == nil {
			messages.StateLogPrintf("unknown", 0, 0, logName, format, more...)
		} else {
			messages.StateLogPrintf(s.FactomNodeName, int(s.LLeaderHeight), int(s.CurrentMinute), logName, format, more...)
		}
	}
}
func (s *State) AddToHolding(hash [32]byte, msg interfaces.IMsg) {
	if !constants.NeedsAck(msg.Type()) {
		s.LogMessage("holding", "add non-ack'd", msg)
	}
	_, ok := s.Holding[hash]
	if !ok {
		s.Holding[hash] = msg
		s.LogMessage("holding", "add", msg)
		TotalHoldingQueueInputs.Inc()
	}
}

func (s *State) DeleteFromHolding(hash [32]byte, msg interfaces.IMsg, reason string) {
	_, ok := s.Holding[hash]
	if ok {
		delete(s.Holding, hash)
		s.LogMessage("holding", "delete "+reason, msg)
		TotalHoldingQueueOutputs.Inc()
	}

	s.Hold.RemoveDependentMsg(hash, reason)

}

var FilterTimeLimit = int64(Range * 60 * 2 * 1000000000) // Filter hold two hours of messages, one in the past one in the future

func (s *State) GetFilterTimeNano() int64 {
	t := s.GetMessageFilterTimestamp().GetTime().UnixNano() // this is the start of the filter
	if t == 0 {
		panic("got 0 time")
	}
	return t
}

// this is the common validation to all messages. they must not be a reply, they must not be out size the time window
// for the replay filter.
func (s *State) Validate(msg interfaces.IMsg) (validToSend int, validToExec int) {
	// check the time frame of messages with ACKs and reject any that are before the message filter time (before boot
	// or outside the replay filter time frame)

	//defer func() {
	//	s.LogMessage("msgvalidation", fmt.Sprintf("send=%d execute=%d local=%v %s", *(&validToSend), *(&validToExec), msg.IsLocal(), atomic.WhereAmIString(1)), msg)
	//}()

	// During boot ignore messages that are more than 15 minutes old...
	if s.IgnoreMissing && msg.Type() != constants.DBSTATE_MSG {
		now := s.GetTimestamp().GetTimeSeconds()
		if now-msg.GetTimestamp().GetTimeSeconds() > 60*15 {
			s.LogMessage("executeMsg", "ignoreMissing", msg)
			return -1, -1
		}
	}

	if constants.NeedsAck(msg.Type()) {
		// Make sure we don't put in an old ack'd message (outside our repeat filter range)
		filterTime := s.GetFilterTimeNano()

		if filterTime == 0 {
			panic("got 0 time")
		}

		msgtime := msg.GetTimestamp().GetTime().UnixNano()

		// Make sure we don't put in an old msg (outside our repeat range)
		{ // debug
			if msgtime < filterTime || msgtime > (filterTime+FilterTimeLimit) {
				s.LogPrintf("executeMsg", "MsgFilter %s", s.GetMessageFilterTimestamp().GetTime().String())

				s.LogPrintf("executeMsg", "Leader    %s", s.GetLeaderTimestamp().GetTime().String())
				s.LogPrintf("executeMsg", "Message   %s", msg.GetTimestamp().GetTime().String())
			}
		}
		// messages before message filter timestamp it's an old message
		if msgtime < filterTime {
			s.LogMessage("executeMsg", "drop message, more than an hour in the past", msg)
			return -1, -1 // Old messages are bad.
		} else if msgtime > (filterTime + FilterTimeLimit) {
			s.LogMessage("executeMsg", "hold message from the future", msg)
			return 0, 0 // Far Future (>1H) stuff I can hold for now.  It might be good later?
		}
	}
	switch msg.Type() {
	case constants.REVEAL_ENTRY_MSG, constants.COMMIT_ENTRY_MSG, constants.COMMIT_CHAIN_MSG:
		if !s.NoEntryYet(msg.GetHash(), nil) {
			s.LogMessage("executeMsg", "drop, already committed", msg)
			return -1, -1
		}
	}

	// Valid to send is a bit different from valid to execute.  Check for valid to send here.
	validToSend = msg.Validate(s)
	if validToSend == 0 { // if the msg says hold then we hold...
		return 0, 0
	}
	if validToSend == -1 { // if the msg says drop then we drop...
		return -1, -1
	}
	if validToSend == -2 { // if the msg says New hold then we don't execute...
		return 0, -2
	}

	if validToSend != 1 { // if the msg says anything other than valid
		s.LogMessage("badEvents", fmt.Sprintf("Invalid validity code %d", validToSend), msg)
		panic("unexpected validity code")
	}

	// if it is valid to send then we check other stuff ...

	_, ok := s.Replay.Valid(constants.INTERNAL_REPLAY, msg.GetRepeatHash().Fixed(), msg.GetTimestamp(), s.GetTimestamp())
	if !ok {
		consenLogger.WithFields(msg.LogFields()).Debug("executeMsg (Replay Invalid)")
		s.LogMessage("executeMsg", "drop, INTERNAL_REPLAY", msg)
		return -1, -1
	}

	// if its local and valid to send, then execute it.
	if msg.IsLocal() {
		return 1, 1
	}

	// only valid to send messages past here

	msg.ComputeVMIndex(s)
	vmIndex := msg.GetVMIndex()
	// If we are not the leader, or this isn't the VM we are responsible for ...
	if !s.Leader || (s.LeaderVMIndex != vmIndex) {
		if constants.NeedsAck(msg.Type()) {
			// don't need to check for a matching ack for ACKs or local messages
			// for messages that get ACK make sure we can expect to process them
			ack, _ := s.Acks[msg.GetMsgHash().Fixed()].(*messages.Ack)
			if ack == nil {
				s.LogPrintf("executeMsg", "LeaderVm = %d, msg vm = %d M-%x", s.LeaderVMIndex, vmIndex, msg.GetMsgHash().Bytes()[:3])
				s.LogMessage("executeMsg", "hold, no ack yet", msg)
				return 1, 0
			}

			pl := s.ProcessLists.Get(ack.DBHeight)
			if pl == nil {
				s.LogMessage("executeMsg", "hold, no process list yet", msg)
				return 1, 0
			}
		}
	}

	return 1, 1
}

func (s *State) executeMsg(msg interfaces.IMsg) (ret bool) {
	// track how long we spend in executeMsg
	preExecuteMsgTime := time.Now()
	defer func() {
		executeMsgTime := time.Since(preExecuteMsgTime)
		TotalExecuteMsgTime.Add(float64(executeMsgTime.Nanoseconds()))
	}()

	if s.executeRecursionDetection == nil {
		s.executeRecursionDetection = make(map[[32]byte]interfaces.IMsg, 10)
	}

	repeatHash := msg.GetRepeatHash().Fixed()
	{
		// detect if we recurse with the same message we are currently executing
		_, ok := s.executeRecursionDetection[repeatHash]
		if ok {
			s.LogMessage("executeMsg", "recursion detected for ", msg)
			panic("executeMsg() recursed")
		}
		s.executeRecursionDetection[repeatHash] = msg
		defer delete(s.executeRecursionDetection, repeatHash)
	}

	if msg.GetHash() == nil || reflect.ValueOf(msg.GetHash()).IsNil() {
		s.LogMessage("badEvents", "Nil hash in executeMsg", msg)
		return false
	}

	s.SetString()
	msg.ComputeVMIndex(s)

	validToSend, validToExecute := s.Validate(msg)

	if validToSend == 1 {
		msg.SendOut(s, msg)
	}

	switch validToExecute {
	case 1:
		switch msg.Type() {
		case constants.REVEAL_ENTRY_MSG:
			if !s.NoEntryYet(msg.GetHash(), nil) {
				s.DeleteFromHolding(msg.GetMsgHash().Fixed(), msg, "AlreadyCommitted1") // delete reveal
				s.Commits.Delete(msg.GetHash().Fixed())
				s.LogMessage("executeMsg", "drop_reveal, already committed", msg)
				return true
			}
			s.AddToHolding(msg.GetMsgHash().Fixed(), msg) // add valid commit/reveal to holding in case it fails to get added
		case constants.COMMIT_ENTRY_MSG, constants.COMMIT_CHAIN_MSG:
			if !s.NoEntryYet(msg.GetHash(), nil) {
				s.DeleteFromHolding(msg.GetMsgHash().Fixed(), msg, "AlreadyCommitted2") // delete commit
				s.DeleteFromHolding(msg.GetHash().Fixed(), msg, "AlreadyCommitted3")    // delete reveal
				s.Commits.Delete(msg.GetHash().Fixed())
				s.LogMessage("executeMsg", "drop_commit, already committed", msg)
				return true
			}
			s.AddToHolding(msg.GetMsgHash().Fixed(), msg) // add valid commit/reveal to holding in case it fails to get added
		}

		var vm *VM = nil
		if s.Leader && s.RunLeader {
			vm = s.LeaderPL.VMs[s.LeaderVMIndex]
		}

		var vml int = 0
		var vmh int = 0
		var vms bool = false

		if vm != nil {
			vms = vm.Synced
			vmh = int(vm.Height)
			if vm.List != nil {
				vml = len(vm.List)
			}
		}
		local := msg.IsLocal()
		vmi := msg.GetVMIndex()
		hkb := s.GetHighestKnownBlock()

		if s.RunLeader &&
			s.Leader &&
			!s.Saving && // if not between blocks
			vm != nil && vmh == vml && // if we have processed to the end of the process list
			(!s.Syncing || !vms) && // if not syncing or this VM is not yet synced
			(local || vmi == s.LeaderVMIndex) && // if it's a local message or it a message for our VM
			s.LeaderPL.DBHeight+1 >= hkb {
			if vml == 0 { // if we have not generated a DBSig ...
				s.SendDBSig(s.LLeaderHeight, s.LeaderVMIndex) // ExecuteMsg()
				TotalXReviewQueueInputs.Inc()
				s.XReview = append(s.XReview, msg)
				s.LogMessage("executeMsg", "Missing DBSig use XReview", msg)
			} else {
				s.LogMessage("executeMsg", fmt.Sprintf("LeaderExecute[%d]", s.LeaderVMIndex), msg)
				msg.LeaderExecute(s)
			}
		} else {
			s.LogMessage("executeMsg", fmt.Sprintf("FollowerExecute[%d]", s.LeaderVMIndex), msg)
			if s.Leader {
				s.LogPrintf("executeMsg", "cause:"+
					" s.RunLeader(%v) && s.Leader(%v) && !s.Saving(%v) &&	vm(%p) != nil && vmh(%v) == vml(%v) && "+
					"(!s.Syncing(%v) || !vms(%v)) && (local(%v) || vmi(%v) == s.LeaderVMIndex(%v)) && "+
					"s.LeaderPL.DBHeight(%v)+1 >= hkb(%v)",
					s.RunLeader, s.Leader, s.Saving, vm, vmh, vml, s.Syncing, vms, local, vmi, s.LeaderVMIndex, s.LeaderPL.DBHeight, hkb)
			}
			msg.FollowerExecute(s)
		}

		return true

	case 0:
		// Sometimes messages we have already processed are in the msgQueue from holding when we execute them
		// this check makes sure we don't put them back in holding after just deleting them
		s.LogMessage("executeMsg", "hold executeMsg", msg)
		s.AddToHolding(msg.GetMsgHash().Fixed(), msg) // Add message where validToExecute==0
		return false

	case -2:
		s.LogMessage("executeMsg", "dependent_hold executeMsg", msg)
		return false

	default:
		s.DeleteFromHolding(msg.GetMsgHash().Fixed(), msg, "InvalidMsg") // delete commit
		if !msg.SentInvalid() {
			msg.MarkSentInvalid(true)
			s.networkInvalidMsgQueue <- msg
		}
		return true
	}

}

func (s *State) Process() (progress bool) {
	if ValidationDebug {
		s.LogPrintf("executeMsg", "start State.Process() msgs=%d acks=%d in=%d in2=%d", len(s.msgQueue), len(s.ackQueue), s.inMsgQueue.Length(), s.inMsgQueue2.Length())
		defer s.LogPrintf("executeMsg", "end State.Process() %v", *&progress)
	}

	s.StateProcessCnt++
	if s.ResetRequest {
		s.ResetRequest = false
		s.DoReset()
		return false
	}

	LeaderPL := s.ProcessLists.Get(s.LLeaderHeight)

	if s.LeaderPL != LeaderPL {
		s.LogPrintf("ExecuteMsg", "Process: Unexpected change in LeaderPL")
		s.LeaderPL = LeaderPL
	}
	now := s.GetTimestamp().GetTimeMilli() // Timestamps are in milliseconds, so wait 20

	// If we are not running the leader, then look to see if we have waited long enough to
	// start running the leader.  If we are, start the clock on Ignoring Missing Messages.  This
	// is so we don't conflict with past version of the network if we have to reboot the network.
	var Leader bool
	var LeaderVMIndex int
	if s.CurrentMinute > 9 {
		Leader, LeaderVMIndex = s.LeaderPL.GetVirtualServers(9, s.IdentityChainID)
	} else {
		Leader, LeaderVMIndex = s.LeaderPL.GetVirtualServers(s.CurrentMinute, s.IdentityChainID)
	}
	if s.LLeaderHeight != 0 { // debug
		if s.Leader != Leader {
			s.LogPrintf("executeMsg", "State.Process() unexpectedly setting s.Leader to %v", Leader)
		}
		if s.LeaderVMIndex != LeaderVMIndex {
			s.LogPrintf("executeMsg", "State.Process()  unexpectedly setting s.LeaderVMIndex to %v", LeaderVMIndex)
		}
	}
	// gotta set them here for the first initialization, maybe ought to call movetoheight()? but not here...
	s.Leader = Leader
	s.LeaderVMIndex = LeaderVMIndex

	if !s.RunLeader {
		if now-s.StartDelay > s.StartDelayLimit {
			if s.DBFinished == true {
				s.RunLeader = true
				if !s.IgnoreDone {
					s.StartDelay = now // Reset StartDelay for Ignore Missing
					s.IgnoreDone = true
				}
			}
		}
	} else if s.IgnoreMissing {
		if now-s.StartDelay > s.StartDelayLimit {
			s.IgnoreMissing = false
		}
	}

	process := []interfaces.IMsg{}

	hsb := s.GetHighestSavedBlk()

	// trim any received DBStatesReceived messages that are fully processed
	completed := s.GetHighestLockedSignedAndSavesBlk()

	cut := int(completed) - s.DBStatesReceivedBase
	if cut > 0 {
		s.LogPrintf("dbstateprocess", "Cut %d (%d to %d) from DBStatesReceived", cut, s.DBStatesReceivedBase, s.DBStatesReceivedBase+cut-1)
		if cut >= len(s.DBStatesReceived) {
			s.DBStatesReceived = s.DBStatesReceived[:0]
			s.DBStatesReceivedBase = int(hsb + 1)
		} else {
			s.DBStatesReceived = append(make([]*messages.DBStateMsg, 0), s.DBStatesReceived[cut:]...)
			s.DBStatesReceivedBase += cut
		}
	}

	/** Process all the DBStatesReceived  that might be pending **/
	if len(s.DBStatesReceived) > 0 {

		for {
			hsb := s.GetHighestSavedBlk()
			// Get the index of the next DBState
			ix := int(hsb) - s.DBStatesReceivedBase + 1
			// Make sure we are in range
			if ix < 0 || ix >= len(s.DBStatesReceived) {
				break // We have nothing for the system, given its current height.
			}
			if msg := s.DBStatesReceived[ix]; msg != nil {
				ret := s.executeMsg(msg)
				s.LogPrintf("dbstateprocess", "Trying to process DBStatesReceived %d, %t", s.DBStatesReceivedBase+ix, ret)
			}

			// if we can not process a DBStatesReceived then go process some messages
			if hsb == s.GetHighestSavedBlk() {
				break
			}
		}
	}

	// Process inbound messages
	preEmptyLoopTime := time.Now()
emptyLoop:
	for i := 0; i < 100; i++ {
		var msg interfaces.IMsg
		select {
		// We have prioritizedMsgQueue listed twice, meaning it has 2 chances to be
		// randomly selected to unblock and execute.
		case msg = <-s.prioritizedMsgQueue:
			s.LogMessage("prioritizedMsgQueue", "Execute", msg)
		case msg = <-s.prioritizedMsgQueue:
			s.LogMessage("prioritizedMsgQueue", "Execute", msg)
		case msg = <-s.msgQueue:
			s.LogMessage("msgQueue", "Execute", msg)
		case msg = <-s.ackQueue:
			s.LogMessage("ackQueue", "Execute", msg)
		case msg = <-s.ackQueue:
			s.LogMessage("ackQueue", "Execute", msg)
		case msg = <-s.ackQueue:
			s.LogMessage("ackQueue", "Execute", msg)
		default:
			break emptyLoop
		}
		progress = s.executeMsg(msg) || progress
	}
	emptyLoopTime := time.Since(preEmptyLoopTime)
	TotalEmptyLoopTime.Add(float64(emptyLoopTime.Nanoseconds()))

	preProcessXReviewTime := time.Now()
	// Reprocess any stalled messages, but not so much compared inbound messages
	// Process last first

	// Check for Ack'd items in holding
	func() {
		count := 0
		if ValidationDebug {
			s.LogPrintf("executeMsg", "Start reviewAcks")
			defer s.LogPrintf("executeMsg", "end reviewAcks executed %d", count)
		}
		for _, a := range s.Acks {
			if s.Holding[a.GetHash().Fixed()] != nil {
				a.FollowerExecute(s)
				count++
			}
		}
	}()

	// only review holding if I am a leader
	if s.RunLeader && s.Leader {
		s.ReviewHolding()
		for _, msg := range s.XReview {
			if msg == nil {
				continue
			}
			// copy the messages we are responsible for and all msg that don't need ack
			// messages that need ack will get processed when thier ack arrives
			if msg.GetVMIndex() == s.LeaderVMIndex || !constants.NeedsAck(msg.Type()) {
				process = append(process, msg)
			}
		}
		// toss everything else
		s.XReview = s.XReview[:0]
	}
	if ValidationDebug {
		s.LogPrintf("executeMsg", "end reviewHolding %d", len(s.XReview))
	}

	processXReviewTime := time.Since(preProcessXReviewTime)
	TotalProcessXReviewTime.Add(float64(processXReviewTime.Nanoseconds()))

	if len(process) != 0 {
		preProcessProcChanTime := time.Now()
		if ValidationDebug {
			s.LogPrintf("executeMsg", "Start processloop %d", len(process))
		}
		for _, msg := range process {
			newProgress := s.executeMsg(msg)
			if ValidationDebug && newProgress {
				s.LogMessage("executeMsg", "progress set by ", msg)
			}
			progress = newProgress || progress //
			s.LogMessage("executeMsg", "From process", msg)
			s.UpdateState()
		} // processLoop for{...}

		if ValidationDebug {
			s.LogPrintf("executeMsg", "end processloop")
		}
		processProcChanTime := time.Since(preProcessProcChanTime)
		TotalProcessProcChanTime.Add(float64(processProcChanTime.Nanoseconds()))
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

	preReviewHoldingTime := time.Now()
	if len(s.XReview) > 0 {
		return
	}

	now := s.GetTimestamp()
	if s.ResendHolding == nil {
		s.ResendHolding = now
	}
	if now.GetTimeMilli()-s.ResendHolding.GetTimeMilli() < 100 {
		return
	}

	if len(s.Holding) == 0 {
		return
	}

	if ValidationDebug {
		s.LogPrintf("executeMsg", "Start reviewHolding")
		defer s.LogPrintf("executeMsg", "end reviewHolding holding=%d, xreview=%d", len(s.Holding), len(s.XReview))
	}

	// moved to once per block in move state to height
	//s.Commits.Cleanup(s)
	//s.DB.Trim()

	// Anything we are holding, we need to reprocess.
	s.XReview = make([]interfaces.IMsg, 0)

	highest := s.GetHighestKnownBlock()
	saved := s.GetHighestSavedBlk()
	if saved > highest {
		highest = saved + 1
	}

	// if the sorted holding is empty or built for the wrong VM
	if len(s.HoldingList) == 0 || s.HoldingVM != s.LeaderVMIndex {
		s.HoldingVM = s.LeaderVMIndex // save the VM I used in making this sorted list
		sorted := []interfaces.IMsg{}
		for _, v := range s.Holding {
			if v.GetVMIndex() == s.LeaderVMIndex || !constants.NeedsAck(v.Type()) {
				sorted = append(sorted, v)
			}
		}
		sort.Slice(sorted,
			func(i, j int) bool {
				a := sorted[i].GetTimestamp().GetTimeMilli()
				b := sorted[j].GetTimestamp().GetTimeMilli()
				return a < b
			})
		for k, v := range sorted {
			if k >= cap(s.HoldingList) {
				break
			}
			s.HoldingList <- v.GetMsgHash().Fixed()
		}
	}

	cnt := 1
processholdinglist:
	for {
		if cnt&0x1F == 0 && s.GetTimestamp().GetTimeMilli()-now.GetTimeMilli() > 200 {

			break processholdinglist
		}
		cnt++
		var k [32]byte
		var v interfaces.IMsg
		var ok bool
		select {
		case k = <-s.HoldingList:
			v, ok = s.Holding[k]
			if !ok || v == nil {
				continue processholdinglist
			}
		default:
			break processholdinglist
		}

		if int(highest)-int(saved) > 1000 {
			TotalHoldingQueueOutputs.Inc()
			//delete(s.Holding, k)
			s.DeleteFromHolding(k, v, "HKB-HSB>1000")
			continue processholdinglist // No point in executing if we think we can't hold this.
		}

		if v.Expire(s) {
			s.LogMessage("executeMsg", "expire from holding", v)
			s.ExpireCnt++
			TotalHoldingQueueOutputs.Inc()
			//delete(s.Holding, k)
			s.DeleteFromHolding(k, v, "expired")
			continue processholdinglist // If the message has expired, don't hold or execute
		}

		eom, ok := v.(*messages.EOM)
		if ok {
			if int(eom.DBHeight)*10+int(eom.Minute) < int(s.LLeaderHeight)*10+s.CurrentMinute {
				s.DeleteFromHolding(k, v, "old EOM")
				continue processholdinglist
			}
			if !eom.IsLocal() && eom.DBHeight > saved {
				s.SetHighestKnownBlock(eom.DBHeight)
			}
		}

		dbsigmsg, ok := v.(*messages.DirectoryBlockSignature)
		if ok {
			if dbsigmsg.DBHeight < s.LLeaderHeight || (s.CurrentMinute > 0 && dbsigmsg.DBHeight == s.LLeaderHeight) {
				TotalHoldingQueueOutputs.Inc()
				//delete(s.Holding, k)
				s.DeleteFromHolding(k, v, "Old DBSig")
				continue processholdinglist
			}
			if !dbsigmsg.IsLocal() && dbsigmsg.DBHeight > saved {
				s.SetHighestKnownBlock(dbsigmsg.DBHeight)
			}
		}

		dbsmsg, ok := v.(*messages.DBStateMsg)
		if ok && dbsmsg.DirectoryBlock.GetHeader().GetDBHeight() <= saved && saved > 0 {

			TotalHoldingQueueOutputs.Inc()
			//delete(s.Holding, k)
			s.DeleteFromHolding(k, v, "old DBState")
			continue processholdinglist
		}

		// If it is an entryCommit/ChainCommit/RevealEntry and it has a duplicate hash to an existing entry throw it away here
		ce, ok := v.(*messages.CommitEntryMsg)
		if ok {
			x := s.NoEntryYet(ce.CommitEntry.EntryHash, ce.CommitEntry.GetTimestamp())
			if !x {
				TotalHoldingQueueOutputs.Inc()
				//delete(s.Holding, k) // Drop commits with the same entry hash from holding because they are blocked by a previous entry
				s.DeleteFromHolding(k, v, "already committed")
				continue processholdinglist
			}
		}

		// If it is an chainCommit and it has a duplicate hash to an existing entry throw it away here

		cc, ok := v.(*messages.CommitChainMsg)
		if ok {
			x := s.NoEntryYet(cc.CommitChain.EntryHash, cc.CommitChain.GetTimestamp())
			if !x {
				TotalHoldingQueueOutputs.Inc()
				//delete(s.Holding, k) // Drop commits with the same entry hash from holding because they are blocked by a previous entry
				s.DeleteFromHolding(k, v, "already committed")
				continue processholdinglist
			}
		}

		validToSend, validToExecute := s.Validate(v)

		if validToSend > 0 {
			v.SendOut(s, v)
		}

		switch validToExecute {
		case -1:
			s.LogMessage("executeMsg", "invalid from holding", v)
			TotalHoldingQueueOutputs.Inc()
			//delete(s.Holding, k)
			s.DeleteFromHolding(k, v, "invalid from holding")
			continue processholdinglist
		case 0:
			continue processholdinglist
		}

		// if it is a Factoid or entry credit transaction then check BLOCK_REPLAY
		switch v.Type() {
		case constants.FACTOID_TRANSACTION_MSG, constants.COMMIT_CHAIN_MSG, constants.COMMIT_ENTRY_MSG:
			ok2 := s.FReplay.IsHashUnique(constants.BLOCK_REPLAY, v.GetRepeatHash().Fixed())
			if !ok2 {
				s.DeleteFromHolding(k, v, "BLOCK_REPLAY")
				continue processholdinglist
			}
		default:
		}
		// If a Reveal Entry has a commit available, then process the Reveal Entry and send it out.
		if re, ok := v.(*messages.RevealEntryMsg); ok {
			if !s.NoEntryYet(re.GetHash(), s.GetLeaderTimestamp()) {
				s.DeleteFromHolding(re.GetMsgHash().Fixed(), re, "already committed reveal")
				s.Commits.Delete(re.GetHash().Fixed())
				continue processholdinglist
			}
			// Needs to be our VMIndex as well, or ignore.
			if re.GetVMIndex() != s.LeaderVMIndex || !s.Leader {
				continue processholdinglist // If we are a leader, but it isn't ours, and it isn't a new minute, ignore.
			}
		}

		TotalXReviewQueueInputs.Inc()
		s.XReview = append(s.XReview, v)
		TotalHoldingQueueOutputs.Inc()
	}
	s.ResendHolding = s.GetTimestamp()
	reviewHoldingTime := time.Since(preReviewHoldingTime)
	TotalReviewHoldingTime.Add(float64(reviewHoldingTime.Nanoseconds()))
}

func (s *State) MoveStateToHeight(dbheight uint32, newMinute int) {
	//	s.LogPrintf("dbstateprocess", "MoveStateToHeight(%d-:-%d) called from %s", dbheight, newMinute, atomic.WhereAmIString(1))
	s.LogPrintf("dbstateprocess", "MoveStateToHeight(%d-:-%d)", dbheight, newMinute)

	if (s.LLeaderHeight+1 == dbheight && newMinute == 0) || (s.LLeaderHeight == dbheight && s.CurrentMinute+1 == newMinute) {
		// these are the allowed cases; move to nextblock-:-0 or move to next minute
	} else {
		s.LogPrintf("dbstateprocess", "State move between non-sequential heights from %d to %d", s.LLeaderHeight, dbheight)
		if s.LLeaderHeight != dbheight {
			fmt.Fprintf(os.Stderr, "State move between non-sequential heights from %d to %d\n", s.LLeaderHeight, dbheight)
		}
		//force sync state to a rational  state for between minutes
		s.Syncing = false    // movestatetoheight
		s.EOM = false        // movestatetoheight
		s.EOMDone = false    // movestatetoheight
		s.DBSig = false      // movestatetoheight
		s.EOMProcessed = 0   // movestatetoheight
		s.DBSigProcessed = 0 // movestatetoheight
		s.DBSigDone = false  // movestatetoheight

		if s.LeaderPL != nil {
			for _, vm := range s.LeaderPL.VMs {
				vm.Synced = false // movestatetoheight
			}
		}
	}
	// normally when loading by DBStates we jump from minute 0 to minute 0
	// when following by minute we jump from minute 10 to minute 0
	if s.LLeaderHeight != dbheight && s.CurrentMinute != 0 && s.CurrentMinute != 10 {
		s.LogPrintf("dbstateprocess", "Jump in current minute from %d-:-%d to %d-:-%d", s.LLeaderHeight, s.CurrentMinute, dbheight, newMinute)
		//fmt.Fprintf(os.Stderr, "Jump in current minute from %d-:-%d to %d-:-%d\n", s.LLeaderHeight, s.CurrentMinute, dbheight, newMinute)
	}
	//s.setCurrentMinute(newMinute)     // MoveStateToHeight() move minute
	//s.SetLLeaderHeight(int(dbheight)) // Update leader height in MoveStateToHeight

	if s.LLeaderHeight != dbheight {
		if newMinute != 0 {
			panic(fmt.Sprintf("Can't jump to the middle of a block minute: %d", newMinute))
		}
		s.DBStates.UpdateState()

		// update cached values that change with current minute
		s.CurrentMinute = 0                // Update height and minute
		s.LLeaderHeight = uint32(dbheight) // Update height and minute

		// update cached values that change with height
		s.LeaderPL = s.ProcessLists.Get(dbheight) // fix up cached values
		if s.LLeaderHeight != s.LeaderPL.DBHeight {
			panic("bad things are happening")
		}

		s.ProcessLists.Get(dbheight + 1) // Make sure next PL exists
		// We are between blocks make sure we are setup to sync
		// should already be true but if a DBSTATE got processed mid block
		// there might be a circumstance where we get here in a weird state
		// so make it the normal starting state

		// update cached values that change with height
		s.dbheights <- int(dbheight) // Notify MMR process we have moved on...

		s.CurrentMinuteStartTime = time.Now().UnixNano()
		s.CurrentBlockStartTime = s.CurrentMinuteStartTime

		// If an we added or removed servers or elections tool place in minute 9, our lists will be unsorted. Fix that
		s.LeaderPL.SortAuditServers()
		s.LeaderPL.SortFedServers()
		// check for identity change every time we start a new block before we look up our VM
		s.CheckForIDChange()
		s.Leader, s.LeaderVMIndex = s.LeaderPL.GetVirtualServers(s.CurrentMinute, s.IdentityChainID) // MoveStateToHeight block

		s.LogPrintf("executeMsg", "MoveStateToHeight set leader=%v, vmIndex = %v", s.Leader, s.LeaderVMIndex)
		// update the elections thread
		authlistMsg := s.EFactory.NewAuthorityListInternal(s.LeaderPL.FedServers, s.LeaderPL.AuditServers, s.LLeaderHeight)
		s.ElectionsQueue().Enqueue(authlistMsg)

		// Do not send out dbsigs while loading from disk
		if s.Leader && !s.LeaderPL.DBSigAlreadySent && s.LLeaderHeight > s.DBHeightAtBoot {
			s.SendDBSig(s.LLeaderHeight, s.LeaderVMIndex) // MoveStateToHeight()
		}
		s.DBStates.UpdateState() // go process the DBSigs

		// Expire old commits and stuff ...
		s.Commits.Cleanup(s)
		s.DB.Trim()

	} else if s.CurrentMinute != newMinute { // And minute
		if newMinute == 1 {
			dbstate := s.GetDBState(dbheight - 1)
			if dbstate != nil && !dbstate.Saved {
				s.LogPrintf("dbstateprocess", "Set ReadyToSave %d", dbstate.DirectoryBlock.GetHeader().GetDBHeight())
				dbstate.ReadyToSave = true
			}
			s.DBStates.UpdateState() // call to get the state signed now that the DBSigs have processed
		}
		s.CurrentMinute = newMinute // Update just the minute
		// We are between blocks make sure we are setup to sync
		// should already be true but if a DBSTATE got processed mid block
		// there might be a circumstance where we get here in a weird state
		// so make it the normal starting state

		s.CurrentMinuteStartTime = time.Now().UnixNano()
		// If an election took place, our lists will be unsorted. Fix that
		s.LeaderPL.SortAuditServers()
		s.LeaderPL.SortFedServers()

		s.Leader, s.LeaderVMIndex = s.LeaderPL.GetVirtualServers(newMinute, s.IdentityChainID) // MoveStateToHeight minute
		s.LogPrintf("executeMsg", "MoveStateToHeight new minute set leader=%v, vmIndex = %v", s.Leader, s.LeaderVMIndex)
	}

	{ // debug
		vmSync := false
		for _, vm := range s.LeaderPL.VMs {
			if vm != nil {
				vmSync = vmSync || vm.Synced
			}
		}

		if s.Syncing || s.EOM || s.EOMDone || s.DBSig || (s.EOMProcessed != 0) || (s.DBSigProcessed != 0) || vmSync {
			s.LogPrintf("executeMsg", "resetting syncstate in moveToHeight")
			s.LogPrintf("executeMsg", "s.Syncing=%v s.EOM=%v s.EOMDone=%v s.DBSig=%v s.DBSigDone=%v s.EOMProcessed=%d s.DBSigProcessed=%v, vmSync = %v",
				s.Syncing, s.EOM, s.EOMDone, s.DBSig, s.DBSigDone, s.EOMProcessed, s.DBSigProcessed, vmSync)

		}
	}

	// set the limits because we might have added servers
	s.EOMLimit = len(s.LeaderPL.FedServers) // We add or remove server only on block boundaries
	s.DBSigLimit = s.EOMLimit               // We add or remove server only on block boundaries
	s.LogPrintf("dbstateprocess", "MoveStateToHeight(%d-:-%d) leader=%v leaderPL=%p, leaderVMIndex=%d", dbheight, newMinute, s.Leader, s.LeaderPL, s.LeaderVMIndex)

	s.Hold.ExecuteForNewHeight(s.LLeaderHeight, s.CurrentMinute) // execute held messages
	s.Hold.Review()                                              // cleanup old messages
}

// Adds blocks that are either pulled locally from a database, or acquired from peers.
func (s *State) AddDBState(isNew bool,
	directoryBlock interfaces.IDirectoryBlock,
	adminBlock interfaces.IAdminBlock,
	factoidBlock interfaces.IFBlock,
	entryCreditBlock interfaces.IEntryCreditBlock,
	eBlocks []interfaces.IEntryBlock,
	entries []interfaces.IEBEntry) *DBState {

	// This is expensive, so only do this once the database is loaded.
	if s.DBFinished {
		s.LogPrintf("dbstateprocess", "AddDBState(isNew %v, directoryBlock %d %x, adminBlock %x, factoidBlock %x, entryCreditBlock %X, eBlocks %d, entries %d)",
			isNew, directoryBlock.GetHeader().GetDBHeight(), directoryBlock.GetHash().Bytes()[:4],
			adminBlock.GetHash().Bytes()[:4], factoidBlock.GetHash().Bytes()[:4], entryCreditBlock.GetHash().Bytes()[:4], len(eBlocks), len(entries))
	}

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

	if ht == s.LLeaderHeight-1 || (ht == s.LLeaderHeight) {
	} else {
		s.LogPrintf("dbstateprocess", "AddDBState out of order! at %d added %d", s.LLeaderHeight, ht)
		fmt.Fprintf(os.Stderr, "AddDBState() out of order! at %d added %d\n", s.LLeaderHeight, ht)
		//panic("AddDBState out of order!")
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

	//s.Holding[m.GetMsgHash().Fixed()] = m
	s.AddToHolding(m.GetMsgHash().Fixed(), m) // FollowerExecuteMsg()
	ack, _ := s.Acks[m.GetMsgHash().Fixed()].(*messages.Ack)

	if ack != nil {
		m.SetLeaderChainID(ack.GetLeaderChainID())
		m.SetMinute(ack.Minute)

		pl := s.ProcessLists.Get(ack.DBHeight)
		pl.AddToProcessList(s, ack, m)

		// Cross Boot Replay
		s.CrossReplayAddSalt(ack.DBHeight, ack.Salt)
	}
}

// execute a msg with an optional delay (in factom seconds)
func (s *State) repost(m interfaces.IMsg, delay int) {
	//whereAmI := atomic.WhereAmIString(1)
	go func() { // This is a trigger to issue the EOM, but we are still syncing.  Wait to retry.
		if delay > 0 {
			time.Sleep(time.Duration(delay) * s.FactomSecond()) // delay in Factom seconds
		}
		//s.LogMessage("MsgQueue", fmt.Sprintf("enqueue_%s(%d)", whereAmI, len(s.msgQueue)), m)
		s.LogMessage("MsgQueue", fmt.Sprintf("enqueue (%d)", len(s.msgQueue)), m)
		s.msgQueue <- m // Goes in the "do this really fast" queue so we are prompt about EOM's while syncing
	}()
}

// FactomSecond finds the time duration of 1 second relative to 10min blocks.
//		Blktime			EOMs		Second
//		600s			60s			1s
//		300s			30s			0.5s
//		120s			12s			0.2s
//		 60s			 6s			0.1s
//		 30s			 3s			0.05s
func (s *State) FactomSecond() time.Duration {
	// Convert to time.second, then divide by 600
	factomsecond := time.Duration(s.DirectoryBlockInSeconds) * time.Second / 600
	if factomsecond < time.Duration(250*time.Millisecond) {
		factomsecond = time.Duration(250 * time.Millisecond) // for really fast block we lie ...
	}
	return factomsecond
}

// Messages that will go into the Process List must match an Acknowledgement.
// The code for this is the same for all such messages, so we put it here.
//
// Returns true if it finds a match, puts the message in holding, or invalidates the message
func (s *State) FollowerExecuteEOM(m interfaces.IMsg) {

	if m.IsLocal() && !s.Leader {
		return // This is an internal EOM message.  We are not a leader so ignore.
	} else if m.IsLocal() {
		s.repost(m, 1) // Goes in the "do this really fast" queue so we are prompt about EOM's while syncing
		return
	}

	eom, ok := m.(*messages.EOM)
	if !ok {
		return
	}

	if eom.DBHeight == s.ProcessLists.Lists[0].DBHeight && int(eom.Minute) < s.CurrentMinute {
		s.LogMessage("executeMsg", "FollowerExecuteEOM drop, wrong minute", m)
		return
	}

	FollowerEOMExecutions.Inc()
	// add it to the holding queue in case AddToProcessList may remove it
	TotalHoldingQueueInputs.Inc()
	//s.Holding[m.GetMsgHash().Fixed()] = m // FollowerExecuteEOM

	s.AddToHolding(m.GetMsgHash().Fixed(), m) // follower execute nonlocal EOM
	ack, _ := s.Acks[m.GetMsgHash().Fixed()].(*messages.Ack)
	if ack != nil {
		s.LogMessage("executeMsg", "FollowerExecuteEOM add2pl", ack)
		pl := s.ProcessLists.Get(ack.DBHeight)
		pl.AddToProcessList(s, ack, m)
	}
}

func (s *State) getMsgFromHolding(h [32]byte) interfaces.IMsg {
	// check if we have a message
	m := s.Holding[h]

	if m != nil {
		return m
	} else {
		return s.Hold.GetDependentMsg(h)
	}
}

// Ack messages always match some message in the Process List.   That is
// done here, though the only msg that should call this routine is the Ack
// message.
func (s *State) FollowerExecuteAck(msg interfaces.IMsg) {
	ack := msg.(*messages.Ack)

	if ack.DBHeight > s.GetHighestKnownBlock() {
		s.SetHighestKnownBlock(ack.DBHeight)
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

	// Add the ack to the list of acks
	TotalAcksInputs.Inc()
	s.Acks[ack.GetHash().Fixed()] = ack // Add the ack to the ask list incase we can't execute the msg yet.

	m := s.getMsgFromHolding(ack.GetHash().Fixed()) // check for a matching message
	if m != nil {
		// We have an ack and a matching message go execute the message!
		s.LogMessage("executeMsg", "FollowerExecuteAck ", m)
		s.executeMsg(m) // Try executing the message, if dependencies are met then it will execute
	} else {
		s.LogMessage("executeMsg", "No Msg, keep", ack)
		pl.VMs[ack.VMIndex].ReportMissing(int(ack.Height), 0) // Ask for it now
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
	//todo: Is this better to be GetEntryDBHeightComplete()
	if dbheight > 0 && dbheight <= s.GetHighestSavedBlk() && dbheight < s.EntryDBHeightComplete {
		return
	}

	pdbstate := s.DBStates.Get(int(dbheight - 1))

	valid := pdbstate.ValidNext(s, dbstatemsg)

	switch valid {
	case 0:
		s.LogPrintf("dbstateprocess", "FollowerExecuteDBState hold for later %d", dbheight)

		ix := int(dbheight) - s.DBStatesReceivedBase
		for len(s.DBStatesReceived) <= ix {
			s.DBStatesReceived = append(s.DBStatesReceived, nil)
		}
		s.DBStatesReceived[ix] = dbstatemsg
		return
	case -1:
		s.LogPrintf("dbstateprocess", "FollowerExecuteDBState Invalid %d", dbheight)
		cntFail()
		if dbstatemsg.IsLast { // this is the last DBState in this load
			panic(fmt.Sprintf("%20s The last DBState %d saved to the database was not valid.", s.FactomNodeName, dbheight))
		}
		return
	}

	if dbstatemsg.IsLast { // this is the last DBState in this load
		s.DBFinished = true // Normal case
	}
	/**************************
	for int(s.ProcessLists.DBHeightBase)+len(s.ProcessLists.Lists) > int(dbheight+1) {
		s.ProcessLists.Lists[len(s.ProcessLists.Lists)-1].Clear()
		s.ProcessLists.Lists = s.ProcessLists.Lists[:len(s.ProcessLists.Lists)-1]
	}
	***************************/

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
		fixed := fct.GetSigHash().Fixed()
		_, valid := s.FReplay.Valid(
			constants.BLOCK_REPLAY,
			fixed,
			fct.GetTimestamp(),
			dbstatemsg.DirectoryBlock.GetHeader().GetTimestamp())
		// FD-682 on 09/27/18 there was a block 160181 that was more than 60 minutes long and this transaction is judged invalid because it was 88 minutes
		// after the block start time. The transaction is valid.
		if dbheight == 160181 && fixed == [32]byte{0xf9, 0x6a, 0xf0, 0x63, 0xff, 0xfd, 0x90, 0xe2, 0x61, 0xe4, 0x5c, 0xbc, 0xdf, 0x34, 0xc3, 0x95,
			0x8c, 0xf5, 0x4e, 0x23, 0x88, 0xfd, 0x3f, 0x8e, 0xf9, 0x07, 0xdc, 0xa9, 0x03, 0xea, 0x2f, 0x2e} {
			valid = true
		}
		// f96af063fffd90e261e45cbcdf34c3958cf54e2388fd3f8ef907dca903ea2f2e

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
	// clear all the syncing flags because we are done with this block
	s.EOM = false
	s.EOMDone = false
	s.DBSig = false
	s.DBSigDone = false
	s.Syncing = false // FollowerExecuteDBState
	s.Saving = true

	// At this point the block is good, make sure not to ask for it anymore
	if !dbstatemsg.IsInDB {
		s.StatesReceived.Notify <- msg.(*messages.DBStateMsg)
	}
	s.DBStates.UpdateState()

}

func (s *State) FollowerExecuteMMR(m interfaces.IMsg) {

	// Just ignore missing messages for a period after going off line or starting up.

	if s.IgnoreMissing {
		s.LogMessage("executeMsg", "drop IgnoreMissing", m)
		return
	}
	// Drop the missing message response if it's already in the process list
	_, valid := s.Replay.Valid(constants.INTERNAL_REPLAY, m.GetRepeatHash().Fixed(), m.GetTimestamp(), s.GetTimestamp())
	if !valid {
		s.LogMessage("executeMsg", "drop, INTERNAL_REPLAY", m)
		return
	}

	mmr, _ := m.(*messages.MissingMsgResponse)

	if mmr.AckResponse == nil {
		s.LogMessage("executeMsg", "drop MMR ack is nil", m)
		return
	}

	if mmr.MsgResponse == nil {
		s.LogMessage("executeMsg", "drop MMR message is nil", m)
		return
	}

	msg := mmr.MsgResponse
	ack, ok := mmr.AckResponse.(*messages.Ack)
	if !ok {
		s.LogMessage("executeMsg", "drop mmr ack is not an ack", m)
		return
	}

	_, validToExecute := s.Validate(ack)
	if validToExecute == -1 {
		s.LogMessage("executeMsg", "drop MMR ack invalid", ack)
		return
	}
	if validToExecute == -2 {
		s.LogMessage("executeMsg", "dependent_hold", ack)
		return
	}

	// If we don't need this message, we don't have to do everything else.
	_, validToExecute = s.Validate(msg)
	if validToExecute == -1 {
		s.LogMessage("executeMsg", "drop MMR message invalid", m)
		return
	}
	if validToExecute == -2 {
		s.LogMessage("executeMsg", "dependent_hold", m)
		return
	}
	ack.Response = true

	pl := s.ProcessLists.Get(ack.DBHeight)

	if pl == nil {
		s.LogMessage("executeMsg", "drop No Processlist", m)
		return
	}

	TotalAcksInputs.Inc()

	s.LogMessage("executeMsg", "FollowerExecute3(MMR)", msg)
	msg.FollowerExecute(s)

	s.LogMessage("executeMsg", "FollowerExecute4(MMR)", ack)
	ack.FollowerExecute(s)

	s.MissingResponseAppliedCnt++
}

func (s *State) FollowerExecuteDataResponse(m interfaces.IMsg) {
	msg, ok := m.(*messages.DataResponse)
	if !ok {
		s.LogMessage("executeMsg", "Drop, not a DataResponce", msg)
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
		s.WriteEntry <- entry // DataResponse
	}
}

func (s *State) FollowerExecuteMissingMsg(msg interfaces.IMsg) {
	s.LogMessage("badEvents", "follower executed missing message, should never happen", msg)
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

	//s.Holding[m.GetMsgHash().Fixed()] = m // hold in  FollowerExecuteRevealEntry
	s.AddToHolding(m.GetMsgHash().Fixed(), m) // hold in  FollowerExecuteRevealEntry

	// still need this because of the call from FollowerExecuteCommitEntry and FollowerExecuteCommitChain
	ack, _ := s.Acks[m.GetMsgHash().Fixed()].(*messages.Ack)

	if ack == nil {
		// todo: prevent this log from double logging
		s.LogMessage("executeMsg", "hold, no ack yet1", m)
		return
	}

	//ack.SendOut(s, ack)
	m.SetLeaderChainID(ack.GetLeaderChainID())
	m.SetMinute(ack.Minute)
	m.SendOut(s, m)

	pl := s.ProcessLists.Get(ack.DBHeight)
	if pl == nil {
		s.LogMessage("executeMsg", "hold, no process list yet1", m)
		return
	}

	// Add the message and ack to the process list.
	pl.AddToProcessList(s, ack, m)

	// Check to make sure AddToProcessList removed it from holding (added it to the list)
	if s.Holding[m.GetMsgHash().Fixed()] != nil {
		s.LogMessage("executeMsg", "add to processlist failed", m)
		return
	}

	msg := m.(*messages.RevealEntryMsg)
	TotalCommitsOutputs.Inc()

	// This is so the api can determine if a chainhead is about to be updated. It fixes a race condition
	// on the api. MUST BE BEFORE THE REPLAY FILTER ADD
	pl.PendingChainHeads.Put(msg.Entry.GetChainID().Fixed(), msg)
}

func (s *State) LeaderExecute(m interfaces.IMsg) {
	vm := s.LeaderPL.VMs[s.LeaderVMIndex]
	if len(vm.List) != vm.Height {
		s.repost(m, 1) // Goes in the "do this really fast" queue so we are prompt about EOM's while syncing
		return
	}
	LeaderExecutions.Inc()
	_, ok := s.Replay.Valid(constants.INTERNAL_REPLAY, m.GetRepeatHash().Fixed(), m.GetTimestamp(), s.GetTimestamp())
	if !ok {
		TotalHoldingQueueOutputs.Inc()
		//delete(s.Holding, m.GetMsgHash().Fixed())
		s.DeleteFromHolding(m.GetMsgHash().Fixed(), m, "INTERNAL_REPLAY")
		if s.DebugExec() {
			s.LogMessage("executeMsg", "drop replay", m)
		}
		return
	}

	ack := s.NewAck(m, nil).(*messages.Ack) // LeaderExecute
	m.SetLeaderChainID(ack.GetLeaderChainID())
	m.SetMinute(ack.Minute)

	s.ProcessLists.Get(ack.DBHeight).AddToProcessList(s, ack, m)
}

func (s *State) LeaderExecuteEOM(m interfaces.IMsg) {
	LeaderEOMExecutions.Inc()
	if !m.IsLocal() {
		s.FollowerExecuteEOM(m)
		return
	}

	pl := s.ProcessLists.Get(s.LLeaderHeight)
	vm := pl.VMs[s.LeaderVMIndex]

	// If we have already issued an EOM for the minute being sync'd
	// then this should be the next EOM but we can't do that just yet.
	if vm.EomMinuteIssued == s.CurrentMinute+1 {
		s.LogMessage("executeMsg", fmt.Sprintf("repost, eomminute issued != s.CurrentMinute+1 : %d - %d", vm.EomMinuteIssued, s.CurrentMinute+1), m)
		s.repost(m, 1) // Do not drop the message, we only generate 1 local eom per height/min, let validate drop it
		return
	}
	// The zero based minute for the message is equal to
	// the one based "LastMinute".  This way we know we are
	// generating minutes in order.

	if len(vm.List) != vm.Height {
		s.LogMessage("executeMsg", "repost, not pl synced", m)
		s.repost(m, 1) // Do not drop the message, we only generate 1 local eom per height/min, let validate drop it
		return
	}
	eom := m.(*messages.EOM)

	// Put the System Height and Serial Hash into the EOM
	eom.SysHeight = uint32(pl.System.Height)

	if vm.Synced {
		s.LogMessage("executeMsg", "drop, already sync'd", m)
		s.repost(m, 1) // Do not drop the message, we only generate 1 local eom per height/min, let validate drop it
		return
	}

	// Committed to sending an EOM now
	vm.EomMinuteIssued = s.CurrentMinute + 1

	fix := false

	if eom.DBHeight != s.LLeaderHeight || eom.VMIndex != s.LeaderVMIndex || eom.Minute != byte(s.CurrentMinute) {
		s.LogPrintf("executeMsg", "EOM has wrong data expected DBH/VM/M %d/%d/%d", s.LLeaderHeight, s.LeaderVMIndex, s.CurrentMinute)
		fix = true
	}

	// make sure EOM has the right data
	eom.DBHeight = s.LLeaderHeight
	eom.VMIndex = s.LeaderVMIndex
	// eom.Minute is zerobased, while LeaderMinute is 1 based.  So
	// a simple assignment works.
	eom.Minute = byte(s.CurrentMinute)
	eom.Sign(s)
	eom.MsgHash = nil                       // delete any existing hash so it will be recomputed
	eom.RepeatHash = nil                    // delete any existing hash so it will be recomputed
	ack := s.NewAck(m, nil).(*messages.Ack) // LeaderExecuteEOM()
	eom.SetLocal(false)

	if fix {
		s.LogMessage("executeMsg", "fixed EOM", eom)
		s.LogMessage("executeMsg", "matching ACK", ack)
	}

	TotalAcksInputs.Inc()
	s.Acks[eom.GetMsgHash().Fixed()] = ack
	ack.SendOut(s, ack)
	eom.SendOut(s, eom)
	s.FollowerExecuteEOM(eom)
	s.UpdateState()
}

func (s *State) LeaderExecuteDBSig(m interfaces.IMsg) {
	LeaderExecutions.Inc()
	dbs := m.(*messages.DirectoryBlockSignature)
	pl := s.ProcessLists.Get(dbs.DBHeight)

	s.LogMessage("executeMsg", "LeaderExecuteDBSig", m)
	if dbs.DBHeight != s.LLeaderHeight {
		s.LogMessage("executeMsg", "followerExec", m)
		m.FollowerExecute(s)
		return
	}

	if pl.VMs[dbs.VMIndex].Height > 0 {
		s.LogPrintf("executeMsg", "DBSig issue height = %d, length = %d", pl.VMs[dbs.VMIndex].Height, len(pl.VMs[dbs.VMIndex].List))
		s.LogMessage("executeMsg", "drop, already processed ", pl.VMs[dbs.VMIndex].List[0])
		return
	}

	if len(pl.VMs[dbs.VMIndex].List) > 0 && pl.VMs[dbs.VMIndex].List[0] != nil {
		s.LogPrintf("executeMsg", "DBSig issue height = %d, length = %d", pl.VMs[dbs.VMIndex].Height, len(pl.VMs[dbs.VMIndex].List))
		s.LogPrintf("executeMsg", "msg=%p pl[0]=%p", m, pl.VMs[dbs.VMIndex].List[0])
		if pl.VMs[dbs.VMIndex].List[0] != m {
			s.LogMessage("executeMsg", "drop, slot 0 taken by", pl.VMs[dbs.VMIndex].List[0])
		} else {
			s.LogMessage("executeMsg", "duplicate execute", pl.VMs[dbs.VMIndex].List[0])
		}

		return
	}

	// Put the System Height and Serial Hash into the EOM
	dbs.SysHeight = uint32(pl.System.Height)

	_, ok := s.Replay.Valid(constants.INTERNAL_REPLAY, m.GetRepeatHash().Fixed(), m.GetTimestamp(), s.GetTimestamp())
	if !ok {
		TotalHoldingQueueOutputs.Inc()
		HoldingQueueDBSigOutputs.Inc()
		//delete(s.Holding, m.GetMsgHash().Fixed())
		s.DeleteFromHolding(m.GetMsgHash().Fixed(), m, "INTERNAL_REPLAY")
		s.LogMessage("executeMsg", "drop INTERNAL_REPLAY", m)
		return
	}

	ack := s.NewAck(m, s.Balancehash).(*messages.Ack)

	m.SetLeaderChainID(ack.GetLeaderChainID())
	m.SetMinute(ack.Minute)

	s.ProcessLists.Get(ack.DBHeight).AddToProcessList(s, ack, m)
}

func (s *State) LeaderExecuteCommitChain(m interfaces.IMsg) {
	vm := s.LeaderPL.VMs[s.LeaderVMIndex]
	if len(vm.List) != vm.Height {
		s.repost(m, 1)
		return
	}
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
	vm := s.LeaderPL.VMs[s.LeaderVMIndex]
	if len(vm.List) != vm.Height {
		s.repost(m, 1)
		return
	}
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
	vm := s.LeaderPL.VMs[s.LeaderVMIndex]
	if len(vm.List) != vm.Height {
		s.repost(m, 1)
		return
	}

	ack := s.NewAck(m, nil).(*messages.Ack)

	// Debugging thing.
	m.SetLeaderChainID(ack.GetLeaderChainID())
	m.SetMinute(ack.Minute)

	// Put the acknowledgement in the Acks so we can tell if AddToProcessList() adds it.
	s.Acks[m.GetMsgHash().Fixed()] = ack
	TotalAcksInputs.Inc()
	s.ProcessLists.Get(ack.DBHeight).AddToProcessList(s, ack, m)

	// If it was not added, then handle as a follower, and leave.
	if s.Acks[m.GetMsgHash().Fixed()] != nil {
		m.FollowerExecute(s)
		return
	}

	TotalCommitsOutputs.Inc()
}

func (s *State) ProcessAddServer(dbheight uint32, addServerMsg interfaces.IMsg) bool {
	as, ok := addServerMsg.(*messages.AddServerMsg)
	if ok && !ProcessIdentityToAdminBlock(s, as.ServerChainID, as.ServerType) {
		s.LogPrintf("process", "Failed to add %x as server type %d", as.ServerChainID.Bytes()[3:6], as.ServerType)
		return true // If it fails it will never work so just move along.
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

	if e := s.GetFactoidState().UpdateECTransaction(true, c.CommitChain); e == nil {
		// save the Commit to match against the Reveal later
		h := c.GetHash()
		s.PutCommit(h, c)
		pl.EntryCreditBlock.GetBody().AddEntry(c.CommitChain)
		entry := s.Holding[h.Fixed()]
		if entry != nil {
			s.repost(entry, 0) // Try and execute the reveal for this commit
		}
		//s.LogMessage("dependentHolding", "process", commitChain)
		s.ExecuteFromHolding(commitChain.GetHash().Fixed()) // process CommitChain
		return true
	}

	//s.AddStatus("Cannot process Commit Chain")

	return false
}

func (s *State) ProcessCommitEntry(dbheight uint32, commitEntry interfaces.IMsg) bool {
	c, _ := commitEntry.(*messages.CommitEntryMsg)

	pl := s.ProcessLists.Get(dbheight)
	if e := s.GetFactoidState().UpdateECTransaction(true, c.CommitEntry); e == nil {
		// save the Commit to match against the Reveal later
		h := c.GetHash()
		s.PutCommit(h, c)
		pl.EntryCreditBlock.GetBody().AddEntry(c.CommitEntry)
		entry := s.Holding[h.Fixed()]
		if entry != nil {
			s.repost(entry, 0) // Try and execute the reveal for this commit
		}
		//		s.LogMessage("dependentHolding", "process", commitEntry)
		s.ExecuteFromHolding(commitEntry.GetHash().Fixed()) // process CommitEntry
		return true
	}
	//s.AddStatus("Cannot Process Commit Entry")

	return false
}

func (s *State) ProcessRevealEntry(dbheight uint32, m interfaces.IMsg) (worked bool) {
	pl := s.ProcessLists.Get(dbheight)
	if pl == nil {
		s.LogMessage("process", "Hold, no processlist", m)
		return false
	}

	msg := m.(*messages.RevealEntryMsg)
	myhash := msg.Entry.GetHash()
	chainID := msg.Entry.GetChainID()

	// Removed because all dependencies are met prior to adding to the process list
	//commit := s.NextCommit(msg.Entry.GetHash())
	//if commit == nil {
	//	s.LogMessage("process", "Hold, no commit", m)
	//	return false // hold for a commit
	//}

	defer func() {
		if worked {
			TotalProcessListProcesses.Inc()
			TotalCommitsOutputs.Inc()
			// This is so the api can determine if a chainhead is about to be updated. It fixes a race condition
			// on the api. MUST BE BEFORE THE REPLAY FILTER ADD
			pl.PendingChainHeads.Put(msg.Entry.GetChainID().Fixed(), msg)
			// Okay the Reveal has been recorded.  Record this as an entry that cannot be duplicated.
			s.Replay.IsTSValidAndUpdateState(constants.REVEAL_REPLAY, msg.Entry.GetHash().Fixed(), msg.Timestamp, s.GetTimestamp())
			s.Commits.Delete(msg.Entry.GetHash().Fixed()) // delete(s.Commits, msg.Entry.GetHash().Fixed())
		}
	}()

	TotalCommitsOutputs.Inc()

	eb := s.GetNewEBlocks(dbheight, chainID)
	eb_db := s.GetNewEBlocks(dbheight-1, chainID)
	if eb_db == nil {
		eb_db, _ = s.DB.FetchEBlockHead(chainID)
	}
	// Handle the case that this is a Entry Chain create
	// Must be built with CommitChain (i.e. !msg.IsEntry).  Also
	// cannot have an existing chain (eb and eb_db == nil)
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
		s.WriteEntry <- msg.Entry
		s.IncEntryChains()
		s.IncEntries()
		//		s.LogMessage("dependentHolding", "process", m)
		s.ExecuteFromHolding(chainID.Fixed()) // Process Reveal for Chain

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
	s.WriteEntry <- msg.Entry

	s.IncEntries()
	return true
}

func (s *State) CreateDBSig(dbheight uint32, vmIndex int) (interfaces.IMsg, interfaces.IMsg) {
	dbstate := s.DBStates.Get(int(dbheight - 1))
	if dbstate == nil && dbheight > 0 {
		s.LogPrintf("executeMsg", "CreateDBSig:Can not create DBSig because %d because there is no dbstate", dbheight)
		return nil, nil
	}
	dbs := new(messages.DirectoryBlockSignature)
	dbs.DirectoryBlockHeader = dbstate.DirectoryBlock.GetHeader()
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

	s.LogMessage("dbstateprocess", "CreateDBSig", dbs)

	return dbs, ack
}

// dbheight is the height of the process list, and vmIndex is the vm
// that is missing the DBSig.  If the DBSig isn't our responsibility, then
// this call will do nothing.  Assumes the state for the leader is set properly
func (s *State) SendDBSig(dbheight uint32, vmIndex int) {
	s.LogPrintf("executeMsg", "SendDBSig(dbht=%d,vm=%d)", dbheight, vmIndex)
	dbslog := consenLogger.WithFields(log.Fields{"func": "SendDBSig"})

	ht := s.GetHighestSavedBlk()
	if dbheight <= ht { // if it's in the past, just return.
		return
	}
	if s.CurrentMinute != 0 {
		s.LogPrintf("executeMsg", "SendDBSig(%d,%d) Only generate DBSig in minute 0 @ %s", dbheight, vmIndex, atomic.WhereAmIString(1))
		return
	}
	pl := s.ProcessLists.Get(dbheight)
	vm := pl.VMs[vmIndex]
	if vm.Height > 0 {
		s.LogPrintf("executeMsg", "SendDBSig(%d,%d) I already have processed a DBSig in this VM @ %s", dbheight, vmIndex, atomic.WhereAmIString(1))
		return // If we already have the DBSIG (it's always in slot 0) then just return
	}
	leader, lvm := pl.GetVirtualServers(vm.LeaderMinute, s.IdentityChainID)
	if !leader || lvm != vmIndex {
		s.LogPrintf("executeMsg", "SendDBSig(%d,%d) Caller lied to me about VMIndex @ %s", dbheight, vmIndex, atomic.WhereAmIString(1))
		return // If I'm not a leader or this is not my VM then return
	}

	if !vm.Signed {

		if !pl.DBSigAlreadySent {

			dbs, _ := s.CreateDBSig(dbheight, vmIndex)
			if dbs == nil {
				return
			}

			dbslog.WithFields(dbs.LogFields()).WithFields(log.Fields{"lheight": s.GetLeaderHeight(), "node-name": s.GetFactomNodeName()}).Infof("Generate DBSig")
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
			ids := s.GetUnsyncedServersString()
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
		return false
	}

	//todo: Why isn't this CurrentMinute?
	if s.EOM && int(e.Minute) > s.EOMMinute {
		//fmt.Println(fmt.Sprintf("SigType PROCESS: %10s vm %2d Will Not Process: return on s.SigType(%v) && int(e.Minute(%v)) > s.EOMMinute(%v)", s.FactomNodeName, e.VMIndex, s.SigType, e.Minute, s.EOMMinute))
		s.LogPrintf("dbsig-eom", "ProcessEOM skip EOM for a future minute e.Minute(%d) > s.EOMMinute(%d)", e.Minute, s.EOMMinute)
		return false
	}

	if s.CurrentMinute == 0 && !s.DBSigDone {
		s.LogPrintf("dbsig-eom", "ProcessEOM wait for DBSig in minute 0")
		return false
	}

	s.LogMessage("dbsig-eom", "ProcessEOM ", msg)

	// If I have done everything for all EOMs for all VMs, then and only then do I
	// let processing continue.
	if s.EOMDone {
		s.LogPrintf("dbsig-eom", "ProcessEOM finalize EOM processing")

		dbstate := s.GetDBState(dbheight - 1)
		if dbstate == nil {
			//fmt.Println(fmt.Sprintf("SigType PROCESS: %10s vm %2d DBState == nil: return on s.SigType(%v) && int(e.Minute(%v)) > s.EOMMinute(%v)", s.FactomNodeName, e.VMIndex, s.SigType, e.Minute, s.EOMMinute))
			s.LogPrintf("dbsig-eom", "ProcessEOM wait prev dbstate == nil")
			return false
		}
		if !dbstate.Saved && s.CurrentMinute > 0 {
			//fmt.Println(fmt.Sprintf("SigType PROCESS: %10s vm %2d DBState not saved: return on s.SigType(%v) && int(e.Minute(%v)) > s.EOMMinute(%v)", s.FactomNodeName, e.VMIndex, s.SigType, e.Minute, s.EOMMinute))
			s.LogPrintf("dbsig-eom", "ProcessEOM wait prev !dbstate.Saved")
			return false
		}

		//fmt.Println(fmt.Sprintf("EOM PROCESS: %10s vm %2d Done! s.EOMDone(%v) && s.EOMSys(%v)", s.FactomNodeName, e.VMIndex, s.EOMDone, s.EOMSys))
		s.EOMProcessed--
		if s.EOMProcessed <= 0 { // why less than or equal?
			s.SendHeartBeat() // Only do this once per minute
			s.LogPrintf("dbsig-eom", "ProcessEOM complete for %d", e.Minute)
			// setup to sync next minute ...
			s.Syncing = false  // ProcessEOM (EOM complete)
			s.EOM = false      // ProcessEOM (EOM complete)
			s.EOMDone = false  // ProcessEOM (EOM complete)
			s.EOMProcessed = 0 // ProcessEOM (EOM complete)
			for _, vm := range pl.VMs {
				vm.Synced = false // ProcessEOM (EOM complete)
			}
			if !s.Leader {
				if s.CurrentMinute != int(e.Minute) {
					s.LogPrintf("dbsig-eom", "Follower jump to minute %d from %d", s.CurrentMinute, int(e.Minute))
				}
				s.MoveStateToHeight(e.DBHeight, int(e.Minute+1))

			} else {
				s.MoveStateToHeight(s.LLeaderHeight, s.CurrentMinute+1)
			}

			switch {
			case s.CurrentMinute < 10:
				if s.CurrentMinute == 1 {
					// Panic had arose when leaders would reboot and the follower was on a future minute
					if dbstate == nil {
						// We recognize that this will leave us "Done" without finishing the process.  But
						// a Follower can heal themselves by asking for a block, and overwriting this block.
						return false
					}
				}
				LeaderPL := s.ProcessLists.Get(s.LLeaderHeight)
				if s.LeaderPL != LeaderPL {
					s.LogPrintf("ExecuteMsg", "ProcessEOM: Unexpected change in LeaderPL")
					s.LeaderPL = LeaderPL
				}

				Leader, LeaderVMIndex := s.LeaderPL.GetVirtualServers(s.CurrentMinute, s.IdentityChainID)
				{ // debug
					if s.Leader != Leader {
						s.LogPrintf("executeMsg", "State.ProcessEOM() unexpectedly setting s.Leader to %v", Leader)
						s.Leader = Leader
					}
					if s.LeaderVMIndex != LeaderVMIndex {
						s.LogPrintf("executeMsg", "State.ProcessEOM()  unexpectedly setting s.LeaderVMIndex to %v", LeaderVMIndex)
						s.LeaderVMIndex = LeaderVMIndex
					}
				}

			case s.CurrentMinute == 10:
				s.LogPrintf("dbsig-eom", "Start new block")
				eBlocks := []interfaces.IEntryBlock{}
				entries := []interfaces.IEBEntry{}
				for _, v := range pl.NewEBlocks {
					eBlocks = append(eBlocks, v)
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

				s.DBSigProcessed = 0
				s.TempBalanceHash = s.FactoidState.GetBalanceHash(true)

				s.Saving = true
			}

			s.Commits.RemoveExpired(s)

			for k := range s.Acks {
				v := s.Acks[k].(*messages.Ack)
				if v.DBHeight < s.LLeaderHeight {
					TotalAcksOutputs.Inc()
					delete(s.Acks, k)
					s.LogMessage("executeMsg", "Drop, expired", v)
				}
			}
			//fmt.Println(fmt.Sprintf("SigType PROCESS: %10s vm %2d Saving: return on s.SigType(%v) && int(e.Minute(%v)) > s.EOMMinute(%v)", s.FactomNodeName, e.VMIndex, s.SigType, e.Minute, s.EOMMinute))
		}
		return true
	}

	// What I do once  for all VMs at the beginning of processing a particular EOM
	if !s.EOM {
		s.LogPrintf("dbsig-eom", "ProcessEOM start EOM processing for %d", e.Minute)

		//fmt.Println(fmt.Sprintf("SigType PROCESS: %10s vm %2d Start SigType Processing: !s.SigType(%v) SigType: %s", s.FactomNodeName, e.VMIndex, s.SigType, e.String()))
		s.Syncing = true   // ProcessEOM start
		s.EOM = true       // ProcessEOM start
		s.EOMDone = false  // ProcessEOM start
		s.EOMProcessed = 0 // ProcessEOM start
		s.EOMLimit = len(pl.FedServers)
		for _, vm := range pl.VMs {
			vm.Synced = false // ProcessEOM start
		}
		if s.CurrentMinute != int(e.Minute) {
			s.LogPrintf("dbsig-eom", "Follower jump to minute %d from %d", s.CurrentMinute, int(e.Minute))
		}
		s.EOMMinute = int(e.Minute)
		s.EOMsyncing = true
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
		vm.Synced = true // ProcessEOM
		markNoFault(pl, msg.GetVMIndex())
		//fmt.Println(fmt.Sprintf("SigType PROCESS: %10s vm %2d Process this SigType: return on s.SigType(%v) && int(e.Minute(%v)) > s.EOMMinute(%v)", s.FactomNodeName, e.VMIndex, s.SigType, e.Minute, s.EOMMinute))
		return false
	}

	// After all EOM markers are processed, Claim we are done.  Now we can unwind

	if s.EOMProcessed == s.EOMLimit && !s.EOMDone {
		s.LogPrintf("dbsig-eom", "ProcessEOM stop EOM processing minute %d", s.CurrentMinute)

		//fmt.Println(fmt.Sprintf("SigType PROCESS: SigType Complete: %10s vm %2d allfaults(%v) && s.EOMProcessed(%v) == s.EOMLimit(%v) && !s.EOMDone(%v)",
		//	s.FactomNodeName,
		//	e.VMIndex, allfaults, s.EOMProcessed, s.EOMLimit, s.EOMDone))

		s.EOMDone = true // ProcessEOM
		s.EOMSyncTime = time.Now().UnixNano()
		for _, eb := range pl.NewEBlocks {
			eb.AddEndOfMinuteMarker(byte(e.Minute + 1))
		}

		s.FactoidState.EndOfPeriod(int(e.Minute))

		ecblk := pl.EntryCreditBlock
		ecbody := ecblk.GetBody()
		mn := entryCreditBlock.NewMinuteNumber(e.Minute + 1)
		ecbody.AddEntry(mn)

	} else {
		//fmt.Println(fmt.Sprintf("SigType PROCESS: %10s vm %2d Do nothing: return on s.SigType(%v) && int(e.Minute(%v)) > s.EOMMinute(%v)", s.FactomNodeName, e.VMIndex, s.SigType, e.Minute, s.EOMMinute))
	}

	return false
}

// GetUnsyncedServers returns an array of the IDs for all unsynced VMs
// when you are not in a sync phase, no VM is considered sync'd
func (s *State) GetUnsyncedServers() (ids []interfaces.IHash, vms []int) {
	p := s.LeaderPL
	c := s.CurrentMinute
	if c == 10 {
		return ids, vms // Must be all sync'd in minute 10
	}
	for index, l := range p.FedServers {
		vmIndex := FedServerVM(p.ServerMap, len(p.FedServers), c, index)
		vm := p.VMs[vmIndex]
		if !vm.Synced {
			ids = append(ids, l.GetChainID())
			vms = append(vms, vmIndex)
		}
	}
	return ids, vms
}

// GetUnsyncedServersString returns a string with the short IDs for all unsynced VMs
func (s *State) GetUnsyncedServersString() string {
	var ids string
	servers, vms := s.GetUnsyncedServers()
	for index, id := range servers {
		ids = ids + "," + id.String()[6:12] + fmt.Sprintf("[%d]{%d}", index, vms[index])
	}
	if len(ids) > 0 {
		ids = ids[1:] // drop the leading comma
	}
	return ids
}

func (s *State) CheckForIDChange() {
	changed, _ := s.GetAckChange()
	if changed {
		s.LogPrintf("AckChange", "AckChange %v", s.AckChange)
	}
	if s.LLeaderHeight == s.AckChange {
		config := util.ReadConfig(s.ConfigFilePath)
		var err error
		prev_ChainID := s.IdentityChainID
		prev_LocalServerPrivKey := s.LocalServerPrivKey
		s.IdentityChainID, err = primitives.NewShaHashFromStr(config.App.IdentityChainID)
		if err != nil {
			panic(err)
		}
		s.LocalServerPrivKey = config.App.LocalServerPrivKey
		s.initServerKeys()
		s.LogPrintf("AckChange", "ReloadIdentity new local_priv: %v ident_chain: %v, prev local_priv: %v ident_chain: %v", s.LocalServerPrivKey, s.IdentityChainID, prev_LocalServerPrivKey, prev_ChainID)
	}
}

// When we process the directory Signature, and we are the leader for said signature, it
// is then that we push it out to the rest of the network.  Otherwise, if we are not the
// leader for the signature, it marks the sig complete for that list
func (s *State) ProcessDBSig(dbheight uint32, msg interfaces.IMsg) bool {
	//fmt.Println(fmt.Sprintf("ProcessDBSig: %10s %s ", s.FactomNodeName, msg.String()))

	// Avoid a race where we try to process DBSig for VM0 before the factoidState is setup.
	if msg.(*messages.DirectoryBlockSignature).VMIndex == 0 && s.FactoidState == nil { // can't process till factoid state is setup
		return false // fix panic in TestMultipleElection7
	}

	dbs := msg.(*messages.DirectoryBlockSignature)
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
	s.LogPrintf("dbsig-eom", "ProcessDBSig@%d/%d/%d minute %d, Syncing %v , DBSig %v, DBSigDone %v, DBSigProcessed %v, DBSigLimit %v DBSigDone %v",
		dbheight, msg.GetVMIndex(), len(vm.List), s.CurrentMinute, s.Syncing, s.DBSig, s.DBSigDone, s.DBSigProcessed, s.DBSigLimit, s.DBSigDone)

	// debug
	if s.DebugExec() {
		if s.Syncing && s.DBSig && !s.DBSigDone {
			ids := s.GetUnsyncedServersString()
			if len(ids) > 0 {
				s.LogPrintf("dbsig-eom", "Waiting for DBSigs from %s", ids)
			}
		}
	}

	s.LogMessage("dbsig-eom", "ProcessDBSig ", msg)
	// If we are done with DBSigs, and this message is processed, then we are done.  Let everything go!
	if s.DBSig && s.DBSigDone {
		s.LogPrintf("dbsig-eom", "ProcessDBSig finalize DBSig processing")
		//fmt.Println(fmt.Sprintf("ProcessDBSig(): %10s Finished with DBSig: s.DBSigSys(%v) && s.DBSig(%v) && s.DBSigDone(%v)", s.FactomNodeName, s.DBSigSys, s.DBSig, s.DBSigDone))
		s.DBSigProcessed--
		if s.DBSigProcessed <= 0 {
			s.Syncing = false //ProcessDBSig done
			s.DBSig = false   //ProcessDBSig done
			for _, vm := range pl.VMs {
				vm.Synced = false // ProcessDBSig finalize
			}
			s.LogPrintf("dbsig-eom", "ProcessDBSig complete for %d", dbs.Minute)
		} else {
			vm.Signed = true
		}
		//s.LeaderPL.AdminBlock
		return true
	}

	// Put the stuff that only executes once at the start of DBSignatures here
	if !s.DBSig {
		s.LogPrintf("dbsig-eom", "ProcessDBSig start DBSig processing for %d", dbs.Minute)

		//fmt.Printf("ProcessDBSig(): %s Start DBSig %s\n", s.FactomNodeName, dbs.String())
		s.Syncing = true     // ProcessDBsig Start
		s.DBSig = true       // ProcessDBsig Start
		s.DBSigDone = false  // ProcessDBsig Start
		s.DBSigProcessed = 0 // ProcessDBsig Start
		s.DBSigLimit = len(pl.FedServers)
		for _, vm := range pl.VMs {
			vm.Synced = false // ProcessDBsig Start
		}
		pl.ResetDiffSigTally()
	}

	// Put the stuff that executes per DBSignature here
	if !vm.Synced {
		s.LogPrintf("dbsig-eom", "ProcessDBSig Handle VM(%v) minute %d", msg.GetVMIndex(), dbs.Minute)

		highestCompletedBlk := s.GetHighestCompletedBlk()
		if s.LLeaderHeight > 0 && highestCompletedBlk+1 < s.LLeaderHeight {

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
			dbsMilli := dbs.Timestamp.GetTimeMilliUInt64()
			fs := s.FactoidState.(*FactoidState)
			s.LogPrintf("dbsig", "1st ProcessDBSig(): %10s DBSig dbht %d leaderheight %d VMIndex %d Timestamp %x %d, leadertimestamp = %x %d",
				s.FactomNodeName, dbs.DBHeight, s.LLeaderHeight, dbs.VMIndex, dbs.GetTimestamp().GetTimeMilli(), dbs.GetTimestamp().GetTimeMilli(), s.LeaderTimestamp.GetTimeMilliUInt64(), s.LeaderTimestamp.GetTimeMilliUInt64())

			cbtx := fs.GetCurrentBlock().(*factoid.FBlock).Transactions[0].(*factoid.Transaction)
			foo := cbtx.MilliTimestamp
			lts := s.LeaderTimestamp.GetTimeMilliUInt64()
			s.LogPrintf("dbsig", "ProcessDBSig(): first  cbtx before %d dbsig %d lts %d", foo, dbsMilli, lts)

			s.SetLeaderTimestamp(dbs.Timestamp) // SetLeaderTimestamp also updates the Message Timestamp filter

			uInt64_3 := dbs.GetTimestamp().GetTimeMilliUInt64()
			foo_3 := cbtx.MilliTimestamp
			lts_3 := s.LeaderTimestamp.GetTimeMilliUInt64()
			s.LogPrintf("dbsig", "ProcessDBSig(): second cbtx before %d dbsig %d lts %d", foo_3, uInt64_3, lts_3)
			s.LogPrintf("dbsig", "ProcessDBSig(): p cbtx %p dbsig %p lts %p", cbtx.GetTimestamp().(*primitives.Timestamp), dbs.GetTimestamp().(*primitives.Timestamp), s.LeaderTimestamp.(*primitives.Timestamp))

			txt, _ := cbtx.CustomMarshalText()
			s.LogPrintf("dbsig", "ProcessDBSig(): coinbase before %s", string(txt))

			uInt64 := dbs.GetTimestamp().GetTimeMilliUInt64()

			foo2 := cbtx.MilliTimestamp
			cbtx.MilliTimestamp = dbsMilli
			s.LogPrintf("dbsig", "ProcessDBSig(): cbtx before %d dbsig %d cbtx after %d", foo2, uInt64, cbtx.MilliTimestamp)

			txt, _ = cbtx.CustomMarshalText()
			s.LogPrintf("dbsig", "ProcessDBSig(): coinbase after  %s", string(txt))

			s.LogPrintf("dbsig", "ProcessDBSig(): 2nd ProcessDBSig(): %10s DBSig dbht %d leaderheight %d VMIndex %d Timestamp %x %d, leadertimestamp = %x %d",
				s.FactomNodeName, dbs.DBHeight, s.LLeaderHeight, dbs.VMIndex, dbs.GetTimestamp().GetTimeMilli(), dbs.GetTimestamp().GetTimeMilli(), s.LeaderTimestamp.GetTimeMilliUInt64(), s.LeaderTimestamp.GetTimeMilliUInt64())
		}

		dblk, err := s.DB.FetchDBlockByHeight(dbheight - 1)
		if dblk != nil {
			hashes := dblk.GetEntryHashes()
			if hashes != nil {
				messages.LogPrintf("marshalsizes.txt", "DirectoryBlock unmarshaled entry count: %d", len(hashes))
			}
		}
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
			s.LogPrintf("processList", "Failed. DBSig and DBlocks do not match Expected-Body-Mr: [%d]%x, Got: [%d]%x",
				dblk.GetHeader().GetDBHeight(), dblk.GetHeader().GetBodyMR().Fixed(), dbs.DirectoryBlockHeader.GetDBHeight(), dbs.DirectoryBlockHeader.GetBodyMR().Fixed())

			// If the Directory block hash doesn't work for me, then the dbsig doesn't work for me, so
			// toss it and ask our neighbors for another one.
			s.LogMessage("processList", "drop from pl", vm.List[0])
			s.LogMessage("processList", "drop from pl", vm.ListAck[0])
			vm.ListAck[0] = nil
			vm.List[0] = nil
			vm.HighestAsk = -1
			vm.HighestNil = 0
			return false
		}

		// Adds DB Sig to be added to Admin block if passes sig checks
		data, err := dbs.DirectoryBlockHeader.MarshalBinary()
		if err != nil {
			return false
		}
		if !dbs.DBSignature.Verify(data) {
			s.LogPrintf("processList", "Failed. DBSig.DBSignature.Verify()")
			// If the signature fails, then ask for another one.
			s.LogMessage("processList", "drop from pl", vm.List[0])
			s.LogMessage("processList", "drop from pl", vm.ListAck[0])
			vm.ListAck[0] = nil
			vm.List[0] = nil
			vm.HighestAsk = -1
			vm.HighestNil = 0
			return false
		}

		valid, err := s.FastVerifyAuthoritySignature(data, dbs.DBSignature, dbs.DBHeight)
		if err != nil || valid != 1 {
			s.LogPrintf("processList", "Failed. DBSig Invalid Auth Sig: Pubkey: %x", dbs.Signature.GetKey())

			// If the authority is bad, toss this signature and ask for another.
			s.LogMessage("processList", "drop from pl", vm.List[0])
			s.LogMessage("processList", "drop from pl", vm.ListAck[0])
			vm.ListAck[0] = nil
			vm.List[0] = nil
			vm.HighestAsk = -1
			vm.HighestNil = 0
			return false
		}

		dbs.Matches = true
		s.AddDBSig(dbheight, dbs.ServerIdentityChainID, dbs.DBSignature)

		s.DBSigProcessed++
		//fmt.Println(fmt.Sprintf("Process DBSig %10s vm %2v DBSigProcessed++ (%2d)", s.FactomNodeName, dbs.VMIndex, s.DBSigProcessed))
		vm.Synced = true // ProcessDBsig

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

	// Put the stuff that executes once for set of DBSignatures (after I have them all) here
	if !s.DBSigDone && s.DBSigProcessed >= s.DBSigLimit {
		s.LogPrintf("dbsig-eom", "ProcessDBSig stop DBSig processing minute %d", s.CurrentMinute)
		//fmt.Println(fmt.Sprintf("All DBSigs are processed: allfaults(%v), && !s.DBSigDone(%v) && s.DBSigProcessed(%v)>= s.DBSigLimit(%v)",
		//	allfaults, s.DBSigDone, s.DBSigProcessed, s.DBSigLimit))
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
	return s.highestAck
}

func (s *State) SetHighestAck(dbht uint32) {
	switch {
	case dbht > s.highestAck + 200 :
		s.highestAck = s.highestAck + 200
		break
	case dbht > s.highestAck:
		s.highestAck = dbht
	}
}

// This is the highest block signed off and recorded in the Database.
func (s *State) GetHighestSavedBlk() uint32 {
	v := s.DBStates.GetHighestSavedBlk()
	HighestSaved.Set(float64(v))
	return v
}

func (s *State) GetDBHeightAtBoot() uint32 {
	return s.DBHeightAtBoot
}

// This is the highest block signed off, but not necessarily validated.
func (s *State) GetHighestCompletedBlk() uint32 {
	v := s.DBStates.GetHighestCompletedBlk()
	HighestCompleted.Set(float64(v))
	return v
}

func (s *State) GetHighestLockedSignedAndSavesBlk() uint32 {
	v := s.DBStates.GetHighestLockedSignedAndSavesBlk()
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
	HighestKnown.Set(float64(s.highestKnown))
	return s.highestKnown
}

func (s *State) SetHighestKnownBlock( dbht uint32 ){
	switch {
	case dbht > s.highestKnown + 200 :
		s.highestKnown = s.highestKnown + 200
		break
	case dbht > s.highestKnown:
		s.highestKnown = dbht
	}
}

// GetF()
// If rt == true, read the Temp balances.  Otherwise read the Permanent balances.
// concurrency safe to call
func (s *State) GetF(rt bool, adr [32]byte) (v int64) {
	ok := false
	if rt {
		pl := s.ProcessLists.Get(s.LLeaderHeight)
		if pl != nil {
			pl.FactoidBalancesTMutex.Lock()
			v, ok = pl.FactoidBalancesT[adr]
			pl.FactoidBalancesTMutex.Unlock()
		}
	}
	if !ok {
		s.FactoidBalancesPMutex.Lock()
		v = s.FactoidBalancesP[adr]
		s.FactoidBalancesPMutex.Unlock()
	}

	return v

}

// PutF()
// If rt == true, update the Temp balances.  Otherwise update the Permanent balances.
// concurrency safe to call
func (s *State) PutF(rt bool, adr [32]byte, v int64) {
	if rt {
		pl := s.ProcessLists.Get(s.LLeaderHeight)
		if pl != nil {
			pl.FactoidBalancesTMutex.Lock()
			pl.FactoidBalancesT[adr] = v
			pl.FactoidBalancesTMutex.Unlock()
		}
	} else {
		s.FactoidBalancesPMutex.Lock()
		s.FactoidBalancesP[adr] = v
		s.FactoidBalancesPMutex.Unlock()
	}
}

func (s *State) GetE(rt bool, adr [32]byte) (v int64) {
	ok := false
	if rt {
		pl := s.ProcessLists.Get(s.LLeaderHeight)
		if pl != nil {
			pl.ECBalancesTMutex.Lock()
			v, ok = pl.ECBalancesT[adr]
			pl.ECBalancesTMutex.Unlock()
		}
	}
	if !ok {
		s.ECBalancesPMutex.Lock()
		v = s.ECBalancesP[adr]
		s.ECBalancesPMutex.Unlock()
	}
	return v

}

// PutE()
// If rt == true, update the Temp balances.  Otherwise update the Permanent balances.
// concurrency safe to call
func (s *State) PutE(rt bool, adr [32]byte, v int64) {
	if rt {
		pl := s.ProcessLists.Get(s.LLeaderHeight)
		if pl != nil {
			pl.ECBalancesTMutex.Lock()
			pl.ECBalancesT[adr] = v
			pl.ECBalancesTMutex.Unlock()
		}
	} else {
		s.ECBalancesPMutex.Lock()
		s.ECBalancesP[adr] = v
		s.ECBalancesPMutex.Unlock()
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

func (s *State) GetNewHash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("State.GetNewHash() saw an interface that was nil")
		}
	}()
	return new(primitives.Hash)
}

// Create a new Acknowledgement.  Must be called by a leader.  This
// call assumes all the pieces are in place to create a new acknowledgement
func (s *State) NewAck(msg interfaces.IMsg, balanceHash interfaces.IHash) interfaces.IMsg {

	vmIndex := msg.GetVMIndex()
	leaderMinute := byte(s.ProcessLists.Get(s.LLeaderHeight).VMs[vmIndex].LeaderMinute)

	// these don't affect the msg hash, just for local use...
	msg.SetLeaderChainID(s.IdentityChainID)
	ack := new(messages.Ack)
	ack.DBHeight = s.LLeaderHeight
	ack.VMIndex = vmIndex
	ack.Minute = leaderMinute
	ack.Timestamp = s.GetTimestamp()
	ack.SaltNumber = s.GetSalt(ack.Timestamp)
	copy(ack.Salt[:8], s.Salt.Bytes()[:8])
	ack.MessageHash = msg.GetMsgHash()
	ack.LeaderChainID = s.IdentityChainID
	ack.BalanceHash = balanceHash
	listlen := s.LeaderPL.VMs[vmIndex].Height
	if listlen == 0 {
		ack.Height = 0
		ack.SerialHash = ack.MessageHash
	} else {
		last := s.LeaderPL.GetAckAt(vmIndex, listlen-1)
		ack.Height = last.Height + 1
		ack.SerialHash, _ = primitives.CreateHash(last.MessageHash, ack.MessageHash)
	}

	ack.Sign(s)
	ack.SetLocal(true)

	return ack
}
