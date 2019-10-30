package state

import (
	"fmt"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
)

// NonBlockingChannelAdd will only add to the channel if the action is non-blocking
func NonBlockingChannelAdd(channel chan interfaces.IMsg, msg interfaces.IMsg) bool {
	select {
	case channel <- msg:
		return true
	default:
		return false
	}
}

// This identifies a specific process list slot
type plRef struct {
	DBH int
	VM  int
	H   int
}

// This is when to next ask for a particular request
type askRef struct {
	plRef
	When int64 // in timestamp ms
}

type MSgPair struct {
	msg, ask interfaces.IMsg
}

type MMRInfo struct {
	// Channels for managing the missing message requests
	asks      chan askRef  // Requests to ask for missing messages
	adds      chan plRef   // notices of slots filled in the process list
	rejects   chan MsgPair // Messages rejected from process list
	dbheights chan int     // Notice that this DBHeight is done
}

// starts the MMR processing for this state
func (s *State) StartMMR() {
	// Missing message request handling.
	s.makeMMRs(s.asks, s.adds, s.dbheights, s.rejects)
}

// MMRDummy is for unit tests that populate the various mmr queues.
// We need to drain the queues to ensure we don't block.
//	ONLY FOR UNIT TESTS
func (s *State) MMRDummy() {
	go func() {
		for {
			select {
			case <-s.asks:
			case <-s.adds:
			case <-s.dbheights:
			case <-s.rejects:
			}
		}
	}()
}

// Ask VM for an MMR for this height with delay ms before asking the network
// called from validation thread to notify MMR that we are missing a message
func (vm *VM) ReportMissing(height int, delay int64) {
	vm.p.State.LogPrintf("missing_messages", "ReportMissing %d/%d/%d, delay %d", vm.p.DBHeight, vm.VmIndex, height, delay)

	now := vm.p.State.GetTimestamp().GetTimeMilli()
	oneSeconds := vm.p.State.FactomSecond()
	if delay < oneSeconds.Milliseconds() {
		delay = oneSeconds.Milliseconds() // Floor for delays is 1 second so there is time to merge adjacent requests
	}
	lenVMList := len(vm.List)
	// ask for all missing messages
	var i int
	for i = vm.Height; i < lenVMList; i++ {
		if i < 0 { // -1 is the default highestask, as we have not asked yet. Obviously this index does not exist
			continue
		}
		if vm.List[i] == nil {
			ok := vm.p.State.Ask(int(vm.p.DBHeight), vm.VmIndex, i, now+delay) // send it to the MMR thread
			if !ok {
				return // If we can't ask for one then don't try the next or highest ask might be set wrong
			}
		}
	}

	// if we are asking above the current list
	if height >= lenVMList {
		vm.p.State.Ask(int(vm.p.DBHeight), vm.VmIndex, height, now+delay) // send it to the MMR thread
	}

}

// Ask is called from ReportMissing which comes from validation thread to notify MMR that we are missing a message
// return false if we are unable to ask
func (s *State) Ask(DBHeight int, vmIndex int, height int, when int64) bool {
	if s.asks == nil { // If it is nil, there is no makemmrs
		return false
	}
	// do not ask for things in the past or very far into the future
	if DBHeight < int(s.LLeaderHeight) || DBHeight > int(s.LLeaderHeight)+1 || DBHeight < int(s.DBHeightAtBoot) {
		return false
	}
	vm := s.LeaderPL.VMs[vmIndex]

	//	Currently if the asks are full, we'd rather just skip
	//	than block the thread. We report missing multiple times, so if
	//	we exit, we will come around and ask again.
	// We have to do this because MMR can provide messages from inmsgqueue by pushing them into msg queue
	// if msgqueue is full then the two threads can deadlock

	if len(vm.p.State.asks) == cap(vm.p.State.asks) {
		vm.p.State.LogPrintf("missing_messages", "drop, asks full %d/%d/%d", vm.p.DBHeight, vm.VmIndex, height)
		return false
	}

	ask := askRef{plRef{DBHeight, vmIndex, height}, when}
	s.asks <- ask
	return true
}

// Used by debug code only
var MMR_enable bool = true

// Receive all asks and all process list adds and create missing message requests any ask that has expired
// and still pending. Add 10 seconds to the ask.
func (s *State) makeMMRs(asks <-chan askRef, adds <-chan plRef, dbheights <-chan int, rejects <-chan MsgPair) {
	type dbhvm struct {
		dbh int
		vm  int
	}

	// Postpone asking for the first 5 seconds so simulations get a chance to get started. Doesn't break things but
	// there is a flurry of unhelpful MMR activity on start up of simulations with followers
	time.Sleep(5 * time.Second)

	var dbheight int // current process list height

	pending := make(map[plRef]*int64)
	ticker := make(chan int64, 50)               // this should deep enough you know that the reading thread is dead if it fills up
	mmrs := make(map[dbhvm]*messages.MissingMsg) // an MMR per DBH/VM
	logname := "missing_messages"
	var now int64

	// Delete any pending ask for a message that was just added to the processlist
	deletePendingAsk := func(add plRef) {
		delete(pending, add) // Delete request that was just added to the process list in the map
		s.LogPrintf(logname, "Add %d/%d/%d %d", add.DBH, add.VM, add.H, len(pending))
	}

	s.LogPrintf(logname, "Start MMR Process")

	// Add an ask to the list of pending asks
	addAsk := func(ask askRef) {
		_, ok := pending[ask.plRef]
		if !ok {
			//fmt.Println("pending[ask.plRef]: ", ok)
			when := ask.When
			pending[ask.plRef] = &when // add the requests to the map
			s.LogPrintf(logname, "Ask %d/%d/%d %d", ask.DBH, ask.VM, ask.H, len(pending))

			// checking if we already have the "missing" message in our maps
			ack, msg := s.RecentMessage.GetAckAndMsg(ask.DBH, ask.VM, ask.H, s)
			if msg != nil && ack != nil {
				// send them to be executed
				s.LogPrintf("mmr", "Found Ask %d/%d/%d. Adding to queues: Msg %d:%d Ack %d:%d Add %d:%d Ask %d:%d", ask.DBH, ask.VM, ask.H, len(s.msgQueue), cap(s.msgQueue), len(s.ackQueue), cap(s.ackQueue), len(s.adds), cap(s.adds), len(s.asks), cap(s.asks))

				// Attempt to add the msg and ack to the prioritized message queue without blocking.
				// If we end up dropping this message, there isn't much we can do without potentially blocking
				// our asks and adds queue.
				// If the s.adds/s.asks queue is backed up, we cannot block here, or the validator loop will be stalled.
				successfulMsg := NonBlockingChannelAdd(s.PrioritizedMsgQueue(), msg)
				ml, mc := len(s.PrioritizedMsgQueue()), cap(s.PrioritizedMsgQueue())

				// Only add the ack if the msg was successful
				successfulAck := !successfulMsg || NonBlockingChannelAdd(s.PrioritizedMsgQueue(), ack)
				al, ac := len(s.PrioritizedMsgQueue()), cap(s.PrioritizedMsgQueue())

				{ // Logging/Debugging
					if successfulMsg {
						s.LogMessage("PrioritizedMsgQueue",
							fmt.Sprintf("enqueue msg makeMMRs_addAsk, PQ %d:%d", ml, mc), msg)
					} else {
						s.LogMessage("PrioritizedMsgQueue",
							fmt.Sprintf("dropped msg makeMMRs_addAsk, PQ %d:%d", ml, mc), msg)
					}
					if successfulMsg && successfulAck { // If the message was not successful, neither was the ack.
						s.LogMessage("PrioritizedMsgQueue",
							fmt.Sprintf("enqueue ack makeMMRs_addAsk, PQ %d:%d", al, ac), msg)
					} else {
						s.LogMessage("PrioritizedMsgQueue",
							fmt.Sprintf("dropped ack makeMMRs_addAsk, PQ %d:%d", al, ac), msg)
					}
				}
			}
		} // don't update the when if it already existed...
	}

	// process all pending asks
	addAllAsks := func() {
		s.LogPrintf("mmr", "AddAllAsks %d asks", len(asks))
	readasks:
		for {
			select {
			case ask := <-asks:
				addAsk(ask)
			default:
				break readasks
			}
		} // process all pending asks before any adds
	}

	addAllAdds := func() {
	readadds:
		for {
			select {
			case add := <-adds:
				deletePendingAsk(add)
			default:
				break readadds
			}
		} // process all pending add before any ticks
	}

	// drain the ticker channel
	readAllTickers := func() {
	readalltickers:
		for {
			select {
			case <-ticker:
			default:
				break readalltickers
			}
		} // process all pending add before any ticks
	}

	// tick every "factom second" to check the  pending MMRs
	go func() {
		for {
			if s.RunState.IsTerminating() {
				return // Factomd is stopping/stopped
			}

			if len(ticker) == cap(ticker) {
				// If we add to the ticker, we will block forever, so just sleep
				// and continue. If factomd is stopped, we will catch this on the continue
				time.Sleep(1 * time.Second)
				s.LogPrintf("mmr", "Ticker queue maxed, %d/%d", len(ticker), cap(ticker))
				continue
			}

			ticker <- s.GetTimestamp().GetTimeMilli()
			askDelay := s.FactomSecond() * 10    // 1/6 of a minute
			if askDelay < time.Millisecond*500 { // Don't go below 500ms. That is just too much
				askDelay = time.Millisecond * 500
			}

			time.Sleep(askDelay)
		}
	}()

	lastAskDelay := int64(0)
	for {
		// You have to compute this at every cycle as you can change the block time in sim control.

		// Take 1/6 of 1 minute boundary (DBlock is 10*min)
		//		This means on 10min block, 10 second delay
		//					  1min block, 1 second delay
		askDelay := int64((s.FactomSecond() * 10).Seconds()) * 1000

		if askDelay < 500 { // Don't go below 500ms. That is just too much
			askDelay = 500
		}

		if askDelay != lastAskDelay {
			s.LogPrintf(logname, "AskDelay %d BlockTime %d", askDelay, s.DirectoryBlockInSeconds)
			lastAskDelay = askDelay
		}

		// process any incoming messages
		select {
		case msgPair := <-rejects:
			s.LogMessage("mmr", "Reject", msgPair.Ack)
			s.RecentMessage.HandleRejection(msgPair.Msg, msgPair.Ack)
		case msg := <-s.RecentMessage.NewMsgs:
			s.LogPrintf("mmr", "start msg handling")
			s.RecentMessage.Add(msg) // adds messages to a message map for MMR

		case dbheight = <-dbheights:
			s.LogPrintf("mmr", "start dbheight handling")
			// toss any old pending requests when the height moves up
			// todo: Keep asks in a  list so cleanup is more efficient
			for ask, _ := range pending {
				if int(ask.DBH) < dbheight {
					s.LogPrintf(logname, "Expire %d/%d/%d %d", ask.DBH, ask.VM, ask.H, len(pending))
					delete(pending, ask)
				}
			}
		case ask := <-asks:
			s.LogPrintf("mmr", "start ask handling")
			addAsk(ask)  // add this ask
			addAllAsks() // add all pending asks

		case add := <-adds:
			s.LogPrintf("mmr", "start add handling")
			addAllAsks() // add all pending asks before any adds
			s.LogPrintf("mmr", "asks done")
			deletePendingAsk(add) // cancel any pending ask for the message just added to the process list

		case now = <-ticker:
			s.LogPrintf("mmr", "Ticker handling")
			addAllAsks()     // process all pending asks before any adds
			addAllAdds()     // process all pending add before any ticks
			readAllTickers() // drain the ticker channel

			//s.LogPrintf(logname, "tick [%v]", pending)

			// time offset to pick asks to

			//build MMRs with all the asks expired asks if we are not in ignore.
			if !s.IgnoreMissing {
				for ref, when := range pending {
					var index dbhvm = dbhvm{ref.DBH, ref.VM}
					// Drop any MMR request that are before the current height
					if ref.DBH < int(s.LLeaderHeight) {
						deletePendingAsk(ref)
						continue
					}
					// if ask is expired or we have an MMR for this DBH/VM and it's not a brand new ask
					if now > *when {

						if mmrs[index] == nil { // If we don't have a message for this DBH/VM
							mmrs[index] = messages.NewMissingMsg(s, ref.VM, uint32(ref.DBH), uint32(ref.H))
						} else {
							mmrs[index].ProcessListHeight = append(mmrs[index].ProcessListHeight, uint32(ref.H))
							// Add an ask for each msg we ask for, even if we bundle the asks.
							// This is so the accounting adds upp.
						}
						s.MissingRequestAskCnt++
						*when = now + askDelay // update when we asked
						// Maybe when asking for past the end of the list we should not ask again?
					}
				} //build a MMRs with all the expired asks in that VM at that DBH.

			}
			for index, mmr := range mmrs {
				s.LogMessage(logname, "sendout", mmr)
				if MMR_enable {
					mmr.SendOut(s, mmr)
				}
				delete(mmrs, index)
			} // Send MMRs that were built
		} // select across all the channels, block till something happens
		s.LogPrintf("mmr", "done")
	} // forever ...
} // func  makeMMRs() {...}

// MissingMessageResponseCache will cache all processlist items from the last 2 blocks.
// It can create MissingMessageResponses to peer requests, and prevent us from asking the network
// if we already have something locally.
type MissingMessageResponseCache struct {
	// MissingMsgRequests is the channel on which we receive acked messages to cache
	MissingMsgRequests chan interfaces.IMsg
	// ProcessedPairs is all the ack+msg pairs that we processed
	ProcessedPairs chan *MsgPair

	// AckMessageCache is the cached acks from the last 2 blocks
	AckMessageCache *AckMsgPairCache

	// We need the state for getting the current timestamp and for logging
	// TODO: Separate logging and current time from state
	localState *State

	quit chan bool
}

func NewMissingMessageReponseCache(s *State) *MissingMessageResponseCache {
	mmrc := new(MissingMessageResponseCache)
	mmrc.MissingMsgRequests = make(chan interfaces.IMsg, 20)
	mmrc.ProcessedPairs = make(chan *MsgPair, 5)
	mmrc.AckMessageCache = NewAckMsgCache()

	mmrc.quit = make(chan bool, 1)
	mmrc.localState = s

	return mmrc
}

// NotifyPeerMissingMsg is the threadsafe way to notify that a peer sent us a missing message
func (mmrc *MissingMessageResponseCache) NotifyPeerMissingMsg(missingMsg interfaces.IMsg) {
	mmrc.MissingMsgRequests <- missingMsg
}

// NotifyNewMsgPair is the threadsafe way to include a new msg pair to respond to missing message requests
// from peers
func (mmrc *MissingMessageResponseCache) NotifyNewMsgPair(ack interfaces.IMsg, msg interfaces.IMsg) {
	mmrc.ProcessedPairs <- &MsgPair{Ack: ack, Msg: msg}
}

func (mmrc *MissingMessageResponseCache) Close() {
	mmrc.quit <- true
}

// answerRequest will attempt to construct a missing message response for a missing message request.
// 		params:
//			request
func (mmrc *MissingMessageResponseCache) answerRequest(request *messages.MissingMsg) {
	// Loop through all requested heights
	for _, plHeight := range request.ProcessListHeight {
		// Check if we have the response for a given plHeight
		pair := mmrc.AckMessageCache.Get(int(request.DBHeight), request.VMIndex, int(plHeight))
		if pair != nil { // Woo! Respond to that peer
			if pair.Msg == nil || pair.Ack == nil {
				mmrc.localState.LogPrintf("mmr_response", "ackpair found, but the msgs were nil")
				continue // This should never happen
			}
			ack := pair.Ack.(*messages.Ack) // For logging, we want the dbheight, vm, and plheight of the ack

			// Pair exists, send out the response
			response := messages.NewMissingMsgResponse(mmrc.localState, pair.Msg, pair.Ack)
			response.SetOrigin(request.GetOrigin())
			response.SetNetworkOrigin(request.GetNetworkOrigin())
			response.SendOut(mmrc.localState, response)
			mmrc.localState.MissingRequestReplyCnt++
			mmrc.localState.LogMessage("mmr_response", fmt.Sprintf("request_fufilled %d/%d/%d", ack.DBHeight, ack.VMIndex, ack.Height), response)
		} else {
			// If we are missing the plheight, we increment the ignore count as we don't have what
			// the peer wanted.
			mmrc.localState.MissingRequestIgnoreCnt++
			mmrc.localState.LogPrintf("mmr_response", "pair_not_found %d/%d/%d", request.DBHeight, request.VMIndex, plHeight)
		}
	}
}

// Run will start the loop to read messages from the channel and build
// the cache to respond to missing message requests
func (mmrc *MissingMessageResponseCache) Run() {
	for {
		select {
		case processedPair := <-mmrc.ProcessedPairs:
			// A new ack/msg pair is processed and ready to respond to missing message requests
			ack := processedPair.Ack.(*messages.Ack)
			mmrc.localState.LogMessage("mmr_response", fmt.Sprintf("add_pair %d/%d/%d", ack.DBHeight, ack.VMIndex, ack.Height), processedPair.Ack)
			mmrc.AckMessageCache.AddMsgPair(processedPair)
		case requestI := <-mmrc.MissingMsgRequests:
			// A missing msg request from a peer
			var _ = requestI
			request, ok := requestI.(*messages.MissingMsg)
			if !ok {
				break // Should never not be a request
			}

			mmrc.answerRequest(request)
		case <-mmrc.quit:
			// Close thread
			return
		}
	}
}

// The pair of messages for a missing message response
type MsgPair struct {
	Ack interfaces.IMsg
	Msg interfaces.IMsg
}

type AckMsgPairCache struct {
	CurrentWorkingHeight int
	// MsgPairMap will contain ack/msg pairs
	MsgPairMap map[int]map[plRef]*MsgPair
}

func NewAckMsgCache() *AckMsgPairCache {
	a := new(AckMsgPairCache)
	a.MsgPairMap = make(map[int]map[plRef]*MsgPair)
	return a
}

// UpdateWorkingHeight will only update the height if it is new
func (a *AckMsgPairCache) UpdateWorkingHeight(newHeight int) {
	// Update working height if it has changed
	if a.CurrentWorkingHeight < int(newHeight) {
		a.CurrentWorkingHeight = int(newHeight)
		a.Expire(newHeight)
	}
}

// Expire for the AckMsgPairCache will expire all acks older than 2 blocks.
//	TODO: Is iterating over a map extra cost? Should we have a sorted list?
//			Technically we can just call delete NewHeight-2 as long as we always
//			Update every height
func (a *AckMsgPairCache) Expire(newHeight int) {
	a.CurrentWorkingHeight = newHeight
	for h, _ := range a.MsgPairMap {
		if a.HeightTooOld(h) {
			delete(a.MsgPairMap, h)
		}
	}
}

// AddMsgPair will add an ack to the cache if it is not too old, and it is an ack+msg pair
//	We assume that all msgs being added have been added to our processlist, and therefore
//	the current working height and they are valid.
func (a *AckMsgPairCache) AddMsgPair(pair *MsgPair) {
	ack, ok := pair.Ack.(*messages.Ack)
	if !ok {
		// Don't add non-acks
		return
	}

	// Verify ack and msg should be paired
	if !ack.MessageHash.IsSameAs(pair.Msg.GetMsgHash()) {
		return
	}

	// Attempt to update working height.
	a.UpdateWorkingHeight(int(ack.DBHeight))

	// Check if we still care about this height
	//	This should never fail, as it is triggered
	//	when we add to the processlist.
	if a.HeightTooOld(int(ack.DBHeight)) {
		// This should never happen
		return // Too old
	}

	plLoc := plRef{int(ack.DBHeight), ack.VMIndex, int(ack.Height)}
	a.ensure(plLoc.DBH)
	a.MsgPairMap[plLoc.DBH][plLoc] = pair
}

func (a *AckMsgPairCache) Get(dbHeight, vmIndex, plHeight int) *MsgPair {
	if a.MsgPairMap[dbHeight] == nil {
		return nil
	}
	return a.MsgPairMap[dbHeight][plRef{dbHeight, vmIndex, plHeight}]
}

func (a *AckMsgPairCache) ensure(height int) {
	if a.MsgPairMap[height] == nil {
		a.MsgPairMap[height] = make(map[plRef]*MsgPair)
	}
}

// HeightTooOld determines if the ack height is too old for the ackcache
func (a *AckMsgPairCache) HeightTooOld(height int) bool {
	// Eg: CurrentWorkingHeight = 10, so saved height is minimum 8. Below 8, we delete
	if height < a.CurrentWorkingHeight-2 {
		return true
	}
	return false
}
