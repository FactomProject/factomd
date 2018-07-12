package state

import (
	"github.com/FactomProject/factomd/common/messages"
)

// Once a second at most, we check to see if we need to pull down some blocks to catch up.
func (list *DBStateList) Catchup(justDoIt bool) {
	// We only check if we need updates once every so often.

	now := list.State.GetTimestamp()

	hs := int(list.State.GetHighestSavedBlk())
	hk := int(list.State.GetHighestAck())
	if list.State.GetHighestKnownBlock() > uint32(hk+2) {
		hk = int(list.State.GetHighestKnownBlock())
	}

	begin := hs + 1
	end := hk

	ask := func() {

		tolerance := 1
		if list.State.Leader {
			tolerance = 2
		}

		if list.TimeToAsk != nil && hk-hs > tolerance && now.GetTime().After(list.TimeToAsk.GetTime()) {

			// Find the first dbstate we don't have.
			for i, v := range list.State.DBStatesReceived {
				ix := i + list.State.DBStatesReceivedBase
				if ix <= hs {
					continue
				}
				if ix >= hk {
					return
				}
				//// if we already have more than 100 waiting to process don't ask for more
				//if ix > hs+100 {
				//	return
				//}
				if v == nil {
					begin = ix
					break
				}
			}

			for len(list.State.DBStatesReceived)+list.State.DBStatesReceivedBase <= hk {
				list.State.DBStatesReceived = append(list.State.DBStatesReceived, nil)
			}

			//  Find the end of the dbstates that we don't have.
			for i, v := range list.State.DBStatesReceived {
				ix := i + list.State.DBStatesReceivedBase

				if ix <= begin {
					continue
				}
				if ix >= end {
					break
				}
				if v != nil {
					end = ix - 1
					break
				}
			}

			if list.State.RunLeader && !list.State.IgnoreMissing {
				msg := messages.NewDBStateMissing(list.State, uint32(begin), uint32(end))

				if msg != nil {
					//		list.State.RunLeader = false
					//		list.State.StartDelay = list.State.GetTimestamp().GetTimeMilli()
					msg.SendOut(list.State, msg)
					list.State.DBStateAskCnt++
					list.TimeToAsk.SetTimeSeconds(now.GetTimeSeconds() + 6)
					list.LastBegin = begin
					list.LastEnd = end
				}
			}
		}
	}

	if end-begin > 200 {
		end = begin + 200
	}

	if end+3 > begin && justDoIt {
		ask()
		return
	}

	// return if we are caught up, and clear our timer
	if end-begin < 1 {
		list.TimeToAsk = nil
		return
	}

	// First Ask.  Because the timer is nil!
	if list.TimeToAsk == nil {
		// Okay, have nothing in play, so wait a bit just in case.
		list.TimeToAsk = list.State.GetTimestamp()
		list.TimeToAsk.SetTimeSeconds(now.GetTimeSeconds() + 6)
		list.LastBegin = begin
		list.LastEnd = end
		return
	}

	if list.TimeToAsk.GetTime().Before(now.GetTime()) {
		ask()
		return
	}

}
