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
	GetCfg() IFactomConfig
	Init(string)
	String() string
	Sign([]byte) IFullSignature

	GetServer() IServer
	SetServer(IServer)

	// Channels
	//==========

	// Network Processor
	NetworkInMsgQueue() chan IMsg // Not sure that IMsg is the right type... TBD
	NetworkOutMsgQueue() chan IMsg
	NetworkInvalidMsgQueue() chan IMsg

	// Consensus
	InMsgQueue() chan IMsg         // Read by Validate
	LeaderInMsgQueue() chan IMsg   // Processed by the Leader
	FollowerInMsgQueue() chan IMsg // Processed by the Follower

	// Maps
	// ====
	// The leader CANNOT touch these maps!  Only the Follower Execution
	// methods can touch them safely.
	GetHolding() map[[32]byte]IMsg // Hold Messages
	GetAcks() map[[32]byte]IMsg    // Hold Acknowledgemets

	// Lists
	// =====
	// The leader CANNOT touch these lists!  Only the FollowerExecution
	// methods can touch them safely.
	GetAuditServers() []IServer   // List of Audit Servers
	GetFedServers() []IServer     // List of Federated Servers
	GetServerOrder() [][]IServer  // 10 lists for Server Order for each minute
	GetProcessList() [][]IMsg     // List of Processed Messages Per server
	GetAuditHeartBeats() []IMsg   // The checklist of HeartBeats for this period
	GetFedServerFaults() [][]IMsg // Keep a fault list for every server

	// Server Configuration
	// ====================

	//Network MAIN = 0, TEST = 1, LOCAL = 2, CUSTOM = 3
	GetNetworkNumber() int  // Encoded into Directory Blocks
	GetNetworkName() string // Some networks have defined names

	// Number of Servers acknowledged by Factom
	GetTotalServers() int
	GetServerState() int    // (0 if client, 1 if server, 2 if audit server
	GetMatryoshka() []IHash // Reverse Hash

	LeaderFor([]byte) bool // Tests if this server is the leader for this key

	// Database
	// ========
	GetDB() DBOverlay
	SetDB(DBOverlay)

	// Directory Block State
	// =====================
	GetPreviousDirectoryBlock() IDirectoryBlock // The previous directory block
	GetCurrentDirectoryBlock() IDirectoryBlock  // The directory block under construction
	GetCurrentEntryCreditBlock() IEntryCreditBlock
	GetCurrentAdminBlock() IAdminBlock
	SetCurrentAdminBlock(IAdminBlock)
	SetCurrentDirectoryBlock(IDirectoryBlock)
	GetDBHeight() uint32 // The index of the directory block under construction.

	// Message State
	GetLastAck() IMsg // Return the last Acknowledgement set by this server

	// Server Methods
	// ==============
	ProcessEndOfBlock()

	// Web Services
	// ============
	SetPort(int)
	GetPort() int

	// Factoid State
	// =============
	GetFactoidState() IFactoidState
	GetPrevFactoidKeyMR() IHash
	SetPrevFactoidKeyMR(IHash)

	// MISC
	// ====
	MatchAckFollowerExecute(m IMsg) error
	GetTimestamp() Timestamp
	GetNewHash() IHash // Return a new Hash object
	CreateDBlock() (b IDirectoryBlock, err error)
	PrintType(int) bool // Debugging

	RecalculateBalances() error
	LogInfo(args ...interface{})
}
