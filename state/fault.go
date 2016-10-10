// Copyright 2016 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"

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

func (fs *FaultState) IsNil() bool {
	if fs.VoteMap == nil || fs.FaultCore.ServerID.IsZero() || fs.FaultCore.AuditServerID.IsZero() {
		return true
	}
	return false
}

func (fs *FaultState) HasEnoughSigs(state interfaces.IState) bool {
	cb, err := fs.FaultCore.MarshalCore()
	if err != nil {
		return false
	}
	validSigCount := 0
	for _, fedSig := range fs.VoteMap {
		check, err := state.FastVerifyAuthoritySignature(cb, fedSig, fs.FaultCore.DBHeight)
		if err == nil && check == 1 {
			validSigCount++
		}
		if validSigCount > len(state.GetFedServers(fs.FaultCore.DBHeight))/2 {
			return true
		}
	}
	return false
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

func (fc *FaultCore) GetHash() interfaces.IHash {
	data, err := fc.MarshalCore()
	if err != nil {
		return nil
	}
	return primitives.Sha(data)
}

func (fc *FaultCore) MarshalCore() (data []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error marshalling Server Fault Core: %v", r)
		}
	}()

	var buf primitives.Buffer

	if d, err := fc.ServerID.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}
	if d, err := fc.AuditServerID.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}

	buf.WriteByte(fc.VMIndex)
	binary.Write(&buf, binary.BigEndian, uint32(fc.DBHeight))
	binary.Write(&buf, binary.BigEndian, uint32(fc.Height))

	return buf.DeepCopyBytes(), nil
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
			if vm.faultInitiatedAlready {
				if now-vm.whenFaulted > 40 {
					if now-vm.lastFaultAction > 3 {
						vm.lastFaultAction = now
						if !couldIFullFault(pl, vmIndex) {
							newVMI := (vmIndex + 1) % len(pl.FedServers)
							//fmt.Println("JUSTIN TIME TO FAULT FAULTER", pl.State.FactomNodeName, newVMI, now)
							fault(pl, pl.VMs[newVMI], newVMI, len(vm.List), 1)
						}
					}
				}
			} else {
				vm.lastFaultAction = now
				// after 20 seconds, we take initiative and
				// issue a server fault vote of our own
				craftAndSubmitFault(pl, vm, vmIndex, height)
				vm.faultInitiatedAlready = true
				//if I am negotiator... {
				//go handleNegotiations(pl)
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
					CraftAndSubmitFullFault(pl, faultID)
					amINego = true
					break
				}
			}
		}

		pl.AmINegotiator = amINego

		time.Sleep(5 * time.Second)
	}
}

func couldIFullFault(pl *ProcessList, vmIndex int) bool {
	leaderMin := getLeaderMin(pl)
	faultedFed := pl.ServerMap[leaderMin][vmIndex]
	id := pl.FedServers[faultedFed].GetChainID()
	stringid := id.String()

	if pl.LenFaultMap() < 1 {
		return true
	}

	faultIDs := pl.GetKeysFaultMap()
	for _, faultID := range faultIDs {
		faultState := pl.GetFaultState(faultID)
		if !faultState.AmINegotiator {
			faultedServerFromFaultState := faultState.FaultCore.ServerID.String()
			if faultedServerFromFaultState == stringid {
				if faultState.PledgeDone && faultState.HasEnoughSigs(pl.State) {
					return false
				}
				/*if faultState.NegotiationOngoing {
					anyNegotiating = true
					if faultState.PledgeDone && faultState.HasEnoughSigs(pl.State) {
						return false
					}
				}*/
			}
		}
	}

	return true
}

func CraftAndSubmitFullFault(pl *ProcessList, faultID [32]byte) {
	fmt.Printf("JUSTIN CRASFF %s %x\n", pl.State.FactomNodeName, faultID)

	faultState := pl.GetFaultState(faultID)
	fc := faultState.FaultCore

	sf := messages.NewServerFault(pl.State.GetTimestamp(), fc.ServerID, fc.AuditServerID, int(fc.VMIndex), fc.DBHeight, fc.Height)

	var listOfSigs []interfaces.IFullSignature
	for _, sig := range faultState.VoteMap {
		listOfSigs = append(listOfSigs, sig)
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

/*func craftFault(pl *ProcessList, vm *VM, vmIndex int, height int) *messages.ServerFault {
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
}*/

func craftAndSubmitFault(pl *ProcessList, vm *VM, vmIndex int, height int) {
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
			pl.State.NetworkOutMsgQueue() <- sf
			pl.State.InMsgQueue() <- sf
			fm := pl.GetFaultState(sf.GetCoreHash().Fixed())
			if !fm.IsNil() {
				fm.LastMatch = time.Now().Unix()
				pl.AddFaultState(sf.GetCoreHash().Fixed(), fm)
			}
			// Now let's set the Audit Server offline (so we don't just re-nominate them over and over)
			replacementServer.SetOnline(false)
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

	coreHash := sf.GetCoreHash().Fixed()
	faultState := pl.GetFaultState(coreHash)
	if faultState.IsNil() {
		// We don't have a map entry yet; let's create one
		fcore := FaultCore{ServerID: sf.ServerID, AuditServerID: sf.AuditServerID, VMIndex: sf.VMIndex, DBHeight: sf.DBHeight, Height: sf.Height}
		faultState = FaultState{FaultCore: fcore, AmINegotiator: false, MyVoteTallied: false, VoteMap: make(map[[32]byte]interfaces.IFullSignature), NegotiationOngoing: false}

		if s.Leader && s.LeaderVMIndex == responsibleFaulterIdx {
			faultState.AmINegotiator = true
			faultState.NegotiationOngoing = true
			pl.AmINegotiator = true
			//go handleNegotiations(pl)
		}

		if faultState.VoteMap == nil {
			faultState.VoteMap = make(map[[32]byte]interfaces.IFullSignature)
		}
		pl.AddFaultState(coreHash, faultState)
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

	sfSigned, err := s.FastVerifyAuthoritySignature(lbytes, sf.Signature, sf.DBHeight)

	if err == nil && (sfSigned > 0 || (sfSigned == 0 && isPledge)) {
		faultState.VoteMap[issuerID] = sf.GetSignature()
	} else {
		return
	}

	if s.Leader || s.IdentityChainID.IsSameAs(sf.AuditServerID) {
		if !faultState.MyVoteTallied {
			if time.Now().Unix()-faultState.LastMatch > 3 {
				s.matchFault(sf)
				faultState.LastMatch = time.Now().Unix()
			}
		}
	}

	//pl.FaultMap[sf.GetCoreHash().Fixed()] = faultState
	pl.AddFaultState(coreHash, faultState)
}

func (s *State) regularFullFaultExecution(sf *messages.FullServerFault, pl *ProcessList) {
	coreHash := sf.GetCoreHash().Fixed()
	var fedServerCnt int

	if pl != nil {
		fedServerCnt = len(pl.FedServers)
	} else {
		fedServerCnt = len(s.GetFedServers(sf.DBHeight))
	}
	responsibleFaulterIdx := (int(sf.VMIndex) + 1) % fedServerCnt

	for _, signature := range sf.SignatureList.List {
		var issuerID [32]byte
		rawIssuerID := signature.GetKey()
		for i := 0; i < 32; i++ {
			if i < len(rawIssuerID) {
				issuerID[i] = rawIssuerID[i]
			}
		}

		faultState := pl.GetFaultState(coreHash)
		if !faultState.IsNil() {
			// We already have a map entry
		} else {
			// We don't have a map entry yet; let's create one
			fcore := FaultCore{ServerID: sf.ServerID, AuditServerID: sf.AuditServerID, VMIndex: sf.VMIndex, DBHeight: sf.DBHeight, Height: sf.Height}
			faultState = FaultState{FaultCore: fcore, AmINegotiator: false, MyVoteTallied: false, VoteMap: make(map[[32]byte]interfaces.IFullSignature), NegotiationOngoing: false}

			if s.Leader && s.LeaderVMIndex == responsibleFaulterIdx {
				faultState.AmINegotiator = true
				faultState.NegotiationOngoing = true
				pl.AmINegotiator = true
			}

			if faultState.VoteMap == nil {
				faultState.VoteMap = make(map[[32]byte]interfaces.IFullSignature)
			}
			pl.AddFaultState(coreHash, faultState)
		}

		if s.Leader || s.IdentityChainID.IsSameAs(sf.AuditServerID) {
			if !faultState.MyVoteTallied {
				nsf := messages.NewServerFault(s.GetTimestamp(), sf.ServerID, sf.AuditServerID, int(sf.VMIndex), sf.DBHeight, sf.Height)
				sfbytes, err := nsf.MarshalForSignature() //fullFault.MarshalForSF()
				myAuth, _ := s.GetAuthority(s.IdentityChainID)
				if myAuth == nil || err != nil {
					return
				}
				valid, err := myAuth.VerifySignature(sfbytes, signature.GetSignature())
				if err == nil && valid {
					faultState.MyVoteTallied = true
					pl.AddFaultState(coreHash, faultState)
				}
			}
		}

		lbytes := sf.GetCoreHash().Bytes()

		isPledge := false
		auth, _ := s.GetAuthority(sf.AuditServerID)
		if auth == nil {
			isPledge = false
		} else {
			valid, err := auth.VerifySignature(lbytes, signature.GetSignature())
			if err == nil && valid {
				isPledge = true
				faultState.PledgeDone = true
				pl.AddFaultState(coreHash, faultState)
			}
		}

		sfSigned, err := s.FastVerifyAuthoritySignature(lbytes, signature, sf.DBHeight)

		if err == nil && (sfSigned > 0 || (sfSigned == 0 && isPledge)) {
			faultState.VoteMap[issuerID] = sf.GetSignature()
			pl.AddFaultState(coreHash, faultState)
		}
	}

	faultState := pl.GetFaultState(coreHash)
	if !faultState.IsNil() {
		if s.Leader || s.IdentityChainID.IsSameAs(sf.AuditServerID) {
			if !faultState.MyVoteTallied {
				nsf := messages.NewServerFault(s.GetTimestamp(), sf.ServerID, sf.AuditServerID, int(sf.VMIndex), sf.DBHeight, sf.Height)
				s.matchFault(nsf)
			}
		}
	}
}

func (s *State) matchFault(sf *messages.ServerFault) {
	if sf != nil {
		sf.Sign(s.serverPrivKey)
		s.NetworkOutMsgQueue() <- sf
		s.InMsgQueue() <- sf
	}
}

func (pl *ProcessList) Unfault() {
	// Delete all entries in FaultMap
	pl.ClearFaultMap()

	// Reset the fault-state of all VMs
	for i := 0; i < len(pl.FedServers); i++ {
		vm := pl.VMs[i]
		vm.faultHeight = -1
		vm.whenFaulted = 0
		vm.lastFaultAction = 0
		vm.faultInitiatedAlready = false
	}
	pl.AmINegotiator = false
}

func (pl *ProcessList) ClearFaultMap() {
	pl.FaultMapMutex.Lock()
	defer pl.FaultMapMutex.Unlock()
	pl.FaultMap = make(map[[32]byte]FaultState)
}

/*
func wipeOutFaultsFor(pl *ProcessList, faultedServerID interfaces.IHash, promotedServerID interfaces.IHash) {

	faultIDs := pl.GetKeysFaultMap()
	for _, faultID := range faultIDs {
		faultState := pl.GetFaultState(faultID)
		if faultState.FaultCore.ServerID.IsSameAs(faultedServerID) {
			//faultState.NegotiationOngoing = false
			//delete(pl.FaultMap, faultID)
			//pl.FaultMap[faultID] = faultState
			//pl.AddFaultState(faultID, faultState)
			pl.DeleteFaultState(faultID)
		}
		if faultState.FaultCore.AuditServerID.IsSameAs(promotedServerID) {
			//faultState.NegotiationOngoing = false
			//delete(pl.FaultMap, faultID)
			//pl.FaultMap[faultID] = faultState
			//pl.AddFaultState(faultID, faultState)
			pl.DeleteFaultState(faultID)
		}
	}
	amINego := false
	for _, faultID := range faultIDs {
		faultState := pl.GetFaultState(faultID)
		if faultState.AmINegotiator && faultState.NegotiationOngoing {
			amINego = true
		}
	}
	pl.AmINegotiator = amINego
}*/

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
	if theAuditReplacement == nil {
		return
	}

	/*theFaultState := relevantPL.GetFaultState(fullFault.GetCoreHash().Fixed())
	if !theFaultState.IsNil() {
		//relevantPL.FaultMap[fullFault.GetCoreHash().Fixed()]
		theFaultState.NegotiationOngoing = true
		//relevantPL.FaultMap[fullFault.GetCoreHash().Fixed()] = theFaultState
		relevantPL.AddFaultState(fullFault.GetCoreHash().Fixed(), theFaultState)
	}*/

	hasSignatureQuorum := fullFault.HasEnoughSigs(s)
	if hasSignatureQuorum > 0 {
		if s.pledgedByAudit(fullFault) {
			//fmt.Println("JUSTIN EXECUTING FULLFAULT", s.FactomNodeName, fullFault.ServerID.String()[:10], fullFault.AuditServerID.String()[:10])
			for listIdx, fedServ := range relevantPL.FedServers {
				if fedServ.GetChainID().IsSameAs(fullFault.ServerID) {
					fmt.Println("FULL FAULT:", s.FactomNodeName, fullFault.ServerID.String()[:10], fullFault.AuditServerID.String()[:10], s.GetTimestamp().GetTimeSeconds())
					relevantPL.FedServers[listIdx] = theAuditReplacement
					relevantPL.FedServers[listIdx].SetOnline(true)
					relevantPL.AddAuditServer(fedServ.GetChainID())
					s.RemoveAuditServer(fullFault.DBHeight, theAuditReplacement.GetChainID())

					/*
						if foundVM, vmindex := relevantPL.GetVirtualServers(s.CurrentMinute, theAuditReplacement.GetChainID()); foundVM {
							relevantPL.VMs[vmindex].faultHeight = -1
							//						relevantPL.VMs[vmindex].faultingEOM = 0
							relevantPL.VMs[vmindex].whenFaulted = 0
						}
						for _, vmReset := range relevantPL.VMs {
							vmReset.whenFaulted = 0
							vmReset.faultInitiatedAlready = false
						}
						wipeOutFaultsFor(relevantPL, fullFault.ServerID, fullFault.AuditServerID)*/
					relevantPL.Unfault()
					break
				}
			}

			s.Leader, s.LeaderVMIndex = s.LeaderPL.GetVirtualServers(s.CurrentMinute, s.IdentityChainID)
			//tempFaultState := relevantPL.FaultMap[fullFault.GetCoreHash().Fixed()]
			/*tempFaultState := relevantPL.GetFaultState(fullFault.GetCoreHash().Fixed())
			if !tempFaultState.IsNil() {
				tempFaultState.NegotiationOngoing = false
				//relevantPL.FaultMap[fullFault.GetCoreHash().Fixed()] = tempFaultState
				relevantPL.AddFaultState(fullFault.GetCoreHash().Fixed(), tempFaultState)
			}
			relevantPL.DeleteFaultState(fullFault.GetCoreHash().Fixed())

			amLeader, myLeaderVMIndex := s.LeaderPL.GetVirtualServers(s.CurrentMinute, s.IdentityChainID)

			if amLeader && myLeaderVMIndex == int(fullFault.VMIndex)+1%(len(relevantPL.FedServers)-1) {
				relevantPL.AmINegotiator = false
			}*/
			return
		} else {
			// MISSING A PLEDGE
			if s.IdentityChainID.IsSameAs(fullFault.AuditServerID) {
				sf := messages.NewServerFault(s.GetTimestamp(), fullFault.ServerID, fullFault.AuditServerID,
					int(fullFault.VMIndex), fullFault.DBHeight, fullFault.Height)
				s.matchFault(sf)
			} else {
				if s.Leader {
					if relevantPL.VMs[fullFault.VMIndex].whenFaulted > 0 {
						craftAndSubmitFault(relevantPL, relevantPL.VMs[fullFault.VMIndex], int(fullFault.VMIndex), int(fullFault.Height))
					}
					/*sf := messages.NewServerFault(s.GetTimestamp(), fullFault.ServerID, fullFault.AuditServerID,
						int(fullFault.VMIndex), fullFault.DBHeight, fullFault.Height)
					s.regularFaultExecution(sf, relevantPL)*/
				}
			}
		}
	} else if hasSignatureQuorum == 0 {
		// NOT ENOUGH SIGNATURES TO EXECUTE
		s.regularFullFaultExecution(fullFault, relevantPL)
	}

	/*
		if !theFaultState.IsNil() {
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
						return
					}
				}
			}
		}*/
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
