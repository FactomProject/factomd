package state

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/FactomProject/factomd/anchor"
	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/factoid/block"
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
	Cfg interfaces.IFactomConfig

	IdentityChainID interfaces.IHash // If this node has an identity, this is it

	networkInMsgQueue      chan interfaces.IMsg
	networkOutMsgQueue     chan interfaces.IMsg
	networkInvalidMsgQueue chan interfaces.IMsg
	inMsgQueue             chan interfaces.IMsg
	leaderInMsgQueue       chan interfaces.IMsg
	followerInMsgQueue     chan interfaces.IMsg

	myServer      interfaces.IServer //the server running on this Federated Server
	serverPrivKey primitives.PrivateKey
	serverPubKey  primitives.PublicKey
	totalServers  int
	serverState   int
	// Maps
	// ====
	// For Follower
	Holding map[[32]byte]interfaces.IMsg // Hold Messages
	Acks    map[[32]byte]interfaces.IMsg // Hold Acknowledgemets

	AuditHeartBeats []interfaces.IMsg   // The checklist of HeartBeats for this period
	FedServerFaults [][]interfaces.IMsg // Keep a fault list for every server

	//Network MAIN = 0, TEST = 1, LOCAL = 2, CUSTOM = 3
	NetworkNumber int // Encoded into Directory Blocks(s.Cfg.(*util.FactomdConfig)).String()

	// Number of Servers acknowledged by Factom
	Matryoshka []interfaces.IHash // Reverse Hash

	// Database
	DB     *databaseOverlay.Overlay
	Logger *logger.FLogger
	Anchor interfaces.IAnchor

	// Directory Block State
	_DBStatesMultex   *sync.Mutex
	_DBHeightComplete uint32     // Holds the DBHeight of the last DirectoryBlock processed
	_DBHeight         uint32     // Holds the index of the DirectoryBlock under construction.
	_DBStates         []*DBState // Holds all DBStates not yet processed.
	_LastDBState      *DBState

	// Having all the state for a particular directory block stored in one structure
	// makes creating the next state, updating the various states, and setting up the next
	// state much more simple.
	//
	// Functions that provide state information take a dbheight param.  I use the current
	// DBHeight to ensure that I return the proper information for the right directory block
	// height, even if it changed out from under the calling code.
	//
	// Process list previous [0], present(@DBHeight) [1], and future (@DBHeight+1) [2]
	_ProcessListsMultex *sync.Mutex
	_ProcessLists       []*ProcessList
	_ProcessListBase    uint32

	// Factom State
	FactoidState interfaces.IFactoidState

	// Web Services
	Port int

	// Message State
	LastAck interfaces.IMsg // The last Acknowledgement set by this server

}

var _ interfaces.IState = (*State)(nil)

func (s *State) Init(filename string) {

	s.ReadCfg(filename)
	// Get our factomd configuration information.
	cfg := s.GetCfg().(*util.FactomdConfig)

	wsapi.InitLogs(cfg.Log.LogPath, cfg.Log.LogLevel)

	s.Logger = logger.NewLogFromConfig(cfg.Log.LogPath, cfg.Log.LogLevel, "State")

	log.SetLevel(cfg.Log.ConsoleLogLevel)

	s.networkInMsgQueue = make(chan interfaces.IMsg, 10000)      //incoming message queue from the network messages
	s.networkInvalidMsgQueue = make(chan interfaces.IMsg, 10000) //incoming message queue from the network messages
	s.networkOutMsgQueue = make(chan interfaces.IMsg, 10000)     //Messages to be broadcast to the network
	s.inMsgQueue = make(chan interfaces.IMsg, 10000)             //incoming message queue for factom application messages
	s.leaderInMsgQueue = make(chan interfaces.IMsg, 10000)       //Leader Messages
	s.followerInMsgQueue = make(chan interfaces.IMsg, 10000)     //Follower Messages

	// Set up maps for the followers
	s.Holding = make(map[[32]byte]interfaces.IMsg)
	s.Acks = make(map[[32]byte]interfaces.IMsg)

	// Setup the FactoidState and Validation Service that holds factoid and entry credit balances
	fs := new(FactoidState)
	fs.ValidationService = NewValidationService()
	s.FactoidState = fs

	// Allocate the original set of Process Lists
	s._ProcessListsMultex = new(sync.Mutex)
	s._ProcessLists = make([]*ProcessList, 0)

	s._DBStatesMultex = new(sync.Mutex)
	s._DBStates = make([]*DBState, 0)

	s.totalServers = 1
	switch cfg.App.NodeMode {
	case "FULL":
		s.serverState = 0
		fmt.Println("\n   +---------------------------+")
		fmt.Println("   +------ Follower Only ------+")
		fmt.Println("   +---------------------------+\n")
	case "SERVER":
		s.serverState = 1
		fmt.Println("\n   +-------------------------+")
		fmt.Println("   |       Leader Node       |")
		fmt.Println("   +-------------------------+\n")
	default:
		panic("Bad Node Mode (must be FULL or SERVER)")
	}

	//Database
	switch cfg.App.DBType {
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

	if cfg.App.ExportData {
		s.DB.SetExportData(cfg.App.ExportDataSubpath)
	}

	//Network
	switch cfg.App.Network {
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

	s.AuditHeartBeats = make([]interfaces.IMsg, 0)
	s.FedServerFaults = make([][]interfaces.IMsg, 0)

	a, _ := anchor.InitAnchor(s)
	s.Anchor = a

	s.loadDatabase()

	s.initServerKeys()
}

func (s *State) __getDBState(height uint32) *DBState {
	index := height - s._DBHeightComplete
	if index > uint32(len(s._DBStates)) {
		return nil
	}
	return s._DBStates[index]
}
func (s *State) GetDBState(height uint32) *DBState {
	s._DBStatesMultex.Lock()
	defer s._DBStatesMultex.Unlock()
	return s.__getDBState(height)
}

func (s *State) UpdateState() {

	s._ProcessListsMultex.Lock()
	// Create DState blocks for all completed Process Lists
	for len(s._ProcessLists) > 0 && s._ProcessLists[0].Complete() {
		fmt.Println("Process List len",len(s._ProcessLists))
		pl := s._ProcessLists[0]
		pl.Process(s)
		s.AddDBState(true, pl.DirectoryBlock, pl.AdminBlock, pl.FactoidBlock, pl.EntryCreditBlock)
		s._ProcessLists = s._ProcessLists[1:]
		s._ProcessListBase = pl.DirectoryBlock.GetHeader().GetDBHeight() + 1
	}
	s._ProcessListsMultex.Unlock()

	fmt.Println("hhhhhhhh1hhhhhhhhhhhhhhhhhhhhhhhhhhhhang")
	
	s._DBStatesMultex.Lock()
	defer s._DBStatesMultex.Unlock()

	// Process all contiguous dbStates from the beginning.  Break once we hit a
	// nil or run out of states to process.
	update := false
	for len(s._DBStates) > 0 && s._DBStates[0] != nil {
		s._LastDBState = s._DBStates[0]
		fmt.Println("hhhhhhhhxxxxx1hhhhhhhhhhhhhhhhhhhhhhhhhhhhang")
		s.ProcessEndOfBlock(s._DBHeightComplete)
		//
		// Need to consider how to deal with the Factoid state
		//
		fmt.Println("Process DState", s._DBHeightComplete)
		fmt.Println("hhhhhhhhaaaaaaaaaaaa2222hhhhhhhhhhhhhhhhhhhhhhhhhhhhang")
		s.FactoidState.ProcessEndOfBlock(s)
		fmt.Println("hhhhhhhhbbbbbbbbbbbbbbbbb2222hhhhhhhhhhhhhhhhhhhhhhhhhhhhang")
		s._LastDBState.Process(s)
		fmt.Println("hhhhhhhhcccccccccccccccccc2222hhhhhhhhhhhhhhhhhhhhhhhhhhhhang")
		
		s._DBHeightComplete = s._LastDBState.DirectoryBlock.GetHeader().GetDBHeight()
		s._DBStates = s._DBStates[1:]
		update = true
	}
	fmt.Println("hhhhhhhh2222hhhhhhhhhhhhhhhhhhhhhhhhhhhhang")
	
	for s._ProcessListBase <= s._DBHeightComplete {
		if len(s._ProcessLists) > 0 {
			s._ProcessLists = s._ProcessLists[1:]
		}
	}

	fmt.Println("hhhhhhhh3333333333hhhhhhhhhhhhhhhhhhhhhhhhhhhhang")
	if update {
		s._DBHeightComplete = s._DBHeightComplete + 1
		s._DBHeight = s._DBHeightComplete + 1
	}

	
}

// Adds blocks that are either pulled locally from a database, or acquired from peers.
func (s *State) AddDBState(isNew bool,
	directoryBlock interfaces.IDirectoryBlock,
	adminBlock interfaces.IAdminBlock,
	factoidBlock interfaces.IFBlock,
	entryCreditBlock interfaces.IEntryCreditBlock) {

	s._DBStatesMultex.Lock()
	defer s._DBStatesMultex.Unlock()

	dbheight := directoryBlock.GetHeader().GetDBHeight()

	// Compute the index of this state in our DBStates array;  we don't keep block info
	// around that we have already processed.  But we do have to keep this info around
	// if we are missing some previous block info.
	index := int(dbheight) - int(s._DBHeightComplete)

	// Ignore repeats
	if index < 0 {
		return
	}

	// Grow the list to include this index
	for len(s._DBStates) <= index {
		s._DBStates = append(s._DBStates, nil)
	}

	// Create a new DBState
	dbState := new(DBState)
	dbState.DirectoryBlock = directoryBlock
	dbState.AdminBlock = adminBlock
	dbState.FactoidBlock = factoidBlock
	dbState.EntryCreditBlock = entryCreditBlock
	dbState.isNew = isNew

	s._DBStates[index] = dbState

}

func (s *State) loadDatabase() {

	dblk, err := s.DB.FetchDirectoryBlockHead()
	if err != nil {
		panic(err.Error())
	}

	if dblk == nil && s.NetworkNumber == constants.NETWORK_LOCAL {
		dblk, err = s.CreateDBlock(0)
		if err != nil {
			panic("Failed to initialize Factoids: " + err.Error())
		}
		ablk := s.NewAdminBlock(0)
		fblk := block.GetGenesisFBlock()
		ecb := entryCreditBlock.NewECBlock()

		s.AddDBState(true, dblk, ablk, fblk, ecb)
		s._DBHeight++

	} else {

		dBlocks, err := s.DB.FetchAllDBlocks()
		if err != nil {
			panic(err.Error())
		}

		aBlocks, err := s.DB.FetchAllABlocks()
		if err != nil {
			panic(err.Error())
		}

		fBlocks, err := s.DB.FetchAllFBlocks()
		if err != nil {
			panic(err.Error())
		}

		ecBlocks, err := s.DB.FetchAllECBlocks()
		if err != nil {
			panic(err.Error())
		}

		if len(dBlocks) != len(aBlocks) ||
			len(dBlocks) != len(fBlocks) ||
			len(dBlocks) != len(ecBlocks) {
			fmt.Printf("\nFailure to load all blocks on Init. D%d A%d F%d Ec%d\n", len(dBlocks), len(aBlocks), len(fBlocks), len(ecBlocks))
		}

		for i := 0; i < len(dBlocks); i++ {
			s.AddDBState(false, dBlocks[i], aBlocks[i], fBlocks[i], ecBlocks[i])
		}
	}
}

// This routine is called once we have everything to create a Directory Block.
// It is called by the follower code.  It is requried to build the Directory Block
// to validate the signatures we will get with the DirectoryBlockSignature messages.
func (s *State) ProcessEndOfBlock(dbheight uint32) {
	s.CreateDBlock(dbheight)
	s.LastAck = nil
}

func (s *State) SetListComplete() {
	s._DBStatesMultex.Lock()
	pl := s.pli(s._DBHeight)
	pl.SigComplete[pl.ServerIndex] = true
	s._DBStatesMultex.Unlock()
}

func (s *State) ListComplete() bool {
	s._ProcessListsMultex.Lock()
	defer s._ProcessListsMultex.Unlock()

	pl := s.pli(s._DBHeight)
	return pl.Complete()
}

// Here we need to validate the signatures of the previous block.  We also need to update
// stuff like a change to the exchange rate for Entry Credits, the number of Federated Servers (until
// such time we reach the 32 server target), the Federated Server profiles (Factom IDs, Bitcoin
// addresses, other chain addresses (like Ethereum, etc.) and more).
func (s *State) AddAdminBlock(interfaces.IAdminBlock) {

}

// Returns the Process List block for the given height.  Returns nil if the Process list
// block specified doesn't exist or is out of range.
func (s *State) pli(height uint32) *ProcessList {
	s._ProcessListsMultex.Lock()
	defer s._ProcessListsMultex.Unlock()
	fmt.Print("Enter Pli")
	defer fmt.Println("Leaving pli")

	i := height - s._ProcessListBase
	if i >= uint32(len(s._ProcessLists)) { // Can't be zero, unsigned. One test tests both
		return nil
	}

	r := s._ProcessLists[i]
	return r
}

func (s *State) NewPli(height uint32) *ProcessList {
	s._ProcessListsMultex.Lock()
	defer s._ProcessListsMultex.Unlock()
	fmt.Print("Enter NewPli")
	defer fmt.Println("Leaving NewPli")
	
	i := int(height) - int(s._ProcessListBase)
	if i < 0 {
		return nil // No blocks before the genesis block
	}

	for i >= len(s._ProcessLists) {
		s._ProcessLists = append(s._ProcessLists, nil)
	}

	if s._ProcessLists[i] != nil { // Do nothing if the process list already exists.
		return s._ProcessLists[i]
	}

	r := NewProcessList(s.totalServers, height)
	s._ProcessLists[i] = r

	return r
}

// Create a new Directory Block at the given height.
// Return the new Current Directory Block
func (s *State) CreateDBlock(height uint32) (interfaces.IDirectoryBlock, error) {

	currPL := s.NewPli(height)
	currPL.SetComplete(false)

	newdb := directoryBlock.NewDirectoryBlock(height)
	currPL.DirectoryBlock = newdb
	var peb interfaces.IEntryCreditBlock

	dstate := s.__getDBState(height - 1)
	if dstate != nil {
		prev := dstate.DirectoryBlock
		bodyMR, err := prev.BuildBodyMR()
		if err != nil {
			return nil, err
		}
		fmt.Println("ooooooooooooooooooooo55555555555oooooooooooo")
		newdb.GetHeader().SetBodyMR(bodyMR)

		prevLedgerKeyMR := prev.GetHash()
		if prevLedgerKeyMR == nil {
			return nil, errors.New("prevLedgerKeyMR is nil")
		}
		fmt.Println("ooooooooooooooooo6666666666oooooooooooooooo")
		newdb.GetHeader().SetPrevLedgerKeyMR(prevLedgerKeyMR)
		newdb.GetHeader().SetPrevKeyMR(prev.GetKeyMR())
		peb = dstate.EntryCreditBlock
		fmt.Println("oooooooooooooo7777ooooooooooooooooooo")
	}
	fmt.Println("oooooooooooooooooooooo8888ooooooooooo")
	
	eb, _ := entryCreditBlock.NextECBlock(peb)
	currPL.EntryCreditBlock = eb
	currPL.AdminBlock = s.NewAdminBlock(height)
	fmt.Println("oooooooooooooooooooooooooo9999ooooooo")
	
	return newdb, nil
}

func (s *State) GetServerIndex(dbheight uint32) int {
	pl := s.pli(dbheight)
	if pl == nil {
		return -1
	}
	return pl.ServerIndex
}

func (s *State) GetNewEBlks(dbheight uint32, key [32]byte) interfaces.IEntryBlock {
	pl := s.pli(dbheight)
	var value interfaces.IEntryBlock
	if pl != nil {
		value = pl.GetNewEBlks(key)
	}
	return value
}

func (s *State) PutNewEBlks(dbheight uint32, key [32]byte, value interfaces.IEntryBlock) {
	pl := s.pli(dbheight)
	if pl != nil {
		pl.PutNewEBlks(key, value)
	}
}

func (s *State) GetCommits(dbheight uint32, key interfaces.IHash) interfaces.IMsg {
	pl := s.pli(dbheight)
	var value interfaces.IMsg
	if pl != nil {
		value = pl.GetCommits(key)
	}
	return value
}

func (s *State) PutCommits(dbheight uint32, key interfaces.IHash, value interfaces.IMsg) {
	pl := s.pli(dbheight)
	if pl != nil {
		pl.PutCommits(key, value)
	}
}

// Messages that will go into the Process List must match an Acknowledgement.
// The code for this is the same for all such messages, so we put it here.
//
// Returns true if it finds a match
func (s *State) FollowerExecuteMsg(m interfaces.IMsg) (bool, error) {
	acks := s.Acks
	ack, ok := acks[m.GetHash().Fixed()].(*messages.Ack)
	if !ok || ack == nil {
		fmt.Println("Msg No Match!")
		s.Holding[m.GetHash().Fixed()] = m
		return false, nil
	} else {
		fmt.Println("Msg Match!")
		pl := s.NewPli(ack.DBHeight)
		pl.AddToProcessList(ack, m)
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
		fmt.Println("Ack Match!")
		match.FollowerExecute(s)
		return true, nil
	}
	fmt.Println("Ack No Match!")
	
	return false, nil
}

func (s *State) GetEntryCreditBlock(dbheight uint32) interfaces.IEntryCreditBlock {
	pl := s.pli(dbheight)
	if pl == nil {
		return nil
	}
	return pl.EntryCreditBlock
}

func (s *State) SetEntryCreditBlock(dbheight uint32, ecblk interfaces.IEntryCreditBlock) {
	pl := s.pli(dbheight)
	pl.EntryCreditBlock = ecblk
}

func (s *State) GetServer() interfaces.IServer {
	return s.myServer
}

func (s *State) SetServer(server interfaces.IServer) {
	s.myServer = server
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

func (s *State) initServerKeys() {
	var err error
	s.serverPrivKey, err = primitives.NewPrivateKeyFromHex(s.GetCfg().(*util.FactomdConfig).App.LocalServerPrivKey)
	if err != nil {
		//panic("Cannot parse Server Private Key from configuration file: " + err.Error())
	}
	s.serverPubKey = primitives.PubKeyFromString(constants.SERVER_PUB_KEY)
}

func (s *State) LogInfo(args ...interface{}) {
	s.Logger.Info(args...)
}

// Lists
// =====
func (s *State) GetAuditServers(dbheight uint32) []interfaces.IServer {
	pl := s.pli(dbheight)
	if pl == nil {
		return nil
	}
	return pl.AuditServers
}
func (s *State) GetFedServers(dbheight uint32) []interfaces.IServer {
	pl := s.pli(dbheight)
	if pl == nil {
		return nil
	}
	return pl.FedServers
}
func (s *State) GetServerOrder(dbheight uint32) [][]interfaces.IServer {
	pl := s.pli(dbheight)
	if pl == nil {
		return nil
	}
	return pl.ServerOrder
}
func (s *State) GetAuditHeartBeats() []interfaces.IMsg {
	return s.AuditHeartBeats
}
func (s *State) GetFedServerFaults() [][]interfaces.IMsg {
	return s.FedServerFaults
}

func (s *State) GetTimestamp() interfaces.Timestamp {
	return interfaces.Timestamp(primitives.GetTimeMilli())
}

func (s *State) Sign([]byte) interfaces.IFullSignature {
	return new(primitives.Signature)
}

func (s *State) GetFactoidKeyMR(dbheight uint32) interfaces.IHash {
	pl := s.pli(dbheight)
	if pl == nil {
		return nil
	}
	return pl.FactoidBlock.GetKeyMR()
}

func (s *State) GetAdminBlock(dbheight uint32) interfaces.IAdminBlock {
	pl := s.pli(dbheight)
	if pl == nil {
		return nil
	}
	return pl.AdminBlock
}

func (s *State) SetAdminBlock(dbheight uint32, adblock interfaces.IAdminBlock) {
	pl := s.pli(dbheight)
	if pl == nil {
		return
	}
	pl.AdminBlock = adblock
}

func (s *State) GetFactoidState(dbheight uint32) interfaces.IFactoidState {
	return s.FactoidState
}

func (s *State) SetFactoidState(dbheight uint32, fs interfaces.IFactoidState) {
	s.FactoidState = fs
}

// Allow us the ability to update the port number at run time....
func (s *State) SetPort(port int) {
	// Get our factomd configuration information.
	cfg := s.GetCfg().(*util.FactomdConfig)
	cfg.Wsapi.PortNumber = port
}

func (s *State) GetPort() int {
	cfg := s.GetCfg().(*util.FactomdConfig)
	return cfg.Wsapi.PortNumber
}

// Tests the given hash, and returns true if this server is the leader for this key.
// For example, keys we test include:
//
// The Hash of the Factoid Hash
// Entry Credit Addresses
// ChainIDs
// ...
func (s *State) LeaderFor([]byte) bool {
	if s.totalServers == 1 && s.serverState == 1 &&
		s.NetworkNumber == constants.NETWORK_LOCAL {
		return true
	}
	return false
}

func (s *State) NetworkInMsgQueue() chan interfaces.IMsg {
	return s.networkInMsgQueue
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

func (s *State) LeaderInMsgQueue() chan interfaces.IMsg {
	return s.leaderInMsgQueue
}

func (s *State) FollowerInMsgQueue() chan interfaces.IMsg {
	return s.followerInMsgQueue
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

func (s *State) GetTotalServers() int {
	return s.totalServers
}

func (s *State) GetProcessListLen(dbheight uint32, list int) int {
	pl := s.pli(dbheight)
	if pl == nil {
		return 0
	}
	if list >= s.totalServers {
		return -1
	}
	return pl.GetLen(list)
}

func (s *State) GetServerState() int {
	return s.serverState
}

func (s *State) GetNetworkNumber() int {
	return s.NetworkNumber
}

func (s *State) GetMatryoshka(dbheight uint32) interfaces.IHash {
	return nil
}

func (s *State) GetLastAck() interfaces.IMsg {
	return s.LastAck
}

func (s *State) SetLastAck(ack interfaces.IMsg) {
	s.LastAck = ack
}

func (s *State) InitLevelDB() error {
	if s.DB != nil {
		return nil
	}

	cfg := s.Cfg.(*util.FactomdConfig)
	path := cfg.App.LdbPath + "/" + cfg.App.Network + "/" + "factoid_level.db"

	log.Printfln("Creating Database at %v", path)

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

	cfg := s.Cfg.(*util.FactomdConfig)
	path := cfg.App.BoltDBPath + "/" + cfg.App.Network + "/"
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
	var out bytes.Buffer

	out.WriteString(fmt.Sprintf("Queues: NetIn %d NetOut %d NetInvalid %d InMsg %d Leader %d Follower %d",
		len(s.NetworkInMsgQueue()),
		len(s.NetworkOutMsgQueue()),
		len(s.NetworkInvalidMsgQueue()),
		len(s.InMsgQueue()),
		len(s.LeaderInMsgQueue()),
		len(s.FollowerInMsgQueue())))

	return out.String()

}

func (s *State) NewAdminBlock(dbheight uint32) interfaces.IAdminBlock {
	ab := new(adminBlock.AdminBlock)
	ab.Header = s.NewAdminBlockHeader(dbheight)

	s.DB.SaveABlockHead(ab)

	return ab
}

func (s *State) NewAdminBlockHeader(dbheight uint32) interfaces.IABlockHeader {
	header := new(adminBlock.ABlockHeader)
	header.DBHeight = dbheight
	pl := s.pli(dbheight)
	if pl == nil || pl.AdminBlock == nil {
		header.PrevLedgerKeyMR = primitives.NewHash(constants.ZERO_HASH)
	} else {
		keymr, err := s.pli(dbheight).AdminBlock.LedgerKeyMR()
		if err != nil {
			panic(err.Error())
		}
		header.PrevLedgerKeyMR = keymr
	}
	header.HeaderExpansionSize = 0
	header.HeaderExpansionArea = make([]byte, 0)
	header.MessageCount = 0
	header.BodySize = 0
	return header
}

func (s *State) PrintType(msgType int) bool {
	r := true
	return r
	r = r && msgType != constants.ACK_MSG
	r = r && msgType != constants.EOM_MSG
	r = r && msgType != constants.DIRECTORY_BLOCK_SIGNATURE_MSG
	return r
}

func (s *State) GetNetworkName() string {
	return (s.Cfg.(util.FactomdConfig)).App.Network

}

func (s *State) GetDirectoryBlock(dbheight uint32) interfaces.IDirectoryBlock {
	pl := s.pli(dbheight)
	if pl != nil {
		return pl.DirectoryBlock
	}
	return nil
}

func (s *State) SetDirectoryBlock(dbheight uint32, dirblk interfaces.IDirectoryBlock) {
	pl := s.pli(dbheight)
	pl.DirectoryBlock = dirblk
}

func (s *State) GetDB() interfaces.DBOverlay {
	return s.DB
}

func (s *State) SetDB(dbase interfaces.DBOverlay) {
	s.DB = databaseOverlay.NewOverlay(dbase)
}

func (s *State) GetDBHeight() uint32 {
	return s._DBHeight
}

func (s *State) GetDBHeightComplete() uint32 {
	return s._DBHeightComplete
}

func (s *State) GetProcessListBase() uint32 {
	return s._ProcessListBase
}

func (s *State) GetNewHash() interfaces.IHash {
	return new(primitives.Hash)
}

func (s *State) RecalculateBalances() error {
	fs := s.FactoidState
	fs.ResetBalances()

	blocks, err := s.DB.FetchAllFBlocks()
	if err != nil {
		return err
	}
	for _, block := range blocks {
		txs := block.GetTransactions()
		for _, tx := range txs {
			err = fs.UpdateTransaction(tx)
			if err != nil {
				fs.ResetBalances()
				return err
			}
		}
	}

	ecBlocks, err := s.DB.FetchAllECBlocks()
	if err != nil {
		return err
	}
	for _, block := range ecBlocks {
		txs := block.GetBody().GetEntries()
		for _, tx := range txs {
			err = fs.UpdateECTransaction(tx)
			if err != nil {
				fs.ResetBalances()
				return err
			}
		}
	}
	return nil
}
