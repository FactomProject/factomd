// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

import ()

// Holds the state information for factomd.  This does imply that we will be
// using accessors to access state information in the consensus algorithm.
// This is a bit tedious, but does provide single choke points where information
// can be logged about the execution of Factom.  Also ensures that we do not
// accidentally
type IState interface {

	// Server
	Clone(number string) IState
	GetCfg() IFactomConfig
	LoadConfig(filename string)
	Init()
	String() string
	GetServerIdentityChainID() IHash
	Sign([]byte) IFullSignature
	GetDirectoryBlockInSeconds() int
	GetServer() IServer
	SetServer(IServer)
	GetDBHeightComplete() uint32
	Print(a ...interface{}) (n int, err error)
	Println(a ...interface{}) (n int, err error)
	SetOut(bool)		// Output is turned on if set to true
	GetOut() bool		// Return true if Print or Println write output
	LoadDBState(dbheight uint32) (IMsg,error)
	LastCompleteDBHeight() uint32
	// Channels
	//==========

	// Network Processor
	NetworkInMsgQueue() chan IMsg // Not sure that IMsg is the right type... TBD
	NetworkOutMsgQueue() chan IMsg
	NetworkInvalidMsgQueue() chan IMsg

	// Consensus
	InMsgQueue() chan IMsg // Read by Validate

	// Lists and Maps
	// =====
	GetAuditHeartBeats() []IMsg   // The checklist of HeartBeats for this period
	GetFedServerFaults() [][]IMsg // Keep a fault list for every server

	GetNewEBlocks(dbheight uint32, hash IHash) IEntryBlock
	PutNewEBlocks(dbheight uint32, hash IHash, eb IEntryBlock)

	GetCommits(dbheight uint32, hash IHash) IMsg
	PutCommits(dbheight uint32, hash IHash, msg IMsg)
	// Server Configuration
	// ====================

	//Network MAIN = 0, TEST = 1, LOCAL = 2, CUSTOM = 3
	GetNetworkNumber() int  // Encoded into Directory Blocks
	GetNetworkName() string // Some networks have defined names

	// Number of Servers acknowledged by Factom
	GetTotalServers() int
	GetServerState() int                 // (0 if client, 1 if server, 2 if audit server
	GetMatryoshka(dbheight uint32) IHash // Reverse Hash

	LeaderFor(hash []byte) bool // Tests if this server is the leader for this key

	// Database
	// ========
	GetDB() DBOverlay
	SetDB(DBOverlay)

	// Directory Block State
	// =====================
	GetDBHeight() uint32
	// Get the last finished directory block
	GetDirectoryBlock() IDirectoryBlock

	GetAnchor() IAnchor

	// Web Services
	// ============
	SetPort(int)
	GetPort() int

	// Factoid State
	// =============
	UpdateState()
	GetFactoidState() IFactoidState

	SetFactoidState(dbheight uint32, fs IFactoidState)
	GetFactoshisPerEC() uint64
	SetFactoshisPerEC(factoshisPerEC uint64)
	// MISC
	// ====

	FollowerExecuteMsg(m IMsg) (bool, error) // Messages that go into the process list
	FollowerExecuteAck(m IMsg) (bool, error) // Ack Msg calls this function.
	FollowerExecuteDBState(IMsg) error       // Add the given DBState to this server
	ProcessAddServer(dbheight uint32, addServerMsg IMsg)
	ProcessCommitChain(dbheight uint32, commitChain IMsg)
	ProcessSignPL(dbheight uint32, commitChain IMsg)
	ProcessEOM(dbheight uint32, eom IMsg)

	// For messages that go into the Process List
	LeaderExecute(m IMsg) error
	LeaderExecuteAddServer(m IMsg) error
	LeaderExecuteEOM(m IMsg) error
	LeaderExecuteDBSig(m IMsg) error

	NewEOM(int) IMsg

	GetTimestamp() Timestamp
	PrintType(int) bool // Debugging

}
