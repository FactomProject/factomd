package snapshot

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/sirupsen/logrus"
)

type Snapshotter struct {
	log           *logrus.Logger
	db            *databaseOverlay.Overlay
	debugHeights  []uint32
	stop          int64
	dumpDir       string
	recordEntries bool

	balances *balanceSnapshot
	entries  *entrySnapshot
}

type Config struct {
	Log           *logrus.Logger
	DB            *databaseOverlay.Overlay
	DebugHeights  []uint32
	Stop          int64
	DumpDir       string
	RecordEntries bool
}

func New(cfg Config) (*Snapshotter, error) {
	s := &Snapshotter{
		log:           cfg.Log,
		db:            cfg.DB,
		debugHeights:  cfg.DebugHeights,
		balances:      newBalanceSnapshot(),
		entries:       NewEntrySnapshot(filepath.Join(cfg.DumpDir, "chains")),
		stop:          cfg.Stop,
		dumpDir:       cfg.DumpDir,
		recordEntries: cfg.RecordEntries,
	}

	if cfg.RecordEntries {
		err := os.MkdirAll(s.entries.Directory, 0777)
		if err != nil && !errors.Is(err, os.ErrExist) {
			return nil, fmt.Errorf("create dir %s: %w", s.entries.Directory, err)
		}
	}

	return s, nil
}
