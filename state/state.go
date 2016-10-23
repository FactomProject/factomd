// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"sync"

	"crypto/rand"
	"encoding/binary"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/hybridDB"
	"github.com/FactomProject/factomd/database/mapdb"
	"github.com/FactomProject/factomd/log"
	"github.com/FactomProject/factomd/logger"
	"github.com/FactomProject/factomd/util"
	"github.com/FactomProject/factomd/wsapi"
)

var _ = fmt.Print

type State struct {
	filename string

	Salt interfaces.IHash

	Cfg interfaces.IFactomConfig

	Prefix            string
	FactomNodeName    string
	FactomdVersion    int
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

	LocalServerPrivKey      string
	DirectoryBlockInSeconds int
	PortNumber              int
	Replay                  *Replay
	DropRate                int
	Delay                   int64 // Simulation delays sending messages this many milliseconds

	ControlPanelPort        int
	ControlPanelSetting     int
	ControlPanelChannel     chan DisplayState
	ControlPanelDataRequest bool // If true, update Display state

	// Network Configuration
	Network           string
	MainNetworkPort   string
	PeersFile         string
	MainSeedURL       string
	MainSpecialPeers  string
	TestNetworkPort   string
	TestSeedURL       string
	TestSpecialPeers  string
	LocalNetworkPort  string
	LocalSeedURL      string
	LocalSpecialPeers string
	CustomNetworkID   []byte

	IdentityChainID      interfaces.IHash // If this node has an identity, this is it
	Identities           []Identity       // Identities of all servers in management chain
	Authorities          []Authority      // Identities of all servers in management chain
	AuthorityServerCount int              // number of federated or audit servers allowed

	// Just to print (so debugging doesn't drive functionaility)
	Status    int // Return a status (0 do nothing, 1 provide queues, 2 provide consensus data)
	serverPrt string
	starttime time.Time
	transCnt  int
	lasttime  time.Time
	tps       float64

	DBStateAskCnt   int
	DBStateAnsCnt   int
	DBStateReplyCnt int
	DBStateFailsCnt int

	MissingRequestSendCnt     int
	MissingRequestReplyCnt    int
	MissingRequestIgnoreCnt   int
	MissingResponseAppliedCnt int

	ResendCnt int
	ExpireCnt int

	tickerQueue            chan int
	timerMsgQueue          chan interfaces.IMsg
	TimeOffset             interfaces.Timestamp
	MaxTimeOffset          interfaces.Timestamp
	networkOutMsgQueue     chan interfaces.IMsg
	networkInvalidMsgQueue chan interfaces.IMsg
	inMsgQueue             chan interfaces.IMsg
	apiQueue               chan interfaces.IMsg
	ackQueue               chan interfaces.IMsg
	msgQueue               chan interfaces.IMsg
	ShutdownChan           chan int // For gracefully halting Factom
	JournalFile            string
	Journaling             bool

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
	RunLeader       bool
	LLeaderHeight   uint32
	Leader          bool
	LeaderVMIndex   int
	LeaderPL        *ProcessList
	PLProcessHeight uint32
	OneLeader       bool
	OutputAllowed   bool
	CurrentMinute   int

	EOMsyncing bool

	EOM          bool // Set to true when the first EOM is encountered
	EOMLimit     int
	EOMProcessed int
	EOMDone      bool
	EOMMinute    int

	DBSig          bool
	DBSigLimit     int
	DBSigProcessed int // Number of DBSignatures received and processed.
	DBSigDone      bool

	// By default, this is false, which means DBstates are discarded
	//when a majority of leaders disagree with the hash we have via DBSigs
	KeepMismatch bool

	DBSigFails int // Keep track of how many blockhash mismatches we've had to correct

	Newblk  bool // True if we are starting a new block, and a dbsig is needed.
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
	resendHolding interfaces.Timestamp           // Timestamp to gate resending holding to neighbors
	Holding       map[[32]byte]interfaces.IMsg   // Hold Messages
	XReview       []interfaces.IMsg              // After the EOM, we must review the messages in Holding
	Acks          map[[32]byte]interfaces.IMsg   // Hold Acknowledgemets
	Commits       map[[32]byte][]interfaces.IMsg // Commit Messages

	InvalidMessages      map[[32]byte]interfaces.IMsg
	InvalidMessagesMutex sync.RWMutex

	AuditHeartBeats []interfaces.IMsg // The checklist of HeartBeats for this period

	FaultTimeout    int
	FaultWait       int
	EOMfaultIndex   int
	LastFaultAction int64
	LastTiebreak    int64

	//Network MAIN = 0, TEST = 1, LOCAL = 2, CUSTOM = 3
	NetworkNumber int // Encoded into Directory Blocks(s.Cfg.(*util.FactomdConfig)).String()

	// Database
	DB     *databaseOverlay.Overlay
	Logger *logger.FLogger
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

	ResetRequest bool // Set to true to trigger a reset
	ProcessLists *ProcessLists

	// Factom State
	FactoidState    interfaces.IFactoidState
	NumTransactions int

	// Permanent balances from processing blocks.
	FactoidBalancesP      map[[32]byte]int64
	FactoidBalancesPMutex sync.Mutex
	ECBalancesP           map[[32]byte]int64
	ECBalancesPMutex      sync.Mutex

	// Web Services
	Port int

	//For Replay / journal
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
	// Height in the DBlock where we have all the entries
	EntryHeightComplete int
	// DBlock Height at which we have started asking for or have all entries
	EntryDBHeightProcessing uint32
	// Height in the Directory Block where we have
	// Entries we don't have that we are asking our neighbors for
	MissingEntries []MissingEntry

	LastPrint    string
	LastPrintCnt int

	// FER section
	FactoshisPerEC               uint64
	FERChainId                   string
	ExchangeRateAuthorityAddress string

	FERChangeHeight      uint32
	FERChangePrice       uint64
	FERPriority          uint32
	FERPrioritySetHeight uint32
}

type MissingEntryBlock struct {
	ebhash   interfaces.IHash
	dbheight uint32
}

type MissingEntry struct {
	ebhash    interfaces.IHash
	entryhash interfaces.IHash
	dbheight  uint32
}

var _ interfaces.IState = (*State)(nil)

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

	newState.FactomNodeName = s.Prefix + "FNode" + number
	newState.FactomdVersion = s.FactomdVersion
	newState.DropRate = s.DropRate
	newState.LogPath = s.LogPath + "/Sim" + number
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
	newState.StartDelayLimit = s.StartDelayLimit
	newState.CustomNetworkID = s.CustomNetworkID

	newState.DirectoryBlockInSeconds = s.DirectoryBlockInSeconds
	newState.PortNumber = s.PortNumber

	newState.ControlPanelPort = s.ControlPanelPort
	newState.ControlPanelSetting = s.ControlPanelSetting

	newState.Identities = s.Identities
	newState.Authorities = s.Authorities
	newState.AuthorityServerCount = s.AuthorityServerCount

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

	return newState
}

func (s *State) AddPrefix(prefix string) {
	s.Prefix = prefix
}

func (s *State) GetFactomNodeName() string {
	return s.FactomNodeName
}

func (s *State) GetDropRate() int {
	return s.DropRate
}

func (s *State) SetDropRate(droprate int) {
	s.DropRate = droprate
}

func (s *State) GetNetStateOff() bool { //	If true, all network communications are disabled
	return s.NetStateOff
}

func (s *State) SetNetStateOff(net bool) {
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

func (s *State) GetCurrentMinute() int {
	return s.CurrentMinute
}

func (s *State) IncDBStateAnswerCnt() {
	s.DBStateAnsCnt++
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

func (s *State) LoadConfig(filename string, networkFlag string) {
	s.FactomNodeName = s.Prefix + "FNode0" // Default Factom Node Name for Simulation

	if len(filename) > 0 {
		s.filename = filename
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
		cfg.Wallet.BoltDBPath = cfg.App.HomeDir + networkName + cfg.Wallet.BoltDBPath
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
		s.LocalNetworkPort = cfg.App.LocalNetworkPort
		s.LocalSeedURL = cfg.App.LocalSeedURL
		s.LocalSpecialPeers = cfg.App.LocalSpecialPeers
		s.LocalServerPrivKey = cfg.App.LocalServerPrivKey
		s.FactoshisPerEC = cfg.App.ExchangeRate
		s.DirectoryBlockInSeconds = cfg.App.DirectoryBlockInSeconds
		s.PortNumber = cfg.Wsapi.PortNumber
		s.ControlPanelPort = cfg.App.ControlPanelPort
		s.RpcUser = cfg.App.FactomdRpcUser
		s.RpcPass = cfg.App.FactomdRpcPass

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
		s.ExchangeRateAuthorityAddress = cfg.App.ExchangeRateAuthorityAddress
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
		s.Network = "LOCAL"
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
		s.ExchangeRateAuthorityAddress = "EC2DKSYyRcNWf7RS963VFYgMExoHRYLHVeCfQ9PGPmNzwrcmgm2r"
		s.DirectoryBlockInSeconds = 6
		s.PortNumber = 8088
		s.ControlPanelPort = 8090
		s.ControlPanelSetting = 1

		// TODO:  Actually load the IdentityChainID from the config file
		s.IdentityChainID = primitives.Sha([]byte(s.FactomNodeName))

	}
	s.JournalFile = s.LogPath + "/journal0" + ".log"
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

	s.StartDelay = s.GetTimestamp().GetTimeMilli() // We cant start as a leader until we know we are upto date
	s.RunLeader = false

	wsapi.InitLogs(s.LogPath+s.FactomNodeName+".log", s.LogLevel)

	s.Logger = logger.NewLogFromConfig(s.LogPath, s.LogLevel, "State")

	log.SetLevel(s.ConsoleLogLevel)

	s.ControlPanelChannel = make(chan DisplayState, 20)
	s.tickerQueue = make(chan int, 10000)                        //ticks from a clock
	s.timerMsgQueue = make(chan interfaces.IMsg, 10000)          //incoming eom notifications, used by leaders
	s.TimeOffset = new(primitives.Timestamp)                     //interfaces.Timestamp(int64(rand.Int63() % int64(time.Microsecond*10)))
	s.networkInvalidMsgQueue = make(chan interfaces.IMsg, 10000) //incoming message queue from the network messages
	s.InvalidMessages = make(map[[32]byte]interfaces.IMsg, 0)
	s.networkOutMsgQueue = make(chan interfaces.IMsg, 10000) //Messages to be broadcast to the network
	s.inMsgQueue = make(chan interfaces.IMsg, 10000)         //incoming message queue for factom application messages
	s.apiQueue = make(chan interfaces.IMsg, 10000)           //incoming message queue from the API
	s.ackQueue = make(chan interfaces.IMsg, 10000)           //queue of Leadership messages
	s.msgQueue = make(chan interfaces.IMsg, 10000)           //queue of Follower messages
	s.ShutdownChan = make(chan int, 1)                       //Channel to gracefully shut down.

	er := os.MkdirAll(s.LogPath, 0777)
	if er != nil {
		// fmt.Println("Could not create " + s.LogPath + "\n error: " + er.Error())
	}
	if s.Journaling {
		_, err := os.Create(s.JournalFile) //Create the Journal File
		if err != nil {
			fmt.Println("Could not create the file: " + s.JournalFile)
			s.JournalFile = ""
		}
	}
	// Set up struct to stop replay attacks
	s.Replay = new(Replay)

	// Set up maps for the followers
	s.Holding = make(map[[32]byte]interfaces.IMsg)
	s.Acks = make(map[[32]byte]interfaces.IMsg)
	s.Commits = make(map[[32]byte][]interfaces.IMsg)

	// Setup the FactoidState and Validation Service that holds factoid and entry credit balances
	s.FactoidBalancesP = map[[32]byte]int64{}
	s.ECBalancesP = map[[32]byte]int64{}

	fs := new(FactoidState)
	fs.State = s
	s.FactoidState = fs

	// Allocate the original set of Process Lists
	s.ProcessLists = NewProcessLists(s)
	s.FaultTimeout = 20
	s.FaultWait = 5
	s.LastFaultAction = 0
	s.LastTiebreak = 0
	s.EOMfaultIndex = 0
	s.FactomdVersion = constants.FACTOMD_VERSION

	s.DBStates = new(DBStateList)
	s.DBStates.LastTime = s.GetTimestamp()
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
			log.Printfln("Error initializing the database: %v", err)
		}
	case "Bolt":
		if err := s.InitBoltDB(); err != nil {
			log.Printfln("Error initializing the database: %v", err)
		}
	case "Map":
		if err := s.InitMapDB(); err != nil {
			log.Printfln("Error initializing the database: %v", err)
		}
	default:
		panic("No Database type specified")
	}

	if s.ExportData {
		s.DB.SetExportData(s.ExportDataSubpath)
	}

	//Network
	switch s.Network {
	case "MAIN":
		s.NetworkNumber = constants.NETWORK_MAIN
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
	s.Println("\nExchange rate Authority Public Key set to ", s.ExchangeRateAuthorityAddress)

	s.AuditHeartBeats = make([]interfaces.IMsg, 0)

	s.initServerKeys()
	s.AuthorityServerCount = 0

	//LoadIdentityCache(s)
	//StubIdentityCache(s)
	//needed for multiple nodes with FER.  remove for singe node launch
	if s.FERChainId == "" {
		s.FERChainId = "111111118d918a8be684e0dac725493a75862ef96d2d3f43f84b26969329bf03"
	}
	if s.ExchangeRateAuthorityAddress == "" {
		s.ExchangeRateAuthorityAddress = "EC2DKSYyRcNWf7RS963VFYgMExoHRYLHVeCfQ9PGPmNzwrcmgm2r"
	}
	// end of FER removal
	s.starttime = time.Now()
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

func (s *State) GetAndLockDB() interfaces.DBOverlay {
	return s.DB
}

func (s *State) UnlockDB() {
}

func (s *State) LoadDBState(dbheight uint32) (interfaces.IMsg, error) {
	dblk, err := s.DB.FetchDBlockByHeight(dbheight)
	if err != nil {
		return nil, err
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

	if len(dblk.GetDBEntries()) > 2 {
		for _, v := range dblk.GetDBEntries()[3:] {
			eBlock, err := s.DB.FetchEBlock(v.GetKeyMR())
			if err == nil && eBlock != nil {
				eBlocks = append(eBlocks, eBlock)
				for _, e := range eBlock.GetEntryHashes() {
					entry, err := s.DB.FetchEntry(e)
					if err == nil && entry != nil {
						entries = append(entries, entry)
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

	msg := messages.NewDBStateMsg(s.GetTimestamp(), dblk, ablk, fblk, ecblk, eBlocks, entries)

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

func (s *State) GetPendingEntryHashes() []interfaces.IHash {
	pLists := s.ProcessLists
	if pLists == nil {
		return nil
	}
	ht := pLists.State.GetHighestCompletedBlock()
	pl := pLists.Get(ht + 1)
	var hashCount int32
	hashCount = 0
	hashResponse := make([]interfaces.IHash, pl.LenNewEntries())
	keys := pl.GetKeysNewEntries()
	for _, k := range keys {
		entry := pl.GetNewEntry(k)
		hashResponse[hashCount] = entry.GetHash()
		hashCount++
	}
	return hashResponse
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

func (s *State) DatabaseContains(hash interfaces.IHash) bool {
	result, _, err := s.LoadDataByHash(hash)
	if result != nil && err == nil {
		return true
	}
	return false
}

func (s *State) MessageToLogString(msg interfaces.IMsg) string {
	bytes, err := msg.MarshalBinary()
	if err != nil {
		panic("Failed MarshalBinary: " + err.Error())
	}
	msgStr := hex.EncodeToString(bytes)

	answer := "\n" + msg.String() + "\n  " + s.ShortString() + "\n" + "\t\t\tMsgHex: " + msgStr + "\n"
	return answer
}

func (s *State) JournalMessage(msg interfaces.IMsg) {
	if s.Journaling && len(s.JournalFile) != 0 {
		f, err := os.OpenFile(s.JournalFile, os.O_APPEND+os.O_WRONLY, 0666)
		if err != nil {
			s.JournalFile = ""
			return
		}
		str := s.MessageToLogString(msg)
		f.WriteString(str)
		f.Close()
	}
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

	dbheight := s.GetHighestCompletedBlock()
	plbase := s.ProcessLists.DBHeightBase
	if dbheight == 0 {
		dbheight++
	}
	if dbheight > 1 {
		dbheight--
	}
	if plbase <= dbheight {
		if !s.Leader || s.RunLeader {
			progress = s.ProcessLists.UpdateState(dbheight)
		}
	}

	p2 := s.DBStates.UpdateState()
	progress = progress || p2

	s.catchupEBlocks()

	s.SetString()
	if s.ControlPanelDataRequest {
		s.CopyStateToControlPanel()
	}
	return
}

func (s *State) NoEntryYet(entryhash interfaces.IHash, ts interfaces.Timestamp) bool {
	_, ok := s.Replay.Valid(constants.REVEAL_REPLAY, entryhash.Fixed(), ts, s.GetTimestamp())
	return ok
}

func (s *State) catchupEBlocks() {
	now := s.GetTimestamp()

	// If we have no Entry Blocks in our queue, reset our timer.
	if len(s.MissingEntryBlocks) == 0 || s.MissingEntryBlockRepeat == nil {
		s.MissingEntryBlockRepeat = now
	}

	// If our delay has been reached, then ask for some missing Entry blocks
	// This is a replay, because sometimes requests are ignored or lost.
	if now.GetTimeSeconds()-s.MissingEntryBlockRepeat.GetTimeSeconds() > 5 {
		s.MissingEntryBlockRepeat = now

		for _, eb := range s.MissingEntryBlocks {
			eBlockRequest := messages.NewMissingData(s, eb.ebhash)
			s.NetworkOutMsgQueue() <- eBlockRequest
		}
	}

	if len(s.MissingEntries) == 0 {
		s.MissingEntryRepeat = nil
		s.EntryDBHeightComplete = s.EntryBlockDBHeightComplete
	} else {
		if s.MissingEntryRepeat == nil {
			s.MissingEntryRepeat = now
		}

		// If our delay has been reached, then ask for some missing Entry blocks
		// This is a replay, because sometimes requests are ignored or lost.
		if now.GetTimeSeconds()-s.MissingEntryRepeat.GetTimeSeconds() > 5 {
			s.MissingEntryRepeat = now

			fmt.Printf("dddd Missing Entry %10s #missing %d Processing %d Complete %d\n",
				s.FactomNodeName,
				len(s.MissingEntries),
				s.EntryDBHeightProcessing,
				s.EntryDBHeightComplete)

			for i, eb := range s.MissingEntries {
				if i > 200 {
					// Only send out 200 requests at a time.
					break
				}
				entryRequest := messages.NewMissingData(s, eb.entryhash)
				s.NetworkOutMsgQueue() <- entryRequest
			}
		}
	}
	// If we still have 10 that we are asking for, then let's not add to the list.
	if len(s.MissingEntryBlocks) < 10 {
		// While we have less than 20 that we are asking for, look for more to ask for.
		for s.EntryBlockDBHeightProcessing < s.GetHighestCompletedBlock() && len(s.MissingEntryBlocks) < 20 {
			db := s.GetDirectoryBlockByHeight(s.EntryBlockDBHeightProcessing)

			for i, ebKeyMR := range db.GetEntryHashes() {
				// The first three entries (0,1,2) in every directory block are blocks we already have by
				// definition.  If we decide to not have Factoid blocks or Entry Credit blocks in some cases,
				// then this assumption might not hold.  But it does for now.
				if i <= 2 {
					continue
				}

				// Ask for blocks we don't have.
				if !s.DatabaseContains(ebKeyMR) {
					s.MissingEntryBlocks = append(s.MissingEntryBlocks,
						MissingEntryBlock{ebhash: ebKeyMR, dbheight: s.EntryBlockDBHeightProcessing})
				} else {
					eblock, _ := s.DB.FetchEBlock(ebKeyMR)

					for _, entryhash := range eblock.GetEntryHashes() {
						if entryhash.IsMinuteMarker() {
							continue
						}
						e, _ := s.DB.FetchEntry(entryhash)
						if e == nil {
							var v struct {
								ebhash    interfaces.IHash
								entryhash interfaces.IHash
								dbheight  uint32
							}

							v.dbheight = eblock.GetHeader().GetDBHeight()
							v.entryhash = entryhash
							v.ebhash = ebKeyMR

							s.MissingEntries = append(s.MissingEntries, v)
						}
						// Save the entry hash, and remove from commits IF this hash is valid in this current timeframe.
						if s.Replay.IsHashUnique(constants.REVEAL_REPLAY, entryhash.Fixed()) {
							s.Replay.SetHashNow(constants.REVEAL_REPLAY, entryhash.Fixed(), now)
						} else {
							delete(s.Commits, entryhash.Fixed())
						}
					}
				}
			}
			if len(s.MissingEntries) == 0 && s.EntryBlockDBHeightComplete == s.EntryBlockDBHeightProcessing {
				s.EntryBlockDBHeightComplete++
			}
			s.EntryBlockDBHeightProcessing++
		}
	}

}

func (s *State) AddDBSig(dbheight uint32, chainID interfaces.IHash, sig interfaces.IFullSignature) {
	s.ProcessLists.Get(dbheight).AddDBSig(chainID, sig)
}

func (s *State) AddFedServer(dbheight uint32, hash interfaces.IHash) int {
	return s.ProcessLists.Get(dbheight).AddFedServer(hash)
}

func (s *State) TrimVMList(dbheight uint32, height uint32, vmIndex int) {
	s.ProcessLists.Get(dbheight).TrimVMList(height, vmIndex)
}

func (s *State) RemoveFedServer(dbheight uint32, hash interfaces.IHash) {
	s.ProcessLists.Get(dbheight).RemoveFedServerHash(hash)
}

func (s *State) AddAuditServer(dbheight uint32, hash interfaces.IHash) int {
	return s.ProcessLists.Get(dbheight).AddAuditServer(hash)
}

func (s *State) RemoveAuditServer(dbheight uint32, hash interfaces.IHash) {
	s.ProcessLists.Get(dbheight).RemoveAuditServerHash(hash)
}

func (s *State) GetFedServers(dbheight uint32) []interfaces.IFctServer {
	pl := s.ProcessLists.Get(dbheight)
	if pl != nil {
		return pl.FedServers
	}
	return nil
}

func (s *State) GetAuditServers(dbheight uint32) []interfaces.IFctServer {
	pl := s.ProcessLists.Get(dbheight)
	if pl != nil {
		return pl.AuditServers
	}
	return nil
}

func (s *State) GetOnlineAuditServers(dbheight uint32) []interfaces.IFctServer {
	allAuditServers := s.GetAuditServers(dbheight)
	var onlineAuditServers []interfaces.IFctServer

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

func (s *State) GetFactomdVersion() int {
	return s.FactomdVersion
}

func (s *State) initServerKeys() {
	var err error
	s.serverPrivKey, err = primitives.NewPrivateKeyFromHex(s.LocalServerPrivKey)
	if err != nil {
		//panic("Cannot parse Server Private Key from configuration file: " + err.Error())
	}
	s.serverPubKey = s.serverPrivKey.Pub
	//s.serverPubKey = primitives.PubKeyFromString(constants.SERVER_PUB_KEY)
}

func (s *State) LogInfo(args ...interface{}) {
	s.Logger.Info(args...)
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

func (s *State) GetPort() int {
	return s.PortNumber
}

func (s *State) TickerQueue() chan int {
	return s.tickerQueue
}

func (s *State) TimerMsgQueue() chan interfaces.IMsg {
	return s.timerMsgQueue
}

func (s *State) NetworkInvalidMsgQueue() chan interfaces.IMsg {
	return s.networkInvalidMsgQueue
}

func (s *State) NetworkOutMsgQueue() chan interfaces.IMsg {
	return s.networkOutMsgQueue
}

func (s *State) InMsgQueue() chan interfaces.IMsg {
	return s.inMsgQueue
}

func (s *State) APIQueue() chan interfaces.IMsg {
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

// Getting the cfg state for Factom doesn't force a read of the config file unless
// it hasn't been read yet.
func (s *State) GetCfg() interfaces.IFactomConfig {
	return s.Cfg
}

// ReadCfg forces a read of the factom config file.  However, it does not change the
// state of any cfg object held by other processes... Only what will be returned by
// future calls to Cfg().(s.Cfg.(*util.FactomdConfig)).String()
func (s *State) ReadCfg(filename string) interfaces.IFactomConfig {
	s.Cfg = util.ReadConfig(filename)
	return s.Cfg
}

func (s *State) GetNetworkNumber() int {
	return s.NetworkNumber
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

func (s *State) GetMatryoshka(dbheight uint32) interfaces.IHash {
	return nil
}

func (s *State) InitLevelDB() error {

	if s.DB != nil {
		return nil
	}

	path := s.LdbPath + "/" + s.Network + "/" + "factoid_level.db"

	s.Println("Database:", path)

	dbase, err := hybridDB.NewLevelMapHybridDB(path, false)

	if err != nil || dbase == nil {
		dbase, err = hybridDB.NewLevelMapHybridDB(path, true)
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
	dbase := hybridDB.NewBoltMapHybridDB(nil, path+"FactomBolt.db")
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
	str := fmt.Sprintf(" %7s %12s %12s %4s %6s %10s %8s %5s %4s %20s %12s %10s %-8s %-9s %15s %9s %s\n",
		"Node",
		"ID   ",
		" ",
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
		"SysHeight")

	return str
}

func (s *State) SetStringConsensus() {
	str := fmt.Sprintf("%10s[%x_%x] ", s.FactomNodeName, s.IdentityChainID.Bytes()[:3], s.IdentityChainID.Bytes()[3:6])

	s.serverPrt = str
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
			if list.AmINegotiator {
				N = "N"
			}
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
		d = s.DBStates.Get(int(s.GetHighestSavedBlock())).DirectoryBlock
		keyMR = d.GetKeyMR().Bytes()
		dHeight = d.GetHeader().GetDBHeight()
	}

	runtime := time.Since(s.starttime)
	shorttime := time.Since(s.lasttime)
	total := s.FactoidTrans + s.NewEntryChains + s.NewEntries
	tps := float64(total) / float64(runtime.Seconds())
	if shorttime > time.Second*3 {
		delta := (s.FactoidTrans + s.NewEntryChains + s.NewEntries) - s.transCnt
		s.tps = ((float64(delta) / float64(shorttime.Seconds())) + 2*s.tps) / 3
		s.lasttime = time.Now()
		s.transCnt = total // transactions accounted for
	}

	str := fmt.Sprintf("%7s[%12x] %4s%4s %2d.%01d%% %2d.%03d",
		s.FactomNodeName,
		s.IdentityChainID.Bytes()[:6],
		stype,
		vmIndex,
		s.DropRate/10, s.DropRate%10,
		s.Delay/1000, s.Delay%1000)

	pls := fmt.Sprintf("%d/%d/%d", s.ProcessLists.DBHeightBase, s.PLProcessHeight, s.GetTrueLeaderHeight())

	str = str + fmt.Sprintf(" %5d[%6x] %-11s ",
		dHeight,
		keyMR[:3],
		pls)

	dbstate := fmt.Sprintf("%d/%d/%d/%d", s.DBStateAskCnt, s.DBStateAnsCnt, s.DBStateReplyCnt, s.DBStateFailsCnt)
	missing := fmt.Sprintf("%d/%d/%d/%d", s.MissingRequestSendCnt, s.MissingRequestReplyCnt, s.MissingRequestIgnoreCnt, s.MissingResponseAppliedCnt)
	str = str + fmt.Sprintf(" %2s/%2d %15s %26s ",
		lmin,
		s.CurrentMinute,
		dbstate,
		missing)

	trans := fmt.Sprintf("%d/%d/%d", s.FactoidTrans, s.NewEntryChains, s.NewEntries-s.NewEntryChains)
	apis := fmt.Sprintf("%d/%d/%d", s.FCTSubmits, s.ECCommits, s.ECommits)
	stps := fmt.Sprintf("%3.2f/%3.2f", tps, s.tps)
	str = str + fmt.Sprintf(" %5d %5d %12s %15s %11s",
		s.ResendCnt,
		s.ExpireCnt,
		trans,
		apis,
		stps)

	str = str + fmt.Sprintf(" %d", list.System.Height)

	if list.System.Height < len(list.System.List) {
		ff, ok := list.System.List[list.System.Height].(*messages.FullServerFault)
		if ok {
			str = str + fmt.Sprintf(" VM:%d %s", int(ff.VMIndex), ff.AuditServerID.String()[6:10])
		} else {
			str = str + fmt.Sprintf(" VM:%s %s", "?", "-nil-")
		}
	} else {
		str = str + " -"
	}

	s.serverPrt = str
}

func (s *State) GetTrueLeaderHeight() uint32 {
	return uint32(int(s.ProcessLists.DBHeightBase) + len(s.ProcessLists.Lists) - 1)
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
