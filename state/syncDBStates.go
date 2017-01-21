package state

import (
	"github.com/FactomProject/factomd/common/messages"
	"time"
)

// Once a second at most, we check to see if we need to pull down some blocks to catch up.
func (list *DBStateList) Catchup() {

	list.State.DBStateMutex.Lock()

theMainLoop:
	for {

		list.State.DBStateMutex.Unlock()

		// Let some others work on stuff.
		time.Sleep(50 * time.Millisecond)

		list.State.DBStateMutex.Lock()

		// We only check if we need updates once every so often.

		if len(list.State.inMsgQueue) > 1000 {
			// If we are behind the curve in processing messages, dump all the dbstates from holding.
			for k := range list.State.Holding {
				if _, ok := list.State.Holding[k].(*messages.DBStateMsg); ok {
					delete(list.State.Holding, k)
				}
			}

			continue theMainLoop
		}

		now := list.State.GetTimestamp()

		hs := int(list.State.GetHighestSavedBlk())
		hk := int(list.State.GetHighestKnownBlock())
		begin := hs
		end := hk

		ask := func() {
			if list.TimeToAsk != nil && hk-hs > 4 && now.GetTime().After(list.TimeToAsk.GetTime()) {

				// Don't ask for more than we already have.
				for i, v := range list.State.DBStatesReceived {
					if i <= hs-list.State.DBStatesReceivedBase {
						continue
					}
					ix := i + list.State.DBStatesReceivedBase
					if v != nil && ix < end {
						end = ix + 1
						if begin > end {
							return
						}
						break
					}
				}

				msg := messages.NewDBStateMissing(list.State, uint32(begin), uint32(end+3))

				if msg != nil {
					//		list.State.RunLeader = false
					//		list.State.StartDelay = list.State.GetTimestamp().GetTimeMilli()
					msg.SendOut(list.State, msg)
					list.State.DBStateAskCnt++
					list.TimeToAsk.SetTimeSeconds(now.GetTimeSeconds() + 3)
					list.LastBegin = begin
					list.LastEnd = end + 3
				}
			}
		}

		if end-begin > 200 {
			end = begin + 200
		}

		if end+3 > begin && list.State.JustDoIt {
			ask()
			list.State.JustDoIt = false
			continue theMainLoop
		}

		// return if we are caught up, and clear our timer
		if end-begin <= 1 {
			list.TimeToAsk = nil
			continue theMainLoop
		}

		// First Ask.  Because the timer is nil!
		if list.TimeToAsk == nil {
			// Okay, have nothing in play, so wait a bit just in case.
			list.TimeToAsk = list.State.GetTimestamp()
			list.TimeToAsk.SetTimeSeconds(now.GetTimeSeconds() + 5)
			list.LastBegin = begin
			list.LastEnd = end
			continue theMainLoop
		}

		if list.TimeToAsk.GetTime().Before(now.GetTime()) {
			ask()
			continue theMainLoop
		}
	}
}
