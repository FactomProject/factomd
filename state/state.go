package state

import (
	"bytes"
	"errors"
	"fmt"
	"sync"

	"os"

	"github.com/FactomProject/factomd/anchor"
	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/directoryBlock/dbInfo"
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

	networkInMsgQueue      chan interfaces.IMsg
	networkOutMsgQueue     chan interfaces.IMsg
	networkInvalidMsgQueue chan interfaces.IMsg
	inMsgQueue             chan interfaces.IMsg
	leaderInMsgQueue       chan interfaces.IMsg
	followerInMsgQueue     chan interfaces.IMsg

	myServer      interfaces.IServer //the server running on this Federated Server
	serverPrivKey primitives.PrivateKey
	serverPubKey  primitives.PublicKey

	// Maps
	// ====
	// For Follower
	Holding map[[32]byte]interfaces.IMsg // Hold Messages
	Acks    map[[32]byte]interfaces.IMsg // Hold Acknowledgemets

	NewEBlksSem *sync.Mutex
	NewEBlks    map[[32]byte]interfaces.IEntryBlock // Entry Blocks added within 10 minutes (follower and leader)

	CommitsSem *sync.Mutex
	Commits    map[[32]byte]interfaces.IMsg // Used by the leader, validate

	// Lists
	// =====
	AuditServers    []interfaces.IServer   // List of Audit Servers
	FedServers      []interfaces.IServer   // List of Federated Servers
	ServerOrder     [][]interfaces.IServer // 10 lists for Server Order for each minute
	ProcessList     [][]interfaces.IMsg    // List of Processed Messages Per server
	PLHeight        []int                  // Process List Height (for processing lists)
	AuditHeartBeats []interfaces.IMsg      // The checklist of HeartBeats for this period
	FedServerFaults [][]interfaces.IMsg    // Keep a fault list for every server

	//Network MAIN = 0, TEST = 1, LOCAL = 2, CUSTOM = 3
	NetworkNumber int // Encoded into Directory Blocks(s.Cfg.(*util.FactomdConfig)).String()

	// Number of Servers acknowledged by Factom
	TotalServers int
	ServerState  int                // (0 if client, 1 if server, 2 if audit server
	Matryoshka   []interfaces.IHash // Reverse Hash

	// Database
	DB *databaseOverlay.Overlay

	// Directory Block State
	PreviousDirectoryBlock interfaces.IDirectoryBlock
	CurrentDirectoryBlock  interfaces.IDirectoryBlock
	DBHeight               uint32

	// Web Services
	Port int

	// Message State
	LastAck interfaces.IMsg // Return the last Acknowledgement set by this server

	FactoidState      interfaces.IFactoidState
	PrevFactoidKeyMR  interfaces.IHash
	CurrentAdminBlock interfaces.IAdminBlock
	EntryCreditBlock  interfaces.IEntryCreditBlock

	Logger *logger.FLogger

	Anchor interfaces.IAnchor
}

var _ interfaces.IState = (*State)(nil)

func (s *State) GetNewEBlks(key [32]byte) interfaces.IEntryBlock {
	s.NewEBlksSem.Lock()
	value := s.NewEBlks[key]
	s.NewEBlksSem.Unlock()
	return value
}

func (s *State) PutNewEBlks(key [32]byte, value interfaces.IEntryBlock) {
	s.NewEBlksSem.Lock()
	s.NewEBlks[key] = value
	s.NewEBlksSem.Unlock()
}

func (s *State) GetCommits(key interfaces.IHash) interfaces.IMsg {
	s.CommitsSem.Lock()
	value := s.Commits[key.Fixed()]
	s.CommitsSem.Unlock()
	return value
}

func (s *State) PutCommits(key interfaces.IHash, value interfaces.IMsg) {
	s.CommitsSem.Lock()
	s.Commits[key.Fixed()] = value
	s.CommitsSem.Unlock()
}

// Messages that match an acknowledgement, and are added to the process list
// all do the same thing.  So that logic is here.
//
// Returns true if it finds a match
func (s *State) MatchAckFollowerExecute(m interfaces.IMsg) (bool, error) {
	acks := s.Acks
	ack, ok := acks[m.GetHash().Fixed()].(*messages.Ack)
	if !ok || ack == nil {
		s.Holding[m.GetHash().Fixed()] = m
		return false, nil
	} else {
		processlist := s.GetProcessList()[ack.ServerIndex]
		for len(processlist) <= ack.Height {
			processlist = append(processlist, nil)
		}
		fmt.Println("Add message at height", ack.Height)
		processlist[ack.Height] = m
		s.GetProcessList()[ack.ServerIndex] = processlist
		// remove the message from the holding/ack maps.
		delete(acks, m.GetHash().Fixed())
		delete(s.Holding, m.GetHash().Fixed())
		s.UpdateProcessLists()
		return true, nil
	}
}

// Match an acknowledgement to a message
func (s *State) FollowerExecuteAck(msg interfaces.IMsg) error {
	ack := msg.(*messages.Ack)
	s.Acks[ack.GetHash().Fixed()] = ack
	match := s.Holding[ack.GetHash().Fixed()]
	if match != nil {
		// If we have a match, the ack is in the Acks, so we
		// can match the message to the ack.  One set of code.
		match.FollowerExecute(s)
	}

	return nil
}

// Run through the process lists, and update the state as required by
// any new entries.  In the near future, we will want to have this in a
// "temp" state, that we push to the regular state at the end of 10 minutes.
// But for now, we will just update.
//
// This routine can only be called by the Follower goroutine.
func (s *State) UpdateProcessLists() {
	for i := 0; i < len(s.ProcessList); i++ {
		plist := s.GetProcessList()[i]
		for j := s.PLHeight[i]; j < len(plist); j++ {
			fmt.Println("UpdatePL: ", j)
			if plist[j] == nil {
				fmt.Println("Nil at", j)
				break
			}
			plist[j].Process(s)   // Process this entry
			s.PLHeight[i] = j + 1 //   and don't process it again.
		}
	}
}

func (s *State) GetCurrentEntryCreditBlock() interfaces.IEntryCreditBlock {
	return s.EntryCreditBlock
}

func (s *State) SetCurrentEntryCreditBlock(ecblk interfaces.IEntryCreditBlock) {
	s.EntryCreditBlock = ecblk
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
func (s *State) GetAuditServers() []interfaces.IServer {
	return s.AuditServers
}
func (s *State) GetFedServers() []interfaces.IServer {
	return s.FedServers
}
func (s *State) GetServerOrder() [][]interfaces.IServer {
	return s.ServerOrder
}
func (s *State) GetProcessList() [][]interfaces.IMsg {
	return s.ProcessList
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

// This routine is called once we have everything to create a Directory Block.
// It is called by the follower code.  It is requried to build the Directory Block
// to validate the signatures we will get with the DirectoryBlockSignature messages.
func (s *State) ProcessEndOfBlock() {

	//Must have all the complete process lists at this point!

	s.UpdateProcessLists() // Do any remaining processing

	for i := 0; i < len(s.ProcessList); i++ { // Reset heights to zero for all lists
		s.PLHeight[i] = 0
	}

	s.PreviousDirectoryBlock = s.CurrentDirectoryBlock
	previousECBlock := s.GetCurrentEntryCreditBlock()

	s.FactoidState.ProcessEndOfBlock(s) // Clean up Factoids

	db, err := s.CreateDBlock()
	if err != nil {
		panic("Failed to create a Directory Block")
	}

	if previousECBlock != nil {
		s.DB.ProcessECBlockBatch(previousECBlock)
	}

	s.SetCurrentDirectoryBlock(db)

	if s.PreviousDirectoryBlock != nil {
		if err = s.DB.SaveDirectoryBlockHead(s.PreviousDirectoryBlock); err != nil {
			panic(err.Error())
		}
		s.Anchor.UpdateDirBlockInfoMap(dbInfo.NewDirBlockInfoFromDirBlock(s.PreviousDirectoryBlock))
	} else {
		log.Println("No old db")
	}

	s.ProcessList = make([][]interfaces.IMsg, 1)
	s.LastAck = nil
	s.NewEBlks = make(map[[32]byte]interfaces.IEntryBlock)
}

func (s *State) GetPrevFactoidKeyMR() interfaces.IHash {
	return s.PrevFactoidKeyMR
}

func (s *State) SetPrevFactoidKeyMR(hash interfaces.IHash) {
	s.PrevFactoidKeyMR = hash
}

func (s *State) GetCurrentAdminBlock() interfaces.IAdminBlock {
	return s.CurrentAdminBlock
}

func (s *State) SetCurrentAdminBlock(adblock interfaces.IAdminBlock) {
	s.CurrentAdminBlock = adblock
}

func (s *State) GetFactoidState() interfaces.IFactoidState {
	return s.FactoidState
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
	if s.TotalServers == 1 && s.ServerState == 1 &&
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
	return s.TotalServers
}

func (s *State) GetServerState() int {
	return s.ServerState
}

func (s *State) GetNetworkNumber() int {
	return s.NetworkNumber
}

func (s *State) GetMatryoshka() []interfaces.IHash {
	return s.Matryoshka
}

func (s *State) GetLastAck() interfaces.IMsg {
	return s.LastAck
}

func (s *State) SetLastAck(ack interfaces.IMsg) {
	s.LastAck = ack
}

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

	s.TotalServers = 1

	// Setup the FactoidState and Validation Service that holds factoid and entry credit balances
	fs := new(FactoidState)
	fs.ValidationService = NewValidationService()
	s.FactoidState = fs

	switch cfg.App.NodeMode {
	case "FULL":
		s.ServerState = 0
	case "SERVER":
		s.ServerState = 1
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
	s.Holding = make(map[[32]byte]interfaces.IMsg)
	s.Acks = make(map[[32]byte]interfaces.IMsg)

	s.NewEBlksSem = new(sync.Mutex)
	s.NewEBlks = make(map[[32]byte]interfaces.IEntryBlock)

	s.CommitsSem = new(sync.Mutex)
	s.Commits = make(map[[32]byte]interfaces.IMsg)

	s.AuditServers = make([]interfaces.IServer, 0)
	s.FedServers = make([]interfaces.IServer, 0)
	s.ServerOrder = make([][]interfaces.IServer, 0)

	s.ProcessList = make([][]interfaces.IMsg, 1)
	s.PLHeight = make([]int, 1)

	s.AuditHeartBeats = make([]interfaces.IMsg, 0)
	s.FedServerFaults = make([][]interfaces.IMsg, 0)

	a, _ := anchor.InitAnchor(s)
	s.Anchor = a

	s.loadDatabase()
	s.initServerKeys()
}

func (s *State) loadDatabase() {
	dblk, err := s.DB.FetchDirectoryBlockHead()
	if err != nil {
		panic(err.Error())
	}

	if dblk == nil && s.NetworkNumber == constants.NETWORK_LOCAL {
		dblk, err = s.CreateDBlock()
		if err != nil {
			panic("Failed to initialize Factoids: " + err.Error())
		}

		//TODO Also need to set Admin block and EC Credit block

		fblk := block.GetGenesisFBlock()

		dblk.GetDBEntries()[2].SetKeyMR(fblk.GetKeyMR())

		s.SetCurrentDirectoryBlock(dblk)

		if s.FactoidState == nil {
			fs := new(FactoidState)
			fs.ValidationService = NewValidationService()
			s.FactoidState = fs
		}
		if err := s.FactoidState.AddTransactionBlock(fblk); err != nil {
			panic("Failed to initialize Factoids: " + err.Error())
		}

		s.EntryCreditBlock = entryCreditBlock.NewECBlock()

		dblk, err = s.CreateDBlock()
		if dblk == nil {
			panic("dblk should never be nil")
		}
		s.ProcessEndOfBlock()
		dblk = s.GetCurrentDirectoryBlock()

	} else {
		if dblk.GetHeader().GetDBHeight() > 0 {
			dbPrev, err := s.DB.FetchDBlockByKeyMR(dblk.GetHeader().GetPrevKeyMR())
			if err != nil {
				panic("Failed to load the Previous Directory Block: " + err.Error())
			}
			if dbPrev == nil {
				panic("Did not find the Previous Directory Block in the database")
			}
			s.PreviousDirectoryBlock = dbPrev.(interfaces.IDirectoryBlock)
		}

		fBlocks, err := s.DB.FetchAllFBlocks()

		fmt.Printf("Processing %d FBlocks\n", len(fBlocks))

		if err != nil {
			panic(err.Error())
		}
		for _, block := range fBlocks {
			s.GetFactoidState().AddTransactionBlock(block)
		}

		ecBlocks, err := s.DB.FetchAllECBlocks()
		if err != nil {
			panic(err.Error())
		}

		fmt.Printf("Processing %d ECBlocks\n", len(ecBlocks))

		for _, block := range ecBlocks {
			s.EntryCreditBlock = block
			s.GetFactoidState().AddECBlock(block)
		}
	}
	s.SetCurrentDirectoryBlock(dblk)
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

func (s *State) NewAdminBlock() interfaces.IAdminBlock {
	ab := new(adminBlock.AdminBlock)
	ab.Header = s.NewAdminBlockHeader()

	s.DB.SaveABlockHead(ab)

	return ab
}

func (s *State) NewAdminBlockHeader() interfaces.IABlockHeader {
	header := new(adminBlock.ABlockHeader)
	header.DBHeight = s.GetDBHeight()
	if s.GetCurrentAdminBlock() == nil {
		header.PrevLedgerKeyMR = primitives.NewHash(constants.ZERO_HASH)
	} else {
		keymr, err := s.GetCurrentAdminBlock().LedgerKeyMR()
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

func (s *State) CreateDBlock() (b interfaces.IDirectoryBlock, err error) {
	prev := s.GetCurrentDirectoryBlock()

	b = new(directoryBlock.DirectoryBlock)

	b.SetHeader(new(directoryBlock.DBlockHeader))
	b.GetHeader().SetVersion(constants.VERSION_0)

	if prev == nil {
		b.GetHeader().SetPrevLedgerKeyMR(primitives.NewZeroHash())
		b.GetHeader().SetPrevKeyMR(primitives.NewZeroHash())
		b.GetHeader().SetDBHeight(0)
		eb, _ := entryCreditBlock.NextECBlock(nil)
		s.EntryCreditBlock = eb
	} else {
		bodyMR, err := prev.BuildBodyMR()
		if err != nil {
			return nil, err
		}
		prev.GetHeader().SetBodyMR(bodyMR)

		prevLedgerKeyMR := prev.GetHash()
		if prevLedgerKeyMR == nil {
			return nil, errors.New("prevLedgerKeyMR is nil")
		}
		b.GetHeader().SetPrevLedgerKeyMR(prevLedgerKeyMR)
		b.GetHeader().SetPrevKeyMR(prev.GetKeyMR())
		b.GetHeader().SetDBHeight(prev.GetHeader().GetDBHeight() + 1)
		eb, _ := entryCreditBlock.NextECBlock(s.EntryCreditBlock)
		s.EntryCreditBlock = eb
	}

	b.SetDBEntries(make([]interfaces.IDBEntry, 0))

	s.CurrentAdminBlock = s.NewAdminBlock()

	b.AddEntry(primitives.NewHash(constants.ADMIN_CHAINID), primitives.NewZeroHash())
	b.AddEntry(primitives.NewHash(constants.EC_CHAINID), primitives.NewZeroHash())
	b.AddEntry(primitives.NewHash(constants.FACTOID_CHAINID), primitives.NewZeroHash())

	return b, err
}

func (s *State) PrintType(msgType int) bool {
	return msgType != constants.EOM_MSG &&
		msgType != constants.DIRECTORY_BLOCK_SIGNATURE_MSG
}

func (s *State) GetNetworkName() string {
	return (s.Cfg.(util.FactomdConfig)).App.Network

}

func (s *State) GetPreviousDirectoryBlock() interfaces.IDirectoryBlock {
	return s.PreviousDirectoryBlock
}

func (s *State) GetCurrentDirectoryBlock() interfaces.IDirectoryBlock {
	return s.CurrentDirectoryBlock
}

func (s *State) SetCurrentDirectoryBlock(dirblk interfaces.IDirectoryBlock) {
	s.PreviousDirectoryBlock = s.CurrentDirectoryBlock
	s.CurrentDirectoryBlock = dirblk
}

func (s *State) GetDB() interfaces.DBOverlay {
	return s.DB
}

func (s *State) SetDB(dbase interfaces.DBOverlay) {
	s.DB = databaseOverlay.NewOverlay(dbase)
}

func (s *State) GetDBHeight() uint32 {
	if s.CurrentDirectoryBlock == nil {
		return 0
	}
	return s.CurrentDirectoryBlock.GetHeader().GetDBHeight()
}

func (s *State) GetNewHash() interfaces.IHash {
	return new(primitives.Hash)
}

func (s *State) RecalculateBalances() error {
	s.FactoidState.ResetBalances()

	blocks, err := s.DB.FetchAllFBlocks()
	if err != nil {
		return err
	}
	for _, block := range blocks {
		txs := block.GetTransactions()
		for _, tx := range txs {
			err = s.FactoidState.UpdateTransaction(tx)
			if err != nil {
				s.FactoidState.ResetBalances()
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
			err = s.FactoidState.UpdateECTransaction(tx)
			if err != nil {
				s.FactoidState.ResetBalances()
				return err
			}
		}
	}
	return nil
}
