package state

import (
	"container/list"
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"time"
)

var _ = fmt.Print

// Create a request for dbsates from begin to end
//
func Ask(state State, begin uint32, end uint32) {
	msg := messages.NewDBStateMissing(state, uint32(begin), uint32(end))
	if msg != nil {
		msg.SendOut(state, msg)
		State.DBStateAskCnt++
	}
}

// Keep track of request history per dbstate
type AskHistory struct {
	LastRequest interfaces.Timestamp
	DBHeight    uint32
}

// All the dbstates we have asked for to date.
type AskHistories struct {
	Base    uint32
	History []*AskHistory
}

// Trim old histories.  We don't need them any more.
func (h *AskHistories) Trim(HighestSaved uint32, HighestKnown uint32) {

	for len(h.History) < int(HighestSaved) {
		h.History = append(h.History, nil)
	}
	if h.Base < HighestSaved {
		cnt := int(HighestSaved - h.Base)
		if len(h.History) > cnt {
			h.History = append(h.History[:0], h.History[cnt:]...)
			copy(h.History, h.History[cnt:])
		}
		h.Base = HighestSaved
	}
}

func (h *AskHistories) Get(DBHeight uint32) *AskHistory {

	index := int(DBHeight) - int(h.Base)
	if index < 0 {
		return nil
	}
	for index > len(h.History) {
		h.History = append(h.History, nil)
	}
	if h.History[index] == nil {
		h.History[index] = new(AskHistory)
		h.History[index].DBHeight = DBHeight
	}
	return h.History[index]
}

func FindBegin(state State, histories AskHistories, start uint32) uint32 {
	begin := start
	now := state.GetTimestamp()
	for {
		switch {
		case begin > state.HighestKnown:
			return -1
		case histories.Get(begin) != nil && now.GetTimeSeconds()-histories.Get(begin).LastRequest > 3:
			begin++
		case histories.Get(begin) == nil:
			begin++
		case state.DBStatesReceived[begin] == nil:
			begin++
		default:
			return begin
		}
	}
}

func FindEnd(state State, histories AskHistories, start uint32) uint32 {
	now := state.GetTimestamp()
	end := start
	for state.DBStatesReceived[end] == nil {
		h := histories.Get(end)
		switch {
		case star
		case h == nil:
			histories.History[end]=new(AskHistory)
			histories.History[end].DBHeight = end
			histories.History[end].LastRequest = now
		}
		now := state.GetTimestamp()
		if h.LastRequest == nil {
			h.LastRequest = now
			continue
		}
		if now.GetTimeSeconds()-h.LastRequest.GetTimeSeconds() < 3 {
			continue
		}
		h.LastRequest = now
		end++
	}
	return end
}

// Once a second at most, we check to see if we need to pull down some blocks to catch up.
func (list *DBStateList) Catchup() {

	histories := new(AskHistories)

	// Just let the system come up before we try to sync.
	time.Sleep(5 * time.Second)

	list.State.DBStateMutex.Lock()
theMainLoop:
	for {

		// So we yeild so we don't lock up the system.
		list.State.DBStateMutex.Unlock()
		time.Sleep(100 * time.Millisecond)
		list.State.DBStateMutex.Lock()

		// While ignoring, don't ask for dbstates
		if list.State.IgnoreMissing {
			continue theMainLoop
		}

		// We only check if we need updates once every so often.
		if len(list.State.inMsgQueue) > 1000 {
			continue theMainLoop
		}

		hs := int(list.State.HighestSaved)
		hk := int(list.State.HighestKnown)

		if list.State.CurrentMinute > 0 && list.State.LLeaderHeight-hs <= 1 {
			continue theMainLoop
		} else if list.State.CurrentMinute == 0 && list.State.LLeaderHeight-hs <= 2 {
			continue theMainLoop
		}

		histories.Trim(hs, hk)

		begin := FindBegin(list.State,histories,hs+1)
		if begin < 0 || begin > hs+1000 {
			continue
		}

		end := FindEnd(list.State, histories, begin)
			Ask(list.State, hs+1, end)
		}

	}
}
