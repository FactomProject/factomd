package state

import (
	"sort"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/messages"
)

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

type MMRInfo struct {
	// Channels for managing the missing message requests
	asks      chan askRef // Requests to ask for missing messages
	adds      chan plRef  // notices of slots filled in the process list
	dbheights chan int    // Notice that this DBHeight is done
}

// starts the MMR processing for this state
func (s *State) StartMMR() {
	// Missing message request handling.
	go s.makeMMRs(s.asks, s.adds, s.dbheights)
}

// Ask VM for an MMR for this height with delay ms before asking the network
// called from validation thread to notify MMR that we are missing a message
func (vm *VM) ReportMissing(height int, delay int64) {
	if height < vm.HighestAsk { // Don't report the same height twice
		return
	}
	now := vm.p.State.GetTimestamp().GetTimeMilli()
	if delay < 500 {
		delay = 500 // Floor for delays is 500ms so there is time to merge adjacent requests
	}
	lenVMList := len(vm.List)
	// ask for all missing messages
	var i int
	for i = vm.HighestAsk; i < lenVMList; i++ {
		if vm.List[i] == nil {
			vm.p.State.Ask(int(vm.p.DBHeight), vm.VmIndex, i, now+delay) // send it to the MMR thread
			vm.HighestAsk = i                                            // We have asked for all nils up to this height
		}
	}

	// if we are asking above the current list
	if height >= lenVMList {
		vm.p.State.Ask(int(vm.p.DBHeight), vm.VmIndex, height, now+delay) // send it to the MMR thread
		vm.HighestAsk = height                                            // We have asked for all nils up to this height
	}

}

// Ask is called from ReportMissing which comes from validation thread to notify MMR that we are missing a message
func (s *State) Ask(DBHeight int, vmIndex int, height int, when int64) {
	if s.asks == nil { // If it is nil, there is no makemmrs
		return
	}
	// do not ask for things in the past or very far into the future
	if DBHeight < int(s.LLeaderHeight) || DBHeight > int(s.LLeaderHeight)+1 || DBHeight < int(s.DBHeightAtBoot) {
		return
	}

	ask := askRef{plRef{DBHeight, vmIndex, height}, when}
	s.asks <- ask
	return
}

// Used by debug code only
var MMR_enable bool = true

// Receive all asks and all process list adds and create missing message requests any ask that has expired
// and still pending. Add 10 seconds to the ask.
func (s *State) makeMMRs(asks <-chan askRef, adds <-chan plRef, dbheights <-chan int) {
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
		} // don't update the when if it already existed...

		// checking if we already have the "missing" message in our maps
		ack, msg := s.RecentMessage.GetAckAndMsg(ask.DBH, ask.VM, ask.H, s)

		if msg != nil && ack != nil {
			// send them to be executed
			s.msgQueue <- msg
			s.ackQueue <- ack
		}
	}

	// process all pending asks
	addAllAsks := func() {
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
			if len(ticker) == cap(ticker) {
				panic("MMR Processing Stalled")
			} // time to die, no one is listening

			ticker <- s.GetTimestamp().GetTimeMilli()
			askDelay := int64(s.DirectoryBlockInSeconds*1000) / 600
			if askDelay < 500 { // Don't go below 500ms. That is just too much
				askDelay = 500
			}

			time.Sleep(time.Duration(askDelay) * time.Millisecond)
		}
	}()

	lastAskDelay := int64(0)
	for {
		// You have to compute this at every cycle as you can change the block time in sim control.

		askDelay := int64(s.DirectoryBlockInSeconds*1000) / 50
		// Take 1/5 of 1 minute boundary (DBlock is 10*min)
		//		This means on 10min block, 12 second delay
		//					  1min block, 1.2 second delay

		if askDelay < 500 { // Don't go below 500ms. That is just too much
			askDelay = 500
		}

		if askDelay != lastAskDelay {
			s.LogPrintf(logname, "AskDelay %d BlockTime %d", askDelay, s.DirectoryBlockInSeconds)
			lastAskDelay = askDelay
		}

		// process any incoming messages
		select {
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

			//build MMRs with all the asks expired asks.
			for ref, when := range pending {
				var index dbhvm = dbhvm{ref.DBH, ref.VM}
				// if ask is expired or we have an MMR for this DBH/VM and it's not a brand new ask
				if now > *when {

					if mmrs[index] == nil { // If we don't have a message for this DBH/VM
						mmrs[index] = messages.NewMissingMsg(s, ref.VM, uint32(ref.DBH), uint32(ref.H))
					} else {
						mmrs[index].ProcessListHeight = append(mmrs[index].ProcessListHeight, uint32(ref.H))
					}
					*when = now + askDelay // update when we asked
					// Maybe when asking for past the end of the list we should not ask again?
				}
			} //build a MMRs with all the expired asks in that VM at that DBH.

			for index, mmr := range mmrs {
				s.LogMessage(logname, "sendout", mmr)
				s.MissingRequestAskCnt++
				if MMR_enable {
					mmr.SendOut(s, mmr)
				}
				delete(mmrs, index)
			} // Send MMRs that were built
		} // select across all the channels, block till something happens
		s.LogPrintf("mmr", "done")
	} // forever ...
} // func  makeMMRs() {...}

// MissingMessageResponseCache will cache all proceslist items from the last 2 blocks.
// It can create MissingMessageResponses to peer requests, and prevent us from asking the network
// if we already have something locally.
type MissingMessageResponseCache struct {
	// NewMsgs is the channel on which we receive acked messages to cache
	NewMsgs chan interfaces.IMsg

	// ACKCache is the cached acks from the last 2 blocks
	AckMessageCache     AckCache
	GeneralMessageCache MsgCache
}

type AckCache struct {
	CurrentHeight int
	// AckMap will only contain ack messages
	AckMap map[int]map[plRef]interfaces.IMsg
}

func NewAckCache() *AckCache {
	a := new(AckCache)
	a.AckMap = make(map[int]map[plRef]interfaces.IMsg)
	return a
}

// Expire for the AckCache will expire all acks older than 2 blocks.
//	TODO: Is iterating over a map extra cost? Should we have a sorted list?
//			Technically we can just call delete NewHeight-2 as long as we always
//			Update every height
func (a *AckCache) Expire(newHeight int) {
	a.CurrentHeight = newHeight
	for h, _ := range a.AckMap {
		if a.HeightTooOld(h) {
			delete(a.AckMap, h)
		}
	}
}

// AddAck will add an ack to the cache if it is not too old, and it is an ack
func (a *AckCache) AddAck(m interfaces.IMsg) {
	ack, ok := m.(*messages.Ack)
	if !ok {
		// Don't add non-acks
		return
	}
	if a.HeightTooOld(int(ack.DBHeight)) || a.HeightTooFuture(int(ack.DBHeight)) {
		return // Too old or too new to care about
	}
	plLoc := plRef{int(ack.DBHeight), ack.VMIndex, int(ack.Height)}
	a.ensure(plLoc.DBH)
	a.AckMap[plLoc.DBH][plLoc] = ack
}

func (a *AckCache) Get(dbHeight, vmIndex, plHeight int) interfaces.IMsg {
	if a.AckMap[dbHeight] == nil {
		return nil
	}
	return a.AckMap[dbHeight][plRef{dbHeight, vmIndex, plHeight}]
}

func (a *AckCache) ensure(height int) {
	if a.AckMap[height] == nil {
		a.AckMap[height] = make(map[plRef]interfaces.IMsg)
	}
}

func (a *AckCache) HeightTooFuture(height int) bool {
	// If the ack is from too far in the future, we can also ignore it
	// TODO: Determine this
	return false
}

// HeightTooOld determines if the ack height is too old for the ackcache
func (a *AckCache) HeightTooOld(height int) bool {
	// Eg: CurrentHeight = 10, so saved height is minimum 8. Below 8, we delete
	if height < a.CurrentHeight-2 {
		return true
	}
	return false
}

type MsgCache struct {
	// MessageMap allows quick lookup for a message hash
	MessageMap map[[32]byte]interfaces.IMsg
	// MessageSlice is the sorted slice of messages by time. This is useful for
	// expiring messages from the map without having to iterate over the entire list.
	MessageSlice []interfaces.IMsg
}

func NewMsgCache() *MsgCache {
	c := new(MsgCache)
	c.MessageMap = make(map[[32]byte]interfaces.IMsg)
	return c
}

func (c *MsgCache) AddMsg(m interfaces.IMsg) {
	// Only add messages that need an
	if !constants.NeedsAck(m.Type()) {
		return
	}

	c.MessageMap[m.GetMsgHash().Fixed()] = m
	c.InsertMsg(m)
}

func (c *MsgCache) InsertMsg(m interfaces.IMsg) {
	index := sort.Search(len(c.MessageSlice), func(i int) bool {
		return c.MessageSlice[i].GetTimestamp().GetTimeMilli() < m.GetTimestamp().GetTimeMilli()
	})
	c.MessageSlice = append(c.MessageSlice, (interfaces.IMsg)(nil))
	copy(c.MessageSlice[index+1:], c.MessageSlice[index:])
	c.MessageSlice[index] = m
}
