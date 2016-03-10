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

	dbState := s.DBStates.NewDBState(isNew, directoryBlock, adminBlock, factoidBlock, entryCreditBlock)
	s.DBStates.Put(dbState)
}

// This routine is called once we have everything to create a Directory Block.
// It is called by the follower code.  It is requried to build the Directory Block
// to validate the signatures we will get with the DirectoryBlockSignature messages.
func (s *State) ProcessEndOfBlock(dbheight uint32) {
	s.LastAck = nil
}

// This returns the DBHeight as defined by the leader, not the follower.
// This value shouldn't be used by follower code.
func (s *State) GetDBHeight() uint32 {
	last := s.DBStates.Last()
	if last == nil {
		return 0
	}
	return last.DirectoryBlock.GetHeader().GetDBHeight()
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
		if pl != nil {
			pl.AddToProcessList(ack, m)
		}
		delete(acks, m.GetHash().Fixed())
		delete(s.Holding, m.GetHash().Fixed())

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
	fmt.Println("DBS Addstate")
	dbstatemsg, ok := msg.(*messages.DBStateMsg)
	if !ok {
		return fmt.Errorf("Cannot execute the given DBStateMsg")
	}

	s.DBStates.last = s.GetTimestamp()

	s.AddDBState(true,
		dbstatemsg.DirectoryBlock,
		dbstatemsg.AdminBlock,
		dbstatemsg.FactoidBlock,
		dbstatemsg.EntryCreditBlock)

	return nil
}

func (s *State) LeaderExecute(m interfaces.IMsg) error {

	hash := m.GetHash()

	ack, err := s.NewAck(hash)
	if err != nil {
		return err
	}

	// Leader Execute creates an acknowledgement and the EOM
	s.NetworkOutMsgQueue() <- ack
	s.NetworkOutMsgQueue() <- m
	ack.FollowerExecute(s)
	m.FollowerExecute(s)
	return nil
}

func (s *State) LeaderExecuteAddServer(server interfaces.IMsg) error {
	return s.LeaderExecute(server)
}

func (s *State) LeaderExecuteEOM(m interfaces.IMsg) error {
	eom, _ := m.(*messages.EOM)
	eom.DirectoryBlockHeight = s.GetHighestKnownBlock()
	return s.LeaderExecute(eom)
}

func (s *State) LeaderExecuteDBSig(m interfaces.IMsg) error {
	bblock := s.GetBuildingBlock()
	s.ProcessLists.Get(bblock).SetComplete(true)
	s.LeaderHeight = bblock + 1
	s.LastAck = nil
	return nil
}

func (s *State) GetNewEBlocks(dbheight uint32, hash interfaces.IHash) interfaces.IEntryBlock {
	return nil
}
func (s *State) PutNewEBlocks(dbheight uint32, hash interfaces.IHash, eb interfaces.IEntryBlock) {
}

func (s *State) GetCommits(dbheight uint32, hash interfaces.IHash) interfaces.IMsg {
	return nil
}
func (s *State) PutCommits(dbheight uint32, hash interfaces.IHash, msg interfaces.IMsg) {
	s.ProcessLists.Get(dbheight).PutCommits(hash, msg)
}

func (s *State) ProcessAddServer(dbheight uint32, addServerMsg interfaces.IMsg) {
	as, ok := addServerMsg.(*messages.AddServerMsg)
	if !ok {
		return
	}
	server := new(interfaces.Server)
	server.ChainID = as.ServerChainID
	plc := s.ProcessLists.Get(dbheight)
	pl := s.ProcessLists.Get(dbheight + 1)
	fmt.Println("Current Server List:")
	for _, fed := range plc.FedServers {
		pl.AddFedServer(fed.(*interfaces.Server))
		fmt.Println("  ", fed.GetChainID().String())
	}
	pl.AddFedServer(server)
	fmt.Println("New Server List:")
	for _, fed := range pl.FedServers {
		fmt.Println("  ", fed.GetChainID().String())
	}
}

func (s *State) ProcessCommitChain(dbheight uint32, commitChain interfaces.IMsg) {
	c, ok := commitChain.(*messages.CommitChainMsg)
	if ok {
		pl := s.ProcessLists.Get(dbheight)
		ecblk := pl.EntryCreditBlock
		ecbody := ecblk.GetBody()
		ecbody.AddEntry(c.CommitChain)
		s.GetFactoidState().UpdateECTransaction(true, c.CommitChain)
		s.PutCommits(dbheight, c.GetHash(), c)
	}
}

func (s *State) ProcessEOM(dbheight uint32, msg interfaces.IMsg) {
	e, ok := msg.(*messages.EOM)
	if !ok {
		panic("Must pass an EOM message to ProcessEOM)")
	}

	pl := s.ProcessLists.Get(dbheight)

	s.FactoidState.EndOfPeriod(int(e.Minute))

	ecblk := pl.EntryCreditBlock

	ecbody := ecblk.GetBody()
	mn := entryCreditBlock.NewMinuteNumber2(e.Minute)

	ecbody.AddEntry(mn)

	if e.Minute == 9 {

		// TODO: This code needs to be reviewed... It works here, but we are effectively
		// executing "leader" code in the compainion "follower" goroutine...
		// Maybe that's okay?
		if s.LeaderFor(e.Bytes()) {
			// What really needs to happen is that we look to make sure all
			// EOM messages have been recieved.  If this is the LAST message,
			// and we have ALL EOM messages from all servers, then we
			// create a DirectoryBlockSignature (if we are the leader) and
			// send it out to the network.
			DBS := messages.NewDirectoryBlockSignature(dbheight)
			DBS.Timestamp = s.GetTimestamp()
			prevDB := s.GetDirectoryBlock()
			if prevDB == nil {
				DBS.DirectoryBlockKeyMR = primitives.NewHash(constants.ZERO_HASH)
			} else {
				DBS.DirectoryBlockKeyMR = prevDB.GetKeyMR()
			}
			DBS.Sign(s)

			ack, err := s.NewAck(DBS.GetHash())
			if err != nil {
				return
			}

			s.NetworkOutMsgQueue() <- ack
			s.NetworkOutMsgQueue() <- DBS
			s.InMsgQueue() <- ack
			s.InMsgQueue() <- DBS
		}
	}
}

// This is the highest block signed off and recorded in the Database.
func (s *State) GetHighestRecordedBlock() uint32 {
	return s.DBStates.GetHighestRecordedBlock()
}

// This is lowest block currently under construction.
func (s *State) GetBuildingBlock() uint32 {
	return s.ProcessLists.GetBuildingBlock()
}

// The highest block for which we have received a message.  Sometimes the same as
// BuildingBlock(), but can be different depending or the order messages are recieved.
func (s *State) GetHighestKnownBlock() uint32 {
	if s.ProcessLists == nil {
		return 0
	}
	return s.ProcessLists.GetHighestKnownBlock()
}

func (s *State) ProcessDBS(dbheight uint32, commitChain interfaces.IMsg) {
	s.ProcessLists.Get(dbheight).SetSigComplete(true)
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
	found, index := s.GetFedServerIndex()

	if !found {
		return false
	}
	if index == 0 {
		return true
	}
	return false
}

func (s *State) GetFedServerIndex() (bool, int) {
	return s.GetFedServerIndexFor(s.IdentityChainID)
}

func (s *State) GetFedServerIndexFor(chainID interfaces.IHash) (bool, int) {
	pl := s.ProcessLists.Get(s.GetBuildingBlock())

	if pl == nil {
		fmt.Println("No Process List", s.GetBuildingBlock())
		return false, 0
	}

	if s.serverState == 1 && len(pl.FedServers) == 0 {
		pl.AddFedServer(&interfaces.Server{ChainID: s.IdentityChainID})
		fmt.Println("Current Servers (Adding):")
		for _, fed := range pl.FedServers {
			fmt.Println("   ", fed.GetChainID().String())
		}
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
	dbstate := s.DBStates.Last()
	if dbstate == nil {
		header.PrevFullHash = primitives.NewHash(constants.ZERO_HASH)
	} else {
		keymr, err := dbstate.AdminBlock.FullHash()
		if err != nil {
			panic(err.Error())
		}
		header.PrevFullHash = keymr
	}
	header.HeaderExpansionSize = 0
	header.HeaderExpansionArea = make([]byte, 0)
	header.MessageCount = 0
	header.BodySize = 0
	return header
}

func (s *State) PrintType(msgType int) bool {
	r := true
	r = r && msgType != constants.DBSTATE_MISSING_MSG
	r = r && msgType != constants.DBSTATE_MSG
	r = r && msgType != constants.ACK_MSG
	r = r && msgType != constants.EOM_MSG
	r = r && msgType != constants.DIRECTORY_BLOCK_SIGNATURE_MSG
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
func (s *State) NewAck(hash interfaces.IHash) (iack interfaces.IMsg, err error) {

	last, ok := s.LastAck.(*messages.Ack)

	ack := new(messages.Ack)
	ack.DBHeight = s.GetBuildingBlock()

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
	s.LastAck = ack

	// TODO:  Add the signature.

	return ack, nil
}

func (s *State) LoadDBState(dbheight uint32) (interfaces.IMsg, error) {

	dblk, err := s.DB.FetchDBlockByHeight(dbheight)
	if err != nil {
		return nil, err
	}
	if dblk == nil {
		return nil, nil
	}
	ablk, err := s.DB.FetchABlockByKeyMR(dblk.GetDBEntries()[0].GetKeyMR())
	if err != nil {
		return nil, err
	}
	if ablk == nil {
		return nil, err
	}
	ecblk, err := s.DB.FetchECBlockByHash(dblk.GetDBEntries()[1].GetKeyMR())
	if err != nil {
		return nil, err
	}
	if ecblk == nil {
		return nil, err
	}
	fblk, err := s.DB.FetchFBlockByKeyMR(dblk.GetDBEntries()[2].GetKeyMR())
	if err != nil {
		return nil, err
	}
	if fblk == nil {
		return nil, err
	}

	msg := messages.NewDBStateMsg(s, dblk, ablk, fblk, ecblk)

	return msg, nil

}

func (s *State) NewEOM(minute int) interfaces.IMsg {
	// The construction of the EOM message needs information from the state of
	// the server to create the proper serial hashes and such.  Right now
	// I am ignoring all of that.
	eom := new(messages.EOM)
	eom.Timestamp = s.GetTimestamp()
	eom.ChainID = s.IdentityChainID
	eom.Minute = byte(minute)
	eom.ServerIndex = s.ServerIndex
	eom.DirectoryBlockHeight = s.GetBuildingBlock()

	return eom
}
