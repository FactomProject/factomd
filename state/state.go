package state

import (
	"bytes"
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
	DBHeight    uint32       // Height as understood by the leader
	ServerIndex int          // Index of the server, as understood by the leader
	DBStates    *DBStateList // Holds all DBStates not yet processed.

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
	s.ProcessLists = NewProcessLists(s)

	s.DBStates = new(DBStateList)
	s.DBStates.state = s
	s.DBStates.multex = new(sync.Mutex)
	s.DBStates.DBStates = make([]*DBState, 0)

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

	fmt.Println("\nRunning on the ", cfg.App.Network, "Network")

	s.AuditHeartBeats = make([]interfaces.IMsg, 0)
	s.FedServerFaults = make([][]interfaces.IMsg, 0)

	a, _ := anchor.InitAnchor(s)
	s.Anchor = a
	log.Println("Loading Database")
	s.loadDatabase()

	s.initServerKeys()
}

func (s *State) GetDBState(height uint32) *DBState {
	return s.DBStates.Get(height)
}

func (s *State) UpdateState() {
	s.ProcessLists.UpdateState()
	s.DBStates.Process()
}

// Adds blocks that are either pulled locally from a database, or acquired from peers.
func (s *State) AddDBState(isNew bool,
	directoryBlock interfaces.IDirectoryBlock,
	adminBlock interfaces.IAdminBlock,
	factoidBlock interfaces.IFBlock,
	entryCreditBlock interfaces.IEntryCreditBlock) {

	dbState := s.DBStates.NewDBState(isNew, directoryBlock, adminBlock, factoidBlock, entryCreditBlock)
	s.DBStates.Put(dbState)
}

func (s *State) loadDatabase() {

	var i uint32
	for i = 0; true; i++ {
		var dhash interfaces.IHash
		var err error
		if dhash, err = s.DB.FetchDBKeyMRByHeight(i); err != nil {
			panic(err)
		}
		if dhash == nil {
			break
		}
		d, err := s.DB.FetchDBlockByKeyMR(dhash)
		if err != nil {
			panic(err)
		}
		if d == nil {
			panic("No DirectoryBlock for " + dhash.String())
		}
		dblk := d
		ablk, err := s.DB.FetchABlockByKeyMR(dblk.GetDBEntries()[0].GetKeyMR())
		if err != nil {
			panic(err)
		}
		if ablk == nil {
			panic("ablk is nil")
		}
		eblk, err := s.DB.FetchECBlockByHash(dblk.GetDBEntries()[1].GetKeyMR())
		if err != nil {
			panic(err)
		}
		if eblk == nil {
			panic("eblk is nil")
		}
		fblk, err := s.DB.FetchFBlockByKeyMR(dblk.GetDBEntries()[2].GetKeyMR())
		if err != nil {
			panic(err)
		}
		if fblk == nil {
			panic("fblk is nil")
		}

		s.DBStates.NewDBState(false, dblk, ablk, fblk, eblk)
		s.DBStates.Process()
	}

	if i == 0 && s.NetworkNumber == constants.NETWORK_LOCAL {
		fmt.Println("\n***********************************")
		fmt.Println("******* New Database **************")
		fmt.Println("***********************************\n")

		dblk := directoryBlock.NewDirectoryBlock(0, nil)
		ablk := s.NewAdminBlock(0)
		fblk := block.GetGenesisFBlock()
		eblk := entryCreditBlock.NewECBlock()

		s.DBStates.NewDBState(true, dblk, ablk, fblk, eblk)
		s.DBStates.Process()
	}
	log.Println(fmt.Sprintf("Loaded %d directory blocks",i))
	s.DBStates.Process()
}

// This routine is called once we have everything to create a Directory Block.
// It is called by the follower code.  It is requried to build the Directory Block
// to validate the signatures we will get with the DirectoryBlockSignature messages.
func (s *State) ProcessEndOfBlock(dbheight uint32) {
	s.LastAck = nil
}

// This returns the DBHeight as defined by the leader, not the follower.
// This value shouldn't be used by follower code.
func (s *State) GetDBHeight() uint32 {
	return s.DBHeight
}

// Messages that will go into the Process List must match an Acknowledgement.
// The code for this is the same for all such messages, so we put it here.
//
// Returns true if it finds a match
func (s *State) FollowerExecuteMsg(m interfaces.IMsg) (bool, error) {
	acks := s.Acks
	ack, ok := acks[m.GetHash().Fixed()].(*messages.Ack)
	if !ok || ack == nil {
		s.Holding[m.GetHash().Fixed()] = m
		return false, nil
	} else {
		pl := s.ProcessLists.Get(ack.DBHeight)
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
		match.FollowerExecute(s)
		return true, nil
	}

	return false, nil
}

func (s *State) LeaderExecute(m interfaces.IMsg) error {
	v := m.Validate(s.GetDBHeight(), s)
	if v <= 0 {
		return fmt.Errorf("Msg not valid: " + m.String())
	}

	hash := m.GetHash()

	ack, err := s.NewAck(hash)
	if err != nil {
		fmt.Println("Ack Error")
		return err
	}

	// Leader Execute creates an acknowledgement and the EOM
	s.NetworkOutMsgQueue() <- ack
	s.FollowerInMsgQueue() <- ack // Send the Ack to follower

	s.NetworkOutMsgQueue() <- m // Send the Message;  It works better if
	s.FollowerInMsgQueue() <- m // the msg follows the Ack, but it doesn't matter much.
	return nil
}

func (s *State) LeaderExecuteEOM(m interfaces.IMsg) error {
	eom, _ := m.(*messages.EOM)
	eom.DirectoryBlockHeight = s.DBHeight
	fmt.Println("EOM", s.DBHeight)
	return s.LeaderExecute(eom)
}

func (s *State) GetNewEBlocks(dbheight uint32, hash interfaces.IHash) interfaces.IEntryBlock {
	return nil
}
func (s *State) PutNewEBlocks(dbheight uint32, hash interfaces.IHash, eb interfaces.IEntryBlock) {
}

func (s *State) GetCommits(dbheight uint32, hash interfaces.IHash) interfaces.IMsg {
	return nil
}
func (s *State) PutCommits(dbheight uint32, hash interfaces.IHash, msg interfaces.IMsg) {
	s.ProcessLists.Get(dbheight).PutCommits(hash, msg)
}

func (s *State) ProcessCommitChain(dbheight uint32, commitChain interfaces.IMsg) {
	c, ok := commitChain.(*messages.CommitChainMsg)
	if ok {
		pl := s.ProcessLists.Get(dbheight)
		ecblk := pl.EntryCreditBlock
		ecbody := ecblk.GetBody()
		ecbody.AddEntry(c.CommitChain)
		s.GetFactoidState().UpdateECTransaction(c.CommitChain)
		s.PutCommits(dbheight, c.GetHash(), c)
	}
}

func (s *State) ProcessEOM(dbheight uint32, msg interfaces.IMsg) {
	e, ok := msg.(*messages.EOM)
	if !ok {
		panic("Must pass an EOM message to ProcessEOM)")
	}

	pl := s.ProcessLists.Get(dbheight)

	fs := s.FactoidState
	pl.FactoidBlock = fs.GetCurrentBlock()
	fs.EndOfPeriod(int(e.Minute))

	pl.AdminBlock = s.NewAdminBlock(dbheight)

	ecblk := pl.EntryCreditBlock
	if ecblk == nil {
		var err error
		dbstate := s.DBStates.Last()
		ecblk, err = entryCreditBlock.NextECBlock(dbstate.EntryCreditBlock)
		if err != nil {
			panic(err.Error())
		}
		pl.EntryCreditBlock = ecblk
	}
	ecbody := ecblk.GetBody()
	mn := entryCreditBlock.NewMinuteNumber2(e.Minute)

	ecbody.AddEntry(mn)

	if e.Minute == 9 {

		// TODO: This code needs to be reviewed... It works here, but we are effectively
		// executing "leader" code in the compainion "follower" goroutine...
		// Maybe that's okay?
		if s.LeaderFor(e.Bytes()) {
			// What really needs to happen is that we look to make sure all
			// EOM messages have been recieved.  If this is the LAST message,
			// and we have ALL EOM messages from all servers, then we
			// create a DirectoryBlockSignature (if we are the leader) and
			// send it out to the network.
			DBM := messages.NewDirectoryBlockSignature()
			prevDB := s.GetDirectoryBlock()
			if prevDB == nil {
				DBM.DirectoryBlockKeyMR = primitives.NewHash(constants.ZERO_HASH)
			} else {
				DBM.DirectoryBlockKeyMR = prevDB.GetKeyMR()
			}
			DBM.Sign(s)

			ack, err := s.NewAck(DBM.GetHash())
			if err != nil {
				fmt.Println("Ack Error")
				return
			}

			s.NetworkOutMsgQueue() <- ack
			s.NetworkOutMsgQueue() <- DBM
			s.InMsgQueue() <- ack
			s.InMsgQueue() <- DBM
		}
	}
}

func (s *State) ProcessSignPL(dbheight uint32, commitChain interfaces.IMsg) {
	s.ProcessLists.Get(dbheight).SetSigComplete(true)
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

func (s *State) GetFactoidState() interfaces.IFactoidState {
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
	return ab
}

func (s *State) NewAdminBlockHeader(dbheight uint32) interfaces.IABlockHeader {
	header := new(adminBlock.ABlockHeader)
	header.DBHeight = dbheight
	dbstate := s.DBStates.Last()
	if dbstate == nil {
		header.PrevFullHash = primitives.NewHash(constants.ZERO_HASH)
	} else {
		keymr, err := dbstate.AdminBlock.FullHash()
		if err != nil {
			panic(err.Error())
		}
		header.PrevFullHash = keymr
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

func (s *State) GetDB() interfaces.DBOverlay {
	return s.DB
}

func (s *State) SetDB(dbase interfaces.DBOverlay) {
	s.DB = databaseOverlay.NewOverlay(dbase)
}

func (s *State) GetDBHeightComplete() uint32 {
	db := s.GetDirectoryBlock()
	if db == nil {
		return 0
	}
	return db.GetHeader().GetDBHeight()
}

func (s *State) GetDirectoryBlock() interfaces.IDirectoryBlock {
	return s.DBStates.Last().DirectoryBlock
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

// Create a new Acknowledgement.  This Acknowledgement
func (s *State) NewAck(hash interfaces.IHash) (iack interfaces.IMsg, err error) {
	var last *messages.Ack
	if s.LastAck != nil {
		last = s.LastAck.(*messages.Ack)
	}
	ack := new(messages.Ack)
	ack.DBHeight = s.DBHeight
	ack.Timestamp = s.GetTimestamp()
	ack.MessageHash = hash
	if last == nil {
		ack.Height = 0
		ack.SerialHash = ack.MessageHash
	} else {
		ack.Height = last.Height + 1
		ack.SerialHash, err = primitives.CreateHash(last.MessageHash, ack.MessageHash)
		if err != nil {
			return nil, err
		}
	}
	s.SetLastAck(ack)

	// TODO:  Add the signature.

	return ack, nil
}

/**********************************************************************
 * Support
 **********************************************************************/

func (s *State) NewEOM(minute int) interfaces.IMsg {
	// The construction of the EOM message needs information from the state of
	// the server to create the proper serial hashes and such.  Right now
	// I am ignoring all of that.
	eom := new(messages.EOM)
	eom.Minute = byte(minute)
	eom.ServerIndex = s.ServerIndex
	eom.DirectoryBlockHeight = s.GetDBHeight()

	return eom
}
