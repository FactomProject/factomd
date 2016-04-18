// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

// Holds the state information for factomd.  This does imply that we will be
// using accessors to access state information in the consensus algorithm.
// This is a bit tedious, but does provide single choke points where information
// can be logged about the execution of Factom.  Also ensures that we do not
// accidentally
type IState interface {

	// Server
	GetFactomNodeName() string
	Clone(number string) IState
	GetCfg() IFactomConfig
	LoadConfig(filename string, folder string) // TODO JAYJAY remove folder here (hack to support multiple factomd processes on one .factom)
	Init()
	String() string
	GetIdentityChainID() IHash
	SetIdentityChainID(IHash)
	Sign([]byte) IFullSignature
	GetDirectoryBlockInSeconds() int
	SetDirectoryBlockInSeconds(int)
	GetServer() IServer
	GetFactomdVersion() int
	GetProtocolVersion() int
	SetServer(IServer)
	GetDBHeightComplete() uint32
	SetOut(bool)  // Output is turned on if set to true
	GetOut() bool // Return true if Print or Println write output
	LoadDBState(dbheight uint32) (IMsg, error)
	LoadSpecificMsg(dbheight uint32, plistheight uint32) (IMsg, error)
	LoadSpecificMsgAndAck(dbheight uint32, plistheight uint32) (IMsg, IMsg, error)
	GetFedServerIndexHash(uint32, IHash) (bool, int)
	SetString()
	ShortString() string

	AddFedServer(uint32, IHash) int
	GetFedServers(uint32) []IFctServer

	Green() bool

	// This is the highest block signed off and recorded in the Database.  This
	// is a follower's state, but it is also critical to validation; we cannot
	// validate transactions where the HighestRecordedBlock+1 != block holding said
	// transaction.
	GetHighestRecordedBlock() uint32
	// This is the Leader's view of the Height. It must be == HighestRecordedBlock+1.  Since
	// Recording a block can take time, messages must be queued until the previous block is
	// recorded (either by processing messages, or timing out and Leaders signing off the block)
	GetLeaderHeight() uint32
	// The highest block for which we have received a message. This is a
	// Follower's understanding of the Height, and reflects what block
	// is receiving messages.
	GetHighestKnownBlock() uint32

	// Find a Directory Block by height
	GetDirectoryBlockByHeight(dbheight uint32) IDirectoryBlock
	// Channels
	//==========

	// Network Processor
	TimerMsgQueue() chan IMsg
	NetworkOutMsgQueue() chan IMsg
	NetworkInvalidMsgQueue() chan IMsg

	// Journalling
	JournalMessage(IMsg)

	// Consensus
	InMsgQueue() chan IMsg     // Read by Validate
	LeaderMsgQueue() chan IMsg // Leader Queue
	Undo() IMsg

	// Lists and Maps
	// =====
	GetAuditHeartBeats() []IMsg   // The checklist of HeartBeats for this period
	GetFedServerFaults() [][]IMsg // Keep a fault list for every server

	GetNewEBlocks(dbheight uint32, hash IHash) IEntryBlock
	PutNewEBlocks(dbheight uint32, hash IHash, eb IEntryBlock)
	PutNewEntries(dbheight uint32, hash IHash, eb IEntry)

	GetCommits(hash IHash) IMsg
	GetReveals(hash IHash) IMsg
	PutCommits(hash IHash, msg IMsg)
	PutReveals(hash IHash, msg IMsg)
	// Server Configuration
	// ====================

	//Network MAIN = 0, TEST = 1, LOCAL = 2, CUSTOM = 3
	GetNetworkNumber() int  // Encoded into Directory Blocks
	GetNetworkName() string // Some networks have defined names

	GetMatryoshka(dbheight uint32) IHash // Reverse Hash

	// These are methods run by the consensus algorithm to track what servers are the leaders
	// and what lists they are responsible for.
	ServerIndexFor(uint32, []byte) int // Returns the serverindex responsible for this hash
	LeaderFor(hash []byte) bool        // Tests if this server is the leader for this key

	// Database
	// ========
	GetDB() DBOverlay
	SetDB(DBOverlay)

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

	ProcessAddServer(dbheight uint32, addServerMsg IMsg) bool
	ProcessCommitChain(dbheight uint32, commitChain IMsg) bool
	ProcessCommitEntry(dbheight uint32, commitChain IMsg) bool
	ProcessDBSig(dbheight uint32, commitChain IMsg) bool
	ProcessEOM(dbheight uint32, eom IMsg) bool

	// For messages that go into the Process List
	LeaderExecute(m IMsg) error
	LeaderExecuteEOM(m IMsg) error
	LeaderExecuteDBSig(m IMsg) error

	GetTimestamp() Timestamp

	PrintType(int) bool // Debugging
	Print(a ...interface{}) (n int, err error)
	Println(a ...interface{}) (n int, err error)

	ValidatorLoop()
	Dethrottle()

	SetIsReplaying()
	SetIsDoneReplaying()
}
