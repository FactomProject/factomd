package blockgen

import (
	"strings"

	"git.factoid.org/factomd/common/constants"

	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/messages/electionMsgs"

	"fmt"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/database/boltdb"
	"github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/leveldb"
	"github.com/FactomProject/factomd/database/mapdb"
	"github.com/FactomProject/factomd/state"
)

type DBGenerator struct {
	// We need to process blocks to get state values
	// and ensure the data is correct
	FactomdState   *state.State
	BlockGenerator *BlockGen

	last *state.DBState
}

func NewDBGenerator(c *DBGeneratorConfig) (*DBGenerator, error) {
	var err error
	db := new(DBGenerator)
	db.FactomdState = NewGeneratorState(c.DBPath, c.DBType, c.FactomdConfigPath)
	db.BlockGenerator, err = NewBlockGen(c.EntryGenConfig)
	if err != nil {
		return nil, err
	}
	db.loadGenesis() // TODO: Load from db?

	return db, nil
}

func NewGeneratorState(dbpath, dbtype string, configpath string) *state.State {
	s := new(state.State)
	s.SetLeaderTimestamp(primitives.NewTimestampFromMilliseconds(0))
	var db interfaces.IDatabase
	var err error
	switch strings.ToLower(dbtype) {
	case "level", "ldb":
		db, err = leveldb.NewLevelDB(dbpath, true)
	case "bolt":
		db = boltdb.NewBoltDB(nil, dbpath)
	case "map":
		db = new(mapdb.MapDB)
	}
	if err != nil {
		panic(err)
	}

	s.DB = databaseOverlay.NewOverlay(db)
	s.LoadConfig(configpath, "CUSTOM")
	s.EFactory = new(electionMsgs.ElectionsFactory)
	s.Init()
	s.NetworkNumber = constants.NETWORK_CUSTOM

	customnetname := "gen"
	s.CustomNetworkID = primitives.Sha([]byte(customnetname)).Bytes()[:4]
	return s
}

func (g *DBGenerator) loadGenesis() {
	var err error
	fmt.Println("\n***********************************")
	fmt.Println("******* New Database **************")
	fmt.Println("***********************************\n")

	var customIdentity interfaces.IHash
	if g.FactomdState.Network == "CUSTOM" {
		customIdentity, err = primitives.HexToHash(g.FactomdState.CustomBootstrapIdentity)
		if err != nil {
			panic(fmt.Sprintf("Could not decode Custom Bootstrap Identity (likely in config file) found: %s\n", g.FactomdState.CustomBootstrapIdentity))
			panic(err)
		}
	}
	dblk, ablk, fblk, ecblk := state.GenerateGenesisBlocks(g.FactomdState.GetNetworkID(), customIdentity)
	msg := messages.NewDBStateMsg(g.FactomdState.GetTimestamp(), dblk, ablk, fblk, ecblk, nil, nil, nil)
	// last block, flag it.
	dbstate, _ := msg.(*messages.DBStateMsg)
	dbstate.IsLast = true // this is the last DBState in this load
	// this will cause s.DBFinished to go true
	sds := g.msgToDBState(dbstate)
	sds.ReadyToSave = true
	sds.Signed = true
	g.FactomdState.DBStates.NewDBState(false, sds.DirectoryBlock, sds.AdminBlock, sds.FactoidBlock, sds.EntryCreditBlock, sds.EntryBlocks, sds.Entries)
	g.FactomdState.DBStates.SaveDBStateToDB(sds)
	sds.Saved = true
	g.last = sds
}

func (g *DBGenerator) SaveDBState(dbstate *state.DBState) {
	dbstate.ReadyToSave = true
	dbstate.Signed = true
	put := g.FactomdState.DBStates.Put(dbstate)
	if !put {
		fmt.Printf("%d Not put in dbstate list\n", dbstate.DirectoryBlock.GetDatabaseHeight())
	}
	g.FactomdState.DBStates.SaveDBStateToDB(dbstate)
}

// dbState := s.DBStates.NewDBState(isNew, directoryBlock, adminBlock, factoidBlock, entryCreditBlock, eBlocks, entries)

func (g *DBGenerator) msgToDBState(msg *messages.DBStateMsg) *state.DBState {
	return g.FactomdState.DBStates.NewDBState(false,
		msg.DirectoryBlock,
		msg.AdminBlock,
		msg.FactoidBlock,
		msg.EntryCreditBlock,
		msg.EBlocks,
		msg.Entries)
}

// There are queues that get filled up from internal state operations.
func (db *DBGenerator) drain() {
	// Election queue?

}

type DBGeneratorConfig struct {
	DBPath            string
	DBType            string
	FactomdConfigPath string

	EntryGenConfig EntryGeneratorConfig
}

func NewDefaultDBGeneratorConfig() *DBGeneratorConfig {
	c := new(DBGeneratorConfig)
	c.DBType = "level"
	c.DBPath = "factoid_level.db"
	c.FactomdConfigPath = "gen.conf"
	c.EntryGenConfig = *NewDefaultEntryGeneratorConfig()
	return c
}

func (g *DBGenerator) CreateBlocks(amt int) error {
	for i := 0; i < amt; i++ {
		dbstate, err := g.BlockGenerator.NewBlock(g.last, g.FactomdState.GetNetworkID())
		if err != nil {
			return err
		}
		g.SaveDBState(dbstate)
		g.last = dbstate
		fmt.Println(i)
	}
	return nil
}
