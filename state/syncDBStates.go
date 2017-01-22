package state

import (
	"fmt"
	"github.com/FactomProject/factomd/common/messages"
	"time"
)

var _ = fmt.Print

// Once a second at most, we check to see if we need to pull down some blocks to catch up.
func (list *DBStateList) Catchup() {

	timeToAsk := list.State.GetTimestamp()

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

		now := list.State.GetTimestamp()

		hs := int(list.State.HighestSaved)
		hk := int(list.State.HighestKnown)

		begin := hs
		end := hk

		ask := func() {
			if timeToAsk != nil && hk-hs > 4 && now.GetTime().After(timeToAsk.GetTime()) {

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
					timeToAsk.SetTimeSeconds(now.GetTimeSeconds() + 3)
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
			timeToAsk = nil
			continue theMainLoop
		}

		// First Ask.  Because the timer is nil!
		if timeToAsk == nil {
			// Okay, have nothing in play, so wait a bit just in case.
			timeToAsk = list.State.GetTimestamp()
			timeToAsk.SetTimeSeconds(now.GetTimeSeconds() + 5)
			list.LastBegin = begin
			list.LastEnd = end
			continue theMainLoop
		}

		if timeToAsk.GetTime().Before(now.GetTime()) {
			ask()
			continue theMainLoop
		}
	}
}
