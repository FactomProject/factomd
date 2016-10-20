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
	FaultCore     FaultCore
	AmINegotiator bool
	MyVoteTallied bool
	VoteMap       map[[32]byte]interfaces.IFullSignature
	PledgeDone    bool
	LastMatch     int64
}

func (fs *FaultState) IsNil() bool {
	if fs.VoteMap == nil || fs.FaultCore.ServerID.IsZero() || fs.FaultCore.AuditServerID.IsZero() {
		return true
	}
	return false
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

func eomFault(pl *ProcessList, vm *VM, vmIndex, height, tag int) {
	now := time.Now().Unix()

	if vm.whenFaulted == 0 {
		// if we did not previously consider this VM faulted
		// we simply mark it as faulted (by assigning it a nonzero whenFaulted time)
		// and keep track of the ProcessList height it has faulted at
		vm.whenFaulted = now
		vm.faultHeight = height
	} else {
		if int(now-vm.whenFaulted) > pl.State.FaultTimeout {
			if int(now-vm.whenFaulted) > pl.State.FaultTimeout*(1+pl.State.EOMfaultIndex) {
				if pl.LenFaultMap() < 1 {
					modFaultIndex := pl.State.EOMfaultIndex % len(pl.FedServers)
					Fault(pl, vm, vmIndex+modFaultIndex, height)
					pl.State.EOMfaultIndex = pl.State.EOMfaultIndex + 1
				}

			}
		}
	}
	if int(now-pl.State.LastFaultAction) > pl.State.FaultWait {
		//THROTTLE
		FaultCheck(pl)
		pl.State.LastFaultAction = now
	}
}

func Fault(pl *ProcessList, vm *VM, vmIndex, height int) {
	now := time.Now().Unix()

	if vm.whenFaulted == 0 {
		// if we did not previously consider this VM faulted
		// we simply mark it as faulted (by assigning it a nonzero whenFaulted time)
		// and keep track of the ProcessList height it has faulted at
		vm.whenFaulted = now
		vm.faultHeight = height
	}

	fedServerCnt := len(pl.FedServers)
	responsibleFaulterIdx := (vmIndex + 1) % fedServerCnt

	if pl.State.Leader && pl.State.LeaderVMIndex == responsibleFaulterIdx {
		CraftAndSubmitFault(pl, vm, vmIndex, height)
	}
	pl.FedServers[pl.ServerMap[pl.State.CurrentMinute][vmIndex]].SetOnline(false)

}

func TopPriorityFaultState(pl *ProcessList) FaultState {
	if pl.LenFaultMap() < 1 {
		return *new(FaultState)
	}
	var currentMax int64
	currentMax = 0
	var winner [32]byte
	for _, faultID := range pl.GetKeysFaultMap() {
		fs := pl.GetFaultState(faultID)
		thisPriority := fs.FaultCore.Timestamp.GetTimeSeconds()
		if thisPriority > currentMax {
			currentMax = thisPriority
			winner = faultID
		}
	}
	return pl.GetFaultState(winner)
}

func FaultCheck(pl *ProcessList) {
	faultState := TopPriorityFaultState(pl)
	if !faultState.IsNil() {
		if isMyNegotiation(faultState.FaultCore, pl) {
			/*faultState.AmINegotiator = true
			pl.AmINegotiator = true
			pl.AddFaultState(faultState.FaultCore.GetHash().Fixed(), faultState)*/
			CraftAndSubmitFullFault(pl, faultState.FaultCore.GetHash().Fixed())
		}
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

	if pl.LenFaultMap() < 1 {
		return false
	}

	faultIDs := pl.GetKeysFaultMap()
	for _, faultID := range faultIDs {
		faultState := pl.GetFaultState(faultID)
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
	}

	return false
}

// CraftAndSubmitFault is how we issue ServerFault messages when we have an
// open choice of which Audit Server we want to vote for (i.e. when we're not
// just matching someone else's existing Fault message vote)
func CraftAndSubmitFault(pl *ProcessList, vm *VM, vmIndex int, height int) {
	// TODO: if I am the Leader being faulted, I should respond by sending out
	// a MissingMsgResponse to everyone for the msg I'm being faulted for

	// Only consider Online Audit servers as candidates for promotion (this
	// allows us to cycle through Audits on successive calls to CraftAndSubmitFault,
	// so that we make sure to (eventually) find one that is ready and able to
	// accept the promotion)
	auditServerList := pl.State.GetOnlineAuditServers(pl.DBHeight)
	if len(auditServerList) > 0 {
		// Nominate the top candidate from the list of Online Audit Servers
		replacementServer := auditServerList[0]
		leaderMin := pl.State.CurrentMinute

		faultedFed := pl.ServerMap[leaderMin][vmIndex]
		pl.FedServers[faultedFed].SetOnline(false)
		faultedFedID := pl.FedServers[faultedFed].GetChainID()

		// Create and send ServerFault (vote) message
		sf := messages.NewServerFault(faultedFedID, replacementServer.GetChainID(), vmIndex, pl.DBHeight, uint32(height), pl.System.Height, pl.State.GetTimestamp())
		if sf != nil {
			sf.Sign(pl.State.serverPrivKey)
			//pl.State.NetworkOutMsgQueue() <- sf
			pl.State.InMsgQueue() <- sf
			fm := pl.GetFaultState(sf.GetCoreHash().Fixed())
			if !fm.IsNil() {
				// If we already have a FaultState saved to our ProcessList's
				// FaultMap, we update its "LastMatch" value to the current time
				// (LastMatch is merely a throttling mechanism)
				fm.LastMatch = time.Now().Unix()
				pl.AddFaultState(sf.GetCoreHash().Fixed(), fm)
			}
			// Now let's set the Audit Server offline (so we don't just re-nominate them over and over)
			replacementServer.SetOnline(false)
		}
	} else {
		// If we don't see any Audit servers as Online, we reset all of
		// them to an Online state and start the cycle anew
		for _, aud := range pl.AuditServers {
			aud.SetOnline(true)
		}
	}
}

// CraftAndSubmitFullFault is called from the Negotiate goroutine
// (which fires once every 5 seconds on each server); most of the time
// these are "incomplete" FullFault messages which serve as status pings
// for the negotiation in progress
func CraftAndSubmitFullFault(pl *ProcessList, faultID [32]byte) *messages.FullServerFault {
	faultState := pl.GetFaultState(faultID)
	fc := faultState.FaultCore

	sf := messages.NewServerFault(fc.ServerID, fc.AuditServerID, int(fc.VMIndex), fc.DBHeight, fc.Height, pl.System.Height, fc.Timestamp)

	var listOfSigs []interfaces.IFullSignature
	for _, sig := range faultState.VoteMap {
		listOfSigs = append(listOfSigs, sig)
	}

	fullFault := messages.NewFullServerFault(sf, listOfSigs, pl.System.Height)
	//adminBlockEntryForFault := fullFault.ToAdminBlockEntry()
	//pl.State.LeaderPL.AdminBlock.AddServerFault(adminBlockEntryForFault)
	if fullFault != nil {
		fullFault.Sign(pl.State.serverPrivKey)
		fullFault.SendOut(pl.State, fullFault)
		fullFault.FollowerExecute(pl.State)
	}

	return fullFault
}

func (s *State) FollowerExecuteSFault(m interfaces.IMsg) {
	sf, _ := m.(*messages.ServerFault)
	pl := s.ProcessLists.Get(sf.DBHeight)

	if pl == nil || pl.VMs[sf.VMIndex].whenFaulted == 0 {
		// If no such ProcessList exists, or if we don't consider
		// the VM in this ServerFault message to be at fault,
		// do not proceed with regularFaultExecution
		return
	}

	s.regularFaultExecution(sf, pl)
}

// regularFaultExecution will create a FaultState in our FaultMap (if none exists already)
// and save the signature of this particular ServerFault to that FaultStat's VoteMap
func (s *State) regularFaultExecution(sf *messages.ServerFault, pl *ProcessList) {
	var issuerID [32]byte
	rawIssuerID := sf.GetSignature().GetKey()
	for i := 0; i < 32; i++ {
		if i < len(rawIssuerID) {
			issuerID[i] = rawIssuerID[i]
		}
	}

	coreHash := sf.GetCoreHash().Fixed()
	faultState := pl.GetFaultState(coreHash)
	if faultState.IsNil() {
		// We don't have a map entry yet; let's create one
		fcore := ExtractFaultCore(sf)
		faultState = FaultState{FaultCore: fcore, AmINegotiator: false, MyVoteTallied: false, VoteMap: make(map[[32]byte]interfaces.IFullSignature)}

		if isMyNegotiation(fcore, pl) {
			faultState.AmINegotiator = true
			pl.AmINegotiator = true
		}

		if faultState.VoteMap == nil {
			faultState.VoteMap = make(map[[32]byte]interfaces.IFullSignature)
		}
		pl.AddFaultState(coreHash, faultState)
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
	}

	if s.Leader || s.IdentityChainID.IsSameAs(sf.AuditServerID) {
		if !faultState.MyVoteTallied {
			if time.Now().Unix()-faultState.LastMatch > 3 {
				s.matchFault(sf)
				faultState.LastMatch = time.Now().Unix()
			}
		}
	}

	// Update the FaultState in our FaultMap with any updates that were applied
	// during execution of this function
	pl.AddFaultState(coreHash, faultState)
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

// regularFullFaultExecution does the same thing as regularFaultExecution, except
// it will make sure to add every signature from the FullFault to the corresponding
// FaultState's VoteMap
func (s *State) regularFullFaultExecution(sf *messages.FullServerFault, pl *ProcessList) {
	coreHash := sf.GetCoreHash().Fixed()

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
			fcore := ExtractFaultCore(sf)
			faultState = FaultState{FaultCore: fcore, AmINegotiator: false, MyVoteTallied: false, VoteMap: make(map[[32]byte]interfaces.IFullSignature)}

			if isMyNegotiation(fcore, pl) {
				faultState.AmINegotiator = true
				pl.AmINegotiator = true
			}

			if faultState.VoteMap == nil {
				faultState.VoteMap = make(map[[32]byte]interfaces.IFullSignature)
			}
			pl.AddFaultState(coreHash, faultState)
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
				nsf := messages.NewServerFault(sf.ServerID, sf.AuditServerID, int(sf.VMIndex), sf.DBHeight, sf.Height, int(sf.SystemHeight), sf.Timestamp)
				s.matchFault(nsf)
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

// Unfault is used to reset to the default state (No One At Fault)
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
		pl.FedServers[i].SetOnline(true)
	}
	pl.AmINegotiator = false
	pl.ChosenNegotiation = [32]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	pl.State.EOMfaultIndex = 0
}

func (pl *ProcessList) ClearFaultMap() {
	pl.FaultMapMutex.Lock()
	defer pl.FaultMapMutex.Unlock()
	pl.FaultMap = make(map[[32]byte]FaultState)
}

// When we execute a FullFault message, it could be complete (includes all
// necessary signatures + pledge) or incomplete, in which case it is just
// a negotiation ping
func (s *State) FollowerExecuteFullFault(m interfaces.IMsg) {
	fullFault, _ := m.(*messages.FullServerFault)

	pl := s.ProcessLists.Get(fullFault.DBHeight)

	if fullFault.ServerID.IsZero() || fullFault.AuditServerID.IsZero() {
		return
	}

	auditServerList := s.GetAuditServers(fullFault.DBHeight)
	var theAuditReplacement interfaces.IFctServer

	for _, auditServer := range auditServerList {
		if auditServer.GetChainID().IsSameAs(fullFault.AuditServerID) {
			theAuditReplacement = auditServer
		}
	}
	if theAuditReplacement == nil {
		// If we don't have any Audit Servers in our Authority set
		// that match the nominated Audit Server in the FullFault,
		// we can't really do anything useful with it
		return
	}

	vm := pl.VMs[int(fullFault.VMIndex)]
	rHt := vm.Height
	ffHt := int(fullFault.Height)
	if false && rHt > ffHt {
		fmt.Printf("dddd  %20s VM[%d] height %d Full Fault ht: %d \n", s.FactomNodeName, fullFault.VMIndex, rHt, ffHt)
		vm.Height = ffHt
		vm.List = vm.List[:ffHt] // Nuke all the extra messages that might annoy us.
	}

	if int(fullFault.SystemHeight) == pl.System.Height {
		if fullFault.HasEnoughSigs(s) && s.pledgedByAudit(fullFault) {
			// COMPLETE
			pl.AddToSystemList(fullFault)
		} else {
			mightMatch := false
			willUpdate := false
			if s.IdentityChainID.IsSameAs(fullFault.AuditServerID) {
				// We are the nominated auditServer
				mightMatch = true
			} else {
				if vm.whenFaulted != 0 {
					//I AGREE
					currentTopPriority := TopPriorityFaultState(pl)
					var tpts int64
					if currentTopPriority.IsNil() {
						tpts = 0
					} else {
						tpts = currentTopPriority.FaultCore.Timestamp.GetTimeSeconds()
					}
					ffts := fullFault.Timestamp.GetTimeSeconds()
					if ffts >= tpts {
						//THIS IS TOP PRIORITY
						if !currentTopPriority.IsNil() && fullFault.ServerID.IsSameAs(currentTopPriority.FaultCore.ServerID) && ffts > tpts {
							//IT IS A RENEWAL
							if int(ffts-tpts) < s.FaultTimeout {
								//TOO SOON
								newVMI := (int(fullFault.VMIndex) + 1) % len(pl.FedServers)
								newVM := pl.VMs[newVMI]
								Fault(pl, newVM, newVMI, int(fullFault.Height))
							} else {
								if !currentTopPriority.IsNil() && couldIFullFault(pl, int(currentTopPriority.FaultCore.VMIndex)) {
									//I COULD FAULT BUT HE HASN'T
									newVMI := (int(fullFault.VMIndex) + 1) % len(pl.FedServers)
									newVM := pl.VMs[newVMI]
									Fault(pl, newVM, newVMI, int(fullFault.Height))
								} else {
									willUpdate = true
								}
							}
						} else {
							willUpdate = true
						}
					}
				}
			}
			if willUpdate {
				mightMatch = true
				s.regularFullFaultExecution(fullFault, pl)
			}
			if mightMatch {
				theFaultState := pl.GetFaultState(fullFault.GetCoreHash().Fixed())
				if !theFaultState.MyVoteTallied {
					now := time.Now().Unix()

					if now-theFaultState.LastMatch > 5 {
						nsf := messages.NewServerFault(fullFault.ServerID, fullFault.AuditServerID, int(fullFault.VMIndex),
							fullFault.DBHeight, fullFault.Height, int(fullFault.SystemHeight), fullFault.Timestamp)
						s.matchFault(nsf)
					}
				}
			}
		}
	}
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
