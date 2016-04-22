// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"fmt"
	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/util"
)

var _ = fmt.Print

//***************************************************************
// Consensus Methods
//***************************************************************

// Adds blocks that are either pulled locally from a database, or acquired from peers.
func (s *State) AddDBState(isNew bool,
	directoryBlock interfaces.IDirectoryBlock,
	adminBlock interfaces.IAdminBlock,
	factoidBlock interfaces.IFBlock,
	entryCreditBlock interfaces.IEntryCreditBlock) {

	// TODO:  Need to validate before we add, or at least validate once we have a contiguous set of blocks.

	// 	fmt.Printf("AddDBState %s: DirectoryBlock %d %x %x %x %x\n",
	// 			   s.FactomNodeName,
	// 			   directoryBlock.GetHeader().GetDBHeight(),
	// 			   directoryBlock.GetKeyMR().Bytes()[:5],
	// 			   adminBlock.GetHash().Bytes()[:5],
	// 			   factoidBlock.GetHash().Bytes()[:5],
	// 			   entryCreditBlock.GetHash().Bytes()[:5])

	dbState := s.DBStates.NewDBState(isNew, directoryBlock, adminBlock, factoidBlock, entryCreditBlock)
	s.DBStates.Put(dbState)

}

// Messages that will go into the Process List must match an Acknowledgement.
// The code for this is the same for all such messages, so we put it here.
//
// Returns true if it finds a match
func (s *State) FollowerExecuteMsg(m interfaces.IMsg) (bool, error) {

	hash := m.GetHash()
	hashf := hash.Fixed()
	ack, ok := s.Acks[hashf].(*messages.Ack)
	if !ok || ack == nil {
		s.Holding[hashf] = m
		return false, nil
	} else {
		pl := s.ProcessLists.Get(ack.DBHeight)

		if m.Type() == constants.COMMIT_CHAIN_MSG || m.Type() == constants.COMMIT_ENTRY_MSG {
			s.PutCommits(hash, m)
		}

		if pl != nil {
			pl.AddToProcessList(ack, m)

			pl.OldAcks[hashf] = ack
			pl.OldMsgs[hashf] = m
			delete(s.Acks, hashf)
			delete(s.Holding, hashf)
		}
		return true, nil
	}
}

// Ack messages always match some message in the Process List.   That is
// done here, though the only msg that should call this routine is the Ack
// message.
func (s *State) FollowerExecuteAck(msg interfaces.IMsg) (bool, error) {
	ack := msg.(*messages.Ack)
	s.Acks[ack.GetHash().Fixed()] = ack
	match := s.Holding[ack.GetHash().Fixed()]
	if match != nil {
		match.FollowerExecute(s)
		return true, nil
	}

	return false, nil
}

func (s *State) FollowerExecuteDBState(msg interfaces.IMsg) error {

	dbstatemsg, ok := msg.(*messages.DBStateMsg)
	if !ok {
		return fmt.Errorf("Cannot execute the given DBStateMsg")
	}

	s.DBStates.LastTime = s.GetTimestamp()
	//	fmt.Println("DBState Message  ")
	s.AddDBState(true,
		dbstatemsg.DirectoryBlock,
		dbstatemsg.AdminBlock,
		dbstatemsg.FactoidBlock,
		dbstatemsg.EntryCreditBlock)

	return nil
}

func (s *State) LeaderExecute(m interfaces.IMsg) error {

	hash := m.GetHash()

	ack, err := s.NewAck(s.LLeaderHeight, m, hash)
	if err != nil {
		return err
	}

	// Leader Execute creates an acknowledgement and the EOM
	s.NetworkOutMsgQueue() <- ack
	s.NetworkOutMsgQueue() <- m
	s.InMsgQueue() <- ack
	m.FollowerExecute(s)
	return nil
}

func (s *State) LeaderExecuteEOM(m interfaces.IMsg) error {
	eom, _ := m.(*messages.EOM)
	eom.DBHeight = s.LLeaderHeight
	if err := s.LeaderExecute(m); err != nil {
		fmt.Println("Error: ", err)
		return err
	}

	return nil
}

func (s *State) LeaderExecuteDBSig(m interfaces.IMsg) error {
	dbs, ok := m.(*messages.DirectoryBlockSignature)
	if !ok {
		return fmt.Errorf("Bad Directory Block Signature")
	}

	dbs.DBHeight = s.LLeaderHeight

	dbs.Timestamp = s.GetTimestamp()
	hash := dbs.GetHash()
	ack, err := s.NewAck(s.LLeaderHeight, dbs, hash)
	if err != nil {
		fmt.Println("Bad Ack")
		s.undo = m
		return nil
	}
	ack.FollowerExecute(s)
	dbs.FollowerExecute(s)

	s.LLeaderHeight++

	return nil
}

func (s *State) ProcessAddServer(dbheight uint32, addServerMsg interfaces.IMsg) bool {
	as, ok := addServerMsg.(*messages.AddServerMsg)
	if !ok {
		fmt.Println("Bad Msg: ", addServerMsg.String())
		return true
	}

	pl := s.ProcessLists.Get(dbheight)
	pl.AdminBlock.AddFedServer(as.ServerChainID)

	return true
}

func (s *State) ProcessCommitChain(dbheight uint32, commitChain interfaces.IMsg) bool {
	c, ok := commitChain.(*messages.CommitChainMsg)
	if !ok {
		return false
	}

	pl := s.ProcessLists.Get(dbheight)
	pl.EntryCreditBlock.GetBody().AddEntry(c.CommitChain)
	s.GetFactoidState().UpdateECTransaction(true, c.CommitChain)

	// save the Commit to match agains the Reveal later
	s.PutCommits(c.GetHash(), c)
	// check for a matching Reveal and, if found, execute it
	if r := s.GetReveals(c.GetHash()); r != nil {
		s.LeaderExecute(r)
	}

	return true
}

func (s *State) ProcessCommitEntry(dbheight uint32, commitEntry interfaces.IMsg) bool {
	c, ok := commitEntry.(*messages.CommitEntryMsg)
	if !ok {
		return false
	}

	pl := s.ProcessLists.Get(dbheight)
	pl.EntryCreditBlock.GetBody().AddEntry(c.CommitEntry)
	s.GetFactoidState().UpdateECTransaction(true, c.CommitEntry)

	// save the Commit to match agains the Reveal later
	s.PutCommits(c.GetHash(), c)
	// check for a matching Reveal and, if found, execute it
	if r := s.GetReveals(c.GetHash()); r != nil {
		s.LeaderExecute(r)
	}

	return true
}

// TODO: Should fault the server if we don't have the proper sequence of EOM messages.
func (s *State) ProcessEOM(dbheight uint32, msg interfaces.IMsg) bool {

	e, ok := msg.(*messages.EOM)
	if !ok {
		panic("Must pass an EOM message to ProcessEOM)")
	}

	pl := s.ProcessLists.Get(dbheight)

	// Set this list complete
	pl.SetMinute(e.ServerIndex, int(e.Minute))

    if pl.MinuteHeight() <= int(e.Minute) {
        return false
    }


	if !e.MarkerSent {
		if s.ServerIndexFor(e.DBHeight, constants.FACTOID_CHAINID) == e.ServerIndex {
			s.FactoidState.EndOfPeriod(int(e.Minute))
		}
		if s.ServerIndexFor(e.DBHeight, constants.ADMIN_CHAINID) == e.ServerIndex {
			pl.AdminBlock.AddEndOfMinuteMarker(e.Minute)
		}
		e.MarkerSent = true
	}

	// We need to have all EOM markers before we start to clean up this height.
	if e.Minute == 9 {

         // Maybe we want to check the block is saved?   
        	// We need to save away the previous state before we begin to process the next height
	        // last := s.DBStates.Last()
	        // if last == nil || !last.Saved {
        	//    return false
	        // }


		if s.ServerIndexFor(e.DBHeight, constants.EC_CHAINID) == e.ServerIndex {
			ecblk := pl.EntryCreditBlock
			ecbody := ecblk.GetBody()
			mn := entryCreditBlock.NewMinuteNumber2(e.Minute)
			ecbody.AddEntry(mn)
		}
	}

	// Add EOM to the EBlocks
	for _, eb := range pl.NewEBlocks {
		eb.AddEndOfMinuteMarker(e.Bytes()[0])
	}

	return true
}

// When we process the directory Signature, and we are the leader for said signature, it
// is then that we push it out to the rest of the network.  Otherwise, if we are not the
// leader for the signature, it marks the sig complete for that list
func (s *State) ProcessDBSig(dbheight uint32, msg interfaces.IMsg) bool {
	pl := s.ProcessLists.Get(dbheight)
	if !pl.EomComplete() {
		return false
	}

	dbs, ok := msg.(*messages.DirectoryBlockSignature)
	if !ok {
		panic("DirectoryBlockSignature is the wrong type.")
	}

	if !pl.Servers[dbs.ServerIndex].SigComplete {
		pl.SetSigComplete(int(dbs.ServerIndex), true)

		s.AddDBState(true, pl.DirectoryBlock, pl.AdminBlock, s.GetFactoidState().GetCurrentBlock(), pl.EntryCreditBlock)
	}

	if dbs.IsLocal() {
		dbstate := s.DBStates.Get(dbheight)

		dbs2 := new(messages.DirectoryBlockSignature)
		dbs2.Timestamp = s.GetTimestamp()
		dbs2.ServerIdentityChainID = dbs.ServerIdentityChainID
		dbs2.DBHeight = dbs.DBHeight
		dbs2.ServerIndex = dbs.ServerIndex
		dbs2.DirectoryBlockKeyMR = dbstate.DirectoryBlock.GetKeyMR()
		err := dbs2.Sign(s)
		if err != nil {
			panic(err)
		}

		hash := dbs2.GetHash()

		// Here we replace out of the process list the local DBS message with one
		// that can be broadcast.  This is a bit of necessary trickery
		pl.UndoLeaderAck(int(dbs.ServerIndex))
		s.LLeaderHeight--
		ack, err := s.NewAck(dbheight, dbs2, hash)
		if err != nil {
			panic(err)
		}
		s.LLeaderHeight++

		// Leader Execute creates an acknowledgement and the EOM
		s.NetworkOutMsgQueue() <- ack
		s.NetworkOutMsgQueue() <- dbs2
	} else {
		// TODO follower should validate signature here.
		resp := dbs.Validate(s)
		if resp == 1 {
			return true
		}
		return false
	}

	return true
}

func (s *State) GetNewEBlocks(dbheight uint32, hash interfaces.IHash) interfaces.IEntryBlock {
	pl := s.ProcessLists.Get(dbheight)
	return pl.GetNewEBlocks(hash)
}

func (s *State) PutNewEBlocks(dbheight uint32, hash interfaces.IHash, eb interfaces.IEntryBlock) {
	pl := s.ProcessLists.Get(dbheight)
	pl.PutNewEBlocks(dbheight, hash, eb)
}

func (s *State) PutNewEntries(dbheight uint32, hash interfaces.IHash, e interfaces.IEntry) {
	pl := s.ProcessLists.Get(dbheight)
	pl.PutNewEntries(dbheight, hash, e)
}

func (s *State) GetCommits(hash interfaces.IHash) interfaces.IMsg {
	return s.Commits[hash.Fixed()]
}

func (s *State) GetReveals(hash interfaces.IHash) interfaces.IMsg {
	return s.Reveals[hash.Fixed()]
}

func (s *State) PutCommits(hash interfaces.IHash, msg interfaces.IMsg) {
	cmsg, ok := msg.(interfaces.ICounted)
	if ok {
		v := s.Commits[hash.Fixed()]
		if v != nil {
			_, ok := v.(interfaces.ICounted)
			if ok {
				cmsg.SetCount(v.(interfaces.ICounted).GetCount() + 1)
			} else {
				panic("Should never happen")
			}
		}
	}
	s.Commits[hash.Fixed()] = msg
}

func (s *State) PutReveals(hash interfaces.IHash, msg interfaces.IMsg) {
	cmsg, ok := msg.(interfaces.ICounted)
	if ok {
		v := s.Reveals[hash.Fixed()]
		if v != nil {
			_, ok := v.(interfaces.ICounted)
			if ok {
				cmsg.SetCount(v.(interfaces.ICounted).GetCount() + 1)
			} else {
				panic("Should never happen")
			}
		}
	}
	s.Reveals[hash.Fixed()] = msg
}

// This is the highest block signed off and recorded in the Database.
func (s *State) GetHighestRecordedBlock() uint32 {
	return s.DBStates.GetHighestRecordedBlock()
}

// If Green, this server is, to the best of its knowledge, caught up with the
// network.  TODO there should be a timeout that requires seeing a message within
// some period of time, but not there yet.
//
// We hare caught up with the network IF:
// The highest recorded block is equal to or just below the highest known block
func (s *State) Green() bool {
	if s.GreenCnt > 1000 {
		return true
	}

	rec := s.DBStates.GetHighestRecordedBlock()
	high := s.GetHighestKnownBlock()
	s.GreenFlg = rec >= high-1
	if s.GreenFlg {
		s.GreenCnt++
	} else {
		s.GreenCnt = 0
	}
	return s.GreenFlg
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
	return s.ProcessLists.DBHeightBase + uint32(len(s.ProcessLists.Lists)-1)
}

func (s *State) GetF(adr [32]byte) int64 {
	if v, ok := s.FactoidBalancesT[adr]; !ok {
		v = s.FactoidBalancesP[adr]
		return v
	} else {
		return v
	}
}

func (s *State) PutF(rt bool, adr [32]byte, v int64) {
	if rt {
		s.FactoidBalancesT[adr] = v
	} else {
		s.FactoidBalancesP[adr] = v
	}
}

func (s *State) GetE(adr [32]byte) int64 {
	if v, ok := s.ECBalancesT[adr]; !ok {
		v = s.ECBalancesP[adr]
		return v
	} else {
		return v
	}
}

func (s *State) PutE(rt bool, adr [32]byte, v int64) {
	if rt {
		s.ECBalancesT[adr] = v
	} else {
		s.ECBalancesP[adr] = v
	}
}

// Tests the given hash, and returns true if this server is the leader for this key.
// For example, keys we test include:
//
// The Hash of the Factoid Hash
// Entry Credit Addresses
// ChainIDs
// ...
func (s *State) ServerIndexFor(dbheight uint32, hash []byte) int {
	pl := s.ProcessLists.Get(dbheight)
	if pl == nil {
		return 0
	}
	n := len(s.ProcessLists.Get(dbheight).FedServers)
	v := 0
	if len(hash) > 0 {
		v = int(hash[0]) % n
	}
	return v
}

func (s *State) LeaderFor(hash []byte) bool {
	found, index := s.GetFedServerIndexHash(s.LLeaderHeight, s.IdentityChainID)
	if !found {
		return false
	}
	return index == s.ServerIndexFor(s.LLeaderHeight, hash)
}

func (s *State) NewAdminBlock(dbheight uint32) interfaces.IAdminBlock {
	ab := new(adminBlock.AdminBlock)
	ab.Header = s.NewAdminBlockHeader(dbheight)
	return ab
}

func (s *State) NewAdminBlockHeader(dbheight uint32) interfaces.IABlockHeader {
	header := new(adminBlock.ABlockHeader)
	header.DBHeight = dbheight
	header.PrevFullHash = primitives.NewHash(constants.ZERO_HASH)
	header.HeaderExpansionSize = 0
	header.HeaderExpansionArea = make([]byte, 0)
	header.MessageCount = 0
	header.BodySize = 0
	return header
}

func (s *State) PrintType(msgType int) bool {
	r := true
	return r
	r = r && msgType != constants.ACK_MSG
	r = r && msgType != constants.EOM_MSG
	r = r && msgType != constants.DIRECTORY_BLOCK_SIGNATURE_MSG
	r = r && msgType != constants.DBSTATE_MISSING_MSG
	r = r && msgType != constants.DBSTATE_MSG
	return r
}

func (s *State) GetNetworkName() string {
	return (s.Cfg.(util.FactomdConfig)).App.Network

}

func (s *State) GetDB() interfaces.DBOverlay {
	return s.DB
}

func (s *State) SetDB(dbase interfaces.DBOverlay) {
	s.DB = databaseOverlay.NewOverlay(dbase)
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

// Create a new Acknowledgement.  This Acknowledgement
func (s *State) NewAck(dbheight uint32, msg interfaces.IMsg, hash interfaces.IHash) (iack interfaces.IMsg, err error) {

	found, index := s.GetFedServerIndexHash(dbheight, s.IdentityChainID)
	if !found {
		return nil, fmt.Errorf(s.FactomNodeName + ": Creation of an Ack attempted by non-server")
	}
	pl := s.ProcessLists.Get(dbheight)
	if pl == nil {
		return nil, fmt.Errorf(s.FactomNodeName + ": No process list at this time")
	}
	last, ok := pl.GetLastLeaderAck(index).(*messages.Ack)

	ack := new(messages.Ack)
	ack.DBHeight = dbheight
	ack.ServerIndex = byte(index)
	ack.Timestamp = s.GetTimestamp()
	ack.MessageHash = hash
	if !ok {
		ack.Height = 0
		ack.SerialHash = ack.MessageHash
	} else {
		ack.Height = last.Height + 1
		ack.SerialHash, err = primitives.CreateHash(last.MessageHash, ack.MessageHash)
		if err != nil {
			return nil, err
		}
	}
	pl.SetLastLeaderAck(index, ack)

	ack.Sign(s)

	return ack, nil
}
