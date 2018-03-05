// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"

	log "github.com/sirupsen/logrus"
)

var faultLogger = packageLogger.WithFields(log.Fields{"subpack": "fault"})

type FaultCore struct {
	// The following 5 fields represent the "Core" of the message
	// This should match the Core of FullServerFault messages
	ServerID      interfaces.IHash
	AuditServerID interfaces.IHash
	VMIndex       byte
	DBHeight      uint32
	Height        uint32
	SystemHeight  uint32
	Timestamp     interfaces.Timestamp
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
	binary.Write(&buf, binary.BigEndian, uint32(fc.SystemHeight))

	if d, err := fc.Timestamp.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}

	return buf.DeepCopyBytes(), nil
}

func markFault(pl *ProcessList, vmIndex int, faultReason int) {
	// We can use the "IgnoreMissing" boolean to track if enough time has elapsed
	// since bootup to start faulting servers on the network
	if pl.State.IgnoreMissing {
		return
	}

	if pl.State.Leader && pl.State.LeaderVMIndex == vmIndex {
		return
	}

	now := time.Now().Unix()
	vm := pl.VMs[vmIndex]

	if vm.WhenFaulted == 0 {
		// if we did not previously consider this VM faulted
		// we simply mark it as faulted (by assigning it a nonzero WhenFaulted time)
		// and keep track of the ProcessList height it has faulted at
		vm.WhenFaulted = now
		vm.FaultFlag = faultReason
	}

	c := pl.State.CurrentMinute
	if c > 9 {
		c = 9
	}
	index := pl.ServerMap[c][vmIndex]
	if index < len(pl.FedServers) {
		pl.FedServers[index].SetOnline(false)
	}
}

func markNoFault(pl *ProcessList, vmIndex int) {
	vm := pl.VMs[vmIndex]

	vm.WhenFaulted = 0
	vm.FaultFlag = -1

	nextIndex := (vmIndex + 1) % len(pl.FedServers)
	if pl.VMs[nextIndex].FaultFlag > 0 {
		markNoFault(pl, nextIndex)
	}

	c := pl.State.CurrentMinute
	if c > 9 {
		c = 9
	}
	index := pl.ServerMap[c][vmIndex]
	if index < len(pl.FedServers) {
		pl.FedServers[index].SetOnline(true)
	}

	cf := pl.CurrentFault()
	if !cf.IsNil() {
		if cf.AmINegotiator {
			ff := CraftFullFault(pl, vmIndex, vm.Height)
			if ff != nil {
				ff.Sign(pl.State.serverPrivKey)
				ff.SendOut(pl.State, ff)
				ff.FollowerExecute(pl.State)
			}
		}
	}

}

func NegotiationCheck(pl *ProcessList) {
	if !pl.State.Leader {
		//If I'm not a leader, do not attempt to negotiate
		return
	}
	prevIdx := precedingVMIndex(pl)
	prevVM := pl.VMs[prevIdx]

	if prevVM.WhenFaulted == 0 {
		//If the VM before me is not faulted, do not attempt
		//to negotiate
		return
	}

	now := time.Now().Unix()
	if now-prevVM.WhenFaulted < int64(pl.State.FaultTimeout) {
		//It hasn't been long enough; wait a little longer
		//before starting negotiation
		return
	}

	if now-pl.State.LastFaultAction > int64(pl.State.FaultWait) {
		//THROTTLE
		ff := CraftFullFault(pl, prevIdx, prevVM.Height)
		if ff != nil {
			ff.Sign(pl.State.serverPrivKey)
			ff.SendOut(pl.State, ff)
			ff.FollowerExecute(pl.State)
		}
		//pl.State.AddStatus(fmt.Sprintf("Sending Negotiation message (because %d) at %d since LFA=%d: %s", prevVM.FaultFlag, now, pl.State.LastFaultAction, ff.String()))
		pl.State.LastFaultAction = now
	}

	return
}

func FaultCheck(pl *ProcessList) {
	NegotiationCheck(pl)

	now := time.Now().Unix()

	currentFault := pl.CurrentFault()
	if currentFault.IsNil() {
		//Do not have a current fault
		pl.SetAmINegotiator(false)

		for i := 0; i < len(pl.FedServers); i++ {
			if i == pl.State.LeaderVMIndex {
				continue
			}
			vm := pl.VMs[i]
			if vm.WhenFaulted > 0 && int(now-vm.WhenFaulted) > pl.State.FaultTimeout*2 {
				newVMI := (i + 1) % len(pl.FedServers)
				markFault(pl, newVMI, 1)
			}

		}
		return
	}

	//If we are here, we have a non-nil CurrentFault

	timeElapsed := now - currentFault.Timestamp.GetTimeSeconds()
	currentFaultCore := ExtractFaultCore(currentFault)
	if isMyNegotiation(currentFaultCore, pl) {
		pl.SetAmINegotiator(true)
		if int(timeElapsed) > pl.State.FaultTimeout {
			if !currentFault.GetPledgeDone() {
				ToggleAuditOffline(pl, currentFaultCore)
			}
			pl.State.LastFaultAction = 0
			NegotiationCheck(pl)
		}
		return
	}

	pl.SetAmINegotiator(false)

	if int(timeElapsed) > pl.State.FaultTimeout*2 {
		// The negotiation has expired; time to fault negotiator
		newVMI := (int(currentFault.VMIndex) + 1) % len(pl.FedServers)
		markFault(pl, newVMI, 1)
	}
}

func precedingVMIndex(pl *ProcessList) int {
	precedingIndex := pl.State.LeaderVMIndex - 1
	if precedingIndex < 0 {
		precedingIndex = len(pl.FedServers) - 1
	}
	return precedingIndex
}

func ToggleAuditOffline(pl *ProcessList, fc FaultCore) {
	auditServerList := pl.State.GetAuditServers(fc.DBHeight)
	var theAuditReplacement interfaces.IServer

	for _, auditServer := range auditServerList {
		if auditServer.GetChainID().IsSameAs(fc.AuditServerID) {
			theAuditReplacement = auditServer
		}
	}
	if theAuditReplacement != nil {
		theAuditReplacement.SetOnline(false)
	}
}

func CraftFault(pl *ProcessList, vmIndex int, height int) *messages.ServerFault {
	// TODO: if I am the Leader being faulted, I should respond by sending out
	// a MissingMsgResponse to everyone for the msg I'm being faulted for

	// Only consider Online Audit servers as candidates for promotion (this
	// allows us to cycle through Audits on successive calls to CraftFullFault,
	// so that we make sure to (eventually) find one that is ready and able to
	// accept the promotion)
	auditServerList := pl.State.GetOnlineAuditServers(pl.DBHeight)
	if len(auditServerList) > 0 {
		// Nominate the top candidate from the list of Online Audit Servers
		audIdx := rand.Int() % len(auditServerList)
		replacementServer := auditServerList[audIdx]
		leaderMin := pl.State.CurrentMinute

		faultedFed := pl.ServerMap[leaderMin][vmIndex]

		if faultedFed >= len(pl.FedServers) {
			return nil
		}
		pl.FedServers[faultedFed].SetOnline(false)
		faultedFedID := pl.FedServers[faultedFed].GetChainID()

		// Create and send ServerFault (vote) message
		sf := messages.NewServerFault(faultedFedID, replacementServer.GetChainID(), vmIndex, pl.DBHeight, uint32(height), pl.System.Height, pl.State.GetTimestamp())
		if sf != nil {
			sf.Sign(pl.State.serverPrivKey)
			return sf
		}
	} else {
		// If we don't see any Audit servers as Online, we reset all of
		// them to an Online state and start the cycle anew
		for _, aud := range pl.AuditServers {
			aud.SetOnline(true)
		}
	}
	return nil
}

// CraftFullFault is called from the Negotiate check from the process list
// (which fires once every 5 seconds on each server); most of the time
// these are "incomplete" FullFault messages which serve as status pings
// for the negotiation in progress
func CraftFullFault(pl *ProcessList, vmIndex int, height int) *messages.FullServerFault {
	faultState := pl.CurrentFault()
	var sf *messages.ServerFault
	var listOfSigs []interfaces.IFullSignature
	var prevFF *messages.FullServerFault
	if pl.System.Height > 0 {
		prevFF = pl.System.List[pl.System.Height-1].(*messages.FullServerFault)
	}

	now := time.Now().Unix()

	if faultState.IsNil() || (now-faultState.GetTimestamp().GetTimeSeconds() > int64(pl.State.FaultTimeout)) && !(faultState.HasEnoughSigs(pl.State) && faultState.GetPledgeDone()) {
		sf = CraftFault(pl, vmIndex, height)
		if sf == nil {
			return nil
		}
		listOfSigs = append(listOfSigs, sf.Signature)
	} else {
		fc := ExtractFaultCore(faultState)
		sf = messages.NewServerFault(fc.ServerID, fc.AuditServerID, int(fc.VMIndex), fc.DBHeight, fc.Height, pl.System.Height, fc.Timestamp)
		for _, sig := range faultState.LocalVoteMap {
			listOfSigs = append(listOfSigs, sig)
		}
		for _, sig := range faultState.SignatureList.List {
			listOfSigs = append(listOfSigs, sig)
		}

	}

	fullFault := messages.NewFullServerFault(prevFF, sf, listOfSigs, pl.System.Height)
	fullFault.SetAmINegotiator(true)
	if pl.VMs[vmIndex].WhenFaulted == 0 {
		fullFault.ClearFault = true
	}

	return fullFault
}

func (s *State) FollowerExecuteSFault(m interfaces.IMsg) {
	sf, ok := m.(*messages.ServerFault)

	if !ok {
		return
	}

	pl := s.ProcessLists.Get(sf.DBHeight)

	if pl == nil || pl.VMs[sf.VMIndex].WhenFaulted == 0 {
		// If no such ProcessList exists, or if we don't consider
		// the VM in this ServerFault message to be at fault,
		// do not proceed with regularFaultExecution
		s.Holding[m.GetMsgHash().Fixed()] = m
		return
	}

	var issuerID [32]byte
	rawIssuerID := sf.GetSignature().GetKey()
	for i := 0; i < 32; i++ {
		if i < len(rawIssuerID) {
			issuerID[i] = rawIssuerID[i]
		}
	}

	currentFault := pl.CurrentFault()
	if currentFault.IsNil() {
		return
	}

	currentFaultCore := ExtractFaultCore(currentFault)
	thisMessageFaultCore := ExtractFaultCore(sf)

	if !currentFaultCore.GetHash().IsSameAs(thisMessageFaultCore.GetHash()) {
		return
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
			currentFault.SetPledgeDone(true)
		}
	}

	sfSigned, err := s.FastVerifyAuthoritySignature(lbytes, sf.Signature, sf.DBHeight)

	if err == nil && (sfSigned > 0 || (sfSigned == 0 && isPledge)) {
		currentFault.AddFaultVote(issuerID, sf.GetSignature())
	}
}

func ExtractFaultCore(sfMsg interfaces.IMsg) FaultCore {
	sf, ok := sfMsg.(*messages.ServerFault)
	if !ok {
		sf, ok2 := sfMsg.(*messages.FullServerFault)
		if !ok2 {
			return *new(FaultCore)
		}
		return FaultCore{ServerID: sf.ServerID, AuditServerID: sf.AuditServerID, VMIndex: sf.VMIndex, DBHeight: sf.DBHeight, Height: sf.Height, SystemHeight: sf.SystemHeight, Timestamp: sf.Timestamp}
	}
	return FaultCore{ServerID: sf.ServerID, AuditServerID: sf.AuditServerID, VMIndex: sf.VMIndex, DBHeight: sf.DBHeight, Height: sf.Height, SystemHeight: sf.SystemHeight, Timestamp: sf.Timestamp}
}

func isMyNegotiation(sf FaultCore, pl *ProcessList) bool {
	var fedServerCnt int

	if pl != nil {
		fedServerCnt = len(pl.FedServers)
	} else {
		fedServerCnt = len(pl.State.GetFedServers(sf.DBHeight))
	}
	responsibleFaulterIdx := (int(sf.VMIndex) + 1) % fedServerCnt

	if pl.State.Leader && pl.State.LeaderVMIndex == responsibleFaulterIdx {
		return true
	}
	return false
}

// matchFault does what it sounds like; given a particular ServerFault
// message, it will copy it, sign it, and send it out to the network
func (s *State) matchFault(sf *messages.ServerFault) {
	if sf != nil {
		sf.Sign(s.serverPrivKey)
		sf.SendOut(s, sf)
		s.InMsgQueue().Enqueue(sf)
	}
}

// When we execute a FullFault message, it could be complete (includes all
// necessary signatures + pledge) or incomplete, in which case it is just
// a negotiation ping
func (s *State) FollowerExecuteFullFault(m interfaces.IMsg) {
	fullFault, _ := m.(*messages.FullServerFault)

	pl := s.ProcessLists.Get(fullFault.DBHeight)

	if pl == nil {
		s.Holding[m.GetMsgHash().Fixed()] = m
		return
	}

	if !pl.CurrentFault().IsNil() && fullFault.GetHash().IsSameAs(pl.CurrentFault().GetHash()) {
		//No need to re-add (just fills up State status unnecessarily)
		return
	}

	//s.AddStatus(fmt.Sprintf("FULL FAULT FOLLOWER EXECUTE Execute Full Fault:  Replacing %x with %x at height %d leader height %d %s",
	//	fullFault.ServerID.Bytes()[2:6],
	//	fullFault.AuditServerID.Bytes()[2:6],
	//	fullFault.DBHeight,
	//	s.LLeaderHeight,
	//	fullFault.String()))

	faultLogger.WithField("func", "AddToSystemList").WithFields(fullFault.LogFields()).Warn("Add to System List")
	pl.AddToSystemList(fullFault)
}

// If a FullFault message includes a signature from the Audit server
// which was nominated in the Fault, pledgedByAudit will return true
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

func (s *State) Reset() {
	s.ResetRequest = true
}

// Set to reprocess all messages and states
func (s *State) DoReset() {
	s.ResetTryCnt++
	//s.AddStatus(fmt.Sprintf("RESET: Trying to Reset for the %d time", s.ResetTryCnt))
	index := len(s.DBStates.DBStates) - 1
	if index < 2 {
		//s.AddStatus("RESET: Failed to Reset because not enough dbstates")
		return
	}

	dbs := s.DBStates.DBStates[index]
	for {
		if dbs == nil || dbs.DirectoryBlock == nil || dbs.AdminBlock == nil || dbs.FactoidBlock == nil || dbs.EntryCreditBlock == nil {
			//s.AddStatus(fmt.Sprintf("RESET: Reset Failed, no dbstate at %d", index))
			return
		}
		if dbs.Saved {
			break
		}
		index--
		dbs = s.DBStates.DBStates[index]
	}
	if index < 0 {
		//s.AddStatus("RESET: Can't reset far enough back")
		return
	}
	s.ResetCnt++
	dbs = s.DBStates.DBStates[index-1]
	s.DBStates.DBStates = s.DBStates.DBStates[:index]

	dbs.AdminBlock = dbs.AdminBlock.New().(interfaces.IAdminBlock)
	dbs.FactoidBlock = dbs.FactoidBlock.New().(interfaces.IFBlock)

	plToReset := s.ProcessLists.Get(s.DBStates.Base + uint32(index) + 1)
	plToReset.Reset()
	s.DBStates.Complete--
	//s.StartDelay = s.GetTimestamp().GetTimeMilli() // We cant start as a leader until we know we are upto date
	//s.RunLeader = false
	s.CurrentMinute = 0

	s.SetLeaderTimestamp(dbs.NextTimestamp)

	s.DBStates.ProcessBlocks(dbs)
	faultLogger.WithFields(log.Fields{"func": "Reset", "count": s.ResetTryCnt}).Warn("DoReset complete")
	//s.AddStatus("RESET: Complete")
}
