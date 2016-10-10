// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"time"

	"github.com/FactomProject/factomd/state"
)

func Negotiate(s *state.State) {
	time.Sleep(3 * time.Second)
	for {
		pl := s.ProcessLists.LastList()
		if pl.LenFaultMap() > 0 {
			faultIDs := pl.GetKeysFaultMap()
			for _, faultID := range faultIDs {
				faultState := pl.GetFaultState(faultID)
				if faultState.AmINegotiator {
					state.CraftAndSubmitFullFault(pl, faultID)
					if faultState.HasEnoughSigs(s) && faultState.PledgeDone {
						break
					}
				}
			}
		}
		time.Sleep(5 * time.Second)
	}
}
