package snapshot

import (
	"github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/sirupsen/logrus"
)

type Snapshotter struct {
	log          *logrus.Logger
	db           *databaseOverlay.Overlay
	debugHeights []uint32
	stop         int64
	dumpDir      string

	balances *balanceSnapshot
}

type Config struct {
	Log          *logrus.Logger
	DB           *databaseOverlay.Overlay
	DebugHeights []uint32
	Stop         int64
	DumpDir      string
}

func New(cfg Config) *Snapshotter {
	s := &Snapshotter{
		log:          cfg.Log,
		db:           cfg.DB,
		debugHeights: cfg.DebugHeights,
		balances:     newBalanceSnapshot(),
		stop:         cfg.Stop,
		dumpDir:      cfg.DumpDir,
	}

	return s
}
