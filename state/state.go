// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"sync"
	"time"

	"github.com/FactomProject/factomd/common"
	"github.com/FactomProject/factomd/common/constants/runstate"
	"github.com/FactomProject/factomd/generated"
	"github.com/FactomProject/factomd/modules/logging"
	"github.com/FactomProject/factomd/pubsub"
	"github.com/FactomProject/factomd/queue"

	"github.com/FactomProject/factomd/activations"
	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	. "github.com/FactomProject/factomd/common/identity"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/database/boltdb"
	"github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/leveldb"
	"github.com/FactomProject/factomd/database/mapdb"
	"github.com/FactomProject/factomd/p2p"
	"github.com/FactomProject/factomd/util"
	"github.com/FactomProject/factomd/util/atomic"
)

// loaded directly from factomParams
type StateConfig struct {
	CheckChainHeads struct {
		CheckChainHeads bool
		Fix             bool
	}
	CloneDBType             string
	ControlPanelPort        int
	ControlPanelSetting     int
	DBType                  string
	DirectoryBlockInSeconds int
	DropRate                int
	FactomdLocations        string
	FactomdTLSCertFile      string
	FactomdTLSEnable        bool
	FactomdTLSKeyFile       string
	FactomNodeName          string
	FastSaveRate            int
	FaultTimeout            int
	LeaderTimestamp         interfaces.Timestamp
	OutputAllowed           bool
	PortNumber              int
	preBootTime             *primitives.Timestamp
	Prefix                  string
	RpcAuthHash             []byte
	RpcPass                 string
	RpcUser                 string
	StartDelayLimit         int64
	TimeOffset              interfaces.Timestamp
	TimestampAtBoot         interfaces.Timestamp
	WaitForEntries          bool
	FactomdVersion          string
	LogLevel                string
	ConsoleLogLevel         string
	ExportData              bool
	// Network Configuration
	Network                 string
	MainNetworkPort         string
	PeersFile               string
	MainSeedURL             string
	MainSpecialPeers        string
	TestNetworkPort         string
	TestSeedURL             string
	TestSpecialPeers        string
	LocalNetworkPort        string
	LocalSeedURL            string
	LocalSpecialPeers       string
	CustomNetworkPort       string
	CustomSeedURL           string
	CustomSpecialPeers      string
	CustomNetworkID         []byte
	CustomBootstrapIdentity string
	CustomBootstrapKey      string
	FaultWait               int
	EOMfaultIndex           int
	FactoshisPerEC          uint64
	Port                    int
	RequestTimeout          time.Duration
	RequestLimit            int
	CorsDomains             []string
	NodeMode                string
}

type State struct {
	common.Name
	StateConfig
	logging           *logging.LayerLogger
	Pub               // Publisher hooks
	RunState          runstate.RunState
	NetworkController *p2p.Controller
	Salt              interfaces.IHash
	Cfg               interfaces.IFactomConfig
	ConfigFilePath    string // $HOME/.factom/m2/factomd.conf by default

	LogPath           string
	LdbPath           string
	BoltDBPath        string
	ExportDataSubpath string

	LogBits int64 // Bit zero is for logging the Directory Block on DBSig [5]

	DBStatesSent         []*interfaces.DBStateSent
	DBStatesReceivedBase int
	DBStatesReceived     []*messages.DBStateMsg
	LocalServerPrivKey   string
	Replay               *Replay
	FReplay              *Replay
	CrossReplay          *CrossReplayFilter
	Delay                int64 // Simulation delays sending inMessages this many milliseconds

	IdentityChainID interfaces.IHash // If this node has an identity, this is it
	//Identities      []*Identity      // Identities of all servers in management chain
	// Authorities          []*Authority     // Identities of all servers in management chain
	AuthorityServerCount int // number of federated or audit servers allowed
	IdentityControl      *IdentityManager

	// Just to print (so debugging doesn't drive functionality)
	Status      int // Return a status (0 do nothing, 1 provide queues, 2 provide consensus data)
	serverPrt   string
	StatusMutex sync.Mutex
	StatusStrs  []string
	Starttime   time.Time
	transCnt    int
	lasttime    time.Time
	tps         float64
	ResetTryCnt int
	ResetCnt    int

	//  pending entry/transaction api calls for the holding queue do not have proper scope
	//  This is used to create a temporary, correctly scoped holding queue snapshot for the calls on demand
	HoldingMutex sync.RWMutex
	HoldingLast  int64
	HoldingMap   map[[32]byte]interfaces.IMsg

	// Elections are managed through the Elections Structure
	EFactory  interfaces.IElectionsFactory
	Elections interfaces.IElections
	Election0 string // Title
	Election1 string // Election state for display
	Election2 string // Election state for display
	Election3 string // Election leader list

	//  pending entry/transaction api calls for the ack queue do not have proper scope
	//  This is used to create a temporary, correctly scoped ackqueue snapshot for the calls on demand
	AcksMutex sync.RWMutex
	AcksLast  int64
	AcksMap   map[[32]byte]interfaces.IMsg

	DBStateAskCnt     int
	DBStateReplyCnt   int
	DBStateIgnoreCnt  int
	DBStateAppliedCnt int

	MissingRequestAskCnt      int
	MissingRequestReplyCnt    int
	MissingRequestIgnoreCnt   int
	MissingResponseAppliedCnt int

	ResendCnt int
	ExpireCnt int

	tickerQueue            chan int
	timerMsgQueue          chan interfaces.IMsg
	MaxTimeOffset          interfaces.Timestamp
	networkOutMsgQueue     *queue.MsgQueue
	networkInvalidMsgQueue chan interfaces.IMsg
	inMsgQueue             *queue.MsgQueue
	inMsgQueue2            *queue.MsgQueue
	electionsQueue         *queue.MsgQueue
	apiQueue               *queue.MsgQueue
	ackQueue               chan interfaces.IMsg
	msgQueue               chan interfaces.IMsg
	// prioritizedMsgQueue contains inMessages we know we need for consensus. (missing from processlist)
	//		Currently inMessages from MMR handling can be put in here to fast track
	//		them to the front.
	prioritizedMsgQueue chan interfaces.IMsg

	ShutdownChan chan int // For gracefully halting Factom

	ServerPrivKey         *primitives.PrivateKey
	ServerPubKey          *primitives.PublicKey
	serverPendingPrivKeys []*primitives.PrivateKey
	serverPendingPubKeys  []*primitives.PublicKey

	// RPC connection config

	// Server State
	StartDelay   int64 // Time in Milliseconds since the last DBState was applied
	DBFinished   bool
	RunLeader    bool
	BootTime     int64 // Time in seconds that we last booted
	EOMIssueTime int64
	EOMSyncEnd   int64

	// Ignore missing inMessages for a period to allow rebooting a network where your
	// own inMessages from the previously executing network can confuse you.
	IgnoreDone    bool
	IgnoreMissing bool

	LLeaderHeight   uint32
	Leader          bool
	LeaderVMIndex   int
	LeaderPL        *ProcessList
	PLProcessHeight uint32
	// Height cutoff where no missing inMessages below this height
	DBHeightAtBoot uint32
	CurrentMinute  int

	// These are the start times for blocks and minutes
	PreviousMinuteStartTime int64
	CurrentMinuteStartTime  int64
	CurrentBlockStartTime   int64

	EOMsyncing   bool
	EOMSyncTime  int64
	EOM          bool // Set to true when the first EOM is encountered
	EOMLimit     int
	EOMProcessed int
	EOMDone      bool
	EOMMinute    int
	EOMSys       bool // At least one EOM has covered the System List

	DBSig          bool
	DBSigLimit     int
	DBSigProcessed int // Number of DBSignatures received and processed.
	DBSigDone      bool
	DBSigSys       bool // At least one DBSig has covered the System List

	CreatedLastBlockFromDBState bool

	// By default, this is false, which means DBstates are discarded
	// when a majority of leaders disagree with the hash we have via DBSigs
	KeepMismatch bool

	DBSigFails int // Keep track of how many blockhash mismatches we've had to correct

	Saving  bool // True if we are in the process of saving to the database
	Syncing bool // Looking for inMessages from leaders to sync

	NetStateOff            bool // Disable if true, Enable if false
	DebugConsensus         bool // If true, dump consensus trace
	FactoidTrans           int
	ECCommits              int
	ECommits               int
	FCTSubmits             int
	NewEntryChains         int
	NewEntries             int
	messageFilterTimestamp interfaces.Timestamp
	// Maps
	// ====
	// For Follower
	ResendHolding interfaces.Timestamp         // Timestamp to gate resending holding to neighbors
	HoldingList   chan [32]byte                // Queue to process Holding in order
	HoldingVM     int                          // VM used to build current holding list
	Holding       map[[32]byte]interfaces.IMsg // Hold Messages
	XReview       []interfaces.IMsg            // After the EOM, we must review the inMessages in Holding
	Acks          map[[32]byte]interfaces.IMsg // Hold Acknowledgements
	Commits       *SafeMsgMap                  //  map[[32]byte]interfaces.IMsg // Commit Messages

	InvalidMessages      map[[32]byte]interfaces.IMsg
	InvalidMessagesMutex sync.RWMutex

	AuditHeartBeats []interfaces.IMsg // The checklist of HeartBeats for this period

	LastTiebreak int64

	AuthoritySetString string
	// Network MAIN = 0, TEST = 1, LOCAL = 2, CUSTOM = 3
	NetworkNumber int // Encoded into Directory Blocks(s.Cfg.(*util.FactomdConfig)).String()

	// Database
	DB     interfaces.DBOverlaySimple
	Anchor interfaces.IAnchor

	// Directory Block State
	DBStates       *DBStateList // Holds all DBStates not yet processed.
	StatesMissing  *StatesMissing
	StatesWaiting  *StatesWaiting
	StatesReceived *StatesReceived

	// Having all the state for a particular directory block stored in one structure
	// makes creating the next state, updating the various states, and setting up the next
	// state much more simple.
	//
	// Functions that provide state information take a dbheight param.  I use the current
	// DBHeight to ensure that I return the proper information for the right directory block
	// height, even if it changed out from under the calling code.
	//
	// Process list previous [0], present(@DBHeight) [1], and future (@DBHeight+1) [2]

	ResetRequest    bool // Set to true to trigger a reset
	ProcessLists    *ProcessLists
	highestKnown    uint32
	highestAck      uint32
	AuthorityDeltas string

	// Factom State
	FactoidState    interfaces.IFactoidState
	NumTransactions int

	// Permanent balances from processing blocks.
	RestoreFCT            map[[32]byte]int64
	RestoreEC             map[[32]byte]int64
	FactoidBalancesPapi   map[[32]byte]int64
	FactoidBalancesP      map[[32]byte]int64
	FactoidBalancesPMutex sync.Mutex
	ECBalancesPapi        map[[32]byte]int64
	ECBalancesP           map[[32]byte]int64
	ECBalancesPMutex      sync.Mutex
	TempBalanceHash       interfaces.IHash
	Balancehash           interfaces.IHash

	// State for the Entry Syncing process
	EntrySyncState *EntrySync

	MissingEntryBlockRepeat interfaces.Timestamp
	// DBlock Height at which node has a complete set of eblocks+entries
	EntryBlockDBHeightComplete uint32
	// DBlock Height at which we have started asking for entry blocks
	EntryBlockDBHeightProcessing uint32
	// Entry Blocks we don't have that we are asking our neighbors for
	MissingEntryBlocks []MissingEntryBlock

	MissingEntryRepeat interfaces.Timestamp
	// DBlock Height at which node has a complete set of eblocks+entries
	EntryDBHeightComplete uint32
	// DBlock Height at which we have started asking for or have all entries
	EntryDBHeightProcessing uint32
	// Height in the Directory Block where we have
	// Entries we don't have that we are asking our neighbors for
	MissingEntries chan *MissingEntry

	// Holds leaders and followers up until all missing entries are processed, if true
	UpdateEntryHash chan *EntryUpdate // SubChannel for updating entry Hashes tracking (repeats and such)
	WriteEntry      chan interfaces.IEBEntry
	// MessageTally causes the node to keep track of (and display) running totals of each
	// type of message received during the tally interval
	MessageTally           bool
	MessageTalliesReceived [constants.NUM_MESSAGES]int
	MessageTalliesSent     [constants.NUM_MESSAGES]int

	LastPrint    string
	LastPrintCnt int

	// FER section
	FERChainId                     string
	ExchangeRateAuthorityPublicKey string

	FERChangeHeight      uint32
	FERChangePrice       uint64
	FERPriority          uint32
	FERPrioritySetHeight uint32

	AckChange uint32

	StateSaverStruct StateSaverStruct

	// Logstash
	UseLogstash bool
	LogstashURL string

	// Plugins
	FastBoot         bool
	FastBootLocation string

	// These stats are collected when we write the dbstate to the database.
	NumNewChains   int // Number of new Chains in this block
	NumNewEntries  int // Number of new Entries, not counting the first entry in a chain
	NumEntries     int // Number of entries in this block (including the entries that create chains)
	NumEntryBlocks int // Number of Entry Blocks
	NumFCTTrans    int // Number of Factoid Transactions in this block

	// debug message about state status rolling queue for ControlPanel
	SyncingState        [256]string
	SyncingStateCurrent int

	ProcessListProcessCnt int64 // count of attempts to process .. so we can see if the thread is running
	StateProcessCnt       int64
	StateUpdateState      int64
	ValidatorLoopSleepCnt int64
	processCnt            int64 // count of attempts to process .. so we can see if the thread is running
	ProcessTime           interfaces.Timestamp

	MMRInfo // fields for MMR processing

	reportedActivations       [activations.ACTIVATION_TYPE_COUNT + 1]bool // flags about which activations we have reported (+1 because we don't use 0)
	validatorLoopThreadID     string
	OutputRegEx               *regexp.Regexp
	OutputRegExString         string
	InputRegEx                *regexp.Regexp
	InputRegExString          string
	executeRecursionDetection map[[32]byte]interfaces.IMsg
	Hold                      *HoldingList

	// MissingMessageResponse is a cache of the last 1000 msgs we receive such that when
	// we send out a missing message, we can find that message locally before we ask the net
	RecentMessage

	// MissingMessageResponseHandler is a cache of the last 2 blocks of processed acks.
	// It can handle and respond to missing message requests on it's own thread.
	MissingMessageResponseHandler *MissingMessageResponseCache
	ChainCommits                  Last100
	Reveals                       Last100

	// publish subscribe hooks for new modules
	leaderTimestampPub *generated.Publish_PubBase_Timestamp_type     // Current Leader Timestamp
	addMsg             *generated.Publish_PubBase_CommitRequest_type // Add Comits/reveals to commit map
	metDependency      *generated.Publish_PubBase_Hash_type          // hash of met dependencies
	//	checkHash          *generated.Publish_PubBase_CommitRequest_type // check commits/reveals against commitmap

}

func (s *State) Publish() {
	s.leaderTimestampPub = generated.Publish_PubBase_Timestamp(pubsub.PubFactory.Base().Publish(pubsub.GetPath(s.GetParentName(), "leadertimestamp")))
	s.addMsg = generated.Publish_PubBase_CommitRequest(pubsub.PubFactory.Base().Publish(pubsub.GetPath(s.GetParentName(), "vms", "commits"))) // Add commits/reveals to commit map
	//	s.checkHash = generated.Publish_PubBase_CommitRequest(pubsub.PubFactory.Base().Publish(pubsub.GetPath(s.GetParentName(), "commitmap", "checks"))) // check commits/reveals against commit map
	s.metDependency = generated.Publish_PubBase_Hash(pubsub.PubFactory.Base().Publish("FNode0/leader/metDependency")) // check commits/reveals against commit map

}

func (s *State) Subscribe() {

}

func (s *State) ClosePublishing() {
	_ = s.leaderTimestampPub.Close()
}

var _ interfaces.IState = (*State)(nil)

type EntryUpdate struct {
	Hash      interfaces.IHash
	Timestamp interfaces.Timestamp
}

func (s *State) GetConfigPath() string {
	return s.ConfigFilePath
}

func (s *State) GetRunState() runstate.RunState {
	return s.RunState
}

func (s *State) AddPrefix(prefix string) {
	s.Prefix = prefix
}

func (s *State) GetFactomNodeName() string {
	return s.FactomNodeName
}

func (s *State) GetDBStatesSent() []*interfaces.DBStateSent {
	return s.DBStatesSent
}

func (s *State) SetDBStatesSent(sents []*interfaces.DBStateSent) {
	s.DBStatesSent = sents
}

func (s *State) GetDelay() int64 {
	return s.Delay
}

func (s *State) SetDelay(delay int64) {
	s.Delay = delay
}

func (s *State) GetBootTime() int64 {
	return s.BootTime
}

func (s *State) GetDropRate() int {
	return s.DropRate
}

func (s *State) SetDropRate(droprate int) {
	s.DropRate = droprate
}

func (s *State) SetAuthoritySetString(authSet string) {
	s.AuthoritySetString = authSet
}

func (s *State) GetAuthoritySetString() string {
	return s.AuthoritySetString
}

func (s *State) AddAuthorityDelta(authSet string) {
	s.AuthorityDeltas += fmt.Sprintf("\n%s", authSet)
}

func (s *State) GetAuthorityDeltas() string {
	return s.AuthorityDeltas
}

func (s *State) GetNetStateOff() bool { //	If true, all network communications are disabled
	return s.NetStateOff
}

func (s *State) SetNetStateOff(net bool) {
	//flag this in everyone!
	s.LogPrintf("executeMsg", "State.SetNetStateOff(%v)", net)
	s.LogPrintf("election", "State.SetNetStateOff(%v)", net)
	s.LogPrintf("InMsgQueue", "State.SetNetStateOff(%v)", net)
	s.LogPrintf("NetworkInputs", "State.SetNetStateOff(%v)", net)
	s.NetStateOff = net
}

func (s *State) GetRpcUser() string {
	return s.RpcUser
}

func (s *State) GetCorsDomains() []string {
	return s.CorsDomains
}
func (s *State) GetRpcPass() string {
	return s.RpcPass
}

func (s *State) SetRpcAuthHash(authHash []byte) {
	s.RpcAuthHash = authHash
}

func (s *State) GetRpcAuthHash() []byte {
	return s.RpcAuthHash
}

func (s *State) GetTlsInfo() (bool, string, string) {
	return s.FactomdTLSEnable, s.FactomdTLSKeyFile, s.FactomdTLSCertFile
}

func (s *State) GetFactomdLocations() string {
	return s.FactomdLocations
}

func (s *State) GetCurrentBlockStartTime() int64 {
	return s.CurrentBlockStartTime
}

func (s *State) GetCurrentMinute() int {
	return s.CurrentMinute
}

func (s *State) GetCurrentMinuteStartTime() int64 {
	return s.CurrentMinuteStartTime
}

func (s *State) GetPreviousMinuteStartTime() int64 {
	return s.PreviousMinuteStartTime
}

func (s *State) GetCurrentTime() int64 {
	return time.Now().UnixNano()
}

func (s *State) IsSyncing() bool {
	return s.Syncing
}

func (s *State) IsSyncingEOMs() bool {
	return s.Syncing && s.EOM && !s.EOMDone
}

func (s *State) IsSyncingDBSigs() bool {
	return s.Syncing && s.DBSig && !s.DBSigDone
}

func (s *State) DidCreateLastBlockFromDBState() bool {
	return s.CreatedLastBlockFromDBState
}

func (s *State) IncDBStateAnswerCnt() {
	s.DBStateReplyCnt++
}

func (s *State) IncFCTSubmits() {
	s.FCTSubmits++
}

func (s *State) IncECCommits() {
	s.ECCommits++
}

func (s *State) IncECommits() {
	s.ECommits++
}

func (s *State) GetAckChange() (bool, error) {
	var flag bool
	change, err := util.GetChangeAcksHeight(s.ConfigFilePath)
	if err != nil {
		return flag, err
	}
	flag = s.AckChange != change
	s.AckChange = change
	return flag, nil
}

func (s *State) GetSalt(ts interfaces.Timestamp) uint32 {
	if s.Salt == nil {
		b := make([]byte, 32)
		_, err := rand.Read(b)
		// Note that err == nil only if we read len(b) bytes.
		if err != nil {
			panic("Random Number Failure")
		}
		s.Salt = primitives.Sha(b)
	}

	var b [32]byte
	copy(b[:], s.Salt.Bytes())
	binary.BigEndian.PutUint64(b[:], uint64(ts.GetTimeMilli()))
	c := primitives.Sha(b[:])
	return binary.BigEndian.Uint32(c.Bytes())
}

func (s *State) GetEntryBlockDBHeightComplete() uint32 {
	return s.EntryBlockDBHeightComplete
}

func (s *State) SetEntryBlockDBHeightComplete(newHeight uint32) {
	s.EntryBlockDBHeightComplete = newHeight
}

func (s *State) GetEntryBlockDBHeightProcessing() uint32 {
	return s.EntryBlockDBHeightProcessing
}

func (s *State) SetEntryBlockDBHeightProcessing(newHeight uint32) {
	s.EntryBlockDBHeightProcessing = newHeight
}

func (s *State) GetLLeaderHeight() uint32 {
	return s.LLeaderHeight
}

func (s *State) GetFaultTimeout() int {
	return s.FaultTimeout
}

func (s *State) GetFaultWait() int {
	return s.FaultWait
}

func (s *State) GetEntryDBHeightComplete() uint32 {
	return s.EntryDBHeightComplete
}

func (s *State) GetMissingEntryCount() uint32 {
	return uint32(len(s.MissingEntries))
}

func (s *State) GetEBlockKeyMRFromEntryHash(entryHash interfaces.IHash) (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("State.GetEBlockKeyMRFromEntryHash() saw an interface that was nil")
		}
	}()
	entry, err := s.DB.FetchEntry(entryHash)
	if err != nil {
		return nil
	}
	if entry != nil {
		dblock := s.GetDirectoryBlockByHeight(entry.GetDatabaseHeight())
		for idx, ebHash := range dblock.GetEntryHashes() {
			if idx > 2 {
				thisBlock, err := s.DB.FetchEBlock(ebHash)
				if err == nil {
					for _, attemptEntryHash := range thisBlock.GetEntryHashes() {
						if attemptEntryHash.IsSameAs(entryHash) {
							return ebHash
						}
					}
				}
			}
		}
	}
	return nil
}

func (s *State) GetDB() interfaces.DBOverlaySimple {
	return s.DB
}

// Checks ChainIDs to determine if we need their entries to process entries and transactions.
func (s *State) Needed(eb interfaces.IEntryBlock) bool {
	id := []byte{0x88, 0x88, 0x88}
	fer := []byte{0x11, 0x11, 0x11}

	if eb.GetDatabaseHeight() < 2 {
		return true
	}
	cid := eb.GetChainID().Bytes()
	if bytes.Compare(id[:3], cid) == 0 {
		return true
	}
	if bytes.Compare(id[:3], fer) == 0 {
		return true
	}
	return false
}

func (s *State) ValidatePrevious(dbheight uint32) error {
	dblk, err := s.DB.FetchDBlockByHeight(dbheight)
	errs := ""
	if dblk != nil && err == nil && dbheight > 0 {

		if dblk2, err := s.DB.FetchDBlock(dblk.GetKeyMR()); err != nil {
			errs += "Don't have the directory block hash indexed %d\n"
		} else if dblk2 == nil {
			errs += fmt.Sprintf("Don't have the directory block hash indexed %d\n", dbheight)
		}

		pdblk, _ := s.DB.FetchDBlockByHeight(dbheight - 1)
		pdblk2, _ := s.DB.FetchDBlock(dblk.GetHeader().GetPrevKeyMR())
		if pdblk == nil {
			errs += fmt.Sprintf("Cannot find the previous block by index at %d", dbheight-1)
		} else {
			if pdblk.GetKeyMR().Fixed() != dblk.GetHeader().GetPrevKeyMR().Fixed() {
				errs += fmt.Sprintf("xxxx KeyMR incorrect at height %d", dbheight-1)
			}
			if pdblk.GetFullHash().Fixed() != dblk.GetHeader().GetPrevFullHash().Fixed() {
				fmt.Println("xxxx Full Hash incorrect at height", dbheight-1)
				return fmt.Errorf("Full hash incorrect block at %d", dbheight-1)
			}
		}
		if pdblk2 == nil {
			errs += fmt.Sprintf("Cannot find the previous block at %d", dbheight-1)
		} else {
			if pdblk2.GetKeyMR().Fixed() != dblk.GetHeader().GetPrevKeyMR().Fixed() {
				errs += fmt.Sprintln("xxxx Hash is incorrect.  Expected: ", dblk.GetHeader().GetPrevKeyMR().String())
				errs += fmt.Sprintln("xxxx Hash is incorrect.  Received: ", pdblk2.GetKeyMR().String())
				errs += fmt.Sprintf("Hash is incorrect at %d", dbheight-1)
			}
		}
	}
	if len(errs) > 0 {
		return errors.New(errs)
	}
	return nil
}

func (s *State) LoadDBState(dbheight uint32) (interfaces.IMsg, error) {
	dblk, err := s.DB.FetchDBlockByHeight(dbheight)
	if err != nil {
		return nil, err
	}

	err = s.ValidatePrevious(dbheight)
	if err != nil {
		panic(err.Error() + " " + s.FactomNodeName)
	}

	if dblk == nil {
		return nil, nil
	}

	ablk, err := s.DB.FetchABlock(dblk.GetDBEntries()[0].GetKeyMR())
	if err != nil {
		return nil, err
	}
	if ablk == nil {
		return nil, fmt.Errorf("%s", "ABlock not found")
	}
	ecblk, err := s.DB.FetchECBlock(dblk.GetDBEntries()[1].GetKeyMR())
	if err != nil {
		return nil, err
	}
	if ecblk == nil {
		return nil, fmt.Errorf("%s", "ECBlock not found")
	}
	fblk, err := s.DB.FetchFBlock(dblk.GetDBEntries()[2].GetKeyMR())
	if err != nil {
		return nil, err
	}
	if fblk == nil {
		return nil, fmt.Errorf("%s", "FBlock not found")
	}
	if bytes.Compare(fblk.GetKeyMR().Bytes(), dblk.GetDBEntries()[2].GetKeyMR().Bytes()) != 0 {
		panic("Should not happen")
	}

	var eBlocks []interfaces.IEntryBlock
	var entries []interfaces.IEBEntry

	ebDBEntries := dblk.GetEBlockDBEntries()
	if len(ebDBEntries) > 0 {
		for _, v := range ebDBEntries {
			eBlock, err := s.DB.FetchEBlock(v.GetKeyMR())
			if err == nil && eBlock != nil {
				eBlocks = append(eBlocks, eBlock)
				if s.Needed(eBlock) {
					for _, e := range eBlock.GetEntryHashes() {
						entry, err := s.DB.FetchEntry(e)
						if err == nil && entry != nil {
							entries = append(entries, entry)
						}
					}
				}
			}
		}
	}

	dbaseID := dblk.GetHeader().GetNetworkID()
	configuredID := s.GetNetworkID()
	if dbaseID != configuredID {
		panic(fmt.Sprintf("The configured network ID (%x) differs from the one in the local database (%x) at height %d", configuredID, dbaseID, dbheight))
	}

	var allSigs []interfaces.IFullSignature

	nextABlock, err := s.DB.FetchABlockByHeight(dbheight + 1)
	if err != nil || nextABlock == nil {
		pl := s.ProcessLists.GetSafe(dbheight + 1)
		if pl == nil {
			dbkl, err := s.DB.FetchDBlockByHeight(dbheight)
			if err != nil || dbkl == nil {
				return nil, fmt.Errorf("Do not have signatures at height %d to validate DBStateMsg", dbheight)
			}
		} else {
			for _, dbsig := range pl.DBSignatures {
				allSigs = append(allSigs, dbsig.Signature)
			}
		}
	} else {
		abEntries := nextABlock.GetABEntries()
		for _, adminEntry := range abEntries {
			data, err := adminEntry.MarshalBinary()
			if err != nil {
				return nil, err
			}
			switch adminEntry.Type() {
			case constants.TYPE_DB_SIGNATURE:
				r := new(adminBlock.DBSignatureEntry)
				err := r.UnmarshalBinary(data)
				if err != nil {
					continue
				}

				blockSig := new(primitives.Signature)
				blockSig.SetSignature(r.PrevDBSig.Bytes())
				blockSig.SetPub(r.PrevDBSig.GetKey())
				allSigs = append(allSigs, blockSig)
			}
		}
	}
	msg := messages.NewDBStateMsg(s.GetTimestamp(), dblk, ablk, fblk, ecblk, eBlocks, entries, allSigs)
	msg.(*messages.DBStateMsg).IsInDB = true

	return msg, nil
}

func (s *State) LoadDataByHash(requestedHash interfaces.IHash) (interfaces.BinaryMarshallable, int, error) {
	if requestedHash == nil {
		return nil, -1, fmt.Errorf("%s", "Requested hash must be non-empty")
	}

	var result interfaces.BinaryMarshallable
	var err error

	// Check for Entry
	result, err = s.DB.FetchEntry(requestedHash)
	if result != nil && err == nil {
		return result, 0, nil
	}

	// Check for Entry Block
	result, err = s.DB.FetchEBlock(requestedHash)
	if result != nil && err == nil {
		return result, 1, nil
	}

	return nil, -1, nil
}

func (s *State) LoadSpecificMsg(dbheight uint32, vm int, plistheight uint32) (interfaces.IMsg, error) {
	msg, _, err := s.LoadSpecificMsgAndAck(dbheight, vm, plistheight)
	return msg, err
}

func (s *State) LoadSpecificMsgAndAck(dbheight uint32, vmIndex int, plistheight uint32) (interfaces.IMsg, interfaces.IMsg, error) {
	pl := s.ProcessLists.Get(dbheight)
	if pl == nil {
		return nil, nil, fmt.Errorf("%s", "Nil Process List")
	}
	if vmIndex < 0 || vmIndex >= len(pl.VMs) {
		return nil, nil, fmt.Errorf("%s", "VM index out of range")
	}
	vm := pl.VMs[vmIndex]

	if plistheight < 0 || int(plistheight) >= len(vm.List) {
		return nil, nil, fmt.Errorf("%s", "Process List too small (lacks requested msg)")
	}

	msg := vm.List[plistheight]
	ackMsg := vm.ListAck[plistheight]

	if msg == nil || ackMsg == nil {
		return nil, nil, fmt.Errorf("%s", "State process list does not include requested message/ack")
	}
	return msg, ackMsg, nil
}

func (s *State) LoadHoldingMap() map[[32]byte]interfaces.IMsg {
	// request holding queue from state from outside state scope
	s.HoldingMutex.RLock()
	defer s.HoldingMutex.RUnlock()
	localMap := s.HoldingMap

	return localMap
}

// this is executed in the state maintenance processes where the holding queue is in scope and can be queried
//  This is what fills the HoldingMap while locking it against a read while building
func (s *State) fillHoldingMap() {
	// once a second is often enough to rebuild the Ack list exposed to api

	if s.HoldingLast < time.Now().Unix() {

		localMap := make(map[[32]byte]interfaces.IMsg)
		for i, msg := range s.Holding {
			localMap[i] = msg
		}
		s.HoldingLast = time.Now().Unix()
		s.HoldingMutex.Lock()
		defer s.HoldingMutex.Unlock()
		s.HoldingMap = localMap

	}
}

// this is called from the APIs that do not have access directly to the Acks.  State makes a copy and puts it in AcksMap
func (s *State) LoadAcksMap() map[[32]byte]interfaces.IMsg {
	// request Acks queue from state from outside state scope
	s.AcksMutex.RLock()
	defer s.AcksMutex.RUnlock()
	localMap := s.AcksMap

	return localMap

}

// this is executed in the state maintenance processes where the Acks queue is in scope and can be queried
//  This is what fills the AcksMap requested in LoadAcksMap
func (s *State) fillAcksMap() {
	// once a second is often enough to rebuild the Ack list exposed to api
	if s.AcksLast < time.Now().Unix() {
		localMap := make(map[[32]byte]interfaces.IMsg)
		for i, msg := range s.Acks {
			localMap[i] = msg
		}
		s.AcksLast = time.Now().Unix()
		s.AcksMutex.Lock()
		defer s.AcksMutex.Unlock()
		s.AcksMap = localMap
	}
}

func (s *State) GetPendingEntries(params interface{}) []interfaces.IPendingEntry {
	resp := make([]interfaces.IPendingEntry, 0)
	repeatmap := make(map[[32]byte]interfaces.IPendingEntry)
	pls := s.ProcessLists.Lists
	var cc messages.CommitChainMsg
	var ce messages.CommitEntryMsg
	var re messages.RevealEntryMsg
	var tmp interfaces.IPendingEntry
	LastComplete := s.GetDBHeightComplete()
	// check all existing processlists/VMs
	for _, pl := range pls {
		if pl != nil {
			if pl.DBHeight > LastComplete {
				for _, v := range pl.VMs {
					for _, plmsg := range v.List {
						if plmsg.Type() == constants.COMMIT_CHAIN_MSG { //5
							enb, err := plmsg.MarshalBinary()
							if err != nil {
								return nil
							}
							err = cc.UnmarshalBinary(enb)
							if err != nil {
								return nil
							}
							tmp.EntryHash = cc.CommitChain.EntryHash

							tmp.ChainID = cc.CommitChain.ChainIDHash
							if pl.DBHeight > s.GetDBHeightComplete() {
								tmp.Status = constants.AckStatusACKString
							} else {
								tmp.Status = constants.AckStatusDBlockConfirmedString
							}
							if _, ok := repeatmap[tmp.EntryHash.Fixed()]; !ok {
								resp = append(resp, tmp)
								repeatmap[tmp.EntryHash.Fixed()] = tmp
							}
						} else if plmsg.Type() == constants.COMMIT_ENTRY_MSG { //6
							enb, err := plmsg.MarshalBinary()
							if err != nil {
								return nil
							}
							err = ce.UnmarshalBinary(enb)
							if err != nil {
								return nil
							}
							tmp.EntryHash = ce.CommitEntry.EntryHash

							tmp.ChainID = nil
							if pl.DBHeight > s.GetDBHeightComplete() {
								tmp.Status = constants.AckStatusACKString
							} else {
								tmp.Status = constants.AckStatusDBlockConfirmedString
							}

							if _, ok := repeatmap[tmp.EntryHash.Fixed()]; !ok {
								resp = append(resp, tmp)
								repeatmap[tmp.EntryHash.Fixed()] = tmp
							}
						} else if plmsg.Type() == constants.REVEAL_ENTRY_MSG { //13
							enb, err := plmsg.MarshalBinary()
							if err != nil {
								return nil
							}
							err = re.UnmarshalBinary(enb)
							if err != nil {
								return nil
							}
							tmp.EntryHash = re.Entry.GetHash()
							tmp.ChainID = re.Entry.GetChainID()
							if pl.DBHeight > s.GetDBHeightComplete() {
								tmp.Status = constants.AckStatusACKString
							} else {
								tmp.Status = constants.AckStatusDBlockConfirmedString
							}
							if _, ok := repeatmap[tmp.EntryHash.Fixed()]; !ok {
								resp = append(resp, tmp)
								repeatmap[tmp.EntryHash.Fixed()] = tmp
							} else {
								//If it is in there, it may not know the chainid because it was from a commit
								if repeatmap[tmp.EntryHash.Fixed()].ChainID == nil {
									repeatmap[tmp.EntryHash.Fixed()] = tmp
									// now update your response entry
									for k, _ := range resp {
										if resp[k].EntryHash.IsSameAs(tmp.EntryHash) {
											if tmp.ChainID != nil {

												resp[k].ChainID = tmp.ChainID
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// check holding queue
	q := s.LoadHoldingMap()
	for _, h := range q {
		if h.Type() == constants.REVEAL_ENTRY_MSG {
			enb, err := h.MarshalBinary()
			if err != nil {
				return nil
			}
			err = re.UnmarshalBinary(enb)
			if err != nil {
				return nil
			}
			tmp.EntryHash = re.Entry.GetHash()

			tmp.ChainID = re.Entry.GetChainID()

			tmp.Status = constants.AckStatusNotConfirmedString

			if _, ok := repeatmap[tmp.EntryHash.Fixed()]; !ok {

				resp = append(resp, tmp)
				repeatmap[tmp.EntryHash.Fixed()] = tmp
			}

		}
	}

	return resp
}

func (s *State) GetPendingTransactions(params interface{}) []interfaces.IPendingTransaction {
	var flgFound bool

	var currentHeightComplete = s.GetDBHeightComplete()
	resp := make([]interfaces.IPendingTransaction, 0)
	pls := s.ProcessLists.Lists
	for _, pl := range pls {
		if pl != nil {
			// ignore old process lists
			if pl.DBHeight > currentHeightComplete {
				cb := pl.State.FactoidState.GetCurrentBlock()
				ct := cb.GetTransactions()
				for _, tran := range ct {
					var tmp interfaces.IPendingTransaction
					tmp.TransactionID = tran.GetSigHash()
					if tran.GetBlockHeight() > 0 {
						tmp.Status = constants.AckStatusDBlockConfirmedString
					} else {
						tmp.Status = constants.AckStatusACKString
					}

					tmp.Inputs = tran.GetInputs()
					tmp.Outputs = tran.GetOutputs()
					tmp.ECOutputs = tran.GetECOutputs()
					ecrate := s.GetPredictiveFER()
					ecrate, _ = tran.CalculateFee(ecrate)
					tmp.Fees = ecrate

					if params.(string) == "" {
						flgFound = true
					} else {
						flgFound = tran.HasUserAddress(params.(string))
					}
					if flgFound == true {
						//working through multiple process lists.  Is this transaction already in the list?
						for _, pt := range resp {
							if pt.TransactionID.String() == tmp.TransactionID.String() {
								flgFound = false
							}
						}
						//  flag was true to be added to the list and not already in the list
						if flgFound == true {
							resp = append(resp, tmp)
						}
					}
				}
			}
		}
	}

	q := s.LoadHoldingMap()
	for _, h := range q {
		if h.Type() == constants.FACTOID_TRANSACTION_MSG {
			var rm messages.FactoidTransaction
			enb, err := h.MarshalBinary()
			if err != nil {
				return nil
			}
			err = rm.UnmarshalBinary(enb)
			if err != nil {
				return nil
			}
			tempTran := rm.GetTransaction()
			var tmp interfaces.IPendingTransaction
			tmp.TransactionID = tempTran.GetSigHash()
			tmp.Status = constants.AckStatusNotConfirmedString
			flgFound = tempTran.HasUserAddress(params.(string))
			tmp.Inputs = tempTran.GetInputs()
			tmp.Outputs = tempTran.GetOutputs()
			tmp.ECOutputs = tempTran.GetECOutputs()
			ecrate := s.GetPredictiveFER()
			ecrate, _ = tempTran.CalculateFee(ecrate)
			tmp.Fees = ecrate
			if flgFound == true {
				//working through multiple process lists.  Is this transaction already in the list?
				for _, pt := range resp {
					if pt.TransactionID.IsSameAs(tmp.TransactionID) {
						flgFound = false
					}
				}
				//  flag was true to be added to the list and not already in the list
				if flgFound == true {
					resp = append(resp, tmp)
				}
			}
		}
	}

	//b, _ := json.Marshal(resp)
	return resp
}

// might want to make this search the database at some point to be more generic
func (s *State) FetchEntryHashFromProcessListsByTxID(txID string) (interfaces.IHash, error) {
	pls := s.ProcessLists.Lists
	var cc messages.CommitChainMsg
	var ce messages.CommitEntryMsg
	var re messages.RevealEntryMsg

	// check all existing processlists (last complete block +1 and greater)
	for _, pl := range pls {
		if pl != nil {
			for _, v := range pl.VMs {
				if v != nil {
					// check chain commits
					for _, plmsg := range v.List {
						if plmsg != nil {
							//	if plmsg.Type() != nil {
							if plmsg.Type() == constants.COMMIT_CHAIN_MSG { //5 other types could be in this VM
								enb, err := plmsg.MarshalBinary()
								if err != nil {
									return nil, err
								}
								err = cc.UnmarshalBinary(enb)
								if err != nil {
									return nil, err
								}
								if cc.CommitChain.GetSigHash().String() == txID {
									return cc.CommitChain.EntryHash, nil
								}
							} else if plmsg.Type() == constants.COMMIT_ENTRY_MSG { //6

								enb, err := plmsg.MarshalBinary()
								if err != nil {
									return nil, err
								}
								err = ce.UnmarshalBinary(enb)
								if err != nil {
									return nil, err
								}

								if ce.CommitEntry.GetSigHash().String() == txID {
									return ce.CommitEntry.EntryHash, nil
								}

							} else if plmsg.Type() == constants.REVEAL_ENTRY_MSG { //13
								enb, err := plmsg.MarshalBinary()
								if err != nil {
									return nil, err
								}
								err = re.UnmarshalBinary(enb)
								if err != nil {
									return nil, err
								}
								if re.Entry.GetHash().String() == txID {
									return re.Entry.GetHash(), nil
								}
							}
						}
						//	} else {
						//		return nil, fmt.Errorf("%s", "Invalid Message in Holding Queue")
						//	}
					}
				}
			}
		}
	}
	return nil, fmt.Errorf("%s", "Transaction not found")
}

func (s *State) IncFactoidTrans() {
	s.FactoidTrans++
}

func (s *State) IncEntryChains() {
	s.NewEntryChains++
}

func (s *State) IncEntries() {
	s.NewEntries++
}

func (s *State) IsStalled() bool {
	if s.CurrentMinuteStartTime == 0 { //0 while syncing.
		return false
	}

	// If we are under height 3, then we won't say stalled by height.
	lh := s.GetTrueLeaderHeight()

	if lh >= 3 && s.GetHighestSavedBlk() < lh-3 {
		return true
	}

	//use 1/10 of the block time times 1.5 in seconds as a timeout on the 'minutes'
	var stalltime float64

	stalltime = float64(int64(s.GetDirectoryBlockInSeconds())) / 10
	stalltime = stalltime * 1.5 * 1e9
	//fmt.Println("STALL 2", s.CurrentMinuteStartTime/1e9, time.Now().UnixNano()/1e9, stalltime/1e9, (float64(time.Now().UnixNano())-stalltime)/1e9)

	if float64(s.CurrentMinuteStartTime) < float64(time.Now().UnixNano())-stalltime { //-90 seconds was arbitrary
		return true
	}

	return false
}

func (s *State) DatabaseContains(hash interfaces.IHash) bool {
	result, _, err := s.LoadDataByHash(hash)
	if result != nil && err == nil {
		return true
	}
	return false
}

func (s *State) GetLeaderVM() int {
	return s.LeaderVMIndex
}

func (s *State) GetDBState(height uint32) *DBState {
	return s.DBStates.Get(int(height))
}

// Return the Directory block if it is in memory, or hit the database if it must
// be loaded.
func (s *State) GetDirectoryBlockByHeight(height uint32) interfaces.IDirectoryBlock {
	dbstate := s.DBStates.Get(int(height))
	if dbstate != nil {
		return dbstate.DirectoryBlock
	}
	dblk, err := s.DB.FetchDBlockByHeight(height)
	if err != nil {
		return nil
	}
	return dblk
}

func (s *State) UpdateState() (progress bool) {
	s.StateUpdateState++
	dbheight := s.LLeaderHeight
	//	s.GetHighestSavedBlk()
	plbase := s.ProcessLists.DBHeightBase

	ProcessLists := s.ProcessLists
	if ProcessLists.SetString {
		ProcessLists.SetString = false
		ProcessLists.Str = ProcessLists.String()
	}

	if plbase <= dbheight { // TODO: This is where we have to fix the fact that syncing with dbstates can fail to transition to inMessages
		if !s.Leader || s.RunLeader {
			progress = ProcessLists.UpdateState(dbheight)
		}
	}

	s.SetString()

	// Update our TPS every ~ 3 seconds at the earliest
	if s.lasttime.Before(time.Now().Add(-3 * time.Second)) {
		s.CalculateTransactionRate()
	}

	// check to see if a holding queue list request has been made
	s.fillHoldingMap()
	s.fillAcksMap()

entryHashProcessing:
	for {
		select {
		case e := <-s.UpdateEntryHash:
			// Save the entry hash, and remove from commits IF this hash is valid in this current timeframe.
			if s.Replay.SetHashNow(constants.REVEAL_REPLAY, e.Hash.Fixed(), e.Timestamp) {
				// If the SetHashNow worked, then we should prohibit any commit that might be pending.
				// Remove any commit that might be around.
				s.Commits.Delete(e.Hash.Fixed())
			}
		default:
			break entryHashProcessing
		}
	}
	return
}

// Returns true if this hash exists nowhere in the Replay structures.  Returns False if we
// have already seen this hash before.  Replay is NOT updated yet.
func (s *State) NoEntryYet(entryhash interfaces.IHash, ts interfaces.Timestamp) bool {
	unique := s.Replay.IsHashUnique(constants.REVEAL_REPLAY, entryhash.Fixed())
	return unique
}

func (s *State) AddDBSig(dbheight uint32, chainID interfaces.IHash, sig interfaces.IFullSignature) {
	s.ProcessLists.Get(dbheight).AddDBSig(chainID, sig)
}

func (s *State) AddFedServer(dbheight uint32, hash interfaces.IHash) int {
	//s.AddStatus(fmt.Sprintf("AddFedServer %x at dbht: %d", hash.Bytes()[2:6], dbheight))
	s.LogPrintf("executeMsg", "AddServer (Federated): ChainID: %x at dbht: %d From %s", hash.Bytes()[3:6], dbheight, atomic.WhereAmIString(1))
	return s.ProcessLists.Get(dbheight).AddFedServer(hash)
}

func (s *State) TrimVMList(dbheight uint32, height uint32, vmIndex int) {
	s.ProcessLists.Get(dbheight).TrimVMList(height, vmIndex)
}

func (s *State) RemoveFedServer(dbheight uint32, hash interfaces.IHash) {
	//s.AddStatus(fmt.Sprintf("RemoveFedServer %x at dbht: %d", hash.Bytes()[2:6], dbheight))
	s.LogPrintf("executeMsg", "RemoveServer (Federated): ChainID: %x at dbht: %d", hash.Bytes()[3:6], dbheight)
	s.ProcessLists.Get(dbheight).RemoveFedServerHash(hash)
}

func (s *State) AddAuditServer(dbheight uint32, hash interfaces.IHash) int {
	//s.AddStatus(fmt.Sprintf("AddAuditServer %x at dbht: %d", hash.Bytes()[2:6], dbheight))
	s.LogPrintf("executeMsg", "AddServer (Audit): ChainID: %x at dbht: %d", hash.Bytes()[3:6], dbheight)
	return s.ProcessLists.Get(dbheight).AddAuditServer(hash)
}

func (s *State) RemoveAuditServer(dbheight uint32, hash interfaces.IHash) {
	//s.AddStatus(fmt.Sprintf("RemoveAuditServer %x at dbht: %d", hash.Bytes()[2:6], dbheight))
	s.LogPrintf("executeMsg", "RemoveServer (Audit): ChainID: %x at dbht: %d", hash.Bytes()[3:6], dbheight)
	s.ProcessLists.Get(dbheight).RemoveAuditServerHash(hash)
}

func (s *State) GetFedServers(dbheight uint32) []interfaces.IServer {
	pl := s.ProcessLists.Get(dbheight)
	if pl != nil {
		return pl.FedServers
	}
	return nil
}

func (s *State) GetAuditServers(dbheight uint32) []interfaces.IServer {
	pl := s.ProcessLists.Get(dbheight)
	if pl != nil {
		return pl.AuditServers
	}
	return nil
}

func (s *State) GetOnlineAuditServers(dbheight uint32) []interfaces.IServer {
	allAuditServers := s.GetAuditServers(dbheight)
	var onlineAuditServers []interfaces.IServer

	for _, server := range allAuditServers {
		if server.IsOnline() {
			onlineAuditServers = append(onlineAuditServers, server)
		}
	}
	return onlineAuditServers
}

func (s *State) IsLeader() bool {
	return s.Leader
}

func (s *State) GetVirtualServers(dbheight uint32, minute int, identityChainID interfaces.IHash) (bool, int) {
	pl := s.ProcessLists.Get(dbheight)
	return pl.GetVirtualServers(minute, identityChainID)
}

func (s *State) GetFactoshisPerEC() uint64 {
	return s.FactoshisPerEC
}

func (s *State) SetFactoshisPerEC(factoshisPerEC uint64) {
	s.FactoshisPerEC = factoshisPerEC
}

func (s *State) GetIdentityChainID() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("State.GetIdentityChainID() saw an interface that was nil")
		}
	}()
	return s.IdentityChainID
}

func (s *State) SetIdentityChainID(chainID interfaces.IHash) {
	if !s.IdentityChainID.IsSameAs(chainID) {
		s.LogPrintf("AckChange", "SetIdentityChainID %v", chainID.String())
	}
	s.IdentityChainID = chainID
}

func (s *State) GetMinuteDuration() time.Duration {
	return time.Duration(s.DirectoryBlockInSeconds) * time.Second / 10
}

func (s *State) GetDirectoryBlockInSeconds() int {
	return s.DirectoryBlockInSeconds
}

func (s *State) SetDirectoryBlockInSeconds(t int) {
	s.DirectoryBlockInSeconds = t
}

func (s *State) GetServerPrivateKey() *primitives.PrivateKey {
	return s.ServerPrivKey
}

func (s *State) GetServerPublicKey() *primitives.PublicKey {
	return s.ServerPubKey
}

func (s *State) GetServerPublicKeyString() string {
	return s.ServerPubKey.String()
}

func (s *State) GetAnchor() interfaces.IAnchor {
	return s.Anchor
}

func (s *State) TallySent(msgType int) {
	s.MessageTalliesSent[msgType]++
}

func (s *State) TallyReceived(msgType int) {
	s.MessageTalliesReceived[msgType]++
}

func (s *State) GetMessageTalliesSent(i int) int {
	return s.MessageTalliesSent[i]
}

func (s *State) GetMessageTalliesReceived(i int) int {
	return s.MessageTalliesReceived[i]
}

func (s *State) GetFactomdVersion() string {
	return s.FactomdVersion
}

func (s *State) initServerKeys() {
	var err error
	s.ServerPrivKey, err = primitives.NewPrivateKeyFromHex(s.LocalServerPrivKey)
	if err != nil {
		//panic("Cannot parse Server Private Key from configuration file: " + err.Error())
	}
	s.ServerPubKey = s.ServerPrivKey.Pub
}

func (s *State) GetAuditHeartBeats() []interfaces.IMsg {
	return s.AuditHeartBeats
}

// Returns a millisecond timestamp
func (s *State) GetTimestamp() interfaces.Timestamp {
	return primitives.NewTimestampNow()
}

func (s *State) GetTimeOffset() interfaces.Timestamp {
	return s.TimeOffset
}

func (s *State) Sign(b []byte) interfaces.IFullSignature {
	return s.ServerPrivKey.Sign(b)
}

func (s *State) GetFactoidState() interfaces.IFactoidState {
	return s.FactoidState
}

func (s *State) SetFactoidState(dbheight uint32, fs interfaces.IFactoidState) {
	s.FactoidState = fs
}

// Allow us the ability to update the port number at run time....
func (s *State) SetPort(port int) {
	s.PortNumber = port
}

func (s *State) GetPort() int {
	return s.PortNumber
}

func (s *State) TickerQueue() chan int {
	if EnableLeaderThread {
		panic("LeaderThreadEnabled")
	}
	return s.tickerQueue
}

func (s *State) TimerMsgQueue() chan interfaces.IMsg {
	return s.timerMsgQueue
}

func (s *State) NetworkInvalidMsgQueue() chan interfaces.IMsg {
	return s.networkInvalidMsgQueue
}

func (s *State) NetworkOutMsgQueue() interfaces.IQueue {
	return s.networkOutMsgQueue
}

func (s *State) InMsgQueue() interfaces.IQueue {
	return s.inMsgQueue
}

func (s *State) InMsgQueue2() interfaces.IQueue {
	return s.inMsgQueue2
}

func (s *State) ElectionsQueue() interfaces.IQueue {
	return s.electionsQueue
}

func (s *State) APIQueue() interfaces.IQueue {
	return s.apiQueue
}

func (s *State) AckQueue() chan interfaces.IMsg {
	return s.ackQueue
}

func (s *State) MsgQueue() chan interfaces.IMsg {
	return s.msgQueue
}

func (s *State) PrioritizedMsgQueue() chan interfaces.IMsg {
	return s.prioritizedMsgQueue
}

func (s *State) GetLeaderTimestamp() interfaces.Timestamp {
	if s.LeaderTimestamp == nil {
		// To leader timestamp?  Then use the boot time less a minute
		s.SetLeaderTimestamp(primitives.NewTimestampFromMilliseconds(s.TimestampAtBoot.GetTimeMilliUInt64() - 60*1000))
	}
	return primitives.NewTimestampFromMilliseconds(s.LeaderTimestamp.GetTimeMilliUInt64())
}

func (s *State) GetMessageFilterTimestamp() interfaces.Timestamp {
	if s.messageFilterTimestamp == nil {
		s.messageFilterTimestamp = primitives.NewTimestampNow()
	}
	return primitives.NewTimestampFromMilliseconds(s.messageFilterTimestamp.GetTimeMilliUInt64())
}

// the MessageFilterTimestamp  is used to filter inMessages from the past or before the replay filter.
// We will not set it to a time that is before (20 minutes before) boot or more than one hour in the past.
// this ensure inMessages from prior boot and inMessages that predate the current replay filter are
// are dropped.
// It marks the start of the replay filter content
func (s *State) SetMessageFilterTimestamp(leaderTS interfaces.Timestamp) {

	// make a copy of the time stamp so we don't change the source
	requestedTS := new(primitives.Timestamp)
	requestedTS.SetTimestamp(leaderTS)

	onehourago := new(primitives.Timestamp)
	onehourago.SetTimeMilli(primitives.NewTimestampNow().GetTimeMilli() - 60*60*1000) // now() - one hour

	if requestedTS.GetTimeMilli() < onehourago.GetTimeMilli() {
		requestedTS.SetTimestamp(onehourago)
	}

	// build a timestamp 20 minutes before boot so we will accept inMessages from nodes who booted before us.
	preBootTime := new(primitives.Timestamp)
	preBootTime.SetTimeMilli(s.TimestampAtBoot.GetTimeMilli() - 20*60*1000)

	if requestedTS.GetTimeMilli() < preBootTime.GetTimeMilli() {
		requestedTS.SetTimestamp(preBootTime)
	}

	if s.messageFilterTimestamp != nil && requestedTS.GetTimeMilli() < s.messageFilterTimestamp.GetTimeMilli() {
		s.LogPrintf("executeMsg", "Set MessageFilterTimestamp attempt to move backward in time from %s", atomic.WhereAmIString(1))
		return
	}

	s.LogPrintf("executeMsg", "SetMessageFilterTimestamp(%s) using %s ", leaderTS, requestedTS.String())
	s.messageFilterTimestamp = primitives.NewTimestampFromMilliseconds(requestedTS.GetTimeMilliUInt64())
}

func (s *State) GotHeartbeat(heartbeatTS interfaces.Timestamp, dbheight uint32) {

	if !s.DBFinished {
		return
	}

	newTS := heartbeatTS.GetTimeMilli()
	leaderTime := s.GetLeaderTimestamp().GetTimeMilli()

	if leaderTime > newTS {
		newTS = leaderTime
	}

	s.LogPrintf("executeMsg", "GotHeartbeat(%s, dbht:"+
		"%d)", heartbeatTS, dbheight)

	// set filter to one hour before target
	s.SetMessageFilterTimestamp(primitives.NewTimestampFromMilliseconds(uint64(newTS - 60*60*1000)))

	// set Highest Ack & Known blocks
	s.SetHighestAck(dbheight)
	s.SetHighestKnownBlock(dbheight)

	// re-center replay filter
	s.Replay.Recenter(primitives.NewTimestampFromMilliseconds(uint64(newTS)))
}

func (s *State) SetLeaderTimestamp(ts interfaces.Timestamp) {
	s.LogPrintf("executeMsg", "SetLeaderTimestamp(%s)", ts.String())

	s.LeaderTimestamp = primitives.NewTimestampFromMilliseconds(ts.GetTimeMilliUInt64())
	s.SetMessageFilterTimestamp(primitives.NewTimestampFromMilliseconds(ts.GetTimeMilliUInt64() - 60*60*1000)) // set message filter to one hour before this block started.

	if s.leaderTimestampPub != nil {
		s.leaderTimestampPub.Write(s.LeaderTimestamp)
	}
}

func (s *State) SetFaultTimeout(timeout int) {
	s.FaultTimeout = timeout
}

func (s *State) SetFaultWait(wait int) {
	s.FaultWait = wait
}

func (s *State) GetElections() interfaces.IElections {
	return s.Elections
}

// GetAuthorities will return a list of the network authorities
func (s *State) GetAuthorities() []interfaces.IAuthority {
	return s.IdentityControl.GetAuthorities()
}

// GetAuthorityInterface will the authority as an interface. Because of import issues
// we cannot access IdentityControl Directly
func (s *State) GetAuthorityInterface(chainid interfaces.IHash) interfaces.IAuthority {
	rval := s.IdentityControl.GetAuthority(chainid)
	if rval == nil {
		return nil
	}
	return rval
}

// GetLeaderPL returns the leader process list from the state. this method is
// for debugging and should not be called in normal production code.
func (s *State) GetLeaderPL() interfaces.IProcessList {
	return s.LeaderPL
}

// Getting the cfg state for Factom doesn't force a read of the config file unless
// it hasn't been read yet.
func (s *State) GetCfg() interfaces.IFactomConfig {
	return s.Cfg
}

// ReadCfg forces a read of the Factom config file.  However, it does not change the
// state of any cfg object held by other processes... Only what will be returned by
// future calls to Cfg().(s.Cfg.(*util.FactomdConfig)).String()
func (s *State) ReadCfg(filename string) interfaces.IFactomConfig {
	s.Cfg = util.ReadConfig(filename)
	return s.Cfg
}

func (s *State) GetNetworkNumber() int {
	return s.NetworkNumber
}

func (s *State) GetNetworkName() string {
	switch s.NetworkNumber {
	case constants.NETWORK_MAIN:
		return "MAIN"
	case constants.NETWORK_TEST:
		return "TEST"
	case constants.NETWORK_LOCAL:
		return "LOCAL"
	case constants.NETWORK_CUSTOM:
		return "CUSTOM"
	}
	return "" // Shouldn't ever get here
}

func (s *State) GetNetworkID() uint32 {
	switch s.NetworkNumber {
	case constants.NETWORK_MAIN:
		return constants.MAIN_NETWORK_ID
	case constants.NETWORK_TEST:
		return constants.TEST_NETWORK_ID
	case constants.NETWORK_LOCAL:
		return constants.LOCAL_NETWORK_ID
	case constants.NETWORK_CUSTOM:
		return binary.BigEndian.Uint32(s.CustomNetworkID)
	}
	return uint32(0)
}

// The initial public key that can sign the first block
func (s *State) GetNetworkBootStrapKey() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("State.GetNetworkBootStrapKey() saw an interface that was nil")
		}
	}()
	switch s.NetworkNumber {
	case constants.NETWORK_MAIN:
		key, _ := primitives.HexToHash("0426a802617848d4d16d87830fc521f4d136bb2d0c352850919c2679f189613a")
		return key
	case constants.NETWORK_TEST:
		key, _ := primitives.HexToHash("49b6edd274e7d07c94d4831eca2f073c207248bde1bf989d2183a8cebca227b7")
		return key
	case constants.NETWORK_LOCAL:
		key, _ := primitives.HexToHash("cc1985cdfae4e32b5a454dfda8ce5e1361558482684f3367649c3ad852c8e31a")
		return key
	case constants.NETWORK_CUSTOM:
		key, err := primitives.HexToHash(s.CustomBootstrapKey)
		if err != nil {
			panic(fmt.Sprintf("Cannot use a CUSTOM network without a CustomBootstrapKey specified in the factomd.conf file. Err: %s", err.Error()))
		}
		return key
	}
	return primitives.NewZeroHash()
}

// The initial identity that can sign the first block
func (s *State) GetNetworkBootStrapIdentity() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("State.GetNetworkBootStrapIdentity() saw an interface that was nil")
		}
	}()
	switch s.NetworkNumber {
	case constants.NETWORK_MAIN:
		return primitives.NewZeroHash()
	case constants.NETWORK_TEST:
		return primitives.NewZeroHash()
	case constants.NETWORK_LOCAL:
		id, _ := primitives.HexToHash("38bab1455b7bd7e5efd15c53c777c79d0c988e9210f1da49a99d95b3a6417be9")
		return id
	case constants.NETWORK_CUSTOM:
		id, err := primitives.HexToHash(s.CustomBootstrapIdentity)
		if err != nil {
			panic(fmt.Sprintf("Cannot use a CUSTOM network without a CustomBootstrapIdentity specified in the factomd.conf file. Err: %s", err.Error()))
		}
		return id
	}
	return primitives.NewZeroHash()
}

// The identity for validating inMessages
func (s *State) GetNetworkSkeletonIdentity() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("State.GetNetworkSkeletonIdentity() saw an interface that was nil")
		}
	}()
	switch s.NetworkNumber {
	case constants.NETWORK_MAIN:
		id, _ := primitives.HexToHash("8888882690706d0d45d49538e64e7c76571d9a9b331256b5b69d9fd2d7f1f14a")
		return id
	case constants.NETWORK_TEST:
		id, _ := primitives.HexToHash("8888888888888888888888888888888888888888888888888888888888888888")
		return id
	case constants.NETWORK_LOCAL:
		id, _ := primitives.HexToHash("8888888888888888888888888888888888888888888888888888888888888888")
		return id
	case constants.NETWORK_CUSTOM:
		id, _ := primitives.HexToHash("88888816d408cd0d7b1b28760f3371a40e98dc2e985c28410e781935954afdf3")
		return id
	}
	id, _ := primitives.HexToHash("8888888888888888888888888888888888888888888888888888888888888888")
	return id
}

func (s *State) GetNetworkIdentityRegistrationChain() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("State.GetNetworkIdentityRegistrationChain() saw an interface that was nil")
		}
	}()
	id, _ := primitives.HexToHash("888888001750ede0eff4b05f0c3f557890b256450cabbb84cada937f9c258327")
	return id
}

func (s *State) GetMatryoshka(dbheight uint32) (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("State.GetMatryoshka saw an interface thatwas nil")
		}
	}()
	return nil
}

func (s *State) InitLevelDB() error {
	if s.DB != nil {
		return nil
	}

	path := s.LdbPath + "/" + s.Network + "/" + "factoid_level.db"

	s.Println("Database:", path)
	fmt.Fprintln(os.Stderr, "Database:", path)

	dbase, err := leveldb.NewLevelDB(path, false)

	if err != nil || dbase == nil {
		dbase, err = leveldb.NewLevelDB(path, true)
		if err != nil {
			return err
		}
	}

	s.DB = databaseOverlay.NewOverlay(dbase)
	return nil
}

func (s *State) InitBoltDB() error {
	if s.DB != nil {
		return nil
	}

	path := s.BoltDBPath + "/" + s.Network + "/"

	s.Println("Database Path for", s.FactomNodeName, "is", path)
	os.MkdirAll(path, 0777)

	dbase := new(boltdb.BoltDB)
	dbase.Init(nil, path+"FactomBolt.db")
	s.DB = databaseOverlay.NewOverlay(dbase)
	return nil
}

func (s *State) InitMapDB() error {
	if s.DB != nil {
		return nil
	}

	dbase := new(mapdb.MapDB)
	dbase.Init(nil)
	s.DB = databaseOverlay.NewOverlay(dbase)
	return nil
}

func (s *State) String() string {
	str := "\n===============================================================\n" + s.serverPrt
	str = fmt.Sprintf("\n%s\n  Leader Height: %d\n", str, s.LLeaderHeight)
	str = str + "===============================================================\n"
	return str
}

func (s *State) ShortString() string {
	return s.serverPrt
}

func (s *State) SetString() {
	switch s.Status {
	case 0:
		return
	case 1:
		s.SetStringQueues()
	case 2:
		s.SetStringConsensus()
	}

	s.Status = 0

}

func (s *State) SummaryHeader() string {
	ht := s.GetHighestSavedBlk()
	dbstate := s.DBStates.Get(int(ht))
	sum := ""
	if dbstate != nil {
		sum = fmt.Sprintf("Ht: %d New Chains: %d New Entries: %d sum: %d Total Entries: %d diff %d Total EBs: %d FCT: %d",
			ht,
			s.NumNewChains,
			s.NumNewEntries,
			s.NumNewEntries+s.NumNewChains,
			s.NumEntries,
			s.NumEntries-s.NumNewEntries-s.NumNewChains,
			s.NumEntryBlocks,
			s.NumFCTTrans)
	}

	str := fmt.Sprintf(" %s\n %10s %6s %12s %5s %4s %6s %10s %8s %5s %4s %20s %14s %10s %-8s %-9s %16s %9s %9s %s\n",
		sum,
		"Node",
		"ID   ",
		" ",
		"Resets",
		"Drop",
		"Delay",
		"DB ",
		"PL  ",
		" ",
		"Min",
		"DBState(ask/rply/drop/apply)",
		"Msg",
		"   Resend",
		"Expire",
		"Fct/EC/E",
		"API:Fct/EC/E",
		"tps t/i",
		"DBH-:-M",
		"BH")

	return str
}

func (s *State) SetStringConsensus() {
	str := fmt.Sprintf("%10s[%x_%x] ", s.FactomNodeName, s.IdentityChainID.Bytes()[:2], s.IdentityChainID.Bytes()[3:6])

	s.serverPrt = str
}

// CalculateTransactionRate calculates how many transactions this node is processing
//		totalTPS	: Transaction rate over life of node (totaltime / totaltrans)
//		instantTPS	: Transaction rate weighted over last 3 seconds
func (s *State) CalculateTransactionRate() (totalTPS float64, instantTPS float64) {
	runtime := time.Since(s.Starttime)
	total := s.FactoidTrans + s.NewEntryChains + s.NewEntries
	tps := float64(total) / float64(runtime.Seconds())
	TotalTransactionPerSecond.Set(tps) // Prometheus
	shorttime := time.Since(s.lasttime)
	if shorttime >= time.Second*3 {
		delta := (s.FactoidTrans + s.NewEntryChains + s.NewEntries) - s.transCnt
		s.tps = ((float64(delta) / float64(shorttime.Seconds())) + 2*s.tps) / 3
		s.lasttime = time.Now()
		s.transCnt = total                     // transactions accounted for
		InstantTransactionPerSecond.Set(s.tps) // Prometheus
	}

	return tps, s.tps
}

func (s *State) SetStringQueues() {
	vmi := -1
	if s.Leader && s.LeaderVMIndex >= 0 {
		vmi = s.LeaderVMIndex
	}
	vmt0 := s.ProcessLists.Get(s.LLeaderHeight)
	var vmt *VM
	lmin := "-"
	if vmt0 != nil && vmi >= 0 {
		vmt = vmt0.VMs[vmi]
		lmin = fmt.Sprintf("%2d", vmt.LeaderMinute)
	}

	vmin := s.CurrentMinute
	if s.CurrentMinute > 9 {
		vmin = 0
	}

	found, vm := s.GetVirtualServers(s.LLeaderHeight, vmin, s.GetIdentityChainID())
	vmIndex := ""
	if found {
		vmIndex = fmt.Sprintf("vm%02d", vm)
	}
	L := "_"
	X := "_"
	W := "_"
	N := "_"
	list := s.ProcessLists.Get(s.LLeaderHeight)
	if found {
		L = "L"
		if list != nil {
		}
	} else {
		if list != nil {
			if foundAudit, _ := list.GetAuditServerIndexHash(s.GetIdentityChainID()); foundAudit {
				if foundAudit {
					L = "A"
				}
			}
		}
	}
	if s.NetStateOff {
		X = "X"
	}
	if !s.RunLeader && found {
		W = "W"
	} else if s.IgnoreMissing {
		W = "I"
	}

	stype := fmt.Sprintf("%1s%1s%1s%1s", L, X, W, N)

	keyMR := primitives.NewZeroHash().Bytes()
	var d interfaces.IDirectoryBlock
	var dHeight uint32
	switch {
	case s.DBStates == nil:

	case s.LLeaderHeight == 0:

	case s.DBStates.Last() == nil:

	case s.DBStates.Last().DirectoryBlock == nil:

	default:
		d = s.DBStates.Get(int(s.GetHighestSavedBlk())).DirectoryBlock
		keyMR = d.GetKeyMR().Bytes()
		dHeight = d.GetHeader().GetDBHeight()
	}

	totalTPS, instantTPS := s.CalculateTransactionRate()

	str := fmt.Sprintf("%10s[%6x] %4s%4s %2d/%2d %2d.%01d%% %2d.%03d",
		s.FactomNodeName,
		s.IdentityChainID.Bytes()[3:6],
		stype,
		vmIndex,
		s.ResetTryCnt,
		s.ResetCnt,
		s.DropRate/10, s.DropRate%10,
		s.Delay/1000, s.Delay%1000)

	pls := fmt.Sprintf("%d/%d/%d", s.ProcessLists.DBHeightBase, s.PLProcessHeight, (s.GetTrueLeaderHeight() + 2))

	str = str + fmt.Sprintf(" %5d[%6x] %-11s ",
		dHeight,
		keyMR[:3],
		pls)

	dbstate := fmt.Sprintf("%d/%d/%d/%d", s.DBStateAskCnt, s.DBStateReplyCnt, s.DBStateIgnoreCnt, s.DBStateAppliedCnt)
	missing := fmt.Sprintf("%d/%d/%d/%d", s.MissingRequestAskCnt, s.MissingRequestReplyCnt, s.MissingRequestIgnoreCnt, s.MissingResponseAppliedCnt)
	str = str + fmt.Sprintf(" %2s/%2d %15s %26s ",
		lmin,
		s.CurrentMinute,
		dbstate,
		missing)

	trans := fmt.Sprintf("%d/%d/%d", s.FactoidTrans, s.NewEntryChains, s.NewEntries-s.NewEntryChains)
	apis := fmt.Sprintf("%d/%d/%d", s.FCTSubmits, s.ECCommits, s.ECommits)
	stps := fmt.Sprintf("%3.2f/%3.2f", totalTPS, instantTPS)
	str = str + fmt.Sprintf(" %5d %5d %12s %15s %11s",
		s.ResendCnt,
		s.ExpireCnt,
		trans,
		apis,
		stps)

	if s.Balancehash == nil {
		s.Balancehash = primitives.NewZeroHash()
	}

	str = str + fmt.Sprintf(" %5d-:-%d", s.LLeaderHeight, s.CurrentMinute)
	if list == nil {
		return
	}

	if list.System.Height < len(list.System.List) {
		str = str + fmt.Sprintf(" VM:%s %s", "?", "-nil-")
	} else {
		str = str + " -"
	}

	str = str + fmt.Sprintf(" %x", s.Balancehash.Bytes()[:3])

	s.serverPrt = str

	authoritiesString := ""
	for _, str := range s.ConstructAuthoritySetString() {
		if len(authoritiesString) > 0 {
			authoritiesString += "\n"
		}
		authoritiesString += str
	}
	// Any updates required to the state as established by the AdminBlock are applied here.
	list.State.SetAuthoritySetString(authoritiesString)

}

func (s *State) ConstructAuthoritySetString() (authSets []string) {
	base := s.ProcessLists.DBHeightBase
	for i, pl := range s.ProcessLists.Lists {
		// If we don't really have a process list at this height, then say no more.
		if i > 8 || pl == nil {
			break
		}
		authoritiesString := fmt.Sprintf("%7s (%4d) Feds:", s.FactomNodeName, int(base)+i)
		for _, fd := range pl.FedServers {
			authoritiesString += " " + fd.GetChainID().String()[6:10]
		}
		authoritiesString += " || Auds :"
		for _, fd := range pl.AuditServers {
			authoritiesString += " " + fd.GetChainID().String()[6:10]
		}
		authSets = append(authSets, authoritiesString)
	}
	return
}

// returns what finished block height this node thinks the leader is at, assuming that the
// local node has the process list the leader is working on plus an extra empty one on top of it.
func (s *State) GetTrueLeaderHeight() uint32 {
	h := int(s.ProcessLists.DBHeightBase) + len(s.ProcessLists.Lists) - 3
	if h < 0 {
		h = 0
	}
	return uint32(h)
}

func (s *State) Print(a ...interface{}) (n int, err error) {
	if s.OutputAllowed {
		str := ""
		for _, v := range a {
			str = str + fmt.Sprintf("%v", v)
		}

		if s.LastPrint == str {
			s.LastPrintCnt++
			fmt.Print(s.LastPrintCnt, " ")
		} else {
			s.LastPrint = str
			s.LastPrintCnt = 0
		}
		return fmt.Print(str)
	}

	return 0, nil
}

func (s *State) Println(a ...interface{}) (n int, err error) {
	if s.OutputAllowed {
		str := ""
		for _, v := range a {
			str = str + fmt.Sprintf("%v", v)
		}
		str = str + "\n"

		if s.LastPrint == str {
			s.LastPrintCnt++
			fmt.Print(s.LastPrintCnt, " ")
		} else {
			s.LastPrint = str
			s.LastPrintCnt = 0
		}
		return fmt.Print(str)
	}

	return 0, nil
}

func (s *State) GetOut() bool {
	return s.OutputAllowed
}

func (s *State) SetOut(o bool) {
	s.OutputAllowed = o
}

func (s *State) GetSystemHeight(dbheight uint32) int {
	pl := s.ProcessLists.Get(dbheight)
	if pl == nil {
		return -1
	}
	return pl.System.Height
}

// Gets the system message at the given dbheight, and given height in the
// System list
func (s *State) GetSystemMsg(dbheight uint32, height uint32) interfaces.IMsg {
	pl := s.ProcessLists.Get(dbheight)
	if pl == nil {
		return nil
	}
	if height >= uint32(len(pl.System.List)) {
		return nil
	}
	return pl.System.List[height]
}

func (s *State) GetInvalidMsg(hash interfaces.IHash) interfaces.IMsg {
	if hash == nil {
		return nil
	}

	s.InvalidMessagesMutex.RLock()
	defer s.InvalidMessagesMutex.RUnlock()

	return s.InvalidMessages[hash.Fixed()]
}

func (s *State) ProcessInvalidMsgQueue() {
	s.InvalidMessagesMutex.Lock()
	defer s.InvalidMessagesMutex.Unlock()
	if len(s.InvalidMessages)+len(s.networkInvalidMsgQueue) > 2048 {
		//Clearing old invalid inMessages
		s.InvalidMessages = map[[32]byte]interfaces.IMsg{}
	}

	for {
		if len(s.networkInvalidMsgQueue) == 0 {
			return
		}
		select {
		case msg := <-s.networkInvalidMsgQueue:
			s.InvalidMessages[msg.GetHash().Fixed()] = msg
		}
	}
}

func (s *State) SetPendingSigningKey(p *primitives.PrivateKey) {
	s.serverPendingPrivKeys = append(s.serverPendingPrivKeys, p)
	s.serverPendingPubKeys = append(s.serverPendingPubKeys, p.Pub)
}

func (s *State) AddStatus(status string) {
	// Don't add duplicates.
	last := s.GetLastStatus()
	if last == status {
		return
	}

	s.StatusMutex.Lock()
	defer s.StatusMutex.Unlock()

	if len(s.StatusStrs) > 1000 {
		copy(s.StatusStrs, s.StatusStrs[1:])
		s.StatusStrs[len(s.StatusStrs)-1] = status
	} else {
		s.StatusStrs = append(s.StatusStrs, status)
	}
}

func (s *State) GetStatus() []string {
	s.StatusMutex.Lock()
	defer s.StatusMutex.Unlock()

	status := make([]string, len(s.StatusStrs))
	status = append(status, s.StatusStrs...)
	return status
}

func (s *State) GetLastStatus() string {
	s.StatusMutex.Lock()
	defer s.StatusMutex.Unlock()

	if len(s.StatusStrs) == 0 {
		return ""
	}
	str := s.StatusStrs[len(s.StatusStrs)-1]
	return str
}

func (s *State) updateNetworkControllerConfig() {
	if s.NetworkController == nil {
		return
	}

	var newPeersConfig string
	switch s.Network {
	case "MAIN", "main":
		newPeersConfig = s.MainSpecialPeers
	case "TEST", "test":
		newPeersConfig = s.TestSpecialPeers
	case "LOCAL", "local":
		newPeersConfig = s.LocalSpecialPeers
	case "CUSTOM", "custom":
		newPeersConfig = s.CustomSpecialPeers
	default:
		// should already be verified earlier
		panic(fmt.Sprintf("Invalid Network: %s", s.Network))
	}

	s.NetworkController.ReloadSpecialPeers(newPeersConfig)
}

// Check and Add a hash to the network replay filter
func (s *State) AddToReplayFilter(mask int, hash [32]byte, timestamp interfaces.Timestamp, systemtime interfaces.Timestamp) (rval bool) {
	return s.Replay.IsTSValidAndUpdateState(constants.NETWORK_REPLAY, hash, timestamp, systemtime)
}

// Return if a feature is active for the current height
func (s *State) IsActive(id activations.ActivationType) bool {
	highestCompletedBlk := s.GetHighestCompletedBlk()

	rval := activations.IsActive(id, int(highestCompletedBlk))

	if rval && !s.reportedActivations[id] {
		s.LogPrintf("executeMsg", "Activating Feature %s at height %v", id.String(), highestCompletedBlk)
		s.reportedActivations[id] = true
	}

	return rval
}

func (s *State) PassOutputRegEx(RegEx *regexp.Regexp, RegExString string) {
	s.OutputRegEx = RegEx
	s.OutputRegExString = RegExString
}

func (s *State) GetOutputRegEx() (*regexp.Regexp, string) {
	return s.OutputRegEx, s.OutputRegExString
}

func (s *State) PassInputRegEx(RegEx *regexp.Regexp, RegExString string) {
	s.InputRegEx = RegEx
	s.InputRegExString = RegExString
}

func (s *State) GetInputRegEx() (*regexp.Regexp, string) {
	return s.InputRegEx, s.InputRegExString
}

func (s *State) GetIgnoreDone() bool {
	return s.IgnoreDone
}

func (s *State) ShutdownNode(exitCode int) {
	fmt.Println(fmt.Sprintf("Initiating a graceful shutdown of node %s. The exit code is %v.", s.FactomNodeName, exitCode))
	s.RunState = runstate.Stopping
	s.ShutdownChan <- exitCode
}

func (s *State) GetDBFinished() bool {
	return s.DBFinished
}

func (s *State) GetRunLeader() bool {
	return s.RunLeader
}
