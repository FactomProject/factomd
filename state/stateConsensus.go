// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"fmt"
	"hash"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/util"
)

var _ = fmt.Print
var _ = (*hash.Hash32)(nil)

//***************************************************************
// Process Loop for Consensus
//
// Returns true if some message was processed.
//***************************************************************

func (s *State) Process() (progress bool) {

	if !s.RunLeader {
		now := s.GetTimestamp().GetTimeMilli() // Timestamps are in milliseconds, so wait 20
		if now-s.StartDelay > s.StartDelayLimit {
			s.RunLeader = true
		}
		s.LeaderPL = s.ProcessLists.Get(s.LLeaderHeight)
		s.Leader, s.LeaderVMIndex = s.LeaderPL.GetVirtualServers(s.CurrentMinute, s.IdentityChainID)
	}

	var vm *VM
	if s.Leader {
		vm = s.LeaderPL.VMs[s.LeaderVMIndex]
	}

	// Executing a message means looking if it is valid, checking if we are a leader.
	executeMsg := func(msg interfaces.IMsg) (ret bool) {
		_, ok := s.Replay.Valid(constants.INTERNAL_REPLAY, msg.GetRepeatHash().Fixed(), msg.GetTimestamp(), s.GetTimestamp())
		if !ok {
			return
		}
		s.SetString()
		msg.ComputeVMIndex(s)

		switch msg.Validate(s) {
		case 1:
			if s.RunLeader &&
				s.Leader &&
				!s.Saving &&
				int(vm.Height) == len(vm.List) &&
				(!s.Syncing || !vm.Synced) &&
				(msg.IsLocal() || msg.GetVMIndex() == s.LeaderVMIndex) {
				msg.LeaderExecute(s)
			} else {
				msg.FollowerExecute(s)
			}
			ret = true
		case 0:
			s.Holding[msg.GetMsgHash().Fixed()] = msg
		default:
			s.Holding[msg.GetMsgHash().Fixed()] = msg
			s.networkInvalidMsgQueue <- msg
		}

		return
	}

	s.ReviewHolding()

	// Reprocess any stalled Acknowledgements
	for i := 0; i < 10 && len(s.XReview) > 0; i++ {
		msg := s.XReview[0]
		executeMsg(msg)
		s.XReview = s.XReview[1:]
	}

	select {
	case ack := <-s.ackQueue:
		a := ack.(*messages.Ack)
		if a.DBHeight >= s.LLeaderHeight && ack.Validate(s) == 1 {
			ack.FollowerExecute(s)
		}
		progress = true
	case msg := <-s.msgQueue:
		if executeMsg(msg) {
			s.networkOutMsgQueue <- msg
		}
	default:
	}
	return
}

//***************************************************************
// Consensus Methods
//***************************************************************

// Places the entries in the holding map back into the XReview list for
// review if this is a leader, and those messages are that leader's
// responsibility
func (s *State) ReviewHolding() {
	if len(s.XReview) > 0 {
		return
	}
	// Anything we are holding, we need to reprocess.
	s.XReview = make([]interfaces.IMsg, 0)

	for k := range s.Holding {
		v := s.Holding[k]

		_, ok := s.Replay.Valid(constants.INTERNAL_REPLAY, v.GetRepeatHash().Fixed(), v.GetTimestamp(), s.GetTimestamp())
		if !ok {
			delete(s.Holding, k)
			continue
		}

		if v.Expire(s) {
			s.ExpireCnt++
			delete(s.Holding, k)
			continue
		}

		if v.Resend(s) {
			_, ok := s.Replay.Valid(constants.INTERNAL_REPLAY, v.GetRepeatHash().Fixed(), v.GetTimestamp(), s.GetTimestamp())
			if !ok {
				delete(s.Holding, k)
				continue
			}
		}

		if s.Leader && v.GetVMIndex() == s.LeaderVMIndex {
			s.XReview = append(s.XReview, v)
			delete(s.Holding, k)
		}

	}
}

// Adds blocks that are either pulled locally from a database, or acquired from peers.
func (s *State) AddDBState(isNew bool,
	directoryBlock interfaces.IDirectoryBlock,
	adminBlock interfaces.IAdminBlock,
	factoidBlock interfaces.IFBlock,
	entryCreditBlock interfaces.IEntryCreditBlock) *DBState {

	dbState := s.DBStates.NewDBState(isNew, directoryBlock, adminBlock, factoidBlock, entryCreditBlock)

	ht := dbState.DirectoryBlock.GetHeader().GetDBHeight()
	if ht > s.LLeaderHeight {
		s.Syncing = false
		s.EOM = false
		s.DBSig = false
		s.LLeaderHeight = ht
		s.ProcessLists.Get(ht + 1)
		s.CurrentMinute = 0
		s.EOMProcessed = 0
		s.DBSigProcessed = 0
		s.StartDelay = s.GetTimestamp().GetTimeMilli()
		s.RunLeader = false
		s.Newblk = true
		s.LeaderPL = s.ProcessLists.Get(s.LLeaderHeight)
		s.Leader, s.LeaderVMIndex = s.LeaderPL.GetVirtualServers(s.CurrentMinute, s.IdentityChainID)
	}
	if ht == 0 && s.LLeaderHeight < 1 {
		s.LLeaderHeight = 1
	}

	return dbState
}

func (s *State) addEBlock(eblock interfaces.IEntryBlock) {
	hash, err := eblock.KeyMR()

	if err == nil {
		if s.HasDataRequest(hash) {

			s.DB.ProcessEBlockBatch(eblock, true)
			delete(s.DataRequests, hash.Fixed())

			if s.GetAllEntries(hash) {
				if s.GetEBDBHeightComplete() < eblock.GetDatabaseHeight() {
					s.SetEBDBHeightComplete(eblock.GetDatabaseHeight())
				}
			}
		}
	}
}

// Messages that will go into the Process List must match an Acknowledgement.
// The code for this is the same for all such messages, so we put it here.
//
// Returns true if it finds a match, puts the message in holding, or invalidates the message
func (s *State) FollowerExecuteMsg(m interfaces.IMsg) {

	s.Holding[m.GetMsgHash().Fixed()] = m
	ack, _ := s.Acks[m.GetMsgHash().Fixed()].(*messages.Ack)
	if ack != nil {
		m.SetLeaderChainID(ack.GetLeaderChainID())
		m.SetMinute(ack.Minute)

		pl := s.ProcessLists.Get(ack.DBHeight)
		pl.AddToProcessList(ack, m)
	}
}

// Messages that will go into the Process List must match an Acknowledgement.
// The code for this is the same for all such messages, so we put it here.
//
// Returns true if it finds a match, puts the message in holding, or invalidates the message
func (s *State) FollowerExecuteEOM(m interfaces.IMsg) {

	if m.IsLocal() {
		return // This is an internal EOM message.  We are not a leader so ignore.
	}

	s.Holding[m.GetMsgHash().Fixed()] = m

	ack, _ := s.Acks[m.GetMsgHash().Fixed()].(*messages.Ack)
	if ack != nil {
		pl := s.ProcessLists.Get(ack.DBHeight)
		pl.AddToProcessList(ack, m)
	}
}

// Ack messages always match some message in the Process List.   That is
// done here, though the only msg that should call this routine is the Ack
// message.
func (s *State) FollowerExecuteAck(msg interfaces.IMsg) {
	ack := msg.(*messages.Ack)
	s.Acks[ack.GetHash().Fixed()] = ack
	m, _ := s.Holding[ack.GetHash().Fixed()]
	if m != nil {
		m.FollowerExecute(s)
	}
}

func (s *State) FollowerExecuteDBState(msg interfaces.IMsg) {
	dbstatemsg, _ := msg.(*messages.DBStateMsg)

	s.DBStates.LastTime = s.GetTimestamp()
	dbstate := s.AddDBState(false, // Not a new block; got it from the network
		dbstatemsg.DirectoryBlock,
		dbstatemsg.AdminBlock,
		dbstatemsg.FactoidBlock,
		dbstatemsg.EntryCreditBlock)
	dbstate.ReadyToSave = true

	s.DBStateReplyCnt++
}

func (s *State) FollowerExecuteAddData(msg interfaces.IMsg) {
	dataResponseMsg, ok := msg.(*messages.DataResponse)
	if !ok {
		return
	}

	switch dataResponseMsg.DataType {
	case 0: // DataType = entry
		entry := dataResponseMsg.DataObject.(interfaces.IEBEntry)

		if entry.GetHash().IsSameAs(dataResponseMsg.DataHash) {

			s.DB.InsertEntry(entry)
			delete(s.DataRequests, entry.GetHash().Fixed())
		}
	case 1: // DataType = eblock
		eblock := dataResponseMsg.DataObject.(interfaces.IEntryBlock)
		dataHash, _ := eblock.KeyMR()
		if dataHash.IsSameAs(dataResponseMsg.DataHash) {
			s.addEBlock(eblock)
		}
	default:
		s.networkInvalidMsgQueue <- msg
	}

}

func (s *State) FollowerExecuteSFault(m interfaces.IMsg) {
	sf, _ := m.(*messages.ServerFault)
	pl := s.ProcessLists.Get(sf.DBHeight)

	if pl == nil {
		return
	}

	if !pl.VMs[sf.VMIndex].isFaulting {
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
	for _, a := range s.Authorities {
		if a.AuthorityChainID.IsSameAs(sf.AuditServerID) {
			marshalledSF, err := sf.MarshalForSignature()
			if err == nil {
				sigVer, err := a.VerifySignature(marshalledSF, sf.Signature.GetSignature())
				if err == nil && sigVer {
					foundAudit, audIdx := pl.GetAuditServerIndexHash(sf.AuditServerID)
					if foundAudit {
						if pl.AuditServers[audIdx].LeaderToReplace() != nil {
							if !pl.AuditServers[audIdx].LeaderToReplace().IsSameAs(sf.ServerID) {
								// illegal vote; audit server has already AOK'd replacing a different leader
								// "punish" them by setting them offline (i.e. make them ineligible for promotion)
								fmt.Println("JUSTIN NODE", s.FactomNodeName, "SETTING OFFLINE:", pl.AuditServers[audIdx].GetChainID().String()[:10])
								pl.AuditServers[audIdx].SetOnline(false)
							}
						} else {
							// AOK: set the Audit Server's "Leader to Replace" field to this ServerID
							//fmt.Println("JUSTIN NODE", s.FactomNodeName, "SET REPLACE (", sf.AuditServerID.String()[:10], ") PLEDGED TO REPLACE", sf.ServerID.String()[:10], "AT DBH:", sf.DBHeight)
							pl.AuditServers[audIdx].SetReplace(sf.ServerID)
						}
					}
				}
			}
		}
	}

	coreHash := sf.GetCoreHash().Fixed()

	if s.FaultMap[coreHash] == nil {
		s.FaultMap[coreHash] = make(map[[32]byte]interfaces.IFullSignature)
	}

	s.FaultMap[coreHash][issuerID] = sf.GetSignature()

	//faultedServerID := sf.ServerID.Fixed()

	cnt := len(s.FaultMap[coreHash])
	var fedServerCnt int
	if pl != nil {
		fedServerCnt = len(pl.FedServers)
	} else {
		fedServerCnt = len(s.GetFedServers(sf.DBHeight))
	}
	if cnt > ((fedServerCnt / 2) - 1) {
		if s.Leader {
			responsibleFaulterIdx := (int(sf.VMIndex) + 1) % fedServerCnt
			if s.LeaderVMIndex == responsibleFaulterIdx {
				foundAudit, aidx := pl.GetAuditServerIndexHash(sf.AuditServerID)
				if foundAudit {
					serverToReplace := pl.AuditServers[aidx].LeaderToReplace()
					if serverToReplace != nil {
						if serverToReplace.IsSameAs(sf.ServerID) {
							// if we have made it this far, that means we have successfully received an "AOK" message
							// from the audit server being promoted... this means we can proceed and issue a FullFault
							var listOfSigs []interfaces.IFullSignature
							for _, sig := range s.FaultMap[coreHash] {
								listOfSigs = append(listOfSigs, sig)
							}
							// Add our own signature to the list before sending out fullFault, to ensure a majority
							sf.Sign(s.serverPrivKey)
							listOfSigs = append(listOfSigs, sf.Signature)

							fullFault := messages.NewFullServerFault(sf, listOfSigs)
							if fullFault != nil {
								//fmt.Println("JUSTIN YEFF:", s.FactomNodeName, sf.ServerID.String()[:10], sf.AuditServerID.String()[:10])
								fullFault.Sign(s.serverPrivKey)
								s.NetworkOutMsgQueue() <- fullFault
								fullFault.FollowerExecute(s)
								delete(s.FaultMap, sf.GetCoreHash().Fixed())
							}
						}
					} else {
						waitingSince, alreadyWaiting := pl.WaitingForPledge[sf.AuditServerID.String()]
						if !alreadyWaiting {
							pl.WaitingForPledge[sf.AuditServerID.String()] = s.GetTimestamp().GetTimeMilli()
						} else {
							if s.GetTimestamp().GetTimeMilli()-waitingSince > 15 {
								// we will grant the audit server 15 seconds to send out its pledge
								// after that, we try nominating a different one
								auditServerList := s.GetOnlineAuditServers(sf.DBHeight)
								for _, audServ := range auditServerList {
									if audServ.GetChainID().IsSameAs(sf.AuditServerID) {
										continue
									}
									_, waitingForThisOneToo := pl.WaitingForPledge[audServ.GetChainID().String()]
									if waitingForThisOneToo {
										continue
									}

									//NOMINATE
									_, servEntryFound := pl.AlreadyNominated[sf.ServerID.String()]
									if !servEntryFound {
										pl.AlreadyNominated[sf.ServerID.String()] = make(map[string]int64)
									}
									pl.AlreadyNominated[sf.ServerID.String()][audServ.GetChainID().String()] = s.GetTimestamp().GetTimeMilli()
									sf := messages.NewServerFault(s.GetTimestamp(), sf.ServerID, audServ.GetChainID(), int(sf.VMIndex), sf.DBHeight, sf.Height)
									if sf != nil {
										sf.Sign(s.serverPrivKey)
										s.NetworkOutMsgQueue() <- sf
										s.InMsgQueue() <- sf
									}
									break
								}
							}
						}
					}
				}
				//fmt.Println("JUSTIN NOFF:", s.FactomNodeName, cnt, "/", fedServerCnt, sf.ServerID.String()[:10], sf.AuditServerID.String()[:10])
			}
		} else {
			if s.IdentityChainID.IsSameAs(sf.AuditServerID) {
				// I am the audit server being promoted
				if !pl.AmIPledged {
					pl.AmIPledged = true
					//fmt.Println("JUSTIN AUDIT SERVER ", s.IdentityChainID.String()[:10], "PLEDGING TO REPLACE", sf.ServerID.String()[:10], "AT DBH:", sf.DBHeight)
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
	/*
		if pl != nil {
			pl.FaultList[sf.ServerID.Fixed()] = append(pl.FaultList[sf.ServerID.Fixed()], sf.GetSignature().GetKey())
			cnt := len(pl.FaultList[sf.ServerID.Fixed()])
			if s.Leader && cnt > len(pl.FedServers)/2 {
				fmt.Println(s.FactomNodeName, "FAULTING", sf.ServerID.String())
			}
		}*/

	//Match a nomination if we haven't nominated the same server already
	if s.Leader {
		existingNominations, exists := pl.AlreadyNominated[sf.ServerID.String()]
		if exists {
			/*alreadyNominatedThisID := false
			for _, nom := range existingNominations {
				if nom.IsSameAs(sf.AuditServerID) {
					alreadyNominatedThisID = true
				}
			}*/
			_, alreadyNom := existingNominations[sf.AuditServerID.String()]
			if !alreadyNom {
				//fmt.Println("JUSTIN ", s.FactomNodeName, "MATCHING FAULT OF:", sf.ServerID.String()[:10], "NEWNOM:", sf.AuditServerID.String()[:10])
				pl.AlreadyNominated[sf.ServerID.String()][sf.AuditServerID.String()] = s.GetTimestamp().GetTimeMilli()
				matchNomination := messages.NewServerFault(s.GetTimestamp(), sf.ServerID, sf.AuditServerID, int(sf.VMIndex), sf.DBHeight, sf.Height)
				if matchNomination != nil {
					matchNomination.Sign(s.serverPrivKey)
					s.NetworkOutMsgQueue() <- matchNomination
					s.InMsgQueue() <- matchNomination
				}
			} /*else {
				fmt.Println("JUSTIN ", s.FactomNodeName, "NOT MATCHING BECAUSE ALREADY EXISTS NOM FOR:", sf.ServerID.String()[:10], "REPL:", sf.AuditServerID.String()[:10])
			}*/
		} /* else {
			fmt.Println("JUSTIN ", s.FactomNodeName, "NOT MATCHING BECAUSE NO EXISTS NOM FOR:", sf.ServerID.String()[:10], "REPL:", sf.AuditServerID.String()[:10])
		}*/
	}
}

func (s *State) FollowerExecuteFullFault(m interfaces.IMsg) {
	fsf, _ := m.(*messages.FullServerFault)
	relevantPL := s.ProcessLists.Get(fsf.DBHeight)
	auditServerList := s.GetOnlineAuditServers(fsf.DBHeight)
	var theAuditReplacement interfaces.IFctServer
	for _, as := range auditServerList {
		if as.GetChainID().IsSameAs(fsf.AuditServerID) {
			theAuditReplacement = as
		}
	}
	if theAuditReplacement != nil {
		for listIdx, fedServ := range relevantPL.FedServers {
			if fedServ.GetChainID().IsSameAs(fsf.ServerID) {
				relevantPL.FedServers[listIdx] = theAuditReplacement
				relevantPL.AddAuditServer(fedServ.GetChainID())
				s.RemoveAuditServer(fsf.DBHeight, theAuditReplacement.GetChainID())
			}
		}

		//addMsg := messages.NewAddServerByHashMsg(s, 0, auditServerList[0].GetChainID())
		//s.InMsgQueue() <- addMsg
		//s.NetworkOutMsgQueue() <- addMsg
	}
	//	s.RemoveFedServer(fsf.DBHeight, fsf.ServerID)

	//removeMsg := messages.NewRemoveServerMsg(s, fsf.ServerID, 0)
	//s.InMsgQueue() <- removeMsg
	//s.NetworkOutMsgQueue() <- removeMsg
	//fmt.Println("JUSTIN FEFF4", s.FactomNodeName, fsf.ServerID.String()[:10], fsf.AuditServerID.String()[:10])

	s.Leader, s.LeaderVMIndex = s.LeaderPL.GetVirtualServers(s.CurrentMinute, s.IdentityChainID)
	delete(s.FaultMap, fsf.GetCoreHash().Fixed())
}

func (s *State) FollowerExecuteNegotiation(m interfaces.IMsg) {
	negotiation, _ := m.(*messages.Negotiation)
	pl := s.ProcessLists.Get(negotiation.DBHeight)
	if pl == nil {
		return
	}

	if s.Leader {
		// TODO: if I am the Leader being faulted, I should respond by sending out
		// a MissingMsgResponse to everyone for the msg I'm being faulted for

		vmAtFault := pl.VMs[negotiation.VMIndex]
		if vmAtFault.isFaulting {
			vmAtFault.isNegotiating = true
			auditServerList := s.GetOnlineAuditServers(negotiation.DBHeight)
			if len(auditServerList) > 0 {
				replacementServer := auditServerList[0]
				if s.GetTimestamp().GetTimeMilli()-vmAtFault.whenFaulted > 30 {
					// If it has been more than 30 seconds since initially faulting vmAtFault
					// we want to try cycling to another online audit server, if possible
					//foundNewAudit := false
					for audIdx, audServ := range auditServerList {
						if audIdx == 0 {
							continue
						}
						//if foundNewAudit {
						//	break
						//}

						//Find an audit server we haven't nominated already
						existingNominations, exists := pl.AlreadyNominated[negotiation.ServerID.String()]
						if exists {
							/*alreadyNominatedThisID := false
							for _, nom := range existingNominations {
								if nom.IsSameAs(audServ.GetChainID()) {
									alreadyNominatedThisID = true
								}
							}*/
							_, alreadyNominatedThisID := existingNominations[audServ.GetChainID().String()]
							if !alreadyNominatedThisID {
								replacementServer = audServ
								//foundNewAudit = true
								break
							}
						}
					}
				}
				/*for auditIndex := 0; auditIndex < len(auditServerList); auditIndex++ {
					if replacementServer.LeaderToReplace() != nil && !replacementServer.LeaderToReplace().IsSameAs(negotiation.ServerID) {
						replacementServer = auditServerList[auditIndex]
					} else {
						break
					}
				}*/
				//fmt.Println("JUSTIN ", s.FactomNodeName, "SENDING NEW SFA BASED OFF NEGO F:", negotiation.ServerID.String()[:10], "AUD:", replacementServer.GetChainID().String()[:10])
				//NOMINATE
				_, servEntryFound := pl.AlreadyNominated[negotiation.ServerID.String()]
				if !servEntryFound {
					pl.AlreadyNominated[negotiation.ServerID.String()] = make(map[string]int64)
				}
				pl.AlreadyNominated[negotiation.ServerID.String()][replacementServer.GetChainID().String()] = s.GetTimestamp().GetTimeMilli()
				sf := messages.NewServerFault(s.GetTimestamp(), negotiation.ServerID, replacementServer.GetChainID(), int(negotiation.VMIndex), negotiation.DBHeight, negotiation.Height)
				if sf != nil {
					sf.Sign(s.serverPrivKey)
					s.NetworkOutMsgQueue() <- sf
					s.InMsgQueue() <- sf
				}
			}
		}
	}

}

func (s *State) FollowerExecuteMMR(m interfaces.IMsg) {
	mmr, _ := m.(*messages.MissingMsgResponse)
	ackResp := mmr.AckResponse.(*messages.Ack)
	//s.Holding[mmr.MsgResponse.GetHash().Fixed()] = mmr.MsgResponse
	//s.Acks[ackResp.GetHash().Fixed()] = ackResp

	pl := s.ProcessLists.Get(ackResp.DBHeight)
	pl.AddToProcessList(ackResp, mmr.MsgResponse)
	s.MissingAnsCnt++
}

func (s *State) LeaderExecute(m interfaces.IMsg) {

	_, ok := s.Replay.Valid(constants.INTERNAL_REPLAY, m.GetRepeatHash().Fixed(), m.GetTimestamp(), s.GetTimestamp())
	if !ok {
		delete(s.Holding, m.GetMsgHash().Fixed())
		return
	}

	ack := s.NewAck(m).(*messages.Ack)
	m.SetLeaderChainID(ack.GetLeaderChainID())
	m.SetMinute(ack.Minute)
	s.ProcessLists.Get(ack.DBHeight).AddToProcessList(ack, m)

}

func (s *State) LeaderExecuteEOM(m interfaces.IMsg) {

	if !m.IsLocal() {
		s.FollowerExecuteEOM(m)
		return
	}

	// The zero based minute for the message is equal to
	// the one based "LastMinute".  This way we know we are
	// generating minutes in order.

	eom := m.(*messages.EOM)
	pl := s.ProcessLists.Get(s.LLeaderHeight)
	vm := pl.VMs[s.LeaderVMIndex]
	if s.Syncing && vm.Synced {
		return
	} else if !s.Syncing {
		s.Syncing = true
		s.EOM = true
		s.EOMsyncing = true
		s.EOMProcessed = 0
		for _, vm := range pl.VMs {
			vm.Synced = false
		}
		s.EOMLimit = len(pl.FedServers)
		s.EOMMinute = int(s.CurrentMinute)
	}

	//_, vmindex := pl.GetVirtualServers(s.EOMMinute, s.IdentityChainID)

	eom.DBHeight = s.LLeaderHeight
	eom.VMIndex = s.LeaderVMIndex
	// eom.Minute is zerobased, while LeaderMinute is 1 based.  So
	// a simple assignment works.
	eom.Minute = byte(s.CurrentMinute)
	eom.Sign(s)
	ack := s.NewAck(m)
	s.Acks[eom.GetMsgHash().Fixed()] = ack
	m.SetLocal(false)
	s.FollowerExecuteEOM(m)
	s.UpdateState()
}

func (s *State) LeaderExecuteRevealEntry(m interfaces.IMsg) {
	re := m.(*messages.RevealEntryMsg)
	commit := s.NextCommit(re.Entry.GetHash())
	if commit == nil {
		m.FollowerExecute(s)
	}
	s.PutCommit(re.Entry.GetHash(), commit)
	s.LeaderExecute(m)
}

func (s *State) ProcessAddServer(dbheight uint32, addServerMsg interfaces.IMsg) bool {
	as, ok := addServerMsg.(*messages.AddServerMsg)
	if !ok {
		return true
	}

	if leader, _ := s.LeaderPL.GetFedServerIndexHash(as.ServerChainID); leader && as.ServerType == 0 {
		return true
	}

	if !ProcessIdentityToAdminBlock(s, as.ServerChainID, as.ServerType) {
		fmt.Printf("dddd %s %s\n", s.FactomNodeName, "Addserver message did not add to admin block.")
		return true
	}
	return true
}

func (s *State) ProcessRemoveServer(dbheight uint32, removeServerMsg interfaces.IMsg) bool {
	rs, ok := removeServerMsg.(*messages.RemoveServerMsg)
	if !ok {
		return true
	}

	if !s.VerifyIsAuthority(rs.ServerChainID) {
		fmt.Printf("dddd %s %s\n", s.FactomNodeName, "RemoveServer message did not add to admin block. Not an Authority")
		return true
	}

	if s.GetAuthorityServerType(rs.ServerChainID) != rs.ServerType {
		fmt.Printf("dddd %s %s\n", s.FactomNodeName, "RemoveServer message did not add to admin block. Servertype of message did not match authority's")
		return true
	}

	if len(s.LeaderPL.FedServers) < 2 && rs.ServerType == 0 {
		fmt.Printf("dddd %s %s\n", s.FactomNodeName, "RemoveServer message did not add to admin block. Only 1 federated server exists.")
		return true
	}
	s.LeaderPL.AdminBlock.RemoveFederatedServer(rs.ServerChainID)

	return true
}

func (s *State) ProcessChangeServerKey(dbheight uint32, changeServerKeyMsg interfaces.IMsg) bool {
	ask, ok := changeServerKeyMsg.(*messages.ChangeServerKeyMsg)
	if !ok {
		return true
	}

	if !s.VerifyIsAuthority(ask.IdentityChainID) {
		fmt.Printf("dddd %s %s\n", s.FactomNodeName, "ChangeServerKey message did not add to admin block.")
		return true
	}

	switch ask.AdminBlockChange {
	case constants.TYPE_ADD_BTC_ANCHOR_KEY:
		var btcKey [20]byte
		copy(btcKey[:], ask.Key.Bytes()[:20])
		s.LeaderPL.AdminBlock.AddFederatedServerBitcoinAnchorKey(ask.IdentityChainID, ask.KeyPriority, ask.KeyType, &btcKey)
	case constants.TYPE_ADD_FED_SERVER_KEY:
		pub := ask.Key.Fixed()
		s.LeaderPL.AdminBlock.AddFederatedServerSigningKey(ask.IdentityChainID, &pub)
	case constants.TYPE_ADD_MATRYOSHKA:
		s.LeaderPL.AdminBlock.AddMatryoshkaHash(ask.IdentityChainID, ask.Key)
	}
	return true
}

func (s *State) ProcessCommitChain(dbheight uint32, commitChain interfaces.IMsg) bool {
	c, _ := commitChain.(*messages.CommitChainMsg)

	pl := s.ProcessLists.Get(dbheight)
	pl.EntryCreditBlock.GetBody().AddEntry(c.CommitChain)
	s.GetFactoidState().UpdateECTransaction(true, c.CommitChain)

	// save the Commit to match agains the Reveal later
	s.PutCommit(c.CommitChain.EntryHash, c)

	return true
}

func (s *State) ProcessCommitEntry(dbheight uint32, commitEntry interfaces.IMsg) bool {
	c, _ := commitEntry.(*messages.CommitEntryMsg)

	pl := s.ProcessLists.Get(dbheight)
	pl.EntryCreditBlock.GetBody().AddEntry(c.CommitEntry)
	s.GetFactoidState().UpdateECTransaction(true, c.CommitEntry)

	// save the Commit to match agains the Reveal later
	s.PutCommit(c.CommitEntry.EntryHash, c)

	return true
}

func (s *State) ProcessRevealEntry(dbheight uint32, m interfaces.IMsg) bool {

	msg := m.(*messages.RevealEntryMsg)
	myhash := msg.Entry.GetHash()

	chainID := msg.Entry.GetChainID()

	s.NextCommit(myhash)

	eb := s.GetNewEBlocks(dbheight, chainID)
	eb_db := s.GetNewEBlocks(dbheight-1, chainID)
	if eb_db == nil {
		eb_db, _ = s.DB.FetchEBlockHead(chainID)
	}
	// Handle the case that this is a Entry Chain create
	// Must be built with CommitChain (i.e. !msg.IsEntry).  Also
	// cannot have an existing chaing (eb and eb_db == nil)
	if !msg.IsEntry && eb == nil && eb_db == nil {
		// Create a new Entry Block for a new Entry Block Chain
		eb = entryBlock.NewEBlock()
		// Set the Chain ID
		eb.GetHeader().SetChainID(chainID)
		// Set the Directory Block Height for this Entry Block
		eb.GetHeader().SetDBHeight(dbheight)
		// Add our new entry
		eb.AddEBEntry(msg.Entry)
		// Put it in our list of new Entry Blocks for this Directory Block
		s.PutNewEBlocks(dbheight, chainID, eb)
		s.PutNewEntries(dbheight, myhash, msg.Entry)

		s.IncEntryChains()
		s.IncEntries()
		return true
	}

	// Create an entry (even if they used commitChain).  Means there must
	// be a chain somewhere.  If not, we return false.
	if eb == nil {
		if eb_db == nil {
			return false
		}
		eb = entryBlock.NewEBlock()
		eb.GetHeader().SetEBSequence(eb_db.GetHeader().GetEBSequence() + 1)
		eb.GetHeader().SetPrevFullHash(eb_db.GetHash())
		// Set the Chain ID
		eb.GetHeader().SetChainID(chainID)
		// Set the Directory Block Height for this Entry Block
		eb.GetHeader().SetDBHeight(dbheight)
		// Set the PrevKeyMR
		key, _ := eb_db.KeyMR()
		eb.GetHeader().SetPrevKeyMR(key)
	}
	// Add our new entry
	eb.AddEBEntry(msg.Entry)
	// Put it in our list of new Entry Blocks for this Directory Block
	s.PutNewEBlocks(dbheight, chainID, eb)
	s.PutNewEntries(dbheight, myhash, msg.Entry)

	// Monitor key changes for fed/audit servers
	LoadIdentityByEntry(msg.Entry, s, dbheight, false)

	s.IncEntries()
	return true
}

// TODO: Should fault the server if we don't have the proper sequence of EOM messages.
func (s *State) ProcessEOM(dbheight uint32, msg interfaces.IMsg) bool {

	e := msg.(*messages.EOM)

	if s.Syncing && !s.EOM {
		return false
	}

	if s.EOM && int(e.Minute) > s.EOMMinute {
		return false
	}

	pl := s.ProcessLists.Get(dbheight)
	vm := s.ProcessLists.Get(dbheight).VMs[msg.GetVMIndex()]

	// If I have done everything for all EOMs for all VMs, then and only then do I
	// let processing continue.
	if s.EOMDone {
		s.EOMProcessed--
		if s.EOMProcessed == 0 {
			s.EOM = false
			s.EOMDone = false
			s.ReviewHolding()
			s.Syncing = false
		}
		return true
	}

	// What I do once  for all VMs at the beginning of processing a particular EOM
	if !s.EOM {
		s.Syncing = true
		s.EOM = true
		s.EOMLimit = len(s.LeaderPL.FedServers)
		s.EOMMinute = int(e.Minute)
		s.EOMsyncing = true
		s.EOMProcessed = 0
		s.Newblk = false

		for _, vm := range pl.VMs {
			vm.Synced = false
		}
		return false
	}

	// What I do for each EOM
	if !e.Processed {
		vm.LeaderMinute++
		s.EOMProcessed++
		e.Processed = true
		vm.Synced = true
		return false
	}

	//vm.missingTime = ask(pl, msg.GetVMIndex(), 1, vm, vm.missingTime, vm.Height, 6)

	// After all EOM markers are processed, Claim we are done.  Now we can unwind
	if s.EOMProcessed == s.EOMLimit && !s.EOMDone {

		s.EOMDone = true
		for _, eb := range pl.NewEBlocks {
			eb.AddEndOfMinuteMarker(byte(e.Minute + 1))
		}

		s.FactoidState.EndOfPeriod(int(e.Minute))

		ecblk := pl.EntryCreditBlock
		ecbody := ecblk.GetBody()
		mn := entryCreditBlock.NewMinuteNumber(e.Minute + 1)
		ecbody.AddEntry(mn)

		if !s.Leader {
			s.CurrentMinute = int(e.Minute)
		}

		s.CurrentMinute++

		switch {
		case s.CurrentMinute < 10:
			s.LeaderPL = s.ProcessLists.Get(s.LLeaderHeight)
			s.Leader, s.LeaderVMIndex = s.LeaderPL.GetVirtualServers(s.CurrentMinute, s.IdentityChainID)
		case s.CurrentMinute == 10:
			dbstate := s.AddDBState(true, s.LeaderPL.DirectoryBlock, s.LeaderPL.AdminBlock, s.GetFactoidState().GetCurrentBlock(), s.LeaderPL.EntryCreditBlock)
			dbht := int(dbstate.DirectoryBlock.GetHeader().GetDBHeight())
			if dbht > 0 {
				prev := s.DBStates.Get(dbht - 1)
				s.DBStates.FixupLinks(prev, dbstate)
			}
			s.DBStates.ProcessBlocks(dbstate)

			s.CurrentMinute = 0
			s.LLeaderHeight++
			s.LeaderPL = s.ProcessLists.Get(s.LLeaderHeight)
			s.Leader, s.LeaderVMIndex = s.LeaderPL.GetVirtualServers(0, s.IdentityChainID)

			s.DBSigProcessed = 0
			if s.Leader {
				dbstate := s.DBStates.Get(int(s.LLeaderHeight - 1))
				dbs := new(messages.DirectoryBlockSignature)
				dbs.DirectoryBlockHeader = dbstate.DirectoryBlock.GetHeader()
				//dbs.DirectoryBlockKeyMR = dbstate.DirectoryBlock.GetKeyMR()
				dbs.ServerIdentityChainID = s.GetIdentityChainID()
				dbs.DBHeight = s.LLeaderHeight
				dbs.Timestamp = s.GetTimestamp()
				dbs.SetVMHash(nil)
				dbs.SetVMIndex(s.LeaderVMIndex)
				dbs.SetLocal(true)
				dbs.Sign(s)
				err := dbs.Sign(s)
				if err != nil {
					panic(err)
				}

				dbs.LeaderExecute(s)

			} else {
				for _, auditServer := range s.GetAuditServers(s.LLeaderHeight) {
					if auditServer.GetChainID().IsSameAs(s.IdentityChainID) {
						hb := new(messages.Heartbeat)
						hb.Timestamp = primitives.NewTimestampNow()
						hb.DBlockHash = dbstate.DBHash
						hb.IdentityChainID = s.IdentityChainID
						hb.Sign(s.GetServerPrivateKey())
						s.NetworkOutMsgQueue() <- hb
					}
				}
			}
			s.Saving = true
		}
	}

	for k := range s.Commits {
		vs := s.Commits[k]
		if len(vs) == 0 {
			delete(s.Commits, k)
			continue
		}
		v, ok := vs[0].(interfaces.IMsg)
		if ok {
			_, ok := s.Replay.Valid(constants.TIME_TEST, v.GetRepeatHash().Fixed(), v.GetTimestamp(), s.GetTimestamp())
			if !ok {
				copy(vs, vs[1:])
				vs[len(vs)-1] = nil
				s.Commits[k] = vs[:len(vs)-1]
			}
		}
	}

	for k := range s.Acks {
		v := s.Acks[k].(*messages.Ack)
		if v.DBHeight < s.LLeaderHeight {
			delete(s.Acks, k)
		}
	}

	return false
}

// When we process the directory Signature, and we are the leader for said signature, it
// is then that we push it out to the rest of the network.  Otherwise, if we are not the
// leader for the signature, it marks the sig complete for that list
func (s *State) ProcessDBSig(dbheight uint32, msg interfaces.IMsg) bool {

	dbs := msg.(*messages.DirectoryBlockSignature)
	// Don't process if syncing an EOM
	if s.Syncing && !s.DBSig {
		return false
	}

	pl := s.ProcessLists.Get(dbheight)
	vm := s.ProcessLists.Get(dbheight).VMs[msg.GetVMIndex()]

	// If we are done with DBSigs, and this message is processed, then we are done.  Let everything go!
	if s.DBSig && s.DBSigDone {
		s.DBSigProcessed--
		if s.DBSigProcessed <= 0 {
			s.DBSig = false
			s.Syncing = false
		}
		//s.LeaderPL.AdminBlock
		return true
	}

	// Put the stuff that only executes once at the start of DBSignatures here
	if !s.DBSig {
		s.DBSigLimit = len(pl.FedServers)
		s.DBSigProcessed = 0
		s.ProcessLists.Get(dbheight).VMs[dbs.VMIndex].Synced = false
		s.DBSig = true
		s.Syncing = true
		s.DBSigDone = false
		for _, vm := range pl.VMs {
			vm.Synced = false
		}
		pl.ResetDiffSigTally()
	}

	// Put the stuff that executes per DBSignature here
	if !dbs.Processed {
		if dbs.VMIndex == 0 {
			s.SetLeaderTimestamp(dbs.GetTimestamp())
		}
		if !dbs.DirectoryBlockHeader.GetBodyMR().IsSameAs(s.GetDBState(dbheight - 1).DirectoryBlock.GetHeader().GetBodyMR()) {
			fmt.Println(s.FactomNodeName, "JUST COMPARED", dbs.DirectoryBlockHeader.GetBodyMR().String()[:10], " : ", s.GetDBState(dbheight - 1).DirectoryBlock.GetHeader().GetBodyMR().String()[:10])
			pl.IncrementDiffSigTally()
		}

		// Adds DB Sig to be added to Admin block if passes sig checks
		allChecks := false
		data, err := dbs.DirectoryBlockHeader.MarshalBinary()
		if err != nil {
			fmt.Println("Debug: DBSig Signature Error, Marshal binary errored")
		} else {
			if !dbs.DBSignature.Verify(data) {
				fmt.Println("Debug: DBSig Signature Error, Verify errored")
			} else {
				if valid, err := s.VerifyAuthoritySignature(data, dbs.DBSignature.GetSignature(), dbs.DBHeight); err == nil && valid == 1 {
					allChecks = true
				}
			}
		}

		if allChecks {
			s.AddDBSig(dbheight, dbs.ServerIdentityChainID, dbs.DBSignature)
		}

		dbs.Processed = true
		s.DBSigProcessed++
		vm.Synced = true
	}

	// Put the stuff that executes once for set of DBSignatures (after I have them all) here
	if s.DBSigProcessed >= s.DBSigLimit {
		dbstate := s.DBStates.Get(int(dbheight - 1))

		// TODO: check signatures here.  Count what match and what don't.  Then if a majority
		// disagree with us, null our entry out.  Otherwise toss our DBState and ask for one from
		// our neighbors.
		if s.KeepMismatch || pl.CheckDiffSigTally() {
			if !dbstate.Saved {
				dbstate.ReadyToSave = true
				s.DBStates.SaveDBStateToDB(dbstate)
				//s.LeaderPL.AddDBSig(dbs.ServerIdentityChainID, dbs.DBSignature)
			}
		} else {
			s.DBSigFails++
			s.DBStates.DBStates = s.DBStates.DBStates[:len(s.DBStates.DBStates)-1]

			msg := messages.NewDBStateMissing(s, uint32(dbheight-1), uint32(dbheight-1))

			if msg != nil {
				s.RunLeader = false
				s.StartDelay = s.GetTimestamp().GetTimeMilli()
				s.NetworkOutMsgQueue() <- msg
			}
		}
		s.ReviewHolding()
		s.Saving = false
		s.DBSigDone = true
	}
	return false
	/*
		err := s.LeaderPL.AdminBlock.AddDBSig(dbs.ServerIdentityChainID, dbs.DBSignature)
		if err != nil {
			fmt.Printf("Error in adding DB sig to admin block, %s\n", err.Error())
		}
	*/
}

func (s *State) ConsiderSaved(dbheight uint32) {
	for _, dbs := range s.DBStates.DBStates {
		if dbs.DirectoryBlock.GetDatabaseHeight() == dbheight {
			dbs.Saved = true
		}
	}
}

func (s *State) GetNewEBlocks(dbheight uint32, hash interfaces.IHash) interfaces.IEntryBlock {
	pl := s.ProcessLists.Get(dbheight)
	if pl == nil {
		return nil
	}
	return pl.GetNewEBlocks(hash)
}

func (s *State) PutNewEBlocks(dbheight uint32, hash interfaces.IHash, eb interfaces.IEntryBlock) {
	pl := s.ProcessLists.Get(dbheight)
	pl.AddNewEBlocks(hash, eb)
}

func (s *State) PutNewEntries(dbheight uint32, hash interfaces.IHash, e interfaces.IEntry) {
	pl := s.ProcessLists.Get(dbheight)
	pl.AddNewEntry(hash, e)
}

// Returns the oldest, not processed, Commit received
func (s *State) NextCommit(hash interfaces.IHash) interfaces.IMsg {
	cs := s.Commits[hash.Fixed()]
	if cs == nil {
		return nil
	}

	if len(cs) == 0 {
		delete(s.Commits, hash.Fixed())
		return nil
	}

	r := cs[0]

	copy(cs[:], cs[1:])
	cs[len(cs)-1] = nil
	s.Commits[hash.Fixed()] = cs[:len(cs)-1]

	return r
}

func (s *State) PutCommit(hash interfaces.IHash, msg interfaces.IMsg) {
	cs := s.Commits[hash.Fixed()]
	if cs == nil {
		cs = make([]interfaces.IMsg, 0)
	}
	s.Commits[hash.Fixed()] = append(cs, msg)
}

// This is the highest block signed off and recorded in the Database.
func (s *State) GetHighestRecordedBlock() uint32 {
	return s.DBStates.GetHighestRecordedBlock()
}

// This is lowest block currently under construction under the "leader".
func (s *State) GetLeaderHeight() uint32 {
	return s.LLeaderHeight
}

// The highest block for which we have received a message.  Sometimes the same as
// BuildingBlock(), but can be different depending or the order messages are recieved.
func (s *State) GetHighestKnownBlock() uint32 {
	if s.ProcessLists == nil {
		return 0
	}
	plh := s.ProcessLists.DBHeightBase + uint32(len(s.ProcessLists.Lists)-1)
	dbsh := s.DBStates.Base + uint32(len(s.DBStates.DBStates))
	if dbsh > plh {
		return dbsh
	}
	return plh
}

func (s *State) GetF(adr [32]byte) int64 {
	s.FactoidBalancesTMutex.Lock()
	defer s.FactoidBalancesTMutex.Unlock()

	if v, ok := s.FactoidBalancesT[adr]; !ok {
		s.FactoidBalancesPMutex.Lock()
		defer s.FactoidBalancesPMutex.Unlock()
		v = s.FactoidBalancesP[adr]
		return v
	} else {
		return v
	}
}

func (s *State) PutF(rt bool, adr [32]byte, v int64) {
	if rt {
		s.FactoidBalancesTMutex.Lock()
		defer s.FactoidBalancesTMutex.Unlock()
		s.FactoidBalancesT[adr] = v
	} else {
		s.FactoidBalancesPMutex.Lock()
		defer s.FactoidBalancesPMutex.Unlock()
		s.FactoidBalancesP[adr] = v
	}
}

func (s *State) GetE(adr [32]byte) int64 {
	s.ECBalancesTMutex.Lock()
	defer s.ECBalancesTMutex.Unlock()

	if v, ok := s.ECBalancesT[adr]; !ok {
		s.ECBalancesPMutex.Lock()
		defer s.ECBalancesPMutex.Unlock()
		v = s.ECBalancesP[adr]
		return v
	} else {
		return v
	}
}

func (s *State) PutE(rt bool, adr [32]byte, v int64) {
	if rt {
		s.ECBalancesTMutex.Lock()
		defer s.ECBalancesTMutex.Unlock()
		s.ECBalancesT[adr] = v
	} else {
		s.ECBalancesPMutex.Lock()
		defer s.ECBalancesPMutex.Unlock()
		s.ECBalancesP[adr] = v
	}
}

// Returns the Virtual Server Index for this hash if this server is the leader;
// returns -1 if we are not the leader for this hash
func (s *State) ComputeVMIndex(hash []byte) int {
	return s.LeaderPL.VMIndexFor(hash)
}

func (s *State) GetNetworkName() string {
	return (s.Cfg.(util.FactomdConfig)).App.Network

}

func (s *State) GetDBHeightComplete() uint32 {
	db := s.GetDirectoryBlock()
	if db == nil {
		return 0
	}
	return db.GetHeader().GetDBHeight()
}

func (s *State) GetDirectoryBlock() interfaces.IDirectoryBlock {
	if s.DBStates.Last() == nil {
		return nil
	}
	return s.DBStates.Last().DirectoryBlock
}

func (s *State) GetNewHash() interfaces.IHash {
	return new(primitives.Hash)
}

// Create a new Acknowledgement.  Must be called by a leader.  This
// call assumes all the pieces are in place to create a new acknowledgement
func (s *State) NewAck(msg interfaces.IMsg) (iack interfaces.IMsg) {

	vmIndex := msg.GetVMIndex()

	msg.SetLeaderChainID(s.IdentityChainID)
	ack := new(messages.Ack)
	ack.DBHeight = s.LLeaderHeight
	ack.VMIndex = vmIndex
	ack.Minute = byte(s.ProcessLists.Get(s.LLeaderHeight).VMs[vmIndex].LeaderMinute)
	ack.Timestamp = s.GetTimestamp()
	ack.MessageHash = msg.GetMsgHash()
	ack.LeaderChainID = s.IdentityChainID

	listlen := len(s.LeaderPL.VMs[vmIndex].List)
	if listlen == 0 {
		ack.Height = 0
		ack.SerialHash = ack.MessageHash
	} else {
		last := s.LeaderPL.GetAckAt(vmIndex, listlen-1)
		ack.Height = last.Height + 1
		ack.SerialHash, _ = primitives.CreateHash(last.MessageHash, ack.MessageHash)
	}

	ack.Sign(s)

	return ack
}
