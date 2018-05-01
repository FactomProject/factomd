// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

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
	"github.com/FactomProject/factomd/wsapi"
	"github.com/FactomProject/logrustash"

	log "github.com/sirupsen/logrus"
)

// packageLogger is the general logger for all package related logs. You can add additional fields,
// or create more context loggers off of this
var packageLogger = log.WithFields(log.Fields{"package": "state"})

var _ = fmt.Print

type State struct {
	Logger            *log.Entry
	IsRunning         bool
	NetworkController *p2p.Controller
	Salt              interfaces.IHash
	Cfg               interfaces.IFactomConfig
	ConfigFilePath    string // $HOME/.factom/m2/factomd.conf by default

	Prefix            string
	FactomNodeName    string
	FactomdVersion    string
	LogPath           string
	LdbPath           string
	BoltDBPath        string
	LogLevel          string
	ConsoleLogLevel   string
	NodeMode          string
	DBType            string
	CloneDBType       string
	ExportData        bool
	ExportDataSubpath string

	LogBits int64 // Bit zero is for logging the Directory Block on DBSig [5]

	DBStatesSent            []*interfaces.DBStateSent
	DBStatesReceivedBase    int
	DBStatesReceived        []*messages.DBStateMsg
	LocalServerPrivKey      string
	DirectoryBlockInSeconds int
	PortNumber              int
	Replay                  *Replay
	FReplay                 *Replay
	CrossReplay             *CrossReplayFilter
	DropRate                int
	Delay                   int64 // Simulation delays sending messages this many milliseconds

	ControlPanelPort    int
	ControlPanelSetting int
	// Keeping the last display state lets us know when to send over the new blocks
	LastDisplayState        *DisplayState
	ControlPanelChannel     chan DisplayState
	ControlPanelDataRequest bool // If true, update Display state

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
	starttime   time.Time
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
	TimeOffset             interfaces.Timestamp
	MaxTimeOffset          interfaces.Timestamp
	networkOutMsgQueue     NetOutMsgQueue
	networkInvalidMsgQueue chan interfaces.IMsg
	inMsgQueue             InMsgMSGQueue
	inMsgQueue2            InMsgMSGQueue
	electionsQueue         ElectionQueue
	apiQueue               APIMSGQueue
	ackQueue               chan interfaces.IMsg
	msgQueue               chan interfaces.IMsg

	ShutdownChan chan int // For gracefully halting Factom
	JournalFile  string
	Journaling   bool

	serverPrivKey         *primitives.PrivateKey
	serverPubKey          *primitives.PublicKey
	serverPendingPrivKeys []*primitives.PrivateKey
	serverPendingPubKeys  []*primitives.PublicKey

	// RPC connection config
	RpcUser     string
	RpcPass     string
	RpcAuthHash []byte

	FactomdTLSEnable   bool
	factomdTLSKeyFile  string
	factomdTLSCertFile string
	FactomdLocations   string

	// Server State
	StartDelay      int64 // Time in Milliseconds since the last DBState was applied
	StartDelayLimit int64
	DBFinished      bool
	RunLeader       bool
	BootTime        int64 // Time in seconds that we last booted

	// Ignore missing messages for a period to allow rebooting a network where your
	// own messages from the previously executing network can confuse you.
	IgnoreDone    bool
	IgnoreMissing bool

	LLeaderHeight   uint32
	Leader          bool
	LeaderVMIndex   int
	LeaderPL        *ProcessList
	PLProcessHeight uint32
	// Height cutoff where no missing messages below this height
	DBHeightAtBoot uint32
	OneLeader      bool
	OutputAllowed  bool
	LeaderNewMin   int
	CurrentMinute  int

	// These are the start times for blocks and minutes
	CurrentMinuteStartTime int64
	CurrentBlockStartTime  int64

	EOMsyncing bool

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

	// By default, this is false, which means DBstates are discarded
	// when a majority of leaders disagree with the hash we have via DBSigs
	KeepMismatch bool

	DBSigFails int // Keep track of how many blockhash mismatches we've had to correct

	Saving  bool // True if we are in the process of saving to the database
	Syncing bool // Looking for messages from leaders to sync

	NetStateOff     bool // Disable if true, Enable if false
	DebugConsensus  bool // If true, dump consensus trace
	FactoidTrans    int
	ECCommits       int
	ECommits        int
	FCTSubmits      int
	NewEntryChains  int
	NewEntries      int
	LeaderTimestamp interfaces.Timestamp
	// Maps
	// ====
	// For Follower
	ResendHolding interfaces.Timestamp         // Timestamp to gate resending holding to neighbors
	Holding       map[[32]byte]interfaces.IMsg // Hold Messages
	XReview       []interfaces.IMsg            // After the EOM, we must review the messages in Holding
	Acks          map[[32]byte]interfaces.IMsg // Hold Acknowledgements
	Commits       *SafeMsgMap                  //  map[[32]byte]interfaces.IMsg // Commit Messages

	InvalidMessages      map[[32]byte]interfaces.IMsg
	InvalidMessagesMutex sync.RWMutex

	AuditHeartBeats []interfaces.IMsg // The checklist of HeartBeats for this period

	FaultTimeout  int
	FaultWait     int
	EOMfaultIndex int
	LastTiebreak  int64

	AuthoritySetString string
	// Network MAIN = 0, TEST = 1, LOCAL = 2, CUSTOM = 3
	NetworkNumber int // Encoded into Directory Blocks(s.Cfg.(*util.FactomdConfig)).String()

	// Database
	DB     interfaces.DBOverlaySimple
	Anchor interfaces.IAnchor

	// Directory Block State
	DBStates *DBStateList // Holds all DBStates not yet processed.

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
	HighestKnown    uint32
	HighestAck      uint32
	AuthorityDeltas string

	// Factom State
	FactoidState    interfaces.IFactoidState
	NumTransactions int

	// Permanent balances from processing blocks.
	FactoidBalancesP      map[[32]byte]int64
	FactoidBalancesPMutex sync.Mutex
	ECBalancesP           map[[32]byte]int64
	ECBalancesPMutex      sync.Mutex
	TempBalanceHash       interfaces.IHash
	Balancehash           interfaces.IHash

	// Web Services
	Port int

	// For Replay / journal
	IsReplaying     bool
	ReplayTimestamp interfaces.Timestamp

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
	WaitForEntries  bool
	UpdateEntryHash chan *EntryUpdate // Channel for updating entry Hashes tracking (repeats and such)
	WriteEntry      chan interfaces.IEBEntry
	// MessageTally causes the node to keep track of (and display) running totals of each
	// type of message received during the tally interval
	MessageTally           bool
	MessageTalliesReceived [constants.NUM_MESSAGES]int
	MessageTalliesSent     [constants.NUM_MESSAGES]int

	LastPrint    string
	LastPrintCnt int

	// FER section
	FactoshisPerEC                 uint64
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
	useTorrents             bool
	torrentUploader         bool
	Uploader                *UploadController // Controls the uploads of torrents. Prevents backups
	DBStateManager          interfaces.IManagerController
	HighestCompletedTorrent uint32
	FastBoot                bool
	FastBootLocation        string

	// These stats are collected when we write the dbstate to the database.
	NumNewChains   int // Number of new Chains in this block
	NumNewEntries  int // Number of new Entries, not counting the first entry in a chain
	NumEntries     int // Number of entries in this block (including the entries that create chains)
	NumEntryBlocks int // Number of Entry Blocks
	NumFCTTrans    int // Number of Factoid Transactions in this block

	// debug message
	pstate              string
	SyncingState        [256]string
	SyncingStateCurrent int
	processCnt          int64 // count of attempts to process .. so we can see if the thread is running
}

var _ interfaces.IState = (*State)(nil)

type EntryUpdate struct {
	Hash      interfaces.IHash
	Timestamp interfaces.Timestamp
}

func (s *State) GetConfigPath() string {
	return s.ConfigFilePath
}

func (s *State) Running() bool {
	return s.IsRunning
}

func (s *State) Clone(cloneNumber int) interfaces.IState {
	newState := new(State)
	number := fmt.Sprintf("%02d", cloneNumber)

	simConfigPath := util.GetHomeDir() + "/.factom/m2/simConfig/"
	configfile := fmt.Sprintf("%sfactomd%03d.conf", simConfigPath, cloneNumber)

	if cloneNumber == 1 {
		os.Stderr.WriteString(fmt.Sprintf("Looking for Config File %s\n", configfile))
	}
	if _, err := os.Stat(simConfigPath); os.IsNotExist(err) {
		os.Stderr.WriteString("Creating simConfig directory\n")
		os.MkdirAll(simConfigPath, 0777)
	}

	config := false
	if _, err := os.Stat(configfile); !os.IsNotExist(err) {
		os.Stderr.WriteString(fmt.Sprintf("   Using the %s config file.\n", configfile))
		newState.LoadConfig(configfile, s.GetNetworkName())
		config = true
	}

	if s.LogPath == "stdout" {
		newState.LogPath = "stdout"
	} else {
		newState.LogPath = s.LogPath + "/Sim" + number
	}

	newState.FactomNodeName = s.Prefix + "FNode" + number
	newState.FactomdVersion = s.FactomdVersion
	newState.DropRate = s.DropRate
	newState.LdbPath = s.LdbPath + "/Sim" + number
	newState.JournalFile = s.LogPath + "/journal" + number + ".log"
	newState.Journaling = s.Journaling
	newState.BoltDBPath = s.BoltDBPath + "/Sim" + number
	newState.LogLevel = s.LogLevel
	newState.ConsoleLogLevel = s.ConsoleLogLevel
	newState.NodeMode = "FULL"
	newState.CloneDBType = s.CloneDBType
	newState.DBType = s.CloneDBType
	newState.ExportData = s.ExportData
	newState.ExportDataSubpath = s.ExportDataSubpath + "sim-" + number
	newState.Network = s.Network
	newState.MainNetworkPort = s.MainNetworkPort
	newState.PeersFile = s.PeersFile
	newState.MainSeedURL = s.MainSeedURL
	newState.MainSpecialPeers = s.MainSpecialPeers
	newState.TestNetworkPort = s.TestNetworkPort
	newState.TestSeedURL = s.TestSeedURL
	newState.TestSpecialPeers = s.TestSpecialPeers
	newState.LocalNetworkPort = s.LocalNetworkPort
	newState.LocalSeedURL = s.LocalSeedURL
	newState.LocalSpecialPeers = s.LocalSpecialPeers
	newState.CustomNetworkPort = s.CustomNetworkPort
	newState.CustomSeedURL = s.CustomSeedURL
	newState.CustomSpecialPeers = s.CustomSpecialPeers
	newState.StartDelayLimit = s.StartDelayLimit
	newState.CustomNetworkID = s.CustomNetworkID

	newState.DirectoryBlockInSeconds = s.DirectoryBlockInSeconds
	newState.PortNumber = s.PortNumber

	newState.ControlPanelPort = s.ControlPanelPort
	newState.ControlPanelSetting = s.ControlPanelSetting

	//newState.Identities = s.Identities
	//newState.Authorities = s.Authorities
	newState.AuthorityServerCount = s.AuthorityServerCount

	newState.IdentityControl = s.IdentityControl.Clone()

	newState.FaultTimeout = s.FaultTimeout
	newState.FaultWait = s.FaultWait
	newState.EOMfaultIndex = s.EOMfaultIndex

	if !config {
		newState.IdentityChainID = primitives.Sha([]byte(newState.FactomNodeName))
		//generate and use a new deterministic PrivateKey for this clone
		shaHashOfNodeName := primitives.Sha([]byte(newState.FactomNodeName)) //seed the private key with node name
		clonePrivateKey := primitives.NewPrivateKeyFromHexBytes(shaHashOfNodeName.Bytes())
		newState.LocalServerPrivKey = clonePrivateKey.PrivateKeyString()
	}

	newState.SetLeaderTimestamp(s.GetLeaderTimestamp())

	//serverPrivKey primitives.PrivateKey
	//serverPubKey  primitives.PublicKey

	newState.FactoshisPerEC = s.FactoshisPerEC

	newState.Port = s.Port

	newState.OneLeader = s.OneLeader
	newState.OneLeader = s.OneLeader

	newState.RpcUser = s.RpcUser
	newState.RpcPass = s.RpcPass
	newState.RpcAuthHash = s.RpcAuthHash

	newState.FactomdTLSEnable = s.FactomdTLSEnable
	newState.factomdTLSKeyFile = s.factomdTLSKeyFile
	newState.factomdTLSCertFile = s.factomdTLSCertFile
	newState.FactomdLocations = s.FactomdLocations

	switch newState.DBType {
	case "LDB":
		newState.StateSaverStruct.FastBoot = s.StateSaverStruct.FastBoot
		newState.StateSaverStruct.FastBootLocation = newState.LdbPath
		break
	case "Bolt":
		newState.StateSaverStruct.FastBoot = s.StateSaverStruct.FastBoot
		newState.StateSaverStruct.FastBootLocation = newState.BoltDBPath
		break
	}
	return newState
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
	return s.FactomdTLSEnable, s.factomdTLSKeyFile, s.factomdTLSCertFile
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

func (s *State) GetCurrentTime() int64 {
	return time.Now().UnixNano()
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

func (s *State) GetAckChange() error {
	change, err := util.GetChangeAcksHeight(s.ConfigFilePath)
	if err != nil {
		return err
	}
	s.AckChange = change
	return nil
}

func (s *State) LoadConfig(filename string, networkFlag string) {
	s.FactomNodeName = s.Prefix + "FNode0" // Default Factom Node Name for Simulation

	if len(filename) > 0 {
		s.ConfigFilePath = filename
		s.ReadCfg(filename)

		// Get our factomd configuration information.
		cfg := s.GetCfg().(*util.FactomdConfig)

		s.Network = cfg.App.Network
		if 0 < len(networkFlag) { // Command line overrides the config file.
			s.Network = networkFlag
		}
		fmt.Printf("\n\nNetwork : %s\n", s.Network)

		networkName := strings.ToLower(s.Network) + "-"
		// TODO: improve the paths after milestone 1
		cfg.App.LdbPath = cfg.App.HomeDir + networkName + cfg.App.LdbPath
		cfg.App.BoltDBPath = cfg.App.HomeDir + networkName + cfg.App.BoltDBPath
		cfg.App.DataStorePath = cfg.App.HomeDir + networkName + cfg.App.DataStorePath
		cfg.Log.LogPath = cfg.App.HomeDir + networkName + cfg.Log.LogPath
		cfg.App.ExportDataSubpath = cfg.App.HomeDir + networkName + cfg.App.ExportDataSubpath
		cfg.App.PeersFile = cfg.App.HomeDir + networkName + cfg.App.PeersFile
		cfg.App.ControlPanelFilesPath = cfg.App.HomeDir + cfg.App.ControlPanelFilesPath

		s.LogPath = cfg.Log.LogPath + s.Prefix
		s.LdbPath = cfg.App.LdbPath + s.Prefix
		s.BoltDBPath = cfg.App.BoltDBPath + s.Prefix
		s.LogLevel = cfg.Log.LogLevel
		s.ConsoleLogLevel = cfg.Log.ConsoleLogLevel
		s.NodeMode = cfg.App.NodeMode
		s.DBType = cfg.App.DBType
		s.ExportData = cfg.App.ExportData // bool
		s.ExportDataSubpath = cfg.App.ExportDataSubpath
		s.MainNetworkPort = cfg.App.MainNetworkPort
		s.PeersFile = cfg.App.PeersFile
		s.MainSeedURL = cfg.App.MainSeedURL
		s.MainSpecialPeers = cfg.App.MainSpecialPeers
		s.TestNetworkPort = cfg.App.TestNetworkPort
		s.TestSeedURL = cfg.App.TestSeedURL
		s.TestSpecialPeers = cfg.App.TestSpecialPeers
		s.CustomBootstrapIdentity = cfg.App.CustomBootstrapIdentity
		s.CustomBootstrapKey = cfg.App.CustomBootstrapKey
		s.LocalNetworkPort = cfg.App.LocalNetworkPort
		s.LocalSeedURL = cfg.App.LocalSeedURL
		s.LocalSpecialPeers = cfg.App.LocalSpecialPeers
		s.LocalServerPrivKey = cfg.App.LocalServerPrivKey
		s.CustomNetworkPort = cfg.App.CustomNetworkPort
		s.CustomSeedURL = cfg.App.CustomSeedURL
		s.CustomSpecialPeers = cfg.App.CustomSpecialPeers
		s.FactoshisPerEC = cfg.App.ExchangeRate
		s.DirectoryBlockInSeconds = cfg.App.DirectoryBlockInSeconds
		s.PortNumber = cfg.App.PortNumber
		s.ControlPanelPort = cfg.App.ControlPanelPort
		s.RpcUser = cfg.App.FactomdRpcUser
		s.RpcPass = cfg.App.FactomdRpcPass
		s.StateSaverStruct.FastBoot = cfg.App.FastBoot
		s.StateSaverStruct.FastBootLocation = cfg.App.FastBootLocation
		s.FastBoot = cfg.App.FastBoot
		s.FastBootLocation = cfg.App.FastBootLocation

		s.FactomdTLSEnable = cfg.App.FactomdTlsEnabled
		if cfg.App.FactomdTlsPrivateKey == "/full/path/to/factomdAPIpriv.key" {
			s.factomdTLSKeyFile = fmt.Sprint(cfg.App.HomeDir, "factomdAPIpriv.key")
		}
		if cfg.App.FactomdTlsPublicCert == "/full/path/to/factomdAPIpub.cert" {
			s.factomdTLSCertFile = fmt.Sprint(cfg.App.HomeDir, "factomdAPIpub.cert")
		}
		externalIP := strings.Split(cfg.Walletd.FactomdLocation, ":")[0]
		if externalIP != "localhost" {
			s.FactomdLocations = externalIP
		}

		switch cfg.App.ControlPanelSetting {
		case "disabled":
			s.ControlPanelSetting = 0
		case "readonly":
			s.ControlPanelSetting = 1
		case "readwrite":
			s.ControlPanelSetting = 2
		default:
			s.ControlPanelSetting = 1
		}
		s.FERChainId = cfg.App.ExchangeRateChainId
		s.ExchangeRateAuthorityPublicKey = cfg.App.ExchangeRateAuthorityPublicKey
		identity, err := primitives.HexToHash(cfg.App.IdentityChainID)
		if err != nil {
			s.IdentityChainID = primitives.Sha([]byte(s.FactomNodeName))
		} else {
			s.IdentityChainID = identity
		}
	} else {
		s.LogPath = "database/"
		s.LdbPath = "database/ldb"
		s.BoltDBPath = "database/bolt"
		s.LogLevel = "none"
		s.ConsoleLogLevel = "standard"
		s.NodeMode = "SERVER"
		s.DBType = "Map"
		s.ExportData = false
		s.ExportDataSubpath = "data/export"
		s.Network = "TEST"
		s.MainNetworkPort = "8108"
		s.PeersFile = "peers.json"
		s.MainSeedURL = "https://raw.githubusercontent.com/FactomProject/factomproject.github.io/master/seed/mainseed.txt"
		s.MainSpecialPeers = ""
		s.TestNetworkPort = "8109"
		s.TestSeedURL = "https://raw.githubusercontent.com/FactomProject/factomproject.github.io/master/seed/testseed.txt"
		s.TestSpecialPeers = ""
		s.LocalNetworkPort = "8110"
		s.LocalSeedURL = "https://raw.githubusercontent.com/FactomProject/factomproject.github.io/master/seed/localseed.txt"
		s.LocalSpecialPeers = ""

		s.LocalServerPrivKey = "4c38c72fc5cdad68f13b74674d3ffb1f3d63a112710868c9b08946553448d26d"
		s.FactoshisPerEC = 006666
		s.FERChainId = "111111118d918a8be684e0dac725493a75862ef96d2d3f43f84b26969329bf03"
		s.ExchangeRateAuthorityPublicKey = "3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29"
		s.DirectoryBlockInSeconds = 6
		s.PortNumber = 8088
		s.ControlPanelPort = 8090
		s.ControlPanelSetting = 1

		// TODO:  Actually load the IdentityChainID from the config file
		s.IdentityChainID = primitives.Sha([]byte(s.FactomNodeName))

	}
	s.JournalFile = s.LogPath + "/journal0" + ".log"

	s.updateNetworkControllerConfig()
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

func (s *State) Init() {

	if s.Salt == nil {
		b := make([]byte, 32)
		_, err := rand.Read(b)
		// Note that err == nil only if we read len(b) bytes.
		if err != nil {
			panic("Random Number Failure")
		}
		s.Salt = primitives.Sha(b)
	}

	salt := fmt.Sprintf("The Instance ID of this node is %s\n", s.Salt.String()[:16])
	fmt.Print(salt)

	s.StartDelay = s.GetTimestamp().GetTimeMilli() // We can't start as a leader until we know we are upto date
	s.RunLeader = false
	s.IgnoreMissing = true
	s.BootTime = s.GetTimestamp().GetTimeSeconds()

	if s.LogPath == "stdout" {
		wsapi.InitLogs(s.LogPath, s.LogLevel)
		//s.Logger = log.NewLogFromConfig(s.LogPath, s.LogLevel, "State")
	} else {
		er := os.MkdirAll(s.LogPath, 0777)
		if er != nil {
			// fmt.Println("Could not create " + s.LogPath + "\n error: " + er.Error())
		}
		wsapi.InitLogs(s.LogPath+s.FactomNodeName+".log", s.LogLevel)
		//s.Logger = log.NewLogFromConfig(s.LogPath, s.LogLevel, "State")
	}

	s.ControlPanelChannel = make(chan DisplayState, 20)
	s.tickerQueue = make(chan int, 100)                        //ticks from a clock
	s.timerMsgQueue = make(chan interfaces.IMsg, 100)          //incoming eom notifications, used by leaders
	s.TimeOffset = new(primitives.Timestamp)                   //interfaces.Timestamp(int64(rand.Int63() % int64(time.Microsecond*10)))
	s.networkInvalidMsgQueue = make(chan interfaces.IMsg, 100) //incoming message queue from the network messages
	s.InvalidMessages = make(map[[32]byte]interfaces.IMsg, 0)
	s.networkOutMsgQueue = NewNetOutMsgQueue(5000)                 //Messages to be broadcast to the network
	s.inMsgQueue = NewInMsgQueue(constants.INMSGQUEUE_HIGH + 100)  //incoming message queue for Factom application messages
	s.inMsgQueue2 = NewInMsgQueue(constants.INMSGQUEUE_HIGH + 100) //incoming message queue for Factom application messages
	s.electionsQueue = NewElectionQueue(10000)                     //incoming message queue for Factom application messages
	s.apiQueue = NewAPIQueue(100)                                  //incoming message queue from the API
	s.ackQueue = make(chan interfaces.IMsg, 100)                   //queue of Leadership messages
	s.msgQueue = make(chan interfaces.IMsg, 400)                   //queue of Follower messages
	s.ShutdownChan = make(chan int, 1)                             //Channel to gracefully shut down.
	s.MissingEntries = make(chan *MissingEntry, 1000)              //Entries I discover are missing from the database
	s.UpdateEntryHash = make(chan *EntryUpdate, 10000)             //Handles entry hashes and updating Commit maps.
	s.WriteEntry = make(chan interfaces.IEBEntry, 3000)            //Entries to be written to the database

	if s.Journaling {
		f, err := os.Create(s.JournalFile)
		if err != nil {
			fmt.Println("Could not create the journal file:", s.JournalFile)
			s.JournalFile = ""
		}
		f.Close()
	}
	// Set up struct to stop replay attacks
	s.Replay = new(Replay)
	s.Replay.s = s
	s.Replay.name = "Replay"

	s.FReplay = new(Replay)
	s.FReplay.s = s
	s.FReplay.name = "FReplay"

	// Set up maps for the followers
	s.Holding = make(map[[32]byte]interfaces.IMsg)
	s.Acks = make(map[[32]byte]interfaces.IMsg)
	s.Commits = NewSafeMsgMap("commits", s) //make(map[[32]byte]interfaces.IMsg)

	// Setup the FactoidState and Validation Service that holds factoid and entry credit balances
	s.FactoidBalancesP = map[[32]byte]int64{}
	s.ECBalancesP = map[[32]byte]int64{}

	fs := new(FactoidState)
	fs.State = s
	s.FactoidState = fs

	// Allocate the original set of Process Lists
	s.ProcessLists = NewProcessLists(s)
	s.FaultWait = 3
	s.LastTiebreak = 0
	s.EOMfaultIndex = 0

	s.DBStates = new(DBStateList)
	s.DBStates.State = s
	s.DBStates.DBStates = make([]*DBState, 0)

	switch s.NodeMode {
	case "FULL":
		s.Leader = false
		s.Println("\n   +---------------------------+")
		s.Println("   +------ Follower Only ------+")
		s.Print("   +---------------------------+\n\n")
	case "SERVER":
		s.Println("\n   +-------------------------+")
		s.Println("   |       Leader Node       |")
		s.Print("   +-------------------------+\n\n")
	default:
		panic("Bad Node Mode (must be FULL or SERVER)")
	}

	//Database
	switch s.DBType {
	case "LDB":
		if err := s.InitLevelDB(); err != nil {
			panic(fmt.Sprintf("Error initializing the database: %v", err))
		}
	case "Bolt":
		if err := s.InitBoltDB(); err != nil {
			panic(fmt.Sprintf("Error initializing the database: %v", err))
		}
	case "Map":
		if err := s.InitMapDB(); err != nil {
			panic(fmt.Sprintf("Error initializing the database: %v", err))
		}
	default:
		panic("No Database type specified")
	}

	if s.ExportData {
		s.DB.SetExportData(s.ExportDataSubpath)
	}

	// Cross Boot Replay
	switch s.DBType {
	case "Map":
		s.SetupCrossBootReplay("Map")
	default:
		s.SetupCrossBootReplay("Bolt")
	}

	//Network
	switch s.Network {
	case "MAIN":
		s.NetworkNumber = constants.NETWORK_MAIN
		s.DirectoryBlockInSeconds = 600
	case "TEST":
		s.NetworkNumber = constants.NETWORK_TEST
	case "LOCAL":
		s.NetworkNumber = constants.NETWORK_LOCAL
	case "CUSTOM":
		s.NetworkNumber = constants.NETWORK_CUSTOM
	default:
		panic("Bad value for Network in factomd.conf")
	}

	s.Println("\nRunning on the ", s.Network, "Network")
	s.Println("\nExchange rate chain id set to ", s.FERChainId)
	s.Println("\nExchange rate Authority Public Key set to ", s.ExchangeRateAuthorityPublicKey)

	s.AuditHeartBeats = make([]interfaces.IMsg, 0)

	// If we cloned the Identity control of another node, don't reset!
	if s.IdentityControl == nil {
		s.IdentityControl = NewIdentityManager()
	}
	s.initServerKeys()
	s.AuthorityServerCount = 0

	//LoadIdentityCache(s)
	//StubIdentityCache(s)
	//needed for multiple nodes with FER.  remove for singe node launch
	if s.FERChainId == "" {
		s.FERChainId = "111111118d918a8be684e0dac725493a75862ef96d2d3f43f84b26969329bf03"
	}
	if s.ExchangeRateAuthorityPublicKey == "" {
		s.ExchangeRateAuthorityPublicKey = "3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29"
	}
	// end of FER removal
	s.starttime = time.Now()

	if s.StateSaverStruct.FastBoot {
		d, err := s.DB.FetchDBlockHead()
		if err != nil {
			panic(err)
		}

		if d == nil || d.GetDatabaseHeight() < 2000 {
			//If we have less than 2k blocks, we wipe SaveState
			//This is to ensure we don't accidentally keep SaveState while deleting a database
			s.StateSaverStruct.DeleteSaveState(s.Network)
		} else {
			err = s.StateSaverStruct.LoadDBStateList(s.DBStates, s.Network)
			if err != nil {
				panic(err)
			}
		}
	}

	s.Logger = log.WithFields(log.Fields{"node-name": s.GetFactomNodeName(), "identity": s.GetIdentityChainID().String()})

	// Set up Logstash Hook for Logrus (if enabled)
	if s.UseLogstash {
		err := s.HookLogstash()
		if err != nil {
			log.Fatal(err)
		}
	}

}

func (s *State) HookLogstash() error {
	hook, err := logrustash.NewAsyncHook("tcp", s.LogstashURL, "factomdLogs")
	if err != nil {
		return err
	}

	hook.ReconnectBaseDelay = time.Second // Wait for one second before first reconnect.
	hook.ReconnectDelayMultiplier = 2
	hook.MaxReconnectRetries = 10

	s.Logger.Logger.Hooks.Add(hook)
	return nil
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

func (s *State) GetEBlockKeyMRFromEntryHash(entryHash interfaces.IHash) interfaces.IHash {
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
//  This is what fills the HoldingMap while locking it againstt a read while building
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

// JournalMessage writes the message to the message journal for debugging
func (s *State) JournalMessage(msg interfaces.IMsg) {
	type journalentry struct {
		Type    byte
		Message interfaces.IMsg
	}

	if s.Journaling && len(s.JournalFile) != 0 {
		f, err := os.OpenFile(s.JournalFile, os.O_APPEND+os.O_WRONLY, 0666)
		if err != nil {
			s.JournalFile = ""
			return
		}
		defer f.Close()

		e := new(journalentry)
		e.Type = msg.Type()
		e.Message = msg

		p, err := json.Marshal(e)
		if err != nil {
			return
		}
		fmt.Fprintln(f, string(p))
	}
}

// GetJournalMessages gets all messages from the message journal
func (s *State) GetJournalMessages() [][]byte {
	ret := make([][]byte, 0)
	if !s.Journaling || len(s.JournalFile) == 0 {
		return nil
	}

	f, err := os.Open(s.JournalFile)
	if err != nil {
		s.JournalFile = ""
		return nil
	}
	defer f.Close()

	r := bufio.NewReader(f)
	for {
		p, err := r.ReadBytes('\n')
		if err != nil {
			break
		}
		ret = append(ret, p)
	}

	return ret
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
	dbheight := s.GetHighestSavedBlk()
	plbase := s.ProcessLists.DBHeightBase
	if dbheight == 0 {
		dbheight++
	}
	if dbheight > 1 {
		dbheight--
	}

	ProcessLists := s.ProcessLists
	if ProcessLists.SetString {
		ProcessLists.SetString = false
		ProcessLists.Str = ProcessLists.String()
	}

	if plbase <= dbheight { // TODO: This is where we have to fix the fact that syncing with dbstates can fail to transition to messages
		if !s.Leader || s.RunLeader {
			progress = ProcessLists.UpdateState(dbheight)
		}
	}

	p2 := s.DBStates.UpdateState()
	progress = progress || p2

	s.SetString()
	if s.ControlPanelDataRequest {
		s.CopyStateToControlPanel()
	}

	// Update our TPS every ~ 3 seconds at the earliest
	if s.lasttime.Before(time.Now().Add(-3 * time.Second)) {
		s.CalculateTransactionRate()
	}

	// check to see ig a holding queue list request has been made
	s.fillHoldingMap()
	s.fillAcksMap()

entryHashProcessing:
	for {
		select {
		case e := <-s.UpdateEntryHash:
			// Save the entry hash, and remove from commits IF this hash is valid in this current timeframe.
			s.Replay.SetHashNow(constants.REVEAL_REPLAY, e.Hash.Fixed(), e.Timestamp)
			// If the SetHashNow worked, then we should prohibit any commit that might be pending.
			// Remove any commit that might be around.
			s.Commits.Delete(e.Hash.Fixed())
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

func (s *State) GetIdentityChainID() interfaces.IHash {
	return s.IdentityChainID
}

func (s *State) SetIdentityChainID(chainID interfaces.IHash) {
	s.IdentityChainID = chainID
}

func (s *State) GetDirectoryBlockInSeconds() int {
	return s.DirectoryBlockInSeconds
}

func (s *State) SetDirectoryBlockInSeconds(t int) {
	s.DirectoryBlockInSeconds = t
}

func (s *State) GetServerPrivateKey() *primitives.PrivateKey {
	return s.serverPrivKey
}

func (s *State) GetServerPublicKey() *primitives.PublicKey {
	return s.serverPubKey
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
	s.serverPrivKey, err = primitives.NewPrivateKeyFromHex(s.LocalServerPrivKey)
	if err != nil {
		//panic("Cannot parse Server Private Key from configuration file: " + err.Error())
	}
	s.serverPubKey = s.serverPrivKey.Pub
}

func (s *State) Log(level string, message string) {
	packageLogger.WithFields(s.Logger.Data).Info(message)
}

func (s *State) Logf(level string, format string, args ...interface{}) {
	llog := packageLogger.WithFields(s.Logger.Data)
	switch level {
	case "emergency":
		llog.Panicf(format, args...)
	case "alert":
		llog.Panicf(format, args...)
	case "critical":
		llog.Panicf(format, args...)
	case "error":
		llog.Errorf(format, args...)
	case "llog":
		llog.Warningf(format, args...)
	case "info":
		llog.Infof(format, args...)
	case "debug":
		llog.Debugf(format, args...)
	default:
		llog.Infof(format, args...)
	}
}

func (s *State) GetAuditHeartBeats() []interfaces.IMsg {
	return s.AuditHeartBeats
}

func (s *State) SetIsReplaying() {
	s.IsReplaying = true
}

func (s *State) SetIsDoneReplaying() {
	s.IsReplaying = false
	s.ReplayTimestamp = nil
}

// Returns a millisecond timestamp
func (s *State) GetTimestamp() interfaces.Timestamp {
	if s.IsReplaying == true {
		fmt.Println("^^^^^^^^ IsReplying is true")
		return s.ReplayTimestamp
	}
	return primitives.NewTimestampNow()
}

func (s *State) GetTimeOffset() interfaces.Timestamp {
	return s.TimeOffset
}

func (s *State) Sign(b []byte) interfaces.IFullSignature {
	return s.serverPrivKey.Sign(b)
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

func (s *State) GetPort() int { return s.PortNumber }

func (s *State) TickerQueue() chan int {
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

func (s *State) GetLeaderTimestamp() interfaces.Timestamp {
	if s.LeaderTimestamp == nil {
		s.LeaderTimestamp = new(primitives.Timestamp)
	}
	return s.LeaderTimestamp
}

func (s *State) SetLeaderTimestamp(ts interfaces.Timestamp) {
	s.LeaderTimestamp = ts
}

func (s *State) SetFaultTimeout(timeout int) {
	s.FaultTimeout = timeout
}

func (s *State) SetFaultWait(wait int) {
	s.FaultWait = wait
}

//var _ IState = (*State)(nil)

// GetAuthorities will return a list of the network authorities
func (s *State) GetAuthorities() []interfaces.IAuthority {
	return s.IdentityControl.GetAuthorities()
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
func (s *State) GetNetworkBootStrapKey() interfaces.IHash {
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
func (s *State) GetNetworkBootStrapIdentity() interfaces.IHash {
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

// The identity for validating messages
func (s *State) GetNetworkSkeletonIdentity() interfaces.IHash {
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

func (s *State) GetNetworkIdentityRegistrationChain() interfaces.IHash {
	id, _ := primitives.HexToHash("888888001750ede0eff4b05f0c3f557890b256450cabbb84cada937f9c258327")
	return id
}

func (s *State) GetMatryoshka(dbheight uint32) interfaces.IHash {
	return nil
}

func (s *State) InitLevelDB() error {
	if s.DB != nil {
		return nil
	}

	path := s.LdbPath + "/" + s.Network + "/" + "factoid_level.db"

	s.Println("Database:", path)

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

	str := fmt.Sprintf(" %s\n %10s %6s %12s %5s %4s %6s %10s %8s %5s %4s %20s %12s %10s %-8s %-9s %15s %9s %9s %s\n",
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
		"SysHeight",
		"BH")

	return str
}

func (s *State) SetStringConsensus() {
	str := fmt.Sprintf("%10s[%x_%x] ", s.FactomNodeName, s.IdentityChainID.Bytes()[:2], s.IdentityChainID.Bytes()[2:5])

	s.serverPrt = str
}

// CalculateTransactionRate calculates how many transactions this node is processing
//		totalTPS	: Transaction rate over life of node (totaltime / totaltrans)
//		instantTPS	: Transaction rate weighted over last 3 seconds
func (s *State) CalculateTransactionRate() (totalTPS float64, instantTPS float64) {
	runtime := time.Since(s.starttime)
	shorttime := time.Since(s.lasttime)
	total := s.FactoidTrans + s.NewEntryChains + s.NewEntries
	tps := float64(total) / float64(runtime.Seconds())
	TotalTransactionPerSecond.Set(tps) // Prometheus
	if shorttime > time.Second*3 {
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

	str = str + fmt.Sprintf(" %d/%d", list.System.Height, len(list.System.List))

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
	if h > 0 && uint32(h-1) > s.HighestKnown {
		s.HighestKnown = uint32(h - 1)
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
		//Clearing old invalid messages
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
