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
	"encoding/json"

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
	ed "github.com/FactomProject/ed25519"
	"github.com/FactomProject/btcutil/base58"

	"time"
	"sync"
)

var _ = fmt.Print

type State struct {
	filename string

	Cfg interfaces.IFactomConfig

	Prefix                  string
	FactomNodeName          string
	FactomdVersion          int
	LogPath                 string
	LdbPath                 string
	BoltDBPath              string
	LogLevel                string
	ConsoleLogLevel         string
	NodeMode                string
	DBType                  string
	CloneDBType             string
	ExportData              bool
	ExportDataSubpath       string
	Network                 string
	PeersFile               string
	LocalServerPrivKey      string
	DirectoryBlockInSeconds int
	PortNumber              int
	Replay                  *Replay
	InternalReplay          *Replay
	GreenFlg                bool
	GreenTimestamp          interfaces.Timestamp
	DropRate                int

	IdentityChainID interfaces.IHash // If this node has an identity, this is it

	// Just to print (so debugging doesn't drive functionaility)
	Status    bool
	starttime time.Time
	serverPrt string

	tickerQueue            chan int
	timerMsgQueue          chan interfaces.IMsg
	networkOutMsgQueue     chan interfaces.IMsg
	networkInvalidMsgQueue chan interfaces.IMsg
	inMsgQueue             chan interfaces.IMsg
	apiQueue               chan interfaces.IMsg
	ackQueue               chan interfaces.IMsg
	msgQueue               chan interfaces.IMsg
	OutOfOrders            []*messages.Ack
	StallAcks              []*messages.Ack
	ShutdownChan           chan int // For gracefully halting Factom
	JournalFile            string

	serverPrivKey primitives.PrivateKey
	serverPubKey  primitives.PublicKey

	// Server State
	LLeaderHeight   uint32
	Leader          bool
	LeaderVMIndex   int
	LeaderPL        *ProcessList
	OutputAllowed   bool
	LeaderMinute    int  // The minute that just was processed by the follower, (1-10), set with EOM
	EOM             int  // Set to true when all Process Lists have finished a minute
	NetStateOff     bool // Disable if true, Enable if false
	DebugConsensus  bool // If true, dump consensus trace
	FactoidTrans    int
	NewEntryChains  int
	NewEntries      int
	LeaderTimestamp uint64
	// Maps
	// ====
	// For Follower
	Holding map[[32]byte]interfaces.IMsg // Hold Messages
	XReview []interfaces.IMsg            // After the EOM, we must review the messages in Holding
	Acks    map[[32]byte]interfaces.IMsg // Hold Acknowledgemets
	Commits map[[32]byte]interfaces.IMsg // Commit Messages
	Reveals map[[32]byte]interfaces.IMsg // Reveal Messages

	InvalidMessages      map[[32]byte]interfaces.IMsg
	InvalidMessagesMutex sync.RWMutex

	AuditHeartBeats []interfaces.IMsg   // The checklist of HeartBeats for this period
	FedServerFaults [][]interfaces.IMsg // Keep a fault list for every server

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

	ProcessLists *ProcessLists

	// Factom State
	FactoidState    interfaces.IFactoidState
	NumTransactions int

	// Permanent balances from processing blocks.
	FactoidBalancesP      map[[32]byte]int64
	FactoidBalancesPMutex sync.Mutex
	ECBalancesP           map[[32]byte]int64
	ECBalancesPMutex      sync.Mutex

	// Temporary balances from updating transactions in real time.
	FactoidBalancesT      map[[32]byte]int64
	FactoidBalancesTMutex sync.Mutex
	ECBalancesT           map[[32]byte]int64
	ECBalancesTMutex      sync.Mutex

	// Web Services
	Port int

	//For Replay / journal
	IsReplaying     bool
	ReplayTimestamp interfaces.Timestamp

	// DBlock Height at which node has a complete set of eblocks+entries
	EBDBHeightComplete uint32

	// For dataRequests made by this node, which it's awaiting dataResponses for
	DataRequests map[[32]byte]interfaces.IHash

	LastPrint    string
	LastPrintCnt int

	// FER section
	FactoshisPerEC uint64
	FERChainId string
	ExchangeRateAuthorityAddress string

	FERChangeHeight uint32
	FERChangePrice uint64
	FERPriority uint32
	FERPrioritySetHeight uint32
}

var _ interfaces.IState = (*State)(nil)

func (s *State) Clone(number string) interfaces.IState {

	clone := new(State)

	clone.FactomNodeName = s.Prefix + "FNode" + number
	clone.FactomdVersion = s.FactomdVersion
	clone.LogPath = s.LogPath + "/Sim" + number
	clone.LdbPath = s.LdbPath + "/Sim" + number
	clone.JournalFile = s.LogPath + "/journal" + number + ".log"
	clone.BoltDBPath = s.BoltDBPath + "/Sim" + number
	clone.LogLevel = s.LogLevel
	clone.ConsoleLogLevel = s.ConsoleLogLevel
	clone.NodeMode = "FULL"
	clone.CloneDBType = s.CloneDBType
	clone.DBType = s.CloneDBType
	clone.ExportData = s.ExportData
	clone.ExportDataSubpath = s.ExportDataSubpath + "sim-" + number
	clone.Network = s.Network
	clone.PeersFile = s.PeersFile
	clone.DirectoryBlockInSeconds = s.DirectoryBlockInSeconds
	clone.PortNumber = s.PortNumber

	clone.IdentityChainID = primitives.Sha([]byte(clone.FactomNodeName))

	//generate and use a new deterministic PrivateKey for this clone
	shaHashOfNodeName := primitives.Sha([]byte(clone.FactomNodeName)) //seed the private key with node name
	clonePrivateKey := primitives.NewPrivateKeyFromHexBytes(shaHashOfNodeName.Bytes())
	clone.LocalServerPrivKey = clonePrivateKey.PrivateKeyString()

	//serverPrivKey primitives.PrivateKey
	//serverPubKey  primitives.PublicKey

	clone.FactoshisPerEC = s.FactoshisPerEC

	clone.Port = s.Port

	return clone
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

// TODO JAYJAY BUGBUG- passing in folder here is a hack for multiple factomd processes on a single machine (sharing a single .factom)
func (s *State) LoadConfig(filename string, folder string) {

	s.FactomNodeName = s.Prefix + "FNode0" // Default Factom Node Name for Simulation
	if len(filename) > 0 {
		s.filename = filename
		s.ReadCfg(filename, folder)

		// Get our factomd configuration information.
		cfg := s.GetCfg().(*util.FactomdConfig)

		s.LogPath = cfg.Log.LogPath + s.Prefix
		s.LdbPath = cfg.App.LdbPath + s.Prefix
		s.BoltDBPath = cfg.App.BoltDBPath + s.Prefix
		s.LogLevel = cfg.Log.LogLevel
		s.ConsoleLogLevel = cfg.Log.ConsoleLogLevel
		s.NodeMode = cfg.App.NodeMode
		s.DBType = cfg.App.DBType
		s.ExportData = cfg.App.ExportData // bool
		s.ExportDataSubpath = cfg.App.ExportDataSubpath
		s.Network = cfg.App.Network
		s.PeersFile = cfg.App.PeersFile
		s.LocalServerPrivKey = cfg.App.LocalServerPrivKey
		s.FactoshisPerEC = cfg.App.ExchangeRate
		s.DirectoryBlockInSeconds = cfg.App.DirectoryBlockInSeconds
		s.PortNumber = cfg.Wsapi.PortNumber
		s.FERChainId = cfg.App.ExchangeRateChainId
		s.ExchangeRateAuthorityAddress = cfg.App.ExchangeRateAuthorityAddress

		// TODO:  Actually load the IdentityChainID from the config file
		s.IdentityChainID = primitives.Sha([]byte(s.FactomNodeName))
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
		s.PeersFile = "peers.json"
		s.LocalServerPrivKey = "4c38c72fc5cdad68f13b74674d3ffb1f3d63a112710868c9b08946553448d26d"
		s.FactoshisPerEC = 006666
		s.FERChainId = "eac57815972c504ec5ae3f9e5c1fe12321a3c8c78def62528fb74cf7af5e7389"
		s.ExchangeRateAuthorityAddress = ""   // default to nothing so that there is no default FER manipulation
		s.DirectoryBlockInSeconds = 6
		s.PortNumber = 8088

		// TODO:  Actually load the IdentityChainID from the config file
		s.IdentityChainID = primitives.Sha([]byte(s.FactomNodeName))

	}
	s.JournalFile = s.LogPath + "/journal0" + ".log"
}

func (s *State) Init() {

	wsapi.InitLogs(s.LogPath+s.FactomNodeName+".log", s.LogLevel)

	s.Logger = logger.NewLogFromConfig(s.LogPath, s.LogLevel, "State")

	log.SetLevel(s.ConsoleLogLevel)

	s.tickerQueue = make(chan int, 10000)                        //ticks from a clock
	s.timerMsgQueue = make(chan interfaces.IMsg, 10000)          //incoming eom notifications, used by leaders
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
		fmt.Println("Could not create "+s.LogPath+"\n error: "+er.Error())
	}
	_, err := os.Create(s.JournalFile) //Create the Journal File
	if err != nil {
		fmt.Println("Could not create the file: " + s.JournalFile)
		s.JournalFile = ""
	}
	// Set up struct to stop replay attacks
	s.Replay = new(Replay)
	s.InternalReplay = new(Replay)

	// Set up maps for the followers
	s.Holding = make(map[[32]byte]interfaces.IMsg)
	s.Acks = make(map[[32]byte]interfaces.IMsg)
	s.Commits = make(map[[32]byte]interfaces.IMsg)
	s.Reveals = make(map[[32]byte]interfaces.IMsg)

	// Setup the FactoidState and Validation Service that holds factoid and entry credit balances
	s.FactoidBalancesP = map[[32]byte]int64{}
	s.ECBalancesP = map[[32]byte]int64{}
	s.FactoidBalancesT = map[[32]byte]int64{}
	s.ECBalancesT = map[[32]byte]int64{}

	fs := new(FactoidState)
	fs.State = s
	s.FactoidState = fs

	// Allocate the original set of Process Lists
	s.ProcessLists = NewProcessLists(s)

	s.FactomdVersion = constants.FACTOMD_VERSION

	s.DBStates = new(DBStateList)
	s.DBStates.State = s
	s.DBStates.DBStates = make([]*DBState, 0)

	s.EBDBHeightComplete = 0
	s.DataRequests = make(map[[32]byte]interfaces.IHash)

	switch s.NodeMode {
	case "FULL":
		s.Leader = false
		s.Println("\n   +---------------------------+")
		s.Println("   +------ Follower Only ------+")
		s.Println("   +---------------------------+\n")
	case "SERVER":
		s.Println("\n   +-------------------------+")
		s.Println("   |       Leader Node       |")
		s.Println("   +-------------------------+\n")
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
	s.Println("\nExchange rate chain id set to ", s.GetFactoshisPerEC())
	s.Println("\nExchange rate Authority Public Key set to ", s.ExchangeRateAuthorityAddress)

	s.AuditHeartBeats = make([]interfaces.IMsg, 0)
	s.FedServerFaults = make([][]interfaces.IMsg, 0)

	s.initServerKeys()

	s.starttime = time.Now()
}

func (s *State) AddDataRequest(requestedHash, missingDataHash interfaces.IHash) {
	s.DataRequests[requestedHash.Fixed()] = missingDataHash
}

func (s *State) HasDataRequest(checkHash interfaces.IHash) bool {
	if _, ok := s.DataRequests[checkHash.Fixed()]; ok {
		return true
	}
	return false
}

func (s *State) GetEBDBHeightComplete() uint32 {
	return s.EBDBHeightComplete
}

func (s *State) SetEBDBHeightComplete(newHeight uint32) {
	s.EBDBHeightComplete = newHeight
}

func (s *State) GetEBlockKeyMRFromEntryHash(entryHash interfaces.IHash) interfaces.IHash {

	entry, err := s.DB.FetchEntryByHash(entryHash)
	if err != nil {
		return nil
	}
	if entry != nil {
		dblock := s.GetDirectoryBlockByHeight(entry.GetDatabaseHeight())
		for idx, ebHash := range dblock.GetEntryHashes() {
			if idx > 2 {
				thisBlock, err := s.DB.FetchEBlockByKeyMR(ebHash)
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
	ablk, err := s.DB.FetchABlockByKeyMR(dblk.GetDBEntries()[0].GetKeyMR())
	if err != nil {
		return nil, err
	}
	if ablk == nil {
		return nil, fmt.Errorf("ABlock not found")
	}
	ecblk, err := s.DB.FetchECBlockByHash(dblk.GetDBEntries()[1].GetKeyMR())
	if err != nil {
		return nil, err
	}
	if ecblk == nil {
		return nil, fmt.Errorf("ECBlock not found")
	}
	fblk, err := s.DB.FetchFBlockByKeyMR(dblk.GetDBEntries()[2].GetKeyMR())
	if err != nil {
		return nil, err
	}
	if fblk == nil {
		return nil, fmt.Errorf("FBlock not found")
	}
	if bytes.Compare(fblk.GetKeyMR().Bytes(), dblk.GetDBEntries()[2].GetKeyMR().Bytes()) != 0 {
		panic("Should not happen")
	}

	msg := messages.NewDBStateMsg(s.GetTimestamp(), dblk, ablk, fblk, ecblk)

	return msg, nil

}

func (s *State) LoadDataByHash(requestedHash interfaces.IHash) (interfaces.BinaryMarshallable, int, error) {
	if requestedHash == nil {
		return nil, -1, fmt.Errorf("Requested hash must be non-empty")
	}

	var result interfaces.BinaryMarshallable
	var err error

	// Check for Entry
	result, err = s.DB.FetchEntryByHash(requestedHash)
	if result != nil && err == nil {
		return result, 0, nil
	}

	// Check for Entry Block
	result, err = s.DB.FetchEBlockByKeyMR(requestedHash)
	if result != nil && err == nil {
		return result, 1, nil
	}
	result, _ = s.DB.FetchEBlockByHash(requestedHash)
	if result != nil && err == nil {
		return result, 1, nil
	}

	return nil, -1, nil
}

func (s *State) LoadSpecificMsg(dbheight uint32, vm int, plistheight uint32) (interfaces.IMsg, error) {
	if dbheight < s.ProcessLists.DBHeightBase {
		return nil, fmt.Errorf("Missing message is too deeply buried in blocks")
	} else if dbheight > (s.ProcessLists.DBHeightBase + uint32(len(s.ProcessLists.Lists))) {
		return nil, fmt.Errorf("Answering node has not reached DBHeight of missing message")
	}

	procList := s.ProcessLists.Get(dbheight)
	if procList == nil {
		return nil, fmt.Errorf("Nil Process List")
	}
	if len(procList.VMs[vm].List) < int(plistheight)+1 {
		return nil, fmt.Errorf("Process List too small (lacks requested msg)")
	}

	msg := procList.VMs[vm].List[plistheight]

	if msg == nil {
		return nil, fmt.Errorf("State process list does not include requested message")
	}

	return msg, nil
}

func (s *State) LoadSpecificMsgAndAck(dbheight uint32, vm int, plistheight uint32) (interfaces.IMsg, interfaces.IMsg, error) {
	if dbheight < s.ProcessLists.DBHeightBase {
		return nil, nil, fmt.Errorf("Missing message is too deeply buried in blocks")
	} else if dbheight > (s.ProcessLists.DBHeightBase + uint32(len(s.ProcessLists.Lists))) {
		return nil, nil, fmt.Errorf("Answering node has not reached DBHeight of missing message")
	}

	procList := s.ProcessLists.Get(dbheight)
	if procList == nil {
		return nil, nil, fmt.Errorf("Nil Process List")
	} else if len(procList.VMs) < 1 {
		return nil, nil, fmt.Errorf("No servers?")
	}
	if len(procList.VMs[vm].List) < int(plistheight)+1 {
		return nil, nil, fmt.Errorf("Process List too small (lacks requested msg)")
	}

	msg := procList.VMs[vm].List[plistheight]

	if msg == nil {
		return nil, nil, fmt.Errorf("State process list does not include requested message")
	}

	ackMsg, ok := s.ProcessLists.Get(dbheight).OldAcks[msg.GetHash().Fixed()]

	if !ok || ackMsg == nil {
		return nil, nil, fmt.Errorf("State process list does not include ack for message")
	}

	return msg, ackMsg, nil
}

// This will issue missingData requests for each entryHash in a particular EBlock
// that is not already saved to the database or requested already.
// It returns True if the EBlock is complete (all entries already exist in database)
func (s *State) GetAllEntries(ebKeyMR interfaces.IHash) bool {
	hasAllEntries := true
	eblock, err := s.DB.FetchEBlockByKeyMR(ebKeyMR)
	if err != nil {
		return false
	}
	if eblock == nil {
		if !s.HasDataRequest(ebKeyMR) {
			eBlockRequest := messages.NewMissingData(s, ebKeyMR)
			s.NetworkOutMsgQueue() <- eBlockRequest
		}
		return false
	}
	for _, entryHash := range eblock.GetEntryHashes() {
		if !strings.HasPrefix(entryHash.String(), "000000000000000000000000000000000000000000000000000000000000000") {
			if !s.DatabaseContains(entryHash) {
				hasAllEntries = false
			} else {
				continue
			}
			if !s.HasDataRequest(entryHash) {
				entryRequest := messages.NewMissingData(s, entryHash)
				s.NetworkOutMsgQueue() <- entryRequest
			}
		}
	}

	return hasAllEntries
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
	if len(s.JournalFile) != 0 {
		fmt.Println("Journal file:",s.JournalFile)
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
	return s.DBStates.Get(height)
}

// Return the Directory block if it is in memory, or hit the database if it must
// be loaded.
func (s *State) GetDirectoryBlockByHeight(height uint32) interfaces.IDirectoryBlock {
	dbstate := s.DBStates.Get(height)
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

	process := func(a *messages.Ack) {
		if _, ok := s.InternalReplay.Valid(a.GetHash().Fixed(), int64(a.GetTimestamp()), int64(s.GetTimestamp())); a.DBHeight < s.LLeaderHeight || !ok {
			delete(s.Holding, a.GetHash().Fixed())
			delete(s.Acks, a.GetHash().Fixed())
			return
		}
		s.ProcessLists.Get(a.DBHeight)
		s.Acks[a.GetHash().Fixed()] = a
		m := s.Holding[a.GetHash().Fixed()]
		if m != nil {
			pl := s.ProcessLists.Get(a.DBHeight)
			if pl != nil {
				pl.AddToProcessList(a, m)
			}
		}
	}

	// Look at all the other out of orders.  Note that if we kept this list sorted,
	// this would be really efficent, and wouldn't require a loop.
	end := len(s.OutOfOrders)
	for i := 0; i < end; i++ {
		a := s.GetOutOfOrder(0)
		if a != nil {
			if s.DebugConsensus {
				fmt.Println("dddd Out of Order Processing", s.FactomNodeName, end, i, a.String())
			}
			process(a)
		}
	}
	// Look at all the other out of orders.  Note that if we kept this list sorted,
	// this would be really efficent, and wouldn't require a loop.
	end = len(s.StallAcks)
	for i := 0; i < end; i++ {
		a := s.GetStalledAck(0)
		if a != nil {
			if s.DebugConsensus {
				fmt.Println("dddd Stall Ack             ", s.FactomNodeName, end, i, a.String())
			}
			process(a)
		}
	}

	for k := range s.Holding {
		m := s.Holding[k]
		if _, ok := s.InternalReplay.Valid(k, int64(m.GetTimestamp()), int64(s.GetTimestamp())); !ok {
			delete(s.Holding, k)
			delete(s.Acks, k)
		}
	}

	// Look at all the other out of orders.  Note that if we kept this list sorted,
	// this would be really efficent, and wouldn't require a loop.
	end = len(s.Acks)
	i := 0
	for k := range s.Acks {
		a, _ := s.Acks[k].(*messages.Ack)
		i++
		if i >= end {
			break
		}
		if a != nil {
			process(a)
		}
	}

	sort := func(msgs []*messages.Ack) {
		for i := 0; i < len(msgs)-1; i++ {
			for j := 0; j < len(msgs)-1-i; j++ {
				dbht1 := msgs[j].DBHeight > msgs[j+1].DBHeight
				dbht2 := msgs[j].DBHeight == msgs[j+1].DBHeight
				ht := msgs[j].Height > msgs[j+1].Height
				if dbht1 || (dbht2 && ht) {
					hld := msgs[j]
					msgs[j] = msgs[j+1]
					msgs[j+1] = hld
				}
			}
		}
	}

	sort(s.OutOfOrders)
	sort(s.StallAcks)

	dbheight := s.GetHighestRecordedBlock()
	plbase := s.ProcessLists.DBHeightBase
	if plbase <= dbheight+1 {
		progress = s.ProcessLists.UpdateState(dbheight + 1)
	}

	p2 := s.DBStates.UpdateState()
	progress = progress || p2

	s.catchupEBlocks()

	s.SetString()
	return
}

func (s *State) catchupEBlocks() {
	isComplete := true
	if s.GetEBDBHeightComplete() < s.GetDBHeightComplete() {
		dblockGathering := s.GetDirectoryBlockByHeight(s.GetEBDBHeightComplete())
		for idx, ebKeyMR := range dblockGathering.GetEntryHashes() {
			if idx > 2 {
				if s.DatabaseContains(ebKeyMR) {
					if !s.GetAllEntries(ebKeyMR) {
						isComplete = false
					}
				} else {
					isComplete = false
					if !s.HasDataRequest(ebKeyMR) {
						eBlockRequest := messages.NewMissingData(s, ebKeyMR)
						s.NetworkOutMsgQueue() <- eBlockRequest
					}
				}
			}
		}
		if isComplete {
			s.SetEBDBHeightComplete(s.GetEBDBHeightComplete() + 1)
		}
	}
}

// Go through the factoid exchange rate chain and determine if an FER change should be scheduled
func (s *State) ProcessRecentFERChainEntries() {

	// Find the FER entry chain
	FERChainHash, err := primitives.HexToHash(s.FERChainId)
	if err != nil {
		s.Println("The FERChainId couldn't be turned into a IHASH")
		return
	}

	//  Get the first eblock from the FERChain
	entryBlock, err := s.DB.FetchEBlockHead(FERChainHash)
	if err != nil {
		s.Println("Couldn't find the FER chain for id ", s.FERChainId)
		return
	}
	if entryBlock == nil {
		s.Println("FER Chain head found to be nil")
		return
	}

	s.Println("Checking last e block of FER chain with height of: ", entryBlock.GetHeader().GetDBHeight())
	s.Println("Current block height: ", s.GetDBHeightComplete())
	s.Println("BEFORE processing recent block: ")
	s.Println("    FERChangePrice: ", s.FERChangePrice)
	s.Println("    FERChangeHeight: ", s.FERChangeHeight)
	s.Println("    FERPriority: ", s.FERPriority)
	s.Println("    FERPrioritySetHeight: ", s.FERPrioritySetHeight)
	s.Println("    FER current: ", s.GetFactoshisPerEC())

	// Check to see if a price change targets the next block
	if (s.FERChangeHeight == (s.GetDBHeightComplete()+1)) {
		s.FactoshisPerEC = s.FERChangePrice
		s.FERChangePrice = 100000000
		s.FERChangeHeight = 0
	}

	// Check for the need to clear the priority
	// (s.GetDBHeightComplete() >= 12) is import because height is a uint and can't break logic if subtracted into false sub-zero
	if ( 	(s.GetDBHeightComplete() >= 12) &&
	(s.GetDBHeightComplete() - 12) >= s.FERPrioritySetHeight) {
		s.FERPrioritySetHeight = 0
		s.FERPriority = 0
		// Now the next entry to come through with a priority of 1 or more will be considered
	}


	// Check last entry block method
	if (entryBlock.GetHeader().GetDBHeight() == s.GetDBHeightComplete()) {
		entryHashes := entryBlock.GetEntryHashes()

		// s.Println("Found FER entry hashes in a block as: ", entryHashes)
		// create a map of possible minute markers that may be found in the EBlock Body
		mins := make(map[string]uint8)
		for i := byte(0); i <= 10; i++ {
			h := make([]byte, 32)
			h[len(h)-1] = i
			mins[hex.EncodeToString(h)] = i
		}

		// Loop through the hashes from the last blocks FER entries and evaluate them individually
		for _, entryHash := range entryHashes {

			// if this entryhash is a minute mark then continue
			if _, exist := mins[entryHash.String()]; exist {
				continue
			}

			// Make sure the entry exists
			anEntry, err := s.DB.FetchEntryByHash(entryHash)
			if (err != nil) {
				s.Println("Error during FetchEntryByHash: ", err)
				continue
			}
			if (anEntry == nil) {
				s.Println("Nil entry during FetchEntryByHash: ", entryHash)
				continue
			}

			if ( ! s.ExchangeRateAuthorityIsValid(anEntry) ) {
				s.Println("Skipping non-authority FER chain entry", entryHash)
				continue
			}

			entryContent := anEntry.GetContent()
			// s.Println("Found content of an FER entry is:  ", string(entryContent))
			anFEREntry := new(FEREntry)
			err = json.Unmarshal(entryContent, &anFEREntry)
			if err != nil {
				s.Println("A FEREntry messgae didn't unmarshall correctly: ", err)
				continue
			}

			// Set it's resident height for validity checking
			anFEREntry.SetResidentHeight(s.GetDBHeightComplete())

			if ((s.FerEntryIsValid(anFEREntry)) && (anFEREntry.Priority > s.FERPriority)) {

				fmt.Println(" Processing FER entry : ", string(entryContent))
				s.FERPriority = anFEREntry.GetPriority()
				s.FERPrioritySetHeight = s.GetDBHeightComplete()
				s.FERChangePrice = anFEREntry.GetTargetPrice()
				s.FERChangeHeight = anFEREntry.GetTargetActivationHeight()

				// Adjust the target if needed
				if ( s.FERChangeHeight < (s.GetDBHeightComplete() + 2)) {
					s.FERChangeHeight = s.GetDBHeightComplete() + 2
				}
			} else {
				fmt.Println(" Failed FER entry : ", string(entryContent))
			}
		}
	}

	s.Println("AFTER processing recent block: ")
	s.Println("    FERChangePrice: ", s.FERChangePrice)
	s.Println("    FERChangeHeight: ", s.FERChangeHeight)
	s.Println("    FERPriority: ", s.FERPriority)
	s.Println("    FERPrioritySetHeight: ", s.FERPrioritySetHeight)
	s.Println("    FER current: ", s.GetFactoshisPerEC())
	s.Println("----------------------------------")

	return
}

func (s *State) ExchangeRateAuthorityIsValid(e interfaces.IEBEntry) bool {

	// convert the conf quthority address into a
	authorityAddress := base58.Decode(s.ExchangeRateAuthorityAddress)
	ecPubPrefix := []byte{0x59, 0x2a}

	if !bytes.Equal(authorityAddress[:2], ecPubPrefix) {
		fmt.Errorf("Invalid Entry Credit Private Address")
		return false
	}

	pub := new([32]byte)
	copy(pub[:], authorityAddress[2:34])

	// in case verify can't handle empty public key
	if (s.ExchangeRateAuthorityAddress == "") {
		return false
	}
	sig := new([64]byte)
	externalIds := e.ExternalIDs()


	// check for number of ext ids
	if len(externalIds) < 1 {
		return false
	}

	copy(sig[:], externalIds[0])   // First ext id needs to be the signature of the content

	if !ed.Verify(pub, e.GetContent(), sig) {
		return false
	}

	return true
}

func (s *State) FerEntryIsValid(passedFEREntry interfaces.IFEREntry) bool {
	// fail if expired
	if (passedFEREntry.GetExpirationHeight() < passedFEREntry.GetResidentHeight()) {
		return false
	}

	// fail if expired height is too far out
	if (passedFEREntry.GetExpirationHeight() > (passedFEREntry.GetResidentHeight() + 6)) {
		return false
	}

	// fail if target is out of range of the expire height
	// The check for expire height >= 6 is import because a lower value results in a uint binary wrap to near maxint, cracks logic
	if (	(passedFEREntry.GetTargetActivationHeight() > (passedFEREntry.GetExpirationHeight() + 6)) ||
	( 	(passedFEREntry.GetExpirationHeight() >= 6) &&
	(passedFEREntry.GetTargetActivationHeight() < (passedFEREntry.GetExpirationHeight() - 6) ))) {

		return false
	}

	return true
}


// Returns the higher of the current factoid exchange rate and what it knows will change in the future
// TODO: this needs to look in the current block being formed and in the process list messages
func (s *State) GetPredictiveFER() (uint64) {
	if (s.FERChangePrice != 0 ) {
		if (s.FERChangePrice > s.GetFactoshisPerEC()) {
			return s.FERChangePrice
		}
	}

	return s.GetFactoshisPerEC()
}

func (s *State) GetEOM() int {
	return s.EOM
}

func (s *State) AddFedServer(dbheight uint32, hash interfaces.IHash) int {
	return s.ProcessLists.Get(dbheight).AddFedServer(hash)
}

func (s *State) AddAuditServer(dbheight uint32, hash interfaces.IHash) int {
	return s.ProcessLists.Get(dbheight).AddAuditServer(hash)
}

func (s *State) GetFedServers(dbheight uint32) []interfaces.IFctServer {
	return s.ProcessLists.Get(dbheight).FedServers
}

func (s *State) GetAuditServers(dbheight uint32) []interfaces.IFctServer {
	return s.ProcessLists.Get(dbheight).AuditServers
}

func (s *State) IsLeader() bool {
	return s.Leader
}

func (s *State) GetVirtualServers(dbheight uint32, minute int, identityChainID interfaces.IHash) (found bool, index int) {
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

func (s *State) GetServerPrivateKey() primitives.PrivateKey {
	return s.serverPrivKey
}

func (s *State) GetServerPublicKey() primitives.PublicKey {
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
	s.serverPubKey = primitives.PubKeyFromString(constants.SERVER_PUB_KEY)
}

func (s *State) LogInfo(args ...interface{}) {
	s.Logger.Info(args...)
}

func (s *State) GetAuditHeartBeats() []interfaces.IMsg {
	return s.AuditHeartBeats
}

func (s *State) GetFedServerFaults() [][]interfaces.IMsg {
	return s.FedServerFaults
}

func (s *State) SetIsReplaying() {
	s.IsReplaying = true
}

func (s *State) SetIsDoneReplaying() {
	s.IsReplaying = false
	s.ReplayTimestamp = 0
}

func (s *State) GetTimestamp() interfaces.Timestamp {
	if s.IsReplaying == true {
		return s.ReplayTimestamp
	}
	return *interfaces.NewTimeStampNow()
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

func (s *State) StallAck(ack *messages.Ack) {
	if ack.IsStalled() {
		return
	}
	ack.SetStall(true)
	s.StallAcks = append(s.StallAcks, ack)
}

// Get the ith message out of the stall queue.  Note getting i=0 makes
// the stall queue into a FIFO, but other options are possible.
func (s *State) GetStalledAck(i int) *messages.Ack {
	if len(s.StallAcks) == 0 || i >= len(s.StallAcks) {
		return nil
	}
	m := s.StallAcks[i]

	copy(s.StallAcks[i:], s.StallAcks[i+1:])
	s.StallAcks[len(s.StallAcks)-1] = nil
	s.StallAcks = s.StallAcks[:len(s.StallAcks)-1]
	m.SetStall(false)
	return m
}

func (s *State) OutOfOrderAck(ack *messages.Ack) {
	if ack.IsStalled() {
		return
	}
	ack.SetStall(true)
	s.OutOfOrders = append(s.OutOfOrders, ack)
}

// Get the ith message out of the stall queue.  Note getting i=0 makes
// the stall queue into a FIFO, but other options are possible.
func (s *State) GetOutOfOrder(i int) *messages.Ack {
	if len(s.OutOfOrders) == 0 || i >= len(s.OutOfOrders) {
		return nil
	}
	m := s.OutOfOrders[i]

	copy(s.OutOfOrders[i:], s.OutOfOrders[i+1:])
	s.OutOfOrders[len(s.OutOfOrders)-1] = nil
	s.OutOfOrders = s.OutOfOrders[:len(s.OutOfOrders)-1]
	m.SetStall(false)
	return m
}

func (s *State) MsgQueue() chan interfaces.IMsg {
	return s.msgQueue
}

func (s *State) GetLeaderTimestamp() uint64 {
	return s.LeaderTimestamp
}

func (s *State) SetLeaderTimestamp(ts uint64) {
	s.LeaderTimestamp = ts
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
func (s *State) ReadCfg(filename string, folder string) interfaces.IFactomConfig {
	s.Cfg = util.ReadConfig(filename, folder)
	return s.Cfg
}

func (s *State) GetNetworkNumber() int {
	return s.NetworkNumber
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

	if !s.Status {
		return
	}
	s.Status = false

	buildingBlock := s.GetHighestRecordedBlock()

	found, _ := s.GetVirtualServers(buildingBlock+1, 0, s.GetIdentityChainID())

	L := ""
	X := ""
	W := ""
	if found {
		L = "L"
	}
	if s.NetStateOff {
		X = "X"
	}
	if !s.GreenFlg {
		W = "W"
	}

	stype := fmt.Sprintf("%1s%1s%1s", L, X, W)

	keyMR := primitives.NewZeroHash().Bytes()
	//abHash := []byte("aaaaa")
	//fbHash := []byte("aaaaa")
	//ecHash := []byte("aaaaa")

	switch {
	case s.DBStates == nil:

	case s.DBStates.Last() == nil:

	case s.DBStates.Last().DirectoryBlock == nil:

	default:
		if s.DBStates.Last().DirectoryBlock.GetHeader().GetDBHeight() > 0 {
			keyMR = s.DBStates.Last().DirectoryBlock.GetKeyMR().Bytes()
		}
	}

	runtime := time.Since(s.starttime)
	tps := float64(s.FactoidTrans+s.NewEntryChains+s.NewEntries)/float64(runtime.Seconds())

	s.serverPrt = fmt.Sprintf("%8s[%6x]%4s Save: %d[%6x] PL:%d/%d Min: %2v DBHT %v Min C/F %02v/%02v EOM %2v %3d-Fct %3d-EC %3d-E  %7.2f tps",
		s.FactomNodeName,
		s.IdentityChainID.Bytes()[:3],
		stype,
		s.GetHighestRecordedBlock(),
		keyMR[:3],
		s.ProcessLists.DBHeightBase,
		int(s.ProcessLists.DBHeightBase)+len(s.ProcessLists.Lists)-1,
		s.LeaderMinute,
		s.LLeaderHeight,
		s.ProcessLists.Get(s.LLeaderHeight).MinuteComplete(),
		s.ProcessLists.Get(s.LLeaderHeight).MinuteFinished(),
		s.EOM,
		s.FactoidTrans,
		s.NewEntryChains,
		s.NewEntries,
		tps)


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
