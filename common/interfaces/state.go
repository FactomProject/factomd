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
	GetSalt(Timestamp) uint32 // A secret number computed from a TS that tests if a message was issued from this server or not
	Clone(number int) IState
	GetCfg() IFactomConfig
	LoadConfig(filename string, networkFlag string)
	Init()
	String() string
	GetIdentityChainID() IHash
	SetIdentityChainID(IHash)
	Sign([]byte) IFullSignature
	GetDirectoryBlockInSeconds() int
	SetDirectoryBlockInSeconds(int)
	GetFactomdVersion() int
	GetDBHeightComplete() uint32
	DatabaseContains(hash IHash) bool
	SetOut(bool)  // Output is turned on if set to true
	GetOut() bool // Return true if Print or Println write output
	LoadDataByHash(requestedHash IHash) (BinaryMarshallable, int, error)
	LoadDBState(dbheight uint32) (IMsg, error)
	LoadSpecificMsg(dbheight uint32, vm int, plistheight uint32) (IMsg, error)
	LoadSpecificMsgAndAck(dbheight uint32, vm int, plistheight uint32) (IMsg, IMsg, error)
	SetString()
	ShortString() string
	GetStatus() []string
	AddStatus(status string)

	AddDBSig(dbheight uint32, chainID IHash, sig IFullSignature)
	AddPrefix(string)
	AddFedServer(uint32, IHash) int
	GetFedServers(uint32) []IFctServer
	RemoveFedServer(uint32, IHash)
	AddAuditServer(uint32, IHash) int
	GetAuditServers(uint32) []IFctServer
	GetOnlineAuditServers(uint32) []IFctServer

	//RPC
	GetRpcUser() string
	GetRpcPass() string
	SetRpcAuthHash(authHash []byte)
	GetRpcAuthHash() []byte
	GetTlsInfo() (bool, string, string)
	GetFactomdLocations() string

	// Routine for handling the syncroniztion of the leader and follower processes
	// and how they process messages.
	Process() (progress bool)
	// This is the highest block completed.  It may or may not be saved in the Database.  This
	// is a follower's state, but it is also critical to validation; we cannot
	// validate transactions where the HighestRecordedBlock+1 != block holding said
	// transaction.
	GetHighestSavedBlk() uint32
	// This is the highest block saved in the Database. A block is completed, then validated
	// then saved.
	GetHighestCompletedBlk() uint32
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
	GetAuditHeartBeats() []IMsg // The checklist of HeartBeats for this period

	GetNewEBlocks(dbheight uint32, hash IHash) IEntryBlock
	PutNewEBlocks(dbheight uint32, hash IHash, eb IEntryBlock)
	PutNewEntries(dbheight uint32, hash IHash, eb IEntry)

	GetPendingEntries(interface{}) []IPendingEntry
	NextCommit(hash IHash) IMsg
	PutCommit(hash IHash, msg IMsg)

	IncEntryChains()
	IncEntries()
	IncECCommits()
	IncECommits()
	IncFCTSubmits()

	// Server Configuration
	// ====================

	//Network MAIN = 0, TEST = 1, LOCAL = 2, CUSTOM = 3
	GetNetworkNumber() int  // Encoded into Directory Blocks
	GetNetworkName() string // Some networks have defined names
	GetNetworkID() uint32

	// Bootstrap Identity Information is dependent on Network
	GetNetworkBootStrapKey() IHash
	GetNetworkBootStrapIdentity() IHash

	GetMatryoshka(dbheight uint32) IHash // Reverse Hash

	// These are methods run by the consensus algorithm to track what servers are the leaders
	// and what lists they are responsible for.
	ComputeVMIndex(hash []byte) int // Returns the VMIndex determined by some hash (usually) for the current processlist
	IsLeader() bool                 // Returns true if this is the leader in the current minute
	GetLeaderVM() int               // Get the Leader VM (only good within a minute)
	// Returns the list of VirtualServers at a given directory block height and minute
	GetVirtualServers(dbheight uint32, minute int, identityChainID IHash) (found bool, index int)
	// Returns true if between minutes

	// Get the message for the given vm index, dbheight, and height.  Returns nil if I
	// have no such message.
	GetMsg(vmIndex int, dbheight int, height int) (IMsg, error)

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
	GetSystemHeight(dbheight uint32) int
	GetFactoidState() IFactoidState

	SetFactoidState(dbheight uint32, fs IFactoidState)
	GetFactoshisPerEC() uint64
	SetFactoshisPerEC(factoshisPerEC uint64)
	IncFactoidTrans()
	IncDBStateAnswerCnt()

	GetPendingTransactions(interface{}) []IPendingTransaction
	// MISC
	// ====

	Reset()                                    // Trim back the state to the last saved block
	GetSystemMsg(dbheight, height uint32) IMsg // Return the system message at the given height.
	SendDBSig(dbheight uint32, vmIndex int)    // If a Leader, we have to send a DBSig out for the previous block

	FollowerExecuteMsg(IMsg)          // Messages that go into the process list
	FollowerExecuteEOM(IMsg)          // Messages that go into the process list
	FollowerExecuteAck(IMsg)          // Ack Msg calls this function.
	FollowerExecuteDBState(IMsg)      // Add the given DBState to this server
	FollowerExecuteSFault(IMsg)       // Handling of Server Fault Messages
	FollowerExecuteFullFault(IMsg)    // Handle Server Full-Fault Messages
	FollowerExecuteMMR(IMsg)          // Handle Missing Message Responses
	FollowerExecuteDataResponse(IMsg) // Handle Data Response
	FollowerExecuteMissingMsg(IMsg)   // Handle requests for missing messages
	FollowerExecuteCommitChain(IMsg)  // CommitChain needs to look for a Reveal Entry
	FollowerExecuteCommitEntry(IMsg)  // CommitEntry needs to look for a Reveal Entry
	FollowerExecuteRevealEntry(IMsg)

	ProcessAddServer(dbheight uint32, addServerMsg IMsg) bool
	ProcessRemoveServer(dbheight uint32, removeServerMsg IMsg) bool
	ProcessChangeServerKey(dbheight uint32, changeServerKeyMsg IMsg) bool
	ProcessCommitChain(dbheight uint32, commitChain IMsg) bool
	ProcessCommitEntry(dbheight uint32, commitChain IMsg) bool
	ProcessDBSig(dbheight uint32, commitChain IMsg) bool
	ProcessEOM(dbheight uint32, eom IMsg) bool
	ProcessRevealEntry(dbheight uint32, m IMsg) bool
	ProcessFullServerFault(dbheight uint32, fullFault IMsg) bool
	// For messages that go into the Process List
	LeaderExecute(IMsg)
	LeaderExecuteEOM(IMsg)
	LeaderExecuteDBSig(IMsg)
	LeaderExecuteRevealEntry(IMsg)
	LeaderExecuteCommitChain(IMsg)
	LeaderExecuteCommitEntry(IMsg)

	GetNetStateOff() bool //	If true, all network communications are disabled
	SetNetStateOff(bool)

	GetTimestamp() Timestamp
	GetTimeOffset() Timestamp

	GetTrueLeaderHeight() uint32
	Print(a ...interface{}) (n int, err error)
	Println(a ...interface{}) (n int, err error)

	ValidatorLoop()

	UpdateECs(IEntryCreditBlock)
	SetIsReplaying()
	SetIsDoneReplaying()
	// No Entry Yet returns true if no Entry Hash is found in the Replay structs.
	// Returns false if we have seen an Entry Replay in the current period.
	NoEntryYet(IHash, Timestamp) bool

	//For ACK
	GetACKStatus(hash IHash) (int, IHash, Timestamp, Timestamp, error)
	FetchPaidFor(hash IHash) (IHash, error)
	FetchFactoidTransactionByHash(hash IHash) (ITransaction, error)
	FetchECTransactionByHash(hash IHash) (IECBlockEntry, error)
	FetchEntryByHash(IHash) (IEBEntry, error)
	FetchEntryHashFromProcessListsByTxID(string) (IHash, error)

	// FER section
	ProcessRecentFERChainEntries()
	ExchangeRateAuthorityIsValid(IEBEntry) bool
	FerEntryIsValid(passedFEREntry IFEREntry) bool
	GetPredictiveFER() uint64

	// Identity Section
	VerifyIsAuthority(cid IHash) bool // True if is authority
	UpdateAuthorityFromABEntry(entry IABEntry) error
	VerifyAuthoritySignature(Message []byte, signature *[64]byte, dbheight uint32) (int, error)
	FastVerifyAuthoritySignature(Message []byte, signature IFullSignature, dbheight uint32) (int, error)
	UpdateAuthSigningKeys(height uint32)

	AddAuthorityDelta(changeString string)

	GetLLeaderHeight() uint32
	GetEntryDBHeightComplete() uint32
	GetMissingEntryCount() uint32
	GetEntryBlockDBHeightProcessing() uint32
	GetEntryBlockDBHeightComplete() uint32
	GetCurrentMinute() int

	// Access to Holding Queue
	LoadHoldingMap() map[[32]byte]IMsg
	LoadAcksMap() map[[32]byte]IMsg
}
