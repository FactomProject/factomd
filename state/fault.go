// Copyright 2016 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"fmt"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"

	"github.com/FactomProject/factomd/common/messages"
)

type FaultState struct {
	FaultCore          FaultCore
	AmINegotiator      bool
	MyVoteTallied      bool
	VoteMap            map[[32]byte]interfaces.IFullSignature
	NegotiationOngoing bool
	PledgeDone         bool
	LastMatch          int64
}

func (fs FaultState) String() string {
	return fmt.Sprintf("Fed: %s Audit: %s, VM: %d, Height: %d, AmINego: %v, MyVote: %v, Votes: %d, Pledged: %v",
		fs.FaultCore.ServerID.String()[:10], fs.FaultCore.AuditServerID.String()[:10], int(fs.FaultCore.VMIndex), fs.FaultCore.Height,
		fs.AmINegotiator, fs.MyVoteTallied, len(fs.VoteMap), fs.PledgeDone)
}

type FaultCore struct {
	// The following 5 fields represent the "Core" of the message
	// This should match the Core of FullServerFault messages
	ServerID      interfaces.IHash
	AuditServerID interfaces.IHash
	VMIndex       byte
	DBHeight      uint32
	Height        uint32
}

func fault(pl *ProcessList, vm *VM, vmIndex, height, tag int) {
	now := time.Now().Unix()

	if vm.whenFaulted == 0 {
		// if we did not previously consider this VM faulted
		// we simply mark it as faulted (by assigning it a nonzero whenFaulted time)
		vm.whenFaulted = now
		vm.faultHeight = height
	} else {
		if now-vm.whenFaulted > 20 {
			if !vm.faultInitiatedAlready {
				// after 20 seconds, we take initiative and
				// issue a server fault vote of our own
				craftAndSubmitFault(pl, vm, vmIndex, height)
				vm.faultInitiatedAlready = true
				//if I am negotiator... {
				go handleNegotiations(pl)
				//}
			}
			if now-vm.whenFaulted > 40 {
				// if !vm(f+1).goodNegotiator {
				//fault(vm(f+1))
				//}
			}
		}
	}
}

func handleNegotiations(pl *ProcessList) {
	for {
		amINego := false
		faultIDs := pl.GetKeysFaultMap()
		for _, faultID := range faultIDs {
			faultState := pl.GetFaultState(faultID)
			if faultState.AmINegotiator {
				if faultState.NegotiationOngoing {
					craftAndSubmitFullFault(pl, faultID)
					amINego = true
					break
				}
			}
		}

		pl.AmINegotiator = amINego

		time.Sleep(3 * time.Second)
	}
}

func craftAndSubmitFullFault(pl *ProcessList, faultID [32]byte) {
	faultState := pl.GetFaultState(faultID)
	fc := faultState.FaultCore

	sf := messages.NewServerFault(pl.State.GetTimestamp(), fc.ServerID, fc.AuditServerID, int(fc.VMIndex), fc.DBHeight, fc.Height)

	var listOfSigs []interfaces.IFullSignature
	for _, sig := range faultState.VoteMap {
		listOfSigs = append(listOfSigs, sig)
	}

	if !pl.State.NetStateOff {
		fmt.Println("JUSTIN CRAFTANDSUB", pl.State.FactomNodeName, len(listOfSigs))
	}

	fullFault := messages.NewFullServerFault(sf, listOfSigs)
	absf := fullFault.ToAdminBlockEntry()
	pl.State.LeaderPL.AdminBlock.AddServerFault(absf)
	if fullFault != nil {
		fullFault.Sign(pl.State.serverPrivKey)
		pl.State.NetworkOutMsgQueue() <- fullFault
		fullFault.FollowerExecute(pl.State)
		//pl.AmINegotiator = false
		//delete(pl.FaultMap, faultID)
	}
}

func craftFault(pl *ProcessList, vm *VM, vmIndex int, height int) *messages.ServerFault {
	// TODO: if I am the Leader being faulted, I should respond by sending out
	// a MissingMsgResponse to everyone for the msg I'm being faulted for
	auditServerList := pl.State.GetOnlineAuditServers(pl.DBHeight)
	if len(auditServerList) > 0 {
		replacementServer := auditServerList[0]
		leaderMin := getLeaderMin(pl)

		faultedFed := pl.ServerMap[leaderMin][vmIndex]

		pl.FedServers[faultedFed].SetOnline(false)
		id := pl.FedServers[faultedFed].GetChainID()
		//NOMINATE
		sf := messages.NewServerFault(pl.State.GetTimestamp(), id, replacementServer.GetChainID(), vmIndex, pl.DBHeight, uint32(height))
		if sf != nil {
			sf.Sign(pl.State.serverPrivKey)
			return sf
		}
	}
	return nil
}

func craftAndSubmitFault(pl *ProcessList, vm *VM, vmIndex int, height int) {
	if !pl.State.NetStateOff {
		fmt.Println("JUSTIN CRASF", pl.State.FactomNodeName)
	}
	// TODO: if I am the Leader being faulted, I should respond by sending out
	// a MissingMsgResponse to everyone for the msg I'm being faulted for
	auditServerList := pl.State.GetOnlineAuditServers(pl.DBHeight)
	if len(auditServerList) > 0 {
		replacementServer := auditServerList[0]
		leaderMin := getLeaderMin(pl)

		faultedFed := pl.ServerMap[leaderMin][vmIndex]

		pl.FedServers[faultedFed].SetOnline(false)
		id := pl.FedServers[faultedFed].GetChainID()
		fmt.Println("JUSTIN CRASF2:", pl.State.FactomNodeName, id.String(), replacementServer.GetChainID().String())
		//NOMINATE
		sf := messages.NewServerFault(pl.State.GetTimestamp(), id, replacementServer.GetChainID(), vmIndex, pl.DBHeight, uint32(height))
		if sf != nil {
			sf.Sign(pl.State.serverPrivKey)
			pl.State.NetworkOutMsgQueue() <- sf
			pl.State.InMsgQueue() <- sf
			fm := pl.GetFaultState(sf.GetCoreHash().Fixed())
			fm.LastMatch = time.Now().Unix()
			pl.AddFaultState(sf.GetCoreHash().Fixed(), fm)

		}
	} else {
		for _, aud := range pl.AuditServers {
			aud.SetOnline(true)
		}
	}
}

func (s *State) FollowerExecuteSFault(m interfaces.IMsg) {
	sf, _ := m.(*messages.ServerFault)
	pl := s.ProcessLists.Get(sf.DBHeight)

	if pl == nil {
		return
	}

	if pl.VMs[sf.VMIndex].whenFaulted == 0 {
		return
	}

	s.regularFaultExecution(sf, pl)
}

func (s *State) regularFaultExecution(sf *messages.ServerFault, pl *ProcessList) {
	var issuerID [32]byte
	rawIssuerID := sf.GetSignature().GetKey()
	for i := 0; i < 32; i++ {
		if i < len(rawIssuerID) {
			issuerID[i] = rawIssuerID[i]
		}
	}
	var fedServerCnt int

	if pl != nil {
		fedServerCnt = len(pl.FedServers)
	} else {
		fedServerCnt = len(s.GetFedServers(sf.DBHeight))
	}
	responsibleFaulterIdx := (int(sf.VMIndex) + 1) % fedServerCnt

	//coreHash := sf.GetCoreHash().Fixed()
	if !s.NetStateOff {
		if s.Leader {
			fmt.Println("JUSTIN COREH:", pl.State.FactomNodeName, sf.GetCoreHash().String()[:10], pl.LenFaultMap())
		}
	}
	faultState := pl.GetFaultState(sf.GetCoreHash().Fixed())
	if faultState.FaultCore.Height > 0 {
		if s.Leader {
			if !s.NetStateOff {
				fmt.Println("JUSTIN FROMAP:", s.FactomNodeName, faultState)
			}
		}
	} else {
		fcore := FaultCore{ServerID: sf.ServerID, AuditServerID: sf.AuditServerID, VMIndex: sf.VMIndex, DBHeight: sf.DBHeight, Height: sf.Height}
		faultState = FaultState{FaultCore: fcore, AmINegotiator: false, MyVoteTallied: false, VoteMap: make(map[[32]byte]interfaces.IFullSignature), NegotiationOngoing: false}

		if s.Leader && s.LeaderVMIndex == responsibleFaulterIdx {
			faultState.AmINegotiator = true
			faultState.NegotiationOngoing = true
			pl.AmINegotiator = true
			go handleNegotiations(pl)
		}

		if faultState.VoteMap == nil {
			faultState.VoteMap = make(map[[32]byte]interfaces.IFullSignature)
		}
		pl.AddFaultState(sf.GetCoreHash().Fixed(), faultState)
		//pl.FaultMap[coreHash] = faultState

	}

	lbytes, err := sf.MarshalForSignature()

	sfSig := sf.Signature.GetSignature()

	isPledge := false
	auth, _ := s.GetAuthority(sf.AuditServerID)
	if auth == nil {
		isPledge = false
	} else {
		valid, err := auth.VerifySignature(lbytes, sfSig)
		if err == nil && valid {
			isPledge = true
			faultState.PledgeDone = true
		}
	}

	sfSigned, err := s.VerifyAuthoritySignature(lbytes, sfSig, sf.DBHeight)

	if err == nil && (sfSigned > 0 || (sfSigned == 0 && isPledge)) {
		faultState.VoteMap[issuerID] = sf.GetSignature()
	} else {
		return
	}

	if s.Leader || s.IdentityChainID.IsSameAs(sf.AuditServerID) {
		if !faultState.MyVoteTallied {
			fmt.Println("JUSTIN MM:", time.Now().Unix(), faultState.LastMatch)
			if time.Now().Unix()-faultState.LastMatch > 3 {
				s.matchFault(sf)
				faultState.LastMatch = time.Now().Unix()
			}
		}
	}

	//pl.FaultMap[sf.GetCoreHash().Fixed()] = faultState
	pl.AddFaultState(sf.GetCoreHash().Fixed(), faultState)
}

func (s *State) matchFault(sf *messages.ServerFault) {
	if sf != nil {
		sf.Sign(s.serverPrivKey)
		s.NetworkOutMsgQueue() <- sf
		s.InMsgQueue() <- sf
		if !s.NetStateOff {
			fmt.Println("JUSTIN MATCHED FAULT", s.FactomNodeName, sf.GetCoreHash().String()[:10])
		}
	}
}

func wipeOutFaultsFor(pl *ProcessList, faultedServerID interfaces.IHash, promotedServerID interfaces.IHash) {

	faultIDs := pl.GetKeysFaultMap()
	for _, faultID := range faultIDs {
		faultState := pl.GetFaultState(faultID)
		if faultState.FaultCore.ServerID.IsSameAs(faultedServerID) {
			faultState.NegotiationOngoing = false
			//delete(pl.FaultMap, faultID)
			//pl.FaultMap[faultID] = faultState
			pl.AddFaultState(faultID, faultState)
		}
		if faultState.FaultCore.AuditServerID.IsSameAs(promotedServerID) {
			faultState.NegotiationOngoing = false
			//delete(pl.FaultMap, faultID)
			//pl.FaultMap[faultID] = faultState
			pl.AddFaultState(faultID, faultState)
		}
	}
	amINego := false
	for _, faultID := range faultIDs {
		faultState := pl.GetFaultState(faultID)
		if faultState.AmINegotiator && faultState.NegotiationOngoing {
			fmt.Println("JUSTIN", pl.State.FactomNodeName, "SETTING AMINEGO TRUE DUE TO", faultState.FaultCore.ServerID.String()[:10], faultState.FaultCore.AuditServerID.String()[:10])
			amINego = true
		}
	}
	fmt.Println("JUSTIN", pl.State.FactomNodeName, "SETTING AMINEGO", amINego)
	pl.AmINegotiator = amINego
}

func (s *State) FollowerExecuteFullFault(m interfaces.IMsg) {
	fullFault, _ := m.(*messages.FullServerFault)
	relevantPL := s.ProcessLists.Get(fullFault.DBHeight)
	fmt.Println("JUSTIN FFFFF", s.FactomNodeName, fullFault.GetCoreHash().String()[:10], len(fullFault.SignatureList.List))

	auditServerList := s.GetAuditServers(fullFault.DBHeight)
	var theAuditReplacement interfaces.IFctServer

	for _, auditServer := range auditServerList {
		if auditServer.GetChainID().IsSameAs(fullFault.AuditServerID) {
			theAuditReplacement = auditServer
		}
	}
	if theAuditReplacement == nil {
		return
	}

	theFaultState := relevantPL.GetFaultState(fullFault.GetCoreHash().Fixed())
	//relevantPL.FaultMap[fullFault.GetCoreHash().Fixed()]
	theFaultState.NegotiationOngoing = true
	//relevantPL.FaultMap[fullFault.GetCoreHash().Fixed()] = theFaultState
	relevantPL.AddFaultState(fullFault.GetCoreHash().Fixed(), theFaultState)

	hasSignatureQuorum := fullFault.HasEnoughSigs(s)
	if hasSignatureQuorum > 0 {
		if s.pledgedByAudit(fullFault) {
			fmt.Println("JUSTIN EXECUTING FULLFAULT", s.FactomNodeName, fullFault.ServerID.String()[:10], fullFault.AuditServerID.String()[:10])
			for listIdx, fedServ := range relevantPL.FedServers {
				if fedServ.GetChainID().IsSameAs(fullFault.ServerID) {
					relevantPL.FedServers[listIdx] = theAuditReplacement
					relevantPL.FedServers[listIdx].SetOnline(true)
					relevantPL.AddAuditServer(fedServ.GetChainID())
					s.RemoveAuditServer(fullFault.DBHeight, theAuditReplacement.GetChainID())
					if foundVM, vmindex := relevantPL.GetVirtualServers(s.CurrentMinute, theAuditReplacement.GetChainID()); foundVM {
						relevantPL.VMs[vmindex].faultHeight = -1
						relevantPL.VMs[vmindex].faultingEOM = 0
						relevantPL.VMs[vmindex].whenFaulted = 0
					}
					wipeOutFaultsFor(relevantPL, fullFault.ServerID, fullFault.AuditServerID)
					fmt.Println("JUSTIN NOW", s.FactomNodeName, "FF:", relevantPL.LenFaultMap())

					break
				}
			}

			s.Leader, s.LeaderVMIndex = s.LeaderPL.GetVirtualServers(s.CurrentMinute, s.IdentityChainID)
			//tempFaultState := relevantPL.FaultMap[fullFault.GetCoreHash().Fixed()]
			tempFaultState := relevantPL.GetFaultState(fullFault.GetCoreHash().Fixed())
			tempFaultState.NegotiationOngoing = false
			//relevantPL.FaultMap[fullFault.GetCoreHash().Fixed()] = tempFaultState
			relevantPL.AddFaultState(fullFault.GetCoreHash().Fixed(), tempFaultState)

			amLeader, myLeaderVMIndex := s.LeaderPL.GetVirtualServers(s.CurrentMinute, s.IdentityChainID)

			if amLeader && myLeaderVMIndex == int(fullFault.VMIndex)+1%(len(relevantPL.FedServers)-1) {
				fmt.Println("JUSTIN POSTEX", s.FactomNodeName, "SETTING AMINEGO false")
				relevantPL.AmINegotiator = false
				fmt.Println("JUSTIN POSTEXX", s.FactomNodeName, "SETTING", time.Now().Unix())
			} else {
				fmt.Println("JUSTIN POSTEX", s.FactomNodeName, "CHECK:", myLeaderVMIndex, int(fullFault.VMIndex)+1%(len(relevantPL.FedServers)-1))
			}
			return
		} else {
			fmt.Println("JUSTIN PLEDGE FAIL", s.FactomNodeName, fullFault.AuditServerID.String()[:10])

		}
	} else if hasSignatureQuorum == 0 {
		fmt.Println("JUSTIN not enough sigs!", s.FactomNodeName, fullFault.GetCoreHash().String()[:10])
	}

	if !theFaultState.MyVoteTallied {
		nsf := messages.NewServerFault(s.GetTimestamp(), fullFault.ServerID, fullFault.AuditServerID, int(fullFault.VMIndex), fullFault.DBHeight, fullFault.Height)
		lbytes, err := nsf.MarshalForSignature() //fullFault.MarshalForSF()
		auth, _ := s.GetAuthority(s.IdentityChainID)
		if auth == nil || err != nil {
			return
		}
		for _, sig := range fullFault.SignatureList.List {
			ffSig := sig.GetSignature()
			valid, err := auth.VerifySignature(lbytes, ffSig)
			if err == nil && valid {
				theFaultState.MyVoteTallied = true
				//relevantPL.FaultMap[fullFault.GetCoreHash().Fixed()] = theFaultState
				relevantPL.AddFaultState(fullFault.GetCoreHash().Fixed(), theFaultState)
				fmt.Println("JUSTIN TALLIED", s.FactomNodeName, fullFault.GetCoreHash().String()[:10])
				return
			}
		}
	}
}

func (s *State) pledgedByAudit(fullFault *messages.FullServerFault) bool {
	for _, a := range s.Authorities {
		if a.AuthorityChainID.IsSameAs(fullFault.AuditServerID) {
			marshalledSF, err := fullFault.MarshalForSF()
			if err == nil {
				for _, sig := range fullFault.SignatureList.List {
					sigVer, err := a.VerifySignature(marshalledSF, sig.GetSignature())
					if err == nil && sigVer {
						return true
					}
				}
			}
			break
		}
	}
	return false
}
