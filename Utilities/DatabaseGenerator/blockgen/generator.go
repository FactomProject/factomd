package blockgen

import (
	"strings"

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
	FactomdState *state.State
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
	s.NetworkNumber = 1000
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
	g.FactomdState.DBStates.SaveDBStateToDB(sds)
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
}

func NewDefaultDBGeneratorConfig() *DBGeneratorConfig {
	c := new(DBGeneratorConfig)
	c.DBType = "level"
	c.DBPath = "factoid_level.db"
	c.FactomdConfigPath = "gen.conf"
	return c
}

func NewDBGenerator(c *DBGeneratorConfig) *DBGenerator {
	db := new(DBGenerator)
	db.FactomdState = NewGeneratorState(c.DBPath, c.DBType, c.FactomdConfigPath)

	db.loadGenesis()
	fmt.Println(db.FactomdState.DB.FetchDBlockHead())

	return db
}
