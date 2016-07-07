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
	GetFactomdVersion() int
	GetDBHeightComplete() uint32
	GetEBDBHeightComplete() uint32
	SetEBDBHeightComplete(uint32)
	DatabaseContains(hash IHash) bool
	SetOut(bool)  // Output is turned on if set to true
	GetOut() bool // Return true if Print or Println write output
	LoadDataByHash(requestedHash IHash) (BinaryMarshallable, int, error)
	LoadDBState(dbheight uint32) (IMsg, error)
	LoadSpecificMsg(dbheight uint32, vm int, plistheight uint32) (IMsg, error)
	LoadSpecificMsgAndAck(dbheight uint32, vm int, plistheight uint32) (IMsg, IMsg, error)
	SetString()
	ShortString() string

	AddPrefix(string)
	AddFedServer(uint32, IHash) int
	GetFedServers(uint32) []IFctServer
	RemoveFedServer(uint32, IHash)
	AddAuditServer(uint32, IHash) int
	GetAuditServers(uint32) []IFctServer

	// Routine for handling the syncroniztion of the leader and follower processes
	// and how they process messages.
	Process() (progress bool)
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
	TickerQueue() chan int
	TimerMsgQueue() chan IMsg
	NetworkOutMsgQueue() chan IMsg
	NetworkInvalidMsgQueue() chan IMsg

	// Journalling
	JournalMessage(IMsg)

	// Consensus
	APIQueue() chan IMsg   // Input Queue from the API
	InMsgQueue() chan IMsg // Read by Validate
	AckQueue() chan IMsg   // Leader Queue
	MsgQueue() chan IMsg   // Follower Queue

	// Lists and Maps
	// =====
	GetAuditHeartBeats() []IMsg   // The checklist of HeartBeats for this period
	GetFedServerFaults() [][]IMsg // Keep a fault list for every server

	GetNewEBlocks(dbheight uint32, hash IHash) IEntryBlock
	PutNewEBlocks(dbheight uint32, hash IHash, eb IEntryBlock)
	PutNewEntries(dbheight uint32, hash IHash, eb IEntry)

	NextCommit(hash IHash) IMsg
	PutCommit(hash IHash, msg IMsg)

	IncEntryChains()
	IncEntries()
	// Server Configuration
	// ====================

	//Network MAIN = 0, TEST = 1, LOCAL = 2, CUSTOM = 3
	GetNetworkNumber() int  // Encoded into Directory Blocks
	GetNetworkName() string // Some networks have defined names

	GetMatryoshka(dbheight uint32) IHash // Reverse Hash

	// These are methods run by the consensus algorithm to track what servers are the leaders
	// and what lists they are responsible for.
	ComputeVMIndex(hash []byte) int // Returns the VMIndex determined by some hash (usually) for the current processlist
	IsLeader() bool                 // Returns true if this is the leader in the current minute
	GetLeaderVM() int               // Get the Leader VM (only good within a minute)
	// Returns the list of VirtualServers at a given directory block height and minute
	GetVirtualServers(dbheight uint32, minute int, identityChainID IHash) (found bool, index int)
	// Returns true if between minutes

	GetEBlockKeyMRFromEntryHash(entryHash IHash) IHash
	GetAnchor() IAnchor

	// Database
	GetAndLockDB() DBOverlay
	UnlockDB()

	// Web Services
	// ============
	SetPort(int)
	GetPort() int

	// Factoid State
	// =============
	UpdateState() bool
	GetFactoidState() IFactoidState

	SetFactoidState(dbheight uint32, fs IFactoidState)
	GetFactoshisPerEC() uint64
	SetFactoshisPerEC(factoshisPerEC uint64)
	IncFactoidTrans()
	// MISC
	// ====

	FollowerExecuteMsg(m IMsg)     // Messages that go into the process list
	FollowerExecuteEOM(m IMsg)     // Messages that go into the process list
	FollowerExecuteAck(m IMsg)     // Ack Msg calls this function.
	FollowerExecuteDBState(IMsg)   // Add the given DBState to this server
	FollowerExecuteAddData(m IMsg) // Add the entry or eblock to this Server
	FollowerExecuteSFault(m IMsg)  // Handle Server Fault Messages
	FollowerExecuteMMR(m IMsg)     // Handle Missing Message Responses

	ProcessAddServer(dbheight uint32, addServerMsg IMsg) bool
	ProcessRemoveServer(dbheight uint32, removeServerMsg IMsg) bool
	ProcessChangeServerKey(dbheight uint32, changeServerKeyMsg IMsg) bool
	ProcessCommitChain(dbheight uint32, commitChain IMsg) bool
	ProcessCommitEntry(dbheight uint32, commitChain IMsg) bool
	ProcessDBSig(dbheight uint32, commitChain IMsg) bool
	ProcessEOM(dbheight uint32, eom IMsg) bool
	ProcessRevealEntry(dbheight uint32, m IMsg) bool
	// For messages that go into the Process List
	LeaderExecute(m IMsg)
	LeaderExecuteEOM(m IMsg)
	LeaderExecuteRevealEntry(m IMsg)

	GetNetStateOff() bool //	If true, all network communications are disabled
	SetNetStateOff(bool)

	GetTimestamp() Timestamp
	GetTimeOffset() Timestamp

	Print(a ...interface{}) (n int, err error)
	Println(a ...interface{}) (n int, err error)

	ValidatorLoop()

	AddDataRequest(requestedHash, missingDataHash IHash)
	HasDataRequest(checkHash IHash) bool
	GetAllEntries(ebKeyMR IHash) bool

	SetIsReplaying()
	SetIsDoneReplaying()

	//For ACK
	GetACKStatus(hash IHash) (int, IHash, Timestamp, Timestamp, error)
	FetchPaidFor(hash IHash) (IHash, error)
	FetchFactoidTransactionByHash(hash IHash) (ITransaction, error)
	FetchECTransactionByHash(hash IHash) (IECBlockEntry, error)
	FetchEntryByHash(IHash) (IEBEntry, error)

	// FER section
	ProcessRecentFERChainEntries()
	ExchangeRateAuthorityIsValid(IEBEntry) bool
	FerEntryIsValid(passedFEREntry IFEREntry) bool
	GetPredictiveFER() uint64

	// Identity Section
	VerifyIsAuthority(cid IHash) bool // True if is authority
	UpdateAuthorityFromABEntry(entry IABEntry) error

	//Authority Section
	VerifyFederatedSignature(Message []byte, signature *[64]byte) (bool, error)
}
