// Copyright 2016 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"

	"github.com/FactomProject/factomd/common/messages"
)

type FaultState struct {
	FaultCore     FaultCore
	AmINegotiator bool
	MyVoteTallied bool
	VoteMap       map[[32]byte]interfaces.IFullSignature
	PledgeDone    bool
	LastMatch     int64
}

var _ interfaces.IFaultState = (*FaultState)(nil)

func (fs *FaultState) GetAmINegotiator() bool {
	return fs.AmINegotiator
}

func (fs *FaultState) SetAmINegotiator(b bool) {
	fs.AmINegotiator = b
}

func (fs *FaultState) GetMyVoteTallied() bool {
	return fs.MyVoteTallied
}

func (fs *FaultState) SetMyVoteTallied(b bool) {
	fs.MyVoteTallied = b
}

func (fs *FaultState) GetPledgeDone() bool {
	return fs.PledgeDone
}

func (fs *FaultState) SetPledgeDone(b bool) {
	fs.PledgeDone = b
}

func (fs *FaultState) GetLastMatch() int64 {
	return fs.LastMatch
}

func (fs *FaultState) SetLastMatch(b int64) {
	fs.LastMatch = b
}

func (fs *FaultState) IsNil() bool {
	if fs == nil {
		return true
	}
	if fs.VoteMap == nil {
		return true
	}
	if (FaultCore{}) == fs.FaultCore {
		return true
	}
	if fs.FaultCore.ServerID.IsZero() {
		return true
	}
	if fs.FaultCore.AuditServerID == nil || fs.FaultCore.AuditServerID.IsZero() {
		return true
	}
	return false
}

func (fs *FaultState) AddFaultVote(issuerID [32]byte, sig interfaces.IFullSignature) {
	if fs.VoteMap == nil {
		fs.VoteMap = make(map[[32]byte]interfaces.IFullSignature)
	}

	fs.VoteMap[issuerID] = sig
}

func (fs *FaultState) SigTally(state interfaces.IState) int {
	validSigCount := 0
	cb, err := fs.FaultCore.MarshalCore()
	if err != nil {
		return validSigCount
	}
	for _, fedSig := range fs.VoteMap {
		check, err := state.FastVerifyAuthoritySignature(cb, fedSig, fs.FaultCore.DBHeight)
		if err == nil && check == 1 {
			validSigCount++
		}

	}
	return validSigCount
}

func (fs *FaultState) HasEnoughSigs(state interfaces.IState) bool {
	if fs.SigTally(state) > len(state.GetFedServers(fs.FaultCore.DBHeight))/2 {
		return true
	}
	return false
}

func (fs FaultState) String() string {
	return fmt.Sprintf("Fed: %s Audit: %s, VM: %d, Height: %d, AmINego: %v, MyVote: %v, Votes: %d, Pledged: %v TS:%d",
		fs.FaultCore.ServerID.String()[:10], fs.FaultCore.AuditServerID.String()[:10], int(fs.FaultCore.VMIndex), fs.FaultCore.Height,
		fs.AmINegotiator, fs.MyVoteTallied, len(fs.VoteMap), fs.PledgeDone, fs.FaultCore.Timestamp.GetTimeSeconds())
}

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

	cf := pl.CurrentFault
	if !cf.IsNil() {
		if int(cf.FaultCore.VMIndex) == vmIndex {
			if index < len(pl.FedServers) && pl.FedServers[index].GetChainID().IsSameAs(cf.FaultCore.ServerID) {
				return
			}
		}
		if cf.AmINegotiator && int(time.Now().Unix()-pl.State.LastFaultAction) > pl.State.FaultWait {
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
		pl.State.AddStatus(fmt.Sprintf("Sending Negotiation message (because %d): %s", prevVM.FaultFlag, ff.String()))
		pl.State.LastFaultAction = now
	}

	return
}

func FaultCheck(pl *ProcessList) {
	NegotiationCheck(pl)

	now := time.Now().Unix()

	faultState := pl.CurrentFault
	if faultState.IsNil() {
		//Do not have a current fault
		pl.SetAmINegotiator(false)

		if now-pl.NegotiatonTimeout > int64(pl.State.FaultTimeout)/2 {
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
		}
		return

	}

	//If we are here, we have a non-nil CurrentFault

	timeElapsed := now - faultState.FaultCore.Timestamp.GetTimeSeconds()

	if isMyNegotiation(faultState.FaultCore, pl) {
		pl.SetAmINegotiator(true)
		if int(timeElapsed) > pl.State.FaultTimeout {
			if !faultState.GetPledgeDone() {
				ToggleAuditOffline(pl, faultState.FaultCore)
			}
			pl.State.LastFaultAction = 0
			NegotiationCheck(pl)
		}
		return
	} else {
		pl.SetAmINegotiator(false)
	}

	if int(timeElapsed) > pl.State.FaultTimeout*2 {
		// The negotiation has expired; time to fault negotiator
		newVMI := (int(faultState.FaultCore.VMIndex) + 1) % len(pl.FedServers)
		markFault(pl, newVMI, 1)
	}
	return
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
	var theAuditReplacement interfaces.IFctServer

	for _, auditServer := range auditServerList {
		if auditServer.GetChainID().IsSameAs(fc.AuditServerID) {
			theAuditReplacement = auditServer
		}
	}
	if theAuditReplacement != nil {
		theAuditReplacement.SetOnline(false)
	}
}

// couldIFullFault is our check to see if there are any negotiations
// ongoing for a particular VM (leader), and if so, have we gathered
// enough ServerFaults (+the Audit Pledge) to issue a FullFault message
// and conclude the faulting process ourselves
func couldIFullFault(pl *ProcessList, vmIndex int) bool {
	leaderMin := getLeaderMin(pl)
	faultedFed := pl.ServerMap[leaderMin][vmIndex]
	id := pl.FedServers[faultedFed].GetChainID()
	stringid := id.String()

	faultState := pl.CurrentFault

	if faultState.IsNil() {
		return false
	}

	if !faultState.AmINegotiator {
		faultedServerFromFaultState := faultState.FaultCore.ServerID.String()
		if faultedServerFromFaultState == stringid {
			if faultState.PledgeDone && faultState.HasEnoughSigs(pl.State) {
				// if the above 2 conditions are satisfied, we could issue
				// a FullFault message (if we were the negotiator for this fault)
				return true
			}
		}
	}

	return false
}

func CraftFault(pl *ProcessList, vmIndex int, height int) {
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
			return
		}
		pl.FedServers[faultedFed].SetOnline(false)
		faultedFedID := pl.FedServers[faultedFed].GetChainID()

		// Create and send ServerFault (vote) message
		sf := messages.NewServerFault(faultedFedID, replacementServer.GetChainID(), vmIndex, pl.DBHeight, uint32(height), pl.System.Height, pl.State.GetTimestamp())
		if sf != nil {
			sf.Sign(pl.State.serverPrivKey)
			//pl.State.NetworkOutMsgQueue() <- sf
			pl.State.InMsgQueue() <- sf
			cf := pl.CurrentFault
			if !cf.IsNil() && cf.FaultCore.GetHash().IsSameAs(sf.GetCoreHash()) {
				// Update the CurrentFault's "LastMatch" value to the current time
				// (LastMatch is merely a throttling mechanism)
				cf.LastMatch = time.Now().Unix()
				pl.SetCurrentFault(cf)
			}
		}
	} else {
		// If we don't see any Audit servers as Online, we reset all of
		// them to an Online state and start the cycle anew
		for _, aud := range pl.AuditServers {
			aud.SetOnline(true)
		}
	}
}

// CraftFullFault is called from the Negotiate goroutine
// (which fires once every 5 seconds on each server); most of the time
// these are "incomplete" FullFault messages which serve as status pings
// for the negotiation in progress
func CraftFullFault(pl *ProcessList, vmIndex int, height int) *messages.FullServerFault {
	faultState := pl.CurrentFault
	if faultState.IsNil() {
		CraftFault(pl, vmIndex, height)
		return nil
	}

	fc := faultState.FaultCore

	sf := messages.NewServerFault(fc.ServerID, fc.AuditServerID, int(fc.VMIndex), fc.DBHeight, fc.Height, pl.System.Height, fc.Timestamp)

	var listOfSigs []interfaces.IFullSignature
	for _, sig := range faultState.VoteMap {
		listOfSigs = append(listOfSigs, sig)
	}
	var pff *messages.FullServerFault
	if pl.System.Height > 0 {
		pff = pl.System.List[pl.System.Height-1].(*messages.FullServerFault)
	}
	fullFault := messages.NewFullServerFault(pff, sf, listOfSigs, pl.System.Height)

	if pl.VMs[int(fc.VMIndex)].WhenFaulted == 0 {
		fullFault.ClearFault = true
	}

	//adminBlockEntryForFault := fullFault.ToAdminBlockEntry()
	//pl.State.LeaderPL.AdminBlock.AddServerFault(adminBlockEntryForFault)

	return fullFault
}

func (s *State) FollowerExecuteSFault(m interfaces.IMsg) {
	sf, _ := m.(*messages.ServerFault)
	pl := s.ProcessLists.Get(sf.DBHeight)

	if pl == nil || pl.VMs[sf.VMIndex].WhenFaulted == 0 {
		// If no such ProcessList exists, or if we don't consider
		// the VM in this ServerFault message to be at fault,
		// do not proceed with regularFaultExecution
		return
	}

	s.regularFaultExecution(sf, pl)
}

// regularFaultExecution will update CurrentFault and save the signature
// of the ServerFault to that FaultState's VoteMap
func (s *State) regularFaultExecution(sf *messages.ServerFault, pl *ProcessList) {
	var issuerID [32]byte
	rawIssuerID := sf.GetSignature().GetKey()
	for i := 0; i < 32; i++ {
		if i < len(rawIssuerID) {
			issuerID[i] = rawIssuerID[i]
		}
	}

	faultState := pl.CurrentFault
	if faultState.IsNil() || (faultState.FaultCore.ServerID.IsSameAs(sf.ServerID) && faultState.FaultCore.Timestamp.GetTimeSeconds() < sf.Timestamp.GetTimeSeconds()) {
		// We don't have a CurrentFault yet; let's create one
		pl.CreateFaultState(sf)
		faultState = pl.CurrentFault
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
		faultState.AddFaultVote(issuerID, sf.GetSignature())
	}

	if s.Leader || s.IdentityChainID.IsSameAs(sf.AuditServerID) {
		if !faultState.MyVoteTallied {
			// Don't fault yourself!
			if !sf.ServerID.IsSameAs(s.IdentityChainID) {
				now := time.Now().Unix()
				if now-faultState.LastMatch > 3 {
					// Don't send multiple tiebreaker votes for different faults
					// too quickly back-to-back
					if int(now-s.LastTiebreak) > s.FaultTimeout/2 {
						if faultState.SigTally(s) >= len(pl.FedServers)-1 {
							s.LastTiebreak = now
						}
						s.matchFault(sf)
						faultState.LastMatch = now
					}
				}
			}
		}
	}

	// Update the CurrentFault with any updates that were applied
	// during execution of this function
	pl.SetCurrentFault(faultState)
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

func (pl *ProcessList) CreateFaultState(sf interfaces.IMsg) {
	fcore := ExtractFaultCore(sf)
	if pl.State.Leader && pl.State.LeaderVMIndex == int(fcore.VMIndex) {
		return
	}
	faultState := FaultState{FaultCore: fcore, AmINegotiator: false, MyVoteTallied: false, VoteMap: make(map[[32]byte]interfaces.IFullSignature)}

	if isMyNegotiation(fcore, pl) {
		faultState.SetAmINegotiator(true)
	}

	if faultState.VoteMap == nil {
		faultState.VoteMap = make(map[[32]byte]interfaces.IFullSignature)
	}

	pl.SetCurrentFault(faultState)
}

// regularFullFaultExecution does the same thing as regularFaultExecution, except
// it will make sure to add every signature from the FullFault to the corresponding
// FaultState's VoteMap
func (s *State) regularFullFaultExecution(sf *messages.FullServerFault, pl *ProcessList) {
	for _, signature := range sf.SignatureList.List {
		var issuerID [32]byte
		rawIssuerID := signature.GetKey()
		for i := 0; i < 32; i++ {
			if i < len(rawIssuerID) {
				issuerID[i] = rawIssuerID[i]
			}
		}

		faultState := pl.CurrentFault
		if faultState.IsNil() {
			// We don't have a map entry yet; let's create one
			pl.CreateFaultState(sf)
			faultState = pl.CurrentFault
		}

		if s.Leader || s.IdentityChainID.IsSameAs(sf.AuditServerID) {
			if !faultState.MyVoteTallied {
				nsf := messages.NewServerFault(sf.ServerID, sf.AuditServerID, int(sf.VMIndex), sf.DBHeight, sf.Height, int(sf.SystemHeight), sf.Timestamp)
				sfbytes, err := nsf.MarshalForSignature()
				myAuth, _ := s.GetAuthority(s.IdentityChainID)
				if myAuth == nil || err != nil {
					return
				}
				valid, err := myAuth.VerifySignature(sfbytes, signature.GetSignature())
				if err == nil && valid {
					faultState.MyVoteTallied = true
					pl.SetCurrentFault(faultState)
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
				pl.SetCurrentFault(faultState)
			}
		}

		sfSigned, err := s.FastVerifyAuthoritySignature(lbytes, signature, sf.DBHeight)

		if err == nil && (sfSigned > 0 || (sfSigned == 0 && isPledge)) {
			faultState.AddFaultVote(issuerID, sf.GetSignature())
			pl.SetCurrentFault(faultState)
		}
	}

	faultState := pl.CurrentFault
	if !faultState.IsNil() {
		if s.Leader || s.IdentityChainID.IsSameAs(sf.AuditServerID) {
			if !faultState.MyVoteTallied {
				now := time.Now().Unix()
				if int(now-s.LastTiebreak) > s.FaultTimeout/2 {
					if faultState.SigTally(s) >= len(pl.FedServers)-1 {
						s.LastTiebreak = now
					}

					nsf := messages.NewServerFault(sf.ServerID, sf.AuditServerID, int(sf.VMIndex), sf.DBHeight, sf.Height, int(sf.SystemHeight), sf.Timestamp)
					s.matchFault(nsf)
				}
			}
		}
	}
}

// matchFault does what it sounds like; given a particular ServerFault
// message, it will copy it, sign it, and send it out to the network
func (s *State) matchFault(sf *messages.ServerFault) {
	if sf != nil {
		sf.Sign(s.serverPrivKey)
		sf.SendOut(s, sf)
		s.InMsgQueue() <- sf
	}
}

// When we execute a FullFault message, it could be complete (includes all
// necessary signatures + pledge) or incomplete, in which case it is just
// a negotiation ping
func (s *State) FollowerExecuteFullFault(m interfaces.IMsg) {
	fullFault, _ := m.(*messages.FullServerFault)

	pl := s.ProcessLists.Get(fullFault.DBHeight)

	if pl == nil {
		s.Holding[m.GetHash().Fixed()] = m
		return
	}

	s.AddStatus(fmt.Sprintf("FULL FAULT FOLLOWER EXECUTE Execute Full Fault:  Replacing %x with %x at height %d leader height %d",
		fullFault.ServerID.Bytes()[2:6],
		fullFault.AuditServerID.Bytes()[2:6],
		fullFault.DBHeight,
		s.LLeaderHeight))

	nextIndex := (int(fullFault.VMIndex) + 1) % len(pl.FedServers)
	if pl.VMs[nextIndex].FaultFlag > 0 {
		markNoFault(pl, nextIndex)
	}

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
	s.AddStatus(fmt.Sprintf("RESET Trying to Reset for the %d time", s.ResetTryCnt))
	index := len(s.DBStates.DBStates) - 1
	if index < 2 {
		s.AddStatus("RESET Failed to Reset because not enough dbstates")
		return
	}

	dbs := s.DBStates.DBStates[index]
	for {
		if dbs == nil {
			s.AddStatus("RESET Reset Failed")
			return
		}
		if dbs.Saved {
			break
		}
		index--
		dbs = s.DBStates.DBStates[index]
	}
	if index > 1 {
		s.ResetCnt++
		dbs = s.DBStates.DBStates[index-1]
		s.DBStates.DBStates = s.DBStates.DBStates[:index]

		dbs.AdminBlock = dbs.AdminBlock.New().(interfaces.IAdminBlock)
		dbs.FactoidBlock = dbs.FactoidBlock.New().(interfaces.IFBlock)

		plToReset := s.ProcessLists.Get(s.DBStates.Base + uint32(index) + 1)
		plToReset.Reset()

		//s.StartDelay = s.GetTimestamp().GetTimeMilli() // We cant start as a leader until we know we are upto date
		//s.RunLeader = false
		s.CurrentMinute = 0

		s.SetLeaderTimestamp(dbs.NextTimestamp)

		s.DBStates.ProcessBlocks(dbs)
	} else {
		s.AddStatus("RESET Can't reset far enough back")
	}
}
