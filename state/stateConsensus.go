// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"bytes"
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
	acks := s.Acks
	ack, ok := acks[m.GetHash().Fixed()].(*messages.Ack)
	if !ok || ack == nil {
		s.Holding[m.GetHash().Fixed()] = m
		return false, nil
	} else {
		pl := s.ProcessLists.Get(ack.DBHeight)

		if m.Type() == constants.COMMIT_CHAIN_MSG || m.Type() == constants.COMMIT_ENTRY_MSG {
			s.PutCommits(m.GetHash(), m)
		}

		if pl != nil {
			pl.AddToProcessList(ack, m)

			pl.OldAcks[ack.GetHash().Fixed()] = ack
			pl.OldMsgs[m.GetHash().Fixed()] = m
			delete(acks, m.GetHash().Fixed())
			delete(s.Holding, m.GetHash().Fixed())
		} else {
			s.Println(">>>>>>>>>>>>>>>>>> Nil Process List at: ", ack.DBHeight)
			s.Println(">>>>>>>>>>>>>>>>>> Ack:                 ", ack.String())
			s.Println(">>>>>>>>>>>>>>>>>> DBStates:\n", s.DBStates.String())
			s.Println("\n\n>>>>>>>>>>>>>>>>>> ProcessLists:\n", s.ProcessLists.String())
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

	ht := dbstatemsg.DirectoryBlock.GetHeader().GetDBHeight()
	if ht >= s.LLeaderHeight {
		s.LLeaderHeight = ht + 1
	}
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

func (s *State) LeaderExecuteAddServer(server interfaces.IMsg) error {
	return s.LeaderExecute(server)
}

func (s *State) LeaderExecuteEOM(m interfaces.IMsg) error {
	eom, _ := m.(*messages.EOM)
	eom.DirectoryBlockHeight = s.LLeaderHeight
	if err := s.LeaderExecute(m); err != nil {
		return err
	}
	if eom.Minute == 9 {
		s.LLeaderHeight++
	}
	return nil
}

func (s *State) ProcessAddServer(dbheight uint32, addServerMsg interfaces.IMsg) bool {
	as, ok := addServerMsg.(*messages.AddServerMsg)
	if !ok {
		return true
	}
	server := new(interfaces.Server)
	server.ChainID = as.ServerChainID
	plc := s.ProcessLists.Get(dbheight)
	pl := s.ProcessLists.Get(dbheight + 1)

	s.Println("Current Server List:")
	for _, fed := range plc.FedServers {
		pl.AddFedServer(fed.(*interfaces.Server))
		s.Println("  ", fed.GetChainID().String())
	}

	pl.AddFedServer(server)

	s.Println("New Server List:")
	for _, fed := range pl.FedServers {
		s.Println("  ", fed.GetChainID().String())
	}

	return true
}

func (s *State) ProcessCommitChain(dbheight uint32, commitChain interfaces.IMsg) bool {
	c, ok := commitChain.(*messages.CommitChainMsg)
	if ok {
		pl := s.ProcessLists.Get(dbheight)
		ecblk := pl.EntryCreditBlock
		ecbody := ecblk.GetBody()
		ecbody.AddEntry(c.CommitChain)
		s.GetFactoidState().UpdateECTransaction(true, c.CommitChain)
		s.PutCommits(c.GetHash(), c)
	}
	return true
}

// TODO: Should fault the server if we don't have the proper sequence of EOM messages.
func (s *State) ProcessEOM(dbheight uint32, msg interfaces.IMsg) bool {

	e, ok := msg.(*messages.EOM)
	if !ok {
		panic("Must pass an EOM message to ProcessEOM)")
	}
	
	last := s.DBStates.Last()
	if e.Minute == 0 && (last == nil || !last.Saved) {
		return false
	}
	

	pl := s.ProcessLists.Get(dbheight)

	s.FactoidState.EndOfPeriod(int(e.Minute))

	if e.Minute == 9 {
		pl.SetEomComplete(e.ServerIndex, true)

		ecblk := pl.EntryCreditBlock
		ecbody := ecblk.GetBody()
		mn := entryCreditBlock.NewMinuteNumber2(e.Minute)
		ecbody.AddEntry(mn)

		//		fmt.Println("Process List ")
		// Should ensure we don't register the directory block multiple times.
		s.AddDBState(true, pl.DirectoryBlock, pl.AdminBlock, pl.FactoidBlock, pl.EntryCreditBlock)

		if s.LLeaderHeight <= dbheight {
			s.LLeaderHeight = dbheight + 1
		}

		found, index := s.GetFedServerIndex(s.LLeaderHeight)
		if found && e.ServerIndex == index {
			dbstate := s.DBStates.Get(dbheight)
			DBS := messages.NewDirectoryBlockSignature(dbheight)
			DBS.DirectoryBlockKeyMR = dbstate.DirectoryBlock.GetKeyMR()
			DBS.Timestamp = s.GetTimestamp()
			DBS.ServerIdentityChainID = s.IdentityChainID
			DBS.Sign(s)

			hash := DBS.GetHash()

			ack, _ := s.NewAck(dbheight, DBS, hash)

			// Leader Execute creates an acknowledgement and the EOM
			s.NetworkOutMsgQueue() <- ack
			s.NetworkOutMsgQueue() <- DBS
			ack.FollowerExecute(s)
			DBS.FollowerExecute(s)
		}
	}

	return true
}

// When we process the directory Signature, and we are the leader for said signature, it
// is then that we push it out to the rest of the network.  Otherwise, if we are not the
// leader for the signature, it marks the sig complete for that list
func (s *State) ProcessDBSig(dbheight uint32, msg interfaces.IMsg) bool {

	dbs := msg.(*messages.DirectoryBlockSignature)

	if msg.Leader(s) {
		hash := dbs.GetHash()
		ack, _ := s.NewAck(dbs.DBHeight, msg, hash)
		s.NetworkOutMsgQueue() <- dbs
		s.NetworkOutMsgQueue() <- ack
	}

	return true
}

func (s *State) GetNewEBlocks(dbheight uint32, hash interfaces.IHash) interfaces.IEntryBlock {
	return nil
}
func (s *State) PutNewEBlocks(dbheight uint32, hash interfaces.IHash, eb interfaces.IEntryBlock) {
}

func (s *State) GetCommits(hash interfaces.IHash) interfaces.IMsg {
	return nil
}
func (s *State) PutCommits(hash interfaces.IHash, msg interfaces.IMsg) {

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
func (s *State) LeaderFor([]byte) bool {
	s.SetString()
	found, index := s.GetFedServerIndex(s.GetLeaderHeight())

	if !found {
		return false
	}
	if index == 0 {
		return true
	}
	return false
}

func (s *State) GetFedServerIndex(dbheight uint32) (bool, int) {
	return s.GetFedServerIndexFor(dbheight, s.IdentityChainID)
}

// Gets the Server Index for an identity chain given the current leader height.
// The follower could be behind this level.
func (s *State) GetFedServerIndexFor(dbheight uint32, chainID interfaces.IHash) (bool, int) {
	pl := s.ProcessLists.Get(dbheight)

	if pl == nil {
		if bytes.Compare(chainID.Bytes(), s.CoreChainID.Bytes()) == 0 {
			return true, 0
		} else {
			s.Println(" No Process List for: ", dbheight)
			return false, 0
		}
	}

	if s.serverState == 1 && len(pl.FedServers) == 0 {
		pl.AddFedServer(&interfaces.Server{ChainID: s.IdentityChainID})
		//fmt.Println("Current Servers (Adding):")
		//for _, fed := range pl.FedServers {
		//	fmt.Println("   ", fed.GetChainID().String())
		//}
	}

	found, index := pl.GetFedServerIndex(chainID)

	return found, index
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

	found, index := s.GetFedServerIndex(dbheight)
	if !found {
		return nil, fmt.Errorf(s.FactomNodeName + ": Creation of an Ack attempted by non-server")
	}
	pl := s.ProcessLists.Get(dbheight)
	if pl == nil {
		return nil, fmt.Errorf(s.FactomNodeName + ": No process list at this time")
	}
	last, ok := pl.GetLastAck(index).(*messages.Ack)

	ack := new(messages.Ack)
	ack.DBHeight = dbheight

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
	pl.SetLastAck(index, ack)

	ack.Sign(s)

	return ack, nil
}
