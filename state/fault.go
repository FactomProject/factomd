// Copyright 2016 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"time"

	"github.com/FactomProject/factomd/common/messages"
)

func fault(p *ProcessList, vmIndex int, waitSeconds int64, vm *VM, thetime int64, height int, tag int) int64 {
	now := time.Now().Unix()

	if thetime == 0 {
		thetime = now
	}

	if now-thetime >= waitSeconds {
		atLeastOneServerOnline := false
		for _, fed := range p.FedServers {
			if fed.IsOnline() {
				atLeastOneServerOnline = true
				break
			}
		}
		if !atLeastOneServerOnline {
			return now
		}
		atLeastOneAuditOnline := false
		for _, aud := range p.AuditServers {
			if aud.IsOnline() {
				atLeastOneAuditOnline = true
				break
			}
		}
		if !atLeastOneAuditOnline {
			for _, aud := range p.AuditServers {
				aud.SetOnline(true)
			}
		}

		leaderMin := getLeaderMin(p)

		myIndex := p.ServerMap[leaderMin][vmIndex]

		p.FedServers[myIndex].SetOnline(false)
		id := p.FedServers[myIndex].GetChainID()

		if vm.faultHeight < 0 {
			vm.whenFaulted = now
			//p.FaultTimes[id.String()] = p.State.GetTimestamp().GetTimeSeconds()
		}

		vm.faultHeight = height

		responsibleFaulterIdx := vmIndex + 1
		if responsibleFaulterIdx >= len(p.FedServers) {
			responsibleFaulterIdx = 0
		}

		if p.State.Leader {
			if p.State.LeaderVMIndex == responsibleFaulterIdx {
				p.NegotiatorVMIndex = vmIndex
				p.AmINegotiator = true
				negotiationMsg := messages.NewNegotiation(p.State.GetTimestamp(), id, vmIndex, p.DBHeight, uint32(height))
				if negotiationMsg != nil {
					negotiationMsg.Sign(p.State.serverPrivKey)
					negotiationMsg.SendOut(p.State, negotiationMsg)
					negotiationMsg.FollowerExecute(p.State)
				}
				thetime = now
			}
		}

		nextVM := p.VMs[responsibleFaulterIdx]

		// tags of 0 and 1 represent faults for EOM duties
		// a tag of 2 instead means that we got a bad ack
		// which means we don't need to fault the negotiator
		// (who may or may not have gotten that bad ack)
		if now-vm.whenFaulted > 20 && tag < 2 {
			_, negotiationInitiated := p.NegotiationInit[id.String()]
			if !negotiationInitiated || now-vm.whenFaulted > 60 {
				if nextVM.faultHeight < 0 {
					for pledger, pledgeSlot := range p.PledgeMap {
						if pledgeSlot == id.String() {
							delete(p.PledgeMap, pledger)
						}
					}
				}
				nextVM.faultingEOM = fault(p, responsibleFaulterIdx, 20, nextVM, nextVM.faultingEOM, height, 1)
			}
		}

		thetime = now
	}

	return thetime
}
