package snapshot

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

func (s *Snapshotter) Dump() error {
	if s.dumpDir == "" {
		s.log.Debug("no dump directory to write to")
		return nil
	}

	_, err := os.Stat(s.dumpDir)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("stat %s: %w", s.dumpDir, err)
	}

	if os.IsNotExist(err) {
		err = os.Mkdir(s.dumpDir, 0766)
		if err != nil {
			return fmt.Errorf("mkdir %s: %w", s.dumpDir, err)
		}
	}

	// Dump balances
	balPath := filepath.Join(s.dumpDir, "balances")
	file, err := os.OpenFile(balPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0766)
	if err != nil {
		return fmt.Errorf("open %s: %w", balPath, err)
	}
	defer file.Close()
	err = s.balances.Dump(file)
	if err != nil {
		return fmt.Errorf("dump balances to %s: %w", balPath, err)
	}

	return nil
}

func (s *Snapshotter) WalkDB() error {
	db := s.db

	topDBlock, err := db.FetchDBlockHead()
	if err != nil {
		return fmt.Errorf("fetch head: %w", err)
	}

	topHeight := topDBlock.GetDatabaseHeight()
	if s.stop >= 0 {
		topHeight = uint32(s.stop)
	}
	start := time.Now()
	for i := uint32(0); i <= topHeight; i++ {
		if i%1000 == 0 && i > 0 {
			bps := float64(i) / time.Since(start).Seconds()
			remain := topHeight - i
			etaSecs := float64(remain) / bps
			eta := time.Duration(etaSecs * 1e9)

			s.log.WithFields(logrus.Fields{
				"done":   i,
				"remain": remain,
				"total":  topHeight,
				"bps":    fmt.Sprintf("%.2f", bps),
				// Take this ETA with a grain of salt. Dense blocks take A LOT longer than smaller ones.
				"eta": eta.String(),
			}).Debug("completed")
		}

		err = s.balances.Process(s.log, db, i, (i%10000 == 0 || i == topHeight) && i > 0)
		if err != nil {
			return fmt.Errorf("balance snapshot, height %d: %w", i, err)
		}
	}

	return nil
}
