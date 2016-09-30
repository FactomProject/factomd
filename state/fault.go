// Copyright 2016 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"time"

	"github.com/FactomProject/factomd/common/interfaces"

	"github.com/FactomProject/factomd/common/messages"
)

type FaultCore struct {
	// The following 5 fields represent the "Core" of the message
	// This should match the Core of FullServerFault messages
	ServerID      interfaces.IHash
	AuditServerID interfaces.IHash
	VMIndex       byte
	DBHeight      uint32
	Height        uint32
}

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
							if pledger == p.State.IdentityChainID.String() {
								p.AmIPledged = false

								for faultKey, sf := range p.State.FaultInfoMap {
									if sf.AuditServerID.String() == pledger && len(p.State.FaultVoteMap[faultKey]) > len(p.FedServers)/2 {
										p.AmIPledged = true
										p.PledgeMap[p.State.IdentityChainID.String()] = sf.ServerID.String()

										nsf := messages.NewServerFault(p.State.GetTimestamp(), sf.ServerID, p.State.IdentityChainID, int(sf.VMIndex), sf.DBHeight, sf.Height)
										if nsf != nil {
											nsf.Sign(p.State.serverPrivKey)
											p.State.NetworkOutMsgQueue() <- nsf
											p.State.InMsgQueue() <- nsf
										}
									}
								}
							}
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

func (s *State) FollowerExecuteSFault(m interfaces.IMsg) {
	sf, _ := m.(*messages.ServerFault)
	pl := s.ProcessLists.Get(sf.DBHeight)

	if pl == nil {
		return
	}

	if pl.VMs[sf.VMIndex].faultHeight < 0 {
		return
	}

	var issuerID [32]byte
	rawIssuerID := sf.GetSignature().GetKey()
	for i := 0; i < 32; i++ {
		if i < len(rawIssuerID) {
			issuerID[i] = rawIssuerID[i]
		}
	}

	// if this fault is an Audit server voting for itself to be promoted
	// (i.e. it is an "AOK" message)
	// then we need to mark the Audit server as "ReadyForPromotion" or
	// alternatively mark it "offline" if it has voted promiscuously
	// during this negotiation
	for _, a := range s.Authorities {
		if a.AuthorityChainID.IsSameAs(sf.AuditServerID) {
			marshalledSF, err := sf.MarshalForSignature()
			if err == nil {
				sigVer, err := a.VerifySignature(marshalledSF, sf.Signature.GetSignature())
				if err == nil && sigVer {
					if foundAudit, audIdx := pl.GetAuditServerIndexHash(sf.AuditServerID); foundAudit {
						if pledgeSlot, pledged := pl.PledgeMap[sf.AuditServerID.String()]; pledged {
							//if pl.AuditServers[audIdx].LeaderToReplace() != nil {
							if pledgeSlot != sf.ServerID.String() {
								// illegal vote; audit server has already AOK'd replacing a different leader
								// "punish" them by setting them offline (i.e. make them ineligible for promotion)
								pl.AuditServers[audIdx].SetOnline(false)
							}
						} else {
							// AOK: set the Audit Server's "Leader to Replace" field to this ServerID
							//pl.AuditServers[audIdx].SetReplace(sf.ServerID)
							pl.PledgeMap[sf.AuditServerID.String()] = sf.ServerID.String()
						}
					}
				}
			}
		}
	}

	coreHash := sf.GetCoreHash().Fixed()

	if s.FaultVoteMap[coreHash] == nil {
		s.FaultVoteMap[coreHash] = make(map[[32]byte]interfaces.IFullSignature)
	}

	s.FaultInfoMap[coreHash] = FaultCore{ServerID: sf.ServerID, AuditServerID: sf.AuditServerID, VMIndex: sf.VMIndex, DBHeight: sf.DBHeight, Height: sf.Height}

	lbytes, err := sf.MarshalForSignature()

	sfSig := sf.Signature.GetSignature()
	sfSigned, err := s.VerifyAuthoritySignature(lbytes, sfSig, sf.DBHeight)
	if err == nil && sfSigned == 1 {
		s.FaultVoteMap[coreHash][issuerID] = sf.GetSignature()
	}

	if s.Leader {
		cnt := len(s.FaultVoteMap[coreHash])
		var fedServerCnt int
		if pl != nil {
			fedServerCnt = len(pl.FedServers)
		} else {
			fedServerCnt = len(s.GetFedServers(sf.DBHeight))
		}
		responsibleFaulterIdx := (int(sf.VMIndex) + 1) % fedServerCnt

		if cnt > (fedServerCnt / 2) {
			if s.LeaderVMIndex == responsibleFaulterIdx {
				if foundAudit, _ := pl.GetAuditServerIndexHash(sf.AuditServerID); foundAudit {
					serverToReplace, pledged := pl.PledgeMap[sf.AuditServerID.String()]
					if pledged {
						if serverToReplace == sf.ServerID.String() {
							var listOfSigs []interfaces.IFullSignature
							for _, sig := range s.FaultVoteMap[coreHash] {
								listOfSigs = append(listOfSigs, sig)
							}
							fullFault := messages.NewFullServerFault(sf, listOfSigs)
							absf := fullFault.ToAdminBlockEntry()
							s.LeaderPL.AdminBlock.AddServerFault(absf)
							if fullFault != nil {
								fullFault.Sign(s.serverPrivKey)
								s.NetworkOutMsgQueue() <- fullFault
								fullFault.FollowerExecute(s)
								pl.AmINegotiator = false
								delete(s.FaultVoteMap, sf.GetCoreHash().Fixed())
								delete(s.FaultInfoMap, sf.GetCoreHash().Fixed())
							}
						}
					}
				}
			}
		}

		//Match a nomination if we haven't nominated the same server already
		existingNominations, exists := pl.AlreadyNominated[sf.ServerID.String()]
		if exists {
			_, alreadyNom := existingNominations[sf.AuditServerID.String()]
			if !alreadyNom {
				pl.AlreadyNominated[sf.ServerID.String()][sf.AuditServerID.String()] = s.GetTimestamp().GetTimeSeconds()
				matchNomination := messages.NewServerFault(s.GetTimestamp(), sf.ServerID, sf.AuditServerID, int(sf.VMIndex), sf.DBHeight, sf.Height)
				if matchNomination != nil {
					matchNomination.Sign(s.serverPrivKey)
					s.NetworkOutMsgQueue() <- matchNomination
					s.InMsgQueue() <- matchNomination
				}
			}
		}
	} else {
		if s.IdentityChainID.IsSameAs(sf.AuditServerID) {
			// I am the audit server being promoted;
			// I will pledge myself to replace the faulted leader
			// (unless I am already pledged elsewhere)
			if !pl.AmIPledged {
				pl.AmIPledged = true
				pl.PledgeMap[s.IdentityChainID.String()] = sf.ServerID.String()

				nsf := messages.NewServerFault(s.GetTimestamp(), sf.ServerID, s.IdentityChainID, int(sf.VMIndex), sf.DBHeight, sf.Height)
				if nsf != nil {
					nsf.Sign(s.serverPrivKey)
					s.NetworkOutMsgQueue() <- nsf
					s.InMsgQueue() <- nsf
				}
			}
		}
	}
}

func (s *State) FollowerExecuteFullFault(m interfaces.IMsg) {
	fullFault, _ := m.(*messages.FullServerFault)
	relevantPL := s.ProcessLists.Get(fullFault.DBHeight)
	auditServerList := s.GetAuditServers(fullFault.DBHeight)
	var theAuditReplacement interfaces.IFctServer

	for _, auditServer := range auditServerList {
		if auditServer.GetChainID().IsSameAs(fullFault.AuditServerID) {
			theAuditReplacement = auditServer
		}
	}
	if theAuditReplacement != nil {
		for listIdx, fedServ := range relevantPL.FedServers {
			if fedServ.GetChainID().IsSameAs(fullFault.ServerID) {
				relevantPL.FedServers[listIdx] = theAuditReplacement
				relevantPL.FedServers[listIdx].SetOnline(true)
				relevantPL.AddAuditServer(fedServ.GetChainID())
				s.RemoveAuditServer(fullFault.DBHeight, theAuditReplacement.GetChainID())
				if foundVM, vmindex := relevantPL.GetVirtualServers(s.CurrentMinute, theAuditReplacement.GetChainID()); foundVM {
					relevantPL.VMs[vmindex].faultHeight = -1
					relevantPL.VMs[vmindex].faultingEOM = 0
				}
				break
			}
		}
	}

	s.Leader, s.LeaderVMIndex = s.LeaderPL.GetVirtualServers(s.CurrentMinute, s.IdentityChainID)
	delete(s.FaultVoteMap, fullFault.GetCoreHash().Fixed())
	delete(s.FaultInfoMap, fullFault.GetCoreHash().Fixed())

	for pledger, pledgeSlot := range relevantPL.PledgeMap {
		if pledgeSlot == fullFault.ServerID.String() {
			delete(relevantPL.PledgeMap, pledger)
			if pledger == s.IdentityChainID.String() {
				relevantPL.AmIPledged = false
			}
		}
	}
}
