package correctChainHeads

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/FactomProject/factomd/Utilities/tools"
	"github.com/FactomProject/factomd/common/interfaces"
	log "github.com/sirupsen/logrus"
)

// NOTE: This code has been deprecated but still exists as a standalone utility.
// See modules/chainheadfix for the new version.

type CorrectChainHeadConfig struct {
	CheckFloating bool
	Fix           bool
	PrintFreq     int
	Logger        *log.Logger
}

func NewCorrectChainHeadConfig() CorrectChainHeadConfig {
	return CorrectChainHeadConfig{
		PrintFreq: 500,
	}
}

func FindHeads(f tools.Fetcher, conf CorrectChainHeadConfig) {
	if conf.Logger == nil {
		conf.Logger = log.New()
		conf.Logger.SetLevel(log.InfoLevel)
	}
	if conf.PrintFreq == 0 {
		conf.PrintFreq = 500
	}

	flog := conf.Logger.WithFields(log.Fields{
		"tool": "chainheadtool",
	})
	checkFloating := conf.CheckFloating
	fix := conf.Fix
	chainHeads := make(map[string]interfaces.IHash)

	var allEblockLock sync.Mutex
	allEblks := make(map[[32]byte]interfaces.IHash)

	var err error
	var dblock interfaces.IDirectoryBlock

	head, err := f.FetchDBlockHead()
	if err != nil {
		panic(fmt.Sprintf("Error fetching head"))
	}

	if head == nil {
		// No head means database is empty
		return
	}

	height := head.GetDatabaseHeight()
	dblock = head
	top := height
	flog.Infof("Checking Chainheads starting at height: %d", height)
	errCount := 0
	waiting := new(int32)
	done := new(int32)
	total := 0

	var wg sync.WaitGroup
	allowedSimulataneous := 1000
	permission := make(chan bool, allowedSimulataneous)
	for i := 0; i < allowedSimulataneous; i++ {
		permission <- true
	}
	start := time.Now()

	doPrint := checkFloating
	go func() {
		for {
			if !doPrint {
				return
			}
			time.Sleep(10 * time.Second)
			v := atomic.LoadInt32(waiting)
			d := atomic.LoadInt32(done)
			flog.Infof("%d are still waiting. %d Done. Permission: %d", v, d, len(permission))
		}
	}()

	for ; height > 0; height-- {
		v := atomic.LoadInt32(waiting)
		for v > 50000 {
			time.Sleep(1 * time.Second)
		}

		dblock, err = f.FetchDBlockByHeight(height)
		if err != nil {
			flog.Errorf("Error fetching height %d: %s", height, err.Error())
			continue
		}

		eblockEnts := dblock.GetEBlockDBEntries()

		total += len(eblockEnts)
		if checkFloating {
			for i := 0; i < len(eblockEnts); i++ {
				wg.Add(1)
				atomic.AddInt32(waiting, 1)
				go func(eb interfaces.IDBEntry) {
					defer wg.Done()
					defer atomic.AddInt32(waiting, -1)
					<-permission
					defer func() {
						permission <- true
						atomic.AddInt32(done, 1)
					}()
					eblkF, err := f.FetchEBlock(eb.GetKeyMR())
					if err != nil {
						flog.Errorf("Error getting eblock %s for %s", eb.GetKeyMR().String(), eb.GetChainID().String())
						return
					}
					kmr, err := eblkF.KeyMR()
					if err != nil {
						flog.Errorf("Error getting eblock keymr %s for %s", eb.GetKeyMR().String(), eb.GetChainID().String())
						return
					}

					allEblockLock.Lock()
					allEblks[kmr.Fixed()] = eblkF.GetHeader().GetPrevKeyMR()
					allEblockLock.Unlock()
				}(eblockEnts[i])
			}
		}

		for _, eblk := range eblockEnts {
			if _, ok := chainHeads[eblk.GetChainID().String()]; ok {
				// Chainhead already exists
				continue
			}
			chainHeads[eblk.GetChainID().String()] = eblk.GetKeyMR()
			ch, err := f.FetchHeadIndexByChainID(eblk.GetChainID())
			if err != nil {
				flog.Errorf("Error getting chainhead for %s", eblk.GetChainID().String())
			} else {
				if !ch.IsSameAs(eblk.GetKeyMR()) {
					if fix {
						f.SetChainHeads([]interfaces.IHash{eblk.GetKeyMR()}, []interfaces.IHash{eblk.GetChainID()})
						flog.Warnf("{FIXED!} Chainhead found: %s, Expected %s :: For Chain: %s at height %d",
							ch.String(), eblk.GetKeyMR().String(), eblk.GetChainID().String(), height)
					} else {
						flog.Errorf("Chainhead found: %s, Expected %s :: For Chain: %s at height %d",
							ch.String(), eblk.GetKeyMR().String(), eblk.GetChainID().String(), height)
					}
					errCount++
				}
			}
		}
		if height%uint32(conf.PrintFreq) == 0 {
			d := atomic.LoadInt32(done)
			ps := float64(top-height) / time.Since(start).Seconds()
			flog.Infof("Currently on %d out of %d at %.3fp/s. %d Eblocks, %d done. %d ChainHeads so far. %d Are bad", height, top, ps, total, d, len(chainHeads), errCount)
		}

		var _ = dblock
	}

	if checkFloating {
		wg.Wait()
	}
	doPrint = false

	flog.Infof("%d Chains found in %f seconds", len(chainHeads), time.Since(start).Seconds())
	if fix {
		flog.Infof("Chainhead Check Complete. %d Errors corrected while checking for bad heads", errCount)
	} else {
		flog.Infof("Chainhead Check Complete. %d Errors found checking for bad heads", errCount)
	}

	errCount = 0
	if checkFloating {
		flog.Infof("Checking all EBLK links")
		for k, h := range chainHeads {
			var prev interfaces.IHash
			prev = h
			for {
				if prev.IsZero() {
					break
				}
				p, ok := allEblks[prev.Fixed()]
				if !ok {
					errCount++
					flog.Infof("Error finding Eblock %s for chain %s", h.String(), k)
				}
				delete(allEblks, prev.Fixed())
				prev = p
			}
		}
		flog.Infof("Floating Check Complete. %d Eblocks remain unaccounted for", len(allEblks))
		for k, h := range allEblks {
			flog.Infof("		|- %x missing. Prev: %s", k, h.String())
		}
	}
}
