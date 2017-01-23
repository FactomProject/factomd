package state

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"time"
)

var _ = fmt.Print

// Create a request for dbsates from begin to end
//
func Ask(state *State, begin uint32, end uint32) {
	if begin > 0 {
		begin--
	}
	msg := messages.NewDBStateMissing(state, uint32(begin), uint32(end))
	if msg != nil {
		msg.SendOut(state, msg)
		state.DBStateAskCnt++
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

	for len(h.History) < int(HighestSaved)-int(h.Base) {
		h.History = append(h.History, nil)
	}
	if h.Base < HighestSaved {
		cnt := int(HighestSaved - h.Base)
		if len(h.History) >= cnt {
			h.History = append(make([]*AskHistory,0), h.History[cnt:]...)
			h.Base = HighestSaved
		}
	}
}

func (h *AskHistories) Get(DBHeight uint32) *AskHistory {

	index := int(DBHeight) - int(h.Base)
	fmt.Println("index",index,"DBHEight", DBHeight,"base",h.Base)
	if index < 0 {
		return nil
	}
	for index >= len(h.History) {
		h.History = append(h.History, nil)
	}
	if h.History[index] == nil {
		h.History[index] = new(AskHistory)
		h.History[index].DBHeight = DBHeight
	}
	return h.History[index]
}

func FindBegin(state *State, histories *AskHistories, start uint32) int {
	begin := start
	now := state.GetTimestamp()
	for {
		fmt.Printf("Highest Known %5d Highest Saved %5d ",state.HighestKnown,state.HighestSaved)
		fmt.Println("begin",begin)
		h := histories.Get(begin)
		ir := int(begin) - state.DBStatesReceivedBase
		switch {
		case begin > state.HighestKnown:
			return -1
		case h == nil:
		case len(state.DBStatesReceived) > int(ir) && state.DBStatesReceived[ir] != nil:
		case h.LastRequest == nil:
			h.LastRequest = now
			return int(begin)
		case now.GetTimeSeconds()-h.LastRequest.GetTimeSeconds() >= 3:
			h.LastRequest = now
			return int(begin)
		}
		begin++
	}
}

func FindEnd(state *State, histories *AskHistories, begin uint32) uint32 {
	now := state.GetTimestamp()
	end := begin+1
	for {
		fmt.Println("end")
		h := histories.Get(end)
		ir := int(end) - state.DBStatesReceivedBase

		switch {
		case len(state.DBStatesReceived) > int(ir) && state.DBStatesReceived[ir] != nil:
			return end-1
		case end >= state.HighestKnown:
			return end
		case end-begin >= 200:
			return end
		case h.LastRequest == nil:
		case now.GetTimeSeconds()-h.LastRequest.GetTimeSeconds() < 3:
			return end-1
		}
		h.LastRequest = now
		end++
	}
	return end
}

// Does one check, and possible submission of a request.
func Step(state *State, histories *AskHistories) {


	// While ignoring, don't ask for dbstates
	if state.IgnoreMissing {
		return
	}

	// We only check if we need updates once every so often.
	if len(state.inMsgQueue) > 1000 {
		return
	}

	hs := state.HighestSaved
	hk := state.HighestKnown

	if state.CurrentMinute > 0 && hk-hs <= 1 {
		return
	} else if state.CurrentMinute == 0 && hk-hs <= 2 {
		return
	}

	//histories.Trim(hs, hk)

	begin := FindBegin(state, histories, hs+1)
	if begin <= 0 || uint32(begin) > hs+1000 {
		return
	}

	end := FindEnd(state, histories, uint32(begin))
	Ask(state, uint32(begin), end)

}

// Once a second at most, we check to see if we need to pull down some blocks to catch up.
func (list *DBStateList) Catchup() {

	histories := new(AskHistories)

	// Just let the system come up before we try to sync.
	time.Sleep(5 * time.Second)

	state := list.State
	state.DBStateMutex.Lock()

	for {
		// So we yeild so we don't lock up the system.
		state.DBStateMutex.Unlock()
		time.Sleep(500 * time.Millisecond)
		state.DBStateMutex.Lock()

		Step(list.State, histories)
	}
}
