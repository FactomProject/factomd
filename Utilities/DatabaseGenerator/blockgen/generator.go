package blockgen

import (
	"fmt"
	"strings"
	"time"

	"github.com/FactomProject/factomd/modules/registry"
	"github.com/FactomProject/factomd/modules/worker"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/identity"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/messages/electionMsgs"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/database/boltdb"
	"github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/leveldb"
	"github.com/FactomProject/factomd/database/mapdb"
	"github.com/FactomProject/factomd/state"
	log "github.com/sirupsen/logrus"
)

// DBGenerator is able to create a database given a defined config
type DBGenerator struct {
	// We need to process blocks to get state values
	// and ensure the data is correct
	FactomdState *state.State
	// Defines the blocks created and data in the db
	BlockGenerator *BlockGen

	last   *state.DBState
	config *DBGeneratorConfig
}

func NewDBGenerator(c *DBGeneratorConfig) (*DBGenerator, error) {
	var err error
	db := new(DBGenerator)
	db.config = c
	starttime := primitives.NewTimestampFromSeconds(uint32(time.Now().Add(-1 * 364 * 24 * time.Hour).Unix()))

	if c.StartTime != "" {
		starttimeT, err := time.Parse(c.TimeFormat(), c.StartTime)
		if err != nil {
			panic(err)
		}
		starttime = primitives.NewTimestampFromSeconds(uint32(starttimeT.Unix()))
	}
	// We need the factomd state to use it's functions
	db.FactomdState = NewGeneratorState(c, starttime)
	db.BlockGenerator, err = NewBlockGen(*c)
	if err != nil {
		return nil, err
	}

	head, err := db.FactomdState.DB.FetchDBlockHead()
	if err != nil || head == nil || head.GetDatabaseHeight() == 0 {
		db.loadGenesis()
	} else {
		// Load from db instead of genesis
		log.Infof("Starting at block height %d", head.GetDatabaseHeight())
		msg, err := db.FactomdState.LoadDBState(head.GetDatabaseHeight())
		if err != nil {
			return nil, err
		}
		dbs := db.msgToDBState(msg.(*messages.DBStateMsg))
		db.last = dbs
		dbs.Saved = true
		db.FactomdState.DBStates.Put(dbs)
	}

	return db, nil
}

func NewGeneratorState(conf *DBGeneratorConfig, starttime interfaces.Timestamp) *state.State {
	s := new(state.State)
	s.TimestampAtBoot = starttime
	s.SetLeaderTimestamp(starttime)
	s.Balancehash = primitives.NewZeroHash()
	var db interfaces.IDatabase
	var err error
	switch strings.ToLower(conf.DBType) {
	case "level", "ldb":
		db, err = leveldb.NewLevelDB(conf.DBPath, true)
	case "bolt":
		db = boltdb.NewBoltDB(nil, conf.DBPath)
	case "map":
		db = new(mapdb.MapDB)
	}
	if err != nil {
		panic(err)
	}

	s.DB = databaseOverlay.NewOverlay(db)
	s.LoadConfig(conf.FactomdConfigPath, "CUSTOM")
	s.StateSaverStruct.FastBoot = false
	s.EFactory = new(electionMsgs.ElectionsFactory)
	p := registry.New()
	p.Register(func(w *worker.Thread) { s.Initialize(w, new(electionMsgs.ElectionsFactory)) })
	go p.Run()
	p.WaitForRunning()
	s.NetworkNumber = constants.NETWORK_CUSTOM

	customnetname := conf.CustomNetID
	s.CustomNetworkID = primitives.Sha([]byte(customnetname)).Bytes()[:4]

	var blkCnt uint32
	head, err := s.DB.FetchDBlockHead()
	if err == nil && head != nil {
		blkCnt = head.GetHeader().GetDBHeight()
	}
	s.DBHeightAtBoot = blkCnt
	list := s.DBStates
	list.Base = blkCnt
	list.ProcessHeight = blkCnt
	s.ProcessLists.DBHeightBase = blkCnt
	return s
}

// loadGenesis loads the gensis block for given config and saves it to disk as a starting point
func (g *DBGenerator) loadGenesis() {
	var err error
	fmt.Println("\n***********************************")
	fmt.Println("******* New Database **************")
	fmt.Println("***********************************")

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
	g.FactomdState.DBStates.Last().Saved = true
}

// SaveDBState will save a dbstate to disk
func (g *DBGenerator) SaveDBState(dbstate *state.DBState) {
	dbstate.ReadyToSave = true
	dbstate.Signed = true
	dbstate.SaveStruct = new(state.SaveState)
	dbstate.SaveStruct.IdentityControl = identity.NewIdentityManager()
	g.FactomdState.DBStates.ProcessHeight = dbstate.DirectoryBlock.GetDatabaseHeight()
	put := g.FactomdState.DBStates.Put(dbstate)
	if !put {
		log.Warnf("%d Not put in dbstate list", dbstate.DirectoryBlock.GetDatabaseHeight())
	}
	progress := g.FactomdState.DBStates.SaveDBStateToDB(dbstate)
	if !progress {
		log.Warnf("%d Not saved to disk", dbstate.DirectoryBlock.GetDatabaseHeight())
	}

EntryLoop:
	for {
		select {
		case ent := <-g.FactomdState.WriteEntry:
			g.FactomdState.GetDB().InsertEntry(ent)
		default:
			break EntryLoop
		}
	}

	dbstate.Saved = true
	g.FactomdState.DBStates.Complete = dbstate.DirectoryBlock.GetDatabaseHeight() - g.FactomdState.DBStates.Base
	g.FactomdState.ProcessLists.DBHeightBase = dbstate.DirectoryBlock.GetDatabaseHeight()
	g.FactomdState.ProcessLists.Lists = make([]*state.ProcessList, 0)
}

func (g *DBGenerator) msgToDBState(msg *messages.DBStateMsg) *state.DBState {
	return g.FactomdState.DBStates.NewDBState(false,
		msg.DirectoryBlock,
		msg.AdminBlock,
		msg.FactoidBlock,
		msg.EntryCreditBlock,
		msg.EBlocks,
		msg.Entries)
}

type DBGeneratorConfig struct {
	DBPath            string
	DBType            string
	FactomdConfigPath string
	CustomNetID       string
	StartTime         string
	EntryGenerator    string
	LoopsPerPrint     int

	EntryGenConfig EntryGeneratorConfig
}

func (DBGeneratorConfig) TimeFormat() string {
	return "02 Jan 2006 15:04"
}

func (c DBGeneratorConfig) FactomLaunch() string {
	return fmt.Sprintf("factomd -network=CUSTOM -customnet=%s", c.CustomNetID)
}

func NewDefaultDBGeneratorConfig() *DBGeneratorConfig {
	c := new(DBGeneratorConfig)
	c.DBType = "level"
	c.DBPath = "factoid_level.db"
	c.FactomdConfigPath = "gen.conf"
	c.EntryGenConfig = *NewDefaultEntryGeneratorConfig()
	c.StartTime = time.Now().Add(-1 * 364 * 24 * time.Hour).Format(c.TimeFormat())
	c.LoopsPerPrint = 10
	return c
}

// CreateBlocks actually creates the blocks and saves them to disk
func (g *DBGenerator) CreateBlocks(amt int) error {
	start := time.Now()
	loop := time.Now()
	loopper := g.config.LoopsPerPrint
	totalEntries := 0
	loopEntries := 0 // Entries per loop
	for i := 0; i < amt; i++ {
		for count := 0; count < g.FactomdState.ElectionsQueue().Length(); count++ {
			g.FactomdState.ElectionsQueue().Dequeue() // Always have to do a dequeue
		}
		if i%loopper == 0 && i != 0 {
			totalDuration := time.Since(start).Seconds()
			duration := time.Since(loop).Seconds()
			avgb := float64(i) / totalDuration // avg block/s
			left := float64(amt - i)
			timeleft := left / avgb

			log.Infof("Current Height %5d:  %6d/%-6d at %6.2f b/s. Entries at %8.2f e/s. Avg Entry Rate: %8.2f e/s. ~%-12s Remain. Total Entries: %d",
				g.last.DirectoryBlock.GetDatabaseHeight(), i, amt,
				avgb, //float64(loopper)/duration,
				float64(loopEntries)/duration,
				float64(totalEntries)/totalDuration,
				time.Duration(timeleft)*time.Second,
				totalEntries,
			)
			loopEntries = 0
			loop = time.Now()
		}

		dbstate, err := g.BlockGenerator.NewBlock(g.last, g.FactomdState.GetNetworkID(), g.FactomdState.GetLeaderTimestamp())
		if err != nil {
			return err
		}

		loopEntries += len(dbstate.Entries)
		totalEntries += len(dbstate.Entries)

		g.SaveDBState(dbstate)
		g.last = dbstate

	}
	log.Infof("%d blocks saved to db", amt)
	return nil
}
