package state

import (
	"github.com/FactomProject/factomd/common/interfaces"

	"fmt"
	"github.com/FactomProject/factomd/database/hybridDB"
	"github.com/FactomProject/factomd/util"
	"log"
	"sync"
)

type State struct {
	once sync.Once
	cfg  interfaces.IFactomConfig

	leaderInMsgQueue   chan interfaces.IMsg
	inMsgQueue         chan interfaces.IMsg
	followerInMsgQueue chan interfaces.IMsg
	outMsgQueue        chan interfaces.IMsg
	//Network
	networkNumber int // Encoded into Directory Blocks

	// Number of Servers acknowledged by Factom
	totalServers int
	serverState  int                // (0 if client, 1 if server, 2 if audit server
	matryoshka   []interfaces.IHash // Reverse Hash

	// Database
	db interfaces.IDatabase

	// Directory Block State
	currentDirectoryBlock interfaces.IDirectoryBlock
	dBHeight              int

	// Message State
	lastAck interfaces.IMsg // Return the last Acknowledgement set by this server

	factoidState *FactoidState
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

func (s *State) OutMsgQueue() chan interfaces.IMsg {
	return s.outMsgQueue
}

//var _ IState = (*State)(nil)

// Getting the cfg state for Factom doesn't force a read of the config file unless
// it hasn;t been read yet.
func (s *State) Cfg() interfaces.IFactomConfig {
	s.once.Do(func() {
		log.Println("read factom config file: ", util.ConfigFilename())
		s.cfg = util.ReadConfig()
	})
	return s.cfg
}

// ReadCfg forces a read of the factom config file.  However, it does not change the
// state of any cfg object held by other processes... Only what will be returned by
// future calls to Cfg().
func (s *State) ReadCfg() interfaces.IFactomConfig {
	s.cfg = util.ReadConfig()
	return s.cfg
}

func (s *State) TotalServers() int {
	return s.totalServers
}

func (s *State) ServerState() int {
	return s.serverState
}

func (s *State) NetworkNumber() int {
	return s.networkNumber
}

func (s *State) Matryoshka() []interfaces.IHash {
	return s.matryoshka
}

func (s *State) LastAck() interfaces.IMsg {
	return s.lastAck
}

func (s *State) Init() {

	// Get our factomd configuration information.
	cfg := s.Cfg().(*util.FactomdConfig)

	s.inMsgQueue = make(chan interfaces.IMsg, 10000)         //incoming message queue for factom application messages
	s.leaderInMsgQueue = make(chan interfaces.IMsg, 10000)   //incoming message queue for factom application messages
	s.followerInMsgQueue = make(chan interfaces.IMsg, 10000) //incoming message queue for factom application messages
	s.outMsgQueue = make(chan interfaces.IMsg, 10000)        //outgoing message queue for factom application messages

	s.totalServers = 1
	s.serverState = 1

	//Database

	//Network
	switch cfg.App.Network {
	case "MAIN":
		s.networkNumber = 0
	case "TEST":
		s.networkNumber = 1
	case "LOCAL":
		s.networkNumber = 2
	case "CUSTOM":
		s.networkNumber = 3
	default:
		panic("Bad value for Network in factomd.conf")
	}

	s.InitLevelDB()
}

func (s *State) InitLevelDB() error {
	cfg := s.Cfg().(*util.FactomdConfig)
	path := cfg.App.LdbPath + "/" + cfg.App.Network + "/" + "factoid_level.db"

	fmt.Println("Creating Database at ", path)

	dbase, err := hybridDB.NewLevelMapHybridDB(path, false)

	if err != nil {
		return err
	}

	if dbase == nil {
		dbase, err = hybridDB.NewLevelMapHybridDB(path, true)
		if err != nil {
			return err
		}
	}

	//s.db = databaseOverlay.NewOverlay(dbase)
	return nil
}

func (s *State) InitBoltDB() error {
	//cfg := s.Cfg().(*util.FactomdConfig)
	//path := cfg.App.BoltDBPath + "/" + cfg.App.Network + "/" + "factoid_bolt.db"
	//dbase := hybridDB.NewBoltMapHybridDB(nil, path)
	//s.db = databaseOverlay.NewOverlay(dbase)
	return nil
}

func (s *State) String() string {
	return (s.cfg.(*util.FactomdConfig)).String()
}

func (s *State) NetworkName() string {
	return (s.Cfg().(util.FactomdConfig)).App.Network

}

func (s *State) NetworkPublicKey() []byte {
	return nil // TODO add our keys here...
}

func (s *State) CurrentDirectoryBlock() interfaces.IDirectoryBlock {
	return s.currentDirectoryBlock
}

func (s *State) SetCurrentDirectoryBlock(dirblk interfaces.IDirectoryBlock) {
	s.currentDirectoryBlock = dirblk
}

func (s *State) DB() interfaces.IDatabase {
	return s.db
}

func (s *State) SetDB(db interfaces.IDatabase) {
	s.db = db
}

func (s *State) DBHeight() int {
	return s.dBHeight
}

func (s *State) SetDBHeight(dbheight int) {
	s.dBHeight = dbheight
}
