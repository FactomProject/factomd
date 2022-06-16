package snapshot

import (
	"path/filepath"

	"github.com/FactomProject/factomd/Utilities/tools"

	"github.com/FactomProject/factomd/Utilities/snapshot/internal"

	"github.com/sirupsen/logrus"
)

var DBHeight uint32

type Snapshotter struct {
	log           *logrus.Logger
	db            tools.Fetcher
	debugHeight   uint32
	stop          int64
	dumpDir       string
	recordEntries bool

	balances *BalanceSnapshot
	entries  *entrySnapshot
}

type Config struct {
	Log           *logrus.Logger
	DB            tools.Fetcher
	DebugHeight   uint32
	Stop          int64
	DumpDir       string
	RecordEntries bool
}

func New(cfg Config) (*Snapshotter, error) {
	s := &Snapshotter{
		log:           cfg.Log,
		db:            cfg.DB,
		debugHeight:   cfg.DebugHeight,
		balances:      newBalanceSnapshot(),
		entries:       NewEntrySnapshot(filepath.Join(cfg.DumpDir, internal.DefaultChainDir)),
		stop:          cfg.Stop,
		dumpDir:       cfg.DumpDir,
		recordEntries: cfg.RecordEntries,
	}

	return s, nil
}
